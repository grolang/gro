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
	"unicode"

	"github.com/grolang/gro/nodes"
	"github.com/grolang/gro/syntax/src"
)

const (
	dynLib = "\"github.com/grolang/gro/ops\""
)

const debug = false
const trace = false

const lineMax = 1<<24 - 1 // TODO(gri) this limit is defined for src.Pos - fix

//--------------------------------------------------------------------------------
type parser struct {
	scanner

	base *src.PosBase
	errh ErrorHandler
	mode Mode

	first  error  // first error encountered
	errcnt int    // number of errors encountered
	pragma Pragma // pragma flags

	fnest  int    // function nesting level (for error handling)
	xnest  int    // expression nesting level (for complit ambiguity resolution)
	indent []byte // tracing support

	//project, package, and section-level...
	currProj *nodes.Project
	currPkg  *nodes.Package
	currSect *nodes.File

	getFile        func(string) (string, error) // function for callback to read in another file
	docComments    string                       // buffer
	lineDirectives bool

	dynamicBlock string
	hashCmdBlock bool
	permits      map[string]bool
	paramdPkgs   map[string]*nodes.Package

	useRegistry  map[string]func([]string, []string)
	stmtRegistry map[string]func(nodes.GeneralParser, ...interface{}) nodes.Stmt
	typeRegistry map[string]func(nodes.GeneralParser, ...interface{}) nodes.Expr
}

//--------------------------------------------------------------------------------
func (p *parser) init(
	base *src.PosBase, r io.Reader, errh ErrorHandler, pragh PragmaHandler, mode Mode,
	getFile func(string) (string, error),
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
			p.ErrorAt(p.PosAt(line, col), msg)
		},
		func(line, col uint, text string) {
			if strings.HasPrefix(text, "line ") {
				p.updateBase(line, col+5, text[5:])
				return
			}
			if pragh != nil {
				p.pragma |= pragh(p.PosAt(line, col), text)
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
	p.paramdPkgs = map[string]*nodes.Package{}
}

//--------------------------------------------------------------------------------
func (p *parser) ProjFromNewParser(filename string, src []byte) (_ *nodes.Project, first error) {
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
	q.init(p.base, &bytesReader{src}, p.errh, nil, p.mode, p.getFile)
	q.Next()
	proj := q.Proj(filename)
	return proj, q.first
}

//--------------------------------------------------------------------------------
func (p *parser) SetPermit(s string)     { p.permits[s] = true }
func (p *parser) UnsetPermit(s string)   { p.permits[s] = false }
func (p *parser) IsPermit(s string) bool { return p.permits[s] }

func (p *parser) DynamicBlock() string     { return p.dynamicBlock }
func (p *parser) SetDynamicBlock(s string) { p.dynamicBlock = s }

func (p *parser) DynamicMode() bool     { return p.dynamicMode }
func (p *parser) SetDynamicMode(b bool) { p.dynamicMode = b }

func (p *parser) DynCharSet() string     { return p.dynCharSet }
func (p *parser) SetDynCharSet(s string) { p.dynCharSet = s }

func (p *parser) LineDirectives() bool     { return p.lineDirectives }
func (p *parser) SetLineDirectives(b bool) { p.lineDirectives = b }

func (p *parser) SetStmtRegistry(s string, f func(nodes.GeneralParser, ...interface{}) nodes.Stmt) {
	p.stmtRegistry[s] = f
}
func (p *parser) UnsetStmtRegistry(s string) {
	delete(p.stmtRegistry, s)
}

func (p *parser) Tok() nodes.Token { return p.tok }

func (p *parser) Lit() string {
	//TODO: only valid if tok is NameT, LiteralT, or SemiT ("semicolon", "newline", or "EOF")
	return p.lit
}

func (p *parser) Kind() nodes.LitKind {
	//TODO: only valid if tok is LiteralT
	return p.kind
}

func (p *parser) Op() nodes.Operator {
	//TODO: only valid if tok is OperatorT, AssignOpT, or IncOpT
	return p.op
}

func (p *parser) Prec() nodes.Prec {
	//TODO: only valid if tok is OperatorT, AssignOpT, or IncOpT
	return p.prec
}

//--------------------------------------------------------------------------------
func (p *parser) ProcImportAlias(lit *nodes.BasicLit, a string) string {
	if a == "" {
		_, a = filepath.Split(strings.Trim(lit.Value, "\""))
	}
	aa := p.NewName(a)
	a = aa.Value //in hash-cmd mode, a could be prepended by underscore
	if p.currSect.InfImpMap[a] != "" && p.currSect.InfImpMap[a] != lit.Value {
		p.SyntaxError(fmt.Sprintf("import alias \"%s\" has already been used but with different import path", a))
		return a
	} else if p.currSect.InfImpMap[a] == "" {
		p.currSect.InfImports = append(p.currSect.InfImports, &nodes.ImportDecl{
			Path:         &nodes.BasicLit{Value: lit.Value, Kind: nodes.StringLit},
			LocalPkgName: aa, //p.NewName(a),
		})
		p.currSect.InfImpMap[a] = lit.Value
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
		p.ErrorAt(p.PosAt(line, col+uint(i+1)), "invalid line number: "+nstr)
		return
	}
	p.base = src.NewLinePragmaBase(src.MakePos(p.base.Pos().Base(), line, col), text[:i], uint(n))
}

//--------------------------------------------------------------------------------
// Transform methods
//--------------------------------------------------------------------------------
// ProjToFiles transforms a project node into a set of named files.
//
// Note: These use parser methods/fields: NewBlankFunc, NewName, SyntaxError/At,
// dynCharSet, dynamicBlock, permits, paramdPkgs
//
func (p *parser) ProjToFiles(proj *nodes.Project) map[string]*nodes.File {
	w := new(nodes.Walker)
	proj.Walk(w)

	fs := map[string]*nodes.File{}
	for _, pkg := range proj.Pkgs { // for each file in each pkg, add to map of files returned (fs)
		if len(proj.Pkgs) > 1 && !p.permits["multiPkg"] {
			p.SyntaxErrorAt(proj.Pkgs[1].Pos(), permitErrorMsgs["multiPkg"])
			return nil
		}
		if p.currProj.DirStr != "" {
			pkg.Dir = filepath.ToSlash(filepath.Join(p.currProj.DirStr, pkg.Dir))
		}
		// if project keyword or more than one package, all explicit packages
		// have package-name both as filename and in directory name
		if p.currProj.HasKw || len(proj.Pkgs) > 1 {
			pkg.Dir = filepath.ToSlash(filepath.Join(pkg.Dir, pkg.Name))
		}
		for _, f := range pkg.Files {
			f.OwnerPkg = pkg
			for _, decl := range f.DeclList {
				if imp, ok := decl.(*nodes.ImportDecl); ok {
					imp.OwnerFile = f
					if imp.Path.Value == dynLib && p.dynCharSet == "utf88" {
						g := p.NewBlankFunc("init")
						gStmt := &nodes.AssignStmt{
							Op: 0, //=
							Lhs: &nodes.SelectorExpr{
								X:   p.NewName(p.dynamicBlock),
								Sel: &nodes.Name{Value: "UseUtf88"},
							},
							Rhs: &nodes.BasicLit{
								Value: "true",
							},
						}
						g.Body.List = append(g.Body.List, gStmt)
						f.DeclList = append(f.DeclList, g) //TODO: fix concurrent modification?
						p.dynCharSet = ""
					}
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
			} else if !p.currProj.HasKw && len(proj.Pkgs) == 1 && pkg.Name != "" {
				f.FileName = p.currProj.Name //use gro-filename as filename
			}
			fs[filepath.ToSlash(filepath.Join(p.currProj.Root, pkg.Dir, f.FileName))+".go"] = f
		}
	}
	for _, ai := range p.currProj.ArgImports {
		for k, v := range p.initGenerics(ai, "", map[string]bool{}) {
			fs[k] = v
		}
	}
	return fs
}

//--------------------------------------------------------------------------------
// initGenerics: for each import that had an argument/s, put its parameters
// into the package as types, then add that to fs.
func (p *parser) initGenerics(ai *nodes.ImportDecl, prefix string, done map[string]bool) map[string]*nodes.File {
	fs := map[string]*nodes.File{}
	if !p.checkPermit("genericCall") {
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
		p.SyntaxError("parameterized package not present in file, or there's a cycle in the parameterized imports")
		return nil
	}
	newpath := filepath.ToSlash(filepath.Join(prefix, "generics", ai.OwnerFile.OwnerPkg.Dir, ai.OwnerFile.FileName, ai.LocalPkgName.Value))
	for _, pf := range pp.Files {
		//recursively call initGenerics on each argImport in the parameterized pkg, where arg is one of pkg's params
		for _, decl := range pf.DeclList {
			if decl, ok := decl.(*nodes.ImportDecl); ok && len(decl.Args) > 0 {
				passThrus := map[int]int{} //map arg to param
				for m, arg := range decl.Args {
					if arg, ok := arg.(*nodes.Name); ok {
						for n, param := range pp.Params {
							if arg.Value == param.Value {
								passThrus[m] = n
							}
						}
					}
				}
				if len(passThrus) > 0 {
					args := []nodes.Expr{}
					for m, arg := range decl.Args {
						if n, ok := passThrus[m]; ok {
							args = append(args, ai.Args[n])
						} else {
							args = append(args, arg)
						}
					}
					infers := []*nodes.ImportDecl{}
					for _, infer := range decl.Infers {
						infers = append(infers, infer)
					}
					declClone := &nodes.ImportDecl{
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
	f := &nodes.File{
		PkgName:  p.NewName(pp.Name),
		DeclList: []nodes.Decl{},
	}
	g := new(nodes.DeclGroup)
	for _, inf := range ai.Infers {
		inf.Group = g
		f.DeclList = append(f.DeclList, inf)
	}
	for n, a := range ai.Args {
		f.DeclList = append(f.DeclList, &nodes.TypeDecl{
			Name:  pp.Params[n],
			Alias: true,
			Type:  a,
		})
	}
	fs[filepath.ToSlash(filepath.Join(p.currProj.Root, newpath, "generic_args"))+".go"] = f
	return fs
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
func (p *parser) Proj(filename string) (proj *nodes.Project) {
	if trace {
		defer p.trace("proj")("")
	}

	proj = new(nodes.Project)

	absName, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		p.Error("error computing absolute name for " + filename)
	}
	proj.Locn = filepath.ToSlash(absName)
	firstGoPath := filepath.SplitList(os.Getenv("GOPATH"))[0]
	srcPath := filepath.ToSlash(filepath.Join(firstGoPath, "src"))
	proj.Root = strings.TrimPrefix(strings.TrimPrefix(proj.Locn, srcPath), "/")
	b := filepath.Base(filename)
	ext := filepath.Ext(b)
	proj.Name = b[:len(b)-len(ext)]
	if len(ext) > 0 { // && ext[0] == '.'
		proj.FileExt = ext[1:]
	}

	p.currProj = proj
	p.setupProfile()
	p.setupRegistries()

	if p.IsName("project") {
		if !p.checkPermit("projectKw") {
			p.Advance(nodes.SemiT)
			return nil
		}
		if len(p.comments) != 0 {
			proj.Doc = p.comments
		}

		p.Next()
		if b := p.OLiteral(); b != nil { // is directory-string present?
			proj.DirStr = strings.Trim(b.Value, "\"")
		}
		proj.Name = p.Name().Value //change project name if explicit
		p.Want(nodes.SemiT)
		proj.HasKw = true
	}
	pkgs := []*nodes.Package{}
	for p.IsName("use") || p.IsName("include") {
		var fn func() *nodes.Project
		switch p.lit {
		case "use":
			fn = p.UseDecl
		case "include":
			fn = p.InclDecl
		}
		if !p.checkPermit(p.lit + "Kw") {
			p.Advance(nodes.SemiT)
			return nil
		}
		p.Want(nodes.NameT)

		if p.tok == nodes.LparenT {
			p.List(nodes.LparenT, nodes.SemiT, nodes.RparenT, func() bool {
				pkgs = append(pkgs, fn().Pkgs...)
				return false
			})
		} else {
			pkgs = append(pkgs, fn().Pkgs...)
		}
		p.Want(nodes.SemiT)
	}
	for p.tok != nodes.EofT {
		pkgs = append(pkgs, p.PkgOrNil())
	}
	proj.Pkgs = pkgs

	if !proj.HasKw && len(pkgs) == 0 {
		p.SyntaxErrorAt(src.MakePos(p.base, 1, 1), "gro-file empty")
		return nil
	} else if len(pkgs) == 0 {
		p.SyntaxError("project keyword but no packages")
		return nil
	}
	//fmt.Printf("name:%s, ext:%s, locn:%s, root:%s, str:%s\n", proj.Name, proj.FileExt, proj.Locn, proj.Root, proj.DirStr)
	return
}

//--------------------------------------------------------------------------------
func (p *parser) InclDecl() *nodes.Project {
	if trace {
		defer p.trace("inclDecl")("")
	}

	lit := p.OLiteral()
	if lit == nil {
		p.SyntaxError("missing name for gro-file after \"include\" keyword")
		p.Advance(nodes.SemiT, nodes.RparenT)
		return nil
	}
	fileLocn := filepath.ToSlash(filepath.Join(p.currProj.Locn, strings.Trim(lit.Value, "\"")))
	src, err := p.getFile(fileLocn)
	if err != nil {
		p.Error(fmt.Sprintf("error \"%s\" retrieving included file %s", err, fileLocn))
		return nil
	}
	proj, err := p.ProjFromNewParser(fileLocn, []byte(src))
	if err != nil {
		p.Error(fmt.Sprintf("error \"%s\" parsing included file %s", err, fileLocn))
		return nil
	}
	p.currProj.ArgImports = append(p.currProj.ArgImports, proj.ArgImports...)
	return proj
}

//--------------------------------------------------------------------------------
func (p *parser) UseDecl() (proj *nodes.Project) {
	if trace {
		defer p.trace("useDecl")("")
	}

	rets := []string{}
	for p.tok == nodes.NameT || p.TokIsKeywordName() {
		rets = append(rets, strings.Trim(p.Name().Value, "\""))
	}
	useStr := p.OLiteral()
	if useStr == nil || useStr.Kind != nodes.StringLit {
		p.SyntaxError("missing use string")
		p.Advance(nodes.SemiT, nodes.RparenT)
		return nil
	}
	use := strings.Trim(useStr.Value, "\"")
	args := []string{}
	if p.tok == nodes.LparenT {
		p.List(nodes.LparenT, nodes.CommaT, nodes.RparenT, func() bool {
			s := strings.Trim(p.OLiteral().Value, "\"") //TODO: generate error if not a string?
			args = append(args, s)
			return false
		})
	}

	if useCase, ok := p.useRegistry[use]; ok {
		useCase(rets, args)
	} else {
		//p.SyntaxError(fmt.Sprintf("use \"%s\" not implemented", use))
		//p.Advance(nodes.SemiT, nodes.RbraceT)
	}
	return &nodes.Project{Pkgs: []*nodes.Package{}} //dud project
}

//--------------------------------------------------------------------------------
// SourceFile = PackageClause ";" { ImportDecl ";" } { TopLevelDecl ";" } .
func (p *parser) PkgOrNil() *nodes.Package {
	if trace {
		defer p.trace("pkgOrNil")("")
	}

	pkg := &nodes.Package{
		IdsUsed: map[string]bool{},
	}
	pkg.Name = ""
	bracesUsed := false

	for { // each section
		f := new(nodes.File)
		f.InfImpMap = map[string]string{}

		f.SetPos(p.Pos())
		pkg.SetPos(f.Pos())
		if p.lineDirectives {
			f.TagLineDirect()
		}
		f.MakeComments()
		if p.currProj.Doc != nil && len(p.currProj.Doc) > 0 {
			//add project-level doc to top of every package file
			f.AppendAloneComment(strings.Join(p.currProj.Doc, "\n"))
		}

		if p.tok == nodes.PackageT || p.IsName("internal") { // first time thru loop when there's a package keyword
			if !p.checkPermit("packageKw") {
				p.Advance(nodes.SemiT)
				return nil
			}
			for _, cg := range p.commentGroups {
				f.AppendAloneComment(strings.Join(cg, "\n"))
			}
			if len(p.comments) > 0 {
				f.SetAboveComment(strings.Join(p.comments, "\n"))
			}
			if p.IsName("internal") {
				if !p.checkPermit("internalKw") {
					p.Advance(nodes.SemiT)
					return nil
				}
				pkg.Dir = "internal"
			}
			p.Next()

			// if directory-string present, add after directory
			if b := p.OLiteral(); b != nil {
				pkg.Dir = filepath.ToSlash(filepath.Join(pkg.Dir, strings.Trim(b.Value, "\"")))
			}
			f.PkgName = p.Name()
			pkg.Name = f.PkgName.Value

			if p.tok == nodes.LparenT {
				if !p.checkPermit("genericDef") {
					p.Advance(nodes.SemiT)
					return nil
				}
				p.List(nodes.LparenT, nodes.CommaT, nodes.RparenT, func() bool {
					pkg.Params = append(pkg.Params, p.Name())
					return false
				})
			}

			if p.Got(nodes.LbraceT) {
				if !p.checkPermit("pkgSectBlocks") {
					p.Advance(nodes.SemiT)
					return nil
				}
				bracesUsed = true
			} else {
				p.Want(nodes.SemiT)
			}

		} else if pkg.Name != "" { // subsequent times thru loop when there was a package keyword
			f.PkgName = p.NewName(pkg.Name)

		} else { // first or subsequent time thru loop but no package keyword
			if !p.checkPermit("inferPkg") {
				p.Advance(nodes.SemiT)
				return nil
			}
			f.PkgName = p.NewName(p.currProj.Name)

		}

		p.currPkg = pkg
		f = p.SectionOrNil(f)
		f.FileName = f.PkgName.Value

		// if package without keyword, and main fn defined, use "main" as package-name
		if pkg.Name == "" && p.currSect.HasMain && p.permits["inferMain"] ||
			p.currSect.HeadKw == "main" && p.currSect.HasMain && p.permits["inferMain"] {
			f.PkgName.Value = "main"
			f.AppendAloneComment("// +build ignore")
		}
		if p.currSect.HasStmts && !p.currSect.HasMain && pkg.Name == "" ||
			p.currSect.HeadKw == "main" && !p.currSect.HasMain {
			if p.permits["inferMain"] {
				f.PkgName.Value = "main"
				f.AppendAloneComment("// +build ignore")
				f.DeclList = append(f.DeclList, p.NewBlankFunc("main"))
			}
		}
		pkg.Files = append(pkg.Files, f)
		if p.tok == nodes.EofT || p.tok == nodes.PackageT || p.IsName("internal") || (bracesUsed && p.tok == nodes.RbraceT) {
			break
		}
	}

	if bracesUsed && !(p.Got(nodes.RbraceT) && p.Got(nodes.SemiT)) {
		p.SyntaxError("missing right brace after package block")
		p.Advance(nodes.SemiT, nodes.RbraceT)
		return nil
	}
	dynTypeGroup := &nodes.DeclGroup{}
	if p.dynamicMode && !pkg.IdsUsed["any"] && len(pkg.Files) > 0 {
		d := &nodes.TypeDecl{
			Name:  p.NewName("any"),
			Alias: true,
			Type:  &nodes.InterfaceType{MethodList: []*nodes.Field{}},
			Group: dynTypeGroup,
			//Pragma: p.pragma,
		}
		pkg.Files[0].DeclList = append(pkg.Files[0].DeclList, d)
		pkg.IdsUsed["any"] = true
	}
	if p.dynamicMode && !pkg.IdsUsed["void"] && len(pkg.Files) > 0 {
		d := &nodes.TypeDecl{
			Name:  p.NewName("void"),
			Alias: true,
			Type:  &nodes.StructType{FieldList: []*nodes.Field{}},
			Group: dynTypeGroup,
			//Pragma: p.pragma,
		}
		pkg.Files[0].DeclList = append(pkg.Files[0].DeclList, d)
		pkg.IdsUsed["void"] = true
	}

	if p.dynamicMode && !pkg.IdsUsed["inf"] && len(pkg.Files) > 0 {
		vd := new(nodes.VarDecl)
		vd.NameList = []*nodes.Name{p.NewName("inf")}
		needDynImport := p.dynamicBlock == ""
		if needDynImport {
			p.dynamicBlock = "groo"
		}
		vd.Values = &nodes.SelectorExpr{
			X:   p.NewName(p.dynamicBlock),
			Sel: &nodes.Name{Value: "Inf"},
		}
		pkg.Files[0].DeclList = append(pkg.Files[0].DeclList, vd)
		if needDynImport {
			id := &nodes.ImportDecl{
				Path:         &nodes.BasicLit{Value: dynLib, Kind: nodes.StringLit},
				LocalPkgName: p.NewName(p.dynamicBlock),
			}
			pkg.Files[0].DeclList = append([]nodes.Decl{id}, pkg.Files[0].DeclList...)
		}
	}
	return pkg
}

//--------------------------------------------------------------------------------
// Section
func (p *parser) SectionOrNil(f *nodes.File) *nodes.File {
	if trace {
		defer p.trace("sectionOrNil")("")
	}

	bracesUsed := false
	currPos := p.Pos()
	if p.IsName("section", "main", "testcode") {
		if !p.checkPermit("sectionKw") || (p.IsName("main") && !p.checkPermit("mainKw")) ||
			!p.checkPermit("testcodeKw") {
			p.Advance(nodes.SemiT)
			return nil
		}
		for _, cg := range p.commentGroups {
			f.DeclList = append(f.DeclList, &nodes.CommentDecl{CommentList: []*nodes.Comment{{Text: strings.Join(cg, "\n")}}})
		}
		if len(p.comments) > 0 {
			f.DeclList = append(f.DeclList, &nodes.CommentDecl{CommentList: []*nodes.Comment{{Text: strings.Join(p.comments, "\n")}}})
		}

		f.HeadKw = p.lit
		p.Next()
		lit := p.OLiteral()
		if lit == nil {
			p.SyntaxError("missing section name")
			p.Advance(nodes.SemiT, nodes.RparenT)
			return nil
		}
		if p.Got(nodes.LbraceT) {
			if !p.checkPermit("pkgSectBlocks") {
				p.Advance(nodes.SemiT)
				return nil
			}
			bracesUsed = true
		} else {
			p.Want(nodes.SemiT)
		}
		f.SectName = strings.Trim(lit.Value, "\"")
		if f.HeadKw == "testcode" {
			f.SectName += "_test"
		}
	}

	p.currSect = f

	// { ImportDecl ";" }
	isHash := p.hash
	for p.Got(nodes.ImportT) {
		if !p.checkPermit("importKw") {
			p.Advance(nodes.SemiT, nodes.RbraceT)
			return nil
		}
		p.CheckHashCmd(isHash, func() {
			f.DeclList = p.AppendGroup(f.DeclList, p.ImportDecl)
		})
		p.Want(nodes.SemiT)
	}
	if p.currSect.HeadKw == "testcode" {
		f.PkgName.Value = p.currPkg.Name
	}

	// { TopLevelDecl ";" }
	for p.tok != nodes.EofT && p.tok != nodes.PackageT && p.tok != nodes.RbraceT && !p.IsName("internal", "section", "main", "testcode") {
		switch p.tok {
		case nodes.ConstT, nodes.VarT, nodes.TypeT, nodes.FuncT: //declarations
			f.DeclList = p.Decl(f.DeclList)
		case nodes.IfT, nodes.ForT, nodes.SwitchT, nodes.SelectT, nodes.GoT, nodes.LbraceT, nodes.SemiT, nodes.LiteralT: //tl-stmts
			f.DeclList = append(f.DeclList, p.TlBlock())
		default:
			if p.IsName("do") { //do-stmt
				f.DeclList = append(f.DeclList, p.TlBlock())
			} else if p.IsName("proc") { //proc-decl
				f.DeclList = p.Decl(f.DeclList)
			} else if _, ok := p.stmtRegistry[p.lit]; p.tok == nodes.NameT && ok { //macros
				f.DeclList = append(f.DeclList, p.TlBlock())
			} else {
				p.SyntaxError(fmt.Sprintf("unexpected %s at top-level", p.tok))
				return nil
			}
		}
	}
	if currPos == p.Pos() && p.tok != nodes.EofT {
		p.SyntaxError(fmt.Sprintf("unexpected token %s", p.tok))
		return nil
	}

	if len(f.InfImports) > 0 {
		g := new(nodes.DeclGroup)
		for i := len(f.InfImports) - 1; i >= 0; i-- {
			imp := f.InfImports[i]
			imp.Group = g
			f.DeclList = append([]nodes.Decl{imp}, f.DeclList...)
		}
	}
	if bracesUsed && !(p.Got(nodes.RbraceT) && p.Got(nodes.SemiT)) {
		p.SyntaxError("missing right brace after section block")
		p.Advance(nodes.SemiT, nodes.RbraceT)
		return nil
	}
	f.Lines = p.source.line
	return f
}

//--------------------------------------------------------------------------------
// Declarations
//--------------------------------------------------------------------------------
// TopLevelDecl
func (p *parser) Decl(declList []nodes.Decl) []nodes.Decl {
	switch {
	case p.tok == nodes.ConstT:
		if !p.checkPermit("constKw") {
			p.Advance(nodes.SemiT)
			return nil
		}
		for _, cg := range p.commentGroups {
			declList = append(declList, &nodes.CommentDecl{CommentList: []*nodes.Comment{{Text: strings.Join(cg, "\n")}}})
		}
		if len(p.comments) > 0 {
			p.docComments = strings.Join(p.comments, "\n")
		}
		p.CheckHashCmd(p.hash, func() {
			p.Next()
			declList = p.AppendGroup(declList, p.ConstDecl)
		})

	case p.tok == nodes.TypeT:
		if !p.checkPermit("typeKw") {
			p.Advance(nodes.SemiT)
			return nil
		}
		for _, cg := range p.commentGroups {
			declList = append(declList, &nodes.CommentDecl{CommentList: []*nodes.Comment{{Text: strings.Join(cg, "\n")}}})
		}
		if len(p.comments) > 0 {
			p.docComments = strings.Join(p.comments, "\n")
		}
		p.CheckHashCmd(p.hash, func() {
			p.Next()
			declList = p.AppendGroup(declList, p.TypeDecl)
		})

	case p.tok == nodes.VarT:
		if !p.checkPermit("varKw") {
			p.Advance(nodes.SemiT)
			return nil
		}
		for _, cg := range p.commentGroups {
			declList = append(declList, &nodes.CommentDecl{CommentList: []*nodes.Comment{{Text: strings.Join(cg, "\n")}}})
		}
		if len(p.comments) > 0 {
			p.docComments = strings.Join(p.comments, "\n")
		}
		p.CheckHashCmd(p.hash, func() {
			p.Next()
			declList = p.AppendGroup(declList, p.VarDecl)
		})

	case p.tok == nodes.FuncT:
		if !p.checkPermit("funcKw") {
			p.Advance(nodes.SemiT)
			return nil
		}
		for _, cg := range p.commentGroups {
			declList = append(declList, &nodes.CommentDecl{CommentList: []*nodes.Comment{{Text: strings.Join(cg, "\n")}}})
		}
		if len(p.comments) > 0 {
			p.docComments = strings.Join(p.comments, "\n")
		}
		p.CheckHashCmd(p.hash, func() {
			p.Next()
			if d := p.FuncDeclOrNil(p.StmtOrNil); d != nil {
				declList = append(declList, d)
			}
		})

	case p.IsName("proc"):
		if !p.checkPermit("procKw") {
			p.Advance(nodes.SemiT)
			return nil
		}
		for _, cg := range p.commentGroups {
			declList = append(declList, &nodes.CommentDecl{CommentList: []*nodes.Comment{{Text: strings.Join(cg, "\n")}}})
		}
		if len(p.comments) > 0 {
			p.docComments = strings.Join(p.comments, "\n")
		}
		p.CheckHashCmd(p.hash, func() {
			p.Next()
			if d := p.FuncDeclOrNil(p.ProcStmt); d != nil {
				declList = append(declList, d)
			}
		})

	default:
		if p.tok == nodes.LbraceT && len(declList) > 0 && isEmptyFuncDecl(declList[len(declList)-1]) {
			// opening { of function declaration on next line
			p.SyntaxError("unexpected semicolon or newline before {")
		} else {
			p.SyntaxError("non-declaration statement outside function body")
		}
		p.Advance(nodes.ConstT, nodes.TypeT, nodes.VarT, nodes.FuncT)
		return declList
	}

	// Reset p.pragma BEFORE advancing to the next token (consuming ';')
	// since comments before may set pragmas for the next function decl.
	p.pragma = 0

	if p.tok != nodes.EofT && !p.Got(nodes.SemiT) {
		p.SyntaxError("after top level declaration")
		p.Advance(nodes.ConstT, nodes.TypeT, nodes.VarT, nodes.FuncT)
	}

	return declList
}

//--------------------------------------------------------------------------------
// appendGroup(f) = f | "(" { f ";" } ")" . // ";" is optional before ")"
func (p *parser) AppendGroup(list []nodes.Decl, f func(*nodes.DeclGroup) nodes.Decl) []nodes.Decl {
	setIdsUsed := func(f nodes.Decl) {
		switch ft := f.(type) {
		case *nodes.ConstDecl:
			for _, name := range ft.NameList {
				if name.Value != "_" {
					p.currPkg.IdsUsed[name.Value] = true
				}
			}
		case *nodes.VarDecl:
			for _, name := range ft.NameList {
				if name.Value != "_" {
					p.currPkg.IdsUsed[name.Value] = true
				}
			}
		case *nodes.TypeDecl:
			if ft.Name.Value != "_" {
				p.currPkg.IdsUsed[ft.Name.Value] = true
			}
		}
	}

	if p.tok == nodes.LparenT {
		g := new(nodes.DeclGroup)

		g.SetPos(p.Pos())
		if p.lineDirectives {
			g.TagLineDirect()
		}
		if p.docComments != "" {
			g.MakeComments()
			g.SetAboveComment(p.docComments)
			p.docComments = ""
		}

		p.List(nodes.LparenT, nodes.SemiT, nodes.RparenT, func() bool {
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
func (p *parser) ImportDecl(group *nodes.DeclGroup) nodes.Decl {
	if trace {
		defer p.trace("importDecl")("")
	}

	d := new(nodes.ImportDecl)
	d.SetPos(p.Pos())
	if p.lineDirectives {
		d.TagLineDirect()
	}

	if p.tok == nodes.NameT || p.TokIsKeywordName() {
		d.LocalPkgName = p.Name()
		//only explicit local names of imports can be analyzed at the syntactic phase...
		if d.LocalPkgName.Value != "_" {
			p.currPkg.IdsUsed[d.LocalPkgName.Value] = true
		}
	} else {
		switch p.tok {
		case nodes.DotT:
			d.LocalPkgName = p.NewName(".")
			p.Next()
		}
	}
	d.Path = p.OLiteral()
	if d.Path == nil {
		p.SyntaxError("missing import path")
		p.Advance(nodes.SemiT, nodes.RparenT)
		return nil
	}
	if p.currSect.HeadKw == "testcode" {
		_, pkgTag := filepath.Split(strings.Trim(d.Path.Value, "\""))
		if d.LocalPkgName != nil && d.LocalPkgName.Value == p.currPkg.Name || pkgTag == p.currPkg.Name {
			p.currPkg.Name += "_test"
		}
	}
	d.Group = group
	if p.tok == nodes.LparenT {
		d.Args = []nodes.Expr{}
		if d.LocalPkgName == nil {
			p.SyntaxError("imported parameterized packages need a local name")
			p.Advance(nodes.RparenT)
			return nil
		}
		p.List(nodes.LparenT, nodes.CommaT, nodes.RparenT, func() bool {
			arg := p.TypeOrNil()
			d.Args = append(d.Args, arg)
			return false
		})

		numPassThrus := 0 //TODO: only need a bool
		for _, arg := range d.Args {
			if arg, ok := arg.(*nodes.Name); ok {
				for _, param := range p.currPkg.Params {
					if arg.Value == param.Value {
						numPassThrus++
					}
				}
			}
		}
		if numPassThrus == 0 {
			p.currProj.ArgImports = append(p.currProj.ArgImports, d)
		}
		if len(p.currSect.InfImports) > 0 {
			d.Infers = p.currSect.InfImports
			p.currSect.InfImports = []*nodes.ImportDecl{}
		}
	}
	if p.dynamicMode && d.LocalPkgName == nil {
		_, a := filepath.Split(strings.Trim(d.Path.Value, "\""))
		d.LocalPkgName = p.NewName(a)
		p.currPkg.IdsUsed[a] = true
	}
	return d
}

//--------------------------------------------------------------------------------
// ConstSpec = IdentifierList [ [ Type ] "=" ExpressionList ] .
func (p *parser) ConstDecl(group *nodes.DeclGroup) nodes.Decl {
	if trace {
		defer p.trace("constDecl")("")
	}

	d := new(nodes.ConstDecl)
	d.SetPos(p.Pos())
	if p.lineDirectives {
		d.TagLineDirect()
	}
	if p.docComments != "" {
		d.MakeComments()
		d.SetAboveComment(p.docComments)
		p.docComments = ""
	}
	d.NameList = p.NameList(p.Name())
	if p.tok != nodes.EofT && p.tok != nodes.SemiT && p.tok != nodes.RparenT {
		d.Type = p.TypeOrNil()
		if p.Got(nodes.AssignT) {
			d.Values = p.ExprList(false)
		}
	}
	d.Group = group
	if len(p.comments) > 0 {
		if d.Comments() == nil {
			d.MakeComments()
		}
		d.SetRightComment(strings.Join(p.comments, "\n"))
	}
	return d
}

//--------------------------------------------------------------------------------
// TypeSpec = identifier [ "=" ] Type .
func (p *parser) TypeDecl(group *nodes.DeclGroup) nodes.Decl {
	if trace {
		defer p.trace("typeDecl")("")
	}

	d := new(nodes.TypeDecl)
	d.SetPos(p.Pos())
	if p.lineDirectives {
		d.TagLineDirect()
	}
	if p.docComments != "" {
		d.MakeComments()
		d.SetAboveComment(p.docComments)
		p.docComments = ""
	}

	d.Name = p.Name()
	d.Alias = p.Got(nodes.AssignT)
	d.Type = p.TypeOrNil()
	if d.Type == nil {
		d.Type = p.BadExpr()
		p.SyntaxError("in type declaration")
		p.Advance(nodes.SemiT, nodes.RparenT)
	}
	d.Group = group
	//d.Pragma = p.pragma
	if len(p.comments) > 0 {
		if d.Comments() == nil {
			d.MakeComments()
		}
		d.SetRightComment(strings.Join(p.comments, "\n"))
	}
	return d
}

//--------------------------------------------------------------------------------
// VarSpec = IdentifierList ( Type [ "=" ExpressionList ] | "=" ExpressionList ) .
func (p *parser) VarDecl(group *nodes.DeclGroup) nodes.Decl {
	if trace {
		defer p.trace("varDecl")("")
	}

	d := new(nodes.VarDecl)
	d.SetPos(p.Pos())
	if p.lineDirectives {
		d.TagLineDirect()
	}
	if p.docComments != "" {
		d.MakeComments()
		d.SetAboveComment(p.docComments)
		p.docComments = ""
	}
	d.NameList = p.NameList(p.Name())
	if p.Got(nodes.AssignT) {
		d.Values = p.ExprList(true)
	} else {
		d.Type = p.Type()
		if p.Got(nodes.AssignT) {
			d.Values = p.ExprList(true)
		}
	}
	d.Group = group
	if len(p.comments) > 0 {
		if d.Comments() == nil {
			d.MakeComments()
		}
		d.SetRightComment(strings.Join(p.comments, "\n"))
	}
	return d
}

//--------------------------------------------------------------------------------
// FunctionDecl = "func" FunctionName ( Function | Signature ) .
// FunctionName = identifier .
// Function     = Signature FunctionBody .
// MethodDecl   = "func" Receiver MethodName ( Function | Signature ) .
// Receiver     = Parameters .
func (p *parser) FuncDeclOrNil(stmt func() nodes.Stmt) *nodes.FuncDecl {
	if trace {
		defer p.trace("funcDecl")("")
	}

	f := new(nodes.FuncDecl)
	f.SetPos(p.Pos())
	if p.lineDirectives {
		f.TagLineDirect()
	}
	if p.docComments != "" {
		f.MakeComments()
		f.SetAboveComment(p.docComments)
		p.docComments = ""
	}

	if p.tok == nodes.LparenT {
		rcvr := p.ParamList()
		switch len(rcvr) {
		case 0:
			p.Error("method has no receiver")
		default:
			p.Error("method has multiple receivers")
			fallthrough
		case 1:
			f.Recv = rcvr[0]
		}
	}

	if p.tok != nodes.NameT && !p.TokIsKeywordName() {
		p.SyntaxError("expecting name or (")
		p.Advance(nodes.LbraceT, nodes.SemiT)
		return nil
	}

	// TODO(gri) check for regular functions only
	// if name.Sym.Name == "init" {
	// 	name = renameinit()
	// 	if params != nil || result != nil {
	// 		p.error("func init must have no arguments and no return values")
	// 	}
	// }

	f.Name = p.Name()
	f.Type = p.FuncType()
	if p.currPkg.Name == "main" && f.Name.Value == "main" {
		if len(f.Type.ParamList) != 0 || len(f.Type.ResultList) != 0 {
			p.Error("func main must have no arguments and no return values")
		}
	}
	if f.Name.Value == "main" && len(f.Type.ParamList) == 0 && len(f.Type.ResultList) == 0 {
		p.currSect.HasMain = true
	}
	if p.tok == nodes.LbraceT {
		f.Body = p.FuncBody(stmt)
	}

	//f.Pragma = p.pragma
	return f
}

//--------------------------------------------------------------------------------
// Statements
//--------------------------------------------------------------------------------
// Possible statement for init() =
// 	IfStmt | ForStmt | SwitchStmt | SelectStmt | DeferStmt | GoStmt | ReturnStmt | DoStmt | ToplevelBlock | StringLit.
func (p *parser) TlBlock() *nodes.FuncDecl {
	if trace {
		defer p.trace("toplevel standalone stmts")("")
	}

	p.currSect.HasStmts = true
	f := p.NewBlankFunc("init")
	l := []nodes.Stmt{}
forloop:
	for {
		switch p.tok {
		//these keyword-based statements can be standalone
		case nodes.IfT, nodes.ForT, nodes.SwitchT, nodes.SelectT, nodes.GoT, nodes.VarT,
			nodes.ConstT, nodes.TypeT, nodes.LbraceT, nodes.LiteralT:
			l = append(l, p.TlStmt())
		default:
			if p.IsName("do") || (p.tok == nodes.NameT && p.stmtRegistry[p.lit] != nil) {
				if p.lineDirectives {
					f.TagLineDirect()
				}
				if len(p.comments) > 0 {
					f.MakeComments()
					if len(p.comments) > 0 {
						f.SetAboveComment(strings.Join(p.comments, "\n"))
					}
				}
				l = append(l, p.TlStmt())
			} else {
				break forloop
			}
		}

		// ";" is optional before "}"
		if p.tok != nodes.EofT && !p.Got(nodes.SemiT) && p.tok != nodes.RbraceT {
			p.SyntaxError("at end of statement")
			p.Advance(nodes.ConstT, nodes.TypeT, nodes.VarT, nodes.FuncT, nodes.IfT, nodes.ForT, nodes.SwitchT,
				nodes.SelectT, nodes.GoT, nodes.SemiT, nodes.RbraceT, nodes.CaseT, nodes.DefaultT)
			p.Got(nodes.SemiT) // avoid spurious empty statement
			return nil
		}
	}
	f.Body.Rbrace = p.Pos()
	f.Body.List = l
	return f
}

//--------------------------------------------------------------------------------
// Possible statement for init() =
// 	IfStmt | ForStmt | SwitchStmt | SelectStmt | DeferStmt | GoStmt | ReturnStmt | DoStmt | ToplevelBlock | StringLit.
func (p *parser) TlStmtList(stmt func() nodes.Stmt) []nodes.Stmt {
	if trace {
		defer p.trace("seq of toplevel-style stmts")("")
	}

	l := []nodes.Stmt{}
forloop:
	for {
		switch p.tok {
		//these keyword-based statements can be standalone
		case nodes.IfT, nodes.ForT, nodes.SwitchT, nodes.SelectT, nodes.GoT, nodes.DeferT, nodes.VarT, nodes.ConstT, nodes.TypeT, nodes.LbraceT, nodes.LiteralT:
			l = append(l, stmt())
		default:
			if p.IsName("do") || (p.tok == nodes.NameT && p.stmtRegistry[p.lit] != nil) {
				l = append(l, stmt())
			} else {
				break forloop
			}
		}

		// ";" is optional before "}"
		if p.tok != nodes.EofT && !p.Got(nodes.SemiT) && p.tok != nodes.RbraceT {
			p.SyntaxError("at end of statement")
			p.Advance(nodes.ConstT, nodes.TypeT, nodes.VarT, nodes.FuncT, nodes.IfT, nodes.ForT, nodes.SwitchT, nodes.SelectT, nodes.GoT, nodes.SemiT, nodes.RbraceT, nodes.CaseT, nodes.DefaultT)
			p.Got(nodes.SemiT) // avoid spurious empty statement
			return nil
		}
	}
	return l
}

//--------------------------------------------------------------------------------
func (p *parser) TlStmt() nodes.Stmt {
	if trace {
		defer p.trace("top-level standalone stmt")("")
	}

	switch p.tok {
	case nodes.LiteralT:
		return p.SimpleStmt(nil, false)

	case nodes.VarT:
		return p.DeclStmt(p.VarDecl)

	case nodes.ConstT:
		return p.DeclStmt(p.ConstDecl)

	case nodes.TypeT:
		return p.DeclStmt(p.TypeDecl)

	case nodes.IfT:
		return p.IfStmt(p.ProcStmt)

	case nodes.ForT:
		return p.ForStmt(p.ProcStmt)

	case nodes.SwitchT:
		return p.SwitchStmt(p.ProcStmt)

	case nodes.SelectT:
		return p.SelectStmt(p.ProcStmt)

	case nodes.GoT:
		return p.CallStmt(p.ProcStmt)

	case nodes.LbraceT:
		return p.BlockStmt("", p.ProcStmt)

	case nodes.SemiT:
		s := new(nodes.EmptyStmt)
		s.SetPos(p.Pos())
		return s

	default:
		if p.IsName("do") {
			if !p.checkPermit("doKw") {
				p.Advance(nodes.SemiT)
				return nil
			}
			var r nodes.Stmt
			p.CheckHashCmd(p.hash, func() {
				p.Want(nodes.NameT)
				if p.tok == nodes.NameT || p.TokIsKeywordName() {
					lhs := p.ExprList(false)
					r = p.DynamicAssignOpOrSimpleStmt(lhs)
				} else {
					switch p.tok {
					//don't include labelled stmts, var, const, func, type,
					// defer, fallthrough, continue, break, goto, return within do-statements
					case nodes.IfT, nodes.ForT, nodes.SwitchT, nodes.SelectT, nodes.GoT:
						r = p.StmtOrNil()
					case nodes.LbraceT:
						r = p.BlockStmt("", p.StmtOrNil)
					case nodes.OperatorT, nodes.StarT:
						switch p.op {
						case nodes.Add, nodes.Sub, nodes.Mul, nodes.And, nodes.Xor, nodes.Not:
							r = p.SimpleStmt(nil, false) // unary operators
						}
					case nodes.LiteralT, nodes.FuncT, nodes.LparenT, // operands
						nodes.LbrackT, nodes.StructT, nodes.MapT, nodes.ChanT, nodes.InterfaceT, // composite types
						nodes.ArrowT: // receive operator
						r = p.SimpleStmt(nil, false)
					}
				}
			})
			if r != nil {
				return r
			}
		}
		if mac := p.stmtRegistry[p.lit]; p.tok == nodes.NameT && mac != nil {
			p.Next()
			return mac(p, p.TlStmt) //p.tlStmt needed for "let"
		}
	}

	return nil //should never reach
}

//--------------------------------------------------------------------------------
func (p *parser) ProcStmt() nodes.Stmt {
	if trace {
		defer p.trace("standalone stmt within proc")("")
	}

	if p.IsName("do") {
		if !p.checkPermit("doKw") {
			p.Advance(nodes.SemiT)
			return nil
		}
		var r nodes.Stmt
		p.CheckHashCmd(p.hash, func() {
			p.Want(nodes.NameT)
			if p.tok == nodes.NameT || p.TokIsKeywordName() {
				lhs := p.ExprList(false)
				r = p.DynamicAssignOpOrSimpleStmt(lhs)
			} else {
				switch p.tok {
				//don't include labelled stmts, var, const, func, type,
				// fallthrough, continue, break, goto, return within do-statements
				case nodes.IfT, nodes.ForT, nodes.SwitchT, nodes.SelectT, nodes.GoT, nodes.DeferT:
					r = p.StmtOrNil()
				case nodes.LbraceT:
					r = p.BlockStmt("", p.StmtOrNil)
				case nodes.OperatorT, nodes.StarT:
					switch p.op {
					case nodes.Add, nodes.Sub, nodes.Mul, nodes.And, nodes.Xor, nodes.Not:
						r = p.SimpleStmt(nil, false) // unary operators
					}
				case nodes.LiteralT, nodes.FuncT, nodes.LparenT, // operands
					nodes.LbrackT, nodes.StructT, nodes.MapT, nodes.ChanT, nodes.InterfaceT, // composite types
					nodes.ArrowT: // receive operator
					r = p.SimpleStmt(nil, false)
				}
			}
		})
		if r != nil {
			return r
		}
	} else if mac := p.stmtRegistry[p.lit]; p.tok == nodes.NameT && mac != nil {
		p.Next()
		return mac(p, p.ProcStmt) //p.procStmt needed for "let"
	} else if p.tok == nodes.NameT || p.TokIsKeywordName() { //TODO: check label as first if-option
		pos := p.Pos()
		lhs := p.ExprList(false)
		if label, ok := lhs.(*nodes.Name); ok && p.tok == nodes.ColonT {
			return p.LabeledStmtOrNil(label)
		}
		p.SyntaxErrorAt(pos, fmt.Sprintf("unexpected name, expecting \"do\", label or macro name"))
		return nil
	}
	switch p.tok {
	case nodes.LiteralT:
		return p.SimpleStmt(nil, false)

	case nodes.VarT:
		return p.DeclStmt(p.VarDecl)

	case nodes.ConstT:
		return p.DeclStmt(p.ConstDecl)

	case nodes.TypeT:
		return p.DeclStmt(p.TypeDecl)

	case nodes.IfT:
		return p.IfStmt(p.ProcStmt)

	case nodes.ForT:
		return p.ForStmt(p.ProcStmt)

	case nodes.SwitchT:
		return p.SwitchStmt(p.ProcStmt)

	case nodes.SelectT:
		return p.SelectStmt(p.ProcStmt)

	case nodes.GoT, nodes.DeferT:
		return p.CallStmt(p.ProcStmt)

	case nodes.LbraceT:
		return p.BlockStmt("", p.ProcStmt)

	case nodes.SemiT:
		s := new(nodes.EmptyStmt)
		s.SetPos(p.Pos())
		return s

	case nodes.FallthroughT:
		if !p.checkPermit("fallthroughKw") {
			p.Advance(nodes.SemiT)
			return nil
		}
		s := new(nodes.BranchStmt)
		s.SetPos(p.Pos())
		p.Next()
		s.Tok = nodes.FallthroughT
		return s

	case nodes.BreakT, nodes.ContinueT:
		return p.BreakOrContinueStmt()

	case nodes.GotoT:
		return p.GotoStmt()

	case nodes.ReturnT:
		return p.ReturnStmt()
	}

	return nil //should never reach
}

//--------------------------------------------------------------------------------
// context must be a non-empty string unless we know that p.tok == _Lbrace.
func (p *parser) BlockStmt(context string, stmt func() nodes.Stmt) *nodes.BlockStmt {
	if trace {
		defer p.trace("blockStmt")("")
	}

	s := new(nodes.BlockStmt)
	s.SetPos(p.Pos())

	// people coming from C may forget that braces are mandatory in Go
	if !p.Got(nodes.LbraceT) {
		p.SyntaxError("expecting { after " + context)
		p.Advance(nodes.NameT, nodes.RbraceT)
		s.Rbrace = p.Pos() // in case we found "}"
		if p.Got(nodes.RbraceT) {
			return s
		}
	}

	s.List = p.StmtList(stmt)
	s.Rbrace = p.Pos()
	p.Want(nodes.RbraceT)

	return s
}

//--------------------------------------------------------------------------------
// StatementList = { Statement ";" } .
func (p *parser) StmtList(stmt func() nodes.Stmt) (l []nodes.Stmt) {
	if trace {
		defer p.trace("stmtList")("")
	}

	for p.tok != nodes.EofT && p.tok != nodes.RbraceT && p.tok != nodes.CaseT && p.tok != nodes.DefaultT {
		s := stmt()
		if s == nil {
			break
		}
		l = append(l, s)

		// ";" is optional before "}"
		if !p.Got(nodes.SemiT) && p.tok != nodes.RbraceT {
			p.SyntaxError("at end of statement")
			p.Advance(nodes.SemiT, nodes.RbraceT, nodes.CaseT, nodes.DefaultT)
			p.Got(nodes.SemiT) // avoid spurious empty statement
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
func (p *parser) StmtOrNil() nodes.Stmt {
	if trace {
		defer p.trace("stmt " + p.tok.String())("")
	}

	// Most statements (assignments) start with an identifier;
	// look for it first before doing anything more expensive.
	if p.tok == nodes.NameT || p.TokIsKeywordName() {
		lhs := p.ExprList(false)
		if label, ok := lhs.(*nodes.Name); ok && p.tok == nodes.ColonT {
			return p.LabeledStmtOrNil(label)
		}
		return p.DynamicAssignOpOrSimpleStmt(lhs)
	}

	switch p.tok {
	case nodes.LbraceT:
		return p.BlockStmt("", p.StmtOrNil)

	case nodes.VarT:
		return p.DeclStmt(p.VarDecl)

	case nodes.ConstT:
		return p.DeclStmt(p.ConstDecl)

	case nodes.TypeT:
		return p.DeclStmt(p.TypeDecl)

	case nodes.OperatorT, nodes.StarT:
		switch p.op {
		case nodes.Add, nodes.Sub, nodes.Mul, nodes.And, nodes.Xor, nodes.Not:
			return p.SimpleStmt(nil, false) // unary operators
		}

	case nodes.LiteralT, nodes.FuncT, nodes.LparenT, // operands
		nodes.LbrackT, nodes.StructT, nodes.MapT, nodes.ChanT, nodes.InterfaceT, // composite types
		nodes.ArrowT: // receive operator
		return p.SimpleStmt(nil, false)

	case nodes.IfT:
		return p.IfStmt(p.StmtOrNil)

	case nodes.ForT:
		return p.ForStmt(p.StmtOrNil)

	case nodes.SwitchT:
		return p.SwitchStmt(p.StmtOrNil)

	case nodes.SelectT:
		return p.SelectStmt(p.StmtOrNil)

	case nodes.GoT, nodes.DeferT:
		return p.CallStmt(p.StmtOrNil)

	case nodes.SemiT:
		s := new(nodes.EmptyStmt)
		s.SetPos(p.Pos())
		return s

	case nodes.FallthroughT:
		if !p.checkPermit("fallthroughKw") {
			p.Advance(nodes.SemiT)
			return nil
		}
		s := new(nodes.BranchStmt)
		s.SetPos(p.Pos())
		p.CheckHashCmd(p.hash, func() {
			p.Next()
			s.Tok = nodes.FallthroughT
		})
		return s

	case nodes.BreakT, nodes.ContinueT:
		return p.BreakOrContinueStmt()

	case nodes.GotoT:
		return p.GotoStmt()

	case nodes.ReturnT:
		return p.ReturnStmt()
	}

	return nil
}

//--------------------------------------------------------------------------------
func (p *parser) DynamicAssignOpOrSimpleStmt(lhs nodes.Expr) nodes.Stmt {
	if _, ok := lhs.(*nodes.ListExpr); !ok && p.tok == nodes.AssignOpT {
		binOp := dynAssignOps[p.op]
		if p.dynamicBlock != "" && binOp != "" {
			p.Next()
			rhs := &nodes.RhsExpr{X: p.Expr()}
			t := &nodes.CallExpr{
				Fun: &nodes.SelectorExpr{
					X:   p.NewName(p.dynamicBlock),
					Sel: &nodes.Name{Value: binOp},
				},
				ArgList: []nodes.Expr{
					&nodes.Operation{
						Op: nodes.And,
						X:  lhs,
						Y:  nil,
					},
				},
			}
			t.SetPos(p.Pos())
			_ = p.ProcImportAlias(&nodes.BasicLit{Value: dynLib, Kind: nodes.StringLit}, p.dynamicBlock)
			t.ArgList = append(t.ArgList, rhs)
			return &nodes.ExprStmt{
				X: t,
			}
		}
	}
	return p.SimpleStmt(lhs, false)
}

var dynAssignOps = map[nodes.Operator]string{
	nodes.Mul:    "MultAssign",       // *=
	nodes.Div:    "DivideAssign",     // /=
	nodes.Rem:    "ModAssign",        // %=
	nodes.And:    "SeqAssign",        // &=
	nodes.AndNot: "SeqXorAssign",     // &^=
	nodes.Shl:    "LeftShiftAssign",  // <<=
	nodes.Shr:    "RightShiftAssign", // >>=
	nodes.LftAnd: "LeftSeqAssign",    // <&=
	nodes.AndRgt: "RightSeqAssign",   // &>=
	nodes.Add:    "PlusAssign",       // +=
	nodes.Sub:    "MinusAssign",      // -=
	nodes.Or:     "AltAssign",        // |=
	nodes.Xor:    "XorAssign",        // ^=
}

//--------------------------------------------------------------------------------
// SimpleStmt = EmptyStmt | ExpressionStmt | SendStmt | IncDecStmt | Assignment | ShortVarDecl .
func (p *parser) SimpleStmt(lhs nodes.Expr, rangeOk bool) nodes.SimpleStmt {
	if trace {
		defer p.trace("simpleStmt")("")
	}

	if rangeOk && p.tok == nodes.RangeT {
		// _Range expr
		if debug && lhs != nil {
			panic("invalid call of simpleStmt")
		}
		return p.NewRangeClause(nil, false)
	}

	if lhs == nil {
		lhs = p.ExprList(false)
	}

	if _, ok := lhs.(*nodes.ListExpr); !ok && p.tok != nodes.AssignT && p.tok != nodes.DefineT {
		// expr
		pos := p.Pos()
		switch p.tok {
		case nodes.AssignOpT:
			// lhs op= rhs
			op := p.op
			p.Next()
			rhs := p.Expr()
			if p.dynamicBlock != "" {
				rhs = &nodes.RhsExpr{X: rhs}
			}
			return p.NewAssignStmt(pos, op, lhs, rhs)

		case nodes.IncOpT:
			// lhs++ or lhs--
			op := p.op
			p.Next()
			if p.dynamicBlock != "" {
				lhs = &nodes.RhsExpr{X: lhs} //parsed as lhs but actually rhs
			}
			return p.NewAssignStmt(pos, op, lhs, nodes.ImplicitOne)

		case nodes.ArrowT:
			// lhs <- rhs
			s := new(nodes.SendStmt)
			s.SetPos(pos)
			p.Next()
			s.Chan = lhs
			rhs := p.Expr()
			if p.dynamicBlock != "" {
				rhs = &nodes.RhsExpr{X: rhs}
			}
			s.Value = rhs
			return s

		default:
			// expr
			s := new(nodes.ExprStmt)
			s.SetPos(lhs.Pos())
			if p.dynamicBlock != "" {
				lhs = &nodes.RhsExpr{X: lhs} //parsed as lhs but actually rhs
			}
			s.X = lhs
			s.MakeComments()
			if len(p.comments) > 0 {
				s.SetRightComment(strings.Join(p.comments, "\n"))
			}
			return s
		}
	}

	// expr_list
	pos := p.Pos()
	switch p.tok {
	case nodes.AssignT:
		p.Next()

		if rangeOk && p.tok == nodes.RangeT {
			// expr_list '=' RangeT expr
			return p.NewRangeClause(lhs, false)
		}

		// expr_list '=' expr_list
		return p.NewAssignStmt(pos, 0, lhs, p.ExprList(true))

	case nodes.DefineT:
		p.Next()

		if rangeOk && p.tok == nodes.RangeT {
			// expr_list ':=' range expr
			return p.NewRangeClause(lhs, true)
		}

		// expr_list ':=' expr_list
		rhs := p.ExprList(true)

		if x, ok := rhs.(*nodes.TypeSwitchGuard); ok {
			switch lhs := lhs.(type) {
			case *nodes.Name:
				x.Lhs = lhs
			case *nodes.ListExpr:
				p.ErrorAt(lhs.Pos(), fmt.Sprintf("cannot assign 1 value to %d variables", len(lhs.ElemList)))
				// make the best of what we have
				if lhs, ok := lhs.ElemList[0].(*nodes.Name); ok {
					x.Lhs = lhs
				}
			default:
				p.ErrorAt(lhs.Pos(), fmt.Sprintf("invalid variable name %s in type switch", String(lhs)))
			}
			s := new(nodes.ExprStmt)
			s.SetPos(x.Pos())
			s.X = x
			return s
		}

		as := p.NewAssignStmt(pos, nodes.Def, lhs, rhs)
		return as

	default:
		p.SyntaxError("expecting := or = or comma")
		p.Advance(nodes.SemiT, nodes.RbraceT)
		// make the best of what we have
		if x, ok := lhs.(*nodes.ListExpr); ok {
			lhs = x.ElemList[0]
		}
		s := new(nodes.ExprStmt)
		s.SetPos(lhs.Pos())
		s.X = lhs
		return s
	}
}

//--------------------------------------------------------------------------------
func (p *parser) NewRangeClause(lhs nodes.Expr, def bool) *nodes.RangeClause {
	r := new(nodes.RangeClause)
	r.SetPos(p.Pos())
	p.CheckHashCmd(p.hash, func() {
		p.Next() // consume _Range
		r.Lhs = lhs
		r.Def = def
		rhs := p.Expr()
		if p.dynamicBlock != "" {
			rhs = &nodes.RhsExpr{X: rhs}
		}
		r.X = rhs
	})
	return r
}

//--------------------------------------------------------------------------------
func (p *parser) NewAssignStmt(pos src.Pos, op nodes.Operator, lhs, rhs nodes.Expr) *nodes.AssignStmt {
	a := new(nodes.AssignStmt)
	a.SetPos(pos)
	a.Op = op
	a.Lhs = lhs
	a.Rhs = rhs
	a.MakeComments()
	if len(p.comments) > 0 {
		a.SetRightComment(strings.Join(p.comments, "\n"))
	}
	return a
}

//--------------------------------------------------------------------------------
func (p *parser) LabeledStmtOrNil(label *nodes.Name) nodes.Stmt {
	if trace {
		defer p.trace("labeledStmt")("")
	}

	s := new(nodes.LabeledStmt)
	s.SetPos(p.Pos())
	s.Label = label

	p.Want(nodes.ColonT)

	if p.tok == nodes.RbraceT {
		// We expect a statement (incl. an empty statement), which must be
		// terminated by a semicolon. Because semicolons may be omitted before
		// an _Rbrace, seeing an _Rbrace implies an empty statement.
		e := new(nodes.EmptyStmt)
		e.SetPos(p.Pos())
		s.Stmt = e
		return s
	}

	s.Stmt = p.StmtOrNil()
	if s.Stmt != nil {
		return s
	}

	// report error at line of ':' token
	p.SyntaxErrorAt(s.Pos(), "missing statement after label")
	// we are already at the end of the labeled statement - no need to advance
	return nil // avoids follow-on errors (see e.g., fixedbugs/bug274.go)
}

//--------------------------------------------------------------------------------
func (p *parser) DeclStmt(f func(*nodes.DeclGroup) nodes.Decl) *nodes.DeclStmt {
	if trace {
		defer p.trace("declStmt")("")
	}

	s := new(nodes.DeclStmt)
	s.SetPos(p.Pos())
	s.MakeComments()
	if len(p.comments) > 0 {
		s.SetAboveComment(strings.Join(p.comments, "\n"))
	}
	p.CheckHashCmd(p.hash, func() {
		p.Next() // ConstT, TypeT, or VarT
		s.DeclList = p.AppendGroup(nil, f)
	})

	return s
}

//--------------------------------------------------------------------------------
func (p *parser) ReturnStmt() *nodes.ReturnStmt {
	if !p.checkPermit("returnKw") {
		p.Advance(nodes.SemiT)
		return nil
	}
	s := new(nodes.ReturnStmt)
	s.SetPos(p.Pos())
	s.MakeComments()
	if len(p.comments) > 0 {
		s.SetAboveComment(strings.Join(p.comments, "\n"))
	}
	p.CheckHashCmd(p.hash, func() {
		p.Next()
		if p.tok != nodes.SemiT && p.tok != nodes.RbraceT {
			s.Results = p.ExprList(true)
		}
		if len(p.comments) > 0 {
			s.SetRightComment(strings.Join(p.comments, "\n"))
		}
	})
	return s
}

//--------------------------------------------------------------------------------
// breakOrContinueStmt parses 'break' or 'continue' statements
func (p *parser) BreakOrContinueStmt() *nodes.BranchStmt {
	if p.tok == nodes.BreakT && !p.checkPermit("breakKw") ||
		p.tok == nodes.ContinueT && !p.checkPermit("continueKw") {
		p.Advance(nodes.SemiT)
		return nil
	}
	s := new(nodes.BranchStmt)
	s.SetPos(p.Pos())
	s.Tok = p.tok
	s.MakeComments()
	if len(p.comments) > 0 {
		s.SetAboveComment(strings.Join(p.comments, "\n"))
	}
	p.CheckHashCmd(p.hash, func() {
		p.Next()
		if p.tok == nodes.NameT || p.TokIsKeywordName() {
			s.Label = p.Name()
		}
		if len(p.comments) > 0 {
			s.SetRightComment(strings.Join(p.comments, "\n"))
		}
	})
	return s
}

//--------------------------------------------------------------------------------
func (p *parser) GotoStmt() *nodes.BranchStmt {
	if !p.checkPermit("gotoKw") {
		p.Advance(nodes.SemiT)
		return nil
	}
	s := new(nodes.BranchStmt)
	s.SetPos(p.Pos())
	s.Tok = nodes.GotoT
	s.MakeComments()
	if len(p.comments) > 0 {
		s.SetAboveComment(strings.Join(p.comments, "\n"))
	}
	p.CheckHashCmd(p.hash, func() {
		p.Next()
		s.Label = p.Name()
		if len(p.comments) > 0 {
			s.SetRightComment(strings.Join(p.comments, "\n"))
		}
	})
	return s
}

//--------------------------------------------------------------------------------
// callStmt parses call-like statements that can be preceded by 'defer' and 'go'.
func (p *parser) CallStmt(stmt func() nodes.Stmt) *nodes.CallStmt {
	if trace {
		defer p.trace("callStmt")("")
	}

	if p.tok == nodes.DeferT && !p.checkPermit("deferKw") ||
		p.tok == nodes.GoT && !p.checkPermit("goKw") {
		p.Advance(nodes.SemiT)
		return nil
	}
	s := new(nodes.CallStmt)
	s.SetPos(p.Pos())
	s.Tok = p.tok // DeferT or GoT
	s.MakeComments()
	if len(p.comments) > 0 {
		s.SetAboveComment(strings.Join(p.comments, "\n"))
	}
	p.CheckHashCmd(p.hash, func() {
		p.Next()
		var x nodes.Expr

		if p.tok == nodes.LbraceT {
			t := new(nodes.FuncType)
			t.SetPos(p.Pos())
			t.ParamList = []*nodes.Field{}
			t.ResultList = nil
			f := new(nodes.FuncLit)
			f.SetPos(p.Pos())
			f.Type = t
			f.Body = p.BlockStmt("", stmt)
			e := new(nodes.CallExpr)
			e.SetPos(p.Pos())
			e.ArgList = []nodes.Expr{}
			e.Fun = f
			x = e
		} else {
			x = p.PExpr(p.tok == nodes.LparenT) // keep_parens so we can report error below
			if t := unparen(x); t != x {
				p.Error(fmt.Sprintf("expression in %s must not be parenthesized", s.Tok))
				// already progressed, no need to advance
				x = t
			}
		}
		cx, ok := x.(*nodes.CallExpr)
		if !ok {
			p.Error(fmt.Sprintf("expression in %s must be function call", s.Tok))
			// already progressed, no need to advance
			cx = new(nodes.CallExpr)
			cx.SetPos(x.Pos())
			cx.Fun = p.BadExpr()
		}
		s.Call = cx
	})
	return s
}

//--------------------------------------------------------------------------------
func (p *parser) IfStmt(stmt func() nodes.Stmt) *nodes.IfStmt {
	if trace {
		defer p.trace("ifStmt")("")
	}

	if !p.checkPermit("ifKw") {
		p.Advance(nodes.RbraceT)
		return nil
	}
	s := new(nodes.IfStmt)
	s.SetPos(p.Pos())

	s.MakeComments()
	if len(p.comments) > 0 {
		s.SetAboveComment(strings.Join(p.comments, "\n"))
	}
	p.CheckHashCmd(p.hash, func() {
		s.Init, s.Cond, _ = p.header(nodes.IfT)
		s.Then = p.BlockStmt("if clause", stmt)
		if !p.checkPermit("elseKw") {
			p.Advance(nodes.SemiT, nodes.RbraceT)
			s = nil
			return
		}
		if p.tok == nodes.ElseT {
			p.CheckHashCmd(p.hash, func() {
				if p.Got(nodes.ElseT) {
					switch p.tok {
					case nodes.IfT:
						s.Else = p.IfStmt(stmt)
					case nodes.SwitchT:
						body := new(nodes.BlockStmt)
						body.SetPos(p.Pos())
						body.List = []nodes.Stmt{p.SwitchStmt(stmt)}
						body.Rbrace = p.Pos()
						s.Else = body
					case nodes.ForT:
						body := new(nodes.BlockStmt)
						body.SetPos(p.Pos())
						body.List = []nodes.Stmt{p.ForStmt(stmt)}
						body.Rbrace = p.Pos()
						s.Else = body
					case nodes.SelectT:
						body := new(nodes.BlockStmt)
						body.SetPos(p.Pos())
						body.List = []nodes.Stmt{p.SelectStmt(stmt)}
						body.Rbrace = p.Pos()
						s.Else = body
					case nodes.GoT, nodes.DeferT:
						body := new(nodes.BlockStmt)
						body.SetPos(p.Pos())
						body.List = []nodes.Stmt{p.CallStmt(stmt)}
						body.Rbrace = p.Pos()
						s.Else = body
					case nodes.LbraceT:
						s.Else = p.BlockStmt("", stmt)
					default:
						p.SyntaxError("else must be followed by if,switch,for,select,go,defer or statement block")
						p.Advance(nodes.NameT, nodes.RbraceT)
					}
				}
			})
		}
	})

	return s
}

//--------------------------------------------------------------------------------
func (p *parser) header(keyword nodes.Token) (init nodes.SimpleStmt, cond nodes.Expr, post nodes.SimpleStmt) {
	p.Want(keyword)

	if p.tok == nodes.LbraceT {
		if keyword == nodes.IfT {
			p.SyntaxError("missing condition in if statement")
		}
		return
	}
	// p.tok != _Lbrace

	outer := p.xnest
	p.xnest = -1

	if p.tok != nodes.SemiT {
		// accept potential varDecl but complain
		if p.Got(nodes.VarT) {
			p.SyntaxError(fmt.Sprintf("var declaration not allowed in %s initializer", keyword.String()))
		}
		init = p.SimpleStmt(nil, keyword == nodes.ForT)
		// If we have a range clause, we are done (can only happen for keyword == _For).
		if _, ok := init.(*nodes.RangeClause); ok {
			if !p.checkPermit("rangeKw") {
				p.Advance(nodes.SemiT, nodes.RbraceT)
				return
			}
			p.xnest = outer
			return
		}
	}

	var condStmt nodes.SimpleStmt
	var semi struct {
		pos src.Pos
		lit string // valid if pos.IsKnown()
	}
	if p.tok == nodes.SemiT {
		semi.pos = p.Pos()
		semi.lit = p.lit
		p.Next()
		if keyword == nodes.ForT {
			if p.tok != nodes.SemiT {
				if p.tok == nodes.LbraceT {
					p.SyntaxError("expecting for loop condition")
					goto done
				}
				condStmt = p.SimpleStmt(nil, false)
			}
			p.Want(nodes.SemiT)
			if p.tok != nodes.LbraceT {
				post = p.SimpleStmt(nil, false)
				if a, _ := post.(*nodes.AssignStmt); a != nil && a.Op == nodes.Def {
					p.SyntaxErrorAt(a.Pos(), "cannot declare in post statement of for loop")
				}
			}
		} else if p.tok != nodes.LbraceT {
			condStmt = p.SimpleStmt(nil, false)
		}
	} else {
		condStmt = init
		init = nil
	}

done:
	// unpack condStmt
	switch s := condStmt.(type) {
	case nil:
		if keyword == nodes.IfT && semi.pos.IsKnown() {
			if semi.lit != "semicolon" {
				p.SyntaxErrorAt(semi.pos, fmt.Sprintf("unexpected %s, expecting { after if clause", semi.lit))
			} else {
				p.SyntaxErrorAt(semi.pos, "missing condition in if statement")
			}
		}
	case *nodes.ExprStmt:
		cond = s.X
	default:
		p.SyntaxError(fmt.Sprintf("%s used as value", String(s)))
	}

	p.xnest = outer
	return
}

//--------------------------------------------------------------------------------
func (p *parser) ForStmt(stmt func() nodes.Stmt) nodes.Stmt {
	if trace {
		defer p.trace("forStmt")("")
	}

	if !p.checkPermit("forKw") {
		p.Advance(nodes.RbraceT)
		return nil
	}
	s := new(nodes.ForStmt)
	s.SetPos(p.Pos())

	s.MakeComments()
	if len(p.comments) > 0 {
		s.SetAboveComment(strings.Join(p.comments, "\n"))
	}
	p.CheckHashCmd(p.hash, func() {
		s.Init, s.Cond, s.Post = p.header(nodes.ForT)
		s.Body = p.BlockStmt("for clause", stmt)
	})

	return s
}

//--------------------------------------------------------------------------------
func (p *parser) SwitchStmt(stmt func() nodes.Stmt) *nodes.SwitchStmt {
	if trace {
		defer p.trace("switchStmt")("")
	}

	if !p.checkPermit("switchKw") {
		p.Advance(nodes.RbraceT)
		return nil
	}
	s := new(nodes.SwitchStmt)
	s.SetPos(p.Pos())

	s.MakeComments()
	if len(p.comments) > 0 {
		s.SetAboveComment(strings.Join(p.comments, "\n"))
	}
	p.CheckHashCmd(p.hash, func() {
		s.Init, s.Tag, _ = p.header(nodes.SwitchT)

		if !p.Got(nodes.LbraceT) {
			p.SyntaxError("missing { after switch clause")
			p.Advance(nodes.CaseT, nodes.DefaultT, nodes.RbraceT)
		}
		for p.tok != nodes.EofT && p.tok != nodes.RbraceT {
			s.Body = append(s.Body, p.CaseClause(stmt))
		}
		if len(s.Body) > 0 {
			s.Body[len(s.Body)-1].Final = true
		}
		s.Rbrace = p.Pos()
		p.Want(nodes.RbraceT)
	})

	return s
}

//--------------------------------------------------------------------------------
func (p *parser) CaseClause(stmt func() nodes.Stmt) *nodes.CaseClause {
	if trace {
		defer p.trace("caseClause")("")
	}

	c := new(nodes.CaseClause)
	c.SetPos(p.Pos())

	if p.tok == nodes.CaseT && !p.checkPermit("caseKw") ||
		p.tok == nodes.DefaultT && !p.checkPermit("defaultKw") {
		p.Advance(nodes.SemiT)
		return nil
	}
	p.CheckHashCmd(p.hash, func() {
		switch p.tok {
		case nodes.CaseT:
			p.Next()
			c.Cases = p.ExprList(true)

		case nodes.DefaultT:
			p.Next()

		default:
			p.SyntaxError("expecting case or default or }")
			p.Advance(nodes.ColonT, nodes.CaseT, nodes.DefaultT, nodes.RbraceT)
		}

		c.Colon = p.Pos()
		p.Want(nodes.ColonT)
		c.Body = p.StmtList(stmt)
	})

	return c
}

//--------------------------------------------------------------------------------
func (p *parser) SelectStmt(stmt func() nodes.Stmt) *nodes.SelectStmt {
	if trace {
		defer p.trace("selectStmt")("")
	}

	if !p.checkPermit("selectKw") {
		p.Advance(nodes.RbraceT)
		return nil
	}
	s := new(nodes.SelectStmt)
	s.SetPos(p.Pos())

	s.MakeComments()
	if len(p.comments) > 0 {
		s.SetAboveComment(strings.Join(p.comments, "\n"))
	}
	p.CheckHashCmd(p.hash, func() {
		p.Want(nodes.SelectT)
		if !p.Got(nodes.LbraceT) {
			p.SyntaxError("missing { after select clause")
			p.Advance(nodes.CaseT, nodes.DefaultT, nodes.RbraceT)
		}
		for p.tok != nodes.EofT && p.tok != nodes.RbraceT {
			s.Body = append(s.Body, p.CommClause(stmt))
		}
		if len(s.Body) > 0 {
			s.Body[len(s.Body)-1].Final = true
		}
		s.Rbrace = p.Pos()
		p.Want(nodes.RbraceT)
	})

	return s
}

//--------------------------------------------------------------------------------
func (p *parser) CommClause(stmt func() nodes.Stmt) *nodes.CommClause {
	if trace {
		defer p.trace("commClause")("")
	}

	c := new(nodes.CommClause)
	c.SetPos(p.Pos())

	if p.tok == nodes.CaseT && !p.checkPermit("caseKw") ||
		p.tok == nodes.DefaultT && !p.checkPermit("defaultKw") {
		p.Advance(nodes.SemiT)
		return nil
	}
	p.CheckHashCmd(p.hash, func() {
		switch p.tok {
		case nodes.CaseT:
			p.Next()
			c.Comm = p.SimpleStmt(nil, false)

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

		case nodes.DefaultT:
			p.Next()

		default:
			p.SyntaxError("expecting case or default or }")
			p.Advance(nodes.ColonT, nodes.CaseT, nodes.DefaultT, nodes.RbraceT)
		}

		c.Colon = p.Pos()
		p.Want(nodes.ColonT)
		c.Body = p.StmtList(stmt)
	})

	return c
}

//--------------------------------------------------------------------------------
// Expressions
//--------------------------------------------------------------------------------
// ExpressionList = Expression { "," Expression } .
func (p *parser) ExprList(inRhs bool) nodes.Expr {
	if trace {
		defer p.trace("exprList")("")
	}

	dynRhs := inRhs && p.dynamicBlock != ""
	x := p.Expr()
	if dynRhs {
		x = &nodes.RhsExpr{X: x}
	}
	if p.Got(nodes.CommaT) {
		y := p.Expr()
		if dynRhs {
			y = &nodes.RhsExpr{X: y}
		}
		list := []nodes.Expr{x, y}
		for p.Got(nodes.CommaT) {
			z := p.Expr()
			if dynRhs {
				z = &nodes.RhsExpr{X: z}
			}
			list = append(list, z)
		}
		t := new(nodes.ListExpr)
		t.SetPos(x.Pos())
		t.ElemList = list
		x = t
	}
	return x
}

//--------------------------------------------------------------------------------
func (p *parser) Expr() nodes.Expr {
	if trace {
		defer p.trace("expr")("")
	}

	return p.BinaryExpr(0)
}

//--------------------------------------------------------------------------------
// Expression = UnaryExpr | Expression binary_op Expression .
func (p *parser) BinaryExpr(prec nodes.Prec) nodes.Expr {
	// don't trace binaryExpr - only leads to overly nested trace output
	x := p.UnaryExpr()
	for (p.tok == nodes.OperatorT || p.tok == nodes.StarT) && p.prec > prec {
		op := p.op
		binOp := dynBinOps[op]
		if p.dynamicBlock == "" || binOp == "" {
			t := new(nodes.Operation)
			t.SetPos(p.Pos())
			t.Op = p.op
			t.X = x
			tprec := p.prec
			p.Next()
			t.Y = p.BinaryExpr(tprec)
			x = t
		} else {
			t := &nodes.CallExpr{
				Fun: &nodes.SelectorExpr{
					X:   p.NewName(p.dynamicBlock),
					Sel: &nodes.Name{Value: binOp},
				},
				ArgList: []nodes.Expr{x},
			}
			t.SetPos(p.Pos())
			_ = p.ProcImportAlias(&nodes.BasicLit{Value: dynLib, Kind: nodes.StringLit}, p.dynamicBlock)
			tprec := p.prec
			p.Next()
			expr := p.BinaryExpr(tprec)
			if op == nodes.AndAnd || op == nodes.OrOr {
				expr = &nodes.FuncLit{
					Type: &nodes.FuncType{
						ParamList:  []*nodes.Field{},
						ResultList: []*nodes.Field{&nodes.Field{Type: &nodes.InterfaceType{MethodList: []*nodes.Field{}}}},
					},
					Body: &nodes.BlockStmt{
						List: []nodes.Stmt{&nodes.ReturnStmt{Results: expr}},
					},
				}
			}
			t.ArgList = append(t.ArgList, expr)
			x = t
		}
	}
	return x
}

var dynBinOps = map[nodes.Operator]string{
	nodes.Mul:    "Mult",             // * //multiplicative prec
	nodes.Div:    "Divide",           // /
	nodes.Rem:    "Mod",              // %
	nodes.And:    "Seq",              // &
	nodes.AndNot: "SeqXor",           // &^
	nodes.Shl:    "LeftShift",        // <<
	nodes.Shr:    "RightShift",       // >>
	nodes.LftAnd: "LeftSeq",          // <&
	nodes.AndRgt: "RightSeq",         // &>
	nodes.Add:    "Plus",             // + //additive prec
	nodes.Sub:    "Minus",            // -
	nodes.Or:     "Alt",              // |
	nodes.Xor:    "Xor",              // ^
	nodes.Eql:    "IsEqual",          // == //comparative prec
	nodes.Neq:    "IsNotEqual",       // !=
	nodes.Lss:    "IsLessThan",       // <
	nodes.Leq:    "IsLessOrEqual",    // <=
	nodes.Gtr:    "IsGreaterThan",    // >
	nodes.Geq:    "IsGreaterOrEqual", // >=
	nodes.AndAnd: "And",              // && //boolean precs
	nodes.OrOr:   "Or",               // ||
}

//TODO: "Matcher", "FindAll", "FindFirst", "Group", "Parenthesize"

//--------------------------------------------------------------------------------
// UnaryExpr = PrimaryExpr | unary_op UnaryExpr .
func (p *parser) UnaryExpr() nodes.Expr {
	if trace {
		defer p.trace("unaryExpr")("")
	}

	switch p.tok {
	case nodes.OperatorT, nodes.StarT:
		switch p.op {
		//nodes.Or used as the dynamic unary "reflect" operator...
		case nodes.Mul, nodes.Add, nodes.Sub, nodes.Not, nodes.Xor, nodes.Or:
			unaryOp := dynUnaryOps[p.op]
			if p.dynamicBlock == "" || unaryOp == "" {
				x := new(nodes.Operation)
				x.SetPos(p.Pos())
				x.Op = p.op
				p.Next()
				x.X = p.UnaryExpr()
				return x
			} else {
				t := &nodes.CallExpr{
					Fun: &nodes.SelectorExpr{
						X:   p.NewName(p.dynamicBlock),
						Sel: &nodes.Name{Value: unaryOp},
					},
					ArgList: []nodes.Expr{},
				}
				t.SetPos(p.Pos())
				p.Next()
				t.ArgList = append(t.ArgList, p.UnaryExpr())
				_ = p.ProcImportAlias(&nodes.BasicLit{Value: dynLib, Kind: nodes.StringLit}, p.dynamicBlock)
				return t
			}

		case nodes.And:
			x := new(nodes.Operation)
			x.SetPos(p.Pos())
			x.Op = nodes.And
			p.Next()
			// unaryExpr may have returned a parenthesized composite literal
			// (see comment in operand) - remove parentheses if any
			x.X = unparen(p.UnaryExpr())
			return x
		}

	case nodes.ArrowT:
		// receive op (<-x) or receive-only channel (<-chan E)
		pos := p.Pos()
		p.Next()

		// If the next token is _Chan we still don't know if it is
		// a channel (<-chan int) or a receive op (<-chan int(ch)).
		// We only know once we have found the end of the unaryExpr.

		x := p.UnaryExpr()

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

		if _, ok := x.(*nodes.ChanType); ok {
			// x is a channel type => re-associate <-
			dir := nodes.SendOnly
			t := x
			for dir == nodes.SendOnly {
				c, ok := t.(*nodes.ChanType)
				if !ok {
					break
				}
				dir = c.Dir
				if dir == nodes.RecvOnly {
					// t is type <-chan E but <-<-chan E is not permitted
					// (report same error as for "type _ <-<-chan E")
					p.SyntaxError("unexpected <-, expecting chan")
					// already progressed, no need to advance
				}
				c.Dir = nodes.RecvOnly
				t = c.Elem
			}
			if dir == nodes.SendOnly {
				// channel dir is <- but channel element E is not a channel
				// (report same error as for "type _ <-chan<-E")
				p.SyntaxError(fmt.Sprintf("unexpected %s, expecting chan", String(t)))
				// already progressed, no need to advance
			}
			return x
		}

		// x is not a channel type => we have a receive op
		o := new(nodes.Operation)
		o.SetPos(pos)
		o.Op = nodes.Recv
		o.X = x
		return o
	}

	// TODO(mdempsky): We need parens here so we can report an
	// error for "(x) := true". It should be possible to detect
	// and reject that more efficiently though.
	return p.PExpr(true)
}

var dynUnaryOps = map[nodes.Operator]string{
	nodes.Add: "Identity",   // +
	nodes.Sub: "Negate",     // -
	nodes.Not: "Not",        // !
	nodes.Xor: "Complement", // ^
	nodes.Or:  "Reflect",    // |
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
func (p *parser) PExpr(keep_parens bool) nodes.Expr {
	if trace {
		defer p.trace("pExpr")("")
	}

	x := p.Operand(keep_parens)

loop:
	for {
		pos := p.Pos()
		switch p.tok {
		case nodes.DotT:
			p.Next()
			if p.tok == nodes.NameT || p.TokIsKeywordName() {
				// pexpr '.' sym
				t := new(nodes.SelectorExpr)
				t.SetPos(pos)
				t.X = x
				t.Sel = p.Name()
				x = t
			} else {
				switch p.tok {
				case nodes.LparenT:
					p.Next()
					if p.Got(nodes.TypeT) {
						t := new(nodes.TypeSwitchGuard)
						t.SetPos(pos)
						t.X = x
						x = t
					} else {
						t := new(nodes.AssertExpr)
						t.SetPos(pos)
						t.X = x
						t.Type = p.Expr()
						x = t
					}
					p.Want(nodes.RparenT)

				default:
					p.SyntaxError("expecting name or (")
					p.Advance(nodes.SemiT, nodes.RparenT)
				}
			}

		case nodes.LbrackT:
			if p.dynamicBlock == "" {
				x = p.IndexExpr(x)
			} else {
				x = p.DynamicIndexing(x)
			}

		case nodes.LparenT:
			//x = p.call(x)
			t := new(nodes.CallExpr)
			t.SetPos(pos)
			t.Fun = x
			t.ArgList, t.HasDots = p.ArgList()
			x = t

		case nodes.LbraceT:
			// operand may have returned a parenthesized complit
			// type; accept it but complain if we have a complit
			t := unparen(x)
			// determine if '{' belongs to a composite literal or a block statement
			complit_ok := false
			switch t.(type) {
			case *nodes.Name, *nodes.SelectorExpr:
				if p.xnest >= 0 {
					// x is considered a composite literal type
					complit_ok = true
				}
			case *nodes.ArrayType, *nodes.SliceType, *nodes.StructType, *nodes.MapType:
				// x is a comptype
				complit_ok = true
			}
			if !complit_ok {
				break loop
			}
			if t != x {
				p.SyntaxError("cannot parenthesize type in composite literal")
				// already progressed, no need to advance
			}
			if u, ok := t.(*nodes.MapType); ok && u.Dyn {
				x = p.DynMapExpr()
			} else {
				n := p.CompLitExpr()
				n.Type = x
				x = n
			}

		default:
			break loop
		}
	}

	return x
}

//--------------------------------------------------------------------------------
// LiteralValue = "{" [ ElementList [ "," ] ] "}" .
func (p *parser) DynMapExpr() *nodes.CallExpr {
	if trace {
		defer p.trace("dynMapExpr")("")
	}

	x := &nodes.CallExpr{
		Fun: &nodes.SelectorExpr{
			X:   p.NewName(p.dynamicBlock),
			Sel: &nodes.Name{Value: "InitMap"},
		},
	}
	_ = p.ProcImportAlias(&nodes.BasicLit{Value: dynLib, Kind: nodes.StringLit}, p.dynamicBlock)

	x.SetPos(p.Pos())
	p.xnest++
	p.List(nodes.LbraceT, nodes.CommaT, nodes.RbraceT, func() bool {
		e := &nodes.CallExpr{
			Fun: &nodes.SelectorExpr{
				X:   p.NewName(p.dynamicBlock),
				Sel: &nodes.Name{Value: "NewPair"},
			},
		}
		e.ArgList = append(e.ArgList, p.BareCompLitExpr())
		p.Want(nodes.ColonT)
		e.ArgList = append(e.ArgList, p.BareCompLitExpr())
		x.ArgList = append(x.ArgList, e)
		return false
	})
	p.xnest--
	return x
}

//--------------------------------------------------------------------------------
// Operand     = Literal | OperandName | MethodExpr | "(" Expression ")" .
// Literal     = BasicLit | CompositeLit | FunctionLit .
// BasicLit    = int_lit | float_lit | imaginary_lit | rune_lit | string_lit .
// OperandName = identifier | QualifiedIdent.
func (p *parser) Operand(keep_parens bool) nodes.Expr {
	if trace {
		defer p.trace("operand " + p.tok.String())("")
	}

	if p.tok == nodes.NameT || p.TokIsKeywordName() {
		return p.Name()
	} else {
		switch p.tok {
		case nodes.LiteralT:
			lit := p.OLiteral()
			if lit.Kind == nodes.StringLit && p.tok == nodes.DotT {
				if !p.checkPermit("inplaceImps") {
					p.Advance(nodes.SemiT)
					return nil
				}
				a := p.ProcImportAlias(lit, "")
				lit.Value = a
				return lit
			}
			if p.dynamicBlock != "" {
				switch lit.Kind {
				case nodes.StringLit:
					t := &nodes.CallExpr{
						Fun: &nodes.SelectorExpr{
							X:   p.NewName(p.dynamicBlock),
							Sel: &nodes.Name{Value: "MakeText"},
						},
						ArgList: []nodes.Expr{lit},
					}
					t.SetPos(p.Pos())
					_ = p.ProcImportAlias(&nodes.BasicLit{Value: dynLib, Kind: nodes.StringLit}, p.dynamicBlock)
					return t
				case nodes.RuneLit:
					t := &nodes.CallExpr{
						Fun: &nodes.SelectorExpr{
							X:   p.NewName(p.dynamicBlock),
							Sel: &nodes.Name{Value: "Runex"},
						},
						ArgList: []nodes.Expr{&nodes.BasicLit{Value: strconv.Quote(strings.Trim(lit.Value, "'")), Kind: nodes.StringLit}},
					}
					t.SetPos(p.Pos())
					_ = p.ProcImportAlias(&nodes.BasicLit{Value: dynLib, Kind: nodes.StringLit}, p.dynamicBlock)
					return t
				case nodes.IntLit, nodes.FloatLit, nodes.ImagLit:
					lit.Value = strings.Replace(lit.Value, "_", "", -1)
					return lit
				case nodes.DateLit:
					lit.Value = strings.Replace(lit.Value, "_", "", -1)
					vals := strings.Split(lit.Value, ".")
					if len(vals) != 3 {
						p.SyntaxError("invalid date format")
					}
					t := &nodes.CallExpr{
						Fun: &nodes.SelectorExpr{
							X:   p.NewName("time"),
							Sel: &nodes.Name{Value: "Date"},
						},
						ArgList: []nodes.Expr{
							&nodes.BasicLit{Value: vals[0], Kind: nodes.IntLit},
							&nodes.BasicLit{Value: vals[1], Kind: nodes.IntLit},
							&nodes.BasicLit{Value: vals[2], Kind: nodes.IntLit},
							&nodes.BasicLit{Value: "0", Kind: nodes.IntLit},
							&nodes.BasicLit{Value: "0", Kind: nodes.IntLit},
							&nodes.BasicLit{Value: "0", Kind: nodes.IntLit},
							&nodes.BasicLit{Value: "0", Kind: nodes.IntLit},
							&nodes.SelectorExpr{
								X:   p.NewName("time"),
								Sel: &nodes.Name{Value: "UTC"},
							},
						},
					}
					t.SetPos(p.Pos())
					_ = p.ProcImportAlias(&nodes.BasicLit{Value: dynLib, Kind: nodes.StringLit}, p.dynamicBlock)
					_ = p.ProcImportAlias(&nodes.BasicLit{Value: "\"time\"", Kind: nodes.StringLit}, "time")
					return t
				}
			}
			return lit

		case nodes.LparenT:
			pos := p.Pos()
			p.Next()
			p.xnest++
			x := p.Expr()
			p.xnest--
			p.Want(nodes.RparenT)

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
			if p.tok == nodes.LbraceT {
				keep_parens = true
			}

			// Parentheses are also not permitted around the expression
			// in a go/defer statement. In that case, operand is called
			// with keep_parens set.
			if keep_parens {
				px := new(nodes.ParenExpr)
				px.SetPos(pos)
				px.X = x
				x = px
			}
			return x

		case nodes.FuncT:
			pos := p.Pos()
			if !p.checkPermit("funcKw") {
				p.Advance(nodes.SemiT)
				return nil
			}
			p.Next()
			if p.tok == nodes.LparenT {
				t := p.FuncType()
				if p.tok == nodes.LbraceT {
					p.xnest++

					f := new(nodes.FuncLit)
					f.SetPos(pos)
					f.Type = t
					/*f.Body = p.blockStmt("", p.stmtOrNil)
					if p.mode&CheckBranches != 0 {
						checkBranches(f.Body, p.errh)
					}*/
					f.Body = p.FuncBody(p.StmtOrNil)

					p.xnest--
					return f
				}
				return t

			} else if p.dynamicBlock != "" { //short-form function
				t := &nodes.FuncType{
					ParamList: []*nodes.Field{{
						Name: p.NewName("groo_it"),
						Type: &nodes.DotsType{Elem: &nodes.InterfaceType{MethodList: []*nodes.Field{}}},
					}},
					ResultList: []*nodes.Field{
						{Type: &nodes.InterfaceType{MethodList: []*nodes.Field{}}},
					},
				}
				t.SetPos(p.Pos())
				if p.tok == nodes.LbraceT {
					p.xnest++
					f := new(nodes.FuncLit)
					f.SetPos(pos)
					f.Type = t
					f.ShortForm = true
					f.Body = p.FuncBody(p.StmtOrNil)
					p.xnest--
					if len(f.Body.List) > 0 {
						lastStmt := f.Body.List[len(f.Body.List)-1]
						switch lastStmt := lastStmt.(type) {
						case *nodes.ExprStmt:
							f.Body.List[len(f.Body.List)-1] = &nodes.ReturnStmt{Results: lastStmt.X}
						//TODO: case nodes.LabeledStmt, etc
						default:
						}
					}
					return f
				}
				return t
			} else {
				p.SyntaxError("invalid func literal syntax")
				return nil
			}

		case nodes.LbrackT, nodes.ChanT, nodes.MapT, nodes.StructT, nodes.InterfaceT:
			return p.Type() // othertype

		default:
			x := p.BadExpr()
			p.SyntaxError("expecting expression")
			p.Advance()
			return x
		}
	}

	// Syntactically, composite literals are operands. Because a complit
	// type may be a qualified identifier which is handled by pexpr
	// (together with selector expressions), complits are parsed there
	// as well (operand is only called from pexpr).
}

//--------------------------------------------------------------------------------
// FunctionBody = Block .
func (p *parser) FuncBody(stmt func() nodes.Stmt) *nodes.BlockStmt {
	if trace {
		defer p.trace("funcBody")("")
	}

	p.fnest++
	errcnt := p.errcnt
	body := p.BlockStmt("", stmt)
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
// Arguments = "(" [ ( ExpressionList | Type [ "," ExpressionList ] ) [ "..." ] [ "," ] ] ")" .
//func (p *parser) call(fun Expr) *CallExpr {
func (p *parser) ArgList() (list []nodes.Expr, hasDots bool) {
	if trace {
		defer p.trace("argList")("")
	}

	p.xnest++
	p.List(nodes.LparenT, nodes.CommaT, nodes.RparenT, func() bool {
		list = append(list, p.Expr())
		hasDots = p.Got(nodes.DotDotDotT)
		return hasDots
	})
	p.xnest--
	return
}

//--------------------------------------------------------------------------------
// LiteralValue = "{" [ ElementList [ "," ] ] "}" .
func (p *parser) CompLitExpr() *nodes.CompositeLit {
	if trace {
		defer p.trace("compLitExpr")("")
	}

	x := new(nodes.CompositeLit)
	x.SetPos(p.Pos())
	p.xnest++
	x.Rbrace = p.List(nodes.LbraceT, nodes.CommaT, nodes.RbraceT, func() bool {
		// value
		e := p.BareCompLitExpr()
		if p.tok == nodes.ColonT {
			// key ':' value
			l := new(nodes.KeyValueExpr)
			l.SetPos(p.Pos())
			p.Next()
			l.Key = e
			l.Value = p.BareCompLitExpr()
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
// Element = Expression | LiteralValue .
func (p *parser) BareCompLitExpr() nodes.Expr {
	if trace {
		defer p.trace("bareCompLitExpr")("")
	}

	if p.tok == nodes.LbraceT {
		// '{' start_complit braced_keyval_list '}'
		return p.CompLitExpr()
	}

	return p.Expr()
}

//--------------------------------------------------------------------------------
func (p *parser) IndexExpr(x nodes.Expr) nodes.Expr {
	pos := p.Pos()
	p.Next()
	p.xnest++

	var i nodes.Expr
	if p.tok != nodes.ColonT {
		i = p.Expr()
		if p.tok == nodes.RbrackT {
			// x[i]
			t := new(nodes.IndexExpr)
			t.SetPos(pos)
			t.X = x
			t.Index = i
			x = t
			p.Want(nodes.RbrackT)
			p.xnest--
			return x
		}
	}

	// x[i:...
	t := new(nodes.SliceExpr)
	t.SetPos(pos)
	t.X = x
	t.Index[0] = i
	p.Want(nodes.ColonT)
	if p.tok != nodes.ColonT && p.tok != nodes.RbrackT {
		// x[i:j...
		t.Index[1] = p.Expr()
	}
	if p.Got(nodes.ColonT) {
		t.Full = true
		// x[i:j:...]
		if t.Index[1] == nil {
			p.Error("middle index required in 3-index slice")
		}
		if p.tok != nodes.RbrackT {
			// x[i:j:k...
			t.Index[2] = p.Expr()
		} else {
			p.Error("final index required in 3-index slice")
		}
	}
	x = t
	p.Want(nodes.RbrackT)
	p.xnest--
	return x
}

//--------------------------------------------------------------------------------
func (p *parser) DynamicIndexing(x nodes.Expr) nodes.Expr { //TODO: mustn't be lhs
	pos := p.Pos()
	p.Next()
	p.xnest++

	var i nodes.Expr
	if p.tok != nodes.ColonT {
		i = p.Expr()
		if p.tok == nodes.RbrackT {
			// x[i]

			t := &nodes.ParenExpr{X: &nodes.Operation{ //work around bug in printer.go ?
				Op: nodes.Mul,
				X: &nodes.CallExpr{
					Fun: &nodes.SelectorExpr{
						X:   p.NewName(p.dynamicBlock),
						Sel: &nodes.Name{Value: "GetIndex"},
					},
					ArgList: []nodes.Expr{
						&nodes.Operation{Op: nodes.And, X: x},
						i,
					},
				}},
			}
			t.SetPos(pos)
			x = t

			_ = p.ProcImportAlias(&nodes.BasicLit{Value: dynLib, Kind: nodes.StringLit}, p.dynamicBlock)
			p.Want(nodes.RbrackT)
			p.xnest--
			return x
		}
	}

	// x[i:...
	p.Want(nodes.ColonT)
	var j nodes.Expr
	if p.tok != nodes.RbrackT {
		// x[i:j...
		j = p.Expr()
	} else {
		// x[i:...
		j = &nodes.SelectorExpr{
			X:   p.NewName(p.dynamicBlock),
			Sel: &nodes.Name{Value: "Inf"},
		}
	}
	if p.Got(nodes.ColonT) {
		p.Error("no more than 2 indexes can be used in a dynamic slice index")
	}
	t := &nodes.ParenExpr{
		X: &nodes.Operation{ //work around bug in printer.go ?
			Op: nodes.Mul,
			X: &nodes.CallExpr{
				Fun: &nodes.SelectorExpr{
					X:   p.NewName(p.dynamicBlock),
					Sel: &nodes.Name{Value: "GetIndex"},
				},
				ArgList: []nodes.Expr{
					&nodes.Operation{Op: nodes.And, X: x},
					i,
					j,
				},
			},
		},
	}
	t.SetPos(pos)
	x = t
	p.Want(nodes.RbrackT)
	p.xnest--
	return x
}

//--------------------------------------------------------------------------------
func (p *parser) BadExpr() *nodes.BadExpr {
	b := new(nodes.BadExpr)
	b.SetPos(p.Pos())
	return b
}

//--------------------------------------------------------------------------------
// Types
//--------------------------------------------------------------------------------
func (p *parser) Type() nodes.Expr {
	if trace {
		defer p.trace("type")("")
	}

	typ := p.TypeOrNil()
	if typ == nil {
		typ = p.BadExpr()
		p.SyntaxError("expecting type")
		p.Advance()
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
func (p *parser) TypeOrNil() nodes.Expr {
	if trace {
		defer p.trace("typeOrNil")("")
	}

	pos := p.Pos()
	if p.tok == nodes.NameT || p.TokIsKeywordName() {
		return p.DotName(p.Name())
	} else {
		switch p.tok {
		case nodes.StarT:
			// ptrtype
			p.Next()
			return newIndirect(pos, p.Type())

		case nodes.ArrowT:
			// recvchantype
			p.Next()
			t := new(nodes.ChanType)
			p.CheckHashCmd(p.hash, func() {
				p.Want(nodes.ChanT)
				t.SetPos(pos)
				t.Dir = nodes.RecvOnly
				t.Elem = p.ChanElem()
			})
			return t

		case nodes.FuncT:
			// fntype
			p.Next()
			return p.FuncType()

		case nodes.LbrackT:
			// '[' oexpr ']' ntype
			// '[' _DotDotDot ']' ntype
			// and if dynamic mode, could be [m:n] or [:n] or [m:] or [:]
			p.Next()
			p.xnest++
			if p.Got(nodes.RbrackT) {
				// []T
				p.xnest--
				t := new(nodes.SliceType)
				t.SetPos(pos)
				if p.dynamicBlock != "" && p.tok == nodes.LbraceT {
					//short slice syntax in literals: "[]" instead of "[T]"
					t.Elem = p.NewName("any")
				} else {
					t.Elem = p.Type()
				}
				return t
			}

			// [n]T or [...]T
			// or, in dynamic mode, [m:n] or [:n] or [m:] or [:]
			var fromval nodes.Expr
			if p.dynamicBlock != "" && p.Got(nodes.ColonT) {
				//[:n] or [:]
				fromval = &nodes.BasicLit{
					Value: "0",
					Kind:  nodes.IntLit,
				}
				var toval nodes.Expr
				if p.Got(nodes.RbrackT) {
					toval = &nodes.SelectorExpr{
						X:   p.NewName(p.dynamicBlock),
						Sel: &nodes.Name{Value: "Inf"},
					}
				} else {
					toval = p.Expr()
					p.Want(nodes.RbrackT)
				}
				p.xnest--
				t := &nodes.CallExpr{ //prev: CompositeLit
					Fun: &nodes.SelectorExpr{ //prev: Type
						X:   p.NewName(p.dynamicBlock),
						Sel: &nodes.Name{Value: "NewPair"},
					},
					ArgList: []nodes.Expr{fromval, toval}, //prev: ElemList
				}
				_ = p.ProcImportAlias(&nodes.BasicLit{Value: dynLib, Kind: nodes.StringLit}, p.dynamicBlock)
				return t
			}
			if !p.Got(nodes.DotDotDotT) {
				fromval = p.Expr()
				if p.dynamicBlock != "" {
					//[m:] or [m:n]
					p.Want(nodes.ColonT)
					var toval nodes.Expr
					if p.tok != nodes.RbrackT {
						toval = p.Expr()
					} else {
						toval = &nodes.SelectorExpr{
							X:   p.NewName(p.dynamicBlock),
							Sel: &nodes.Name{Value: "Inf"},
						}
					}
					p.Want(nodes.RbrackT)
					p.xnest--
					t := &nodes.CallExpr{
						Fun: &nodes.SelectorExpr{
							X:   p.NewName(p.dynamicBlock),
							Sel: &nodes.Name{Value: "NewPair"},
						},
						ArgList: []nodes.Expr{fromval, toval},
					}
					_ = p.ProcImportAlias(&nodes.BasicLit{Value: dynLib, Kind: nodes.StringLit}, p.dynamicBlock)
					return t
				}
			}
			p.Want(nodes.RbrackT)
			p.xnest--
			t := new(nodes.ArrayType)
			t.SetPos(pos)
			t.Len = fromval
			t.Elem = p.Type()
			return t

		case nodes.ChanT:
			// _Chan non_recvchantype
			// _Chan _Comm ntype
			if !p.checkPermit("chanKw") {
				return nil
			}
			t := new(nodes.ChanType)
			p.CheckHashCmd(p.hash, func() {
				p.Next()
				t.SetPos(pos)
				if p.Got(nodes.ArrowT) {
					t.Dir = nodes.SendOnly
				}
				t.Elem = p.ChanElem()
			})
			return t

		case nodes.MapT:
			// _Map '[' ntype ']' ntype
			if !p.checkPermit("mapKw") {
				p.Advance(nodes.SemiT)
				return nil
			}
			t := new(nodes.MapType)
			p.CheckHashCmd(p.hash, func() {
				p.Next()
				t.SetPos(pos)
				if p.dynamicBlock != "" && p.tok == nodes.LbraceT {
					//short map syntax in literals: "map" instead of "map[T]U" with backing slice "[]T"
					t.Key = p.NewName("any")
					t.Value = p.NewName("any")
					t.Dyn = true
				} else {
					p.Want(nodes.LbrackT)
					t.Key = p.Type()
					p.Want(nodes.RbrackT)
					t.Value = p.Type()
				}
			})
			return t

		case nodes.StructT:
			return p.StructType()

		case nodes.InterfaceT:
			return p.InterfaceType()

		case nodes.LparenT:
			p.Next()
			t := p.Type()
			p.Want(nodes.RparenT)
			return t

		case nodes.LiteralT:
			lit := p.OLiteral()
			if lit.Kind == nodes.StringLit {
				a := p.ProcImportAlias(lit, "")
				return p.DotName(p.NewName(a))
			} else {
				return nil
			}
		}
	}

	return nil
}

//--------------------------------------------------------------------------------
func (p *parser) ChanElem() nodes.Expr {
	if trace {
		defer p.trace("chanElem")("")
	}

	typ := p.TypeOrNil()
	if typ == nil {
		typ = p.BadExpr()
		p.SyntaxError("missing channel element type")
		// assume element type is simply absent - don't advance
	}

	return typ
}

//--------------------------------------------------------------------------------
func (p *parser) FuncType() *nodes.FuncType {
	if trace {
		defer p.trace("funcType")("")
	}

	typ := new(nodes.FuncType)
	typ.SetPos(p.Pos())
	typ.ParamList = p.ParamList()
	typ.ResultList = p.FuncResult()

	return typ
}

//--------------------------------------------------------------------------------
// Parameters    = "(" [ ParameterList [ "," ] ] ")" .
// ParameterList = ParameterDecl { "," ParameterDecl } .
func (p *parser) ParamList() (list []*nodes.Field) {
	if trace {
		defer p.trace("paramList")("")
	}

	pos := p.Pos()
	var named int // number of parameters that have an explicit name and type
	p.List(nodes.LparenT, nodes.CommaT, nodes.RparenT, func() bool {
		if par := p.ParamDeclOrNil(); par != nil {
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
		var typ nodes.Expr
		for i := len(list) - 1; i >= 0; i-- {
			if par := list[i]; par.Type != nil {
				typ = par.Type
				if par.Name == nil {
					ok = false
					n := p.NewName("_")
					n.SetPos(typ.Pos()) // correct position
					par.Name = n
				}
			} else if typ != nil {
				par.Type = typ
			} else {
				// par.Type == nil && typ == nil => we only have a par.Name
				ok = false
				t := p.BadExpr()
				t.SetPos(par.Name.Pos()) // correct position
				par.Type = t
			}
		}
		if !ok {
			p.SyntaxErrorAt(pos, "mixed named and unnamed function parameters")
		}
	}

	return
}

//--------------------------------------------------------------------------------
// ParameterDecl = [ IdentifierList ] [ "..." ] Type .
func (p *parser) ParamDeclOrNil() *nodes.Field {
	if trace {
		defer p.trace("paramDecl")("")
	}

	f := new(nodes.Field)
	f.SetPos(p.Pos())

	if p.tok == nodes.NameT || p.TokIsKeywordName() {
		f.Name = p.Name()
		if p.tok == nodes.NameT || p.TokIsKeywordName() {
			f.Type = p.Type()
		} else {
			switch p.tok {
			//case nodes.NameT, nodes.StarT, nodes.ArrowT, nodes.FuncT, nodes.LbrackT,
			case nodes.StarT, nodes.ArrowT, nodes.FuncT, nodes.LbrackT,
				nodes.ChanT, nodes.MapT, nodes.StructT, nodes.InterfaceT, nodes.LparenT:
				// sym name_or_type
				f.Type = p.Type()

			case nodes.DotDotDotT:
				// sym dotdotdot
				f.Type = p.DotsType()

			case nodes.DotT:
				// name_or_type
				// from dotname
				f.Type = p.DotName(f.Name)
				f.Name = nil
			}
		}
	} else {
		switch p.tok {
		case nodes.ArrowT, nodes.StarT, nodes.FuncT, nodes.LbrackT, nodes.ChanT, nodes.MapT, nodes.StructT, nodes.InterfaceT, nodes.LparenT:
			// name_or_type
			f.Type = p.Type()

		case nodes.DotDotDotT:
			// dotdotdot
			f.Type = p.DotsType()

		default:
			p.SyntaxError("expecting )")
			p.Advance(nodes.CommaT, nodes.RparenT)
			return nil
		}
	}

	return f
}

//--------------------------------------------------------------------------------
// ...Type
func (p *parser) DotsType() *nodes.DotsType {
	if trace {
		defer p.trace("dotsType")("")
	}

	t := new(nodes.DotsType)
	t.SetPos(p.Pos())

	p.Want(nodes.DotDotDotT)
	t.Elem = p.TypeOrNil()
	if t.Elem == nil {
		t.Elem = p.BadExpr()
		p.SyntaxError("final argument in variadic function missing type")
	}

	return t
}

//--------------------------------------------------------------------------------
func (p *parser) DotName(name *nodes.Name) nodes.Expr {
	if trace {
		defer p.trace("dotName")("")
	}

	if p.tok == nodes.DotT {
		s := new(nodes.SelectorExpr)
		s.SetPos(p.Pos())
		p.Next()
		s.X = name
		s.Sel = p.Name()
		return s
	}
	return name
}

//--------------------------------------------------------------------------------
// Result = Parameters | Type .
func (p *parser) FuncResult() []*nodes.Field {
	if trace {
		defer p.trace("funcResult")("")
	}

	if p.tok == nodes.LparenT {
		return p.ParamList()
	}

	pos := p.Pos()
	if typ := p.TypeOrNil(); typ != nil {
		f := new(nodes.Field)
		f.SetPos(pos)
		f.Type = typ
		return []*nodes.Field{f}
	}

	return nil
}

//--------------------------------------------------------------------------------
// StructType = "struct" "{" { FieldDecl ";" } "}" .
func (p *parser) StructType() *nodes.StructType {
	if trace {
		defer p.trace("structType")("")
	}

	typ := new(nodes.StructType)
	typ.SetPos(p.Pos())
	if !p.checkPermit("structKw") {
		p.Advance(nodes.SemiT)
		return nil
	}

	p.CheckHashCmd(p.hash, func() {
		p.Want(nodes.StructT)
		p.List(nodes.LbraceT, nodes.SemiT, nodes.RbraceT, func() bool {
			p.FieldDecl(typ)
			return false
		})
	})
	return typ
}

//--------------------------------------------------------------------------------
// FieldDecl      = (IdentifierList Type | AnonymousField) [ Tag ] .
// AnonymousField = [ "*" ] TypeName .
// Tag            = string_lit .
func (p *parser) FieldDecl(styp *nodes.StructType) {
	if trace {
		defer p.trace("fieldDecl")("")
	}

	pos := p.Pos()
	if p.tok == nodes.NameT || p.TokIsKeywordName() {
		name := p.Name()
		if p.tok == nodes.DotT || p.tok == nodes.LiteralT || p.tok == nodes.SemiT || p.tok == nodes.RbraceT {
			// embed oliteral
			typ := p.QualifiedName(name)
			tag := p.OLiteral()
			p.AddField(styp, pos, nil, typ, tag)
			return
		}

		// new_name_list ntype oliteral
		names := p.NameList(name)
		typ := p.Type()
		tag := p.OLiteral()

		for _, name := range names {
			p.AddField(styp, name.Pos(), name, typ, tag)
		}
	} else {
		switch p.tok {
		case nodes.LparenT:
			p.Next()
			if p.tok == nodes.StarT {
				// '(' '*' embed ')' oliteral
				pos := p.Pos()
				p.Next()
				typ := newIndirect(pos, p.QualifiedName(nil))
				p.Want(nodes.RparenT)
				tag := p.OLiteral()
				p.AddField(styp, pos, nil, typ, tag)
				p.SyntaxError("cannot parenthesize embedded type")

			} else {
				// '(' embed ')' oliteral
				typ := p.QualifiedName(nil)
				p.Want(nodes.RparenT)
				tag := p.OLiteral()
				p.AddField(styp, pos, nil, typ, tag)
				p.SyntaxError("cannot parenthesize embedded type")
			}

		case nodes.StarT:
			p.Next()
			if p.Got(nodes.LparenT) {
				// '*' '(' embed ')' oliteral
				typ := newIndirect(pos, p.QualifiedName(nil))
				p.Want(nodes.RparenT)
				tag := p.OLiteral()
				p.AddField(styp, pos, nil, typ, tag)
				p.SyntaxError("cannot parenthesize embedded type")

			} else {
				// '*' embed oliteral
				typ := newIndirect(pos, p.QualifiedName(nil))
				tag := p.OLiteral()
				p.AddField(styp, pos, nil, typ, tag)
			}

		default:
			p.SyntaxError("expecting field name or embedded type")
			p.Advance(nodes.SemiT, nodes.RbraceT)
		}
	}
}

//--------------------------------------------------------------------------------
func (p *parser) AddField(styp *nodes.StructType, pos src.Pos, name *nodes.Name, typ nodes.Expr, tag *nodes.BasicLit) {
	if tag != nil {
		for i := len(styp.FieldList) - len(styp.TagList); i > 0; i-- {
			styp.TagList = append(styp.TagList, nil)
		}
		styp.TagList = append(styp.TagList, tag)
	}

	f := new(nodes.Field)
	f.SetPos(pos)
	f.Name = name
	f.Type = typ
	styp.FieldList = append(styp.FieldList, f)

	if debug && tag != nil && len(styp.FieldList) != len(styp.TagList) {
		panic("inconsistent struct field list")
	}
}

//--------------------------------------------------------------------------------
// InterfaceType = "interface" "{" { MethodSpec ";" } "}" .
func (p *parser) InterfaceType() *nodes.InterfaceType {
	if trace {
		defer p.trace("interfaceType")("")
	}

	if !p.checkPermit("interfaceKw") {
		p.Advance(nodes.SemiT)
		return nil
	}
	typ := new(nodes.InterfaceType)
	typ.SetPos(p.Pos())
	p.CheckHashCmd(p.hash, func() {
		p.Want(nodes.InterfaceT)
		p.List(nodes.LbraceT, nodes.SemiT, nodes.RbraceT, func() bool {
			if m := p.MethodDecl(); m != nil {
				typ.MethodList = append(typ.MethodList, m)
			}
			return false
		})
	})
	return typ
}

//--------------------------------------------------------------------------------
// MethodSpec        = MethodName Signature | InterfaceTypeName .
// MethodName        = identifier .
// InterfaceTypeName = TypeName .
func (p *parser) MethodDecl() *nodes.Field {
	if trace {
		defer p.trace("methodDecl")("")
	}

	if p.tok == nodes.NameT || p.TokIsKeywordName() {
		name := p.Name()

		// accept potential name list but complain
		hasNameList := false
		for p.Got(nodes.CommaT) {
			p.Name()
			hasNameList = true
		}
		if hasNameList {
			p.SyntaxError("name list not allowed in interface type")
			// already progressed, no need to advance
		}

		f := new(nodes.Field)
		f.SetPos(name.Pos())
		if p.tok != nodes.LparenT {
			// packname
			f.Type = p.QualifiedName(name)
			return f
		}

		f.Name = name
		f.Type = p.FuncType()
		return f

	} else {
		switch p.tok {
		case nodes.LparenT:
			p.SyntaxError("cannot parenthesize embedded type")
			f := new(nodes.Field)
			f.SetPos(p.Pos())
			p.Next()
			f.Type = p.QualifiedName(nil)
			p.Want(nodes.RparenT)
			return f

		default:
			p.SyntaxError("expecting method or interface name")
			p.Advance(nodes.SemiT, nodes.RbraceT)
			return nil
		}
	}
}

//--------------------------------------------------------------------------------
// Common productions
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
func (p *parser) List(open, sep, close nodes.Token, f func() bool) src.Pos {
	p.Want(open)

	var done bool
	for p.tok != nodes.EofT && p.tok != close && !done {
		done = f()

		// sep is optional before close
		if !p.Got(sep) && p.tok != close {
			p.SyntaxError(fmt.Sprintf("expecting %s or %s", tokstring(sep), tokstring(close)))
			p.Advance(nodes.RparenT, nodes.RbrackT, nodes.RbraceT)
			if p.tok != close {
				return p.Pos()
				// position could be better but we had an error so we don't care
			}
		}
	}

	pos := p.Pos()
	p.Want(close)
	return pos
}

//--------------------------------------------------------------------------------
func (p *parser) Got(tok nodes.Token) bool {
	if p.tok == tok {
		p.Next()
		return true
	}
	return false
}

//--------------------------------------------------------------------------------
func (p *parser) Want(tok nodes.Token) {
	if !p.Got(tok) {
		p.SyntaxError("expecting " + tokstring(tok))
		p.Advance()
	}
}

//--------------------------------------------------------------------------------
func (p *parser) NewName(value string) *nodes.Name {
	n := new(nodes.Name)
	n.SetPos(p.Pos())
	n.Value = value

	//if p.hashCmdBlock {
	if p.hashCmdBlock && !p.hash {
		for _, r := range value { //look at 1st rune only
			if unicode.IsLower(r) {
				n.Value = "_" + n.Value
			}
			break
		}
	}
	return n
}

//--------------------------------------------------------------------------------
func (p *parser) TokIsKeywordName() bool {
	return p.hashCmdMode && p.hashCmdBlock && !p.hash && p.tok.IsKeyword()
}

//--------------------------------------------------------------------------------
func (p *parser) Name() *nodes.Name {
	// no tracing to avoid overly verbose output

	if p.tok == nodes.NameT {
		n := p.NewName(p.lit)
		p.Next()
		return n
	} else if p.TokIsKeywordName() {
		n := p.NewName(p.tok.String())
		p.Next()
		return n
	}

	n := p.NewName("_")
	p.SyntaxError("expecting name")
	p.Advance()
	return n
}

//--------------------------------------------------------------------------------
// IdentifierList = identifier { "," identifier } .
// The first name must be provided.
func (p *parser) NameList(first *nodes.Name) []*nodes.Name {
	if trace {
		defer p.trace("nameList")("")
	}

	if debug && first == nil {
		panic("first name not provided")
	}

	l := []*nodes.Name{first}
	for p.Got(nodes.CommaT) {
		l = append(l, p.Name())
	}

	return l
}

//--------------------------------------------------------------------------------
// The first name may be provided, or nil.
func (p *parser) QualifiedName(name *nodes.Name) nodes.Expr {
	if trace {
		defer p.trace("qualifiedName")("")
	}

	switch {
	case name != nil:
		// name is provided
	case p.tok == nodes.NameT, p.TokIsKeywordName():
		name = p.Name()
	default:
		name = p.NewName("_")
		p.SyntaxError("expecting name")
		p.Advance(nodes.DotT, nodes.SemiT, nodes.RbraceT)
	}

	return p.DotName(name)
}

//--------------------------------------------------------------------------------
func (p *parser) OLiteral() *nodes.BasicLit {
	if p.tok == nodes.LiteralT {
		b := new(nodes.BasicLit)
		b.SetPos(p.Pos())
		b.Value = p.lit
		b.Kind = p.kind
		p.Next()
		return b
	}
	return nil
}

//--------------------------------------------------------------------------------
func (p *parser) IsName(ss ...string) bool {
	for _, s := range ss {
		if p.tok == nodes.NameT && p.lit == s {
			return true
		}
	}
	return false
}

//--------------------------------------------------------------------------------
func (p *parser) NewBlankFunc(s string) *nodes.FuncDecl {
	fn := &nodes.FuncDecl{
		Name: p.NewName(s),
		Type: &nodes.FuncType{
			ParamList:  []*nodes.Field{},
			ResultList: nil,
		},
		Body: &nodes.BlockStmt{
			List:   []nodes.Stmt{},
			Rbrace: p.Pos(),
		},
	}
	fn.SetPos(p.Pos())
	fn.Type.SetPos(p.Pos())
	fn.Body.SetPos(p.Pos())
	return fn
}

//--------------------------------------------------------------------------------
func (p *parser) NewBlankFuncLit() *nodes.FuncLit {
	fn := &nodes.FuncLit{
		Type: &nodes.FuncType{
			ParamList:  []*nodes.Field{},
			ResultList: nil,
		},
		Body: &nodes.BlockStmt{
			List:   []nodes.Stmt{},
			Rbrace: p.Pos(),
		},
	}
	fn.SetPos(p.Pos())
	fn.Type.SetPos(p.Pos())
	fn.Body.SetPos(p.Pos())
	return fn
}

//--------------------------------------------------------------------------------
// Error handling
//--------------------------------------------------------------------------------
// posAt returns the Pos value for (line, col) and the current position base.
func (p *parser) PosAt(line, col uint) src.Pos {
	return src.MakePos(p.base, line, col)
}

//--------------------------------------------------------------------------------
// error reports an error at the given position.
func (p *parser) ErrorAt(pos src.Pos, msg string) {
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
func (p *parser) SyntaxErrorAt(pos src.Pos, msg string) {
	if trace {
		defer p.trace("syntaxError (" + msg + ")")("")
		//p.print("syntax error: " + msg)
	}

	if p.tok == nodes.EofT && p.first != nil {
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
		p.ErrorAt(pos, "syntax error: "+msg)
		return
	}

	// determine token string
	var tok string
	switch p.tok {
	case nodes.NameT, nodes.SemiT:
		tok = p.lit
	case nodes.LiteralT:
		tok = "literal " + p.lit
	case nodes.OperatorT:
		tok = p.op.String()
	case nodes.AssignOpT:
		tok = p.op.String() + "="
	case nodes.IncOpT:
		tok = p.op.String()
		tok += tok
	default:
		tok = tokstring(p.tok)
	}

	p.ErrorAt(pos, "syntax error: unexpected "+tok+msg)
}

//--------------------------------------------------------------------------------
// Convenience methods using the current token position.
func (p *parser) Pos() src.Pos           { return p.PosAt(p.line, p.col) }
func (p *parser) Error(msg string)       { p.ErrorAt(p.Pos(), msg) }
func (p *parser) SyntaxError(msg string) { p.SyntaxErrorAt(p.Pos(), msg) }

// The stopset contains keywords that start a statement.
// They are good synchronization points in case of syntax
// errors and (usually) shouldn't be skipped over.
const stopset uint64 = 1<<nodes.BreakT |
	1<<nodes.ConstT |
	1<<nodes.ContinueT |
	1<<nodes.DeferT |
	1<<nodes.FallthroughT |
	1<<nodes.ForT |
	//1<<_Func |
	1<<nodes.GoT |
	1<<nodes.GotoT |
	1<<nodes.IfT |
	1<<nodes.ReturnT |
	1<<nodes.SelectT |
	1<<nodes.SwitchT |
	1<<nodes.TypeT |
	1<<nodes.VarT

//--------------------------------------------------------------------------------
// Advance consumes tokens until it finds a token of the stopset or followlist.
// The stopset is only considered if we are inside a function (p.fnest > 0).
// The followlist is the list of valid tokens that can follow a production;
// if it is empty, exactly one (non-EOF) token is consumed to ensure progress.
func (p *parser) Advance(followlist ...nodes.Token) {
	if trace {
		p.print(fmt.Sprintf("advance %s", followlist))
	}

	// compute follow set
	// (not speed critical, advance is only called in error situations)
	var followset uint64 = 1 << nodes.EofT // don't skip over EOF
	if len(followlist) > 0 {
		if p.fnest > 0 {
			followset |= stopset
		}
		for _, tok := range followlist {
			followset |= 1 << tok
		}
	}

	for !nodes.Contains(followset, p.tok) {
		if trace {
			p.print("skip " + p.tok.String())
		}
		p.Next()
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
// functions not attached to parser
//--------------------------------------------------------------------------------
// tokstring returns the English word for selected punctuation tokens
// for more readable error messages.
func tokstring(tok nodes.Token) string {
	switch tok {
	//case _EOF:
	//return "EOF"
	case nodes.CommaT:
		return "comma"
	case nodes.SemiT:
		return "semicolon"
	}
	return tok.String()
}

//--------------------------------------------------------------------------------
func isEmptyFuncDecl(dcl nodes.Decl) bool {
	f, ok := dcl.(*nodes.FuncDecl)
	return ok && f.Body == nil
}

//--------------------------------------------------------------------------------
func newIndirect(pos src.Pos, typ nodes.Expr) nodes.Expr {
	o := new(nodes.Operation)
	o.SetPos(pos)
	o.Op = nodes.Mul
	o.X = typ
	return o
}

//--------------------------------------------------------------------------------
// unparen removes all parentheses around an expression.
func unparen(x nodes.Expr) nodes.Expr {
	for {
		p, ok := x.(*nodes.ParenExpr)
		if !ok {
			break
		}
		x = p.X
	}
	return x
}

//--------------------------------------------------------------------------------
