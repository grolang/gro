// Copyright 2016 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

//Package for macroized command 鲜.
package fresh

import (
	"github.com/grolang/gro/ast"
	//"github.com/grolang/gro/token"
	"github.com/grolang/gro/macro"
)

type Fresh struct{}

func (m Fresh) Init(p macro.Parser) {
}

func (m Fresh) Main(p macro.Parser) ast.Stmt {
	p.Next()
	/*
	p.CheckOrConvertIdentifier()
	p.OpenScope()
	ss, _ := p.ParseSimpleStmt(macro.Basic)
	var list []ast.Stmt
	list = append(list, ss)
	for p.Tok() != token.CASE && p.Tok() != token.DEFAULT && p.Tok() != token.RBRACE && p.Tok() != token.EOF {
		list = append(list, p.ParseStmt())
	}
	p.CloseScope()
	s:= &ast.BlockStmt{Lbrace: token.NoPos, List: list, Rbrace: token.NoPos}
	return s
	*/
	return &ast.EmptyStmt{Semicolon: p.Pos(), Implicit: true}
}

