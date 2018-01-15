// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"testing"
)

//================================================================================
func TestBlacklist(t *testing.T) {
	groTest(t, groTestData{
		//--------------------------------------------------------------------------------
		//use macro "blacklist" (do-for)
		{
			num: 100,
			fnm: "dud.gro",
			src: `project eggs
use "blacklist"("ifKw", "fallthrough")
package abc
import "fmt"
do for a:= range as {
	fmt.Printf("Hello, %s of Mars!", a)
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"abc/abc.go": `package abc

import "fmt"

func init() {
	for a := range as {
		fmt.Printf("Hello, %s of Mars!", a)
	}
}
`}},

		//--------------------------------------------------------------------------------
		//use macro "blacklist" (project)
		{
			num: 110,
			fnm: "dud.gro",
			src: `project eggs
use "blacklist"("project") //should have no effect because "project" kw used before blacklisting
package abc
import "fmt"
do for a:= range as {
	fmt.Printf("Hello, %s of Mars!", a)
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"abc/abc.go": `package abc

import "fmt"

func init() {
	for a := range as {
		fmt.Printf("Hello, %s of Mars!", a)
	}
}
`}},

		//--------------------------------------------------------------------------------
		//use macro "blacklist" (use)
		{
			num: 120,
			fnm: "dud.gro",
			src: `project eggs
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
			prt: map[string]string{
				"abc/abc.go": `package abc

import "fmt"

func init() {
	for a := range as {
		fmt.Printf("Hello, %s of Mars!", a)
	}
}
`}},

		//--------------------------------------------------------------------------------
		//use macro "blacklist" (inferMain)
		{
			num: 130,
			fnm: "dud.gro",
			src: `use "blacklist"("inferMain") //test from gro_test.go
import "fmt"
func main() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package dud

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//generate errors by using blacklisted keywords
		{
			num: 210,
			fnm: "dud.gro",
			src: `project eggs
use "blacklist"("ifKw", "gotoKw")
package abc
import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
			err: "dud.gro:5:1: syntax error: if-statement has been disabled but is present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 220,
			fnm: "dud.gro",
			src: `project eggs
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
			err: "dud.gro:8:1: syntax error: multi-packages disabled but more than one package present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 230,
			fnm: "dud.gro",
			src: `project eggs
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
			err: "dud.gro:3:1: syntax error: infer-packages disabled but no explicit \"package\" keyword present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 240,
			fnm: "dud.gro",
			src: `project eggs
use (
	"blacklist"("useKw")
	"open"
)
use "close"

import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
			err: "dud.gro:6:1: syntax error: \"use\" keywords are disabled but keyword is present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 250,
			fnm: "dud.gro",
			src: `project eggs
use (
	"blacklist"("packageKw")
)
package abc
import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
			err: "dud.gro:5:1: syntax error: \"package\" (and similar) keywords are disabled but keyword is present",
		},
		//--------------------------------------------------------------------------------
		{
			num: 260,
			fnm: "dud.gro",
			src: `project eggs
use (
	"blacklist"("internalKw")
)
internal abc
import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
			err: "dud.gro:5:1: syntax error: \"internal\" keywords disabled but keyword is present",
		},
		//--------------------------------------------------------------------------------
		{
			num: 270,
			fnm: "dud.gro",
			src: `project eggs
use (
	"blacklist"("sectionKw")
)
section "sth"
import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
			err: "dud.gro:5:1: syntax error: \"section\" keywords are disabled but keyword is present",
		},
		//--------------------------------------------------------------------------------
		{
			num: 271,
			fnm: "dud.gro",
			src: `project eggs
use (
	"blacklist"("mainKw")
)
main "swhere"
import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
			err: "dud.gro:5:1: syntax error: \"main\" keywords are disabled but keyword is present",
		},
		//--------------------------------------------------------------------------------
		{
			num: 272,
			fnm: "dud.gro",
			src: `project eggs
use (
	"blacklist"("testcodeKw")
)
testcode "sth"
import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
			err: "dud.gro:5:1: syntax error: \"testcode\" keywords are disabled but keyword is present",
		},
		//--------------------------------------------------------------------------------
		{
			num: 280,
			fnm: "dud.gro",
			src: `project eggs
use "blacklist" ("importKw")
import "fmt"
if true {
	"fmt".Println("Hi!")
}`,
			err: "dud.gro:3:8: syntax error: \"import\" keywords are disabled but keyword is present",
			//TODO: correct pos-info to keyword at :3:1, not import-string
		},

		//--------------------------------------------------------------------------------
		{
			num: 290,
			fnm: "dud.gro",
			src: `project eggs
use "blacklist" ("varKw")
import "fmt"
var a = "world"
"fmt".Printf("Hi, %s!", a)`,
			err: "dud.gro:4:1: syntax error: \"var\" keywords are disabled but keyword is present",
		},
		//--------------------------------------------------------------------------------
		{
			num: 291,
			fnm: "dud.gro",
			src: `project eggs
use "blacklist" ("constKw")
import "fmt"
const a = "world"
"fmt".Printf("Hi, %s!", a)`,
			err: "dud.gro:4:1: syntax error: \"const\" keywords are disabled but keyword is present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 292,
			fnm: "dud.gro",
			src: `project eggs
use "blacklist" ("typeKw")
import "fmt"
type a int32
"fmt".Printf("Hi, %s!", a(123))`,
			err: "dud.gro:4:1: syntax error: \"type\" keywords are disabled but keyword is present",
		},
		//--------------------------------------------------------------------------------
		{
			num: 293,
			fnm: "dud.gro",
			src: `project eggs
use "blacklist" ("procKw")
import "fmt"
proc hi() {
	"fmt".Printf("Hi, %s!", a)
}`,
			err: "dud.gro:4:1: syntax error: \"proc\" keywords are disabled but keyword is present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 300,
			fnm: "dud.gro",
			src: `project eggs
use "blacklist" ("genericCall")
import sp "somePkg" (int)
func hi() {
	sp.Printf("Hi")
}`,
			err: "dud.gro:6:2: syntax error: calling generic-type packages disabled but import arguments present",
			//TODO: correct pos-info to generic-call, i.e. :3:1 or :3:21
		},

		//--------------------------------------------------------------------------------
		{
			num: 310,
			fnm: "dud.gro",
			src: `project eggs
use "blacklist" ("genericDef")
package abc (U)
func hi() {
	sp.Printf("Hi %T", U)
}`,
			err: "dud.gro:3:13: syntax error: defining generic packages is disabled but one is present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 320,
			fnm: "dud.gro",
			src: `project eggs
use "blacklist" ("inplaceImps")
package abc
func hi() {
	"fmt".Printf("Hi there\n")
}`,
			err: "dud.gro:5:7: syntax error: inplace-imports disabled but are present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 330,
			fnm: "dud.gro",
			src: `use "blacklist" ("pkgSectBlocks") //test from gro_test.go
package abc {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
}`,
			err: "dud.gro:3:2: syntax error: using block-style notation for packages and sections is disabled but it is being used",
			//TODO: correct pos-info to :2:13
		},

		//--------------------------------------------------------------------------------
		{
			num: 340,
			fnm: "dud.gro",
			src: `use "blacklist" ("pkgSectBlocks") //test from gro_test.go
package abc
section "this" {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
}`,
			err: "dud.gro:4:2: syntax error: using block-style notation for packages and sections is disabled but it is being used",
			//TODO: correct pos-info to :3:16
		},
		//--------------------------------------------------------------------------------
		{
			num: 400,
			fnm: "dud.gro",
			src: `use "blacklist" ("interfaceKw")
type myIface interface {
	myfunc(int) bool
}`,
			err: "dud.gro:2:14: syntax error: \"interface\" keywords are disabled but keyword is present",
		},
		//--------------------------------------------------------------------------------
		{
			num: 410,
			fnm: "dud.gro",
			src: `use "blacklist" ("chanKw")
type myChan chan *int
`,
			err: "dud.gro:2:13: syntax error: \"chan\" keywords are disabled but keyword is present",
		},
		//--------------------------------------------------------------------------------
		{
			num: 420,
			fnm: "dud.gro",
			src: `use "blacklist" ("mapKw")
type myMap map[int]bool
`,
			err: "dud.gro:2:12: syntax error: \"map\" keywords are disabled but keyword is present",
		},
		//--------------------------------------------------------------------------------
		{
			num: 430,
			fnm: "dud.gro",
			src: `use "blacklist" ("deferKw")
func main(){
	defer func(){return}()
}
`,
			err: "dud.gro:3:2: syntax error: defer-statement has been disabled but is present",
		},
		//--------------------------------------------------------------------------------
		{
			num: 440,
			fnm: "dud.gro",
			src: `use "blacklist" ("goKw")
func pain(){
	go func(){return}()
}
`,
			err: "dud.gro:3:2: syntax error: go-statement has been disabled but is present",
		},
		//--------------------------------------------------------------------------------
		{
			num: 450,
			fnm: "dud.gro",
			src: `use "blacklist" ("rangeKw")
for n:= range ns {
	break
}
`,
			err: "dud.gro:2:18: syntax error: \"range\" keywords are disabled but keyword is present",
			//TODO: correct pos-info to :2:9
		},
		//--------------------------------------------------------------------------------
		{
			num: 460,
			fnm: "dud.gro",
			src: `use "blacklist" ("caseKw")
switch {
	case a: break
	case b: fallthrough
	default: println("abc")
}
`,
			err: "dud.gro:3:2: syntax error: \"case\" keywords are disabled but keyword is present",
		},
		//--------------------------------------------------------------------------------
		{
			num: 470,
			fnm: "dud.gro",
			src: `use "blacklist" ("defaultKw")
switch {
	case a: break
	case b: fallthrough
	default: println("abc")
}
`,
			err: "dud.gro:5:2: syntax error: \"default\" keywords are disabled but keyword is present",
		},
		//--------------------------------------------------------------------------------
		{
			num: 480,
			fnm: "dud.gro",
			src: `use "blacklist" ("caseKw")
select {
	case <-a: break
	case <-b: break
	default: println("abc")
}
`,
			err: "dud.gro:3:2: syntax error: \"case\" keywords are disabled but keyword is present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 490,
			fnm: "dud.gro",
			src: `use "blacklist" ("defaultKw")
select {
	case <-a: break
	case <-b: break
	default: println("abc")
}
`,
			err: "dud.gro:5:2: syntax error: \"default\" keywords are disabled but keyword is present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 500,
			fnm: "dud.gro",
			src: `use "blacklist" ("fallthroughKw")
switch {
	case a: break
	case b: fallthrough
	default: println("abc")
}
`,
			err: "dud.gro:4:10: syntax error: fallthrough-statement has been disabled but is present",
		},
		//--------------------------------------------------------------------------------
		{
			num: 510,
			fnm: "dud.gro",
			src: `use "blacklist" ("funcKw")
func myfunc() {
	return
}
`,
			err: "dud.gro:2:1: syntax error: \"func\" keywords are disabled but keyword is present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 520,
			fnm: "dud.gro",
			src: `use "blacklist" ("funcKw")
var myfunc = func() {
	return
}
`,
			err: "dud.gro:2:14: syntax error: \"func\" keywords are disabled but keyword is present",
		},
		//--------------------------------------------------------------------------------
		{
			num: 530,
			fnm: "dud.gro",
			src: `use "blacklist" ("structKw")
type myStruct struct {
	a, b *int
}
`,
			err: "dud.gro:2:15: syntax error: \"struct\" keywords are disabled but keyword is present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 540,
			fnm: "dud.gro",
			src: `use "blacklist" ("elseKw")
func main (){
	if a<10 {
		return
	} else {
		break
	}
}
`,
			err: "dud.gro:5:4: syntax error: \"else\" keywords are disabled but keyword is present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 550,
			fnm: "dud.gro",
			src: `use "blacklist" ("breakKw")
for i:= 0; i<10; i++ {
	break
}
`,
			err: "dud.gro:3:2: syntax error: break-statement has been disabled but is present",
		},
		//--------------------------------------------------------------------------------
		{
			num: 560,
			fnm: "dud.gro",
			src: `use "blacklist" ("continueKw")
for i:= 0; i<10; i++ {
	continue
}
`,
			err: "dud.gro:3:2: syntax error: continue-statement has been disabled but is present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 570,
			fnm: "dud.gro",
			src: `use "blacklist" ("gotoKw")
func main(){
a:
	b:= 7
	goto a
}
`,
			err: "dud.gro:5:2: syntax error: goto-statement has been disabled but is present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 580,
			fnm: "dud.gro",
			src: `use "blacklist" ("selectKw")
select {
	case <-a: break
	case <-b: break
	default: println("abc")
}
`,
			err: "dud.gro:2:1: syntax error: select-statement has been disabled but is present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 590,
			fnm: "dud.gro",
			src: `use "blacklist" ("switchKw")
switch {
	case a: break
	case b: fallthrough
	default: println("abc")
}
`,
			err: "dud.gro:2:1: syntax error: switch-statement has been disabled but is present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 600,
			fnm: "dud.gro",
			src: `use "blacklist" ("returnKw")
func main (){
	if a<10 {
		return
	} else {
		break
	}
}
`,
			err: "dud.gro:4:3: syntax error: return-statement has been disabled but is present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 610,
			fnm: "dud.gro",
			src: `use "blacklist" ("forKw")
for i:= 0; i<10; i++ {
	break
}
`,
			err: "dud.gro:2:1: syntax error: for-statement has been disabled but is present",
		},

		//--------------------------------------------------------------------------------
		{
			num: 700,
			fnm: "dud.g", // <--- NOTE: g-file should allow func, for, break, then halt on "range"
			src: `package main
func main() {
	for i:= 0; i<10; i++ {
		break
	}
	for n:= range ns {
		continue
	}
}
`,
			err: "dud.g:6:19: syntax error: \"range\" keywords are disabled but keyword is present",
			//TODO: correct pos-info to :6:10
		},

		//--------------------------------------------------------------------------------

	})
}
