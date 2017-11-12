// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"testing"
)

//================================================================================
func TestMacros(t *testing.T) {
	groTest(t, groTestData{

		//--------------------------------------------------------------------------------
		{
			num: 100,
			fnm: "dud.gro",
			src: `import "fmt"
prepare "afile.gro"
func main() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `// +build ignore

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
		{
			num: 110,
			fnm: "dud.gro",
			src: `import "fmt"
execute "afile.gro"
func main() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `// +build ignore

package main

import (
	sys "github.com/grolang/gro/sys"
)

import "fmt"

func init() {
	sys.Execute("afile.gro")
}

func main() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
		{
			num: 120,
			fnm: "dud.gro",
			src: `import "fmt"
run "afile.go"
func main() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `// +build ignore

package main

import (
	sys "github.com/grolang/gro/sys"
)

import "fmt"

func init() {
	sys.Run("afile.go")
}

func main() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
		{
			num: 200,
			fnm: "dud.gro",
			src: `do a:= 1
do a = 2
let b = 3
do c:= 4
let d, a = 7, 8
do e = 9
`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `// +build ignore

package main

func init() {
	a := 1
	a = 2
	{
		b := 3
		c := 4
		{
			d, a := 7, 8
			e = 9
		}
	}
}

func main() {}`}},

		//--------------------------------------------------------------------------------
		{
			num: 250,
			fnm: "dud.gro",
			src: `assert 1 == 1
assert !false
do d:= 1
assert d == 1`,

			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `// +build ignore

package main

import (
	assert "github.com/grolang/gro/assert"
)

func init() {
	assert.AssertTrue(1 == 1)
	assert.AssertTrue(!false)
	d := 1
	assert.AssertTrue(d == 1)
}

func main() {}`}},

		//--------------------------------------------------------------------------------
	})
}

//================================================================================
func TestUseDecls(t *testing.T) {
	groTest(t, groTestData{
		//--------------------------------------------------------------------------------
		{
			num: 100,
			fnm: "dud.gro",
			src: `project myproj
use a b c "myuse"
package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
		// "include" cmd
		{
			num: 200,
			fnm: "dud.gro",
			src: `include "mymy"
package def
func run() {
	fmt.Println("Hello, world!")
}`,
			xtr: map[string]string{
				"mymy": ``},

			// - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package def

func run() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
		// "include" cmd group
		{
			num: 201,
			fnm: "dud.gro",
			src: `include (
	"mymy"
	"youyou"
)
package def
func run() {
	fmt.Println("Hello, world!")
}`,
			xtr: map[string]string{
				"mymy":   ``,
				"youyou": ``},

			// - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package def

func run() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
		// multiple "include" cmds
		{
			num: 202,
			fnm: "dud.gro",
			src: `include (
	"mymy"
	"youyou"
)
include "itit"
package def
func run() {
	fmt.Println("Hello, world!")
}`,
			xtr: map[string]string{
				"mymy":   ``,
				"youyou": ``,
				"itit":   ``},

			// - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package def

func run() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
		// multiple "include" cmds using subdirs
		{
			num: 203,
			fnm: "adir/dud.gro",
			src: `include (
	"mymy"
	"anotherdir/youyou"
)
include "itit"
package def
func run() {
	fmt.Println("Hello, world!")
}`,
			xtr: map[string]string{
				"adir/mymy":              ``,
				"adir/anotherdir/youyou": ``,
				"adir/itit":              ``},

			// - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"adir/dud.go": `package def

func run() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
		// multiple "include" and "use" cmds mixed up
		{
			num: 220,
			fnm: "dud.gro",
			src: `include (
	"mymy"
	"anotherdir/youyou"
)
use "blacklist"
include "itit"
package def
func run() {
	fmt.Println("Hello, world!")
}`,
			xtr: map[string]string{
				"mymy":              ``,
				"anotherdir/youyou": ``,
				"itit":              ``},

			// - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package def

func run() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
	})
}

//================================================================================
func TestDynamic(t *testing.T) {
	groTest(t, groTestData{
		//--------------------------------------------------------------------------------
		//use "dynamic" causes "type any" to be added
		{
			num: 100,
			fnm: "dud.gro",
			src: `use"dynamic"
package abc
import "fmt"
var v int; const c = 2; type t int
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import fmt "fmt"

var v int

const c = 2

type t int

func run() {
	fmt.Println("Hello, world!")
}

type any = interface{}`}},

		//--------------------------------------------------------------------------------
		//use "dynamic" but "any" already in use (as TYPE) causes "type any" NOT to be added
		{
			num: 110,
			fnm: "dud.gro",
			src: `use"dynamic"
package abc
import "fmt"
var v int; const c = 2; type t int

type any string

func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import fmt "fmt"

var v int

const c = 2

type t int
type any string

func run() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
		//use "dynamic" but "any" already in use (as VAR) causes "type any" NOT to be added
		{
			num: 120,
			fnm: "dud.gro",
			src: `use"dynamic"
package abc
import (
	"fmt"
	"path/filepath"
)
var v int; const c = 2; type t int

var any = 2

func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import (
	fmt "fmt"
	filepath "path/filepath"
)

var v int

const c = 2

type t int

var any = 2

func run() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
		//use "dynamic" but "any" already in use (as CONST inside a group) causes "type any" NOT to be added
		{
			num: 130,
			fnm: "dud.gro",
			src: `use"dynamic"
package abc
import "fmt"
var v int; const c = 2; type t int
const (
	any = 2
	another = 3
)
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import fmt "fmt"

var v int

const c = 2

type t int

const (
	any = 2
	another = 3
)

func run() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
		//use "dynamic" but "any" already in use (as CONST in a list) causes "type any" NOT to be added
		{
			num: 140,
			fnm: "dud.gro",
			src: `use"dynamic"
package abc
import "fmt"
var v int; const c = 2; type t int

const any, another = 2, 3

func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import fmt "fmt"

var v int

const c = 2

type t int

const any, another = 2, 3

func run() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
		//use "dynamic" but "any" already in use (as IMPORT local name) causes "type any" NOT to be added
		{
			num: 150,
			fnm: "dud.gro",
			src: `use"dynamic"
package abc
import "fmt"
import any "my/path/to/swh"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import fmt "fmt"
import any "my/path/to/swh"

func run() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
		//use "dynamic" but "any" already in use (as default IMPORT name) causes "type any" NOT to be added
		{
			num: 160,
			fnm: "dud.gro",
			src: `use"dynamic"
package abc
import "fmt"
import "my/path/to/any"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import fmt "fmt"
import any "my/path/to/any"

func run() {
	fmt.Println("Hello, world!")
}`}},

		//--------------------------------------------------------------------------------
		//use "dynamic" with return value (here, "dyn") causes "any" to be added, ...
		//...and some operator overloading
		{
			num: 200,
			fnm: "dud.gro",
			src: `use dyn "dynamic"
package abc
import "fmt"
do fmt.Println(4 + 5)
do {
	fmt.Println(true && false)
	fmt.Println(-true)
}
dyn {
	"fmt".Println(4 + 5) //in real code, better not use "fmt" in-place when "fmt" already imported
	"fmt".Println(true && false)
	do fmt.Println(-true)
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import (
	fmt "fmt"
	dyn "github.com/grolang/gro/ops"
)

import fmt "fmt"

func init() {
	fmt.Println(4 + 5)
	{
		fmt.Println(true && false)
		fmt.Println(-true)
	}
	{
		fmt.Println(dyn.Plus(4, 5))
		fmt.Println(dyn.And(true, func() interface{} {
			return false
		}))
		fmt.Println(dyn.Negate(true))
	}
}

type any = interface{}`}},

		//--------------------------------------------------------------------------------
		//use "dynamic" with return value (here, "dyn") causes dynamic interp of strings and runes
		{
			num: 210,
			fnm: "dud.gro",
			src: `use dyn "dynamic"
package abc
do {
	"fmt".Println("abcdefg")
	"fmt".Println('a')
}
dyn {
	"fmt".Println("abcdefg")
	"fmt".Println("hijklmnop")
	"fmt".Println('a')
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import (
	fmt "fmt"
	utf88 "github.com/grolang/gro/utf88"
	dyn "github.com/grolang/gro/ops"
)

func init() {
	{
		fmt.Println("abcdefg")
		fmt.Println('a')
	}
	{
		fmt.Println(utf88.Desur("abcdefg"))
		fmt.Println(utf88.Desur("hijklmnop"))
		fmt.Println(dyn.Runex("a"))
	}
}

type any = interface{}`}},

		//--------------------------------------------------------------------------------
		//use _ "dynamic" should enable full dynamic mode
		{
			num: 220,
			fnm: "dud.gro",
			src: `use _ "dynamic"
package abc
do {
	"fmt".Println("abcdefg")
	"fmt".Println("hijklmnop")
	"fmt".Println('a')
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import (
	fmt "fmt"
	utf88 "github.com/grolang/gro/utf88"
	groo "github.com/grolang/gro/ops"
)

func init() {
	{
		fmt.Println(utf88.Desur("abcdefg"))
		fmt.Println(utf88.Desur("hijklmnop"))
		fmt.Println(groo.Runex("a"))
	}
}

type any = interface{}`}},

		//--------------------------------------------------------------------------------
		//the "groo" file type should enable full dynamic mode
		{
			num: 230,
			fnm: "dud.groo", // <- note ".groo"
			src: `package abc
do {
	"fmt".Println("abcdefg")
	"fmt".Println("hijklmnop")
	"fmt".Println('a')
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import (
	fmt "fmt"
	utf88 "github.com/grolang/gro/utf88"
	groo "github.com/grolang/gro/ops"
)

func init() {
	{
		fmt.Println(utf88.Desur("abcdefg"))
		fmt.Println(utf88.Desur("hijklmnop"))
		fmt.Println(groo.Runex("a"))
	}
}

type any = interface{}`}},

		//--------------------------------------------------------------------------------
	})
}

//================================================================================
