// Copyright 2016 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

//Package for processing blacklists.
package whitelist

import (
	"go/ast"
	"github.com/grolang/gro/macro"
)

type Struct struct{}

func (m Struct) Init(p macro.Parser) {
	p.ProcBlacklist("如")
}

func (m Struct) Main(p macro.Parser) ast.Stmt {
	return &ast.EmptyStmt{Semicolon: p.Pos(), Implicit: true}
}

