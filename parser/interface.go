// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains the exported entry points for invoking the parser.

package parser

import (
	"bytes"
	"errors"
	"go/ast"
	"go/token"
	"io"
	"io/ioutil"
	"github.com/grolang/gro/macro"
)

// ============================================================================

// If src != nil, readSource converts src to a []byte if possible;
// otherwise it returns an error. If src == nil, readSource returns
// the result of reading the file specified by filename.
//
func readSource(filename string, src interface{}) ([]byte, error) {
	if src != nil {
		switch s := src.(type) {
		case string:
			return []byte(s), nil
		case []byte:
			return s, nil
		case *bytes.Buffer:
			// is io.Reader, but src is already available in []byte form
			if s != nil {
				return s.Bytes(), nil
			}
		case io.Reader:
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, s); err != nil {
				return nil, err
			}
			return buf.Bytes(), nil
		}
		return nil, errors.New("invalid source")
	}
	return ioutil.ReadFile(filename)
}

// ============================================================================

// A Mode value is a set of flags (or 0).
// They control the amount of source code parsed and other optional
// parser functionality.
//
type Mode uint

const (
	PackageClauseOnly Mode             = 1 << iota // stop parsing after package clause
	ImportsOnly                                    // stop parsing after import declarations
	ParseComments                                  // parse comments and add them to AST
	Trace                                          // print a trace of parsed productions
	DeclarationErrors                              // report declaration errors
	SpuriousErrors                                 // same as AllErrors, for backward-compatibility
	AllErrors         = SpuriousErrors             // report all errors (not just the first 10 on different lines)
	IgnoreUsesClause                               // don't produce fragment based on uses clause, but continue parsing instead
)

// ============================================================================

// ParseMultiFile parses the source code of a single source file which 
// contains many Go source files, and returns the corresponding map of
// ast.File nodes. The source code may be provided via
// the filename of the source file, or via the src parameter.
//
// If src != nil, ParseMultiFile parses the source from src and the filename is
// only used when recording position information. The type of the argument
// for the src parameter must be string, []byte, or io.Reader.
// If src == nil, ParseMultiFile parses the file specified by filename.
//
// The mode parameter controls the amount of source text parsed and other
// optional parser functionality. Position information is recorded in the
// file set fset.
//
// If the source couldn't be read, the returned AST is nil and the error
// indicates the specific failure. If the source was read but syntax
// errors were found, the result is a partial AST (with ast.Bad* nodes
// representing the fragments of erroneous source code). Multiple errors
// are returned via a scanner.ErrorList which is sorted by file position.
//
func ParseMultiFile(
		fset *token.FileSet, filename string, src interface{}, aliases map[rune]macro.StmtMacro, mode Mode) (
		fm map[string]*ast.File, frag string, err error) {
	// get source
	text, err := readSource(filename, src)
	if err != nil {
		return nil, "", err
	}

	if mode & IgnoreUsesClause == 0 {
		var q parser
		q.init(token.NewFileSet(), filename, text, nil, mode)
		frag = q.parseUsesOnly()
	}

	// parse source
	if frag == "" {
		var p parser
		defer func() {
			if e := recover(); e != nil {
				// resume same panic if it's not a bailout
				if _, ok := e.(bailout); !ok {
					panic(e)
				}
			}
			p.errors.Sort()
			err = p.errors.Err()
		}()

		p.init(fset, filename, text, aliases, mode)
		fm = p.parseMultiFile()
	}
	return
}

// ============================================================================

// ParseFile parses the source code of a single source file which must contain
// at most one Gro source file with no macros, and returns the corresponding
// ast.File node. The remaining functionality is the same as for ParseMultiFile
func ParseFile(fset *token.FileSet, filename string, src interface{}, mode Mode) (f *ast.File, err error) {
	fm, frag, err:= ParseMultiFile(fset, filename, src, nil, mode)
	if frag != "" {
		err = errors.New("Source file contains macros.")
	} else if len(fm) != 1 {
		err = errors.New("Exactly one file wasn't returned.")
	}
	for _, v:= range fm {
		f = v
		break
	}
	return
}

// ============================================================================

const (
	NoCmd = iota
	ExitCmd
	PrevCmd
	NextCmd
	LearnCmd
)

type ReplState struct {
	CmdFound       int
	VarFragment    string
	ImpFragment    string
	Tutorial       string
}

// ParseRepl parses the code of a single REPL command.
//
func ParseRepl(fset *token.FileSet, src interface{}, mode Mode) (fm map[string]*ast.File, rs ReplState, err error) {
	text:= []byte(src.(string))

	var p parser
	defer func() {
		if e := recover(); e != nil {
			// resume same panic if it's not a bailout
			if _, ok := e.(bailout); !ok {
				panic(e)
			}
		}
		p.errors.Sort()
		err = p.errors.Err()
	}()

	// parse source
	p.init(fset, "", text, nil, mode)
	fm = p.parseRepl()
	rs = p.replState

	return
}

// ============================================================================

