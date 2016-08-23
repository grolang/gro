// Copyright 2016 The Gro Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package basic

import (
	"github.com/grolang/gro/tutor"
)

//================================================================================

var TutStmts = &tutor.Tutorial{
	Name:      "stmts",
	Short:     "explains the statements of Gro",
	Long:      `
Stmts is a tutorial on the statements of Gro.

	`,
	Pages:     []*tutor.Page{
		stmtsPkg,
		stmtsSource,
		stmtsImport,
	},
}

//================================================================================
var stmtsPkg = &tutor.Page{
	English: []string{`
包 is an alias for the package header, and has many forms:

    包mypkg            //defines package with specified name just like in Go
    包正               //正 is an alias for "main"
    包                 //package name "main" is used as default
    包"oneLib/hi"      //defines package hi, and uses oneLib/hi as base location for source files
    包"someDir"thinking //defines package thinking, and uses someDir as base location for source files

A single .gro file can contain more than one 包(package) header. When it's missing at the top, package "main" is used.

`},
}

//================================================================================
var stmtsSource = &tutor.Page{
	English: []string{`
Each package statement may have one of more source statements, with the package name followed by .go used if it's missing. It has form:

    源"xyz.go"         //defines source file "xyz.go" within package
    源                 //defines source with same name as package name, but with ".go" appended

Each package in a .gro file can contain more than one 源(source) header.

`},
}

//================================================================================
var stmtsImport = &tutor.Page{
	English: []string{`
The 入(import) statement has many forms, and cause the public members of the imported package to be available in the importing one.
Some forms are the same as in Go:

    入"abc/defg"        //makes members available if qualified with last segment and dot, e.g. defg.Member
    入hij"fmt"          //accepts identifier as alias, just like in Go
    入."xyz"            //accepts dot form, just like in Go
    入卟"unicode/utf8"  //accepts 口-radical Unihan as alias
    入㕤hij"fmt"        //accepts identifier and Unihan together
    入叨叩kl"unicode/utf8" //accepts more than one Unihan alias
    入("math/rand";"sync/atomic") //accepts parenthesized forms

`},
}

//================================================================================

