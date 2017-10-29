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
import dint "github.com/grolang/gro/syntax/defg" (int) //...called with an argument
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"hij/hij.go": `package hij

import dint "github.com/grolang/gro/syntax/generics/hij/hij/dint"

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
		fnm: "adir/dud.gro",
		src:`package defg (T) //package with parameter...
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package hij
import dint "github.com/grolang/gro/syntax/adir/defg" (int) //...called with an argument
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"adir/hij/hij.go": `package hij

import dint "github.com/grolang/gro/syntax/adir/generics/hij/hij/dint"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"adir/generics/hij/hij/dint/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"adir/generics/hij/hij/dint/generic_args.go": `package defg

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
import dint "github.com/grolang/gro/syntax/defg" (int)
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

import dint "github.com/grolang/gro/syntax/generics/hij/hij/dint"

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
import dint   "github.com/grolang/gro/syntax/defg" (int)
import dfloat "github.com/grolang/gro/syntax/defg" (float)
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
import dint "github.com/grolang/gro/syntax/generics/hij/hij/dint"
import dfloat "github.com/grolang/gro/syntax/generics/hij/hij/dfloat"

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
import dintfl "github.com/grolang/gro/syntax/defg" (int, float)
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
import dintfl "github.com/grolang/gro/syntax/generics/dud/dintfl"

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
	dint   "github.com/grolang/gro/syntax/defg" (int, float)
	dfloat "github.com/grolang/gro/syntax/defg" (float, int)
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
	dint "github.com/grolang/gro/syntax/generics/dud/dint"
	dfloat "github.com/grolang/gro/syntax/generics/dud/dfloat"
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
	dif   "github.com/grolang/gro/syntax/defg" (int, float)
	dib   "github.com/grolang/gro/syntax/defg" (int, byte)
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
	dif "github.com/grolang/gro/syntax/generics/dud/dif"
	dib "github.com/grolang/gro/syntax/generics/dud/dib"
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
import dint "github.com/grolang/gro/syntax/defg" (int)
func run() {
	fmt.Println("Hello, world!")
}
package defg (T)
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package hij
import dint "github.com/grolang/gro/syntax/defg" (int)
func run() {
	fmt.Println("Hello, world!")
}`,

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"abc/abc.go": `package abc

import "fmt"
import dint "github.com/grolang/gro/syntax/generics/abc/abc/dint"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"hij/hij.go": `package hij

import dint "github.com/grolang/gro/syntax/generics/hij/hij/dint"

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
import dint "github.com/grolang/gro/syntax/xyz/defg" (int)
func run() {
	fmt.Println("Hello, world!")
}
package "xyz" defg (T)
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package hij
import deger "github.com/grolang/gro/syntax/xyz/defg" (int)
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"abc/abc.go": `package abc

import "fmt"
import dint "github.com/grolang/gro/syntax/generics/abc/abc/dint"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"hij/hij.go": `package hij

import deger "github.com/grolang/gro/syntax/generics/hij/hij/deger"

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
import dparam "github.com/grolang/gro/syntax/defg" (int)
func run() {
	fmt.Println("Hello, world!")
}
package hij
import dparam "github.com/grolang/gro/syntax/defg" (float)
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
import dparam "github.com/grolang/gro/syntax/generics/abc/abc/dparam"

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"hij/hij.go": `package hij

import dparam "github.com/grolang/gro/syntax/generics/hij/hij/dparam"

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
	dif   "github.com/grolang/gro/syntax/defg" (int, float)
	dib   "github.com/grolang/gro/syntax/defg" (int, byte)
)
func run() {
	fmt.Println("Hello, world!")
}
package abc
import "fmt"
import (
	dif   "github.com/grolang/gro/syntax/defg" (int, int32)
	dib   "github.com/grolang/gro/syntax/defg" (int, int64)
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
	dif "github.com/grolang/gro/syntax/generics/dud/dif"
	dib "github.com/grolang/gro/syntax/generics/dud/dib"
)

func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - -
"abc/abc.go": `package abc

import "fmt"

import (
	dif "github.com/grolang/gro/syntax/generics/abc/abc/dif"
	dib "github.com/grolang/gro/syntax/generics/abc/abc/dib"
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
import dint "github.com/grolang/gro/syntax/defg" ("some/other/path".MyStruct)
func run() {
	fmt.Println("Hello, world!")
}`,
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"hij/hij.go": `package hij

import dint "github.com/grolang/gro/syntax/generics/hij/hij/dint"

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
import dint "github.com/grolang/gro/syntax/defg" (
	"github.com/grolang/gro/syntax/some/other/path".MyStruct,
	"github.com/grolang/gro/syntax/yet/another".Strooct)
func run() {
	fmt.Println("Hello, world!")
}`,

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"hij/hij.go": `package hij

import dint "github.com/grolang/gro/syntax/generics/hij/hij/dint"

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
	path "github.com/grolang/gro/syntax/some/other/path"
	another "github.com/grolang/gro/syntax/yet/another"
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
import powpal "github.com/grolang/gro/syntax/abc" ("yet/first".Streect)
import dint "github.com/grolang/gro/syntax/defg" (
	"some/other/path".MyStruct,
	"github.com/grolang/gro/syntax/yet/another".Strooct,
)
func run() {
	fmt.Println("Hello, world!")
}`,

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"hij/hij.go": `package hij

import powpal "github.com/grolang/gro/syntax/generics/hij/hij/powpal"
import dint "github.com/grolang/gro/syntax/generics/hij/hij/dint"

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
	another "github.com/grolang/gro/syntax/yet/another"
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
import dint "github.com/grolang/gro/syntax/defg" (int)
func run() {
	fmt.Println("Hello, world!")
}`,

		xtr:map[string]string{
"defg.gro":`package defg (T)
`},
// - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"hij/hij.go": `package hij

import dint "github.com/grolang/gro/syntax/generics/hij/hij/dint"

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
// multiple "include" cmds using subdirs
	{
	num: 403,
		fnm: "adir/dud.gro",
		src:`include (
	"mymy"
	"anotherdir/youyou"
)
include "itit"
package def
func run() {
	fmt.Println("Hello, world!")
}`,
		xtr:map[string]string{
"adir/mymy":``,
"adir/anotherdir/youyou":``,
"adir/itit":``},

// - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"adir/dud.go": `package def

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
import dint "github.com/grolang/gro/syntax/internal/defg" (int) //...called with an argument
func run() {
	fmt.Println("Hello, world!")
}`,

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"hij/hij.go": `package hij

import dint "github.com/grolang/gro/syntax/generics/hij/hij/dint"

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
// param to package must be type
	{
	num: 510,
		fnm: "dud.gro",
		src:`package def ("bad param")
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
		err: "dud.gro:1:14: syntax error: unexpected literal \"bad param\", expecting name",
},

//--------------------------------------------------------------------------------
// imported arg pkgs need local name
	{
	num: 520,
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
		err: "dud.gro:6:15: syntax error: imported parameterized packages need a local name",
},

//--------------------------------------------------------------------------------
//nested generics
	{
		num: 530,
		fnm: "grotest/dud.gro",
		src:`import yourstruct "github.com/grolang/gro/syntax/grotest/yourthree" (float64, struct{a,b int})
"fmt".Println("'Hi' from src/grotest/sixthseashell.gro")
do yourstruct.RunIt()

package yourthree (S, T)
import mycomplex128 "github.com/grolang/gro/syntax/grotest/somedir/myfour" (complex128)
func RunIt(){
	var s S
	var t T
	"fmt".Printf("'Hi' from src/grotest/sixthseashell.gro:yourthree(%T, %T).RunIt\n", s, t)
	mycomplex128.DoIt()
}

package "somedir" myfour (T)
func DoIt(){
	var t T
	"fmt".Printf("'Hi' from src/grotest/sixthseashell.gro:somedir/myfour(%T).DoIt\n", t)
}`,

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"grotest/dud.go": `// +build ignore

package main

import (
	fmt "fmt"
)

import yourstruct "github.com/grolang/gro/syntax/grotest/generics/dud/yourstruct"

func init() {
	fmt.Println("'Hi' from src/grotest/sixthseashell.gro")
	yourstruct.RunIt()
}

func main() {}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/yourthree.go": `package yourthree

import (
	fmt "fmt"
)

import mycomplex128 "github.com/grolang/gro/syntax/grotest/generics/yourthree/yourthree/mycomplex128"

func RunIt() {
	var s S
	var t T
	fmt.Printf("'Hi' from src/grotest/sixthseashell.gro:yourthree(%T, %T).RunIt\n", s, t)
	mycomplex128.DoIt()
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/generic_args.go": `package yourthree

type S = float64
type T = struct {
	a, b int
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/yourthree/yourthree/mycomplex128/myfour.go": `package myfour

import (
	fmt "fmt"
)

func DoIt() {
	var t T
	fmt.Printf("'Hi' from src/grotest/sixthseashell.gro:somedir/myfour(%T).DoIt\n", t)
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/yourthree/yourthree/mycomplex128/generic_args.go": `package myfour

type T = complex128`}},

//--------------------------------------------------------------------------------
//nested generics: parameterized package has 2 invoked imports
	{
		num: 540,
		fnm: "grotest/dud.gro",
		src:`import yourstruct "github.com/grolang/gro/syntax/grotest/yourthree" (float64, struct{a,b int})
"fmt".Println("'Hi' from src/grotest/sixthseashell.gro")
do yourstruct.RunIt()

package yourthree (S, T)
import mycomplex128 "github.com/grolang/gro/syntax/grotest/somedir/myfour" (complex128)
import myint "github.com/grolang/gro/syntax/grotest/somedir/myfour" (int)
func RunIt(){
	var s S
	var t T
	"fmt".Printf("'Hi' from src/grotest/sixthseashell.gro:yourthree(%T, %T).RunIt\n", s, t)
	mycomplex128.DoIt()
}

package "somedir" myfour (T)
func DoIt(){
	var t T
	"fmt".Printf("'Hi' from src/grotest/sixthseashell.gro:somedir/myfour(%T).DoIt\n", t)
}`,

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"grotest/dud.go": `// +build ignore

package main

import (
	fmt "fmt"
)

import yourstruct "github.com/grolang/gro/syntax/grotest/generics/dud/yourstruct"

func init() {
	fmt.Println("'Hi' from src/grotest/sixthseashell.gro")
	yourstruct.RunIt()
}

func main() {}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/yourthree.go": `package yourthree

import (
	fmt "fmt"
)

import mycomplex128 "github.com/grolang/gro/syntax/grotest/generics/yourthree/yourthree/mycomplex128"
import myint "github.com/grolang/gro/syntax/grotest/generics/yourthree/yourthree/myint"

func RunIt() {
	var s S
	var t T
	fmt.Printf("'Hi' from src/grotest/sixthseashell.gro:yourthree(%T, %T).RunIt\n", s, t)
	mycomplex128.DoIt()
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/generic_args.go": `package yourthree

type S = float64
type T = struct {
	a, b int
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/yourthree/yourthree/mycomplex128/myfour.go": `package myfour

import (
	fmt "fmt"
)

func DoIt() {
	var t T
	fmt.Printf("'Hi' from src/grotest/sixthseashell.gro:somedir/myfour(%T).DoIt\n", t)
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/yourthree/yourthree/mycomplex128/generic_args.go": `package myfour

type T = complex128`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/yourthree/yourthree/myint/myfour.go": `package myfour

import (
	fmt "fmt"
)

func DoIt() {
	var t T
	fmt.Printf("'Hi' from src/grotest/sixthseashell.gro:somedir/myfour(%T).DoIt\n", t)
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/yourthree/yourthree/myint/generic_args.go": `package myfour

type T = int`}},

//--------------------------------------------------------------------------------
//nested generics: using package parameter as invoked import argument
	{
		num: 550,
		fnm: "grotest/dud.gro",
		src:`import yourstruct "github.com/grolang/gro/syntax/grotest/yourthree" (float64, struct{a,b int})
"fmt".Println("'Hi' from src/grotest/sixthseashell.gro")
do yourstruct.RunIt()

package yourthree (S, T)
import mycomplex128 "github.com/grolang/gro/syntax/grotest/somedir/myfour" (complex128)
import myess "github.com/grolang/gro/syntax/grotest/somedir/myfour" (S)
func RunIt(){
	var s S
	var t T
	"fmt".Printf("'Hi' from src/grotest/sixthseashell.gro:yourthree(%T, %T).RunIt\n", s, t)
	mycomplex128.DoIt()
	myess.DoIt()
}

package "somedir" myfour (T)
func DoIt(){
	var t T
	"fmt".Printf("'Hi' from src/grotest/sixthseashell.gro:somedir/myfour(%T).DoIt\n", t)
}`,

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"grotest/dud.go": `// +build ignore

package main

import (
	fmt "fmt"
)

import yourstruct "github.com/grolang/gro/syntax/grotest/generics/dud/yourstruct"

func init() {
	fmt.Println("'Hi' from src/grotest/sixthseashell.gro")
	yourstruct.RunIt()
}

func main() {}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/yourthree.go": `package yourthree

import (
	fmt "fmt"
)

import mycomplex128 "github.com/grolang/gro/syntax/grotest/generics/yourthree/yourthree/mycomplex128"
import myess "github.com/grolang/gro/syntax/grotest/generics/dud/yourstruct/generics/yourthree/yourthree/myess"

func RunIt() {
	var s S
	var t T
	fmt.Printf("'Hi' from src/grotest/sixthseashell.gro:yourthree(%T, %T).RunIt\n", s, t)
	mycomplex128.DoIt()
	myess.DoIt()
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/generic_args.go": `package yourthree

type S = float64
type T = struct {
	a, b int
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/yourthree/yourthree/mycomplex128/myfour.go": `package myfour

import (
	fmt "fmt"
)

func DoIt() {
	var t T
	fmt.Printf("'Hi' from src/grotest/sixthseashell.gro:somedir/myfour(%T).DoIt\n", t)
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/yourthree/yourthree/mycomplex128/generic_args.go": `package myfour

type T = complex128`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/generics/yourthree/yourthree/myess/myfour.go": `package myfour

import (
	fmt "fmt"
)

func DoIt() {
	var t T
	fmt.Printf("'Hi' from src/grotest/sixthseashell.gro:somedir/myfour(%T).DoIt\n", t)
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/generics/yourthree/yourthree/myess/generic_args.go": `package myfour

type T = float64`}},

//--------------------------------------------------------------------------------
//nested generics: using package parameter as invoked import argument -- two such imports both with two pass-thru args
	{
		num: 560,
		fnm: "grotest/dud.gro",
		src:`import yourstruct "github.com/grolang/gro/syntax/grotest/yourthree" (
	complex128, int, float64, struct{a,b int},
)
"fmt".Println("'Hi' from src/grotest/sixthseashell.gro")
do yourstruct.RunIt()

package yourthree (Q, R, S, T)
import myque "github.com/grolang/gro/syntax/grotest/somedir/myfour" (Q, R)
import myess "github.com/grolang/gro/syntax/grotest/somedir/myfour" (S, T)
func RunIt(){
	var q Q
	var r R
	var s S
	var t T
	"fmt".Printf("'Hi' from src/grotest/sixthseashell.gro:yourthree(%T, %T, %T, %T).RunIt\n", q, r, s, t)
	myque.DoIt()
	myess.DoIt()
}

package "somedir" myfour (T, U)
func DoIt(){
	var t T
	var u U
	"fmt".Printf("'Hi' from src/grotest/sixthseashell.gro:somedir/myfour(%T, %T).DoIt\n", t, u)
}`,

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"grotest/dud.go": `// +build ignore

package main

import (
	fmt "fmt"
)

import yourstruct "github.com/grolang/gro/syntax/grotest/generics/dud/yourstruct"

func init() {
	fmt.Println("'Hi' from src/grotest/sixthseashell.gro")
	yourstruct.RunIt()
}

func main() {}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/yourthree.go": `package yourthree

import (
	fmt "fmt"
)

import myque "github.com/grolang/gro/syntax/grotest/generics/dud/yourstruct/generics/yourthree/yourthree/myque"
import myess "github.com/grolang/gro/syntax/grotest/generics/dud/yourstruct/generics/yourthree/yourthree/myess"

func RunIt() {
	var q Q
	var r R
	var s S
	var t T
	fmt.Printf("'Hi' from src/grotest/sixthseashell.gro:yourthree(%T, %T, %T, %T).RunIt\n", q, r, s, t)
	myque.DoIt()
	myess.DoIt()
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/generic_args.go": `package yourthree

type Q = complex128
type R = int
type S = float64
type T = struct {
	a, b int
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/generics/yourthree/yourthree/myque/myfour.go": `package myfour

import (
	fmt "fmt"
)

func DoIt() {
	var t T
	var u U
	fmt.Printf("'Hi' from src/grotest/sixthseashell.gro:somedir/myfour(%T, %T).DoIt\n", t, u)
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/generics/yourthree/yourthree/myque/generic_args.go": `package myfour

type T = complex128
type U = int`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/generics/yourthree/yourthree/myess/myfour.go": `package myfour

import (
	fmt "fmt"
)

func DoIt() {
	var t T
	var u U
	fmt.Printf("'Hi' from src/grotest/sixthseashell.gro:somedir/myfour(%T, %T).DoIt\n", t, u)
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/generics/yourthree/yourthree/myess/generic_args.go": `package myfour

type T = float64
type U = struct {
	a, b int
}`}},

//--------------------------------------------------------------------------------
//nested generics
	{
		num: 570,
		fnm: "grotest/dud.gro",
		src:`import yourstruct "github.com/grolang/gro/syntax/grotest/yourthree" (float64, struct{a,b int})
"fmt".Println("'Hi' from src/grotest/sixthseashell.gro")
do yourstruct.RunIt()

package yourthree (S, T)
import myess "github.com/grolang/gro/syntax/grotest/somedir/myfour" (S)
func RunIt(){
	var s S
	var t T
	"fmt".Printf("'Hi' from src/grotest/sixthseashell.gro:yourthree(%T, %T).RunIt\n", s, t)
	myess.DoIt()
}

package "somedir" myfour (T)
import mytee "github.com/grolang/gro/syntax/grotest/somedir/theirfive" (T)
func DoIt(){
	var t T
	"fmt".Printf("'Hi' from src/grotest/sixthseashell.gro:somedir/myfour(%T).DoIt\n", t)
	mytee.DoIt()
}

package "somedir" theirfive (U)
func DoIt(){
	var u U
	"fmt".Printf("'Hi' from src/grotest/sixthseashell.gro:somedir/theirfive(%T).DoIt\n", u)
}`,

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		prt:map[string]string{
"grotest/dud.go": `// +build ignore

package main

import (
	fmt "fmt"
)

import yourstruct "github.com/grolang/gro/syntax/grotest/generics/dud/yourstruct"

func init() {
	fmt.Println("'Hi' from src/grotest/sixthseashell.gro")
	yourstruct.RunIt()
}

func main() {}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/yourthree.go": `package yourthree

import (
	fmt "fmt"
)

import myess "github.com/grolang/gro/syntax/grotest/generics/dud/yourstruct/generics/yourthree/yourthree/myess"

func RunIt() {
	var s S
	var t T
	fmt.Printf("'Hi' from src/grotest/sixthseashell.gro:yourthree(%T, %T).RunIt\n", s, t)
	myess.DoIt()
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/generic_args.go": `package yourthree

type S = float64
type T = struct {
	a, b int
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/generics/yourthree/yourthree/myess/myfour.go": `package myfour

import (
	fmt "fmt"
)

import mytee "github.com/grolang/gro/syntax/grotest/generics/dud/yourstruct/generics/yourthree/yourthree/myess/generics/somedir/myfour/myfour/mytee"

func DoIt() {
	var t T
	fmt.Printf("'Hi' from src/grotest/sixthseashell.gro:somedir/myfour(%T).DoIt\n", t)
	mytee.DoIt()
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/generics/yourthree/yourthree/myess/generic_args.go": `package myfour

type T = float64`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/generics/yourthree/yourthree/myess/generics/somedir/myfour/myfour/mytee/theirfive.go": `package theirfive

import (
	fmt "fmt"
)

func DoIt() {
	var u U
	fmt.Printf("'Hi' from src/grotest/sixthseashell.gro:somedir/theirfive(%T).DoIt\n", u)
}`,

// - - - - - - - - - - - - - - - - - - - -
"grotest/generics/dud/yourstruct/generics/yourthree/yourthree/myess/generics/somedir/myfour/myfour/mytee/generic_args.go": `package theirfive

type U = float64`}},

//--------------------------------------------------------------------------------
//nested generics: using package parameter as invoked import argument, with a cycle
	{
		num: 580,
		fnm: "grotest/dud.gro",
		src:`import yourstruct "github.com/grolang/gro/syntax/grotest/yourthree" (complex128, int, float64, struct{a,b int})
"fmt".Println("'Hi' from src/grotest/sixthseashell.gro")
do yourstruct.RunIt()

package yourthree (Q, R, S, T)
import myque "github.com/grolang/gro/syntax/grotest/somedir/myfour" (Q, R)
import myess "github.com/grolang/gro/syntax/grotest/somedir/myfour" (S, T)
func RunIt(){
	var q Q
	var r R
	var s S
	var t T
	"fmt".Printf("'Hi' from src/grotest/sixthseashell.gro:yourthree(%T, %T, %T, %T).RunIt\n", q, r, s, t)
	myque.DoIt()
	myess.DoIt()
}

package "somedir" myfour (T, U)
import yourints "github.com/grolang/gro/syntax/grotest/yourthree" (int, int, int, int)
func DoIt(){
	var t T
	var u U
	"fmt".Printf("'Hi' from src/grotest/sixthseashell.gro:somedir/myfour(%T, %T).DoIt\n", t, u)
}`,
		//err: "grotest/dud.gro:24:2: syntax error: cycles not allowed for parameterized imports",
		err: "grotest/dud.gro:24:2: syntax error: parameterized package not present in file, or there's a cycle in the parameterized imports",
},

//--------------------------------------------------------------------------------
})}

