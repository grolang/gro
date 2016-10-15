// Copyright 2016 The Gro Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package basic

import (
	"github.com/grolang/gro/tutor"
)

//================================================================================

var TutTodo = &tutor.Tutorial{
	Name:      "todo",
	Short:     "describes tentative features of Gro",
	Long:      `
Todo is a tutorial to tentative features of Gro.

`,
	Pages:     []*tutor.Page{
		todoMain,
		todoPackage,
		todoSyntacticMacros,
	},
}

//================================================================================

var todoMain = &tutor.Page{
	English: []string{`
Tentative Features

The following features aren't yet supported.

`},
}

//================================================================================

var todoPackage = &tutor.Page{
	English: []string{`
Tentative Package Features

Packages defined within a ".gro" file can also be aliased with Unihan. Use "包吧thinking", or 包"someDir"吧thinking if the directory is specified.

TODO: Packages can be pre-declared with the "预" Unihan. Put "预吧thinking" at the top level somewhere before its first use, followed up with 包"someDir"thinking or 包"someDir"thinking at the package definition.

There are thus 3 ways to use packages in Gro:

	use the globally known Unihan alias, e.g. "形" for "fmt"
	import it explicitly at the top of the source just like in Go
	define it in the same source file

`},
}

//================================================================================

var todoSyntacticMacros = &tutor.Page{
	Code: []string{`
	准"src/github.com/grolang/samples/goByEg4.gro" //runs "gro prepare" command
	跑"src/github.com/grolang/samples/goByEg4.go" //runs "go run" command
`,
`
	用㕧"github.com/grolang/samples/moremacs" //before the package; directory must be used
	用㕨"github.com/grolang/samples/somemacs"
	包正       //package
	源"hey.go" //源(source)
	入"fmt"    //import
	种A整;     //type
	功正(){    //func main()
		㕧a:=9 //macro could head a statement
		㕧{    //macro could enclose a block
			a=123
		}
		fmt.Println(㕨(a)) //macro could act like a builtin function
	}
`},

	English: []string{`
Syntactic macros

Gro brings hygienic macros to Go. The 2 ways to use macros are:

	use a globally known Unihan alias
	import it explicitly from another directory at the top of the source

Besides the Unihan for Go keywords and special identifiers, there are many more that can be used for defining statements, built-in functions, and built-in types, and they are specified with macros.

Two of these are "准" and "跑", used for building Go and Gro source files:

`,

`
The "用"(use) command is for importing a macro for use in a package. It must be used with a temporary Unihan, any with the "口" radical on the left. The macro Unihan could be used like a statement keyword, a built-in function name, or a built-in type name.

`,

`
TODO: Macros as Statements:

Macro syntax should follow the syntactic style of elements already in the Go language. For statements, they typically begin with a keyword (e.g. "为"for, "去"go),perhaps followed by an expression (e.g. 回"abc"), then perhaps a curly-enclosed block. Perhaps there's many semicolon-separated expressions or simple statements (e.g. "为i:=0;i<10;i++{}") before the curlies, or perhaps there's many statements following (e.g. case's "事789: f(x); g(y); h(z)"). Perhaps another keyword follows the curlies (e.g. if-else's "如a==b{回1}否{回2}").

Optional Unihan 让let (TODO: and 叫call) help bring consistency to this syntactic style.

TODO: Macros as Built-in Expressions:

Built-in expressions typically begin with a function name followed by other expressions or types within parens (e.g. append's "加(x,a,b)"). Sometimes the parens are omitted (e.g. range's "围x"), empty (e.g. recover's "抓()"), or both (e.g. "毫"iota). Because Gro's Unihan syntax blurs the line between keywords and special identifiers, "range" is considered a built-in expression.

TODO: Macros as Built-in Types:

Built-in types typically begin with a name (e.g. "串"string), and can be followed by brackets (e.g. map's "图[X]Y"), curlies (e.g. structs "构{X节}"), a number (e.g. int's "整8"), or something even more complicated (e.g. func's "功(串)双").

Optional Unihan 这 for same-package names helps bring consistency to this syntactic style. (TODO: Optional Unihan 切 for slices and arrays also helps this.)

`},
}

//================================================================================

