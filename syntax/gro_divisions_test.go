// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"testing"
)

//================================================================================
func TestDivisions(t *testing.T) {
	groTest(t, groTestData{
		//--------------------------------------------------------------------------------
		//single "package" directive
		{
			num: 10,
			fnm: "dud.gro",
			src: `package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "package" directive; file in subdirectory
		{
			num: 11,
			fnm: "adir/dud.gro",
			src: `package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"adir/dud.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "package" directive with explicit directory "."
		{
			num: 15,
			fnm: "dud.gro",
			src: `package "." abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single implicit package
		{
			num: 20,
			fnm: "dud.gro",
			src: `import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//implicit package and explicit "package"; semicolon after 1st package defn
		{
			num: 30,
			fnm: "dud.gro",
			src: `import "fmt"
func run() {
	fmt.Println("Hello, world!")
};
package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,

				"abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//implicit package and explicit "package"; semicolon after 1st package defn
		{
			num: 31,
			fnm: "adir/dud.gro",
			src: `import "fmt"
func run() {
	fmt.Println("Hello, world!")
};
package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"adir/dud.go": `package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,

				"adir/abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		// 2 explicit "package"s - doc to recommend using "project" keyword
		{
			num: 40,
			fnm: "dud.gro",
			src: `package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package defg
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
}
`,
				"defg/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		// 2 explicit "package"s, one with "." directory - doc to recommend using "project" keyword
		{
			num: 45,
			fnm: "dud.gro",
			src: `package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package "." defg
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
}
`,
				"defg/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//implicit package and 3 explicit "package"s -- 2 with directory strings
		{
			num: 50,
			fnm: "dud.gro",
			src: `import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package "somebase" defg
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package "somebase/defg" hij
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{

				"dud.go": `package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,

				"abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,

				"somebase/defg/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,

				"somebase/defg/hij/hij.go": `package hij

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		// 2 packages, one headed with "package", one headed with "internal"
		{
			num: 60,
			fnm: "dud.gro",
			src: `package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
internal defg
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
}
`,
				"internal/defg/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		// project dir-string and both "internal" and normal packages
		{
			num: 61,
			fnm: "dud.gro",
			src: `project "projlevel" myproj
package "packlevel" abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
internal "another" defg
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"projlevel/packlevel/abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"projlevel/internal/another/defg/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "project" and "package" directives
		{
			num: 100,
			fnm: "dud.gro",
			src: `project myproj
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
}
`}},

		//--------------------------------------------------------------------------------
		//single "project" with 2 explicit "package"s
		{
			num: 110,
			fnm: "dud.gro",
			src: `project mypack
package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package defg
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
}
`,

				"defg/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "project" with 2 explicit "package"s, and project dir-string
		{
			num: 111,
			fnm: "dud.gro",
			src: `project "deep/down" mypack
package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package defg
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"deep/down/abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,

				"deep/down/defg/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "project" with implicit package and explicit "package"
		{
			num: 120,
			fnm: "dud.gro",
			src: `project mypack
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"mypack.go": `package mypack

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,

				"abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "package" directive and single "section" directive
		{
			num: 200,
			fnm: "dud.gro",
			src: `package abc
section "this"
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"this.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "package" directive and 2 "section" directives
		{
			num: 210,
			fnm: "dud.gro",
			src: `package abc
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
				"this.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"that.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "section" directive
		{
			num: 220,
			fnm: "dud.gro",
			src: `section "this"
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"this.go": `package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single implicit "package", with both implicit and explicit "section"s
		{
			num: 230,
			fnm: "dud.gro",
			src: `import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
section "this"
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"this.go": `package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single explicit "package", with both implicit and explicit "section"s
		{
			num: 240,
			fnm: "dud.gro",
			src: `package abc
import "fmt"
func run() {
	fmt.Println("Nihao, world!")
}
section "this"
import "fmt"
func run() {
	fmt.Println("Konichiwa, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import "fmt"

func run() {
	fmt.Println("Nihao, world!")
}
`,
				"this.go": `package abc

import "fmt"

func run() {
	fmt.Println("Konichiwa, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//implicit package (with both implicit and explicit "section") and explicit "package" (with both implicit and explicit "section")
		{
			num: 250,
			fnm: "dud.gro",
			src: `import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
section "afile"
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
section "bfile"
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{

				"dud.go": `package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"afile.go": `package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,

				"abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"abc/bfile.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//implicit package (with sections) and 2 explicit "package"s with directory strings (each with sections)
		{
			num: 260,
			fnm: "dud.gro",
			src: `import "fmt"
func run() { fmt.Println("Hello, world!") }

section "afile"
import "fmt"
func run() { fmt.Println("Hello, world!") }

package "somebase" defg
section "bfile"
import "fmt"
func run() { fmt.Println("Hello, world!") }

section "cfile"
import "fmt"
func run() { fmt.Println("Hello, world!") }

package "somebase/defg" hij
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}

section "dfile"
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"afile.go": `package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"somebase/defg/bfile.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"somebase/defg/cfile.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"somebase/defg/hij/hij.go": `package hij

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"somebase/defg/hij/dfile.go": `package hij

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "project", "package" and "section" directives
		{
			num: 300,
			fnm: "dud.gro",
			src: `project mypack
package abc
section "myfile"
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"abc/myfile.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "project" with implicit package (with implicit and explicit "section") and explicit "package" (with implicit and explicit "section")
		{
			num: 310,
			fnm: "dud.gro",
			src: `project mypack
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
section "afile"
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
section "bfile"
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"mypack.go": `package mypack

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"afile.go": `package mypack

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,

				"abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"abc/bfile.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//empty file
		{
			num: 400,
			fnm: "dud.gro",
			src: ``,
			err: "dud.gro:1:1: syntax error: gro-file empty",
		},

		//--------------------------------------------------------------------------------
		//"project" clause only
		{
			num: 410,
			fnm: "dud.gro",
			src: `project meta`,
			err: "dud.gro:1:13: syntax error: project keyword but no packages",
		},

		//--------------------------------------------------------------------------------
		//bad top-level keyword at beginning
		{
			num: 411,
			fnm: "dud.gro",
			src: `gobblemeup 789`,
			err: "dud.gro:1:1: syntax error: unexpected name at top-level",
		},

		//--------------------------------------------------------------------------------
		//bad top-level keyword after project clause
		{
			num: 421,
			fnm: "dud.gro",
			src: `project abc
gobbledegook "this"
import "fmt"
func main() {
	fmt.Println("abc/this")
}`,
			err: "dud.gro:2:1: syntax error: unexpected name at top-level",
		},

		//--------------------------------------------------------------------------------
		//bad top-level keyword after package clause
		{
			num: 422,
			fnm: "dud.gro",
			src: `package abc
gobbledegook "this"
import "fmt"
func main() {
	fmt.Println("abc/this")
}`,
			err: "dud.gro:2:1: syntax error: unexpected name at top-level",
		},

		//--------------------------------------------------------------------------------
		//TODO: dangling close paren
		{
			num: 510,
			fnm: "dud.gro",
			src: `"fmt".Println("'Hi' from src/grotest/thirdshot.gro")
"grotest/basicsecond".RunIt()
}
`,
			err: "dud.gro:3:1: syntax error: unexpected token }",
		},

		//--------------------------------------------------------------------------------
		//TODO: dangling close paren
		{
			num: 511,
			fnm: "dud.gro",
			src: `}`,
			err: "dud.gro:1:1: syntax error: unexpected token }",
		},

		//--------------------------------------------------------------------------------
		//TODO: dangling open paren
		{
			num: 512,
			fnm: "dud.gro",
			src: `{`,
			err: "dud.gro:1:2: syntax error: unexpected EOF, expecting }",
		},

		//--------------------------------------------------------------------------------
	})
}

//================================================================================
func TestMain(t *testing.T) {
	groTest(t, groTestData{
		//--------------------------------------------------------------------------------
		//single "package" directive (with main fn)
		{
			num: 10,
			fnm: "dud.gro",
			src: `package abc
import "fmt"
func main() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single implicit package (with main fn)
		{
			num: 20,
			fnm: "dud.gro",
			src: `import "fmt"
func main() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `// +build ignore

package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//implicit package (with main fn) and explicit "package"; semicolon after 1st package defn
		{
			num: 30,
			fnm: "dud.gro",
			src: `import "fmt"
func main() {
	fmt.Println("Hello, world!")
};
package abc
import "fmt"
func main() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `// +build ignore

package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}
`,

				"abc/abc.go": `package abc

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "project" and "package" directives (with main fn)
		{
			num: 100,
			fnm: "dud.gro",
			src: `project myproj
package abc
import "fmt"
func main() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"abc/abc.go": `package abc

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "project" with implicit package (with main fn) and explicit "package"
		{
			num: 120,
			fnm: "dud.gro",
			src: `project mypack
import "fmt"
func main() {
	fmt.Println("Hello, world!")
}
package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"mypack.go": `// +build ignore

package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}
`,

				"abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "package" directive and single "main" directive
		{
			num: 200,
			fnm: "dud.gro",
			src: `package abc
main "this"
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"this.go": `// +build ignore

package main

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}

func main() {}
`}},

		//--------------------------------------------------------------------------------
		//single "package" directive and single "testcode" directive
		{
			num: 201,
			fnm: "dud.gro",
			src: `package abc
testcode "this"
import "fmt"
func TestRun() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"this_test.go": `package abc

import "fmt"

func TestRun() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "package" directive and single "testcode" directive, with imported package of same name
		{
			num: 202,
			fnm: "dud.gro",
			src: `package abc
testcode "this"
import "abc"
func TestRun() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"this_test.go": `package abc_test

import "abc"

func TestRun() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
	})
}

//================================================================================
func TestCurlies(t *testing.T) {
	groTest(t, groTestData{
		//--------------------------------------------------------------------------------
		//single curlied "package"
		{
			num: 100,
			fnm: "dud.gro",
			src: `package abc {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//implicit package and explicit curlied "package"
		{
			num: 110,
			fnm: "dud.gro",
			src: `import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package abc {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
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

				"abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		// 2 explicit curlied "package"s
		{
			num: 120,
			fnm: "dud.gro",
			src: `package abc {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
}
package defg {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{

				"abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"defg/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//implicit package, 2 curlied packages (1 with directory string), 1 go-style package (with dir string)
		{
			num: 130,
			fnm: "dud.gro",
			src: `import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
package abc {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
};
package "somebase" defg {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
}
package "somebase/defg" hij
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{

				"dud.go": `package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,

				"abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,

				"somebase/defg/defg.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,

				"somebase/defg/hij/hij.go": `package hij

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "project" and curlied package
		{
			num: 140,
			fnm: "dud.gro",
			src: `project mypack
package abc {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "project" and curlied internal
		{
			num: 150,
			fnm: "dud.gro",
			src: `project mypack
internal abc {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"internal/abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single go-style package and single curlied section
		{
			num: 200,
			fnm: "dud.gro",
			src: `package abc
section "this" {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"this.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single curlied package and single go-style section
		{
			num: 201,
			fnm: "dud.gro",
			src: `package abc {
section "this"
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"this.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single curlied package and single curlied section
		{
			num: 202,
			fnm: "dud.gro",
			src: `package abc {
	section "this" {
		import "fmt"
		func run() {
			fmt.Println("Hello, world!")
		}
	}
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"this.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "package" directive and 2 curlied "section" directives
		{
			num: 210,
			fnm: "dud.gro",
			src: `package abc
section "this" {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
};
section "that" {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"this.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"that.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single curlied "section" directive
		{
			num: 220,
			fnm: "dud.gro",
			src: `section "this" {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"this.go": `package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single implicit "package", with both implicit and curlied "section"s
		{
			num: 230,
			fnm: "dud.gro",
			src: `import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
section "this" {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
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
				"this.go": `package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single go-style "package", with both implicit and curlied "section"s
		{
			num: 240,
			fnm: "dud.gro",
			src: `package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
}
section "this" {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"this.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single curlied "package", with both implicit and curlied "section"s
		{
			num: 241,
			fnm: "dud.gro",
			src: `package abc {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
	section "this" {
		import "fmt"
		func run() {
			fmt.Println("Hello, world!")
		}
	}
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"this.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//implicit package (with curlied sections) and 2 curlied "package"s with directory strings (each with curlied sections)
		{
			num: 250,
			fnm: "dud.gro",
			src: `import "fmt"
func run() { fmt.Println("Hello, world!") }

section "afile" {
	import "fmt"
	func run() { fmt.Println("Hello, world!") }
}
package "somebase" defg {
	section "bfile" {
		import "fmt"
		func run() { fmt.Println("Hello, world!") }
		//semicolon follows...
	};
	section "cfile" {
		import "fmt"
		func run() { fmt.Println("Hello, world!") }
	}
//semicolon follows...
};
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
				"afile.go": `package dud

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"somebase/defg/bfile.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"somebase/defg/cfile.go": `package defg

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"somebase/defg/hij/hij.go": `package hij

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`,
				"somebase/defg/hij/dfile.go": `package hij

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//single "project" and curlied "package" directive
		{
			num: 260,
			fnm: "dud.gro",
			src: `project mypack
package abc {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"abc/abc.go": `package abc

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//project keyword, and single curlied "section" directive
		{
			num: 270,
			fnm: "dud.gro",
			src: `project myproj
section "this" {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"this.go": `package myproj

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
		//project keyword, and single curlied "main" directive
		{
			num: 280,
			fnm: "dud.gro",
			src: `project myproj
main "this" {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"this.go": `// +build ignore

package main

import "fmt"

func run() {
	fmt.Println("Hello, world!")
}

func main() {}
`}},

		//--------------------------------------------------------------------------------
		//single curlied "package", with both implicit "section" and curlied "testcode"
		{
			num: 290,
			fnm: "dud.gro",
			src: `package abc {
	import "fmt"
	func run(s string) {
		fmt.Println("Hello, world!")
	}
	testcode "this" {
		func TestRun() {
			run("Hello, world!")
		}
	}
}`,
			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package abc

import "fmt"

func run(s string) {
	fmt.Println("Hello, world!")
}
`,
				"this_test.go": `package abc

func TestRun() {
	run("Hello, world!")
}
`}},

		//--------------------------------------------------------------------------------
	})
}

//================================================================================
func TestShorthandAliases(t *testing.T) {
	groTest(t, groTestData{
		//--------------------------------------------------------------------------------
		//string shorthand for imported packages
		{
			num: 100,
			fnm: "dud.gro",
			src: `package you
func main(){
	"my/dir/path".Join("Hi, you!")
	"another/fmt".Println("Bye bye.")
	"third/way/hoot".Println("Bye bye.")
}`,

			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `package you

import (
	path "my/dir/path"
	fmt "another/fmt"
	hoot "third/way/hoot"
)

func main() {
	path.Join("Hi, you!")
	fmt.Println("Bye bye.")
	hoot.Println("Bye bye.")
}
`}},

		//--------------------------------------------------------------------------------
		//string shorthand in one-liner
		{
			num: 110,
			fnm: "dud.gro",
			src: `do"fmt".Println("Hi, earth!")`,

			// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			prt: map[string]string{
				"dud.go": `// +build ignore

package main

import (
	fmt "fmt"
)

func init() {
	fmt.Println("Hi, earth!")
}

func main() {}
`}},

		//--------------------------------------------------------------------------------
		{
			num: 120,
			fnm: "dud.gro",
			src: `package abc
func main() {
	"my/path/here".SomeFunc()
	"another/path/here".AnotherFn()
}`,
			err: "dud.gro:4:21: syntax error: import alias \"here\" has already been used but with different import path",
		},

		//--------------------------------------------------------------------------------
	})
}
