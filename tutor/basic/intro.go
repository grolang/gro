// Copyright 2016 The Gro Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package basic

import (
	"github.com/grolang/gro/tutor"
)

//================================================================================

var TutIntro = &tutor.Tutorial{
	Name:      "intro",
	Short:     "introduces the basics of Gro",
	Long:      `
Intro is a tutorial to introduce the basics of Gro.

`,
	Pages:     []*tutor.Page{
		introMain,
		introSummary,
		introHierarchy,
		introCommands,
		introSpaceless,
		introRule1,
		introRule2,
		introRule3,
		introRule4,
		introRule5,
		introRule6,
		introImpliedPack,
		introMultipack,
		introExample,
	},
}

//================================================================================

var introMain = &tutor.Page{
	English: []string{`
What Gro Is

Gro is both a syntax that extends Go's, and a tool that generates Go code from Gro's.

Gro uses single Unihan characters as optional substitutes for keywords, special identifiers, and common package names, and as the names for hygienic macros. It does so by introducing one small prohibition to Go's syntax: prohibiting the use of Unihan in identifier names. Because every Unihan has an implied space both before and after it in the Gro grammar, using Unihan enables Gro source code to not have any whitespace, just as using semicolons enables Go code not to have any newlines.

Gro code and Unihan-free Go code can be mixed freely in the same source file. But when a program is written using Unihan in all the places it can be in the code, any of the 25 Go keywords can be used as identifier or label names, so a dedicated Gro programmer doesn't need to know any Go-specific naming exceptions to write Gro code. The Unihan, unlike other non-Ascii characters, are easily enterable via the many IME's (input method editors) available for Chinese and Japanese that ship for free on OS's such as Linux and Windows, so Gro code can be typed in quickly.

`},
}

//================================================================================

var introSummary = &tutor.Page{
	English: []string{`
What Gro adds to Go

Gro adds to Go:

	* ability to use Unihan as keywords, special identifiers, and package names
	* "package main" and "func main()" are optional so a one-line statement is a valid Gro script
	* many Go packages and sources can be in a single Gro source file
	* redundant imports are ignored, and missing common imports are inferred
	* commands (e.g. "run", "test") called the same way as statements (e.g. "for", "if")
	* "assert" keyword available

`},
}

//================================================================================

var introHierarchy = &tutor.Page{
	English: []string{`
Directory Hierarchy

After getting the source the hierarchy should be:

	GOPATH
	  +--bin
	  +--pkg
	  +--src
	  |   +--github.com
	  |       +--grolang
	  |           +--gro
	  |           |   +--[various libraries of go source code for gro]
	  |           |   +--LICENSE.txt, etc
	  |           +--groo (i.e. sample dynamic language addon for gro)
	  |           +--vy (i.e. sample IDE for groo)
	  |           +--samples
	  |               +--gro10
	  |               +--gro11
	  |               +--groo
	  +-tmp (i.e. temporary files used by grolang)

Build the gro runtime with "go install src/github.com/grolang/gro/cmd/gro.go".
You can add your own projects anywhere else under the src/ directory.

`},
}

//================================================================================

var introCommands = &tutor.Page{
	English: []string{`
Gro Commands

"gro prepare" translates Gro code and Unihan-free Go code mixed together into standard Go code.
"gro execute" translates the code into Go code then runs it.

`},
}

//================================================================================

var introSpaceless = &tutor.Page{
	Code: []string{`
	包正;种A整;功正(){a:=123;形Println(整64(a));变b这A;变c这A;形Println(b,c)}
	//example of totally spaceless code (except this comment)
`},

	English: []string{`
Spaceless Programming

It's possible to write any valid Gro program without any whitespace, e.g:

`},
}

//================================================================================

var introRule1 = &tutor.Page{
	English: []string{`
Rule 1 - No Unihan in identifier names

The use of Unihan in identifier names is prohibited in the Gro grammar. So "myName" is valid in both Go and Gro, but "my名" and "性名" are invalid in Gro. This is the only way in which Go code is restricted in Gro. Virtually no-one uses Unihan in identifiers anyway -- even Chinese and Japanese programmers really only use them inside strings and comments -- so in practise this shouldn't be a problem for anyone wanting to program in Gro.

`},
}

//================================================================================

var introRule2 = &tutor.Page{
	Code: []string{`
	包正 //包 is short for "package" with an implied space after it
		//正 is short for identifier "main" which is required here
	入"fmt" //入 is short for "import"

	//功 is short for "func", and 正 for identifier "main" which is required here
	功正(){
		fmt.Println("你好,世界")
	}
`,
`
	package main
	import("fmt")
	//we use 整 for "int" and 漂64 for "float64"...
	功plus(a整,b整)漂64{回漂64(a+b)} //回 for "return"
	func main(){
		a:= 复(0, plus(4,6))
		fmt.Println("a is: ", a)
	}
`,
`
	包正;功正(){
		//形 is short for "fmt." which is automatically imported when used
		形Println("你好,世界")
	}
`},

	English: []string{`
Rule 2 - Unihan aliases for keywords and such

Various single Unihan are used as aliases for Go's keywords, special identifiers, and certain package names in the Gro grammar. Each Unihan has an implied space both before and after it so the spaces needn't be written.

The 25 keywords of Go can be substituted by any of their respective Unihan below:

	"包" package, "入" import
	"变" var, "久" const, "种" type, "功" func
	"构" struct, "图" map, "面" interface, "通" chan
	"考" switch, "事" case, "别" default, "掉" fallthrough
	"如" if, "否" else, "为" for, "围" range
	"选" select, "去" go, "终" defer
	"回" return, "破" break, "继" continue, "跳" goto

So we can write Gro code using keyword aliases:

`,

`
The 39 special identifiers in Go can also be substituted by their associated Unihan:

	"真" true, "假" false, "空" nil, "毫" iota
	"双" bool, "节" byte, "字" rune, "串" string, "错" error, "镇" uintptr
	"整" int, "整8" int8, "整16" int16, "整32" int32, "整64" int64
	"绝" uint, "绝8" uint8, "绝16" uint16, "绝32" uint32, "绝64" uint64
	"漂32" float32, "漂64" float64, "复" complex, "复64" complex64, "复128" complex128
	"造" make, "新" new, "关" close, "删" delete, "能" cap, "度" len, "加" append, "副" copy
	"实" real, "虚" imag, "丢" panic, "抓" recover, "写" print, "线" println

The ones suffixed with a number have special support in the grammar, and are the only cases in Gro of Unihan having extra tokens associated with them.

As well as Go aliases "byte" ("节") and "rune" ("字"), Gro adds alias "任" for "interface{}", best verbalized as "any".

`,

`
We also enable Unihan aliases for package names, and they aren't followed by a dot when used. Only 6 packages are implemented for now:

	"形" fmt, "网" net, "序" sort, "数" math, "大" math/big, "时" time

When a package name is aliased by a Unihan, it needn't and mustn't be explicitly imported.

`,

`
More packages will progressively be added over time.

`},
}

//================================================================================

var introRule3 = &tutor.Page{
	Code: []string{`
	//Go by Example: Collection Functions

	package正 //Go code headed by ascii "package"
	入"strings" //Gro code headed by Unihan 入: translated to: import _strings "strings"
	import"fmt" //Go code

	功正(){ //Gro code headed by 功, so all identifiers within have underscore prefixed
		变strs=[]串{"peach","apple","pear","plum"} //"_strs" is actually generated
		形Println(Index(strs,"pear"))
		if 真{ 形Println(Include(strs,"grape")) }
		英{ //Go code embedded within Gro -- signified by 英
			形Println(Any(_strs,功(v串)双{ //referencing Gro-defined identifier from Go code, so prefix _
			回strings.HasPrefix(v,"p")
			}))

			var strs = []string{"peach", "apple", "pear", "plum"} //"strs" generated
			fmt.Println(Index(strs, "pear"))
			fmt.Println(Include(strs, "grape"))
			fmt.Println(Any(strs, func(v string) bool {
				//referencing Go-defined package identifier from Gro -- must prefix underscore...
				return _strings.HasPrefix(v, "p")
			}))
		}
	}

	//We can translate public functions to Gro...
	功Include(vs[]串,t串)双{回Index(vs,t)>=0}

	功Any(vs[]串,f功(串)双)双{
		为_,v:=围vs{如f(v){回真}}
		回假
	}

	//...or we can leave function defn unchanged and everything still just works
	func Index(vs []string, t string) int {
		for i, v := range vs {
			if v == t {
				return i
			}
		}
		return -1
	}
`},

	English: []string{`
Rule 3 - Go and Gro code can be mixed

Code conforming to the Unihan-free Go grammar and that conforming to the Gro grammar can be mixed easily. Gro code is embedded in Go code simply by using the Unihan alias of the keyword or special identifier at the head of the scope, or using special Unihan "做", best verbalized as "do", at the head of a block. Go code is embedded in Gro code by using special Unihan "英", best verbalized as "ascii", at the head of a block.

To understand the details, we must first understand the categories of identifier in Go and in Gro. Just as Go has 3 categories of identifier, i.e. global (the 25 keywords and 39 special identifiers), public (uppercase-initial identifiers), and private (all identifiers beginning with an underscore or lowercase letter, except the 25 keywords), so Gro also has 3 categories. The categories of identifier in Gro match more intuitively to their lexical class, however. They are:

* Unihan. All single-token Unihan which are aliases for keywords, special identifiers, and package names.

* public. Uppercase-initial identifiers are visible outside a package in Gro, just like in Go. They have the same format as in Go.

* protected. Identifiers that begin with an underscore followed by lowercase. They are accessible by both Gro and Go code within a single file. When defined or used within Gro code (i.e. within the static scope of a Unihan), the initial underscore is omitted.

Go's private identifiers are inaccessible within Gro code, which generally isn't a problem because they're usually used as parameters and local variables. If a private variable needs to be accessed by both Go and Gro code, put an underscore in front of it when declaring it in Go context.

The identifier "main" can't have an underscore in front of it in the generated Go code, so we provide the built-in Unihan "正" to use with "包"(package) and "功"(func).

We can see how Go code and Gro code is easily mixed:

`,

`
It is good style, though, to use either Go or Gro but not both as much as possible in a single file. Being able to mix both helps with gradual conversions of code from one syntax to the other.

The rules for Unihan acting as a header for Gro code are:

	"包"package heads everything after it in the package.
	"入"import, "变"var, "久"const, "种"type each head everything they define.
	"功"func heads everything in the receiver, parameters, results, and block after it.
	"构"struct, "图"map, "面"interface, "通"chan each head the types after them. They also head the literal data when used that way, as also do slices and arrays.
	"考"switch, "如"if, "否"else, "为"for, "选"select, "去"go, "终"defer each head everything up to the end of the following block.
	"事"case, "别"default each head the block of statements afterwards.
	"围"range, "回"return each head the expression immediately following it.
	"破"break, "继"continue, "跳"goto each head any labels that follow them.
	36 special identifiers (all, plus "任", except "真"true, "假"false, "空"nil, "毫"iota) head the expression they convert in a type conversion.

`},
}

//================================================================================

var introRule4 = &tutor.Page{
	Code: []string{`
	包正;功正(){
		//added to demo 让:
		让range:="abc" //when used with 让, Go keywords like "range" can be used as identifiers...
		让range="abcdefg" //...and this style should be the prefered style for Gro programmers
		形Printf("range: %v\n",range)
	}
`},

	English: []string{`
Rule 4 - Keywords are identifiers within Gro

Because keywords are identifiers within Gro, all possible lowercase-initial identifiers, including Go keywords, are permitted there.

Gro code permits all possible identifiers beginning with a lowercase letter, including all 25 keywords. It does this by automatically putting an underscore at the front of all lowercase-initial identifiers. Such an identifier referenced by surrounding or embedded Go code must explicitly have the initial underscore. Special identifier "让" is available to optionally head declarations and assignments, and even required when an identifier that doubles as a Go keyword is used.

`},
}

//================================================================================

var introRule5 = &tutor.Page{
	Code: []string{`
	package main
	func main() {
		a:= true
		b:= 真 //Unihan for "true" used on right-hand side
		nil:= true //Gro still allows special identifiers (here, "nil") to be used on the left-hand side, ...
		iota:= 真
		//假:= true // ... but when the Unihan version is used, e.g. 假 for false,
				// generates a parse error "expected non-Unihan special identifier on left hand side"
		形Printf("a: %v, b: %v, nil: %v, iota: %v\n", a, b, nil, iota)

		abc:=图[双]整{}
		abc[真]=789 //of course, Unihan can still be on the LHS when not being assigned to
	}
`},

	English: []string{`
Rule 5 - Special identifier Unicode are protected

Go only has 25 reserved words which can't be used as variables, but all the special identifiers such as "false" can be. Unihan used as special identifiers in Gro can't be locally declared and assigned to as they can in Go:

`},
}

//================================================================================

var introRule6 = &tutor.Page{
	Code: []string{`
	package main

	入"fmt"
	import 吧"fmt" //we can use any Unihan with 口-radical on LHS...
	import 哪_fg"fmt" //...with imports that don't have their own dedicated Unihan
	入㕤hij"fmt"
	入卟"unicode/utf8"
	入吗嗎kl"unicode/utf8" //can even put in two aliases

	type A int
	变n = 50
	功正(){
		var b A
		var c这A //can use "这" with locally-defined type to achieve spaceless program
		形Println(b, c)

		如假{
			fg.Printf("Len: %d\n", 度("hijk") + n)
		}否{
			hij.Printf("Len: %d\n", 度("hi") + n)
		}
		fr,_:= utf8.DecodeRune([]节("lmnop"))
		fmt.Printf("1st rune: %s; Len: %d\n", 串(fr), 度("lmnop") + n)
		让_,_=吗DecodeRune([]节("lmnop"))
		㕤Printf("Fifty: %d\n", n)
		哪Printf("Fifty: %d\n", n)
		吧Printf("Fifty: %d\n", n)
	}
`},

	English: []string{`
Rule 6 - Packages can be aliased with Unihan

A package not in the registry of Unihan-aliased packages can be given a temporary Unihan when imported. The Unihan is any of those with the "口" radical on the left hand side. There's about 2000 such Unihan to choose from out of the 80,000 in Unicode. Types defined in the current package can be prefixed with "这", best verbalized as "this", if desired.

`},
}

//================================================================================
var introImpliedPack = &tutor.Page{
	Code: []string{`
	形Println("abc", 789)
`},

	English: []string{`
Implied packages and main functions

Because the "package main" and "func main()" lines are optional, we can write single-line scripts in Gro:

`},
}

//================================================================================

var introMultipack = &tutor.Page{
	Code: []string{`
	包正 //package "main" contains two explicit sources
	源"tiful.go" // 1st source
	种A整;
	功正(){
		a:=123
		形Println(整64(a))
	}
	源"beau.go" // 2nd source
	种B整;
	功hello(){
		b:=789
		形Println(整64(b))
	}
	包hi //package "hi" contains one implied source, called "hi.go"
	功hoorah(){
		c:=666
		形Println(整32(c))
	}
	包thinking //one implied source, "thinking.go", and one explicit source, "psyche.go"
	功mind(){
		e:= -12345
		形Println(整8(e))
	}
	源"psyche.go"
	功psycheGo(){
		g:= 54321
		形Println(整8(g))
	}
	包bye //one explicit source, "booyah.go"
	源"booyah.go"
	功cherio(){
		d:= -99
		形Println(整16(d))
	}
`,
`
	功正(){
		形Println('a')
	}
`,
`
	形Println("Hey, world!") //standalone stmt for package "main", "main.go" in current directory
	包正 //package "main" uses current directory
	源"tiful.go" //source "tiful.go"
	种A整;
	功正(){
		a:=123
		形Println(整64(a))
	}
	源"beau.go" //source "beau.go"
	种B整;
	功hello(){
		b:=789
		形Println(整64(b))
	}
	包"oneLib/hi" //one implied src-name with specified string for pkg:hi, dir:oneLib/hi, src:hi.go
	功hoorah(){
		c:=666
		形Println(整32(c))
	}
	包"someDir"thinking //one implied and one explicit src-name with specified pkg-name and string for pkg:thinking, dir:someDir, srcs:thinking.go & psyche.go
	功mind(){
		e:= -12345
		形Println(整8(e))
	}
	源"psyche.go"
	功psycheGo(){
		g:= 54321
		形Println(整8(g))
	}
	包"lets/wave"bye //one explicit src-name with specified pkg-name and string for pkg:bye, dir:lets/wave, src:booyah.go
	源"booyah.go"
	功cherio(){
		d:= -99
		形Println(整16(d))
	}
`},

	English: []string{`
Multisource packages and multipackaging

The code from a single package in a Gro source file can be distributed among many Go source files using the "源" command, best verbalized as "source". As well as this, a single Gro source file can contain the code for many packages.

If the "源"(source) command is missing, the source file will have the same name as the package, with ".go" appended.

`,

`
If the first package command is missing, it will be called "main". If the "源"(source) command is also missing, the generated source file will be 'main.go'.

`,

`
Code beginning with a statement or a label can appear outside a function in Gro. Such code is automatically wrapped inside "func main()". Any "type", "var", and "const" declarations mixed in with such statements will still be interpreted as having package scope, and will be plucked out. To include such declarations with the statements for wrapping within "func main()", use Go's "_:" label.

It's a common idiom to write a quick script consisting only of statements, which will be wrapped inside "func main()", given package name "main", and put in file "main.go". Because the "gro execute" command runs the "main.go" file after preparing the given ".gro" file, such statement-only code can be run with "gro execute". Such statements can act as build code for packages following it in the same ".gro" file, though because of Go's simple builds, that's usually not necessary.

A directory can be specified in the package command, either before or instead of the package name.

`},
}

//================================================================================

var introExample = &tutor.Page{
	Code: []string{`
	包正 //包 is shorthand Unihan for "package"; 正 is short for "main"
	久s串="constant" //久 short for "const"
	功正(){ //功 short for "func"

		//Go by Example: Hello world
		//形 is short for "fmt." which is automatically imported when used
		形Println("你好,世界")

		//Go by Example: Values
		形Println("go"+"lang") //Canonical "spaceless" Gro style: no spaces within statements and expressions
		形Println("1+1 =",1+1)
		形Println("7.0/3.0 =",7.0/3.0)
		形Println(真&&假) //真 for "true"; 假 for "false"
		形Println(真||假)
		形Println(!真)

		//Go by Example: Variables
		做{ //optionally, begin standalone block with "做"
			变a串= "initial";形Println(a) //变 for "var"; 串 for "string"
			变b,c整=1,2;形Println(b, c) //整 for "int"
			变d=真;形Println(d)
			变e整;形Println(e)
			f:="short";形Println(f)
			//canonical Gro style: join stmts on one line with semicolons if they're written
			//as separate lines between blank lines in corresponding Go style
		}
	}
	功plus(a整,b整)整{回a+b} //回 for "return"
	功plusPlus(a,b,c整)整{回a+b+c}
`},

	English: []string{`
Examples

Most of the Go examples from gobyexample.com (or its Chinese translation gobyexample.everyx.in ) are available as Gro source in github.com/grolang/samples , all except the last few. (Go By Example is copyrighted by Mark McGranaghan.)

For example, the code from the first 3 pages of "Go by Example" can be replaced by the terser:

`,

`
If we shorten the local identifier names, and use semicolon (";") to join lines together, we can achieve much greater tersity.

`},
}

//================================================================================

