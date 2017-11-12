// Copyright 2016-17 The Go and Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/grolang/gro/macros"
	"github.com/grolang/gro/syntax/src"
)

const (
	sysLib    = "\"github.com/grolang/gro/sys\""
	dynLib    = "\"github.com/grolang/gro/ops\""
	utf88Lib  = "\"github.com/grolang/gro/utf88\""
	assertLib = "\"github.com/grolang/gro/assert\""
)

const debug = false
const trace = false

const lineMax = 1<<24 - 1 // TODO(gri) this limit is defined for src.Pos - fix

//--------------------------------------------------------------------------------
type parser struct {
	base *src.PosBase
	errh ErrorHandler
	mode Mode
	scanner

	first  error  // first error encountered
	errcnt int    // number of errors encountered
	pragma Pragma // pragma flags

	fnest  int    // function nesting level (for error handling)
	xnest  int    // expression nesting level (for complit ambiguity resolution)
	indent []byte // tracing support

	//project, package, and section-level...
	getFile func(string) (string, error) // function for callback to read in another file
	permits map[string]bool

	currProj   *Project
	argImports []*ImportDecl
	paramdPkgs map[string]*Package
	infImports []*ImportDecl // infered imports based on occurrences of, say, "fmt".Println
	infImpMap  map[string]string

	currPkg      *Package
	pkgName      string // name of current package
	lineDirects  bool
	dynamicBlock string

	sectHasMain  bool // does the current section have a main() function?
	sectHasStmts bool // does the current section have standalone stmts?
	sectIsMain   bool // was the current section headed with the "main" keyword
	sectIsTc     bool // was the current section headed with the "testcode" keyword

	docComments []string // buffer

	useRegistry   map[string]func(int, []string) []interface{}
	macroRegistry map[string]func(*parser, ...interface{}) Stmt
}

//--------------------------------------------------------------------------------
func (p *parser) init(
	filename string, base *src.PosBase, r io.Reader, errh ErrorHandler, pragh PragmaHandler, mode Mode, getFile func(string) (string, error),
) {
	p.base = base
	p.errh = errh
	p.mode = mode
	p.scanner.init(
		r,
		// Error and pragma handlers for scanner.
		// Because the (line, col) positions passed to these
		// handlers are always at or after the current reading
		// position, it is save to use the most recent position
		// base to compute the corresponding Pos value.
		func(line, col uint, msg string) {
			p.errorAt(p.posAt(line, col), msg)
		},
		func(line, col uint, text string) {
			if strings.HasPrefix(text, "line ") {
				p.updateBase(line, col+5, text[5:])
				return
			}
			if pragh != nil {
				p.pragma |= pragh(p.posAt(line, col), text)
			}
		},
	)

	p.first = nil
	p.errcnt = 0
	p.pragma = 0

	p.fnest = 0
	p.xnest = 0
	p.indent = nil

	p.getFile = getFile
	p.argImports = []*ImportDecl{}
	p.paramdPkgs = map[string]*Package{}

	p.currProj = &Project{}
	absName, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		p.error("error computing absolute name for " + filename)
	}
	p.currProj.Locn = filepath.ToSlash(absName)
	firstGoPath := filepath.SplitList(os.Getenv("GOPATH"))[0]
	srcPath := filepath.ToSlash(filepath.Join(firstGoPath, "src"))
	p.currProj.Root = strings.TrimPrefix(strings.TrimPrefix(p.currProj.Locn, srcPath), "/")

	b := filepath.Base(filename)
	ext := filepath.Ext(b)
	p.currProj.Name = b[:len(b)-len(ext)]
	if len(ext) > 0 && ext[0] == '.' {
		p.currProj.FileExt = ext[1:]
	}
	p.permits = map[string]bool{}
	switch p.currProj.FileExt {
	case "groo":
		p.dynamicMode = true
		p.dynamicBlock = "groo"
		fallthrough
	case "gro", "":
		for _, kw := range groProfile {
			p.permits[kw] = true
		}
		fallthrough
	case "go":
		for _, kw := range goProfile {
			p.permits[kw] = true
		}
	}

	p.macroRegistry = map[string]func(*parser, ...interface{}) Stmt{
		"prepare": func(p *parser, _ ...interface{}) Stmt {
			if !p.permits["prepare"] {
				p.syntaxError("\"prepare\" command disabled but is present")
				return nil
			}
			return GroSystemCmd(p, "Prepare")
		},
		"execute": func(p *parser, _ ...interface{}) Stmt {
			if !p.permits["execute"] {
				p.syntaxError("\"execute\" command disabled but is present")
				return nil
			}
			return GroSystemCmd(p, "Execute")
		},
		"run": func(p *parser, _ ...interface{}) Stmt {
			if !p.permits["run"] {
				p.syntaxError("\"run\" command disabled but is present")
				return nil
			}
			return GroSystemCmd(p, "Run")
		},
		"test": func(p *parser, _ ...interface{}) Stmt {
			if !p.permits["test"] {
				p.syntaxError("\"test\" command disabled but is present")
				return nil
			}
			return GroSystemCmd(p, "Test")
		},
		"assert": func(p *parser, _ ...interface{}) Stmt {
			if !p.permits["assert"] {
				p.syntaxError("\"assert\" macro disabled but is present")
				return nil
			}
			return Assert(p)
		},
		"let": func(p *parser, rest ...interface{}) Stmt {
			if !p.permits["let"] {
				p.syntaxError("\"let\" macro disabled but is present")
				return nil
			}
			if len(rest) != 1 {
				panic("argument error with \"let\" macro")
			}
			if stmt, ok := rest[0].(func() Stmt); ok {
				return Let(p, stmt)
			}
			panic("argument error with \"let\" macro")
		},
	}
	p.useRegistry = map[string]func(int, []string) []interface{}{
		"blacklist": func(numRets int, args []string) []interface{} {
			macros.InitBlacklist(p, args)
			return nil
		},
		"grosys": func(numRets int, args []string) []interface{} {
			macros.GroSys(p, args)
			return nil
		},
	}
}

//--------------------------------------------------------------------------------
// GroSystemCmd: to xfer to macros, we need to make public:
// * _Semi, _Rparen, <etc>
// * StringLit
// * p.oliteral, p.syntaxError, p.advance, p.procImportAlias
// * Name, ExprStmt, CallExpr, <etc>
//
func GroSystemCmd(p *parser, s string) Stmt {
	fn := p.oliteral()
	if fn == nil || fn.Kind != StringLit {
		p.syntaxError("missing filename for " + s)
		p.advance(_Semi, _Rparen)
		return nil
	} else {
		a := p.procImportAlias(&BasicLit{Value: sysLib, Kind: StringLit}, "")
		es := &ExprStmt{
			X: &CallExpr{
				Fun: &SelectorExpr{
					X:   &Name{Value: a},
					Sel: &Name{Value: s},
				},
				ArgList: []Expr{fn},
			},
		}
		return es
	}
}

//--------------------------------------------------------------------------------
// Assert: to xfer to macros, we need to make public:
// * _Semi, _Rparen, <etc>
// * BasicLit, StringLit
// * p.expr, p.procImportAlias
// * Name, ExprStmt, CallExpr, <etc>
//
func Assert(p *parser) Stmt {
	e := p.expr()
	p.procImportAlias(&BasicLit{Value: assertLib, Kind: StringLit}, "assert")
	es := &ExprStmt{
		X: &CallExpr{
			Fun: &SelectorExpr{
				X:   &Name{Value: "assert"},
				Sel: &Name{Value: "AssertTrue"},
			},
			ArgList: []Expr{e},
		},
	}
	return es
}

//--------------------------------------------------------------------------------
func Let(p *parser, stmt func() Stmt) Stmt {
	pos := p.pos()
	lhs := p.exprList()
	if p.tok == _Assign {
		// expr_list '=' expr_list
		p.next()
		l := []Stmt{
			p.newAssignStmt(pos, Def, lhs, p.exprList()),
		}
		p.got(_Semi)
		l = append(l, p.tlStmtList(stmt)...)
		return &BlockStmt{
			List: l,
		}
	}
	p.syntaxError("invalid syntax in \"let\" statement")
	return nil
}

//--------------------------------------------------------------------------------
var goProfile = [...]string{
	"package", "import", "const", "var", "type" /*TODO:*/, "func", "map", "chan", "struct", "interface",
	/*TEST:*/ "if", "else", "switch", "case", "default", "fallthrough", "select",
	"for" /*TEST:*/, "range", "goto", "return" /*TODO:*/, "defer", "go", "break", "continue",
}

var groProfile = [...]string{
	"inferPkg", "multiPkg", "genericCall", "genericDef", "inplaceImps", "inferMain", "pkgSectBlocks",
	"project", "use", "include", "internal", "section", "main", "testcode", "proc" /*TODO:*/, "do",
	/*TODO:*/ "assert", "let", "prepare", "execute", "run", "test",
}

//--------------------------------------------------------------------------------
func (p *parser) SetPermit(s string) {
	p.permits[s] = true
}

func (p *parser) UnsetPermit(s string) {
	p.permits[s] = false
}

func (p *parser) IsPermit(s string) bool {
	return p.permits[s]
}

//--------------------------------------------------------------------------------
func (p *parser) procImportAlias(lit *BasicLit, a string) string {
	if a == "" {
		_, a = filepath.Split(strings.Trim(lit.Value, "\""))
	}
	if p.infImpMap[a] != "" && p.infImpMap[a] != lit.Value {
		p.syntaxError(fmt.Sprintf("import alias \"%s\" has already been used but with different import path", a))
		return a
	} else if p.infImpMap[a] == "" {
		p.infImports = append(p.infImports, &ImportDecl{
			Path:         &BasicLit{Value: lit.Value, Kind: StringLit},
			LocalPkgName: &Name{Value: a},
		})
		p.infImpMap[a] = lit.Value
	}
	return a
}

//--------------------------------------------------------------------------------
func (p *parser) updateBase(line, col uint, text string) {
	// Want to use LastIndexByte below but it's not defined in Go1.4 and bootstrap fails.
	i := strings.LastIndex(text, ":") // look from right (Windows filenames may contain ':')
	if i < 0 {
		return // ignore (not a line directive)
	}
	nstr := text[i+1:]
	n, err := strconv.Atoi(nstr)
	if err != nil || n <= 0 || n > lineMax {
		p.errorAt(p.posAt(line, col+uint(i+1)), "invalid line number: "+nstr)
		return
	}
	p.base = src.NewLinePragmaBase(src.MakePos(p.base.Pos().Base(), line, col), text[:i], uint(n))
}

//--------------------------------------------------------------------------------
// Package files
//--------------------------------------------------------------------------------
//
// Parse methods are annotated with matching Go productions as appropriate.
// The annotations are intended as guidelines only since a single Go grammar
// rule may be covered by multiple parse methods and vice versa.
//
// Excluding methods returning slices, parse methods named xOrNil may return
// nil; all others are expected to return a valid non-nil node.

//--------------------------------------------------------------------------------
// Project
func (p *parser) files() map[string]*File {
	if trace {
		defer p.trace("filesOrNil")("")
	}

	pkgs := p.pkgs()
	if !p.currProj.ExplicitKw && len(pkgs) == 0 {
		p.syntaxErrorAt(src.MakePos(p.base, 1, 1), "gro-file empty")
		return nil
	} else if len(pkgs) == 0 {
		p.syntaxError("project keyword but no packages")
		return nil
	}
	fs := map[string]*File{}
	for _, pkg := range pkgs { // for each file in each pkg, add to map of files returned (fs)
		if len(pkgs) > 1 && !p.permits["multiPkg"] {
			p.syntaxErrorAt(pkgs[1].Pos(), "multi-packages disabled but more than one package present")
			return nil
		}
		if p.currProj.Str != "" {
			pkg.Dir = filepath.ToSlash(filepath.Join(p.currProj.Str, pkg.Dir))
		}
		// if project keyword or more than one package, all explicit packages have package-name both as filename and in directory name
		if p.currProj.ExplicitKw || len(pkgs) > 1 {
			pkg.Dir = filepath.ToSlash(filepath.Join(pkg.Dir, pkg.Name))
		}
		for _, f := range pkg.Files {
			f.OwnerPkg = pkg
			for _, decl := range f.DeclList {
				if imp, ok := decl.(*ImportDecl); ok {
					imp.OwnerFile = f
				}
			}
		}
		if len(pkg.Params) > 0 {
			p.paramdPkgs[filepath.ToSlash(filepath.Join(p.currProj.Root, pkg.Dir))] = pkg
			continue
		}
		for _, f := range pkg.Files {
			if f.SectName != "" {
				f.FileName = f.SectName
			} else if !p.currProj.ExplicitKw && len(pkgs) == 1 && pkg.Kw {
				f.FileName = p.currProj.Name //use gro-filename as filename
			}
			fs[filepath.ToSlash(filepath.Join(p.currProj.Root, pkg.Dir, f.FileName))+".go"] = f
		}
	}
	for _, ai := range p.argImports {
		for k, v := range p.initGenerics(ai, "", map[string]bool{}) {
			fs[k] = v
		}
	}
	return fs
}

//--------------------------------------------------------------------------------
// initGenerics: for each import that had an argument/s, put its parameters into the package as types, then add that to fs.
func (p *parser) initGenerics(ai *ImportDecl, prefix string, done map[string]bool) map[string]*File {
	fs := map[string]*File{}
	if !p.permits["genericCall"] {
		p.syntaxError("calling generic-type packages disabled but import arguments present")
		return nil
	}
	aiPkgLocn := strings.Trim(ai.Path.Value, "\"")
	/*if done[aiPkgLocn] {
		p.syntaxError("cycles not allowed in parameterized imports") //TODO: fix so error is different to that below
		return nil
	}*/
	pp := p.paramdPkgs[aiPkgLocn]
	if pp == nil {
		/*pps:= ""
		for k, _:= range p.paramdPkgs {
			pps += fmt.Sprintf("%s\n", k)
		}
		p.syntaxError(fmt.Sprintf("parameterized package \"%s\" not present in file. Files available are:\n%s", aiPkgLocn, pps))
		*/
		p.syntaxError("parameterized package not present in file, or there's a cycle in the parameterized imports")
		return nil
	}
	newpath := filepath.ToSlash(filepath.Join(prefix, "generics", ai.OwnerFile.OwnerPkg.Dir, ai.OwnerFile.FileName, ai.LocalPkgName.Value))
	for _, pf := range pp.Files {
		//recursively call initGenerics on each argImport in the parameterized pkg, where arg is one of pkg's params
		for _, decl := range pf.DeclList {
			if decl, ok := decl.(*ImportDecl); ok && len(decl.Args) > 0 {
				passThrus := map[int]int{} //map arg to param
				for m, arg := range decl.Args {
					if arg, ok := arg.(*Name); ok {
						for n, param := range pp.Params {
							if arg.Value == param.Value {
								passThrus[m] = n
							}
						}
					}
				}
				if len(passThrus) > 0 {
					args := []Expr{}
					for m, arg := range decl.Args {
						if n, ok := passThrus[m]; ok {
							args = append(args, ai.Args[n])
						} else {
							args = append(args, arg)
						}
					}
					infers := []*ImportDecl{}
					for _, infer := range decl.Infers {
						infers = append(infers, infer)
					}
					declClone := &ImportDecl{
						LocalPkgName: decl.LocalPkgName,
						Path:         decl.Path,
						Group:        decl.Group,
						OwnerFile:    decl.OwnerFile,
						Args:         args,
						Infers:       infers,
					}
					done[aiPkgLocn] = true
					for k, subFile := range p.initGenerics(declClone, newpath, done) {
						fs[k] = subFile
					}
					delete(done, aiPkgLocn)
				}
			}
		}
		fs[filepath.ToSlash(filepath.Join(p.currProj.Root, newpath, pf.FileName))+".go"] = pf
	}
	ai.Path.Value = "\"" + filepath.ToSlash(filepath.Join(p.currProj.Root, newpath)) + "\""
	f := &File{
		PkgName:  p.newName(pp.Name),
		DeclList: []Decl{},
	}
	g := new(Group)
	for _, inf := range ai.Infers {
		inf.Group = g
		f.DeclList = append(f.DeclList, inf)
	}
	for n, a := range ai.Args {
		f.DeclList = append(f.DeclList, &TypeDecl{
			Name:  pp.Params[n],
			Alias: true,
			Type:  a,
		})
	}
	fs[filepath.ToSlash(filepath.Join(p.currProj.Root, newpath, "generic_args"))+".go"] = f
	return fs
}

//--------------------------------------------------------------------------------
func (p *parser) forkAndParse(filename string, src []byte) (_ []*Package, first error) {
	defer func() {
		if pnc := recover(); pnc != nil {
			if err, ok := pnc.(Error); ok {
				first = err
				return
			}
			panic(pnc)
		}
	}()

	var q parser
	q.init(filename, p.base, &bytesReader{src}, p.errh, nil, p.mode, p.getFile)
	q.next()
	pkgs := q.pkgs()
	p.argImports = append(p.argImports, q.argImports...)
	return pkgs, q.first
}

//--------------------------------------------------------------------------------
func (p *parser) pkgs() (pkgs []*Package) {
	if trace {
		defer p.trace("pkgs")("")
	}

	p.currProj.ExplicitKw = false
	p.currProj.Str = ""
	if p.isName("project") {
		if !p.permits["project"] {
			p.syntaxError("\"project\" keyword disabled but keyword is present")
			return nil
		}
		if len(p.comments) != 0 {
			p.currProj.Doc = p.comments
		}
		p.next()
		if b := p.oliteral(); b != nil { // is directory-string present?
			p.currProj.Str = strings.Trim(b.Value, "\"")
		}
		p.currProj.Name = p.name().Value
		p.want(_Semi)
		p.currProj.ExplicitKw = true
	}
	for p.isName("use") || p.isName("include") {
		var fn func() []*Package
		switch p.lit {
		case "use":
			fn = p.useDecl
		case "include":
			fn = p.inclDecl
		}
		if !p.permits[p.lit] {
			p.syntaxError(fmt.Sprintf("\"%s\" keywords are disabled but keyword is present", p.lit))
			return nil
		}
		p.want(_Name)

		if p.tok == _Lparen {
			p.list(_Lparen, _Semi, _Rparen, func() bool {
				pkgs = append(pkgs, fn()...)
				return false
			})
		} else {
			pkgs = append(pkgs, fn()...)
		}
		p.want(_Semi)
	}
	for p.tok != _EOF {
		pkgs = append(pkgs, p.pkgOrNil())
	}
	return
}

//--------------------------------------------------------------------------------
func (p *parser) useDecl() (pkgs []*Package) {
	if trace {
		defer p.trace("useDecl")("")
	}

	rets := []string{}
	for p.tok == _Name {
		rets = append(rets, p.name().Value)
	}
	useStr := p.oliteral()
	if useStr == nil || useStr.Kind != StringLit {
		p.syntaxError("missing use string")
		p.advance(_Semi, _Rparen)
		return nil
	}
	use := strings.Trim(useStr.Value, "\"")
	args := []string{}
	if p.tok == _Lparen {
		p.list(_Lparen, _Comma, _Rparen, func() bool {
			args = append(args, p.oliteral().Value)
			return false
		})
	}

	switch use {
	/*default:
	p.syntaxError(fmt.Sprintf("use \"%s\" not implemented", use))
	p.advance(_Semi, _Rbrace)
	return
	*/
	case "blacklist":
		//TODO: ensure len(rets) == 0
		//p.useRegistry["blacklist"](args)
		p.useRegistry["blacklist"](len(rets), args)
		return
	case "linedirectives":
		//TODO: ensure len(rets) == 0 && len(args) == 0
		p.lineDirects = true
		return
	case "dynamic":
		if len(args) != 0 {
			p.syntaxError("use \"dynamic\" shouldn't take any arguments but does")
			p.advance(_Semi, _Rbrace)
			return
		}
		p.dynamicMode = true
		if len(rets) == 1 {
			ret := strings.Trim(rets[0], "\"")
			if ret == "_" {
				p.dynamicBlock = "groo"
			} else {
				p.macroRegistry[ret] = func(p *parser, _ ...interface{}) Stmt {
					dynBlock := p.dynamicBlock
					p.dynamicBlock = ret
					bs := p.blockStmt("", p.tlStmt)
					p.dynamicBlock = dynBlock
					return bs
				}
			}
		} else if len(rets) > 0 {
			p.syntaxError("use \"dynamic\" has too many return values")
			p.advance(_Semi, _Rbrace)
			return
		}
		return
	default:
		return
	}
}

//--------------------------------------------------------------------------------
func (p *parser) inclDecl() []*Package {
	if trace {
		defer p.trace("inclDecl")("")
	}

	lit := p.oliteral()
	if lit == nil {
		p.syntaxError("missing name for gro-file after \"include\" keyword")
		p.advance(_Semi, _Rparen)
		return nil
	}
	fileLocn := filepath.ToSlash(filepath.Join(p.currProj.Locn, strings.Trim(lit.Value, "\"")))
	src, err := p.getFile(fileLocn)
	if err != nil {
		p.error(fmt.Sprintf("error \"%s\" retrieving included file %s", err, fileLocn))
		return nil
	}
	pkgs, err := p.forkAndParse(fileLocn, []byte(src))
	if err != nil {
		p.error(fmt.Sprintf("error \"%s\" parsing included file %s", err, fileLocn))
		return nil
	}
	return pkgs
}

//--------------------------------------------------------------------------------
func (p *parser) newBlankFunc(s string) *FuncDecl {
	fn := &FuncDecl{
		Name: p.newName(s),
		Type: &FuncType{
			ParamList:  []*Field{},
			ResultList: nil,
		},
		Body: &BlockStmt{
			List:   []Stmt{},
			Rbrace: p.pos(),
		},
	}
	fn.pos = p.pos()
	fn.Type.pos = p.pos()
	fn.Body.pos = p.pos()
	return fn
}

//--------------------------------------------------------------------------------
// SourceFile = PackageClause ";" { ImportDecl ";" } { TopLevelDecl ";" } .
func (p *parser) pkgOrNil() *Package {
	if trace {
		defer p.trace("pkgOrNil")("")
	}

	pkg := &Package{
		Files:   []*File{},
		Params:  []*Name{},
		IdsUsed: map[string]bool{},
	}
	p.currPkg = pkg
	p.pkgName = "" // used by funcDeclOrNil()
	bracesUsed := false

	for { // each section
		f := new(File)
		f.pos = p.pos()
		pkg.pos = f.pos
		if p.lineDirects {
			f.linedirect = f.pos
		}
		f.comments = &Comments{Alone: []Comment{}, Above: []Comment{}, Left: []Comment{}, Right: []Comment{}, Below: []Comment{}}
		if p.currProj.Doc != nil {
			//add project-level doc to top of every package file
			for _, c := range p.currProj.Doc {
				f.comments.Alone = append(f.comments.Alone, Comment{Text: c})
			}
		}
		if p.tok == _Package || p.isName("internal") { // first time thru loop only, maybe
			if !p.permits["package"] {
				p.syntaxError("\"package\" (and similar) keywords are disabled but keyword is present")
				return nil
			}
			for i, c := range p.comments {
				if i >= len(p.comments)-p.numDocComments {
					f.comments.Above = append(f.comments.Above, Comment{Text: c})
				}
			}
			if p.isName("internal") {
				if !p.permits["internal"] {
					p.syntaxError("\"internal\" keywords disabled but keyword is present")
					return nil
				}
				pkg.Dir = "internal"
			}
			p.next()
			// if directory-string present, add after directory
			if b := p.oliteral(); b != nil {
				pkg.Dir = filepath.ToSlash(filepath.Join(pkg.Dir, strings.Trim(b.Value, "\"")))
			}
			pkg.Kw = true
			f.PkgName = p.name()
			pkg.Name = f.PkgName.Value
			p.pkgName = f.PkgName.Value

			if p.tok == _Lparen {
				if !p.permits["genericDef"] {
					p.syntaxError("defining generic packages is disabled but one is present")
					return nil
				}
				p.list(_Lparen, _Comma, _Rparen, func() bool {
					pkg.Params = append(pkg.Params, p.name())
					return false
				})
			}

			if p.got(_Lbrace) {
				if !p.permits["pkgSectBlocks"] {
					p.syntaxError("using block-style notation for packages and sections is disabled but it is being used")
					return nil
				}
				bracesUsed = true
			} else {
				p.want(_Semi)
			}
		} else if p.pkgName != "" { // subsequent times thru loop (section keyword)
			f.PkgName = p.newName(p.pkgName)
		} else { // first time thru loop, no package keyword
			if !p.permits["inferPkg"] {
				p.syntaxError("infer-packages disabled but no explicit \"package\" keyword present")
				return nil
			}
			// use project-name (or gro-filename) as filename and package-name
			f.PkgName = p.newName(p.currProj.Name)
		}

		f = p.sectionOrNil(f)
		f.FileName = f.PkgName.Value

		// if package without keyword, and main fn defined, use "main" as package-name
		if !pkg.Kw && p.sectHasMain && p.permits["inferMain"] ||
			p.sectIsMain && p.sectHasMain && p.permits["inferMain"] {
			f.PkgName.Value = "main"
			f.comments.Alone = append(f.comments.Alone, Comment{Text: "// +build ignore"})
		}
		if p.sectHasStmts && !p.sectHasMain && !pkg.Kw ||
			p.sectIsMain && !p.sectHasMain {
			if p.permits["inferMain"] {
				f.PkgName.Value = "main"
				f.comments.Alone = append(f.comments.Alone, Comment{Text: "// +build ignore"})
				f.DeclList = append(f.DeclList, p.newBlankFunc("main"))
			}
		}
		pkg.Files = append(pkg.Files, f)
		if p.tok == _EOF || p.tok == _Package || p.isName("internal") || (bracesUsed && p.tok == _Rbrace) {
			break
		}
	}
	if bracesUsed && !(p.got(_Rbrace) && p.got(_Semi)) {
		p.syntaxError("missing right brace after package block")
		p.advance(_Semi, _Rbrace)
		return nil
	}
	if p.dynamicMode && !pkg.IdsUsed["any"] && len(pkg.Files) > 0 {
		d := &TypeDecl{
			Name:   &Name{Value: "any"},
			Alias:  true,
			Type:   &InterfaceType{MethodList: []*Field{}},
			Pragma: p.pragma,
		}
		pkg.Files[0].DeclList = append(pkg.Files[0].DeclList, d)
		pkg.IdsUsed["any"] = true
	}
	return pkg
}

//--------------------------------------------------------------------------------
// Section
func (p *parser) sectionOrNil(f *File) *File {
	if trace {
		defer p.trace("sectionOrNil")("")
	}

	p.sectHasMain = false  // used by funcDeclOrNil()
	p.sectHasStmts = false // used by tlBlock()
	p.sectIsMain = false
	p.sectIsTc = false
	p.infImports = []*ImportDecl{} // used by operand()
	p.infImpMap = map[string]string{}
	bracesUsed := false

	currPos := p.pos()
	if p.isName("section", "main", "testcode") {
		if !p.permits["section"] {
			p.syntaxError("\"section\" keywords are disabled but keyword is present")
			return nil
		}
		if p.isName("main") && !p.permits["main"] {
			p.syntaxError("\"main\" keywords are disabled but keyword is present")
			return nil
		}
		if p.isName("testcode") && !p.permits["testcode"] {
			p.syntaxError("\"testcode\" keywords are disabled but keyword is present")
			return nil
		}
		p.sectIsMain = p.isName("main")
		for i, c := range p.comments {
			if i >= len(p.comments)-p.numDocComments {
				f.DeclList = append(f.DeclList, &CommentDecl{CommentList: []Comment{{Text: c}}})
			}
		}
		p.sectIsTc = p.isName("testcode")
		p.next()
		lit := p.oliteral()
		if lit == nil {
			p.syntaxError("missing section name")
			p.advance(_Semi, _Rparen)
			return nil
		}
		if p.got(_Lbrace) {
			if !p.permits["pkgSectBlocks"] {
				p.syntaxError("using block-style notation for packages and sections is disabled but it is being used")
				return nil
			}
			bracesUsed = true
		} else {
			p.want(_Semi)
		}
		f.SectName = strings.Trim(lit.Value, "\"")
		if p.sectIsTc {
			f.SectName += "_test"
		}
	}

	// { ImportDecl ";" }
	for p.got(_Import) {
		if !p.permits["import"] {
			p.syntaxError("\"import\" keywords are disabled but keyword is present")
			return nil
		}
		f.DeclList = p.appendGroup(f.DeclList, p.importDecl)
		p.want(_Semi)
	}
	if p.sectIsTc {
		f.PkgName.Value = p.pkgName
	}

	// { TopLevelDecl ";" }
	for p.tok != _EOF && p.tok != _Package && p.tok != _Rbrace && !p.isName("internal", "section", "main", "testcode") {
		switch p.tok {
		case _Const, _Var, _Type, _Func: //declarations
			f.DeclList = p.decl(f.DeclList)
		case _If, _For, _Switch, _Select, _Go, _Lbrace, _Semi, _Literal: //tl-stmts
			f.DeclList = append(f.DeclList, p.tlBlock())
		default:
			if p.isName("do") { //do-stmt
				f.DeclList = append(f.DeclList, p.tlBlock())
			} else if p.isName("proc") { //proc-decl
				f.DeclList = p.decl(f.DeclList)
			} else if _, ok := p.macroRegistry[p.lit]; p.tok == _Name && ok { //macros
				f.DeclList = append(f.DeclList, p.tlBlock())
			} else {
				p.syntaxError(fmt.Sprintf("unexpected %s at top-level", p.tok))
				return nil
			}
		}
	}
	if currPos == p.pos() && p.tok != _EOF {
		p.syntaxError(fmt.Sprintf("unexpected token %s", p.tok))
		return nil
	}

	if len(p.infImports) > 0 {
		g := new(Group)
		for i := len(p.infImports) - 1; i >= 0; i-- {
			imp := p.infImports[i]
			imp.Group = g
			f.DeclList = append([]Decl{imp}, f.DeclList...)
		}
	}
	if bracesUsed && !(p.got(_Rbrace) && p.got(_Semi)) {
		p.syntaxError("missing right brace after section block")
		p.advance(_Semi, _Rbrace)
		return nil
	}
	f.Lines = p.source.line
	return f
}

//--------------------------------------------------------------------------------
// TopLevelDecl
func (p *parser) decl(declList []Decl) []Decl {
	switch {
	case p.tok == _Const:
		if !p.permits["const"] {
			p.syntaxError("\"const\" keywords are disabled but keyword is present")
			return nil
		}
		for i, c := range p.comments {
			if i >= len(p.comments)-p.numDocComments {
				p.docComments = append(p.docComments, c)
			}
		}
		p.next()
		declList = p.appendGroup(declList, p.constDecl)

	case p.tok == _Type:
		if !p.permits["type"] {
			p.syntaxError("\"type\" keywords are disabled but keyword is present")
			return nil
		}
		for i, c := range p.comments {
			if i >= len(p.comments)-p.numDocComments {
				p.docComments = append(p.docComments, c)
			}
		}
		p.next()
		declList = p.appendGroup(declList, p.typeDecl)

	case p.tok == _Var:
		if !p.permits["var"] {
			p.syntaxError("\"var\" keywords are disabled but keyword is present")
			return nil
		}
		for i, c := range p.comments {
			if i >= len(p.comments)-p.numDocComments {
				p.docComments = append(p.docComments, c)
			}
		}
		p.next()
		declList = p.appendGroup(declList, p.varDecl)

	case p.tok == _Func:
		for i, c := range p.comments {
			if i >= len(p.comments)-p.numDocComments {
				p.docComments = append(p.docComments, c)
			}
		}
		p.next()
		if d := p.funcDeclOrNil(p.stmtOrNil); d != nil {
			declList = append(declList, d)
		}

	case p.isName("proc"):
		if !p.permits["proc"] {
			p.syntaxError("\"proc\" keywords are disabled but keyword is present")
			return nil
		}
		for i, c := range p.comments {
			if i >= len(p.comments)-p.numDocComments {
				p.docComments = append(p.docComments, c)
			}
		}
		p.next()
		if d := p.funcDeclOrNil(p.procStmt); d != nil {
			declList = append(declList, d)
		}

	default:
		if p.tok == _Lbrace && len(declList) > 0 && isEmptyFuncDecl(declList[len(declList)-1]) {
			// opening { of function declaration on next line
			p.syntaxError("unexpected semicolon or newline before {")
		} else {
			p.syntaxError("non-declaration statement outside function body")
		}
		p.advance(_Const, _Type, _Var, _Func)
		return declList
	}

	// Reset p.pragma BEFORE advancing to the next token (consuming ';')
	// since comments before may set pragmas for the next function decl.
	p.pragma = 0

	if p.tok != _EOF && !p.got(_Semi) {
		p.syntaxError("after top level declaration")
		p.advance(_Const, _Type, _Var, _Func)
	}

	return declList
}

//--------------------------------------------------------------------------------
// Declarations
//--------------------------------------------------------------------------------
// list parses a possibly empty, sep-separated list, optionally
// followed by sep and enclosed by ( and ) or { and }. open is
// one of _Lparen, or _Lbrace, sep is one of _Comma or _Semi,
// and close is expected to be the (closing) opposite of open.
// For each list element, f is called. After f returns true, no
// more list elements are accepted. list returns the position
// of the closing token.
//
// list = "(" { f sep } ")" |
//        "{" { f sep } "}" . // sep is optional before ")" or "}"
//
func (p *parser) list(open, sep, close token, f func() bool) src.Pos {
	p.want(open)

	var done bool
	for p.tok != _EOF && p.tok != close && !done {
		done = f()

		// sep is optional before close
		if !p.got(sep) && p.tok != close {
			p.syntaxError(fmt.Sprintf("expecting %s or %s", tokstring(sep), tokstring(close)))
			p.advance(_Rparen, _Rbrack, _Rbrace)
			if p.tok != close {
				return p.pos()
				// position could be better but we had an error so we don't care
			}
		}
	}

	pos := p.pos()
	p.want(close)
	return pos
}

//--------------------------------------------------------------------------------
// appendGroup(f) = f | "(" { f ";" } ")" . // ";" is optional before ")"
func (p *parser) appendGroup(list []Decl, f func(*Group) Decl) []Decl {
	setIdsUsed := func(f Decl) {
		switch ft := f.(type) {
		case *ConstDecl:
			for _, name := range ft.NameList {
				if name.Value != "_" {
					p.currPkg.IdsUsed[name.Value] = true
				}
			}
		case *VarDecl:
			for _, name := range ft.NameList {
				if name.Value != "_" {
					p.currPkg.IdsUsed[name.Value] = true
				}
			}
		case *TypeDecl:
			if ft.Name.Value != "_" {
				p.currPkg.IdsUsed[ft.Name.Value] = true
			}
		}
	}

	if p.tok == _Lparen {
		g := new(Group)
		p.list(_Lparen, _Semi, _Rparen, func() bool {
			fg := f(g)
			setIdsUsed(fg)
			list = append(list, fg)
			return false
		})

	} else {
		fnil := f(nil)
		setIdsUsed(fnil)
		list = append(list, fnil)
	}

	if debug {
		for _, d := range list {
			if d == nil {
				panic("nil list entry")
			}
		}
	}

	return list
}

//--------------------------------------------------------------------------------
// ImportSpec = [ "." | PackageName ] ImportPath .
// ImportPath = string_lit .
func (p *parser) importDecl(group *Group) Decl {
	if trace {
		defer p.trace("importDecl")("")
	}

	d := new(ImportDecl)
	d.pos = p.pos()
	if p.lineDirects {
		d.linedirect = d.pos
	}

	switch p.tok {
	case _Name:
		d.LocalPkgName = p.name()
		//only explicit local names of imports can be analyzed at the syntactic phase...
		if d.LocalPkgName.Value != "_" {
			p.currPkg.IdsUsed[d.LocalPkgName.Value] = true
		}
	case _Dot:
		d.LocalPkgName = p.newName(".")
		p.next()
	}
	d.Path = p.oliteral()
	if d.Path == nil {
		p.syntaxError("missing import path")
		p.advance(_Semi, _Rparen)
		return nil
	}
	if p.sectIsTc {
		_, pkgTag := filepath.Split(strings.Trim(d.Path.Value, "\""))
		if d.LocalPkgName != nil && d.LocalPkgName.Value == p.pkgName || pkgTag == p.pkgName {
			p.pkgName += "_test"
		}
	}
	d.Group = group
	if p.tok == _Lparen {
		d.Args = []Expr{}
		if d.LocalPkgName == nil {
			p.syntaxError("imported parameterized packages need a local name")
			p.advance(_Rparen)
			return nil
		}
		p.list(_Lparen, _Comma, _Rparen, func() bool {
			arg := p.typeOrNil()
			d.Args = append(d.Args, arg)
			return false
		})

		numPassThrus := 0 //TODO: only need a bool
		for _, arg := range d.Args {
			if arg, ok := arg.(*Name); ok {
				for _, param := range p.currPkg.Params {
					if arg.Value == param.Value {
						numPassThrus++
					}
				}
			}
		}
		if numPassThrus == 0 {
			p.argImports = append(p.argImports, d)
		}
		if len(p.infImports) > 0 {
			d.Infers = p.infImports
			p.infImports = []*ImportDecl{}
		}
	}
	if p.dynamicMode && d.LocalPkgName == nil {
		_, a := filepath.Split(strings.Trim(d.Path.Value, "\""))
		d.LocalPkgName = &Name{Value: a}
		p.currPkg.IdsUsed[a] = true
	}
	return d
}

//--------------------------------------------------------------------------------
// ConstSpec = IdentifierList [ [ Type ] "=" ExpressionList ] .
func (p *parser) constDecl(group *Group) Decl {
	if trace {
		defer p.trace("constDecl")("")
	}

	d := new(ConstDecl)
	d.pos = p.pos()
	if p.lineDirects {
		d.linedirect = d.pos
	}
	if p.docComments != nil {
		d.comments = &Comments{Alone: []Comment{}, Above: []Comment{}, Left: []Comment{}, Right: []Comment{}, Below: []Comment{}}
		for _, c := range p.docComments {
			d.comments.Above = append(d.comments.Above, Comment{Text: c})
		}
		p.docComments = nil
	}
	d.NameList = p.nameList(p.name())
	if p.tok != _EOF && p.tok != _Semi && p.tok != _Rparen {
		d.Type = p.typeOrNil()
		if p.got(_Assign) {
			d.Values = p.exprList()
		}
	}
	d.Group = group

	return d
}

//--------------------------------------------------------------------------------
// TypeSpec = identifier [ "=" ] Type .
func (p *parser) typeDecl(group *Group) Decl {
	if trace {
		defer p.trace("typeDecl")("")
	}

	d := new(TypeDecl)
	d.pos = p.pos()
	if p.lineDirects {
		d.linedirect = d.pos
	}
	if p.docComments != nil {
		d.comments = &Comments{Alone: []Comment{}, Above: []Comment{}, Left: []Comment{}, Right: []Comment{}, Below: []Comment{}}
		for _, c := range p.docComments {
			d.comments.Above = append(d.comments.Above, Comment{Text: c})
		}
		p.docComments = nil
	}

	d.Name = p.name()
	d.Alias = p.got(_Assign)
	d.Type = p.typeOrNil()
	if d.Type == nil {
		d.Type = p.bad()
		p.syntaxError("in type declaration")
		p.advance(_Semi, _Rparen)
	}
	d.Group = group
	d.Pragma = p.pragma

	return d
}

//--------------------------------------------------------------------------------
// VarSpec = IdentifierList ( Type [ "=" ExpressionList ] | "=" ExpressionList ) .
func (p *parser) varDecl(group *Group) Decl {
	if trace {
		defer p.trace("varDecl")("")
	}

	d := new(VarDecl)
	d.pos = p.pos()
	if p.lineDirects {
		d.linedirect = d.pos
	}
	if p.docComments != nil {
		d.comments = &Comments{Alone: []Comment{}, Above: []Comment{}, Left: []Comment{}, Right: []Comment{}, Below: []Comment{}}
		for _, c := range p.docComments {
			d.comments.Above = append(d.comments.Above, Comment{Text: c})
		}
		p.docComments = nil
	}

	d.NameList = p.nameList(p.name())
	if p.got(_Assign) {
		d.Values = p.exprList()
	} else {
		d.Type = p.type_()
		if p.got(_Assign) {
			d.Values = p.exprList()
		}
	}
	d.Group = group

	return d
}

//--------------------------------------------------------------------------------
// FunctionDecl = "func" FunctionName ( Function | Signature ) .
// FunctionName = identifier .
// Function     = Signature FunctionBody .
// MethodDecl   = "func" Receiver MethodName ( Function | Signature ) .
// Receiver     = Parameters .
func (p *parser) funcDeclOrNil(stmt func() Stmt) *FuncDecl {
	if trace {
		defer p.trace("funcDecl")("")
	}

	f := new(FuncDecl)
	f.pos = p.pos()
	if p.lineDirects {
		f.linedirect = f.pos
	}
	if p.docComments != nil {
		f.comments = &Comments{Alone: []Comment{}, Above: []Comment{}, Left: []Comment{}, Right: []Comment{}, Below: []Comment{}}
		for _, c := range p.docComments {
			f.comments.Above = append(f.comments.Above, Comment{Text: c})
		}
		p.docComments = nil
	}

	if p.tok == _Lparen {
		rcvr := p.paramList()
		switch len(rcvr) {
		case 0:
			p.error("method has no receiver")
		default:
			p.error("method has multiple receivers")
			fallthrough
		case 1:
			f.Recv = rcvr[0]
		}
	}

	if p.tok != _Name {
		p.syntaxError("expecting name or (")
		p.advance(_Lbrace, _Semi)
		return nil
	}

	// TODO(gri) check for regular functions only
	// if name.Sym.Name == "init" {
	// 	name = renameinit()
	// 	if params != nil || result != nil {
	// 		p.error("func init must have no arguments and no return values")
	// 	}
	// }

	f.Name = p.name()
	f.Type = p.funcType()
	if p.pkgName == "main" && f.Name.Value == "main" {
		if len(f.Type.ParamList) != 0 || len(f.Type.ResultList) != 0 {
			p.error("func main must have no arguments and no return values")
		}
	}
	if f.Name.Value == "main" && len(f.Type.ParamList) == 0 && len(f.Type.ResultList) == 0 {
		p.sectHasMain = true
	}
	if p.tok == _Lbrace {
		f.Body = p.funcBody(stmt)
	}

	f.Pragma = p.pragma
	return f
}

//--------------------------------------------------------------------------------
// Statements
//--------------------------------------------------------------------------------
// Possible statement for init() =
// 	IfStmt | ForStmt | SwitchStmt | SelectStmt | DeferStmt | GoStmt | ReturnStmt | DoStmt | ToplevelBlock | StringLit.
func (p *parser) tlBlock() *FuncDecl {
	if trace {
		defer p.trace("toplevel standalone stmts")("")
	}

	p.sectHasStmts = true
	f := p.newBlankFunc("init")
	l := []Stmt{}
forloop:
	for {
		switch p.tok {
		//these keyword-based statements can be standalone
		case _If, _For, _Switch, _Select, _Go, _Var, _Const, _Type, _Lbrace, _Literal:
			l = append(l, p.tlStmt())
		default:
			if p.isName("do") || (p.tok == _Name && p.macroRegistry[p.lit] != nil) {
				if p.lineDirects {
					f.linedirect = f.pos
				}
				if len(p.comments) > 0 {
					f.comments = &Comments{Alone: []Comment{}, Above: []Comment{}, Left: []Comment{}, Right: []Comment{}, Below: []Comment{}}
					for i, c := range p.comments {
						if i >= len(p.comments)-p.numDocComments {
							f.comments.Above = append(f.comments.Above, Comment{Text: c})
						}
					}
				}
				l = append(l, p.tlStmt())
			} else {
				break forloop
			}
		}

		// ";" is optional before "}"
		if p.tok != _EOF && !p.got(_Semi) && p.tok != _Rbrace {
			p.syntaxError("at end of statement")
			p.advance(_Const, _Type, _Var, _Func, _If, _For, _Switch, _Select, _Go, _Semi, _Rbrace, _Case, _Default)
			p.got(_Semi) // avoid spurious empty statement
			return nil
		}
	}
	f.Body.Rbrace = p.pos()
	f.Body.List = l
	return f
}

//--------------------------------------------------------------------------------
// Possible statement for init() =
// 	IfStmt | ForStmt | SwitchStmt | SelectStmt | DeferStmt | GoStmt | ReturnStmt | DoStmt | ToplevelBlock | StringLit.
func (p *parser) tlStmtList(stmt func() Stmt) []Stmt {
	if trace {
		defer p.trace("seq of toplevel-style stmts")("")
	}

	l := []Stmt{}
forloop:
	for {
		switch p.tok {
		//these keyword-based statements can be standalone
		case _If, _For, _Switch, _Select, _Go, _Defer, _Var, _Const, _Type, _Lbrace, _Literal:
			l = append(l, stmt())
		default:
			if p.isName("do") || (p.tok == _Name && p.macroRegistry[p.lit] != nil) {
				l = append(l, stmt())
			} else {
				break forloop
			}
		}

		// ";" is optional before "}"
		if p.tok != _EOF && !p.got(_Semi) && p.tok != _Rbrace {
			p.syntaxError("at end of statement")
			p.advance(_Const, _Type, _Var, _Func, _If, _For, _Switch, _Select, _Go, _Semi, _Rbrace, _Case, _Default)
			p.got(_Semi) // avoid spurious empty statement
			return nil
		}
	}
	return l
}

//--------------------------------------------------------------------------------
func (p *parser) tlStmt() Stmt {
	if trace {
		defer p.trace("top-level standalone stmt")("")
	}

	switch p.tok {
	case _Literal:
		return p.simpleStmt(nil, false)

	case _Var:
		return p.declStmt(p.varDecl)

	case _Const:
		return p.declStmt(p.constDecl)

	case _Type:
		return p.declStmt(p.typeDecl)

	case _If:
		return p.ifStmt(p.procStmt)

	case _For:
		return p.forStmt(p.procStmt)

	case _Switch:
		return p.switchStmt(p.procStmt)

	case _Select:
		return p.selectStmt(p.procStmt)

	case _Go:
		return p.callStmt(p.procStmt)

	case _Lbrace:
		return p.blockStmt("", p.procStmt)

	case _Semi:
		s := new(EmptyStmt)
		s.pos = p.pos()
		return s

	default:
		if p.isName("do") {
			p.want(_Name)
			switch p.tok {
			//don't include labelled stmts, var, const, func, type,
			// defer, fallthrough, continue, break, goto, return within do-statements
			case _If, _For, _Switch, _Select, _Go:
				return p.stmtOrNil()
			case _Name:
				return p.simpleStmt(p.exprList(), false)
			case _Lbrace:
				return p.blockStmt("", p.stmtOrNil)
			case _Operator, _Star:
				switch p.op {
				case Add, Sub, Mul, And, Xor, Not:
					return p.simpleStmt(nil, false) // unary operators
				}
			case _Literal, _Func, _Lparen, // operands
				_Lbrack, _Struct, _Map, _Chan, _Interface, // composite types
				_Arrow: // receive operator
				return p.simpleStmt(nil, false)
			}
		}
		if mac := p.macroRegistry[p.lit]; p.tok == _Name && mac != nil {
			p.next()
			return mac(p, p.tlStmt) //p.tlStmt needed for "let"
		}
	}

	return nil //should never reach
}

//--------------------------------------------------------------------------------
func (p *parser) procStmt() Stmt {
	if trace {
		defer p.trace("standalone stmt within proc")("")
	}

	if p.isName("do") {
		p.want(_Name)
		switch p.tok {
		//don't include labelled stmts, var, const, func, type,
		// fallthrough, continue, break, goto, return within do-statements
		case _If, _For, _Switch, _Select, _Go, _Defer:
			return p.stmtOrNil()
		case _Name:
			return p.simpleStmt(p.exprList(), false)
		case _Lbrace:
			return p.blockStmt("", p.stmtOrNil)
		case _Operator, _Star:
			switch p.op {
			case Add, Sub, Mul, And, Xor, Not:
				return p.simpleStmt(nil, false) // unary operators
			}
		case _Literal, _Func, _Lparen, // operands
			_Lbrack, _Struct, _Map, _Chan, _Interface, // composite types
			_Arrow: // receive operator
			return p.simpleStmt(nil, false)
		}
	} else if mac := p.macroRegistry[p.lit]; p.tok == _Name && mac != nil {
		p.next()
		return mac(p, p.procStmt) //p.procStmt needed for "let"
	} else if p.tok == _Name { //TODO: check label as first if-option
		pos := p.pos()
		lhs := p.exprList()
		if label, ok := lhs.(*Name); ok && p.tok == _Colon {
			return p.labeledStmtOrNil(label)
		}
		p.syntaxErrorAt(pos, fmt.Sprintf("unexpected name, expecting \"do\", label or macro name"))
		return nil
	}
	switch p.tok {
	case _Literal:
		return p.simpleStmt(nil, false)

	case _Var:
		return p.declStmt(p.varDecl)

	case _Const:
		return p.declStmt(p.constDecl)

	case _Type:
		return p.declStmt(p.typeDecl)

	case _If:
		return p.ifStmt(p.procStmt)

	case _For:
		return p.forStmt(p.procStmt)

	case _Switch:
		return p.switchStmt(p.procStmt)

	case _Select:
		return p.selectStmt(p.procStmt)

	case _Go, _Defer:
		return p.callStmt(p.procStmt)

	case _Lbrace:
		return p.blockStmt("", p.procStmt)

	case _Semi:
		s := new(EmptyStmt)
		s.pos = p.pos()
		return s

	case _Fallthrough:
		s := new(BranchStmt)
		s.pos = p.pos()
		p.next()
		s.Tok = _Fallthrough
		return s

	case _Break, _Continue:
		s := new(BranchStmt)
		s.pos = p.pos()
		s.Tok = p.tok
		p.next()
		if p.tok == _Name {
			s.Label = p.name()
		}
		return s

	case _Goto:
		if !p.permits["goto"] {
			p.syntaxError("goto-statement has been prohibited by blacklist")
			p.advance(_Semi)
			return nil
		}
		s := new(BranchStmt)
		s.pos = p.pos()
		s.Tok = _Goto
		p.next()
		s.Label = p.name()
		return s

	case _Return:
		if !p.permits["return"] {
			p.syntaxError("return-statement has been prohibited by blacklist")
			p.advance(_Semi)
			return nil
		}
		s := new(ReturnStmt)
		s.pos = p.pos()
		p.next()
		if p.tok != _Semi && p.tok != _Rbrace {
			s.Results = p.exprList()
		}
		return s
	}

	return nil //should never reach
}

//--------------------------------------------------------------------------------
// StatementList = { Statement ";" } .
func (p *parser) stmtList(stmt func() Stmt) (l []Stmt) {
	if trace {
		defer p.trace("stmtList")("")
	}

	for p.tok != _EOF && p.tok != _Rbrace && p.tok != _Case && p.tok != _Default {
		s := stmt()
		if s == nil {
			break
		}
		l = append(l, s)

		// ";" is optional before "}"
		if !p.got(_Semi) && p.tok != _Rbrace {
			p.syntaxError("at end of statement")
			p.advance(_Semi, _Rbrace, _Case, _Default)
			p.got(_Semi) // avoid spurious empty statement
		}

	}
	return
}

//--------------------------------------------------------------------------------
// Statement =
// 	Declaration | LabeledStmt | SimpleStmt |
// 	GoStmt | ReturnStmt | BreakStmt | ContinueStmt | GotoStmt |
// 	FallthroughStmt | Block | IfStmt | SwitchStmt | SelectStmt | ForStmt |
// 	DeferStmt .
func (p *parser) stmtOrNil() Stmt {
	if trace {
		defer p.trace("stmt " + p.tok.String())("")
	}

	// Most statements (assignments) start with an identifier;
	// look for it first before doing anything more expensive.
	if p.tok == _Name {
		lhs := p.exprList()
		if label, ok := lhs.(*Name); ok && p.tok == _Colon {
			return p.labeledStmtOrNil(label)
		}
		return p.simpleStmt(lhs, false)
	}

	switch p.tok {
	case _Lbrace:
		return p.blockStmt("", p.stmtOrNil)

	case _Var:
		return p.declStmt(p.varDecl)

	case _Const:
		return p.declStmt(p.constDecl)

	case _Type:
		return p.declStmt(p.typeDecl)

	case _Operator, _Star:
		switch p.op {
		case Add, Sub, Mul, And, Xor, Not:
			return p.simpleStmt(nil, false) // unary operators
		}

	case _Literal, _Func, _Lparen, // operands
		_Lbrack, _Struct, _Map, _Chan, _Interface, // composite types
		_Arrow: // receive operator
		return p.simpleStmt(nil, false)

	case _If:
		return p.ifStmt(p.stmtOrNil)

	case _For:
		return p.forStmt(p.stmtOrNil)

	case _Switch:
		return p.switchStmt(p.stmtOrNil)

	case _Select:
		return p.selectStmt(p.stmtOrNil)

	case _Fallthrough:
		s := new(BranchStmt)
		s.pos = p.pos()
		p.next()
		s.Tok = _Fallthrough
		return s

	case _Break, _Continue:
		s := new(BranchStmt)
		s.pos = p.pos()
		s.Tok = p.tok
		p.next()
		if p.tok == _Name {
			s.Label = p.name()
		}
		return s

	case _Go, _Defer:
		return p.callStmt(p.stmtOrNil)

	case _Goto:
		if !p.permits["goto"] {
			p.syntaxError("goto-statement has been prohibited by blacklist")
			p.advance(_Semi)
			return nil
		}
		s := new(BranchStmt)
		s.pos = p.pos()
		s.Tok = _Goto
		p.next()
		s.Label = p.name()
		return s

	case _Return:
		if !p.permits["return"] {
			p.syntaxError("return-statement has been prohibited by blacklist")
			p.advance(_Semi)
			return nil
		}
		s := new(ReturnStmt)
		s.pos = p.pos()
		p.next()
		if p.tok != _Semi && p.tok != _Rbrace {
			s.Results = p.exprList()
		}
		return s

	case _Semi:
		s := new(EmptyStmt)
		s.pos = p.pos()
		return s
	}

	return nil
}

//--------------------------------------------------------------------------------
// We represent x++, x-- as assignments x += ImplicitOne, x -= ImplicitOne.
// We use ImplicitOne so they'll be printed by printer.go as x++/-- instead of x +=/-= 1.
// ImplicitOne should not be used elsewhere.
var ImplicitOne = &BasicLit{Value: "1"}

//--------------------------------------------------------------------------------
// SimpleStmt = EmptyStmt | ExpressionStmt | SendStmt | IncDecStmt | Assignment | ShortVarDecl .
func (p *parser) simpleStmt(lhs Expr, rangeOk bool) SimpleStmt {
	if trace {
		defer p.trace("simpleStmt")("")
	}

	if rangeOk && p.tok == _Range {
		// _Range expr
		if debug && lhs != nil {
			panic("invalid call of simpleStmt")
		}
		return p.newRangeClause(nil, false)
	}

	if lhs == nil {
		lhs = p.exprList()
	}

	if _, ok := lhs.(*ListExpr); !ok && p.tok != _Assign && p.tok != _Define {
		// expr
		pos := p.pos()
		switch p.tok {
		case _AssignOp:
			// lhs op= rhs
			op := p.op
			p.next()
			return p.newAssignStmt(pos, op, lhs, p.expr())

		case _IncOp:
			// lhs++ or lhs--
			op := p.op
			p.next()
			return p.newAssignStmt(pos, op, lhs, ImplicitOne)

		case _Arrow:
			// lhs <- rhs
			s := new(SendStmt)
			s.pos = pos
			p.next()
			s.Chan = lhs
			s.Value = p.expr()
			return s

		default:
			// expr
			s := new(ExprStmt)
			s.pos = lhs.Pos()
			s.X = lhs
			return s
		}
	}

	// expr_list
	pos := p.pos()
	switch p.tok {
	case _Assign:
		p.next()

		if rangeOk && p.tok == _Range {
			// expr_list '=' _Range expr
			return p.newRangeClause(lhs, false)
		}

		// expr_list '=' expr_list
		return p.newAssignStmt(pos, 0, lhs, p.exprList())

	case _Define:
		p.next()

		if rangeOk && p.tok == _Range {
			// expr_list ':=' range expr
			return p.newRangeClause(lhs, true)
		}

		// expr_list ':=' expr_list
		rhs := p.exprList()

		if x, ok := rhs.(*TypeSwitchGuard); ok {
			switch lhs := lhs.(type) {
			case *Name:
				x.Lhs = lhs
			case *ListExpr:
				p.errorAt(lhs.Pos(), fmt.Sprintf("cannot assign 1 value to %d variables", len(lhs.ElemList)))
				// make the best of what we have
				if lhs, ok := lhs.ElemList[0].(*Name); ok {
					x.Lhs = lhs
				}
			default:
				p.errorAt(lhs.Pos(), fmt.Sprintf("invalid variable name %s in type switch", String(lhs)))
			}
			s := new(ExprStmt)
			s.pos = x.Pos()
			s.X = x
			return s
		}

		as := p.newAssignStmt(pos, Def, lhs, rhs)
		return as

	default:
		p.syntaxError("expecting := or = or comma")
		p.advance(_Semi, _Rbrace)
		// make the best of what we have
		if x, ok := lhs.(*ListExpr); ok {
			lhs = x.ElemList[0]
		}
		s := new(ExprStmt)
		s.pos = lhs.Pos()
		s.X = lhs
		return s
	}
}

//--------------------------------------------------------------------------------
func (p *parser) newRangeClause(lhs Expr, def bool) *RangeClause {
	r := new(RangeClause)
	r.pos = p.pos()
	p.next() // consume _Range
	r.Lhs = lhs
	r.Def = def
	r.X = p.expr()
	return r
}

//--------------------------------------------------------------------------------
func (p *parser) newAssignStmt(pos src.Pos, op Operator, lhs, rhs Expr) *AssignStmt {
	a := new(AssignStmt)
	a.pos = pos
	a.Op = op
	a.Lhs = lhs
	a.Rhs = rhs
	return a
}

//--------------------------------------------------------------------------------
func (p *parser) labeledStmtOrNil(label *Name) Stmt {
	if trace {
		defer p.trace("labeledStmt")("")
	}

	s := new(LabeledStmt)
	s.pos = p.pos()
	s.Label = label

	p.want(_Colon)

	if p.tok == _Rbrace {
		// We expect a statement (incl. an empty statement), which must be
		// terminated by a semicolon. Because semicolons may be omitted before
		// an _Rbrace, seeing an _Rbrace implies an empty statement.
		e := new(EmptyStmt)
		e.pos = p.pos()
		s.Stmt = e
		return s
	}

	s.Stmt = p.stmtOrNil()
	if s.Stmt != nil {
		return s
	}

	// report error at line of ':' token
	p.syntaxErrorAt(s.pos, "missing statement after label")
	// we are already at the end of the labeled statement - no need to advance
	return nil // avoids follow-on errors (see e.g., fixedbugs/bug274.go)
}

//--------------------------------------------------------------------------------
// context must be a non-empty string unless we know that p.tok == _Lbrace.
func (p *parser) blockStmt(context string, stmt func() Stmt) *BlockStmt {
	if trace {
		defer p.trace("blockStmt")("")
	}

	s := new(BlockStmt)
	s.pos = p.pos()

	// people coming from C may forget that braces are mandatory in Go
	if !p.got(_Lbrace) {
		p.syntaxError("expecting { after " + context)
		p.advance(_Name, _Rbrace)
		s.Rbrace = p.pos() // in case we found "}"
		if p.got(_Rbrace) {
			return s
		}
	}

	s.List = p.stmtList(stmt)
	s.Rbrace = p.pos()
	p.want(_Rbrace)

	return s
}

//--------------------------------------------------------------------------------
func (p *parser) declStmt(f func(*Group) Decl) *DeclStmt {
	if trace {
		defer p.trace("declStmt")("")
	}

	s := new(DeclStmt)
	s.pos = p.pos()

	p.next() // _Const, _Type, or _Var
	s.DeclList = p.appendGroup(nil, f)

	return s
}

//--------------------------------------------------------------------------------
// callStmt parses call-like statements that can be preceded by 'defer' and 'go'.
func (p *parser) callStmt(stmt func() Stmt) *CallStmt {
	if trace {
		defer p.trace("callStmt")("")
	}

	s := new(CallStmt)
	s.pos = p.pos()
	s.Tok = p.tok // _Defer or _Go
	p.next()
	var x Expr

	if p.tok == _Lbrace {
		t := new(FuncType)
		t.pos = p.pos()
		t.ParamList = []*Field{}
		t.ResultList = nil
		f := new(FuncLit)
		f.pos = p.pos()
		f.Type = t
		f.Body = p.blockStmt("", stmt)
		e := new(CallExpr)
		e.pos = p.pos()
		e.ArgList = []Expr{}
		e.Fun = f
		x = e
	} else {
		x = p.pexpr(p.tok == _Lparen) // keep_parens so we can report error below
		if t := unparen(x); t != x {
			p.error(fmt.Sprintf("expression in %s must not be parenthesized", s.Tok))
			// already progressed, no need to advance
			x = t
		}
	}
	cx, ok := x.(*CallExpr)
	if !ok {
		p.error(fmt.Sprintf("expression in %s must be function call", s.Tok))
		// already progressed, no need to advance
		cx = new(CallExpr)
		cx.pos = x.Pos()
		cx.Fun = p.bad()
	}
	s.Call = cx
	return s
}

//--------------------------------------------------------------------------------
func (p *parser) ifStmt(stmt func() Stmt) *IfStmt {
	if trace {
		defer p.trace("ifStmt")("")
	}
	if !p.permits["if"] {
		p.syntaxError("if-statement has been prohibited by blacklist")
		p.advance(_Rbrace)
		return nil
	}
	s := new(IfStmt)
	s.pos = p.pos()

	s.Init, s.Cond, _ = p.header(_If)
	s.Then = p.blockStmt("if clause", stmt)

	if p.got(_Else) {
		switch p.tok {
		case _If:
			s.Else = p.ifStmt(stmt)
		case _Switch:
			body := new(BlockStmt)
			body.pos = p.pos()
			body.List = []Stmt{p.switchStmt(stmt)}
			body.Rbrace = p.pos()
			s.Else = body
		case _For:
			body := new(BlockStmt)
			body.pos = p.pos()
			body.List = []Stmt{p.forStmt(stmt)}
			body.Rbrace = p.pos()
			s.Else = body
		case _Select:
			body := new(BlockStmt)
			body.pos = p.pos()
			body.List = []Stmt{p.selectStmt(stmt)}
			body.Rbrace = p.pos()
			s.Else = body
		case _Go, _Defer:
			body := new(BlockStmt)
			body.pos = p.pos()
			body.List = []Stmt{p.callStmt(stmt)}
			body.Rbrace = p.pos()
			s.Else = body
		case _Lbrace:
			s.Else = p.blockStmt("", stmt)
		default:
			p.syntaxError("else must be followed by if,switch,for,select,go,defer or statement block")
			p.advance(_Name, _Rbrace)
		}
	}

	return s
}

//--------------------------------------------------------------------------------
func (p *parser) forStmt(stmt func() Stmt) Stmt {
	if trace {
		defer p.trace("forStmt")("")
	}
	if !p.permits["for"] {
		p.syntaxError("for-statement has been prohibited by blacklist")
		p.advance(_Rbrace)
		return nil
	}

	s := new(ForStmt)
	s.pos = p.pos()

	s.Init, s.Cond, s.Post = p.header(_For)
	s.Body = p.blockStmt("for clause", stmt)

	return s
}

//--------------------------------------------------------------------------------
func (p *parser) header(keyword token) (init SimpleStmt, cond Expr, post SimpleStmt) {
	p.want(keyword)

	if p.tok == _Lbrace {
		if keyword == _If {
			p.syntaxError("missing condition in if statement")
		}
		return
	}
	// p.tok != _Lbrace

	outer := p.xnest
	p.xnest = -1

	if p.tok != _Semi {
		// accept potential varDecl but complain
		if p.got(_Var) {
			p.syntaxError(fmt.Sprintf("var declaration not allowed in %s initializer", keyword.String()))
		}
		init = p.simpleStmt(nil, keyword == _For)
		// If we have a range clause, we are done (can only happen for keyword == _For).
		if _, ok := init.(*RangeClause); ok {
			p.xnest = outer
			return
		}
	}

	var condStmt SimpleStmt
	var semi struct {
		pos src.Pos
		lit string // valid if pos.IsKnown()
	}
	if p.tok == _Semi {
		semi.pos = p.pos()
		semi.lit = p.lit
		p.next()
		if keyword == _For {
			if p.tok != _Semi {
				if p.tok == _Lbrace {
					p.syntaxError("expecting for loop condition")
					goto done
				}
				condStmt = p.simpleStmt(nil, false)
			}
			p.want(_Semi)
			if p.tok != _Lbrace {
				post = p.simpleStmt(nil, false)
				if a, _ := post.(*AssignStmt); a != nil && a.Op == Def {
					p.syntaxErrorAt(a.Pos(), "cannot declare in post statement of for loop")
				}
			}
		} else if p.tok != _Lbrace {
			condStmt = p.simpleStmt(nil, false)
		}
	} else {
		condStmt = init
		init = nil
	}

done:
	// unpack condStmt
	switch s := condStmt.(type) {
	case nil:
		if keyword == _If && semi.pos.IsKnown() {
			if semi.lit != "semicolon" {
				p.syntaxErrorAt(semi.pos, fmt.Sprintf("unexpected %s, expecting { after if clause", semi.lit))
			} else {
				p.syntaxErrorAt(semi.pos, "missing condition in if statement")
			}
		}
	case *ExprStmt:
		cond = s.X
	default:
		p.syntaxError(fmt.Sprintf("%s used as value", String(s)))
	}

	p.xnest = outer
	return
}

//--------------------------------------------------------------------------------
func (p *parser) switchStmt(stmt func() Stmt) *SwitchStmt {
	if trace {
		defer p.trace("switchStmt")("")
	}
	if !p.permits["switch"] {
		p.syntaxError("switch-statement has been prohibited by blacklist")
		p.advance(_Rbrace)
		return nil
	}

	s := new(SwitchStmt)
	s.pos = p.pos()

	s.Init, s.Tag, _ = p.header(_Switch)

	if !p.got(_Lbrace) {
		p.syntaxError("missing { after switch clause")
		p.advance(_Case, _Default, _Rbrace)
	}
	for p.tok != _EOF && p.tok != _Rbrace {
		s.Body = append(s.Body, p.caseClause(stmt))
	}
	s.Rbrace = p.pos()
	p.want(_Rbrace)

	return s
}

//--------------------------------------------------------------------------------
func (p *parser) caseClause(stmt func() Stmt) *CaseClause {
	if trace {
		defer p.trace("caseClause")("")
	}

	c := new(CaseClause)
	c.pos = p.pos()

	switch p.tok {
	case _Case:
		p.next()
		c.Cases = p.exprList()

	case _Default:
		p.next()

	default:
		p.syntaxError("expecting case or default or }")
		p.advance(_Colon, _Case, _Default, _Rbrace)
	}

	c.Colon = p.pos()
	p.want(_Colon)
	c.Body = p.stmtList(stmt)

	return c
}

//--------------------------------------------------------------------------------
func (p *parser) selectStmt(stmt func() Stmt) *SelectStmt {
	if trace {
		defer p.trace("selectStmt")("")
	}
	if !p.permits["select"] {
		p.syntaxError("select-statement has been prohibited by blacklist")
		p.advance(_Rbrace)
		return nil
	}

	s := new(SelectStmt)
	s.pos = p.pos()

	p.want(_Select)
	if !p.got(_Lbrace) {
		p.syntaxError("missing { after select clause")
		p.advance(_Case, _Default, _Rbrace)
	}
	for p.tok != _EOF && p.tok != _Rbrace {
		s.Body = append(s.Body, p.commClause(stmt))
	}
	s.Rbrace = p.pos()
	p.want(_Rbrace)

	return s
}

//--------------------------------------------------------------------------------
func (p *parser) commClause(stmt func() Stmt) *CommClause {
	if trace {
		defer p.trace("commClause")("")
	}

	c := new(CommClause)
	c.pos = p.pos()

	switch p.tok {
	case _Case:
		p.next()
		c.Comm = p.simpleStmt(nil, false)

		// The syntax restricts the possible simple statements here to:
		//
		//     lhs <- x (send statement)
		//     <-x
		//     lhs = <-x
		//     lhs := <-x
		//
		// All these (and more) are recognized by simpleStmt and invalid
		// syntax trees are flagged later, during type checking.
		// TODO(gri) eventually may want to restrict valid syntax trees
		// here.

	case _Default:
		p.next()

	default:
		p.syntaxError("expecting case or default or }")
		p.advance(_Colon, _Case, _Default, _Rbrace)
	}

	c.Colon = p.pos()
	p.want(_Colon)
	c.Body = p.stmtList(stmt)

	return c
}

//--------------------------------------------------------------------------------
// Expressions
//--------------------------------------------------------------------------------
func (p *parser) expr() Expr {
	if trace {
		defer p.trace("expr")("")
	}

	return p.binaryExpr(0)
}

//--------------------------------------------------------------------------------
// Arguments = "(" [ ( ExpressionList | Type [ "," ExpressionList ] ) [ "..." ] [ "," ] ] ")" .
//func (p *parser) call(fun Expr) *CallExpr {
func (p *parser) argList() (list []Expr, hasDots bool) {
	if trace {
		defer p.trace("argList")("")
	}

	p.xnest++
	p.list(_Lparen, _Comma, _Rparen, func() bool {
		list = append(list, p.expr())
		hasDots = p.got(_DotDotDot)
		return hasDots
	})
	p.xnest--
	return
}

//--------------------------------------------------------------------------------
// Expression = UnaryExpr | Expression binary_op Expression .
func (p *parser) binaryExpr(prec int) Expr {
	// don't trace binaryExpr - only leads to overly nested trace output
	x := p.unaryExpr()
	for (p.tok == _Operator || p.tok == _Star) && p.prec > prec {
		op := p.op
		binOp := dynBinOps[op]
		if p.dynamicBlock == "" || binOp == "" {
			t := new(Operation)
			t.pos = p.pos()
			t.Op = p.op
			t.X = x
			tprec := p.prec
			p.next()
			t.Y = p.binaryExpr(tprec)
			x = t
		} else {
			t := &CallExpr{
				Fun: &SelectorExpr{
					X:   &Name{Value: p.dynamicBlock},
					Sel: &Name{Value: binOp},
				},
				ArgList: []Expr{x},
			}
			t.pos = p.pos()
			_ = p.procImportAlias(&BasicLit{Value: dynLib, Kind: StringLit}, p.dynamicBlock)
			tprec := p.prec
			p.next()
			expr := p.binaryExpr(tprec)
			if op == AndAnd || op == OrOr {
				expr = &FuncLit{
					Type: &FuncType{
						ParamList:  []*Field{},
						ResultList: []*Field{&Field{Type: &InterfaceType{MethodList: []*Field{}}}},
					},
					Body: &BlockStmt{
						List: []Stmt{&ReturnStmt{Results: expr}},
					},
				}
			}
			t.ArgList = append(t.ArgList, expr)
			x = t
		}
	}
	return x
}

var dynBinOps = map[Operator]string{
	Add:    "Plus",             // +
	Sub:    "Minus",            // -
	Or:     "Alt",              // |
	Xor:    "Xor",              // ^
	Mul:    "Mult",             // *
	Div:    "Divide",           // /
	Rem:    "Mod",              // %
	And:    "Seq",              // &
	AndNot: "SeqXor",           // &^
	Shl:    "LeftShift",        // <<
	Shr:    "RightShift",       // >>
	Eql:    "IsEqual",          // ==
	Neq:    "IsNotEqual",       // !=
	Lss:    "IsLessThan",       // <
	Leq:    "IsLessOrEqual",    // <=
	Gtr:    "IsGreaterThan",    // >
	Geq:    "IsGreaterOrEqual", // >=
	AndAnd: "And",              // &&
	OrOr:   "Or",               // ||
}

var dynUnaryOps = map[Operator]string{
	Add: "Identity", // +
	Sub: "Negate",   // -
	Not: "Not",      // !
	//Xor: "", // ^
}

//TODO: "Matcher", "Finder", "FindAll", "FindFirst", "RegexRepeat", "Group", "Parenthesize"

//--------------------------------------------------------------------------------
// UnaryExpr = PrimaryExpr | unary_op UnaryExpr .
func (p *parser) unaryExpr() Expr {
	if trace {
		defer p.trace("unaryExpr")("")
	}

	switch p.tok {
	case _Operator, _Star:
		switch p.op {
		case Mul, Add, Sub, Not, Xor:
			unaryOp := dynUnaryOps[p.op]
			if p.dynamicBlock == "" || unaryOp == "" {
				x := new(Operation)
				x.pos = p.pos()
				x.Op = p.op
				p.next()
				x.X = p.unaryExpr()
				return x
			} else {
				t := &CallExpr{
					Fun: &SelectorExpr{
						X:   &Name{Value: p.dynamicBlock},
						Sel: &Name{Value: unaryOp},
					},
					ArgList: []Expr{},
				}
				t.pos = p.pos()
				p.next()
				t.ArgList = append(t.ArgList, p.unaryExpr())
				_ = p.procImportAlias(&BasicLit{Value: dynLib, Kind: StringLit}, p.dynamicBlock)
				return t
			}

		case And:
			x := new(Operation)
			x.pos = p.pos()
			x.Op = And
			p.next()
			// unaryExpr may have returned a parenthesized composite literal
			// (see comment in operand) - remove parentheses if any
			x.X = unparen(p.unaryExpr())
			return x
		}

	case _Arrow:
		// receive op (<-x) or receive-only channel (<-chan E)
		pos := p.pos()
		p.next()

		// If the next token is _Chan we still don't know if it is
		// a channel (<-chan int) or a receive op (<-chan int(ch)).
		// We only know once we have found the end of the unaryExpr.

		x := p.unaryExpr()

		// There are two cases:
		//
		//   <-chan...  => <-x is a channel type
		//   <-x        => <-x is a receive operation
		//
		// In the first case, <- must be re-associated with
		// the channel type parsed already:
		//
		//   <-(chan E)   =>  (<-chan E)
		//   <-(chan<-E)  =>  (<-chan (<-E))

		if _, ok := x.(*ChanType); ok {
			// x is a channel type => re-associate <-
			dir := SendOnly
			t := x
			for dir == SendOnly {
				c, ok := t.(*ChanType)
				if !ok {
					break
				}
				dir = c.Dir
				if dir == RecvOnly {
					// t is type <-chan E but <-<-chan E is not permitted
					// (report same error as for "type _ <-<-chan E")
					p.syntaxError("unexpected <-, expecting chan")
					// already progressed, no need to advance
				}
				c.Dir = RecvOnly
				t = c.Elem
			}
			if dir == SendOnly {
				// channel dir is <- but channel element E is not a channel
				// (report same error as for "type _ <-chan<-E")
				p.syntaxError(fmt.Sprintf("unexpected %s, expecting chan", String(t)))
				// already progressed, no need to advance
			}
			return x
		}

		// x is not a channel type => we have a receive op
		o := new(Operation)
		o.pos = pos
		o.Op = Recv
		o.X = x
		return o
	}

	// TODO(mdempsky): We need parens here so we can report an
	// error for "(x) := true". It should be possible to detect
	// and reject that more efficiently though.
	return p.pexpr(true)
}

//--------------------------------------------------------------------------------
// Operand     = Literal | OperandName | MethodExpr | "(" Expression ")" .
// Literal     = BasicLit | CompositeLit | FunctionLit .
// BasicLit    = int_lit | float_lit | imaginary_lit | rune_lit | string_lit .
// OperandName = identifier | QualifiedIdent.
func (p *parser) operand(keep_parens bool) Expr {
	if trace {
		defer p.trace("operand " + p.tok.String())("")
	}

	switch p.tok {
	case _Name:
		return p.name()
	case _Literal:
		lit := p.oliteral()
		if lit.Kind == StringLit && p.tok == _Dot {
			if !p.permits["inplaceImps"] {
				p.syntaxError("inplace-imports disabled but are present")
				return nil
			}
			a := p.procImportAlias(lit, "")
			lit.Value = a
			return lit
		}
		if p.dynamicBlock != "" {
			switch lit.Kind {
			case StringLit:
				t := &CallExpr{
					Fun: &SelectorExpr{
						X:   &Name{Value: "utf88"},
						Sel: &Name{Value: "Desur"},
					},
					ArgList: []Expr{lit},
				}
				t.pos = p.pos()
				_ = p.procImportAlias(&BasicLit{Value: utf88Lib, Kind: StringLit}, "utf88")
				return t
			case RuneLit:
				t := &CallExpr{
					Fun: &SelectorExpr{
						X:   &Name{Value: p.dynamicBlock},
						Sel: &Name{Value: "Runex"},
					},
					ArgList: []Expr{&BasicLit{Value: strconv.Quote(strings.Trim(lit.Value, "'")), Kind: StringLit}},
				}
				t.pos = p.pos()
				_ = p.procImportAlias(&BasicLit{Value: dynLib, Kind: StringLit}, p.dynamicBlock)
				return t
			}
		}
		return lit

	case _Lparen:
		pos := p.pos()
		p.next()
		p.xnest++
		x := p.expr()
		p.xnest--
		p.want(_Rparen)

		// Optimization: Record presence of ()'s only where needed
		// for error reporting. Don't bother in other cases; it is
		// just a waste of memory and time.

		// Parentheses are not permitted on lhs of := .
		// switch x.Op {
		// case ONAME, ONONAME, OPACK, OTYPE, OLITERAL, OTYPESW:
		// 	keep_parens = true
		// }

		// Parentheses are not permitted around T in a composite
		// literal T{}. If the next token is a {, assume x is a
		// composite literal type T (it may not be, { could be
		// the opening brace of a block, but we don't know yet).
		if p.tok == _Lbrace {
			keep_parens = true
		}

		// Parentheses are also not permitted around the expression
		// in a go/defer statement. In that case, operand is called
		// with keep_parens set.
		if keep_parens {
			px := new(ParenExpr)
			px.pos = pos
			px.X = x
			x = px
		}
		return x

	case _Func:
		pos := p.pos()
		p.next()
		t := p.funcType()
		if p.tok == _Lbrace {
			p.xnest++

			f := new(FuncLit)
			f.pos = pos
			f.Type = t
			/*f.Body = p.blockStmt("", p.stmtOrNil)
			if p.mode&CheckBranches != 0 {
				checkBranches(f.Body, p.errh)
			}*/
			f.Body = p.funcBody(p.stmtOrNil)

			p.xnest--
			return f
		}
		return t

	case _Lbrack, _Chan, _Map, _Struct, _Interface:
		return p.type_() // othertype

	default:
		x := p.bad()
		p.syntaxError("expecting expression")
		p.advance()
		return x
	}

	// Syntactically, composite literals are operands. Because a complit
	// type may be a qualified identifier which is handled by pexpr
	// (together with selector expressions), complits are parsed there
	// as well (operand is only called from pexpr).
}

//--------------------------------------------------------------------------------
// PrimaryExpr =
// 	Operand |
// 	Conversion |
// 	PrimaryExpr Selector |
// 	PrimaryExpr Index |
// 	PrimaryExpr Slice |
// 	PrimaryExpr TypeAssertion |
// 	PrimaryExpr Arguments .
//
// Selector       = "." identifier .
// Index          = "[" Expression "]" .
// Slice          = "[" ( [ Expression ] ":" [ Expression ] ) |
//                      ( [ Expression ] ":" Expression ":" Expression )
//                  "]" .
// TypeAssertion  = "." "(" Type ")" .
// Arguments      = "(" [ ( ExpressionList | Type [ "," ExpressionList ] ) [ "..." ] [ "," ] ] ")" .
func (p *parser) pexpr(keep_parens bool) Expr {
	if trace {
		defer p.trace("pexpr")("")
	}

	x := p.operand(keep_parens)

loop:
	for {
		pos := p.pos()
		switch p.tok {
		case _Dot:
			p.next()
			switch p.tok {
			case _Name:
				// pexpr '.' sym
				t := new(SelectorExpr)
				t.pos = pos
				t.X = x
				t.Sel = p.name()
				x = t

			case _Lparen:
				p.next()
				if p.got(_Type) {
					t := new(TypeSwitchGuard)
					t.pos = pos
					t.X = x
					x = t
				} else {
					t := new(AssertExpr)
					t.pos = pos
					t.X = x
					t.Type = p.expr()
					x = t
				}
				p.want(_Rparen)

			default:
				p.syntaxError("expecting name or (")
				p.advance(_Semi, _Rparen)
			}

		case _Lbrack:
			p.next()
			p.xnest++

			var i Expr
			if p.tok != _Colon {
				i = p.expr()
				if p.got(_Rbrack) {
					// x[i]
					t := new(IndexExpr)
					t.pos = pos
					t.X = x
					t.Index = i
					x = t
					p.xnest--
					break
				}
			}

			// x[i:...
			t := new(SliceExpr)
			t.pos = pos
			t.X = x
			t.Index[0] = i
			p.want(_Colon)
			if p.tok != _Colon && p.tok != _Rbrack {
				// x[i:j...
				t.Index[1] = p.expr()
			}
			if p.got(_Colon) {
				t.Full = true
				// x[i:j:...]
				if t.Index[1] == nil {
					p.error("middle index required in 3-index slice")
				}
				if p.tok != _Rbrack {
					// x[i:j:k...
					t.Index[2] = p.expr()
				} else {
					p.error("final index required in 3-index slice")
				}
			}
			p.want(_Rbrack)

			x = t
			p.xnest--

		case _Lparen:
			//x = p.call(x)
			t := new(CallExpr)
			t.pos = pos
			t.Fun = x
			t.ArgList, t.HasDots = p.argList()
			x = t

		case _Lbrace:
			// operand may have returned a parenthesized complit
			// type; accept it but complain if we have a complit
			t := unparen(x)
			// determine if '{' belongs to a composite literal or a block statement
			complit_ok := false
			switch t.(type) {
			case *Name, *SelectorExpr:
				if p.xnest >= 0 {
					// x is considered a composite literal type
					complit_ok = true
				}
			case *ArrayType, *SliceType, *StructType, *MapType:
				// x is a comptype
				complit_ok = true
			}
			if !complit_ok {
				break loop
			}
			if t != x {
				p.syntaxError("cannot parenthesize type in composite literal")
				// already progressed, no need to advance
			}
			n := p.complitexpr()
			n.Type = x
			x = n

		default:
			break loop
		}
	}

	return x
}

//--------------------------------------------------------------------------------
// Element = Expression | LiteralValue .
func (p *parser) bare_complitexpr() Expr {
	if trace {
		defer p.trace("bare_complitexpr")("")
	}

	if p.tok == _Lbrace {
		// '{' start_complit braced_keyval_list '}'
		return p.complitexpr()
	}

	return p.expr()
}

//--------------------------------------------------------------------------------
// LiteralValue = "{" [ ElementList [ "," ] ] "}" .
func (p *parser) complitexpr() *CompositeLit {
	if trace {
		defer p.trace("complitexpr")("")
	}

	x := new(CompositeLit)
	x.pos = p.pos()
	p.xnest++
	x.Rbrace = p.list(_Lbrace, _Comma, _Rbrace, func() bool {
		// value
		e := p.bare_complitexpr()
		if p.tok == _Colon {
			// key ':' value
			l := new(KeyValueExpr)
			l.pos = p.pos()
			p.next()
			l.Key = e
			l.Value = p.bare_complitexpr()
			e = l
			x.NKeys++
		}
		x.ElemList = append(x.ElemList, e)
		return false
	})
	p.xnest--
	return x
}

//--------------------------------------------------------------------------------
func (p *parser) bad() *BadExpr {
	b := new(BadExpr)
	b.pos = p.pos()
	return b
}

//--------------------------------------------------------------------------------
// Types
//--------------------------------------------------------------------------------
func (p *parser) type_() Expr {
	if trace {
		defer p.trace("type_")("")
	}

	typ := p.typeOrNil()
	if typ == nil {
		typ = p.bad()
		p.syntaxError("expecting type")
		p.advance()
	}

	return typ
}

//--------------------------------------------------------------------------------
// typeOrNil is like type_ but it returns nil if there was no type
// instead of reporting an error.
//
// Type     = TypeName | TypeLit | "(" Type ")" .
// TypeName = identifier | QualifiedIdent .
// TypeLit  = ArrayType | StructType | PointerType | FunctionType | InterfaceType |
// 	      SliceType | MapType | Channel_Type .
func (p *parser) typeOrNil() Expr {
	if trace {
		defer p.trace("typeOrNil")("")
	}

	pos := p.pos()
	switch p.tok {
	case _Star:
		// ptrtype
		p.next()
		return newIndirect(pos, p.type_())

	case _Arrow:
		// recvchantype
		p.next()
		p.want(_Chan)
		t := new(ChanType)
		t.pos = pos
		t.Dir = RecvOnly
		t.Elem = p.chanElem()
		return t

	case _Func:
		// fntype
		p.next()
		return p.funcType()

	case _Lbrack:
		// '[' oexpr ']' ntype
		// '[' _DotDotDot ']' ntype
		p.next()
		p.xnest++
		if p.got(_Rbrack) {
			// []T
			p.xnest--
			t := new(SliceType)
			t.pos = pos
			t.Elem = p.type_()
			return t
		}

		// [n]T
		t := new(ArrayType)
		t.pos = pos
		if !p.got(_DotDotDot) {
			t.Len = p.expr()
		}
		p.want(_Rbrack)
		p.xnest--
		t.Elem = p.type_()
		return t

	case _Chan:
		// _Chan non_recvchantype
		// _Chan _Comm ntype
		p.next()
		t := new(ChanType)
		t.pos = pos
		if p.got(_Arrow) {
			t.Dir = SendOnly
		}
		t.Elem = p.chanElem()
		return t

	case _Map:
		// _Map '[' ntype ']' ntype
		p.next()
		p.want(_Lbrack)
		t := new(MapType)
		t.pos = pos
		t.Key = p.type_()
		p.want(_Rbrack)
		t.Value = p.type_()
		return t

	case _Struct:
		return p.structType()

	case _Interface:
		return p.interfaceType()

	case _Name:
		return p.dotname(p.name())

	case _Lparen:
		p.next()
		t := p.type_()
		p.want(_Rparen)
		return t

	case _Literal:
		lit := p.oliteral()
		if lit.Kind == StringLit {
			a := p.procImportAlias(lit, "")
			return p.dotname(&Name{Value: a})
		} else {
			return nil
		}
	}

	return nil
}

//--------------------------------------------------------------------------------
func (p *parser) funcType() *FuncType {
	if trace {
		defer p.trace("funcType")("")
	}

	typ := new(FuncType)
	typ.pos = p.pos()
	typ.ParamList = p.paramList()
	typ.ResultList = p.funcResult()

	return typ
}

//--------------------------------------------------------------------------------
func (p *parser) chanElem() Expr {
	if trace {
		defer p.trace("chanElem")("")
	}

	typ := p.typeOrNil()
	if typ == nil {
		typ = p.bad()
		p.syntaxError("missing channel element type")
		// assume element type is simply absent - don't advance
	}

	return typ
}

//--------------------------------------------------------------------------------
func (p *parser) dotname(name *Name) Expr {
	if trace {
		defer p.trace("dotname")("")
	}

	if p.tok == _Dot {
		s := new(SelectorExpr)
		s.pos = p.pos()
		p.next()
		s.X = name
		s.Sel = p.name()
		return s
	}
	return name
}

//--------------------------------------------------------------------------------
// StructType = "struct" "{" { FieldDecl ";" } "}" .
func (p *parser) structType() *StructType {
	if trace {
		defer p.trace("structType")("")
	}

	typ := new(StructType)
	typ.pos = p.pos()
	p.want(_Struct)
	p.list(_Lbrace, _Semi, _Rbrace, func() bool {
		p.fieldDecl(typ)
		return false
	})
	return typ
}

//--------------------------------------------------------------------------------
// InterfaceType = "interface" "{" { MethodSpec ";" } "}" .
func (p *parser) interfaceType() *InterfaceType {
	if trace {
		defer p.trace("interfaceType")("")
	}

	typ := new(InterfaceType)
	typ.pos = p.pos()
	p.want(_Interface)
	p.list(_Lbrace, _Semi, _Rbrace, func() bool {
		if m := p.methodDecl(); m != nil {
			typ.MethodList = append(typ.MethodList, m)
		}
		return false
	})
	return typ
}

//--------------------------------------------------------------------------------
// FunctionBody = Block .
func (p *parser) funcBody(stmt func() Stmt) *BlockStmt {
	if trace {
		defer p.trace("funcBody")("")
	}

	p.fnest++
	errcnt := p.errcnt
	body := p.blockStmt("", stmt)
	p.fnest--
	// Don't check branches if there were syntax errors in the function
	// as it may lead to spurious errors (e.g., see test/switch2.go) or
	// possibly crashes due to incomplete syntax trees.
	if p.mode&CheckBranches != 0 && errcnt == p.errcnt {
		checkBranches(body, p.errh)
	}
	/*if body == nil {
		body = []Stmt{new(EmptyStmt)}
	}*/
	return body
}

//--------------------------------------------------------------------------------
// Result = Parameters | Type .
func (p *parser) funcResult() []*Field {
	if trace {
		defer p.trace("funcResult")("")
	}

	if p.tok == _Lparen {
		return p.paramList()
	}

	pos := p.pos()
	if typ := p.typeOrNil(); typ != nil {
		f := new(Field)
		f.pos = pos
		f.Type = typ
		return []*Field{f}
	}

	return nil
}

//--------------------------------------------------------------------------------
func (p *parser) addField(styp *StructType, pos src.Pos, name *Name, typ Expr, tag *BasicLit) {
	if tag != nil {
		for i := len(styp.FieldList) - len(styp.TagList); i > 0; i-- {
			styp.TagList = append(styp.TagList, nil)
		}
		styp.TagList = append(styp.TagList, tag)
	}

	f := new(Field)
	f.pos = pos
	f.Name = name
	f.Type = typ
	styp.FieldList = append(styp.FieldList, f)

	if debug && tag != nil && len(styp.FieldList) != len(styp.TagList) {
		panic("inconsistent struct field list")
	}
}

//--------------------------------------------------------------------------------
// FieldDecl      = (IdentifierList Type | AnonymousField) [ Tag ] .
// AnonymousField = [ "*" ] TypeName .
// Tag            = string_lit .
func (p *parser) fieldDecl(styp *StructType) {
	if trace {
		defer p.trace("fieldDecl")("")
	}

	pos := p.pos()
	switch p.tok {
	case _Name:
		name := p.name()
		if p.tok == _Dot || p.tok == _Literal || p.tok == _Semi || p.tok == _Rbrace {
			// embed oliteral
			typ := p.qualifiedName(name)
			tag := p.oliteral()
			p.addField(styp, pos, nil, typ, tag)
			return
		}

		// new_name_list ntype oliteral
		names := p.nameList(name)
		typ := p.type_()
		tag := p.oliteral()

		for _, name := range names {
			p.addField(styp, name.Pos(), name, typ, tag)
		}

	case _Lparen:
		p.next()
		if p.tok == _Star {
			// '(' '*' embed ')' oliteral
			pos := p.pos()
			p.next()
			typ := newIndirect(pos, p.qualifiedName(nil))
			p.want(_Rparen)
			tag := p.oliteral()
			p.addField(styp, pos, nil, typ, tag)
			p.syntaxError("cannot parenthesize embedded type")

		} else {
			// '(' embed ')' oliteral
			typ := p.qualifiedName(nil)
			p.want(_Rparen)
			tag := p.oliteral()
			p.addField(styp, pos, nil, typ, tag)
			p.syntaxError("cannot parenthesize embedded type")
		}

	case _Star:
		p.next()
		if p.got(_Lparen) {
			// '*' '(' embed ')' oliteral
			typ := newIndirect(pos, p.qualifiedName(nil))
			p.want(_Rparen)
			tag := p.oliteral()
			p.addField(styp, pos, nil, typ, tag)
			p.syntaxError("cannot parenthesize embedded type")

		} else {
			// '*' embed oliteral
			typ := newIndirect(pos, p.qualifiedName(nil))
			tag := p.oliteral()
			p.addField(styp, pos, nil, typ, tag)
		}

	default:
		p.syntaxError("expecting field name or embedded type")
		p.advance(_Semi, _Rbrace)
	}
}

//--------------------------------------------------------------------------------
// MethodSpec        = MethodName Signature | InterfaceTypeName .
// MethodName        = identifier .
// InterfaceTypeName = TypeName .
func (p *parser) methodDecl() *Field {
	if trace {
		defer p.trace("methodDecl")("")
	}

	switch p.tok {
	case _Name:
		name := p.name()

		// accept potential name list but complain
		hasNameList := false
		for p.got(_Comma) {
			p.name()
			hasNameList = true
		}
		if hasNameList {
			p.syntaxError("name list not allowed in interface type")
			// already progressed, no need to advance
		}

		f := new(Field)
		f.pos = name.Pos()
		if p.tok != _Lparen {
			// packname
			f.Type = p.qualifiedName(name)
			return f
		}

		f.Name = name
		f.Type = p.funcType()
		return f

	case _Lparen:
		p.syntaxError("cannot parenthesize embedded type")
		f := new(Field)
		f.pos = p.pos()
		p.next()
		f.Type = p.qualifiedName(nil)
		p.want(_Rparen)
		return f

	default:
		p.syntaxError("expecting method or interface name")
		p.advance(_Semi, _Rbrace)
		return nil
	}
}

//--------------------------------------------------------------------------------
// ParameterDecl = [ IdentifierList ] [ "..." ] Type .
func (p *parser) paramDeclOrNil() *Field {
	if trace {
		defer p.trace("paramDecl")("")
	}

	f := new(Field)
	f.pos = p.pos()

	switch p.tok {
	case _Name:
		f.Name = p.name()
		switch p.tok {
		case _Name, _Star, _Arrow, _Func, _Lbrack, _Chan, _Map, _Struct, _Interface, _Lparen:
			// sym name_or_type
			f.Type = p.type_()

		case _DotDotDot:
			// sym dotdotdot
			f.Type = p.dotsType()

		case _Dot:
			// name_or_type
			// from dotname
			f.Type = p.dotname(f.Name)
			f.Name = nil
		}

	case _Arrow, _Star, _Func, _Lbrack, _Chan, _Map, _Struct, _Interface, _Lparen:
		// name_or_type
		f.Type = p.type_()

	case _DotDotDot:
		// dotdotdot
		f.Type = p.dotsType()

	default:
		p.syntaxError("expecting )")
		p.advance(_Comma, _Rparen)
		return nil
	}

	return f
}

//--------------------------------------------------------------------------------
// ...Type
func (p *parser) dotsType() *DotsType {
	if trace {
		defer p.trace("dotsType")("")
	}

	t := new(DotsType)
	t.pos = p.pos()

	p.want(_DotDotDot)
	t.Elem = p.typeOrNil()
	if t.Elem == nil {
		t.Elem = p.bad()
		p.syntaxError("final argument in variadic function missing type")
	}

	return t
}

//--------------------------------------------------------------------------------
// Parameters    = "(" [ ParameterList [ "," ] ] ")" .
// ParameterList = ParameterDecl { "," ParameterDecl } .
func (p *parser) paramList() (list []*Field) {
	if trace {
		defer p.trace("paramList")("")
	}

	pos := p.pos()
	var named int // number of parameters that have an explicit name and type
	p.list(_Lparen, _Comma, _Rparen, func() bool {
		if par := p.paramDeclOrNil(); par != nil {
			if debug && par.Name == nil && par.Type == nil {
				panic("parameter without name or type")
			}
			if par.Name != nil && par.Type != nil {
				named++
			}
			list = append(list, par)
		}
		return false
	})

	// distribute parameter types
	if named == 0 {
		// all unnamed => found names are named types
		for _, par := range list {
			if typ := par.Name; typ != nil {
				par.Type = typ
				par.Name = nil
			}
		}
	} else if named != len(list) {
		// some named => all must be named
		ok := true
		var typ Expr
		for i := len(list) - 1; i >= 0; i-- {
			if par := list[i]; par.Type != nil {
				typ = par.Type
				if par.Name == nil {
					ok = false
					n := p.newName("_")
					n.pos = typ.Pos() // correct position
					par.Name = n
				}
			} else if typ != nil {
				par.Type = typ
			} else {
				// par.Type == nil && typ == nil => we only have a par.Name
				ok = false
				t := p.bad()
				t.pos = par.Name.Pos() // correct position
				par.Type = t
			}
		}
		if !ok {
			p.syntaxErrorAt(pos, "mixed named and unnamed function parameters")
		}
	}

	//p.want(_Rparen)
	return
}

//--------------------------------------------------------------------------------
// Common productions
//--------------------------------------------------------------------------------
func (p *parser) got(tok token) bool {
	if p.tok == tok {
		p.next()
		return true
	}
	return false
}

//--------------------------------------------------------------------------------
func (p *parser) want(tok token) {
	if !p.got(tok) {
		p.syntaxError("expecting " + tokstring(tok))
		p.advance()
	}
}

//--------------------------------------------------------------------------------
func (p *parser) newName(value string) *Name {
	n := new(Name)
	n.pos = p.pos()
	n.Value = value
	return n
}

//--------------------------------------------------------------------------------
func (p *parser) name() *Name {
	// no tracing to avoid overly verbose output

	if p.tok == _Name {
		n := p.newName(p.lit)
		p.next()
		return n
	}

	n := p.newName("_")
	p.syntaxError("expecting name")
	p.advance()
	return n
}

//--------------------------------------------------------------------------------
// IdentifierList = identifier { "," identifier } .
// The first name must be provided.
func (p *parser) nameList(first *Name) []*Name {
	if trace {
		defer p.trace("nameList")("")
	}

	if debug && first == nil {
		panic("first name not provided")
	}

	l := []*Name{first}
	for p.got(_Comma) {
		l = append(l, p.name())
	}

	return l
}

//--------------------------------------------------------------------------------
// The first name may be provided, or nil.
func (p *parser) qualifiedName(name *Name) Expr {
	if trace {
		defer p.trace("qualifiedName")("")
	}

	switch {
	case name != nil:
		// name is provided
	case p.tok == _Name:
		name = p.name()
	default:
		name = p.newName("_")
		p.syntaxError("expecting name")
		p.advance(_Dot, _Semi, _Rbrace)
	}

	return p.dotname(name)
}

//--------------------------------------------------------------------------------
// ExpressionList = Expression { "," Expression } .
func (p *parser) exprList() Expr {
	if trace {
		defer p.trace("exprList")("")
	}

	x := p.expr()
	if p.got(_Comma) {
		list := []Expr{x, p.expr()}
		for p.got(_Comma) {
			list = append(list, p.expr())
		}
		t := new(ListExpr)
		t.pos = x.Pos()
		t.ElemList = list
		x = t
	}
	return x
}

//--------------------------------------------------------------------------------
func (p *parser) oliteral() *BasicLit {
	if p.tok == _Literal {
		b := new(BasicLit)
		b.pos = p.pos()
		b.Value = p.lit
		b.Kind = p.kind
		p.next()
		return b
	}
	return nil
}

//--------------------------------------------------------------------------------
// Error handling
//--------------------------------------------------------------------------------
// posAt returns the Pos value for (line, col) and the current position base.
func (p *parser) posAt(line, col uint) src.Pos {
	return src.MakePos(p.base, line, col)
}

//--------------------------------------------------------------------------------
// error reports an error at the given position.
func (p *parser) errorAt(pos src.Pos, msg string) {
	err := Error{pos, msg}
	if p.first == nil {
		p.first = err
	}
	p.errcnt++
	if p.errh == nil {
		panic(p.first)
	}
	p.errh(err)
}

//--------------------------------------------------------------------------------
// syntax_error_at reports a syntax error at the given position.
func (p *parser) syntaxErrorAt(pos src.Pos, msg string) {
	if trace {
		defer p.trace("syntaxError (" + msg + ")")("")
		//p.print("syntax error: " + msg)
	}

	if p.tok == _EOF && p.first != nil {
		return // avoid meaningless follow-up errors
	}

	// add punctuation etc. as needed to msg
	switch {
	case msg == "":
		// nothing to do
	case strings.HasPrefix(msg, "in "), strings.HasPrefix(msg, "at "), strings.HasPrefix(msg, "after "):
		msg = " " + msg
	case strings.HasPrefix(msg, "expecting"):
		msg = ", " + msg
	default:
		// plain error - we don't care about current token
		p.errorAt(pos, "syntax error: "+msg)
		return
	}

	// determine token string
	var tok string
	switch p.tok {
	case _Name, _Semi:
		tok = p.lit
	case _Literal:
		tok = "literal " + p.lit
	case _Operator:
		tok = p.op.String()
	case _AssignOp:
		tok = p.op.String() + "="
	case _IncOp:
		tok = p.op.String()
		tok += tok
	default:
		tok = tokstring(p.tok)
	}

	p.errorAt(pos, "syntax error: unexpected "+tok+msg)
}

//--------------------------------------------------------------------------------
// Convenience methods using the current token position.
func (p *parser) pos() src.Pos           { return p.posAt(p.line, p.col) }
func (p *parser) error(msg string)       { p.errorAt(p.pos(), msg) }
func (p *parser) syntaxError(msg string) { p.syntaxErrorAt(p.pos(), msg) }

// The stopset contains keywords that start a statement.
// They are good synchronization points in case of syntax
// errors and (usually) shouldn't be skipped over.
const stopset uint64 = 1<<_Break |
	1<<_Const |
	1<<_Continue |
	1<<_Defer |
	1<<_Fallthrough |
	1<<_For |
	//1<<_Func |
	1<<_Go |
	1<<_Goto |
	1<<_If |
	1<<_Return |
	1<<_Select |
	1<<_Switch |
	1<<_Type |
	1<<_Var

//--------------------------------------------------------------------------------
// Advance consumes tokens until it finds a token of the stopset or followlist.
// The stopset is only considered if we are inside a function (p.fnest > 0).
// The followlist is the list of valid tokens that can follow a production;
// if it is empty, exactly one (non-EOF) token is consumed to ensure progress.
func (p *parser) advance(followlist ...token) {
	if trace {
		p.print(fmt.Sprintf("advance %s", followlist))
	}

	// compute follow set
	// (not speed critical, advance is only called in error situations)
	var followset uint64 = 1 << _EOF // don't skip over EOF
	if len(followlist) > 0 {
		if p.fnest > 0 {
			followset |= stopset
		}
		for _, tok := range followlist {
			followset |= 1 << tok
		}
	}

	for !contains(followset, p.tok) {
		if trace {
			p.print("skip " + p.tok.String())
		}
		p.next()
		if len(followlist) == 0 {
			break
		}
	}
	if trace {
		p.print("next " + p.tok.String())
	}
}

//--------------------------------------------------------------------------------
// usage: defer p.trace(msgs...)(endMsgs...)
func (p *parser) trace(msg string, args ...interface{}) func(string, ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	fmt.Printf("// %5d: %s%s (\n", p.line, p.indent, msg)
	const tab = ". "
	p.indent = append(p.indent, tab...)
	return func(endMsg string, endArgs ...interface{}) {
		if len(endArgs) > 0 {
			endMsg = fmt.Sprintf(endMsg, endArgs...)
		}
		p.indent = p.indent[:len(p.indent)-len(tab)]
		if x := recover(); x != nil {
			panic(x) // skip print_trace
		}
		fmt.Printf("// %5d: %s)", p.line, p.indent)
		if endMsg != "" {
			fmt.Printf(" %s", endMsg)
		}
		fmt.Println()
	}
}

//--------------------------------------------------------------------------------
func (p *parser) print(msg string) {
	fmt.Printf("%5d: %s%s\n", p.line, p.indent, msg)
}

//--------------------------------------------------------------------------------
// usage: p.tag(msgs...)
func (p *parser) tag(msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	fmt.Printf("// %5d: %s%s\n", p.line, p.indent, msg)
}

//--------------------------------------------------------------------------------
func (p *parser) isName(ss ...string) bool {
	for _, s := range ss {
		if p.tok == _Name && p.lit == s {
			return true
		}
	}
	return false
}

//--------------------------------------------------------------------------------
// functions not attached to parser
//--------------------------------------------------------------------------------
// tokstring returns the English word for selected punctuation tokens
// for more readable error messages.
func tokstring(tok token) string {
	switch tok {
	//case _EOF:
	//return "EOF"
	case _Comma:
		return "comma"
	case _Semi:
		return "semicolon"
	}
	return tok.String()
}

//--------------------------------------------------------------------------------
func isEmptyFuncDecl(dcl Decl) bool {
	f, ok := dcl.(*FuncDecl)
	return ok && f.Body == nil
}

//--------------------------------------------------------------------------------
func newIndirect(pos src.Pos, typ Expr) Expr {
	o := new(Operation)
	o.pos = pos
	o.Op = Mul
	o.X = typ
	return o
}

//--------------------------------------------------------------------------------
// unparen removes all parentheses around an expression.
func unparen(x Expr) Expr {
	for {
		p, ok := x.(*ParenExpr)
		if !ok {
			break
		}
		x = p.X
	}
	return x
}

//--------------------------------------------------------------------------------
