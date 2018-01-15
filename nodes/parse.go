// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nodes

import (
	"github.com/grolang/gro/syntax/src"
)

//--------------------------------------------------------------------------------
type GeneralParser interface {
	Proj(string) *Project
	ProjFromNewParser(string, []byte) (*Project, error)
	UseDecl() *Project
	InclDecl() *Project
	PkgOrNil() *Package
	SectionOrNil(*File) *File

	DeclParser
	StmtParser
	ExprParser
	TypeParser

	Got(Token) bool
	Want(tok Token)
	NewName(string) *Name
	Name() *Name
	NameList(*Name) []*Name
	QualifiedName(*Name) Expr
	OLiteral() *BasicLit

	NewBlankFunc(string) *FuncDecl
	NewBlankFuncLit() *FuncLit
	ProcImportAlias(*BasicLit, string) string
	IsName(ss ...string) bool
	Advance(...Token)
	List(Token, Token, Token, func() bool) src.Pos

	Pos() src.Pos
	Error(string)
	SyntaxError(string)
	PosAt(uint, uint) src.Pos
	ErrorAt(src.Pos, string)
	SyntaxErrorAt(src.Pos, string)

	DynamicParser
	ScannerState
	StmtRegistryParser
	PermitParser
	LineDirectiveParser
}

type ScannerState interface {
	Next()
	Tok() Token
	Lit() string
	Kind() LitKind
	Op() Operator
	Prec() Prec
}

type DeclParser interface {
	Decl([]Decl) []Decl
	AppendGroup([]Decl, func(*DeclGroup) Decl) []Decl
	ImportDecl(*DeclGroup) Decl
	ConstDecl(*DeclGroup) Decl
	TypeDecl(*DeclGroup) Decl
	VarDecl(*DeclGroup) Decl
	FuncDeclOrNil(func() Stmt) *FuncDecl
}

type StmtParser interface {
	TlBlock() *FuncDecl
	TlStmtList(func() Stmt) []Stmt
	TlStmt() Stmt
	ProcStmt() Stmt
	BlockStmt(string, func() Stmt) *BlockStmt
	StmtList(func() Stmt) []Stmt
	StmtOrNil() Stmt
	SimpleStmt(Expr, bool) SimpleStmt
	NewRangeClause(Expr, bool) *RangeClause
	NewAssignStmt(src.Pos, Operator, Expr, Expr) *AssignStmt
	LabeledStmtOrNil(*Name) Stmt
	DeclStmt(func(*DeclGroup) Decl) *DeclStmt
	BreakOrContinueStmt() *BranchStmt
	GotoStmt() *BranchStmt
	ReturnStmt() *ReturnStmt
	CallStmt(func() Stmt) *CallStmt
	IfStmt(func() Stmt) *IfStmt
	ForStmt(func() Stmt) Stmt
	SwitchStmt(func() Stmt) *SwitchStmt
	CaseClause(func() Stmt) *CaseClause
	SelectStmt(func() Stmt) *SelectStmt
	CommClause(func() Stmt) *CommClause
}

type ExprParser interface {
	Expr() Expr
	ExprList(bool) Expr
	ArgList() ([]Expr, bool)
	BinaryExpr(Prec) Expr
	UnaryExpr() Expr
	Operand(bool) Expr
	PExpr(bool) Expr
	BareCompLitExpr() Expr
	CompLitExpr() *CompositeLit
	BadExpr() *BadExpr
}

type TypeParser interface {
	Type() Expr
	TypeOrNil() Expr
	FuncType() *FuncType
	ChanElem() Expr
	DotName(*Name) Expr
	StructType() *StructType
	InterfaceType() *InterfaceType

	//TODO: put these in right place...
	FuncBody(func() Stmt) *BlockStmt
	FuncResult() []*Field
	AddField(*StructType, src.Pos, *Name, Expr, *BasicLit)
	FieldDecl(styp *StructType)
	MethodDecl() *Field
	ParamDeclOrNil() *Field
	DotsType() *DotsType
	ParamList() []*Field
}

type DynamicParser interface {
	DynamicBlock() string
	SetDynamicBlock(string)
	DynamicMode() bool
	SetDynamicMode(bool)
	DynCharSet() string
	SetDynCharSet(string)
}

type PermitParser interface {
	SetPermit(string)
	UnsetPermit(string)
	IsPermit(string) bool
}

type LineDirectiveParser interface {
	LineDirectives() bool
	SetLineDirectives(bool)
}

type StmtRegistryParser interface {
	SetStmtRegistry(string, func(GeneralParser, ...interface{}) Stmt)
	UnsetStmtRegistry(string)
}

//--------------------------------------------------------------------------------
