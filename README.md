![Gro Logo](gro.jpg)

Gro is both a syntax that extends that of Go's, and a tool that generates Go code from Gro's.

## License

Copyright © 2016 The Gro authors

Distributed under the same BSD-style license as Go that can be found in the LICENSE file.

## Status

Version 0.5.

All documented functionality is implemented.

## Operation

### Installation

First, make sure Go is already installed. Gro has tested successfully on Go 1.7.

Run `go get github.com/grolang/gro` to get the command and packages.

Run `go install github.com/grolang/gro/cmd/gro` to compile and install the `gro` command from the downloaded source.

Run `go get github.com/grolang/samples` to get some sample gro code, including most of the examples from the GoByExample website.

### Documentation

Run `gro help` to see a list of commands available. All commands are also available via the browser as doc.go files in the source hierarchy.

Run `gro learn` to see a list of tutorials available. You can learn about the Unihan characters and Gro syntax rules. All tutorials are also available via the browser as doc.go files.

### Execution

Run `gro prepare src/github.com/grolang/samples/goByEg1.gro` to format one of the supplied gro code samples, which can then be run using the standard `go run src/github.com/grolang/samples/goByEg1.go`. Most of the examples from GoByExample have been translated into Gro.

Run `gro execute src/github.com/grolang/samples/goByEg1.gro` to both format one of the supplied gro code samples, and run it..

Run `gro repl` to run the rudimentary REPL.

## Rationale

The Gro syntax is an extension of Go's, parsed by a modified edition of the recursive descent parser shipped in Go 1.6. The `parser`, `scanner`, and `cmd/gofmt` packages were cloned and modified, and the `golang.org/x/tools/go/ast/astutil` package copied.

Cloning parts of this code (while adhering to the license restrictions) to create other scripting languages for Go is encouraged. The primary focus for my own work on this fork of the codebase (i.e. Gro) remains that of showcasing Unihan, and later other Unicode characters, in the language grammar.

