// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"testing"
)

//================================================================================
func TestGenerics(t *testing.T){
	groTest(t, groTestData{
//--------------------------------------------------------------------------------
//one plain "package" and one parameterized package only
	{
		num: 10,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package def (T) //don't want this one to appear
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`}},

//--------------------------------------------------------------------------------
//one plain "package" with one explicit section, and one 2-param package
	{
		num: 20,
		fnm: "dud.gro",
		src:`package abc
section "myfile"
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package def (T, U) //don't want this one to appear
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"abc/myfile.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`}},

//--------------------------------------------------------------------------------
//one 2-param package
	{
		num: 30,
		fnm: "dud.gro",
		src:`package def (T, U)
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
//expect package to be dropped
}},

//--------------------------------------------------------------------------------
//one plain "package" (with explicit section), and one 2-param package (with explicit section), both in block-style
	{
		num: 40,
		fnm: "dud.gro",
		src:`package abc {
  section "myfile" {
    import "fmt"
    func run() {
      fmt.Println("Hello, world!")
    }
  }
}
package def (T, U) {
  import "fmt"
  func run() {
    fmt.Println("Hello, world!")
  }
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"abc/myfile.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`}},

//--------------------------------------------------------------------------------
//a 1-param package and a calling package
	{
		num: 100,
		fnm: "dud.gro",
		src:`package defg (T) //package with parameter...
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package hij
import dint "defg" (int) //...called with an argument
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"hij/hij.go": `package hij

import dint "generics/hij/hij/dint"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/generic_args.go": `package defg

type T = int`}},

//--------------------------------------------------------------------------------
//calling & 1-param packages, gro-file in subdir
	{
		num: 101,
		fnm: "somedir/dud.gro",
		src:`package defg (T) //package with parameter...
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package hij
import dint "defg" (int) //...called with an argument
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"hij/hij.go": `package hij

import dint "somedir/generics/hij/hij/dint"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/generic_args.go": `package defg

type T = int`}},

//--------------------------------------------------------------------------------
//one plain "package", a 1-param package, and one calling package
	{
		num: 110,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package defg (T)
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package hij
import dint "defg" (int)
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"hij/hij.go": `package hij

import dint "generics/hij/hij/dint"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/generic_args.go": `package defg

type T = int`}},

//--------------------------------------------------------------------------------
//an implicit plain package, a 1-param package, and a package calling it twice
	{
		num: 120,
		fnm: "dud.gro",
		src:`import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package defg (T)
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package hij
import "fmt"
import dint   "defg" (int)
import dfloat "defg" (float)
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"hij/hij.go": `package hij

import "fmt"
import dint "generics/hij/hij/dint"
import dfloat "generics/hij/hij/dfloat"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/generic_args.go": `package defg

type T = int`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dfloat/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dfloat/generic_args.go": `package defg

type T = float`}},

//--------------------------------------------------------------------------------
//a 2-param package, and an implicit package calling it
	{
		num: 130,
		fnm: "dud.gro",
		src:`import "fmt"
import dintfl "defg" (int, float)
func run() {
	fmt.Println("Hello, world!")
}
package defg (T, U)
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package dud

import "fmt"
import dintfl "generics/dud/dintfl"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/dud/dintfl/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/dud/dintfl/generic_args.go": `package defg

type T = int
type U = float`}},

//--------------------------------------------------------------------------------
//a 2-param package, and an implicit package calling it twice
	{
		num: 140,
		fnm: "dud.gro",
		src:`import "fmt"
import (
	dint   "defg" (int, float)
	dfloat "defg" (float, int)
)
func run() {
	fmt.Println("Hello, world!")
}
package defg (T, U)
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package dud

import "fmt"

import (
	dint "generics/dud/dint"
	dfloat "generics/dud/dfloat"
)

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/dud/dint/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/dud/dint/generic_args.go": `package defg

type T = int
type U = float`,
// - - - - - - - - - - - - - - - - - - - -
"generics/dud/dfloat/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/dud/dfloat/generic_args.go": `package defg

type T = float
type U = int`}},

//--------------------------------------------------------------------------------
//a 2-param package, and an implicit package calling it twice (with same first arg in each call)
	{
		num: 150,
		fnm: "dud.gro",
		src:`import "fmt"
import (
	dif   "defg" (int, float)
	dib   "defg" (int, byte)
)
func run() {
	fmt.Println("Hello, world!")
}
package defg (T, U)
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package dud

import "fmt"

import (
	dif "generics/dud/dif"
	dib "generics/dud/dib"
)

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/dud/dif/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/dud/dif/generic_args.go": `package defg

type T = int
type U = float`,
// - - - - - - - - - - - - - - - - - - - -
"generics/dud/dib/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/dud/dib/generic_args.go": `package defg

type T = int
type U = byte`}},

//--------------------------------------------------------------------------------
//one 1-param package, and two calling packages (each using same arg and alias name)
	{
		num: 200,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
import dint "defg" (int)
func run() {
	fmt.Println("Hello, world!")
}
package defg (T)
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package hij
import dint "defg" (int)
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"abc/abc.go": `package abc

import "fmt"
import dint "generics/abc/abc/dint"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"hij/hij.go": `package hij

import dint "generics/hij/hij/dint"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/abc/abc/dint/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/abc/abc/dint/generic_args.go": `package defg

type T = int`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/generic_args.go": `package defg

type T = int`}},

//--------------------------------------------------------------------------------
//one 1-param package, and two calling packages (each using same arg, but different alias name)
	{
		num: 210,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
import dint "xyz/defg" (int)
func run() {
	fmt.Println("Hello, world!")
}
package "xyz" defg (T)
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package hij
import deger "xyz/defg" (int)
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"abc/abc.go": `package abc

import "fmt"
import dint "generics/abc/abc/dint"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"hij/hij.go": `package hij

import deger "generics/hij/hij/deger"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/abc/abc/dint/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/abc/abc/dint/generic_args.go": `package defg

type T = int`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/deger/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/deger/generic_args.go": `package defg

type T = int`}},

//--------------------------------------------------------------------------------
//one 1-param package, and two calling packages (each using different arg, but same alias name)
	{
		num: 220,
		fnm: "dud.gro",
		src:`package abc
import "fmt"
import dparam "defg" (int)
func run() {
	fmt.Println("Hello, world!")
}
package hij
import dparam "defg" (float)
func run() {
	fmt.Println("Hello, world!")
}
package defg (T)
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"abc/abc.go": `package abc

import "fmt"
import dparam "generics/abc/abc/dparam"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"hij/hij.go": `package hij

import dparam "generics/hij/hij/dparam"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/abc/abc/dparam/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/abc/abc/dparam/generic_args.go": `package defg

type T = int`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dparam/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dparam/generic_args.go": `package defg

type T = float`}},

//--------------------------------------------------------------------------------
//a 2-param package, and 2 packages each calling it twice (with same first arg in each call)
	{
		num: 230,
		fnm: "dud.gro",
		src:`import "fmt"
import (
	dif   "defg" (int, float)
	dib   "defg" (int, byte)
)
func run() {
	fmt.Println("Hello, world!")
}
package abc
import "fmt"
import (
	dif   "defg" (int, int32)
	dib   "defg" (int, int64)
)
func run() {
	fmt.Println("Hello, world!")
}
package defg (T, U)
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package dud

import "fmt"

import (
	dif "generics/dud/dif"
	dib "generics/dud/dib"
)

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"abc/abc.go": `package abc

import "fmt"

import (
	dif "generics/abc/abc/dif"
	dib "generics/abc/abc/dib"
)

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/abc/abc/dif/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/abc/abc/dif/generic_args.go": `package defg

type T = int
type U = int32`,
// - - - - - - - - - - - - - - - - - - - -
"generics/abc/abc/dib/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/abc/abc/dib/generic_args.go": `package defg

type T = int
type U = int64`,
// - - - - - - - - - - - - - - - - - - - -
"generics/dud/dif/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/dud/dif/generic_args.go": `package defg

type T = int
type U = float`,
// - - - - - - - - - - - - - - - - - - - -
"generics/dud/dib/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/dud/dib/generic_args.go": `package defg

type T = int
type U = byte`}},

//--------------------------------------------------------------------------------
//a 1-param package and a calling package using in-place string package names
	{
		num: 300,
		fnm: "dud.gro",
		src:`package defg (T)
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package hij
import dint "defg" ("some/other/path".MyStruct)
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"hij/hij.go": `package hij

import dint "generics/hij/hij/dint"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/generic_args.go": `package defg

import (
	path "some/other/path"
)

type T = path.MyStruct`}},

//--------------------------------------------------------------------------------
//a 2-param package and a calling package using 2 in-place string package names
	{
		num: 310,
		fnm: "dud.gro",
		src:`package defg (T, U)
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package hij
import dint "defg" ("some/other/path".MyStruct, "yet/another".Strooct)
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"hij/hij.go": `package hij

import dint "generics/hij/hij/dint"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/generic_args.go": `package defg

import (
	path "some/other/path"
	another "yet/another"
)

type T = path.MyStruct
type U = another.Strooct`}},

//--------------------------------------------------------------------------------
//a two parameterized packages and a calling package calling both using in-place string package names
	{
		num: 320,
		fnm: "dud.gro",
		src:`package abc (S)
func pow() {
	fmt.Println("Pow wow!")
}
package defg (T, U)
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package hij
import powpal "abc" ("yet/first".Streect)
import dint "defg" ("some/other/path".MyStruct, "yet/another".Strooct)
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"hij/hij.go": `package hij

import powpal "generics/hij/hij/powpal"
import dint "generics/hij/hij/dint"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/powpal/abc.go": `package abc

func pow() {
	fmt.Println("Pow wow!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/powpal/generic_args.go": `package abc

import (
	first "yet/first"
)

type S = first.Streect`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/generic_args.go": `package defg

import (
	path "some/other/path"
	another "yet/another"
)

type T = path.MyStruct
type U = another.Strooct`}},

//--------------------------------------------------------------------------------
//a package calling a parameterized package supplied in another file
	{
		num: 330,
		fnm: "dud.gro",
		src:`include "defg.gro"
package hij
import dint "defg" (int)
func run() {
	fmt.Println("Hello, world!")
}`,

		xtr:map[string]string{
"defg.gro":`package defg (T)
`},
// - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"hij/hij.go": `package hij

import dint "generics/hij/hij/dint"

func run() {
	fmt.Println("Hello, world!")
}`,
"generics/hij/hij/dint/defg.go": `package defg`,
"generics/hij/hij/dint/generic_args.go": `package defg

type T = int`}},

//--------------------------------------------------------------------------------
// "include" cmd
	{
	num: 400,
		fnm: "dud.gro",
		src:`include "mymy"
package def
func run() {
	fmt.Println("Hello, world!")
}`,
		xtr:map[string]string{
"mymy":``},

// - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package def

func run() {
	fmt.Println("Hello, world!")
}`}},

//--------------------------------------------------------------------------------
// "include" cmd group
	{
	num: 401,
		fnm: "dud.gro",
		src:`include (
	"mymy"
	"youyou"
)
package def
func run() {
	fmt.Println("Hello, world!")
}`,
		xtr:map[string]string{
"mymy":``,
"youyou":``},

// - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package def

func run() {
	fmt.Println("Hello, world!")
}`}},

//--------------------------------------------------------------------------------
// multiple "include" cmds
	{
	num: 402,
		fnm: "dud.gro",
		src:`include (
	"mymy"
	"youyou"
)
include "itit"
package def
func run() {
	fmt.Println("Hello, world!")
}`,
		xtr:map[string]string{
"mymy":``,
"youyou":``,
"itit":``},

// - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"dud.go": `package def

func run() {
	fmt.Println("Hello, world!")
}`}},

//--------------------------------------------------------------------------------
//a 1-param "internal" package and a calling package
	{
		num: 410,
		fnm: "dud.gro",
		src:`internal defg (T) //internal package with parameter...
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package hij
import dint "internal/defg" (int) //...called with an argument
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"hij/hij.go": `package hij

import dint "generics/hij/hij/dint"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"generics/hij/hij/dint/generic_args.go": `package defg

type T = int`}},

//--------------------------------------------------------------------------------
})}

//================================================================================
func TestGenericsFails(t *testing.T){
	groFail(t, groFailData{
//--------------------------------------------------------------------------------
// param to package must be type
	{
	num: 10,
		fnm: "dud.gro",
		src:`package def ("bad param")
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
		err: ":1:14: syntax error: unexpected literal \"bad param\", expecting name",
},

//--------------------------------------------------------------------------------
// imported arg pkgs need local name
	{
	num: 20,
		fnm: "dud.gro",
		src:`package abc (T)
func run() T {
	"fmt".Println("abc")
}
package def
import "abc" (int)
func run() {
	fmt.Println("Hello, world!")
}`,
		err: ":6:15: syntax error: imported parameterized packages need a local name",
},

//--------------------------------------------------------------------------------
})}

