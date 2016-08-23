// Copyright 2016 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

//Package for macroized expressions such as 叫.
package expr

import (
	"go/ast"
	"go/token"
	"github.com/grolang/gro/macro"
)

type Call struct{}

func (m Call) Init(p interface{}) {
}

func (m Call) Main(p macro.Parser) ast.Expr {

	p.Next()
	return &ast.BasicLit{ValuePos: token.NoPos, Kind: token.STRING, Value: "\"\""}

}

