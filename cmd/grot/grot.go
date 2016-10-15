// Copyright 2016 The Go and Gro Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command grot provides testing utilities to complement those in the Gro command.
package main

import (
	"fmt"
	"errors"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"github.com/grolang/gro/parser"
	"github.com/grolang/gro/sys"
	"github.com/grolang/gro/tutor"
)

const (
	pgmName = "grot"
	pgmVers = "0.4.0.0"
)

// ============================================================================
func main() {
	sys.SetProgData(pgmName, "managing Gro scripts, with root services", pgmVers)
	sys.AddCommand(cmdDoc)
	sys.Main()
}

// ============================================================================
var cmdDoc = &sys.Command{
	Run:         runDoc,
	UsageLine:   "doc [type] [name]",
	Short:       "create various forms of documentation from the Gro source",
	Long: `
Doc creates various types of documentation from the Gro source.

The types are:

	cmd      prints documentation for all available gro sub-commands in
	         a single go file.

	tut      prints tutorial "name" in the form of a doc.go file.
	         If no "name", prints all available tutorials in a single file.

	test     test runs all gro code in the tutorial "name".
	         If no "name", test runs code in all available tutorials.
`,
}

// ============================================================================
var cmdDocTemplate = `// Copyright 2016 The Gro Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// DO NOT EDIT THIS FILE. GENERATED BY "grot doc".
// After changing the documentation text in source files,
// rerun "grot doc cmd >> filename.go" to regenerate this file.

/*
Command gro provides utilities to complement those in the Go command.

{{range .}}{{if .Short}}{{.Short | capitalize}}

{{end}}{{if .Runnable}}Usage:

	gro {{.UsageLine}}

{{end}}{{.Long | trim}}


{{end}}*/
package main
`

// ============================================================================
var tutDocTemplate = `// Copyright 2016 The Gro Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// DO NOT EDIT THIS FILE. GENERATED BY "grot doc".
// After changing the English language tutorial in source files,
// rerun "grot doc tut tutname >> filename.go" to regenerate this file.

/*
Package {{.Name}} provides in doc format a tutorial that {{if .Short}}{{.Short}}{{end}}.

{{range .Pages}}{{.EnglishPage}}

{{end}}*/
package {{.Name}}

`

// ============================================================================
func runDoc(cmd *sys.Command, cmds []*sys.Command, args []string) {
	if len(args) == 0 {
		sys.Report(errors.New("No args were given."))
		return
	}

	var outf *os.File
	if sys.OutFile != "" {

		parentPath, _:= path.Split(sys.OutFile)
		err := os.MkdirAll(parentPath, os.ModeDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "creating directory: %s\n", err)
			return
		}

		f, err := os.Create(sys.OutFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "creating output file: %s\n", err)
			sys.SetExitStatus(2)
			return
		}
		defer f.Close()
		outf = f
	} else {
		outf = os.Stdout
	}

	switch args[0] {
	case "cmd":
		sys.MakeTemplate(outf, cmdDocTemplate, cmds)

	case "tut":
		if len(args) >= 2 {
			for _, t:= range sys.Tutorials() {
				if t.Name == args[1] {
					sys.MakeTemplate(outf, tutDocTemplate, t)
					return
				}
			}
		} else {
			for _, t:= range sys.Tutorials() {
				sys.MakeTemplate(outf, tutDocTemplate, t)
			}
		}

	case "test":
		if len(args) >= 2 {
			for _, t:= range sys.Tutorials() {
				if t.Name == args[1] {
					testTutorial(t)
					return
				}
			}
		} else {
			for _, t:= range sys.Tutorials() {
				testTutorial(t)
			}
		}

	}
}

// ============================================================================
func testTutorial (t *tutor.Tutorial) {
	fileSet:= token.NewFileSet()
	var parserMode parser.Mode

	holdDir, _:= sys.GetHoldDir()
	holdFile:= holdDir + "testTut.gro"
	for i, p:= range t.Pages {
		for j, c:= range p.Code {
			_= ioutil.WriteFile(holdFile, []byte(c), os.ModeExclusive)
			if err := sys.ProcessFile(holdFile, nil, os.Stdout, fileSet, parserMode); err != nil {
				fmt.Fprintf(os.Stderr, "... page %d code %d:\n", i, j)
				sys.Report(err)
			}
		}
	}
}

// ============================================================================

