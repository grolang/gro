// Copyright 2016 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package fresh

import (
	"go/ast"
	"github.com/grolang/gro/macro"
)

type Fresh struct{}

func (m Fresh) Init(p interface{}) {
}

func (m Fresh) Main(p macro.Parser) ast.Stmt {
	s:= &ast.EmptyStmt{Semicolon: p.Pos(), Implicit: true}
	p.Next()
	return s
}

