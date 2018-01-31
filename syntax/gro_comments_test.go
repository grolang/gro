// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"testing"
)

//================================================================================
func TestHashCmds(t *testing.T) {
	groTest(t, groTestData{
		//--------------------------------------------------------------------------------
		//hashbangs
		{
			num: 100,
			fnm: "dud", // <--- no suffix is normal use case
			src: `#!/usr/local/go/bin/gro
package main
func main(){
	"fmt".Printf("\n")
}
`,

			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package main

import (
	fmt "fmt"
)

func main() {
	fmt.Printf("\n")
}
`}},

		//--------------------------------------------------------------------------------
		{
			num: 110,
			fnm: "dud.gro",
			src: `#unexpected name
package main
func main(){
	"fmt".Printf("\n")
}
`,
			err: "dud.gro:1:1: syntax error: unexpected name at top-level"},

		//--------------------------------------------------------------------------------
		//hash-cmds (both go-kw and gro-kw) incl #import, #const, #var, #type
		{
			num: 120,
			fnm: "dud.grooy", // <---- ".grooy" extension for hash-cmds
			src: `
#project myproj
package abc
#import "fmt"
#const const = 7890
#const a = 789
#var b = 123
#var var = 1234
#type (t = #int; u #string; range u )
func run() {
	fmt.Println("Hello, world!")
	#const c, for = 777, 888
	#var (d = "defg"; e ="hij")
	#type v #struct{i, j, struct int}
}`,

			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"abc/abc.go": `package abc

import (
	groo "github.com/grolang/gro/ops"
	_groo "github.com/grolang/gro/ops"
)

import _fmt "fmt"

const _const = 7890
const _a = 789

var _b = 123
var _var = 1234

type (
	_t = int
	_u string
	_range _u
)

func run() {
	fmt.Println(groo.MakeText("Hello, world!"))
	const _c, _for = 777, 888
	var (
		_d = _groo.MakeText("defg")
		_e = _groo.MakeText("hij")
	)
	type _v struct {
		_i, _j, _struct _int
	}
}

type (
	any = interface{}
	void = struct{}
)

var inf = groo.Inf

func init() {
	groo.UseUtf88 = true
}
`}},

		//--------------------------------------------------------------------------------
		{
			num: 121,
			fnm: "dud.gro",
			src: `
package main
const c, for = 777, 888
func main(){
	"fmt".Printf("\n")
}
`,
			err: "dud.gro:3:10: syntax error: unexpected for, expecting name"},

		//--------------------------------------------------------------------------------
		//hash-cmds in combinations: #func, #proc, #do, #if, #else
		{
			num: 130,
			fnm: "dud.grooy",
			src: `#project myproj
package abc
#import fmt "fmt"

	if 1==0 {do fmt.Println()}
	if 1==0 {do fmt.Println()} else {do fmt.Println()}
	if 1==0 {do fmt.Println()} #else {#do fmt.Println()}
	#if 1==0 {#do fmt.Println()}
	#if 1==0 {#do fmt.Println()} #else {#do fmt.Println()}

	do if 1==0 {fmt.Println()}
	do if 1==0 {fmt.Println()} else {fmt.Println()}
	do if 1==0 {fmt.Println()} #else {fmt.Println()}
	do #if 1==0 {fmt.Println()}
	do #if 1==0 {fmt.Println()} #else {fmt.Println()}

	#do #if 1==0 {fmt.Println()}
	#do #if 1==0 {fmt.Println()} #else {fmt.Println()}
	#do #if 1==0 {fmt.Println()}
	#do #if 1==0 {fmt.Println()} #else {fmt.Println()}
{
	do if 1==0 {fmt.Println()}
	do if 1==0 {fmt.Println()} else {fmt.Println()}
	do if 1==0 {fmt.Println()} #else {fmt.Println()}
	do #if 1==0 {fmt.Println()}
	do #if 1==0 {fmt.Println()} #else {fmt.Println()}
}
{
	#do #if 1==0 {fmt.Println()}
	#do #if 1==0 {fmt.Println()} #else {fmt.Println()}
}

func run() {
	fmt.Println("Hello, world!")
	if 1==0 {fmt.Println()}
	if 1==0 {fmt.Println()} else {fmt.Println()}
	if 1==0 {fmt.Println()} #else {fmt.Println()}
	#if 1==0 {fmt.Println()}
	#if 1==0 {fmt.Println()} #else {fmt.Println()}
}
proc sim() {
	do fmt.Println("Horray, world!")
	do if 1==0 {fmt.Println()}
	do if 1==0 {fmt.Println()} else {fmt.Println()}
	do if 1==0 {fmt.Println()} #else {fmt.Println()}
	do #if 1==0 {fmt.Println()}
	do #if 1==0 {fmt.Println()} #else {fmt.Println()}
}
proc ted() {
	#do fmt.Println("Horray, world!")
	#do #if 1==0 {fmt.Println()}
	#do #if 1==0 {fmt.Println()} #else {fmt.Println()}
	#do #if 1==0 {fmt.Println()}
	#do #if 1==0 {fmt.Println()} #else {fmt.Println()}
}
#func runner() {
	fmt.Println("Hello, world!")
	#if 1==0 {fmt.Println()}
	#if 1==0 {fmt.Println()} #else {fmt.Println()}
}
#proc tedder() {
	#do fmt.Println("Horray, world!")
	#do #if 1==0 {fmt.Println()}
	#do #if 1==0 {fmt.Println()} #else {fmt.Println()}
}
`,

			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"abc/abc.go": `package abc

import (
	groo "github.com/grolang/gro/ops"
	_groo "github.com/grolang/gro/ops"
)

import _fmt "fmt"

func init() {
	if groo.IsEqual(1, 0) {
		fmt.Println()
	}
	if groo.IsEqual(1, 0) {
		fmt.Println()
	} else {
		fmt.Println()
	}
	if groo.IsEqual(1, 0) {
		fmt.Println()
	} else {
		_fmt.Println()
	}
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	}
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	} else {
		_fmt.Println()
	}
	if groo.IsEqual(1, 0) {
		fmt.Println()
	}
	if groo.IsEqual(1, 0) {
		fmt.Println()
	} else {
		fmt.Println()
	}
	if groo.IsEqual(1, 0) {
		fmt.Println()
	} else {
		_fmt.Println()
	}
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	}
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	} else {
		_fmt.Println()
	}
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	}
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	} else {
		_fmt.Println()
	}
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	}
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	} else {
		_fmt.Println()
	}
	{
		if groo.IsEqual(1, 0) {
			fmt.Println()
		}
		if groo.IsEqual(1, 0) {
			fmt.Println()
		} else {
			fmt.Println()
		}
		if groo.IsEqual(1, 0) {
			fmt.Println()
		} else {
			_fmt.Println()
		}
		if _groo.IsEqual(1, 0) {
			_fmt.Println()
		}
		if _groo.IsEqual(1, 0) {
			_fmt.Println()
		} else {
			_fmt.Println()
		}
	}
	{
		if _groo.IsEqual(1, 0) {
			_fmt.Println()
		}
		if _groo.IsEqual(1, 0) {
			_fmt.Println()
		} else {
			_fmt.Println()
		}
	}
}

func run() {
	fmt.Println(groo.MakeText("Hello, world!"))
	if groo.IsEqual(1, 0) {
		fmt.Println()
	}
	if groo.IsEqual(1, 0) {
		fmt.Println()
	} else {
		fmt.Println()
	}
	if groo.IsEqual(1, 0) {
		fmt.Println()
	} else {
		_fmt.Println()
	}
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	}
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	} else {
		_fmt.Println()
	}
}

func sim() {
	fmt.Println(groo.MakeText("Horray, world!"))
	if groo.IsEqual(1, 0) {
		fmt.Println()
	}
	if groo.IsEqual(1, 0) {
		fmt.Println()
	} else {
		fmt.Println()
	}
	if groo.IsEqual(1, 0) {
		fmt.Println()
	} else {
		_fmt.Println()
	}
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	}
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	} else {
		_fmt.Println()
	}
}

func ted() {
	_fmt.Println(_groo.MakeText("Horray, world!"))
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	}
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	} else {
		_fmt.Println()
	}
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	}
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	} else {
		_fmt.Println()
	}
}

func _runner() {
	_fmt.Println(_groo.MakeText("Hello, world!"))
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	}
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	} else {
		_fmt.Println()
	}
}

func _tedder() {
	_fmt.Println(_groo.MakeText("Horray, world!"))
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	}
	if _groo.IsEqual(1, 0) {
		_fmt.Println()
	} else {
		_fmt.Println()
	}
}

type (
	any = interface{}
	void = struct{}
)

var inf = groo.Inf

func init() {
	groo.UseUtf88 = true
}
`}},

		//--------------------------------------------------------------------------------
		{
			num: 131,
			fnm: "dud.grooy",
			src: `#project myproj
package abc
#import fmt "fmt"

if 1==0 {do fmt.Println()} #else {do fmt.Println()}
`,
			err: "dud.grooy:5:35: syntax error: Keywords must be prepended with # within scope of other #-keywords"},

		//--------------------------------------------------------------------------------
		{
			num: 132,
			fnm: "dud.grooy",
			src: `#project myproj
package abc
#import fmt "fmt"

#do #if 1==0 {fmt.Println()} else {fmt.Println()}
`,
			err: "dud.grooy:5:30: syntax error: Keywords must be prepended with # within scope of other #-keywords"},

		//--------------------------------------------------------------------------------
		//hash-cmds in combinations: #for, #range, #continue, #goto
		{
			num: 140,
			fnm: "dud.grooy",
			src: `#project myproj
package abc
#import fmt "fmt"
func run() {
	fmt.Println("Hello, world!")
	#goto sala
	#for i:= 0; i<10; i++ {
		fmt.Println("Goodbye, cruel world.")
	}
	#for n:= #range ns {
		fmt.Println("Malah.")
	}
_sala:
	for n:= #range ns {
		fmt.Println("Malah.")
		#continue sala
	}
	#for n:= #range ns {
		fmt.Println("Malah.")
	}
}
#for i:= 0; i<10; i++ {
	#do fmt.Println("Goodbye, cruel world.")
}
#for n:= #range ns {
	#do fmt.Println("Malah.")
}
for n:= #range ns {
	do fmt.Println("Malah.")
}
#for n:= #range ns {
	#do fmt.Println("Malah.")
}
`,

			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"abc/abc.go": `package abc

import (
	groo "github.com/grolang/gro/ops"
	_groo "github.com/grolang/gro/ops"
)

import _fmt "fmt"

func run() {
	fmt.Println(groo.MakeText("Hello, world!"))
	goto _sala
	for _i := 0; _groo.IsLessThan(_i, 10); _i++ {
		_fmt.Println(_groo.MakeText("Goodbye, cruel world."))
	}
	for _n := range _ns {
		_fmt.Println(_groo.MakeText("Malah."))
	}
_sala:
	for n := range _ns {
		fmt.Println(groo.MakeText("Malah."))
		continue _sala
	}
	for _n := range _ns {
		_fmt.Println(_groo.MakeText("Malah."))
	}
}

func init() {
	for _i := 0; _groo.IsLessThan(_i, 10); _i++ {
		_fmt.Println(_groo.MakeText("Goodbye, cruel world."))
	}
	for _n := range _ns {
		_fmt.Println(_groo.MakeText("Malah."))
	}
	for n := range _ns {
		fmt.Println(groo.MakeText("Malah."))
	}
	for _n := range _ns {
		_fmt.Println(_groo.MakeText("Malah."))
	}
}

type (
	any = interface{}
	void = struct{}
)

var inf = groo.Inf

func init() {
	groo.UseUtf88 = true
}
`}},

		//--------------------------------------------------------------------------------
		//hash-cmds: #defer, #go
		{
			num: 150,
			fnm: "dud.grooy",
			src: `#project myproj
package abc
#import fmt "fmt"
func run() {
	fmt.Println("Hello, world!")
	#defer #func(){
		fmt.Println("Hello, world!")
	}()
	#defer{
		fmt.Println("Hello, world!")
	}
	#go #func(){
		fmt.Println("Hello, world!")
	}()
	#go{
		fmt.Println("Hello, world!")
	}
}

	#go #func(){
		fmt.Println("Hello, world!")
	}()
	#go{
		#do fmt.Println("Hello, world!")
	}
`,

			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"abc/abc.go": `package abc

import (
	groo "github.com/grolang/gro/ops"
	_groo "github.com/grolang/gro/ops"
)

import _fmt "fmt"

func run() {
	fmt.Println(groo.MakeText("Hello, world!"))
	defer func() {
		_fmt.Println(_groo.MakeText("Hello, world!"))
	}()
	defer func() {
		_fmt.Println(_groo.MakeText("Hello, world!"))
	}()
	go func() {
		_fmt.Println(_groo.MakeText("Hello, world!"))
	}()
	go func() {
		_fmt.Println(_groo.MakeText("Hello, world!"))
	}()
}

func init() {
	go func() {
		_fmt.Println(_groo.MakeText("Hello, world!"))
	}()
	go func() {
		_fmt.Println(_groo.MakeText("Hello, world!"))
	}()
}

type (
	any = interface{}
	void = struct{}
)

var inf = groo.Inf

func init() {
	groo.UseUtf88 = true
}
`}},

		//--------------------------------------------------------------------------------
		{
			num: 151,
			fnm: "dud.grooy",
			src: `#project myproj
package abc
#import fmt "fmt"
#go{
	do fmt.Println("Hello, world!")
}
`,
			err: "dud.grooy:5:2: syntax error: Keywords must be prepended with # within scope of other #-keywords"},

		//--------------------------------------------------------------------------------
		//hash-cmds: #switch, #select, #case, #default, #break, #fallthrough
		{
			num: 160,
			fnm: "dud.grooy",
			src: `#project myproj
package abc
#import fmt "fmt"
func run() {
	fmt.Println("Hello, world!")
	#switch i {
	#case 789:
		fmt.Println("Hello, world!")
	#default:
		fmt.Println("Hello, world!")
	}
	#select {
		#case <-a: #break
		#case <-b: #break
		#default: fmt.Println("abc")
	}
_mylab:
	switch i {
	case 123:
		fmt.Println("Hello, world!")
		#break mylab
	#case 789:
		fmt.Println("Hello, world!")
	default:
		fmt.Println("Hello, world!")
	}
}
#switch i {
#case 789:
	#do fmt.Println("Hello, world!")
#default:
	#do fmt.Println("Hello, world!")
}
select {
	case <-a: break
	#case <-b: #break
	default: do fmt.Println("abc")
}
switch i {
#case 123:
	#do fmt.Println("Hello, world!")
case 789:
	do fmt.Println("Hello, world!")
	#fallthrough
#default:
	#do fmt.Println("Hello, world!")
}
`,

			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"abc/abc.go": `package abc

import (
	groo "github.com/grolang/gro/ops"
	_groo "github.com/grolang/gro/ops"
)

import _fmt "fmt"

func run() {
	fmt.Println(groo.MakeText("Hello, world!"))
	switch _i {
	case 789:
		_fmt.Println(_groo.MakeText("Hello, world!"))
	default:
		_fmt.Println(_groo.MakeText("Hello, world!"))
	}
	select {
	case <-_a:
		break
	case <-_b:
		break
	default:
		_fmt.Println(_groo.MakeText("abc"))
	}
_mylab:
	switch i {
	case 123:
		fmt.Println(groo.MakeText("Hello, world!"))
		break _mylab
	case 789:
		_fmt.Println(_groo.MakeText("Hello, world!"))
	default:
		fmt.Println(groo.MakeText("Hello, world!"))
	}
}

func init() {
	switch _i {
	case 789:
		_fmt.Println(_groo.MakeText("Hello, world!"))
	default:
		_fmt.Println(_groo.MakeText("Hello, world!"))
	}
	select {
	case <-a:
		break
	case <-_b:
		break
	default:
		fmt.Println(groo.MakeText("abc"))
	}
	switch i {
	case 123:
		_fmt.Println(_groo.MakeText("Hello, world!"))
	case 789:
		fmt.Println(groo.MakeText("Hello, world!"))
		fallthrough
	default:
		_fmt.Println(_groo.MakeText("Hello, world!"))
	}
}

type (
	any = interface{}
	void = struct{}
)

var inf = groo.Inf

func init() {
	groo.UseUtf88 = true
}
`}},

		//--------------------------------------------------------------------------------
		{
			num: 161,
			fnm: "dud.grooy",
			src: `#project myproj
switch i {
#case 123:
	do fmt.Println("Hello, world!")
}
`,
			err: "dud.grooy:4:2: syntax error: Keywords must be prepended with # within scope of other #-keywords"},

		//--------------------------------------------------------------------------------
		//hash-cmds: #struct, #map, #chan, #interface; also spec id's: #int, #string
		{
			num: 170,
			fnm: "dud.grooy",
			src: `#project myproj
package abc
#import fmt "fmt"
func run() {
	type St #struct{a, b #int}
	var m #map[#string]#int = map[string]int{"abc":123}
	type ch #chan #int
	type ci <-#chan #int
	type I #interface{f();g()}
	fmt.Println("Hello, world!")
}
`,

			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"abc/abc.go": `package abc

import (
	groo "github.com/grolang/gro/ops"
)

import _fmt "fmt"

func run() {
	type St struct {
		_a, _b int
	}
	var m map[string]int = map[string]int{
		groo.MakeText("abc"): 123,
	}
	type ch chan int
	type ci <-chan int
	type I interface {
		_f()
		_g()
	}
	fmt.Println(groo.MakeText("Hello, world!"))
}

type (
	any = interface{}
	void = struct{}
)

var inf = groo.Inf

func init() {
	groo.UseUtf88 = true
}
`}},

		//--------------------------------------------------------------------------------
		{
			num: 180,
			fnm: "dud.grooy",
			src: `#project myproj
package abc
#import fmt "fmt"
func run() {
	fmt.Println("Hello, world!")
}
`,

			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"abc/abc.go": `package abc

import (
	groo "github.com/grolang/gro/ops"
)

import _fmt "fmt"

func run() {
	fmt.Println(groo.MakeText("Hello, world!"))
}

type (
	any = interface{}
	void = struct{}
)

var inf = groo.Inf

func init() {
	groo.UseUtf88 = true
}
`}},

		//--------------------------------------------------------------------------------
	})
}

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
}
`}},

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
}
`,

				// - - - - - - - - - - - - - - - - - - - -
				"defg/defg.go": `/* (c)2017 Grolang.*/

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
`,

				// - - - - - - - - - - - - - - - - - - - -
				"hij/hij.go": `/* (c)2017 Grolang.*/

/* Not part of package hij's doc-comments */

package hij

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//comments above "project" keyword with 1 implicit and 1 "internal" packages, and comments above "internal" kw
		{
			num: 120,
			fnm: "dud.gro",
			src: `/* (c)2017 Grolang.*/
//Another project comment
//and another
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
//Another project comment
//and another

package mypack

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,

				// - - - - - - - - - - - - - - - - - - - -
				"internal/defg/defg.go": `/* (c)2017 Grolang.*/
//Another project comment
//and another

/* Package defg is amazing! */
package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

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
}
`,
				// - - - - - - - - - - - - - - - - - - - -
				"that.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

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
}
`}},

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
}
`,
				// - - - - - - - - - - - - - - - - - - - -
				"afile.go": `package dud

// This file is called afile.go
// and ends with .go.

import "fmt"

//don't want this here to be in const pi's doc-comment

// a comment for pi
const pi = 3.14

func run() {
	fmt.Println("Hello, world!")
}
`,
				// - - - - - - - - - - - - - - - - - - - -
				"somebase/defg/bfile.go": `package defg

import "fmt"

//don't want this here to be in var v's doc-comment

// last 6 letters backwds
var v = "zyxwvu"

func run() {
	fmt.Println("Hello, world!")
}
`,
				// - - - - - - - - - - - - - - - - - - - -
				"somebase/defg/cfile.go": `package defg

//don't want this here to be in cfile's doc-comment

// This file is called cfile.go
// and ends with .go.

import "fmt"

//don't want this here to be in type T's doc-comment

// nice type
type T int32

func run() {
	fmt.Println("Hello, world!")
}
`,
				// - - - - - - - - - - - - - - - - - - - -
				"somebase/defg/hij/hij.go": `// Package hij runs sth.
package hij

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				// - - - - - - - - - - - - - - - - - - - -
				"somebase/defg/hij/dfile.go": `package hij

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//comments above and after const/type/var, both inside spec and standalone decl
		{
			num: 210,
			fnm: "dud.gro",
			src: `//import spec
import (
	"path/filepath" //import comment //TODO: make this comment pass thru
	"strings"
)
//import decl
import "fmt"

//const spec
const(
	a=7 //const comment
	b=8
)
//const decl
const c=9 //const comment

//type spec
type(
	T bool
	U= string //type comment
)
//type decl
type V int
type W string //type comment

//var spec
var(
	x int = 1 //var comment
	y uint =2
)
//var decl
var z string = "3" //var comment
`,

			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package dud

import (
	"path/filepath"
	"strings"
)

import "fmt"

//const spec
const (
	a = 7 //const comment
	b = 8
)

//const decl
const c = 9 //const comment

//type spec
type (
	T bool
	U = string //type comment
)

//type decl
type V int
type W string //type comment

//var spec
var (
	x int = 1 //var comment
	y uint = 2
)

//var decl
var z string = "3" //var comment
`}},

		//--------------------------------------------------------------------------------
		//"go:" and "line" comments
		{
			num: 220,
			fnm: "dud.gro",
			src: `package abc
import "fmt"

//go:linkname stringsIndexByte strings.IndexByte
func stringsIndexByte(s string, c byte) int

//line :21
var v int

//live comment but not //line
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import "fmt"

//go:linkname stringsIndexByte strings.IndexByte
func stringsIndexByte(s string, c byte) int

//line :21
var v int

//live comment but not //line
func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//typical doc.go file
		{
			num: 230,
			fnm: "doc.gro",
			src: `// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package race implements data race detection logic.
// No public interface is provided.
// For details about the race detector see
// https://golang.org/doc/articles/race_detector.html
package race
`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"doc.go": `// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package race implements data race detection logic.
// No public interface is provided.
// For details about the race detector see
// https://golang.org/doc/articles/race_detector.html
package race
`}},

		//--------------------------------------------------------------------------------
		//comments above various stmts:
		//if, for, switch, select, defer, go, const, const(), return, goto, break, continue
		{
			num: 240,
			fnm: "dud.gro",
			src: `package main
func main(){
	//comment here!
	if true {
		fmt.Println("abc")
	}
	//another comment!
	for r:= range rr {
		fmt.Println("defg")
	}

	//yet another comment!
	switch n {
	case 7:
		fmt.Println("defg")
	default:
		fmt.Println("hij")
	}
	//comment four
	select {
	}
	//comment five
	defer func() {
		return
	}()
	//comment six
	go func() {
		return
	}()
	//comment seven
	const a=7
sevenofnine:
	//comment eight
	const (
		b=8
		c=9
	)
	//comment nine
	goto sevenofnine

	for true {
		//cmt-10
		break
	}
	for true {
		//cmt-11
		continue
	}

	//cmt-12
	return true
}
`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package main

func main() {
	//comment here!
	if true {
		fmt.Println("abc")
	}
	//another comment!
	for r := range rr {
		fmt.Println("defg")
	}
	//yet another comment!
	switch n {
	case 7:
		fmt.Println("defg")
	default:
		fmt.Println("hij")
	}
	//comment four
	select {}
	//comment five
	defer func() {
		return
	}()
	//comment six
	go func() {
		return
	}()
	//comment seven
	const a = 7
sevenofnine:
	//comment eight
	const (
		b = 8
		c = 9
	)
	//comment nine
	goto sevenofnine
	for true {
		//cmt-10
		break
	}
	for true {
		//cmt-11
		continue
	}
	//cmt-12
	return true
}
`}},

		//--------------------------------------------------------------------------------
		//comments to right of various stmts:
		//return, goto, break, continue
		{
			num: 250,
			fnm: "dud.gro",
			src: `package main
func main(){
sevenofnine:
	//comment nine
	goto sevenofnine //comment 9-a

	for true {
		//cmt-10
		break //comment 10-a
	}
	for true {
		//cmt-11
		continue //comment 11-a
	}

	//cmt-12
	return true //comment 12-a
}
`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package main

func main() {
sevenofnine:
	//comment nine
	goto sevenofnine //comment 9-a
	for true {
		//cmt-10
		break //comment 10-a
	}
	for true {
		//cmt-11
		continue //comment 11-a
	}
	//cmt-12
	return true //comment 12-a
}
`}},

		//--------------------------------------------------------------------------------
		//multiple comments comprising doc-comment
		{
			num: 260,
			fnm: "dud.gro",
			src: `package abc
import "fmt"

//comment

/*hey
here*/
//comment
/*hey
there*/
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import "fmt"

//comment

/*hey
here*/
//comment
/*hey
there*/
func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//comments to right of various stmts
		{
			num: 270,
			fnm: "dud.gro",
			src: `package main
func main(){
	a:= 7 //here
	a++ //and here
	a = 18 //and here
	a++

	//and here
	b = 24 //and here
}
`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package main

func main() {
	a := 7 //here
	a++ //and here
	a = 18 //and here
	a++
	b = 24 //and here
}
`}},

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
}
`}},

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
}
`}},

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
}
`,

				// - - - - - - - - - - - - - - - - - - - -
				"afile.go": `//line dud.gro:7
package dud

// This file is called afile.go
// and ends with .go.

//line dud.gro:8
import "fmt"

//don't want this here to be in const pi's doc-comment

//line dud.gro:11
// a comment for pi
const pi = 3.14 //#### 12

//line dud.gro:13
func run() {
	fmt.Println("Hello, world!") //#### 14
}
`,

				// - - - - - - - - - - - - - - - - - - - -
				"somebase/defg/bfile.go": `//line dud.gro:16
package defg

import "fmt"

//don't want this here to be in var v's doc-comment

//line dud.gro:21
// last 6 letters backwds
var v = "zyxwvu" //#### 22

//line dud.gro:23
func run() {
	fmt.Println("Hello, world!")
}
`,

				// - - - - - - - - - - - - - - - - - - - -
				"somebase/defg/cfile.go": `//line dud.gro:29
package defg

//don't want this here to be in cfile's doc-comment

// This file is called cfile.go
// and ends with .go.

//line dud.gro:30
import "fmt"

//don't want this here to be in type T's doc-comment

//line dud.gro:33
// nice type
type T int32 //#### 34

//line dud.gro:35
func run() {
	fmt.Println("Hello, world!")
}
`,

				// - - - - - - - - - - - - - - - - - - - -
				"somebase/defg/hij/hij.go": `//line dud.gro:38
// Package hij runs sth.
package hij

//line dud.gro:40
import "fmt"

//line dud.gro:41
func run() {
	fmt.Println("Hello, world!") //#### 42
}
`,

				// - - - - - - - - - - - - - - - - - - - -
				"somebase/defg/hij/dfile.go": `//line dud.gro:44
package hij

//line dud.gro:45
import "fmt"

//line dud.gro:46
func run() {
	fmt.Println("Hello, world!") //#### 47
}
`}},

		//--------------------------------------------------------------------------------
	})
}

//================================================================================
func TestNewSyntax(t *testing.T) {
	groTest(t, groTestData{
		//--------------------------------------------------------------------------------
		//dangling "switch" from "else"
		{
			num: 100,
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
}
`}},

		//--------------------------------------------------------------------------------
		//dangling "for" from "else"
		{
			num: 110,
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
}
`}},

		//--------------------------------------------------------------------------------
		//underscores in numbers - dynamic mode, i.e. groo-file
		{
			num: 120,
			fnm: "dud.groo",
			src: `package abc
func main() {
	a:= 123_456_789
	a2:= 1_234
	b:= 012_37
	b2:= 0_1237
	c:= 0xab_cd_ef
	c2:= 0x_a16f

	_0123:= "Hey"
	d:= _0123 //identifier, not number
}`,

			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import (
	groo "github.com/grolang/gro/ops"
)

func main() {
	a := 123456789
	a2 := 1234
	b := 01237
	b2 := 01237
	c := 0xabcdef
	c2 := 0xa16f
	_0123 := groo.MakeText("Hey")
	d := _0123 //identifier, not number
}

type (
	any = interface{}
	void = struct{}
)

var inf = groo.Inf

func init() {
	groo.UseUtf88 = true
}
`}},

		//--------------------------------------------------------------------------------
		//underscores in numbers
		//non-dynamic mode will pass values unchanged for Go to reject
		//TODO: Gro should reject them here
		{
			num: 121,
			fnm: "dud.gro",
			src: `package abc
func main() {
	a:= 123_456_789
	a2:= 1_234
	b:= 012_37
	b2:= 0_1237
	c:= 0xab_cd_ef
	c2:= 0x_a16f
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

func main() {
	a := 123_456_789
	a2 := 1_234
	b := 012_37
	b2 := 0_1237
	c := 0xab_cd_ef
	c2 := 0x_a16f
}
`}},

		//--------------------------------------------------------------------------------
		//dates in dynamic mode
		{
			num: 130,
			fnm: "dud.groo",
			src: `package abc
func main() {
	a:= 2003.8.29
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import (
	groo "github.com/grolang/gro/ops"
	time "time"
)

func main() {
	a := time.Date(2003, 8, 29, 0, 0, 0, 0, time.UTC)
}

type (
	any = interface{}
	void = struct{}
)

var inf = groo.Inf

func init() {
	groo.UseUtf88 = true
}
`}},

		//--------------------------------------------------------------------------------
		//underscores on RHS of short func notation in dynamic mode
		{
			num: 200,
			fnm: "dud.groo",
			src: `package abc
func main() {
	a:= func{return 7}
	b:= func{89}
	c:= func{10*_}
	d:= func{_+_*_}
	a();b();c();d()
}`,

			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import (
	groo "github.com/grolang/gro/ops"
)

func main() {
	a := func(groo_it ...interface{}) interface{} {
		return 7
	}
	b := func(groo_it ...interface{}) interface{} {
		return 89
	}
	c := func(groo_it ...interface{}) interface{} {
		return groo.Mult(10, groo_it[0])
	}
	d := func(groo_it ...interface{}) interface{} {
		return groo.Plus(groo_it[0], groo.Mult(groo_it[1], groo_it[2]))
	}
	a()
	b()
	c()
	d()
}

type (
	any = interface{}
	void = struct{}
)

var inf = groo.Inf

func init() {
	groo.UseUtf88 = true
}
`}},

		//--------------------------------------------------------------------------------
	})
}

//================================================================================
