// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"fmt"
	"io"
	"os"
	"github.com/grolang/gro/syntax/src"
)

//--------------------------------------------------------------------------------
// Mode describes the parser mode.
type Mode uint

// Modes supported by the parser.
const (
	CheckBranches Mode = 1 << iota // check correct use of labels, break, continue, and goto statements
)

// Error describes a syntax error. Error implements the error interface.
type Error struct {
	Pos src.Pos
	Msg string
}

func (err Error) Error() string {
	return fmt.Sprintf("%s: %s", err.Pos, err.Msg)
}

var _ error = Error{} // verify that Error implements error

// An ErrorHandler is called for each error encountered reading a .go file.
type ErrorHandler func(err error)

// A Pragma value is a set of flags that augment a function or
// type declaration. Callers may assign meaning to the flags as
// appropriate.
type Pragma uint16

// A PragmaHandler is used to process //line and //go: directives as
// they're scanned. The returned Pragma value will be unioned into the
// next FuncDecl node.
type PragmaHandler func(pos src.Pos, text string) Pragma

//--------------------------------------------------------------------------------

// Parse parses a single Go source file from src and returns the corresponding
// syntax tree. If there are errors, Parse will return the first error found,
// and a possibly partially constructed syntax tree, or nil if no correct package
// clause was found. The base argument is only used for position information.
//
// If errh != nil, it is called with each error encountered, and Parse will
// process as much source as possible. If errh is nil, Parse will terminate
// immediately upon encountering an error.
//
// If a PragmaHandler is provided, it is called with each pragma encountered.
//
// The Mode argument is currently ignored.
func Parse(filename string, base *src.PosBase, src io.Reader, errh ErrorHandler, pragh PragmaHandler, mode Mode, f func(string)(string,error)) (
		_ map[string]*File, first error) {
	defer func() {
		if p := recover(); p != nil {
			if err, ok := p.(Error); ok {
				first = err
				return
			}
			panic(p)
		}
	}()

	var p parser
	p.init(filename, base, src, errh, pragh, mode, f)
	p.next()
	files:= p.files()
	return files, p.first
}

//--------------------------------------------------------------------------------

// ParseBytes behaves like Parse but it reads the source from the []byte slice provided.
func ParseBytes(filename string, base *src.PosBase, src []byte, errh ErrorHandler, pragh PragmaHandler, mode Mode, f func(string)(string,error)) (
		map[string]*File, error) {
	return Parse(filename, base, &bytesReader{src}, errh, pragh, mode, f)
}

type bytesReader struct {
	data []byte
}

func (r *bytesReader) Read(p []byte) (int, error) {
	if len(r.data) > 0 {
		n := copy(p, r.data)
		r.data = r.data[n:]
		return n, nil
	}
	return 0, io.EOF
}

//--------------------------------------------------------------------------------

// ParseFile behaves like Parse but it reads the source from the named file.
func ParseFile(filename string, errh ErrorHandler, pragh PragmaHandler, mode Mode, getFile func(string)(string,error)) (map[string]*File, error) {
	f, err := os.Open(filename)
	if err != nil {
		if errh != nil {
			errh(err)
		}
		return nil, err
	}
	defer f.Close()
	return Parse(filename, src.NewFileBase(filename, filename), f, errh, pragh, mode, getFile)
}

//--------------------------------------------------------------------------------

// ParseReplCmd accepts as input a .gro file written out to disk from the previous
// run and a sequence of entered strings, joins them together and parses them,
// then returns a .gro file for execution, which will write out another .gro file
// to be read by the next call of ParseReplCmd.
func ParseReplCmd(fin string, sin []string) (string, error) {

	replVars:= []string{"a", "b"}
	replConsts:= []string{"a", "b"}
	outp:= "\"var(\\n"
	for _, r:= range replVars {
		outp = outp + r + "=\" + fmt.Sprintf(\"%#v\", " + r + ") + \"\\n"
	}
	outp = outp + ")\\n"
	outp = outp + "const(\\n"
	for _, r:= range replConsts {
		outp = outp + r + "=\" + fmt.Sprintf(\"%#v\", " + r + ") + \"\\n"
	}
	outp = outp + ")\\n\""
	impFragment:= "import(\n\t\"io/ioutil\"\n\t\"os\"\n)\n"
	varFragment:= "_= ioutil.WriteFile(\"temp/repl-var.gro\", []byte(" + outp + "), os.ModeExclusive)"

	return impFragment + varFragment, nil
}

//--------------------------------------------------------------------------------
