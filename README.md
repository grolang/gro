![Gro Logo](gro.jpg)

Gro is both a syntax that extends that of Go, and a tool that generates Go code from Gro.


## License

Copyright © 2017 The Gro and Go authors

Distributed under the same BSD-style license as Go that can be found in the LICENSE file.


## Status

Version 0.6.

All documented functionality is implemented.


## Operation

### Installation

First, make sure Go is already installed. Gro has tested successfully on Go 1.7.

Run `go get github.com/grolang/gro` to get the command and packages.

The hierarchy should be:

```
	GOPATH
	  +--bin
	  +--pkg
	  +--src
	  |   +--github.com
	  |   |   +--grolang
	  |   |      +--gro
	  |   |      |   +--[various libraries of go source code for gro]
	  |   |      |   +--LICENSE.txt, etc
	  |   |      +--groo (i.e. sample dynamic language addon for gro)
	  |   |      +--samples
	  |   +--[put your own directories and files here]
	  +--tmp (i.e. temporary files used by gro)
```

Run `go install github.com/grolang/gro/cmd/gro` to compile and install the `gro` command from the downloaded source.

Run `go get github.com/grolang/samples/gro` to get some sample gro code, including most of the examples from the GoByExample website.

You can add your own projects anywhere else under the src/ directory.


### Execution

gro runs various utilities to complement those in the Go command.

Run `gro prepare src/github.com/grolang/samples/gro/goByEg1.gro` to format one of the supplied gro code samples, which can then be run using the standard `go run src/github.com/grolang/samples/gro/goByEg1.go`.

Run `gro execute src/github.com/grolang/samples/gro/goByEg1.gro` to both format one of the supplied gro code samples, and run it.

Most of the examples from GoByExample have been translated into Gro.


### Documentation

Run `gro help` to see a list of commands available. All commands are also viewable via the browser as doc.go files in the source hierarchy.

Run `gro learn` to see a list of tutorials available. You can learn about the Gro tool, syntax rules, and Unihan characters. All tutorials are also viewable via the browser as doc.go files.


### Rudimentary

Run `gro repl` to run the rudimentary REPL.

```
do a:=7
do b:=8
fmt.Println(a+b,a*b)
exit
```


## Syntax

The syntax for Grolang is described in detail in [a separate document](SPEC.md).


## Rationale

The Gro tools comprise a modified edition of the `go/*` packages from the Go standard library.

