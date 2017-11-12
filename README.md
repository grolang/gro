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

Visit the [wiki](https://github.com/grolang/gro/wiki/Home) for:

* [Basic Features](https://github.com/grolang/gro/wiki/Features)

* [1. Gro's syntax](https://github.com/grolang/gro/wiki/Syntax)

* [2. One-liners](https://github.com/grolang/gro/wiki/Oneliners)

* [3. Generics in Gro](https://github.com/grolang/gro/wiki/Generics)

* [4. Gro Macros](https://github.com/grolang/gro/wiki/Macros)

* [5. Dynamic Typing](https://github.com/grolang/gro/wiki/Dynamic)


## Operation

### Installation

First, make sure Go is already installed. Gro has tested successfully on Go 1.9.2.

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

Version 0.7.3, released on 12 November 2017.

All functionality documented in the wiki is implemented.


## License

Copyright © 2017 The Gro and Go authors

Distributed under the same BSD-style license as Go that can be found in the LICENSE file.


## Rationale

Gro aims ultimately to supplement Go's functionality similar to how the original Groovy supplemented Java, but in a way that makes sense for Go. The Gro tool is based on Go's `cmd/compile/internal/syntax` package.


## Future Direction

The next steps in building Gro will be to:

* ensure all comments in gro source code are in the generated go code
* allow Go developers to add their own macros to Gro

The core of Gro is restricted to a small handful of features, in the same minimal style as Go. In fact, version 0.7 removed many core features from version 0.6, intending to re-introduce them as optional macros later on.

