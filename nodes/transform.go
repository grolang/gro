// Copyright 2016-17 The Go and Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nodes

import (
	"fmt"
)

type Walker struct {
	inRhs         bool
	shortFormFunc struct {
		active bool
		count  int
	}
}

// ----------------------------------------------------------------------------
// Files

func (proj *Project) Walk(w *Walker) Node {
	for i, pkg := range proj.Pkgs {
		proj.Pkgs[i] = pkg.Walk(w).(*Package)
	}
	//don't walk the ArgImports
	return proj
}

func (pkg *Package) Walk(w *Walker) Node {
	//don't walk the Params yet
	for i, file := range pkg.Files {
		pkg.Files[i] = file.Walk(w).(*File)
	}
	return pkg
}

func (file *File) Walk(w *Walker) Node {
	//don't walk InfImports/InfImpMap yet
	for i, decl := range file.DeclList {
		file.DeclList[i] = decl.Walk(w).(Decl)
	}
	return file
}

// ----------------------------------------------------------------------------
// Decls

func (g *DeclGroup) Walk(w *Walker) Node {
	//do nothing
	return g
}

func (decl *CommentDecl) Walk(w *Walker) Node {
	return decl
}

func (decl *EmptyDecl) Walk(w *Walker) Node {
	return decl
}

func (decl *ImportDecl) Walk(w *Walker) Node {
	//don't walk the Group
	//don't walk LocalPkgName, Path, Args, Infers yet
	return decl
}

func (decl *TypeDecl) Walk(w *Walker) Node {
	decl.Name = decl.Name.Walk(w).(*Name)
	decl.Type = decl.Type.Walk(w).(Expr)
	return decl
}

func (decl *VarDecl) Walk(w *Walker) Node {
	for i, name := range decl.NameList {
		decl.NameList[i] = name.Walk(w).(*Name)
	}
	if decl.Type != nil {
		decl.Type = decl.Type.Walk(w).(Expr)
	}
	if decl.Values != nil {
		decl.Values = decl.Values.Walk(w).(Expr)
	}
	return decl
}

func (decl *ConstDecl) Walk(w *Walker) Node {
	for i, name := range decl.NameList {
		decl.NameList[i] = name.Walk(w).(*Name)
	}
	if decl.Type != nil {
		decl.Type = decl.Type.Walk(w).(Expr)
	}
	if decl.Values != nil {
		decl.Values = decl.Values.Walk(w).(Expr)
	}
	return decl
}

func (decl *FuncDecl) Walk(w *Walker) Node {
	//don't walk Attr -- yet???
	if decl.Recv != nil {
		decl.Recv = decl.Recv.Walk(w).(*Field)
	}
	decl.Name = decl.Name.Walk(w).(*Name)
	decl.Type = decl.Type.Walk(w).(*FuncType)
	if decl.Body != nil {
		decl.Body = decl.Body.Walk(w).(*BlockStmt)
	}
	return decl
}

// ----------------------------------------------------------------------------
// Stmts

func (stmt *EmptyStmt) Walk(w *Walker) Node {
	return stmt
}

func (stmt *CommentStmt) Walk(w *Walker) Node {
	return stmt
}

func (stmt *LabeledStmt) Walk(w *Walker) Node {
	stmt.Label = stmt.Label.Walk(w).(*Name)
	stmt.Stmt = stmt.Stmt.Walk(w).(Stmt)
	return stmt
}

func (block *BlockStmt) Walk(w *Walker) Node {
	for i, stmt := range block.List {
		block.List[i] = stmt.Walk(w).(Stmt)
	}
	return block
}

func (stmt *ExprStmt) Walk(w *Walker) Node {
	stmt.X = stmt.X.Walk(w).(Expr)
	return stmt
}

func (stmt *SendStmt) Walk(w *Walker) Node {
	stmt.Chan = stmt.Chan.Walk(w).(Expr)
	stmt.Value = stmt.Value.Walk(w).(Expr)
	return stmt
}

func (stmt *DeclStmt) Walk(w *Walker) Node {
	for i, decl := range stmt.DeclList {
		stmt.DeclList[i] = decl.Walk(w).(Decl)
	}
	return stmt
}

func (stmt *AssignStmt) Walk(w *Walker) Node {
	stmt.Lhs = stmt.Lhs.Walk(w).(Expr)
	stmt.Rhs = stmt.Rhs.Walk(w).(Expr)
	return stmt
}

func (stmt *BranchStmt) Walk(w *Walker) Node {
	if stmt.Label != nil {
		stmt.Label = stmt.Label.Walk(w).(*Name)
	}
	//don't walk the target
	return stmt
}

func (stmt *CallStmt) Walk(w *Walker) Node {
	stmt.Call = stmt.Call.Walk(w).(*CallExpr)
	return stmt
}

func (stmt *ReturnStmt) Walk(w *Walker) Node {
	if stmt.Results != nil {
		stmt.Results = stmt.Results.Walk(w).(Expr)
	}
	return stmt
}

func (stmt *IfStmt) Walk(w *Walker) Node {
	if stmt.Init != nil {
		stmt.Init = stmt.Init.Walk(w).(SimpleStmt)
	}
	stmt.Cond = stmt.Cond.Walk(w).(Expr)
	stmt.Then = stmt.Then.Walk(w).(*BlockStmt)
	if stmt.Else != nil {
		stmt.Else = stmt.Else.Walk(w).(Stmt)
	}
	return stmt
}

func (stmt *ForStmt) Walk(w *Walker) Node {
	if stmt.Init != nil {
		stmt.Init = stmt.Init.Walk(w).(SimpleStmt)
	}
	if stmt.Cond != nil {
		stmt.Cond = stmt.Cond.Walk(w).(Expr)
	}
	if stmt.Post != nil {
		stmt.Post = stmt.Post.Walk(w).(SimpleStmt)
	}
	stmt.Body = stmt.Body.Walk(w).(*BlockStmt)
	return stmt
}

func (stmt *RangeClause) Walk(w *Walker) Node {
	if stmt.Lhs != nil {
		stmt.Lhs = stmt.Lhs.Walk(w).(Expr)
	}
	stmt.X = stmt.X.Walk(w).(Expr)
	return stmt
}

func (stmt *SwitchStmt) Walk(w *Walker) Node {
	if stmt.Init != nil {
		stmt.Init = stmt.Init.Walk(w).(SimpleStmt)
	}
	if stmt.Tag != nil {
		stmt.Tag = stmt.Tag.Walk(w).(Expr)
	}
	for i, clause := range stmt.Body {
		stmt.Body[i] = clause.Walk(w).(*CaseClause)
	}
	return stmt
}

func (expr *TypeSwitchGuard) Walk(w *Walker) Node {
	if expr.Lhs != nil {
		expr.Lhs = expr.Lhs.Walk(w).(*Name)
	}
	expr.X = expr.X.Walk(w).(Expr)
	return expr
}

func (clause *CaseClause) Walk(w *Walker) Node {
	if clause.Cases != nil {
		clause.Cases = clause.Cases.Walk(w).(Expr)
	}
	for i, stmt := range clause.Body {
		clause.Body[i] = stmt.Walk(w).(Stmt)
	}
	return clause
}

func (stmt *SelectStmt) Walk(w *Walker) Node {
	for i, clause := range stmt.Body {
		stmt.Body[i] = clause.Walk(w).(*CommClause)
	}
	return stmt
}

func (clause *CommClause) Walk(w *Walker) Node {
	if clause.Comm != nil {
		clause.Comm.Walk(w)
	}
	for i, stmt := range clause.Body {
		clause.Body[i] = stmt.Walk(w).(Stmt)
	}
	return clause
}

// ----------------------------------------------------------------------------
// Exprs

func (expr *BadExpr) Walk(w *Walker) Node {
	return expr
}

func (expr *Name) Walk(w *Walker) Node {
	if expr.Value == "_" {
		//if w.shortFormFunc.active || w.inRhs {
		//fmt.Printf(">>> %s, sff:%v, inRhs:%t\n", expr.Value, w.shortFormFunc, w.inRhs)
	}
	if expr.Value == "_" && w.shortFormFunc.active && w.inRhs {
		e := &IndexExpr{
			X:     &Name{Value: "groo_it"},
			Index: &BasicLit{Value: fmt.Sprint(w.shortFormFunc.count), Kind: IntLit},
		}
		w.shortFormFunc.count++
		return e
	} else {
		return expr
	}
}

func (expr *BasicLit) Walk(w *Walker) Node {
	return expr
}

func (expr *CompositeLit) Walk(w *Walker) Node {
	if expr.Type != nil {
		expr.Type = expr.Type.Walk(w).(Expr)
	}
	for i, elem := range expr.ElemList {
		expr.ElemList[i] = elem.Walk(w).(Expr)
	}
	return expr
}

func (expr *KeyValueExpr) Walk(w *Walker) Node {
	expr.Key = expr.Key.Walk(w).(Expr)
	expr.Value = expr.Value.Walk(w).(Expr)
	return expr
}

func (expr *FuncLit) Walk(w *Walker) Node {
	expr.Type = expr.Type.Walk(w).(*FuncType)

	oldShortFormFunc := w.shortFormFunc
	w.shortFormFunc = struct {
		active bool
		count  int
	}{active: true, count: 0}

	expr.Body = expr.Body.Walk(w).(*BlockStmt)

	w.shortFormFunc = oldShortFormFunc
	return expr
}

func (expr *ParenExpr) Walk(w *Walker) Node {
	expr.X = expr.X.Walk(w).(Expr)
	return expr
}

func (expr *RhsExpr) Walk(w *Walker) Node {
	oldRhs := w.inRhs
	w.inRhs = true

	newX := expr.X.Walk(w)

	w.inRhs = oldRhs
	return newX
}

func (expr *SelectorExpr) Walk(w *Walker) Node {
	expr.X = expr.X.Walk(w).(Expr)
	expr.Sel = expr.Sel.Walk(w).(*Name)
	return expr
}

func (expr *IndexExpr) Walk(w *Walker) Node {
	expr.X = expr.X.Walk(w).(Expr)
	expr.Index = expr.Index.Walk(w).(Expr)
	return expr
}

func (expr *SliceExpr) Walk(w *Walker) Node {
	expr.X = expr.X.Walk(w).(Expr)
	if len(expr.Index) > 0 && expr.Index[0] != nil {
		expr.Index[0] = expr.Index[0].Walk(w).(Expr)
	}
	if len(expr.Index) > 1 && expr.Index[1] != nil {
		expr.Index[1] = expr.Index[1].Walk(w).(Expr)
	}
	if expr.Full && len(expr.Index) > 2 && expr.Index[2] != nil {
		expr.Index[2] = expr.Index[2].Walk(w).(Expr)
	}
	return expr
}

func (expr *AssertExpr) Walk(w *Walker) Node {
	expr.X = expr.X.Walk(w).(Expr)
	expr.Type = expr.Type.Walk(w).(Expr)
	return expr
}

func (expr *Operation) Walk(w *Walker) Node {
	expr.X = expr.X.Walk(w).(Expr)
	if expr.Y != nil {
		expr.Y = expr.Y.Walk(w).(Expr)
	}
	return expr
}

func (expr *CallExpr) Walk(w *Walker) Node {
	expr.Fun = expr.Fun.Walk(w).(Expr)
	for i, arg := range expr.ArgList {
		expr.ArgList[i] = arg.Walk(w).(Expr)
	}
	return expr
}

func (expr *ListExpr) Walk(w *Walker) Node {
	for i, elem := range expr.ElemList {
		expr.ElemList[i] = elem.Walk(w).(Expr)
	}
	return expr
}

func (field *Field) Walk(w *Walker) Node {
	if field.Name != nil {
		field.Name = field.Name.Walk(w).(*Name)
	}
	field.Type = field.Type.Walk(w).(Expr)
	return field
}

// ----------------------------------------------------------------------------
// Types

func (expr *ArrayType) Walk(w *Walker) Node {
	if expr.Len != nil {
		expr.Len = expr.Len.Walk(w).(Expr)
	}
	expr.Elem = expr.Elem.Walk(w).(Expr)
	return expr
}

func (expr *SliceType) Walk(w *Walker) Node {
	expr.Elem = expr.Elem.Walk(w).(Expr)
	return expr
}

func (expr *DotsType) Walk(w *Walker) Node {
	expr.Elem = expr.Elem.Walk(w).(Expr)
	return expr
}

func (expr *StructType) Walk(w *Walker) Node {
	for i, field := range expr.FieldList {
		expr.FieldList[i] = field.Walk(w).(*Field)
	}
	for i, tag := range expr.TagList {
		expr.TagList[i] = tag.Walk(w).(*BasicLit)
	}
	return expr
}

func (expr *FuncType) Walk(w *Walker) Node {
	for i, param := range expr.ParamList {
		expr.ParamList[i] = param.Walk(w).(*Field)
	}
	for i, result := range expr.ResultList {
		expr.ResultList[i] = result.Walk(w).(*Field)
	}
	return expr
}

func (expr *InterfaceType) Walk(w *Walker) Node {
	for i, method := range expr.MethodList {
		expr.MethodList[i] = method.Walk(w).(*Field)
	}
	return expr
}

func (expr *MapType) Walk(w *Walker) Node {
	expr.Key = expr.Key.Walk(w).(Expr)
	expr.Value = expr.Value.Walk(w).(Expr)
	return expr
}

func (expr *ChanType) Walk(w *Walker) Node {
	expr.Elem = expr.Elem.Walk(w).(Expr)
	return expr
}

// ----------------------------------------------------------------------------
