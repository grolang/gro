# Gro Language

Gro is both a syntax that extends Go's, and a tool that generates Go code from Gro's.

Various features:

* when a gro-file is valid Go syntax, its functionality is exactly the same as the equivalent go-file
* various syntactic shortcuts enable one-line programs to be written in Gro
* packages can be defined with parameters and imported with arguments, thus supporting basic Generics
* Go's 6 top-level keywords are supplemented with more top-level Gro keywords to give extra functionality
* various macros are available, including dynamic typing
* Go and Gro code can be mixed freely in the same source file


## Wiki contents

If you know Go but nothing about Gro, see the [Basic Features](https://github.com/grolang/gro/wiki/Features).


After that, you can read these in any sequence:

* [Gro's syntax](https://github.com/grolang/gro/wiki/Syntax)

* [One-liners in Gro](https://github.com/grolang/gro/wiki/Oneliners)

* [Generics in Gro](https://github.com/grolang/gro/wiki/Generics)

* [Language Spec](https://github.com/grolang/gro/wiki/Spec)

* [Macros in Gro](https://github.com/grolang/gro/wiki/Macros)

All the syntax described is implemented in Gro 0.8.


Work in progress:

* [Dynamic Typing with Groo](https://github.com/grolang/gro/wiki/Dynamic)

* [Low-level version of Go](https://github.com/grolang/gro/wiki/Lowlevel)


## Operation

### Installation

First, make sure Go is already installed. Gro has tested successfully on Go 1.9.3.

Run `go get github.com/grolang/gro` to get the command and support packages for Gro, and `go get github.com/grolang/samples` to get some Gro sample code.

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
          |      +--samples
          |          +--[various samples of gro source code]
          +--[put your own directories and files here]
```

Run `go install github.com/grolang/gro/cmd/gro` to compile and install the `gro` command from the downloaded source.

You can add your own projects anywhere else under the `src/` directory.


### Execution

Gro runs various utilities to complement those in the Go command.

From your `GOPATH` as working directory, run `gro prepare src/github.com/grolang/samples/container/list_run.gro` to format one of the supplied gro code samples, which can then be run using the standard `go run src/github.com/grolang/samples/container/list_run.go`.

Or, run `gro execute src/github.com/grolang/samples/container/list_run.gro` to both format and run that gro code sample.


### Documentation

Run `gro help` to see a list of commands available, or visit the [wiki](https://github.com/grolang/gro/wiki/Home) for help on Gro's language features.


## Status

Version 0.8, released on 15 January 2018.

All functionality documented in the wiki is implemented, unless otherwise stated.


## License

Copyright © 2018 The Gro and Go authors

Distributed under the same BSD-style license as Go that can be found in the LICENSE file.


## Rationale

Gro aims ultimately to supplement Go's functionality similar to how the original Groovy supplemented Java, but in a way that makes sense for Go. The Gro tool is based on Go's `cmd/compile/internal/syntax` package.


## Future Direction

The core of Gro is restricted to a small handful of features, in the same minimal style as Go. In fact, version 0.7 removed many core features from version 0.6, intending to re-introduce them as optional macros later on.

