// Copyright 2016 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package macro

import (
	"go/token"
	"go/ast"
)

type StmtMacro interface {
	Init(Parser)
	Main(Parser) ast.Stmt
}

type ExprMacro interface {
	Init(Parser)
	Main(Parser) ast.Expr
}

type Scanner interface{
	Src() []byte
	Ch() rune
	Offset() int
	Next()
	Error(int, string)
}

// Parsing modes for ParseSimpleStmt.
const (
	Basic = iota
	LabelOk
	RangeOk
)

type Parser interface{
	Pos() token.Pos
	Tok() token.Token
	Lit() string

	AppendUnihanImports(*ast.ImportSpec)
	ErrorExpected(token.Pos, string)
	SetTok(token.Token)
	Next()
	CheckExpr(ast.Expr) ast.Expr

	OpenScope()
	CloseScope()
	CheckOrConvertIdentifier()
	ParseSimpleStmt(mode int) (ast.Stmt, bool)

	TryIdentOrType() ast.Expr
	ParseExpr(bool) ast.Expr

	ParseStmt() ast.Stmt

	AddSpecialProc(token.Token, func(Parser)ast.Expr)
	AddSpecialUnary(token.Token, func(Parser, ast.Expr)ast.Expr)
	AddSpecialBinary(token.Token, func(Parser, ast.Expr, ast.Expr)ast.Expr)
	AddSpecialRuneProc(func(Scanner) string)

	DeleteSpecialProc(token.Token)
	DeleteSpecialUnary(token.Token)
	DeleteSpecialBinary(token.Token)
	DeleteSpecialRuneProc()

	ProcBlacklist(string)
}

