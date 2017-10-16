// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"testing"
)

//================================================================================
func TestBlacklist(t *testing.T){
	groTest(t, groTestData{
//--------------------------------------------------------------------------------
//use macro "blacklist" (do-for)
	{
		num: 100,
		fnm: "dud.gro",
		src:`project eggs
use "blacklist"("if", "fallthrough")
package abc
import "fmt"
do for a:= range as {
	fmt.Printf("Hello, %s of Mars!", a)
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"abc/abc.go": `package abc

import "fmt"

func init() {
	for a := range as {
		fmt.Printf("Hello, %s of Mars!", a)
	}
}`}},

//--------------------------------------------------------------------------------
//use macro "blacklist" (project)
	{
		num: 110,
		fnm: "dud.gro",
		src:`project eggs
use "blacklist"("project") //should have no effect because "project" kw used before blacklisting
package abc
import "fmt"
do for a:= range as {
	fmt.Printf("Hello, %s of Mars!", a)
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"abc/abc.go": `package abc

import "fmt"

func init() {
	for a := range as {
		fmt.Printf("Hello, %s of Mars!", a)
	}
}`}},

//--------------------------------------------------------------------------------
//use macro "blacklist" (use)
	{
		num: 120,
		fnm: "dud.gro",
		src:`project eggs
use (
	"blacklist"("use") //should have no effect because only "use" kw used before blacklisting
	"open"
)
package abc
import "fmt"
do for a:= range as {
	fmt.Printf("Hello, %s of Mars!", a)
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"abc/abc.go": `package abc

import "fmt"

func init() {
	for a := range as {
		fmt.Printf("Hello, %s of Mars!", a)
	}
}`}},

//--------------------------------------------------------------------------------
//use macro "blacklist" (inferMain)
	{
		num: 130,
		fnm: "dud.gro",
		src:`use "blacklist"("inferMain") //test from gro_test.go
import "fmt"
func main() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package dud

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`}},

//--------------------------------------------------------------------------------
})}

//================================================================================
func TestBlacklistFails(t *testing.T){
	groFail(t, groFailData{
//--------------------------------------------------------------------------------
	{
		num: 10,
		fnm: "dud.gro",
		src:`project eggs
use "blacklist"("if", "goto")
package abc
import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
		err: ":5:1: syntax error: if-statement has been prohibited by blacklist",
},
//--------------------------------------------------------------------------------
	{
		num: 20,
		fnm: "dud.gro",
		src:`project eggs
use "blacklist"("multiPkg")
package abc
import "fmt"
if true {
	"fmt".Println("Hi!")
}
package defg
import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
		err: ":8:1: syntax error: multi-packages disabled but more than one package present",
},
//--------------------------------------------------------------------------------
	{
		num: 30,
		fnm: "dud.gro",
		src:`project eggs
use "blacklist"("inferPkg")
import "fmt"
if true {
	"fmt".Println("Hi!")
}
package defg
import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
		err: ":3:1: syntax error: unexpected import infer-packages disabled but no explicit \"package\" keyword present",
		//TODO: remove "unexpected import" from error msg
},
//--------------------------------------------------------------------------------
	{
		num: 40,
		fnm: "dud.gro",
		src:`project eggs
use (
	"blacklist"("use")
	"open"
)
use "close"

import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
		err: ":6:1: syntax error: \"use\" keywords disabled but keyword is present",
},
//--------------------------------------------------------------------------------
	{
		num: 50,
		fnm: "dud.gro",
		src:`project eggs
use (
	"blacklist"("package")
)
package abc
import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
		err: ":5:1: syntax error: \"package\" (and similar) keywords are disabled but keyword is present",
},
//--------------------------------------------------------------------------------
	{
		num: 60,
		fnm: "dud.gro",
		src:`project eggs
use (
	"blacklist"("internal")
)
internal abc
import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
		err: ":5:1: syntax error: \"internal\" keywords disabled but keyword is present",
},
//--------------------------------------------------------------------------------
	{
		num: 70,
		fnm: "dud.gro",
		src:`project eggs
use (
	"blacklist"("section")
)
section "sth"
import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
		err: ":5:1: syntax error: \"section\" keywords are disabled but keyword is present",
},
//--------------------------------------------------------------------------------
	{
		num: 71,
		fnm: "dud.gro",
		src:`project eggs
use (
	"blacklist"("main")
)
main "swhere"
import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
		err: ":5:1: syntax error: \"main\" keywords are disabled but keyword is present",
},
//--------------------------------------------------------------------------------
	{
		num: 72,
		fnm: "dud.gro",
		src:`project eggs
use (
	"blacklist"("testcode")
)
testcode "sth"
import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
		err: ":5:1: syntax error: \"testcode\" keywords are disabled but keyword is present",
},
//--------------------------------------------------------------------------------
	{
		num: 80,
		fnm: "dud.gro",
		src:`project eggs
use "blacklist" ("import")
import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
		err: ":3:8: syntax error: \"import\" keywords are disabled but keyword is present", //TODO: change pos-info to keyword, not import-string
},
//--------------------------------------------------------------------------------
	{
		num: 90,
		fnm: "dud.gro",
		src:`project eggs
use "blacklist" ("var")
import "fmt"
var a = "world"
"fmt".Printf("Hi, %s!", a)`,
		err: ":4:1: syntax error: \"var\" keywords are disabled but keyword is present",
},
//--------------------------------------------------------------------------------
	{
		num: 91,
		fnm: "dud.gro",
		src:`project eggs
use "blacklist" ("const")
import "fmt"
const a = "world"
"fmt".Printf("Hi, %s!", a)`,
		err: ":4:1: syntax error: \"const\" keywords are disabled but keyword is present",
},
//--------------------------------------------------------------------------------
	{
		num: 92,
		fnm: "dud.gro",
		src:`project eggs
use "blacklist" ("type")
import "fmt"
type a int32
"fmt".Printf("Hi, %s!", a(123))`,
		err: ":4:1: syntax error: \"type\" keywords are disabled but keyword is present",
},
//--------------------------------------------------------------------------------
	{
		num: 93,
		fnm: "dud.gro",
		src:`project eggs
use "blacklist" ("proc")
import "fmt"
proc hi() {
	"fmt".Printf("Hi, %s!", a)
}`,
		err: ":4:1: syntax error: \"proc\" keywords are disabled but keyword is present",
},
//--------------------------------------------------------------------------------
	{
		num: 100,
		fnm: "dud.gro",
		src:`project eggs
use "blacklist" ("genericCall")
import sp "somePkg" (int)
func hi() {
	sp.Printf("Hi")
}`,
		err: ":6:2: syntax error: calling generic-type packages disabled but import arguments present", //TODO: correct pos-info
},
//--------------------------------------------------------------------------------
	{
		num: 110,
		fnm: "dud.gro",
		src:`project eggs
use "blacklist" ("genericDef")
package abc (U)
func hi() {
	sp.Printf("Hi %T", U)
}`,
		err: ":3:14: syntax error: defining generic packages is disabled but one is present",
},
//--------------------------------------------------------------------------------
	{
		num: 120,
		fnm: "dud.gro",
		src:`project eggs
use "blacklist" ("inplaceImps")
package abc
func hi() {
	"fmt".Printf("Hi there\n")
}`,
		err: ":5:7: syntax error: unexpected . inplace-imports disabled but are present", //TODO: correct pos-info; AND: remove "unexpected ."
},
//--------------------------------------------------------------------------------
	{
		num: 130,
		fnm: "dud.gro",
		src:`use "blacklist" ("pkgSectBlocks") //test from gro_test.go
package abc {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
}`,
		err: ":3:2: syntax error: using block-style notation for packages and sections is disabled but it is being used", //TODO: correct pos-info
},
//--------------------------------------------------------------------------------
	{
		num: 140,
		fnm: "dud.gro",
		src:`use "blacklist" ("pkgSectBlocks") //test from gro_test.go
package abc
section "this" {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
}`,
		err: ":4:2: syntax error: using block-style notation for packages and sections is disabled but it is being used", //TODO: correct pos-info
},
//--------------------------------------------------------------------------------
})}

