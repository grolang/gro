// Copyright 2016 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package scanner

import(
	"github.com/grolang/gro/macro"
	"github.com/grolang/gro/macro/expr"
	"github.com/grolang/gro/macro/fresh"
	"github.com/grolang/gro/macro/command"
)

var StmtMacros = map[rune]macro.StmtMacro {
	'鲜': fresh.Fresh{},
	'准': command.Prepare{},
	'执': command.Execute{},
	'跑': command.Run{},
}

var ExprMacros = map[rune]macro.ExprMacro {
	'叫': expr.Call{},
}

