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

Run `gro help` to see a list of commands available. Visit the wiki for [a spec of Gro's syntax](https://github.com/grolang/gro/wiki/Spec).


## Features of Gro

### One-line programs

Various syntactic shortcuts enable one-line programs to be written in Gro. Here is a one-liner:

```go
"fmt".Println("Hello, world!")
```

A path-string with a dot following it, anywhere in the syntax, is automatically converted to an import. Also, when statements exist at the top level outside of an enclosing function, the `package main` and `func main()` will be inferred.


### Compatibility with Go

When a gro-file is valid go syntax, its functionality is exactly the same as the equivalent go-file. So executing this gro-file in Gro does exactly the same as running it in Go:

```go
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

```go
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

Go has 6 top-level keywords, i.e. `package`, `import`, `var`, `const`, `type` and `func`. These are supplemented in Gro with 8 more keywords, available at the top level:

* `include` allows gro-files to load and run other gro-files
* `internal` operates at the same level as `package`
* `section` enables many go-files for one project to be included in one gro-file
* `main` operates at the same syntactic level as `section`
* `testcode` operates at the same syntactic level as `section`
* `project` is available to use at the top of the gro-file, and overrides the name of the gro-file
* `do` enables blocks and various statements to be available at the top-level
* `proc` enables top-level style syntax to be used within functions

The `package` keyword is optional, and inferred from the name of the gro-file. More than one package is allowed in a single gro-file, each with its own `package` keyword.

Block notation is available for defining packages and sections.

Go's doc-comments, project-wide comments, and section-wide comments are supported in a gro-file.

There is a small set of syntax shortcuts:

* the `else` keyword can be followed by other statement keywords as well as `if`, such as `switch` and `for`
* bare `if`, `for`, `switch`, `select`, `go`, `defer` and block statements can be called from the top-level
* the `defer` and `go` keywords have a special shortcut syntax when calling functions with no parameters or return values


## Rationale

Gro aims ultimately to supplement Go's functionality similar to how the original Groovy supplemented Java, but in a way that makes sense for Go. The Gro tool is based on Go's `cmd/compile/internal/syntax` package.


## Future Direction

The next steps in building Gro will be to:

* ensure all comments in gro source code are in the generated go code
* add macros, allowing Go developers to add their own top-level keywords to Gro

The core of Gro is restricted to a small handful of features, in the same spirit as the minimal style of Go. In fact, version 0.7 removed many core features from version 0.6, intending to re-introduce them as optional macros later on.

