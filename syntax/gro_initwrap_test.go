// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"testing"
)

//================================================================================
func TestInitwrap(t *testing.T){
	groTest(t, groTestData{
//--------------------------------------------------------------------------------
//single standalone stmt
	{
		num: 10,
		fnm: "dud.gro",
		src:`"fmt".Println("Hi, world!")`,

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `// +build ignore

package main

import (
	fmt "fmt"
)

func init() {
	fmt.Println("Hi, world!")
}

func main() {}`}},

//--------------------------------------------------------------------------------
//single "package" directive, with single standalone stmt (if)
	{
		num: 100,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
if true {
	do a:= "world"
	do fmt.Printf("Goodbye, %s.", a)
}
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package abc

import "fmt"

func init() {
	if true {
		a := "world"
		fmt.Printf("Goodbye, %s.", a)
	}
}

func run() {
	fmt.Println("Hello, world!")
}`}},

//--------------------------------------------------------------------------------
//single "package" directive, with two standalone stmts together (if, for)
	{
		num: 110,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
if true {
	do fmt.Println("Goodbye, world.")
}
for i:= 0; i < 10; i++ {
	do fmt.Println("Welcome back, world.")
}
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package abc

import "fmt"

func init() {
	if true {
		fmt.Println("Goodbye, world.")
	}
	for i := 0; i < 10; i++ {
		fmt.Println("Welcome back, world.")
	}
}

func run() {
	fmt.Println("Hello, world!")
}`}},

//--------------------------------------------------------------------------------
//single "package" directive, with two separated standalone stmts (if, for)
	{
		num: 120,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
if true {
	do fmt.Println("Goodbye, world.")
}
func run() {
	fmt.Println("Hello, world!")
}
for i:= 0; i < 10; i++ {
	do fmt.Println("Welcome back, world.")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package abc

import "fmt"

func init() {
	if true {
		fmt.Println("Goodbye, world.")
	}
}

func run() {
	fmt.Println("Hello, world!")
}

func init() {
	for i := 0; i < 10; i++ {
		fmt.Println("Welcome back, world.")
	}
}`}},

//--------------------------------------------------------------------------------
//single "package" directive and 2 "section" directives with standalone stmts (if, do-block, do-standalone, for)
	{
		num: 130,
		fnm: "dud.gro",
		src:`package abc
section "this"
import "fmt"
var n int //should be interpreted as top-level decl
if true { do n++ }
do fmt.Println("n is:", n)
func run() {
	fmt.Println("Hello, world!")
}
section "that"
import "fmt"
var n int
do {
	n++
	fmt.Println("n is:", n)
}
for i:= 1; i<=10; i++ {
	do fmt.Println("i is:", i)
}
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"this.go": `package abc

import "fmt"

var n int

func init() {
	if true {
		n++
	}
	fmt.Println("n is:", n)
}

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"that.go": `package abc

import "fmt"

var n int

func init() {
	{
		n++
		fmt.Println("n is:", n)
	}
	for i := 1; i <= 10; i++ {
		fmt.Println("i is:", i)
	}
}

func run() {
	fmt.Println("Hello, world!")
}`}},

//--------------------------------------------------------------------------------
//single "project" directive with standalone stmts (do-block, bare-block)
	{
		num: 140,
		fnm: "dud.gro",
		src:`project abc
do {
	a:= 7
	b:= 8
	fmt.Println(a+b)
	{
		c:= 9
	}
}
{
	do a:= 17
	do b:= 18
	do fmt.Println(a*b)
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"abc.go": `// +build ignore

package main

func init() {
	{
		a := 7
		b := 8
		fmt.Println(a + b)
		{
			c := 9
		}
	}
	{
		a := 17
		b := 18
		fmt.Println(a * b)
	}
}

func main() {}`}},

//--------------------------------------------------------------------------------
//single "section" directive with standalone stmts (do-block, do-standalone)
	{
		num: 150,
		fnm: "dud.gro",
		src:`section "abcde"
do {
	a:= 7
	b:= 8
	fmt.Println(a+b)
	{
		c:= 9
	}
}
do a:= 17
do fmt.Println(a*b)`,

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"abcde.go": `// +build ignore

package main

func init() {
	{
		a := 7
		b := 8
		fmt.Println(a + b)
		{
			c := 9
		}
	}
	a := 17
	fmt.Println(a * b)
}

func main() {}`}},

//--------------------------------------------------------------------------------
//standalone stmts only (do-standalone)
	{
		num: 160,
		fnm: "dud.gro",
		src:`do a:= 7
do b:= 8
do fmt.Println(a+b)`,

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `// +build ignore

package main

func init() {
	a := 7
	b := 8
	fmt.Println(a + b)
}

func main() {}`}},

//--------------------------------------------------------------------------------
//single standalone stmt (for with embedded do-standalone)
	{
		num: 170,
		fnm: "dud.gro",
		src:`for { do a:= 7 }`,

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `// +build ignore

package main

func init() {
	for {
		a := 7
	}
}

func main() {}`}},

//--------------------------------------------------------------------------------
//single standalone block (bare-block)
	{
		num: 180,
		fnm: "dud.gro",
		src:`{ do a:= 7 }`,

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `// +build ignore

package main

func init() {
	{
		a := 7
	}
}

func main() {}`}},

//--------------------------------------------------------------------------------
//do-for, for-standalone
	{
		num: 200,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
do for a:= range as {
	fmt.Printf("Hello, %s of Mars!", a)
}
for a:= range as {
	do b:= a
	do fmt.Printf("Hello, %s of Mars!", a)
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package abc

import "fmt"

func init() {
	for a := range as {
		fmt.Printf("Hello, %s of Mars!", a)
	}
	for a := range as {
		b := a
		fmt.Printf("Hello, %s of Mars!", a)
	}
}`}},

//--------------------------------------------------------------------------------
//do-if, if-standalone
	{
		num: 210,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
do if a < 10 {
	fmt.Printf("Hello, %s of Mars!", a)
}
if a < 10 {
	do b:= a
	do fmt.Printf("Hello, %s of Mars!", a)
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package abc

import "fmt"

func init() {
	if a < 10 {
		fmt.Printf("Hello, %s of Mars!", a)
	}
	if a < 10 {
		b := a
		fmt.Printf("Hello, %s of Mars!", a)
	}
}`}},

//--------------------------------------------------------------------------------
//do-switch
	{
		num: 220,
		fnm: "dud.gro",
		src:`do switch {
case 1:
	"fmt".Println("abc")
case 2:
	a:= 7
default:
	"fmt".Println("defg")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go":`// +build ignore

package main

import (
	fmt "fmt"
)

func init() {
	switch {
	case 1:
		fmt.Println("abc")
	case 2:
		a := 7
	default:
		fmt.Println("defg")
	}
}

func main() {}`}},

//--------------------------------------------------------------------------------
//do-go, go-standalone with syntax shortcut
	{
		num: 230,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
do go func() {
	fmt.Printf("Hello, %s of Mars!", a)
}()
go func() { //also parses OK
	fmt.Printf("Hello, %s of Mars!", a)
}()
do go {
	fmt.Printf("Hello, %s of Jupiter!", a)
}
go {
	do fmt.Printf("Hello, %s of Saturn!", a)
}
`,

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package abc

import "fmt"

func init() {
	go func() {
		fmt.Printf("Hello, %s of Mars!", a)
	}()
	go func() {
		fmt.Printf("Hello, %s of Mars!", a)
	}()
	go func() {
		fmt.Printf("Hello, %s of Jupiter!", a)
	}()
	go func() {
		fmt.Printf("Hello, %s of Saturn!", a)
	}()
}`}},

//--------------------------------------------------------------------------------
//do-defer, defer-standalone with syntax shortcut
	{
		num: 231,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
do defer func() {
	fmt.Printf("Hello, %s of Mars!", a)
}()
do defer {
	fmt.Printf("Hello, %s of Jupiter!", a)
}
defer {
	do fmt.Printf("Hello, %s of Saturn!", a)
}
`,

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package abc

import "fmt"

func init() {
	defer func() {
		fmt.Printf("Hello, %s of Mars!", a)
	}()
	defer func() {
		fmt.Printf("Hello, %s of Jupiter!", a)
	}()
	defer func() {
		fmt.Printf("Hello, %s of Saturn!", a)
	}()
}`}},

//--------------------------------------------------------------------------------
//compare "func" and "proc" keywords
	{
		num: 300,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
proc runtwo() {
	do fmt.Println("Hello, Mars!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}

func runtwo() {
	fmt.Println("Hello, Mars!")
}`}},

//--------------------------------------------------------------------------------
//const both at top-level and within proc
	{
		num: 310,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
const a = 123
do fmt.Println("a is:", a)
proc runtwo() {
	const b = 789
	do fmt.Println("b is:", b)
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package abc

import "fmt"

const a = 123

func init() {
	fmt.Println("a is:", a)
}

func runtwo() {
	const b = 789
	fmt.Println("b is:", b)
}`}},

//--------------------------------------------------------------------------------
//var and type, each both at top-level and within proc
	{
		num: 311,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
type ii int32
var a = 123
do fmt.Println("a is:", a)
proc runtwo() {
	type ij int32
	var b = 789
	do fmt.Println("b is:", b)
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package abc

import "fmt"

type ii int32

var a = 123

func init() {
	fmt.Println("a is:", a)
}

func runtwo() {
	type ij int32
	var b = 789
	fmt.Println("b is:", b)
}`}},

//--------------------------------------------------------------------------------
//check Gro-sans-Go keywords can still be used as names
	{
		num: 400,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
var section = 123
const internal = "abc"
do fmt.Println("sec is:", section, "; intern is:", internal)
section "defg"
import "fmt"
var do = 123
const main = "zyx"
do fmt.Println("do is:", do, "; main is:", main)
do do:= 7
`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package abc

import "fmt"

var section = 123

const internal = "abc"

func init() {
	fmt.Println("sec is:", section, "; intern is:", internal)
}`,
"defg.go": `package abc

import "fmt"

var do = 123

const main = "zyx"

func init() {
	fmt.Println("do is:", do, "; main is:", main)
	do := 7
}`}},

//--------------------------------------------------------------------------------
//dangling "switch" from "else"
	{
		num: 410,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
do if true {
	fmt.Println("true")
} else switch {
default:
	fmt.Println("false")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
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
		src:`package abc
import "fmt"
do if true {
	fmt.Println("true")
} else for {
	fmt.Println("false")
	break
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
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
//
	{
		num: 510,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
for a:= range as {
	fmt.Printf("Hello, %s of Mars!", a) //need "do" to introduce Go-style statement
}
`,
		err: ":4:2: syntax error: unexpected fmt, expecting }",
},

//--------------------------------------------------------------------------------
//can't have "proc" as expression within "go"
	{
		num: 521,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
go proc() {
	do fmt.Printf("Hello, %s of Mars!", a)
}()
`,
		err: ":3:11: syntax error: unexpected { at end of statement",
},

//--------------------------------------------------------------------------------
//need either "do" or "func" containing Go statements
	{
		num: 522,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
go {
	fmt.Printf("Hello, %s of Neptune!", a)
}
`,
		err: ":4:2: syntax error: unexpected fmt, expecting }",
},

//--------------------------------------------------------------------------------
//can't have "do" inside "func"
	{
		num: 523,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
go func() {
	do fmt.Printf("Hello, %s of Uranus!", a)
}()
`,
		err: ":4:5: syntax error: unexpected fmt at end of statement",
},

//--------------------------------------------------------------------------------
})}

