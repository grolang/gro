// Copyright 2009-16 The Go and Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

import (
	"errors"
	"fmt"
	"github.com/grolang/gro/parser"
	"github.com/grolang/gro/token"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
)

// ============================================================================
var cmdPrepare = &Command{
	Run:       groPrepare,
	UsageLine: "prepare [flags] [path ...]",
	Short:     "generate the go files",
	Long: `
Prepare generates formatted Go programs from Gro scripts.
It uses the same whitespace as gofmt.

Without an explicit path, it processes the standard input and prints the reformatted source to the standard output.
Given a file, it operates on that file; given a directory, it operates on all .gro files in that directory, recursively.
(Files starting with a period are ignored.)
It then prints the generated Go source to the output files as determined by the Gro source.

`,
}

// ============================================================================
func groPrepare(cmd *Command, cmds []*Command, args []string) {
	fileSet:= token.NewFileSet()
	var parserMode parser.Mode

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "creating cpu profile: %s\n", err)
			SetExitStatus(2)
			return
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	parserMode = parser.ParseComments
	if allErrors {
		parserMode |= parser.AllErrors
	}

	if len(args) == 0 {
		Report(errors.New("No args were given."))
		return
	}

	for i := 0; i < len(args); i++ {
		path := args[i]
		switch dir, err := os.Stat(path); {
		case err != nil:
			Report(err)
		case dir.IsDir():
			visitFile:= func(path string, f os.FileInfo, err error) error {
				name := f.Name()
				if err == nil && !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".gro") {
					if wantMsgs {
						fmt.Fprintf(os.Stderr, ">>> preparing %s\n", path)
					}
					err = ProcessFile(path, nil, os.Stdout, fileSet, parserMode)
				}
				if err != nil {
					Report(err)
				}
				return nil
			}
			filepath.Walk(path, visitFile)
		default:
			if wantMsgs {
				fmt.Fprintf(os.Stderr, ">>> preparing %s\n", path)
			}
			if err := ProcessFile(path, nil, os.Stdout, fileSet, parserMode); err != nil {
				Report(err)
			}
		}
	}
}

// ============================================================================

