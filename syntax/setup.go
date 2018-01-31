// Copyright 2016-17 The Go and Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"github.com/grolang/gro/macros"
	"github.com/grolang/gro/nodes"
)

//--------------------------------------------------------------------------------
var permitErrorMsgs = map[string]string{
	"inferPkg":      "infer-packages disabled but no explicit \"package\" keyword present",
	"multiPkg":      "multi-packages disabled but more than one package present",
	"inplaceImps":   "inplace-imports disabled but are present",
	"inferMain":     "infer-main disabled but no explicit \"main\" function present",
	"pkgSectBlocks": "using block-style notation for packages and sections is disabled but it is being used",
	"genericCall":   "calling generic-type packages disabled but import arguments present",
	"genericDef":    "defining generic packages is disabled but one is present",

	"projectKw":  "\"project\" keyword disabled but keyword is present",
	"useKw":      "\"use\" keywords are disabled but keyword is present",
	"includeKw":  "\"include\" keywords are disabled but keyword is present",
	"internalKw": "\"internal\" keywords disabled but keyword is present",
	"sectionKw":  "\"section\" keywords are disabled but keyword is present",
	"mainKw":     "\"main\" keywords are disabled but keyword is present",
	"testcodeKw": "\"testcode\" keywords are disabled but keyword is present",
	"procKw":     "\"proc\" keywords are disabled but keyword is present",
	"doKw":       "\"do\" keywords are disabled but keyword is present",

	"packageKw": "\"package\" (and similar) keywords are disabled but keyword is present",
	"importKw":  "\"import\" keywords are disabled but keyword is present",
	"constKw":   "\"const\" keywords are disabled but keyword is present",
	"typeKw":    "\"type\" keywords are disabled but keyword is present",
	"varKw":     "\"var\" keywords are disabled but keyword is present",
	"funcKw":    "\"func\" keywords are disabled but keyword is present",

	"structKw":    "\"struct\" keywords are disabled but keyword is present",
	"mapKw":       "\"map\" keywords are disabled but keyword is present",
	"chanKw":      "\"chan\" keywords are disabled but keyword is present",
	"interfaceKw": "\"interface\" keywords are disabled but keyword is present",

	"ifKw":      "if-statement has been disabled but is present",
	"elseKw":    "\"else\" keywords are disabled but keyword is present",
	"forKw":     "for-statement has been disabled but is present",
	"rangeKw":   "\"range\" keywords are disabled but keyword is present",
	"switchKw":  "switch-statement has been disabled but is present",
	"caseKw":    "\"case\" keywords are disabled but keyword is present",
	"defaultKw": "\"default\" keywords are disabled but keyword is present",
	"selectKw":  "select-statement has been disabled but is present",
	"deferKw":   "defer-statement has been disabled but is present",
	"goKw":      "go-statement has been disabled but is present",

	"returnKw":      "return-statement has been disabled but is present",
	"gotoKw":        "goto-statement has been disabled but is present",
	"continueKw":    "continue-statement has been disabled but is present",
	"breakKw":       "break-statement has been disabled but is present",
	"fallthroughKw": "fallthrough-statement has been disabled but is present",
}

//--------------------------------------------------------------------------------
func (p *parser) CheckHashCmd(hashFlag bool, f func()) {
	if p.hashCmdBlock && !hashFlag {
		p.SyntaxError("Keywords must be prepended with # within scope of other #-keywords")
		return
	}
	oldHashCmdBlock := p.hashCmdBlock
	if p.hashCmdMode && hashFlag {
		p.hashCmdBlock = true
	}
	f()
	p.hashCmdBlock = oldHashCmdBlock
}

//--------------------------------------------------------------------------------
func (p *parser) checkPermit(permit string) bool {
	if !p.permits[permit] {
		p.SyntaxError(permitErrorMsgs[permit])
		return false
	} else {
		return true
	}
}

//--------------------------------------------------------------------------------
func (p *parser) setupProfile() {
	p.permits = map[string]bool{}

	switch p.currProj.FileExt {
	//- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	//hash-cmds, which rely on dynamic typing
	case "grooy":
		p.hashCmdMode = true
		fallthrough

	//- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	//dynamic typing
	case "groo":
		p.dynamicMode = true
		p.dynamicBlock = "groo"
		p.dynCharSet = "utf88"
		fallthrough

	//- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	//generic typing
	case "grog":
		for _, kw := range [...]string{
			"genericCall", //enable imports of generic packages
			"genericDef",  //enable definitions of generic packages
		} {
			p.permits[kw] = true
		}
		fallthrough

	//- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	//standard grolang extensions
	case "gro", "":
		for _, kw := range [...]string{
			"assert",  //enable "assert" macro
			"let",     //enable "let" macro
			"prepare", //enable "prepare" macro
			"execute", //enable "execute" macro
			"run",     //enable "run" macro
			"test",    //enable "test" macro

			"inferPkg",      //enable package names to be inferred
			"multiPkg",      //enable more than one package in a single file
			"inplaceImps",   //enable in-place spec strings for package names
			"inferMain",     //enable main function to be inferred
			"pkgSectBlocks", //enable block notation for packages and sections

			"projectKw",  //enable "project" keyword
			"useKw",      //enable "use" keyword
			"includeKw",  //enable "include" keyword
			"internalKw", //enable "internal" keyword
			"sectionKw",  //enable "section" keyword
			"mainKw",     //enable "main" keyword
			"testcodeKw", //enable "testcode" keyword
			"procKw",     //enable "proc" keyword
			"doKw",       //enable "do" keyword

			//TODO: yet to code blacklist for these...
			"escapeEscapeInStrings", //enable \e in runes/strings
		} {
			p.permits[kw] = true
		}
		fallthrough

	//- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	//interfaces, i.e. final layer for exact golang syntax
	case "go":
		for _, kw := range [...]string{
			"interfaceKw", //enable use of the interface keyword

			//TODO: yet to code blacklist for these...
			"embeddedInterface",            //allow embedded interfaces
			"typeAssertion",                //allow type assertions
			"twoValuedTypeAssertion",       //allow two-valued type assertions
			"typeSwitchStmt",               //allow type-switch stmt
			"simpleStmtOnTypeSwitch",       //allow simple stmt prefix on type-switch stmt
			"requireDefaultInTypeSwitch",   //require default clause in type-switch stmt
			"multivaluedCasesInTypeSwitch", //allow multi-valued case clauses in type-switch stmt
			"emptyClausesInTypeSwitch",     //allow empty case/default stmt sequences in type-switch stmt
			"shortDeclInTypeSwitch",        //allow short-declaration in type-switch stmt
			"breakKwInTypeSwitch",          //allow break kw in type-switch stmt; labeled/unlabeled
		} {
			p.permits[kw] = true
		}
		fallthrough

	//methods
	case "g0750":
		for _, pf := range [...]string{

			//TODO: yet to code blacklist for these...
			"methods",                  //allow defining methods on types
			"valueMethodsOnly",         //restrict methods to value only
			"pointerMethodsOnly",       //whether to restrict methods to pointer only
			"matchingMethodRcvrTarget", //restrict method set to either all values or all pointers
			"omitRcvrName",             //allow omitting receiver name
		} {
			p.permits[pf] = true
		}
		fallthrough

	//channels
	case "g0740":
		for _, pf := range [...]string{
			"selectKw", //enable use of the select keyword
			"chanKw",   //enable use of the chan keyword

			//TODO: yet to code blacklist for these...
			"chanSendStmt",      //allow send stmts, i.e. r <- c stmt
			"chanReceiveOp",     //allow channel receives, i.e. unary <-
			"directedChans",     //allow directed channels, i.e. both send and receive
			"closeChannels",     //enable close on channels
			"twoValuedReceives", //allow two-valued receive op
			//"makeChannels",              //allow make on channels
			//"lenCapChannels",            //allow len, cap on channels
			"defaultsInSelectStmt",      //require default clause in select stmts
			"twoValSelectStmt",          //require two-value lhs in select stmts
			"emptyCaseDefaultsInSelect", //allow empty case/default stmt sequences in select stmts
			"breakKwInSelectStmts",      //allow break kw in select stmts; labeled/unlabeled
			//"chanForRangeStmt",          //allow for-range stmts for channels
			//"twoValChanForRangeStmt",    //prohibit two-value lhs in for-range stmts for channels
		} {
			p.permits[pf] = true
		}
		fallthrough

	//functions
	case "g0730":
		for _, pf := range [...]string{
			"deferKw", //enable use of the defer keyword
			"goKw",    //enable use of the go keyword

			//TODO: yet to code blacklist for these...
			"absentFuncParamNames",  //allow function type param names to be absent in param lists
			"absentFuncResultNames", //allow function type param names to be absent in result lists
			"blankFuncResultNames",  //allow blank name in param list and/or result list
			"variadicFuncParams",    //allow variadic param
			"funcLits",              //allow func literals (with closures)
			//"nonterminatingReturn",    //allow return as non-terminating stmt in function
			"emptyReturnsWhenResults", //disallow empty-valued return stmts when enclosing function has results
			"panicAndRecover",         //allow panic and recover spec-ids
			"variadicArgs",            //allow calls with variadic ...
		} {
			p.permits[pf] = true
		}
		fallthrough

	//maps
	case "g0720":
		for _, pf := range [...]string{
			"mapKw", //enable use of the map keyword

			//TODO: yet to code blacklist for these...
			"makeMaps",     //allow make on maps
			"lenMaps",      //enable len on maps
			"deleteMaps",   //allow delete on maps
			"forRangeMaps", //allow for-range stmts for maps
		} {
			p.permits[pf] = true
		}
		fallthrough

	//slices
	case "g0710":
		for _, pf := range [...]string{

			//TODO: yet to code blacklist for these...
			"sliceDecls",              //allow slice declarations
			"lenCapSlices",            //allow len, cap for slices
			"appendCopySlices",        //allow append, copy for slices
			"makeSlices",              //allow make for slices
			"rangeSlices",             //allow for-range stmts on slices
			"twoIndexSlices",          //allow slice expressions with 2 indexes
			"threeIndexSlices",        //allow slice exprs with 3 indexes
			"elideFirstIndexInSlice",  //allow elided first index in slice expression with 2 or 3 indexes
			"elideSecondIndexInSlice", //allow elided second index in slice expression with 2 indexes
		} {
			p.permits[pf] = true
		}
		fallthrough

	//for-range
	case "g0630":
		for _, pf := range [...]string{
			"rangeKw", //enable use of the range keyword

			//TODO: yet to code blacklist for these...
			"shortDeclInRanges", //allow short-declaration in for-range stmts
			"oneValForRanges",   //require at least one value lhs in for-range stmts
			"twoValRangesOnly",  //require two-value lhs in for-range stmts //except for channels
		} {
			p.permits[pf] = true
		}
		fallthrough

	//std switch stmt
	case "g0610":
		for _, pf := range [...]string{
			"switchKw",      //enable use of the switch keyword
			"caseKw",        //enable use of the case keyword
			"defaultKw",     //enable use of the default keyword
			"fallthroughKw", //enable use of the fallthrough keyword

			//TODO: yet to code blacklist for these...
			"blankSwitchStmt",          //allow blank expression in std-switch stmt
			"simpleStmtPrefixInSwitch", //allow simple stmt prefix on std-switch stmt
			"defaultInSwitch",          //require default clause in std-switch stmt
			"multiValCases",            //allow multi-valued case clauses in std-switch stmt
			"breakInSwitch",            //allow break kw in std-switch stmt
			"fallthruInSwitch",         //allow fallthrough kw in std-switch stmt
			"emptyCaseDefaultInSwitch", //allow empty case/default stmt sequences in std-switch stmt
		} {
			p.permits[pf] = true
		}
		fallthrough

	//complex numbers
	case "g0520":
		for _, pf := range [...]string{

			//TODO: yet to code blacklist for these...
			"complexLits",        //allow complex lits
			"complexRealImagIds", //allow complex, real, and imag
		} {
			p.permits[pf] = true
		}
		fallthrough

	//not in g
	case "g0500":
		for _, pf := range [...]string{
			//TODO: also prohibit in: switch x.(type)

			//TODO: yet to code blacklist for these...
			"structEmbeddedFields", //allow struct embedded fields
			"structPointerFields",  //allow pointers to embedded struct fields
			"structFieldTags",      //allow struct field tags

			"simpleStmtOnIfStmt", //allow simple stmt prefix on if stmt

			"typeDefns", //allow type definitions

			"rangeArrays", //allow for-range stmts on arrays

			"rawStringSyntax", //allow raw strings
			"octalInStrings",  //allow '\077' in strings

			"octalNums", //allow octal

			"elidedZeroInFloats",  //allow 0. or .1 in floats/complexes
			"leadingZeroInFloats", //allow 072.34 in floats

			"byteAlias", //allow byte alias
			"runeAlias", //allow "rune" alias for int32

			"blankIdInAssigns",  //allow blank identifier in assignments
			"multivalueAssigns", //allow multi-value assignments
			"opAssigns",         //allow op-assignments based on permission of op
			"incrDecrs",         //allow incr/decr stmts

			"unicodeInIdNames",    //otherwise, restricted to ASCII in identifier names
			"cGo",                 //allow cgo function declarations
			"blankLabels",         //allow blank labels
			"declarePredeclareds", //allow top-level declarations of predeclared special identifiers

			"unsafePkg", //allow use of unsafe pkg
		} {
			p.permits[pf] = true
		}
		fallthrough

	//- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	//std subset of go known as g
	case "g0450", "g":
		for _, pf := range [...]string{
			"typeKw",   //enable use of the type keyword
			"returnKw", //enable use of the return keyword
			"funcKw",   //enable use of the func keyword

			//TODO: yet to code blacklist for these...
			"funcDecls",            //allow func as declarations
			"callExprsAndConverts", //allow call expressions and conversions
		} {
			p.permits[pf] = true
		}
		fallthrough

	//for-while
	case "g0210":
		for _, pf := range [...]string{
			"forKw",      //enable use of the for keyword
			"breakKw",    //enable use of the break keyword
			"continueKw", //enable use of the continue keyword

			//TODO: yet to code blacklist for these...
			"forWhileStmts",      //allow for-while stmts
			"breakInForStmt",     //allow break kw in for stmt; labeled/unlabeled
			"continueInForStmt",  //allow continue kw in for stmt; labeled/unlabeled
			"blankHeadForStmt",   //allow empty head in for-while stmt
			"threeClauseForStmt", //allow 3-clause for-while stmts
			"initInForStmt",      //require init in 3-clause for-clause stmts
			"postInForStmt",      //require post-stmt in 3-clause for-clause stmts
		} {
			p.permits[pf] = true
		}
		fallthrough

	//if stmt
	case "g0200":
		for _, pf := range [...]string{
			"ifKw",   //enable use of the if keyword
			"elseKw", //enable use of the else keyword

			//TODO: yet to code blacklist for these...
			"ifElseClause", //allow else-if clause on if stmt
		} {
			p.permits[pf] = true
		}
		fallthrough

	//pointers
	case "g0190":
		for _, pf := range [...]string{

			//TODO: yet to code blacklist for these...
			"pointerTypes",       //allow pointer types
			"newSpecId",          //allow pointers with `new(T)`
			"addrOfCompositeLit", //allow pointers with `&T{...}`
			"unaryIndirection",   //allow unary "*"
			"unaryAddressOf",     //allow unary "&"
		} {
			p.permits[pf] = true
		}
		fallthrough

	//arrays
	case "g0180":
		for _, pf := range [...]string{

			//TODO: yet to code blacklist for these...
			"arrays",             //allow arrays
			"lenCapArrays",       //allow len (and cap) for arrays
			"inferredArraySizes", //allow ... in array literals
		} {
			p.permits[pf] = true
		}
		fallthrough

	//structs
	case "g0170":
		for _, pf := range [...]string{
			"structKw", //enable use of the struct keyword

			//TODO: yet to code blacklist for these...
			"structMultiFieldOfSameType", //allow struct multi-field with same type
			"structPadding",              //allow struct padding fields
			"structSelectors",            //allow struct selectors and qualified names
			"structComposites",           //allow composite struct literals
		} {
			p.permits[pf] = true
		}
		fallthrough

	//types
	case "g0160":
		for _, pf := range [...]string{

			//TODO: yet to code blacklist for these...
			"typeAliases", //allow type aliases
			"typeGroups",  //allow type groups
		} {
			p.permits[pf] = true
		}
		fallthrough

	//strings
	case "g0150":
		for _, pf := range [...]string{

			//TODO: yet to code blacklist for these...
			"lenOfStrings",          //enable len on strings
			"hexInStrings",          //allow '\x1f' in strings
			"shortUnicodeInStrings", //allow '\uFFe1' in strings
			"longUnicodeInStrings",  //allow '\U0001FFFF' in strings
			"escapesInStrings",      //allow \a, \b, \f, \n, \r, \t, \v
			"indexStrings",          //Indexing: allow index expressions - 1 index
		} {
			p.permits[pf] = true
		}
		fallthrough

	//booleans
	case "g0130":
		for _, pf := range [...]string{

			//TODO: yet to code blacklist for these...
			"trueFalseIds",  //allow true and false
			"logicalOps",    //allow logical ops  !  ||  &&
			"equalityOps",   //allow equality ops  ==  !=
			"comparisonOps", //allow ordering ops  <  <=  >  >=
		} {
			p.permits[pf] = true
		}
		fallthrough

	//floats
	case "g0110":
		for _, pf := range [...]string{

			//TODO: yet to code blacklist for these...
			"standardFloats", //allow standard float notation
		} {
			p.permits[pf] = true
		}
		fallthrough

	//integers
	case "g0100":
		for _, pf := range [...]string{

			//TODO: yet to code blacklist for these...
			"hexNums",           //allow hex
			"archDependentInts", //allow int, uint, and uintptr
			"sizedInts",         //allow int8, int16, int32, int64
			"sizedUnsigneds",    //allow uint8, uint16, uint32, uint64
			"binPlusOp",         //allow math/str "+"
			"unaryNumericOps",   //allow math u"+", u"-", "-", "*", "/"
			"modOp",             //allow integer "%"
			"bitwiseOps",        //allow bitwise u"^", "|", "^", "&", "&^"
			"shiftOps",          //allow shift "<<", ">>"
		} {
			p.permits[pf] = true
		}
		fallthrough

	//assignments
	case "g0070":
		for _, pf := range [...]string{

			//TODO: yet to code blacklist for these...
			"assignments", //allow assignments
		} {
			p.permits[pf] = true
		}
		fallthrough

	//variables
	case "g0060":
		for _, pf := range [...]string{
			"varKw", //enable use of the var keyword

			//TODO: yet to code blacklist for these...
			"typedVars",            //allow typed variables
			"defaultZeroesForVars", //allow default zero values for variables
			"multivalueVarDecls",   //allow multi-value var declarations
			"varGroups",            //allow var groups
			"shortVarDecls",        //allow short-variable declarations
			"blankIdInShortDecls",  //allow blank identifier in short-declarations
			"multivalueShortDecls", //allow multi-value short declarations
		} {
			p.permits[pf] = true
		}
		fallthrough

	//constants
	case "g0050":
		for _, pf := range [...]string{
			"constKw", //enable use of the const keyword

			//TODO: yet to code blacklist for these...
			"typedConsts",       //allow typed constants
			"multivaluedConsts", //allow multi-value const declarations
			"constGroups",       //allow const groups
			"iota",              //allow iota
			"inferedLinesInConstGroup", //allow omitted infered values in const group
			"iotaMultiuse",             //allow multi-use of iota within a const declaration
			"blankConsts",              //allow blank constants
		} {
			p.permits[pf] = true
		}
		fallthrough

	//imports
	case "g0040":
		for _, pf := range [...]string{
			"importKw", //allow imports

			//TODO: yet to code blacklist for these...
			"importGroups",     //allow imports in groups
			"importAliases",    //allow aliases on imports
			"unaliasedImports", //allow aliases with default package name
			"importBlankAlias", //allow underscore as alias on imports
			"importDotAlias",   //allow dot as alias on imports
		} {
			p.permits[pf] = true
		}
		fallthrough

	//basic stuff
	case "g0030":
		for _, pf := range [...]string{
			"gotoKw", //allow goto-stmts

			//TODO: yet to code blacklist for these...
			"blockComments", //otherwise, restricted to line comments only
			"exportedIds",   //allow exported identifiers - top-level, fields, methods
			"initFuncs",     //allow init functions
			"useLabels",     //allow labels
		} {
			p.permits[pf] = true
		}
		fallthrough

	//non-main functions and packages
	case "g0020":
		for _, pf := range [...]string{

			//TODO: yet to code blacklist for these...
			"nonMainFunc", //allow non-main function
			"nonMainPkg",  //allow non-main package
		} {
			p.permits[pf] = true
		}
		fallthrough

	//- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	//enough for a minimal program
	case "g0010":
		for _, pf := range [...]string{
			"packageKw", //enable use of the package keyword

			//TODO: yet to code blacklist for these...
			"mainPkgAndFunc", //allow package main and func main()
			"printSpecIds",   //allow print and println spec-ids
		} {
			p.permits[pf] = true
		}
		fallthrough

	default:
	}
}

//--------------------------------------------------------------------------------
func (p *parser) setupRegistries() {
	p.useRegistry = map[string]func([]string, []string){
		"blacklist": func(rets []string, args []string) {
			macros.InitBlacklist(p, rets, args)
		},
		"dynamic": func(rets []string, args []string) {
			macros.InitDynamic(p, rets, args)
		},
		"generics": func(rets []string, args []string) {
			macros.InitGenerics(p, rets, args)
		},
		"linedirectives": func(rets []string, args []string) {
			macros.InitLineDirectives(p, rets, args)
		},
	}

	p.stmtRegistry = map[string]func(nodes.GeneralParser, ...interface{}) nodes.Stmt{
		"assert": func(p nodes.GeneralParser, _ ...interface{}) nodes.Stmt {
			return macros.Assert(p)
		},
		"let": func(p nodes.GeneralParser, rest ...interface{}) nodes.Stmt {
			if len(rest) != 1 {
				panic("argument error with \"let\" macro")
			}
			if stmt, ok := rest[0].(func() nodes.Stmt); ok {
				return macros.Let(p, stmt)
			}
			panic("argument error with \"let\" macro")
		},
		"prepare": func(p nodes.GeneralParser, _ ...interface{}) nodes.Stmt {
			return macros.GroSystemCmd(p, "prepare")
		},
		"execute": func(p nodes.GeneralParser, _ ...interface{}) nodes.Stmt {
			return macros.GroSystemCmd(p, "execute")
		},
		"run": func(p nodes.GeneralParser, _ ...interface{}) nodes.Stmt {
			return macros.GroSystemCmd(p, "run")
		},
		"test": func(p nodes.GeneralParser, _ ...interface{}) nodes.Stmt {
			return macros.GroSystemCmd(p, "test")
		},
	}

	p.typeRegistry = map[string]func(nodes.GeneralParser, ...interface{}) nodes.Expr{
		"propertied": func(p nodes.GeneralParser, _ ...interface{}) nodes.Expr {
			return macros.Propertied(p)
		},
	}
}

//--------------------------------------------------------------------------------
