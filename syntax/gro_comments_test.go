// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"testing"
)

//================================================================================
func TestComments(t *testing.T) {
	groTest(t, groTestData{
		//--------------------------------------------------------------------------------
		//comments above "project" keyword
		{
			num: 100,
			fnm: "dud.gro",
			src: `/* (c)2017 Grolang.*/
project myproj
package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"abc/abc.go": `/* (c)2017 Grolang.*/

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
			src: `/* (c)2017 Grolang.*/
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
			prt: map[string]string{
				"abc/abc.go": `/* (c)2017 Grolang.*/

package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,

				// - - - - - - - - - - - - - - - - - - - -
				"defg/defg.go": `/* (c)2017 Grolang.*/

/* Package defg is amazing! */
package defg

import "fmt"

//here am I!
func run() {
	fmt.Println("Hello, world!")
}`,

				// - - - - - - - - - - - - - - - - - - - -
				"hij/hij.go": `/* (c)2017 Grolang.*/

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
			src: `/* (c)2017 Grolang.*/
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
			prt: map[string]string{
				"mypack.go": `/* (c)2017 Grolang.*/

package mypack

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,

				// - - - - - - - - - - - - - - - - - - - -
				"internal/defg/defg.go": `/* (c)2017 Grolang.*/

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
			src: `/* Package abc section "this" is amazing! */
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
			prt: map[string]string{
				"this.go": `/* Package abc section "this" is amazing! */
package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
				// - - - - - - - - - - - - - - - - - - - -
				"that.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
		//single comment attached to "do"
		{
			num: 140,
			fnm: "dud.gro",
			src: `package abc
import "fmt"

//my comment attached to 'do'
do {
	fmt.Println("Hello, world!")
}`,

			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import "fmt"

//my comment attached to 'do'
func init() {
	{
		fmt.Println("Hello, world!")
	}
}`}},

		//--------------------------------------------------------------------------------
		//comments above sections, both form-style and curlied
		{
			num: 200,
			fnm: "dud.gro",
			src: `import "fmt"
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
			prt: map[string]string{
				"dud.go": `package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
				// - - - - - - - - - - - - - - - - - - - -
				"afile.go": `package dud

// This file is called afile.go
// and ends with .go.

import "fmt"

// a comment for pi
const pi = 3.14

func run() {
	fmt.Println("Hello, world!")
}`,
				// - - - - - - - - - - - - - - - - - - - -
				"somebase/defg/bfile.go": `package defg

import "fmt"

// last 6 letters backwds
var v = "zyxwvu"

func run() {
	fmt.Println("Hello, world!")
}`,
				// - - - - - - - - - - - - - - - - - - - -
				"somebase/defg/cfile.go": `package defg

// This file is called cfile.go
// and ends with .go.

import "fmt"

// nice type
type T int32

func run() {
	fmt.Println("Hello, world!")
}`,
				// - - - - - - - - - - - - - - - - - - - -
				"somebase/defg/hij/hij.go": `// Package hij runs sth.
package hij

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
				// - - - - - - - - - - - - - - - - - - - -
				"somebase/defg/hij/dfile.go": `package hij

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
	})
}

//================================================================================
func TestLineTags(t *testing.T) {
	groTest(t, groTestData{
		//--------------------------------------------------------------------------------
		//assorted line directives -- including absent ones
		{
			num: 100,
			fnm: "dud.gro",
			src: `use("linedirectives")
package abc
import "fmt"

func run() {
	fmt.Println("Hello, world!")
}

var a = 123
const b = "abc"

type C = struct{d, e rune}
`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `//line dud.gro:2
package abc

//line dud.gro:3
import "fmt"

func run() {
	fmt.Println("Hello, world!")
}

var a = 123

//line dud.gro:10
const b = "abc"

type C = struct {
	d, e rune
}`}},

		//--------------------------------------------------------------------------------
		//all linedirs present
		{
			num: 110,
			fnm: "dud.gro",
			src: `use("linedirectives")
package abc
import "fmt"
var u, v int32
const pi= 3.1416
type T struct{a,b int}
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `//line dud.gro:2
package abc

//line dud.gro:3
import "fmt"

//line dud.gro:4
var u, v int32

//line dud.gro:5
const pi = 3.1416

//line dud.gro:6
type T struct {
	a, b int
}

//line dud.gro:7
func run() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
		//mix line directives and comments
		{
			num: 120,
			fnm: "dud.gro",
			src: `use "linedirectives" //#### 1
import "fmt" //#### 2
func run() { fmt.Println("Hello, world!") } //#### 3

// This file is called afile.go
// and ends with .go.
section "afile" //#### 7
import "fmt" //#### 8
//don't want this here to be in const pi's doc-comment

// a comment for pi
const pi = 3.14 //#### 12
func run() {
	fmt.Println("Hello, world!") //#### 14
}
package "somebase" defg {
	section "bfile" { //#### 17
		import "fmt" //#### 18
		//don't want this here to be in var v's doc-comment

		// last 6 letters backwds
		var v = "zyxwvu" //#### 22
		func run() { fmt.Println("Hello, world!") } //#### 23
	}
	//don't want this here to be in cfile's doc-comment

	// This file is called cfile.go
	// and ends with .go.
	section "cfile" { //#### 29
		import "fmt" //#### 30
		//don't want this here to be in type T's doc-comment

		// nice type
		type T int32 //#### 34
		func run() { fmt.Println("Hello, world!") } //#### 35
	}
}
// Package hij runs sth.
package "somebase/defg" hij { //#### 39
	import "fmt" //#### 40
	func run() {
		fmt.Println("Hello, world!") //#### 42
	}
	section "dfile" { //#### 44
		import "fmt" //#### 45
		func run() {
			fmt.Println("Hello, world!") //#### 47
		}
	}
} //#### 50`,

			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `//line dud.gro:2
package dud

//line dud.gro:2
import "fmt"

//line dud.gro:3
func run() {
	fmt.Println("Hello, world!")
}`,

				// - - - - - - - - - - - - - - - - - - - -
				"afile.go": `//line dud.gro:7
package dud

// This file is called afile.go
// and ends with .go.

//line dud.gro:8
import "fmt"

//line dud.gro:11
// a comment for pi
const pi = 3.14

//line dud.gro:13
func run() {
	fmt.Println("Hello, world!")
}`,

				// - - - - - - - - - - - - - - - - - - - -
				"somebase/defg/bfile.go": `//line dud.gro:16
package defg

import "fmt"

//line dud.gro:21
// last 6 letters backwds
var v = "zyxwvu"

//line dud.gro:23
func run() {
	fmt.Println("Hello, world!")
}`,

				// - - - - - - - - - - - - - - - - - - - -
				"somebase/defg/cfile.go": `//line dud.gro:29
package defg

// This file is called cfile.go
// and ends with .go.

//line dud.gro:30
import "fmt"

//line dud.gro:33
// nice type
type T int32

//line dud.gro:35
func run() {
	fmt.Println("Hello, world!")
}`,

				// - - - - - - - - - - - - - - - - - - - -
				"somebase/defg/hij/hij.go": `//line dud.gro:38
// Package hij runs sth.
package hij

//line dud.gro:40
import "fmt"

//line dud.gro:41
func run() {
	fmt.Println("Hello, world!")
}`,

				// - - - - - - - - - - - - - - - - - - - -
				"somebase/defg/hij/dfile.go": `//line dud.gro:44
package hij

//line dud.gro:45
import "fmt"

//line dud.gro:46
func run() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
	})
}

//================================================================================
func TestElseClause(t *testing.T) {
	groTest(t, groTestData{
		//--------------------------------------------------------------------------------
		//dangling "switch" from "else"
		{
			num: 410,
			fnm: "dud.gro",
			src: `package abc
import "fmt"
do if true {
	fmt.Println("true")
} else switch {
default:
	fmt.Println("false")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import "fmt"

func init() {
	if true {
		fmt.Println("true")
	} else {
		switch {
		default:
			fmt.Println("false")
		}
	}
}`}},

		//--------------------------------------------------------------------------------
		//dangling "for" from "else"
		{
			num: 411,
			fnm: "dud.gro",
			src: `package abc
import "fmt"
do if true {
	fmt.Println("true")
} else for {
	fmt.Println("false")
	break
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import "fmt"

func init() {
	if true {
		fmt.Println("true")
	} else {
		for {
			fmt.Println("false")
			break
		}
	}
}`}},

		//--------------------------------------------------------------------------------
	})
}

//================================================================================
