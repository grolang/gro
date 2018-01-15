// Copyright 2016-17 The Go and Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nodes

//--------------------------------------------------------------------------------
// ImplicitOne represents x++, x-- as assignments x += ImplicitOne, x -= ImplicitOne.
// It's used ImplicitOne in parser.go so they'll be printed by printer.go
// as x++/-- instead of x +=/-= 1. ImplicitOne should not be used elsewhere.
var ImplicitOne = &BasicLit{Value: "1"}

type printer interface {
	Print(...interface{})
	PrintNode(Node)
	Linebreaks() bool
}

//--------------------------------------------------------------------------------
func groupFor(d Decl) (Token, *DeclGroup) {
	switch d := d.(type) {
	case *ImportDecl:
		return ImportT, d.Group
	case *ConstDecl:
		return ConstT, d.Group
	case *TypeDecl:
		return TypeT, d.Group
	case *VarDecl:
		return VarT, d.Group
	case *FuncDecl:
		return FuncT, nil
	case *CommentDecl:
		return CommentT, d.Group
	default:
		panic("unreachable")
	}
}

//================================================================================
// project-level
// ----------------------------------------------------------------------------
func (n Project) Print(p printer) {}
func (n Package) Print(p printer) {}

//--------------------------------------------------------------------------------
func (n File) Print(p printer) {
	p.Print(PackageT, BlankSym, n.PkgName)
	if len(n.DeclList) > 0 {
		p.Print(SemiT, NewlineSym, NewlineSym)
		printDeclList(p, n.DeclList)
	}
}

//- - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
func printDeclList(p printer, list []Decl) {
	i0 := 0
	var tok Token
	var group *DeclGroup
	for i, x := range list {
		if s, g := groupFor(x); g == nil || g != group {
			if i0 < i {
				printDecl(p, list[i0:i])
				p.Print(SemiT, NewlineSym)
				// print empty line between different declaration groups,
				// different kinds of declarations, or between functions
				if g != group || s != tok || s == FuncT {
					p.Print(NewlineSym)
				}
				i0 = i
			}
			tok, group = s, g
		}
	}
	printDecl(p, list[i0:])
}

//================================================================================
// declarations
//--------------------------------------------------------------------------------
func (n DeclGroup) Print(p printer) {
	p.Print(n.Tok, BlankSym, LparenT)
	if len(n.Decls) > 0 {
		p.Print(NewlineSym, IndentSym)
		for _, d := range n.Decls {
			p.PrintNode(d)
			p.Print(SemiT, NewlineSym)
		}
		p.Print(OutdentSym)
	}
	p.Print(RparenT)
}

//--------------------------------------------------------------------------------
func (n ImportDecl) Print(p printer) {
	if n.Group == nil {
		p.Print(ImportT, BlankSym)
	}
	if n.LocalPkgName != nil {
		p.Print(n.LocalPkgName, BlankSym)
	}
	p.Print(n.Path)
}

//--------------------------------------------------------------------------------
func (n ConstDecl) Print(p printer) {
	if n.Group == nil {
		p.Print(ConstT, BlankSym)
	}
	printNameList(p, n.NameList)
	if n.Type != nil {
		p.Print(BlankSym, n.Type)
	}
	if n.Values != nil {
		p.Print(BlankSym, AssignT, BlankSym, n.Values)
	}
}

//--------------------------------------------------------------------------------
func (n VarDecl) Print(p printer) {
	if n.Group == nil {
		p.Print(VarT, BlankSym)
	}
	printNameList(p, n.NameList)
	if n.Type != nil {
		p.Print(BlankSym, n.Type)
	}
	if n.Values != nil {
		p.Print(BlankSym, AssignT, BlankSym, n.Values)
	}
}

//--------------------------------------------------------------------------------
func (n TypeDecl) Print(p printer) {
	if n.Group == nil {
		p.Print(TypeT, BlankSym)
	}
	p.Print(n.Name, BlankSym)
	if n.Alias {
		p.Print(AssignT, BlankSym)
	}
	p.Print(n.Type)
}

//--------------------------------------------------------------------------------
func (n FuncDecl) Print(p printer) {
	p.Print(FuncT, BlankSym)
	if r := n.Recv; r != nil {
		p.Print(LparenT)
		if r.Name != nil {
			p.Print(r.Name, BlankSym)
		}
		p.PrintNode(r.Type)
		p.Print(RparenT, BlankSym)
	}
	p.Print(n.Name)
	printSignature(p, n.Type)
	if n.Body != nil {
		p.Print(BlankSym, n.Body)
	}
}

//--------------------------------------------------------------------------------
func (n CommentDecl) Print(p printer) {
	for _, c := range n.CommentList {
		p.Print(NewlineSym, c, CommentT)
	}
}

//--------------------------------------------------------------------------------
func (n EmptyDecl) Print(p printer) {
	// nothing to print
}

//================================================================================
// statements
//--------------------------------------------------------------------------------
func (n CommentStmt) Print(p printer) {
	// nothing to print ???
}

//--------------------------------------------------------------------------------
func (n LabeledStmt) Print(p printer) {
	p.Print(OutdentSym, n.Label, ColonT, IndentSym, NewlineSym, n.Stmt)
}

//--------------------------------------------------------------------------------
func (n BlockStmt) Print(p printer) {
	p.Print(LbraceT)
	if len(n.List) > 0 {
		p.Print(NewlineSym, IndentSym)
		printStmtList(p, n.List, true)
		p.Print(OutdentSym, NewlineSym)
	}
	p.Print(RbraceT)
}

//--------------------------------------------------------------------------------
func (n ExprStmt) Print(p printer) {
	p.Print(n.X)
}

//--------------------------------------------------------------------------------
func (n SendStmt) Print(p printer) {
	p.Print(n.Chan, BlankSym, ArrowT, BlankSym, n.Value)
}

//--------------------------------------------------------------------------------
func (n DeclStmt) Print(p printer) {
	printDecl(p, n.DeclList)
}

//--------------------------------------------------------------------------------
func (n AssignStmt) Print(p printer) {
	p.Print(n.Lhs)
	if n.Rhs == ImplicitOne {
		// TODO(gri) This is going to break the mayCombine
		//           check once we enable that again.
		p.Print(n.Op, n.Op) // ++ or --
	} else {
		p.Print(BlankSym, n.Op, AssignT, BlankSym)
		p.Print(n.Rhs)
	}
}

//--------------------------------------------------------------------------------
func (n BranchStmt) Print(p printer) {
	p.Print(n.Tok)
	if n.Label != nil {
		p.Print(BlankSym, n.Label)
	}
}

//--------------------------------------------------------------------------------
func (n CallStmt) Print(p printer) {
	p.Print(n.Tok, BlankSym, n.Call)
}

//--------------------------------------------------------------------------------
func (n ReturnStmt) Print(p printer) {
	p.Print(ReturnT)
	if n.Results != nil {
		p.Print(BlankSym, n.Results)
	}
}

//--------------------------------------------------------------------------------
func (n IfStmt) Print(p printer) {
	p.Print(IfT, BlankSym)
	if n.Init != nil {
		p.Print(n.Init, SemiT, BlankSym)
	}
	p.Print(n.Cond, BlankSym, n.Then)
	if n.Else != nil {
		p.Print(BlankSym, ElseT, BlankSym, n.Else)
	}
}

//--------------------------------------------------------------------------------
func (n RangeClause) Print(p printer) {
	if n.Lhs != nil {
		tok := AssignT
		if n.Def {
			tok = DefineT
		}
		p.Print(n.Lhs, BlankSym, tok, BlankSym)
	}
	p.Print(RangeT, BlankSym, n.X)
}

//--------------------------------------------------------------------------------
func (n ForStmt) Print(p printer) {
	p.Print(ForT, BlankSym)
	if n.Init == nil && n.Post == nil {
		if n.Cond != nil {
			p.Print(n.Cond, BlankSym)
		}
	} else {
		if n.Init != nil {
			p.Print(n.Init)
			if _, ok := n.Init.(*RangeClause); ok {
				p.Print(BlankSym, n.Body)
				return //TODO: clean this up???
			}
		}
		p.Print(SemiT, BlankSym)
		if n.Cond != nil {
			p.Print(n.Cond)
		}
		p.Print(SemiT, BlankSym)
		if n.Post != nil {
			p.Print(n.Post, BlankSym)
		}
	}
	p.Print(n.Body)
}

//--------------------------------------------------------------------------------
func (n TypeSwitchGuard) Print(p printer) {
	if n.Lhs != nil {
		p.Print(n.Lhs, BlankSym, DefineT, BlankSym)
	}
	p.Print(n.X, DotT, LparenT, TypeT, RparenT)
}

//--------------------------------------------------------------------------------
func (n SwitchStmt) Print(p printer) {
	p.Print(SwitchT, BlankSym)
	if n.Init != nil {
		p.Print(n.Init, SemiT, BlankSym)
	}
	if n.Tag != nil {
		p.Print(n.Tag, BlankSym)
	}
	p.Print(LbraceT)
	if len(n.Body) > 0 {
		p.Print(NewlineSym)
		for _, c := range n.Body {
			p.Print(c, NewlineSym)
		}
	}
	p.Print(RbraceT)
}

//--------------------------------------------------------------------------------
func (c CaseClause) Print(p printer) {
	if c.Cases != nil {
		p.Print(CaseT, BlankSym, c.Cases)
	} else {
		p.Print(DefaultT)
	}
	p.Print(ColonT)
	if len(c.Body) > 0 {
		p.Print(NewlineSym, IndentSym)
		printStmtList(p, c.Body, c.Final)
		p.Print(OutdentSym)
	}
}

//--------------------------------------------------------------------------------
func (n SelectStmt) Print(p printer) {
	p.Print(SelectT, BlankSym, LbraceT)
	if len(n.Body) > 0 {
		p.Print(NewlineSym)
		for _, c := range n.Body {
			p.Print(c, NewlineSym)
		}
	}
	p.Print(RbraceT)
}

//--------------------------------------------------------------------------------
func (c CommClause) Print(p printer) {
	if c.Comm != nil {
		p.Print(CaseT, BlankSym)
		p.Print(c.Comm)
	} else {
		p.Print(DefaultT)
	}
	p.Print(ColonT)
	if len(c.Body) > 0 {
		p.Print(NewlineSym, IndentSym)
		printStmtList(p, c.Body, c.Final)
		p.Print(OutdentSym)
	}
}

//--------------------------------------------------------------------------------
func (n EmptyStmt) Print(p printer) {
	// nothing to print
}

//================================================================================
// expressions
//--------------------------------------------------------------------------------
func (n BadExpr) Print(p printer) {
	p.Print(NameT, "<bad expr>")
}

//--------------------------------------------------------------------------------
func (n Name) Print(p printer) {
	p.Print(NameT, n.Value) // NameT requires actual value following immediately
}

//--------------------------------------------------------------------------------
func (n BasicLit) Print(p printer) {
	p.Print(NameT, n.Value) // NameT requires actual value following immediately
}

//--------------------------------------------------------------------------------
func (n CompositeLit) Print(p printer) {
	if n.Type != nil {
		p.Print(n.Type)
	}
	p.Print(LbraceT)
	if n.NKeys > 0 && n.NKeys == len(n.ElemList) {
		printExprLines(p, n.ElemList)
	} else {
		printExprList(p, n.ElemList)
	}
	p.Print(RbraceT)
}

//- - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
func printExprLines(p printer, list []Expr) {
	if len(list) > 0 {
		p.Print(NewlineSym, IndentSym)
		for _, x := range list {
			p.Print(x, CommaT, NewlineSym)
		}
		p.Print(OutdentSym)
	}
}

//--------------------------------------------------------------------------------
func (n KeyValueExpr) Print(p printer) {
	p.Print(n.Key, ColonT, BlankSym, n.Value)
}

//--------------------------------------------------------------------------------
func (n FuncLit) Print(p printer) {
	p.Print(n.Type, BlankSym, n.Body)
}

//--------------------------------------------------------------------------------
func (n ParenExpr) Print(p printer) {
	p.Print(LparenT, n.X, RparenT)
}

//--------------------------------------------------------------------------------
func (n RhsExpr) Print(p printer) {
	p.Print(n.X)
}

//--------------------------------------------------------------------------------
func (n SelectorExpr) Print(p printer) {
	p.Print(n.X, DotT, n.Sel)
}

//--------------------------------------------------------------------------------
func (n IndexExpr) Print(p printer) {
	p.Print(n.X, LbrackT, n.Index, RbrackT)
}

//--------------------------------------------------------------------------------
func (n SliceExpr) Print(p printer) {
	p.Print(n.X, LbrackT)
	if i := n.Index[0]; i != nil {
		p.PrintNode(i)
	}
	p.Print(ColonT)
	if j := n.Index[1]; j != nil {
		p.PrintNode(j)
	}
	if k := n.Index[2]; k != nil {
		p.Print(ColonT, k)
	}
	p.Print(RbrackT)
}

//--------------------------------------------------------------------------------
func (n AssertExpr) Print(p printer) {
	p.Print(n.X, DotT, LparenT)
	if n.Type != nil {
		p.PrintNode(n.Type)
	} else {
		p.Print(TypeT)
	}
	p.Print(RparenT)
}

//--------------------------------------------------------------------------------
func (n Operation) Print(p printer) {
	if n.Y == nil {
		// unary expr
		p.Print(n.Op)
		// if n.Op == lexical.Range {
		// 	p.print(BlankSym)
		// }
		p.Print(n.X)
	} else {
		// binary expr
		// TODO(gri) eventually take precedence into account
		// to control possibly missing parentheses
		p.Print(n.X, BlankSym, n.Op, BlankSym, n.Y)
	}
}

//--------------------------------------------------------------------------------
func (n CallExpr) Print(p printer) {
	p.Print(n.Fun, LparenT)
	printExprList(p, n.ArgList)
	if n.HasDots {
		p.Print(DotDotDotT)
	}
	p.Print(RparenT)
}

//--------------------------------------------------------------------------------
func (n ListExpr) Print(p printer) {
	printExprList(p, n.ElemList)
}

//--------------------------------------------------------------------------------
func (n Field) Print(p printer) {
	// nothing to print yet
}

//================================================================================
// types
//--------------------------------------------------------------------------------
func (n ArrayType) Print(p printer) {
	var len interface{} = DotDotDotT
	if n.Len != nil {
		len = n.Len
	}
	p.Print(LbrackT, len, RbrackT, n.Elem)
}

//--------------------------------------------------------------------------------
func (n SliceType) Print(p printer) {
	p.Print(LbrackT, RbrackT, n.Elem)
}

//--------------------------------------------------------------------------------
func (n DotsType) Print(p printer) {
	p.Print(DotDotDotT, n.Elem)
}

//--------------------------------------------------------------------------------
func (n StructType) Print(p printer) {
	p.Print(StructT)
	if len(n.FieldList) > 0 && p.Linebreaks() {
		p.Print(BlankSym)
	}
	p.Print(LbraceT)
	if len(n.FieldList) > 0 {
		p.Print(NewlineSym, IndentSym)
		printFieldList(p, n.FieldList, n.TagList)
		p.Print(OutdentSym, NewlineSym)
	}
	p.Print(RbraceT)
}

//- - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
func printFieldList(p printer, fields []*Field, tags []*BasicLit) {
	i0 := 0
	var typ Expr
	for i, f := range fields {
		if f.Name == nil || f.Type != typ {
			if i0 < i {
				printFields(p, fields, tags, i0, i)
				p.Print(SemiT, NewlineSym)
				i0 = i
			}
			typ = f.Type
		}
	}
	printFields(p, fields, tags, i0, len(fields))
}

//- - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
func printFields(p printer, fields []*Field, tags []*BasicLit, i, j int) {
	if i+1 == j && fields[i].Name == nil {
		// anonymous field
		p.PrintNode(fields[i].Type)
	} else {
		for k, f := range fields[i:j] {
			if k > 0 {
				p.Print(CommaT, BlankSym)
			}
			p.PrintNode(f.Name)
		}
		p.Print(BlankSym)
		p.PrintNode(fields[i].Type)
	}
	if i < len(tags) && tags[i] != nil {
		p.Print(BlankSym)
		p.PrintNode(tags[i])
	}
}

//--------------------------------------------------------------------------------
func (n FuncType) Print(p printer) {
	p.Print(FuncT)
	printSignature(p, &n)
}

//--------------------------------------------------------------------------------
func (n InterfaceType) Print(p printer) {
	p.Print(InterfaceT)
	if len(n.MethodList) > 0 && p.Linebreaks() {
		p.Print(BlankSym)
	}
	p.Print(LbraceT)
	if len(n.MethodList) > 0 {
		p.Print(NewlineSym, IndentSym)
		printMethodList(p, n.MethodList)
		p.Print(OutdentSym, NewlineSym)
	}
	p.Print(RbraceT)
}

//- - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
func printMethodList(p printer, methods []*Field) {
	for i, m := range methods {
		if i > 0 {
			p.Print(SemiT, NewlineSym)
		}
		if m.Name != nil {
			p.PrintNode(m.Name)
			printSignature(p, m.Type.(*FuncType))
		} else {
			p.PrintNode(m.Type)
		}
	}
}

//--------------------------------------------------------------------------------
func (n MapType) Print(p printer) {
	p.Print(MapT, LbrackT, n.Key, RbrackT, n.Value)
}

//--------------------------------------------------------------------------------
func (n ChanType) Print(p printer) {
	if n.Dir == RecvOnly {
		p.Print(ArrowT)
	}
	p.Print(ChanT)
	if n.Dir == SendOnly {
		p.Print(ArrowT)
	}
	p.Print(BlankSym, n.Elem)
}

//================================================================================
// node print utilities
//--------------------------------------------------------------------------------
func printDecl(p printer, list []Decl) {
	tok, group := groupFor(list[0])

	if group == nil {
		if len(list) != 1 {
			panic("unreachable")
		}
		p.PrintNode(list[0])
		return
	}

	if _, ok := list[0].(*EmptyDecl); ok {
		if len(list) != 1 {
			panic("unreachable")
		}
		// TODO(gri) if there are comments inside the empty
		// group, we may need to keep the list non-nil
		list = nil
	}

	// for consistent comment handling
	group.Tok = tok
	group.Decls = list
	p.PrintNode(group)
}

//--------------------------------------------------------------------------------
func printNameList(p printer, list []*Name) {
	for i, x := range list {
		if i > 0 {
			p.Print(CommaT, BlankSym)
		}
		p.PrintNode(x)
	}
}

//--------------------------------------------------------------------------------
func printSignature(p printer, sig *FuncType) {
	printParameterList(p, sig.ParamList)
	if list := sig.ResultList; list != nil {
		p.Print(BlankSym)
		if len(list) == 1 && list[0].Name == nil {
			p.PrintNode(list[0].Type)
		} else {
			printParameterList(p, list)
		}
	}
}

//--------------------------------------------------------------------------------
func printParameterList(p printer, list []*Field) {
	p.Print(LparenT)
	if len(list) > 0 {
		for i, f := range list {
			if i > 0 {
				p.Print(CommaT, BlankSym)
			}
			if f.Name != nil {
				p.PrintNode(f.Name)
				if i+1 < len(list) {
					f1 := list[i+1]
					if f1.Name != nil && f1.Type == f.Type {
						continue // no need to print type
					}
				}
				p.Print(BlankSym)
			}
			p.PrintNode(f.Type)
		}
	}
	p.Print(RparenT)
}

//--------------------------------------------------------------------------------
func printStmtList(p printer, list []Stmt, braces bool) {
	for i, x := range list {
		p.Print(x, SemiT)
		if i+1 < len(list) {
			p.Print(NewlineSym)
		} else if braces {
			// Print an extra semicolon if the last statement is
			// an empty statement and we are in a braced block
			// because one semicolon is automatically removed.
			if _, ok := x.(*EmptyStmt); ok {
				p.Print(x, SemiT)
			}
		}
	}
}

//--------------------------------------------------------------------------------
func printExprList(p printer, list []Expr) {
	for i, x := range list {
		if i > 0 {
			p.Print(CommaT, BlankSym)
		}
		p.PrintNode(x)
	}
}

//================================================================================
