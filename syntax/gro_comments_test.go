// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"testing"
)

//================================================================================
func TestComments(t *testing.T){
	groTest(t, groTestData{
//--------------------------------------------------------------------------------
//comments above "project" keyword
	{
		num: 100,
		fnm: "dud.gro",
		src:`/* (c)2017 Grolang.*/
project myproj
package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"abc/abc.go":`/* (c)2017 Grolang.*/

package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`}},

//--------------------------------------------------------------------------------
//comments above "project" keyword with 2 explicit packages, and comments above "package" kw
	{
		num: 110,
		fnm: "dud.gro",
		src:`/* (c)2017 Grolang.*/
project mypack
package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
//another comment
//end of mypack

/* Package defg is amazing! */
package defg
import "fmt"

//don't want this here to be in func run's doc-comment

//here am I!
func run() {
	fmt.Println("Hello, world!")
}

/* Not part of package hij's doc-comments */

package hij
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"abc/abc.go":`/* (c)2017 Grolang.*/

package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,

"defg/defg.go":`/* (c)2017 Grolang.*/

/* Package defg is amazing! */
package defg

import "fmt"

//here am I!
func run() {
	fmt.Println("Hello, world!")
}`,

"hij/hij.go":`/* (c)2017 Grolang.*/

package hij

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`}},

//--------------------------------------------------------------------------------
//comments above "project" keyword with 1 implicit and 1 "internal" packages, and comments above "internal" kw
	{
		num: 120,
		fnm: "dud.gro",
		src:`/* (c)2017 Grolang.*/
project mypack
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
/* Package defg is amazing! */
internal defg
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"mypack.go":`/* (c)2017 Grolang.*/

package mypack

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,

"internal/defg/defg.go":`/* (c)2017 Grolang.*/

/* Package defg is amazing! */
package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`}},

//--------------------------------------------------------------------------------
//single "package" comment attached to only 1 of 2 "section"s
	{
		num: 130,
		fnm: "dud.gro",
		src:`/* Package abc section "this" is amazing! */
package abc
section "this"
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
section "that"
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"this.go": `/* Package abc section "this" is amazing! */
package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
"that.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`}},

//--------------------------------------------------------------------------------
//comments above sections, both form-style and curlied
	{
		num: 200,
		fnm: "dud.gro",
		src:`import "fmt"
func run() { fmt.Println("Hello, world!") }

// This file is called afile.go
// and ends with .go.
section "afile"
import "fmt"
//don't want this here to be in const pi's doc-comment

// a comment for pi
const pi = 3.14
func run() {
	fmt.Println("Hello, world!")
}
package "somebase" defg {
	section "bfile" {
		import "fmt"
		//don't want this here to be in var v's doc-comment

		// last 6 letters backwds
		var v = "zyxwvu"
		func run() { fmt.Println("Hello, world!") }
	}
	//don't want this here to be in cfile's doc-comment

	// This file is called cfile.go
	// and ends with .go.
	section "cfile" {
		import "fmt"
		//don't want this here to be in type T's doc-comment

		// nice type
		type T int32
		func run() { fmt.Println("Hello, world!") }
	}
}
// Package hij runs sth.
package "somebase/defg" hij {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
	section "dfile" {
		import "fmt"
		func run() {
			fmt.Println("Hello, world!")
		}
	}
}`,

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go":`package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
"afile.go":`package dud

// This file is called afile.go
// and ends with .go.

import "fmt"

// a comment for pi
const pi = 3.14

func run() {
	fmt.Println("Hello, world!")
}`,
"somebase/defg/bfile.go":`package defg

import "fmt"

// last 6 letters backwds
var v = "zyxwvu"

func run() {
	fmt.Println("Hello, world!")
}`,
"somebase/defg/cfile.go":`package defg

// This file is called cfile.go
// and ends with .go.

import "fmt"

// nice type
type T int32

func run() {
	fmt.Println("Hello, world!")
}`,
"somebase/defg/hij/hij.go":`// Package hij runs sth.
package hij

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
"somebase/defg/hij/dfile.go":`package hij

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`}},

//--------------------------------------------------------------------------------
})}

//================================================================================
func TestUseDecls(t *testing.T){
	groTest(t, groTestData{
//--------------------------------------------------------------------------------
	{
		num: 100,
		fnm: "dud.gro",
		src:`project myproj
use a b c "myuse"
package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"abc/abc.go":`package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`}},

//--------------------------------------------------------------------------------
	{
		num: 110,
		fnm: "dud.gro",
		src:`import "fmt"
prepare "afile.gro"
func main() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go":`// +build ignore

package main

import (
	sys "github.com/grolang/gro/sys"
)

import "fmt"

func init() {
	sys.Prepare("afile.gro")
}

func main() {
	fmt.Println("Hello, world!")
}`}},

//--------------------------------------------------------------------------------
})}

//================================================================================
