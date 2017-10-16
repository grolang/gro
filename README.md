# Gro Language

Gro is both a syntax that extends Go's, and a tool that generates Go code from Gro's.

Various features:

* when a gro-file is valid Go syntax, its functionality is exactly the same as the equivalent go-file
* various syntactic shortcuts enable one-line programs to be written in Gro
* packages can be defined with parameters and imported with arguments, thus supporting basic Generics
* Go's 6 top-level keywords are supplemented with 8 more top-level Gro keywords to give extra functionality
* Go and Gro code can be mixed freely in the same source file


## License

Copyright © 2017 The Gro and Go authors

Distributed under the same BSD-style license as Go that can be found in the LICENSE file.


## Status

Version 0.7, released 16 October 2017.

All functionality documented below is implemented in 0.7.


## Operation

### Installation

First, make sure Go is already installed. Gro has tested successfully on Go 1.9.1.

Run `go get github.com/grolang/gro` to get the command and packages.

The hierarchy should then be:

```
	GOPATH
	  +--bin
	  +--pkg
	  +--src
	      +--github.com
	      |   +--grolang
	      |      +--gro
	      |      |   +--[various libraries of go source code for gro]
	      |      |   +--LICENSE.txt, etc
	      +--[put your own directories and files here]
```

Run `go install github.com/grolang/gro/cmd/gro` to compile and install the `gro` command from the downloaded source.

You can add your own projects anywhere else under the src/ directory.


### Execution

gro runs various utilities to complement those in the Go command.

Run `gro prepare src/github.com/grolang/gro/cmd/gro/testdata/list.gro` to format one of the supplied gro code samples, which can then be run using the standard `go run src/github.com/grolang/gro/cmd/gro/testdata/list.go`.

Or, run `gro execute src/github.com/grolang/gro/cmd/gro/testdata/list.gro` to both format and run that gro code sample.


### Documentation

Run `gro help` to see a list of commands available.


## Features of Gro

### One-line programs

Various syntactic shortcuts enable one-line programs to be written in Gro. Here is a one-liner:

```
"fmt".Println("Hello, world!")
```

A path-string with a dot following it, anywhere in the syntax, is automatically converted to an import. Also, when statements exist at the top level outside of an enclosing function, the "package main" and "func main()" will be inferred.


### Compatibility with Go

When a gro-file is valid go syntax, its functionality is exactly the same as the equivalent go-file. So executing this gro-file in Gro does exactly the same as running it in Go:

```
package main
import "fmt"
func main() {
	fmt.Println("Hello, world!")
}
```

Go and Gro code can be mixed freely in the same source file.


### Generics

Packages can be defined with parameters and imported with arguments, enabling generic classes.

To see this work:

* copy `container/list/list.go` from the Go standard library into your own directory and change its suffix to `gro`, e.g. `mydir/list.gro`
* change the 7 occurences of `interface{}` to `T`
* extend the package clause so it reads `package list (T)`
* create a file `mydir/runlist.gro` with these statements (borrowed from one of the tests in `container/list/list_test.go`):

```
include "mydir/list.gro"

import (
	listint "list" (int) //local name compulsory when importing a generic package with arguments
	"fmt"
)

do {
	// Create a new list and put some numbers in it.
	l := listint.New()
	e4 := l.PushBack(4)
	e1 := l.PushFront(1)
	l.InsertBefore(3, e4)
	l.InsertAfter(2, e1)

	// Iterate through list and print its contents.
	for e := l.Front(); e != nil; e = e.Next() {
		fmt.Println(e.Value)
	}

	// Output:
	// 1
	// 2
	// 3
	// 4
}
```

* run `gro execute mydir/runlist.gro` to see that code run


### Extra top-level keywords

Go has 6 top-level keywords, i.e. "package", "import", "var", "const", "type" and "func". These are supplemented in Gro with 8 more keywords, available at the top level:

* "include" allows gro-files to load and run other gro-files
* "internal" operates at the same level as "package"
* "section" enables many project go-files to be included in one gro-file
* "main" operates at the same syntactic level as "section"
* "testcode" operates at the same syntactic level as "section"
* "project" is available, which overrides the name of the gro-file
* "do" enables blocks and various statements to be available at top-level
* "proc" enables top-level style syntax within functions

The "package" keyword is optional, and inferred from the name of the gro-file. More than one package is allowed in a single gro-file.

Block notation is available for defining packages and sections.

Go's doc-comments, project-wide comments, and section-wide comments are supported in a gro-file.

There is a small set of syntax shortcuts:

* the "else" keyword can be followed by other statement keywords as well as "if"
* bare "if", "for", "switch", "select", "go" and block statements can be called from top-level
* there's a special shortcut syntax for the "defer" and "go" keywords


## Behaviour

Files with extension `.gro` are parsed and those with extension `.go` are usually generated.


### Subsets of Go programs

When a file `abc.gro` with valid go syntax is parsed, the output is `abc.go` in the same directory with nothing changed except it having been run through gofmt. For now, the doc-comments are generated but other comments will be dropped.

Gro programs can also be a subset of Go. The Gro program:
```
"fmt".Println("Hello, world!")
```
is a subset of Go program:
```
package main
import "fmt"
func main() {
	fmt.Println("Hello, world!")
}
```
and they do exactly the same thing when executed. The package keyword and name `package main`, the `import` keyword and imported package name `fmt.`, and the `func main(){}` function wrapping are all dropped in the Gro version.

When there's a dot-path on a string, it's converted to an alias and an import added:
```
package main
func main() {
	"fmt".Println("Hello, world!")
}
```

So instead of importing a package at the top of a file, it can be referred to by a string where it's used. It will assume the last directory name in the specified sequence is the package name, which is true for all standard library packages.
```
package main
func main(){
	"fmt".Println("path/filepath".Join("Hello, ", "world!"))
}
```
will generate:
```
package main
import (
	"fmt"
	"path/filepath"
)
func main(){
	fmt.Println(filepath.Join("Hello, ", "world!"))
}
```

These can be included anywhere in a Gro file, even deeply nested within Go syntax.

### Implicit packages

When the package name is absent, it will be `main` if there's a function called `main` with no parameters and no return values defined:
```
import "fmt"
func main() {
	fmt.Println("Hello, world!")
}
```
or even if there's statements at the top-level in the .gro file:
```
import "fmt"
fmt.Println("Hello, world!")
```
If the file containing the main() function is called `dud.gro`, the generated file will be called `dud.go`:
```
// +build ignore
package main
import "fmt"
func main() {}
func init() {
	fmt.Println("abc")
}
```

If the package keyword is missing but no `main` function or standalone statements, the package name will be inferred from the name of the gro file, e.g. file `mypack.gro` containing lines which would be valid go syntax if there had been a package keyword:
```
import "fmt"
func run() {
	fmt.Println("abc")
}
```
will produce file `mypack.go` with lines:
```
package mypack
import "fmt"
func run() {
	fmt.Println("abc")
}
```

### Multi-sections

Gro also provides a large superset of Go syntax, such as the `section` keyword.

When a file `abc.gro` has a section keyword with accompanying string straight after the package keyword and all other lines are valid go syntax, e.g.
```
package somepkg
section "xyz"
//...
```
then a file with the same name as the string but with ".go" appended, here `xyz.go`, in the same directory is produced:
```
package somepkg
//...
```

When `abc.gro` has 2 or more section keywords, the first straight after the package keyword, e.g.
```
package somepkg

section "uvw"
//code u

section "xyz"
//code x
```
then a file for each section keyword is produced:
"uvw.go" will have:
```
package somepkg
//...
```
and "xyz.go" will have:
```
package somepkg
//...
```

If there's one or more section keywords, but none are directly after the package keyword, and all other lines are valid go, e.g. in `dud.go`:
```
package somepkg

//code t

section "xyz"
//code x
```
then an extra file is produced:
"dud.go" will have:
```
package somepkg
//code t
```
and "xyz.go" will have:
```
package somepkg
//code x
```

The behaviour when there's section keyword/s but no package keyword is a hybrid of what happens in each case.

All doc comments in the .gro file will be included in the generated .go files. The doc-comment directly above the package keyword will only be included in the first file generated. The comment above the section keyword will be included in that specific file after a blank line below the package keyword.

### "main" and "testcode" keywords

The keyword `main` can appear whereever the section keyword can. It causes the produced file to have package main in the same directory as the other files using the primary package name.
```
package somepkg

section "uvw"
//code u

main "xyz"
//code x
```
produces files "uvw.go":
```
package somepkg
//code u
```
and "xyz.go":
```
// +build ignore
package main
//code x
```

Similarly, the keyword `testcode` causes the produced file to have "_test" appended to the given filename. If the package is explicitly imported, "_test" will be appended to the package name also:
```
package somepkg
section "thecode"
//code u
testcode "super"
//code v
testcode "sub"
import . "somepkg"
//code w
```
produces "thecode.go":
```
package somepkg
```
and "super_test.go":
```
package somepkg
```
and "sub_test.go":
```
package somepkg_test
```

### Multi-packages

When there's two or more package keywords, the resulting files will be split between two different directories, e.g. in file `mypacks.gro`:
```
package somepkg
//code s
package secondpkg
//code t
```
will generate "somepkg/somepkg.go":
```
package somepkg
//code s
```
and "secondpkg/secondpkg.go":
```
package secondpkg
//code t
```
So defining a second package in a gro-file changes the location of the package already defined.

When there's code before the first package keyword, then that code will be wrapped within an implied package within the current directory based on the name of the .gro file. E.g. file "mypacks.gro" containing
```
//code r
package somepkg
//code s
package secondpkg
//code t
```
will produce "mypacks.go":
```
package mypacks
//code r
```
and "somepkg/somepkg.go":
```
package somepkg
//code s
```
and "secondpkg/secondpkg.go":
```
package secondpkg
//code t
```

When there's section keywords mixed in with multi-packages, the resulting behaviour is a hybrid of both the multi-package and multi-file cases. E.g. in file `mypacks.gro`:
```
//code r

section "uvw"
//code s

package somepkg

//code t

section "xyz"
//code x
```
then four files are produced:
"mypacks.go":
```
package mypacks
//code r
```
"uvw.go":
```
package mypacks
//code s
```
"somepkg/somepkg.go":
```
package somepkg
//code t
```
"somepkg/xyz.go":
```
package somepkg
//code x
```

So when defining a second package in a gro-file, to keep the same location for the first package, we can remove its package keyword. If the gro-file is the same name as the first package, none of the files produced will change its name. Otherwise, we can change the `package firstpkg` to `section "firstpkg"` to force all produced files to keep their existing names.

### Projects

A file `something.gro` can optionally have a single project keyword at the very top, e.g.
```
project myproj
package somepkg
//...
```
Here, file `myproj.go` is generated. The project name overrides the base name of the gro file.

The comments above the project keyword will be included with a blank line afterwards at the top of all generated .go files, so it's useful to put a copyright notice there.

### "internal" keyword

The internal keyword can appear whereever the package keyword can, to signify the special-use package called internal:
```
package somepkg
//code s
internal
//code t
internal packone
//code u
internal "somedir" mypack
//code v
section "more"
//code w
```
will generate files:
"somepkg/somepkg.go":
```
package somepkg
//code s
```
"internal/internal.go":
```
package internal
//code t
```
"internal/packone/packone.go":
```
package packone
//code u
```
"internal/somedir/mypack/mypack.go":
```
package mypack
//code v
```
"internal/somedir/mypack/more.go":
```
package mypack
//code w
```

### Directories

A string can be specified in the package command before the package name. The string specifies a root directory under which all files in that package will be placed. E.g. a file called `something.gro`:
```
package "some/path" pkgone
//code x (with no section-specs)
package "some/path" mypkg
//code y (with no section-specs)
```
will generate "some/path/pkgone/pkgone.go":
```
//code x
```
and "some/path/mypkg/mypkg.go":
```
//code y
```

The project keyword can also have a string argument. As with packages, it specifies a root directory under which all code appearing after it will be placed. E.g. a file containing
```
project "my/path" myproj
//code r
package somepkg
//code s
```
will produce "my/path/myproj.go":
```
package myproj
//code r
```
and "my/path/somepkg/somepkg.go":
```
package somepkg
//code s
```

When both the project and package commands have a string argument, their individual effects are combined. So a file:
```
project "my/path" myproj
package "another/path"
//code x
```
will generate "my/path/another/path/myproj.go":
```
package myproj
//code x
```

### Generics

Entire packages can have one or more type parameters:
```
package chanpack (T1)
type MyChan chan T1

package mappack (T2, T3)
type MyMap map[T2]T3
```

The type arguments for the package are supplied in the import definition of the package where they are instantiated, which can optionally be in the same gro file. Each imported package must have an alias.
```
package "somepath" mypack (T)
type List struct { element T; next *List }
//...

package mappack (T2, T3)
type MyMap map[T2]T3

package "caller" callingpackage
import intpack   "somepath/mypack" (int)
import floatpack "somepath/mypack" (float)
import (
	mymappack "somepath/mappack" (int, string)
	clubpack  "somepath/mypack" ("somepath/auxpkg".Club) //circular import refs not allowed
)

type ListInt intpack.List
var v1 intpack.List
var v2 floatpack.List
var v3 mymappack.MyMap

func main () {
	cp:= clubpack.List
	cp.element
}
package "somepath" auxpkg
type Club //defined here...
```
generates "caller/callingpackage.go":
```
package callingpackage
import "generics/caller/callingpackage/intpack"
import "generics/caller/callingpackage/floatpack"
import (
	"generics/caller/callingpackage/mymappack"
	"generics/caller/callingpackage/clubpack"
)

type ListInt intpack.List
var v1 intpack.List
var v2 floatpack.List
var v3 mymappack.MyMap

func main () {
	cp:= clubpack.List
	cp.element
}
```
"somepath/auxpkg.go":
```
package auxpkg
type Club //defined here...
```
"generics/caller/callingpackage/intpack/intpack.go":
```
package intpack
type T = int
type List struct { element T; next *List }
//...
```
"generics/caller/callingpackage/floatpack/floatpack.go":
```
package floatpack
type T = float
type List struct { element T; next *List }
```
"generics/caller/callingpackage/clubpack/clubpack.go":
```
package clubpack
import "somepath/auxpkg"
type T = auxpkg.Club
type List struct { element T; next *List }
```
"generics/caller/callingpackage/mymappack.go":
```
package mappack
type (
	T1 = int
	T2 = string
)
type MyMap map[T1]T2
```

### Block-style package and section code

package and section keywords can be written in block-style between curlies instead of form-style:

```
project hello
package abc {
	import "fmt"
	func run() {
		fmt.Println("Hello, world!")
	}
	section "this" {
		import "fmt"
		func run() {
			fmt.Println("Goodbye, world!")
		}
	}
}
```
will generate "dud.go":
```
package abc
import "fmt"
func run() {
	fmt.Println("Hello, world!")
```
and "this.go":
```
package abc
import "fmt"
func run() {
	fmt.Println("Goodbye, world!")
}
```

Any mixture of block-style and form-style is valid.

The other keywords similar to package (i.e. internal) and section (i.e. main and testcode) also work this way.

### "include" directive

The "include" keyword will read and parse the included files being parsing the current gro-file. It's put after the "use" keyword:
```
project carrot
use "groot"
include "other.gro"
//...
```
or, if the optional "project" and "use" keywords aren't being used, at the top:
```
include (
	"other.gro"
	"another.gro"
)
//rest of code
```


### Implicit init functions with "do" statements

Statements that don't begin with keywords, such as assignments and post-crements, can be included in the top-level by using a do-statement. They will be automatically wrapped inside "func init()". The "type", "var", "const", and "func" declarations will still be included in the top-level declarations:
```
package abc
do a:= 7
do "fmt".Println(a)
var z string
do b:= 8
"fmt".Println(b) //"do" optional here
```
will generate:
```
package abc
import "fmt"
func init() {
	a := 7
	fmt.Println(a)
}
var z string
func init() {
	b := 8
	fmt.Println(b)
}
```

Code without a package directive or main() function, but with such standalone statements, will, as we've seen, have "package main" and "func main(){}" added to the generated go file.

"do" can also precede a bare block at the top level scope to indicate the contents can be any valid Go statement:
```
do {
	a:= 7
	b:= 8
	"fmt".Println(a+b)
}
```
will generate:
```
package main
import "fmt"
func main() {
	{
		a := 7
		b := 8
		fmt.Println(a+b)
	}
}
```

"do" can also precede a "if", "for", "switch", "select" or "return" statement at the top level:
```
do for {
	a:= 17
	b:= 18
}
```
will generate:
```
package main
import "fmt"
func main() {
	for {
		a := 17
		b := 18
	}
}
```

Gro files can also contain bare "if", "for", "switch", "select", and "return" statements and standalone blocks mixed in with the "do" statements and "type", "var", "const", and "func" declarations at the top level. Their blocks (except for "return") contain statements in the same syntactic style as those at the top level:
```
for {
	do a:= 7
	do b:= 8
}
do for {
	a:= 17
}
```
will generate:
```
package main
import "fmt"
func main() {
	for {
		a := 7
		b := 8
	}
	for {
		a := 17
	}
}
```

Top-level standalone block:
```
{
	do a:= 17
	do b:= 18
	do "fmt".Println(a*b)
}
do {
	a:= 7
}
```
will generate:
```
package main
import "fmt"
func main() {
	{
		a := 17
		b := 18
		fmt.Println(a*b)
	}
	{
		a := 7
	}
}
```

"const", "var" and "type" declarations can be put inside the block for such a "if", "for", "switch", or "select" statement, "proc" definition, or standalone block, but they are interpreted as local declarations, not top-level. Labelled statements, "fallthrough", "continue", "break", and "goto" statements, and "func" declarations can't be put there.

### "proc" keyword

When the "proc" keyword is used instead of the "func" keyword when defining a function at the top-level, the statements must have the same syntax as top-level statements:
```
import "fmt"
proc fa(){
	do a:= 7
	do fmt.Println(a)
}
func fb(){
	b:= 8
	fmt.Println(b)
}
```
will generate:
```
package main
import "fmt"
func fa () {
	a := 7
	fmt.Println(a)
}
func fb () {
	b := 8
	fmt.Println(b)
}
```

### "go" and "defer" keyword syntax

The "go" keyword can also be used at the top level with "do":
```
do go func(){
	a:= 7
}()
```

However, this isn't allowed:
```
//Syntax Error:
go proc(){
	do a:= 7
}()
```
because the extra Gro keywords like proc can only be used when beginning a directive at the top-level.

But when the function has no parameters or return value, as is the usual case, a shortcut can be used:
```
do go {
	a:= 7
}
go {
	do a:= 7
}
```

The "defer" keyword also has this shortcut syntax:
```
//Ok:
do defer {
	a:= 7
}
defer {
	do a:= 7
}
```

### Expanded "else" keyword

The "else" keyword in an if statement be followed by a block or another if-statement in Go. In Gro, it can also be followed by a switch-statement, for-statement, or select-statement:
```
if true {
	fmt.Println("true")
} else switch {
default:
	fmt.Println("false")
}
```

## Rationale

Gro aims ultimately to supplement Go's functionality similar to how the original Groovy supplemented Java, but in a way that makes sense for Go. The Gro tool is based on Go's `cmd/compile/internal/syntax` package.


## Future Direction

The next steps in building Gro will be to:

* ensure all comments in gro source code are in the generated go code
* add macros, allowing Go developers to add their own top-level keywords to Gro

The core of Gro is restricted to a small handful of features, in the same spirit as the minimal style of Go. In fact, version 0.7 removed many core features from version 0.6, intending to re-introduce them as optional macros later on.

