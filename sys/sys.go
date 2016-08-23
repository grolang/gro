// Copyright 2011 The Gro Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"github.com/grolang/gro/macro"
	"github.com/grolang/gro/parser"
)

// ============================================================================

type StringWriter struct{
	data string
}

func (sw *StringWriter) Write(b []byte) (int, error) {
	sw.data += string(b)
	return len(b), nil
}

//================================================================================

func GetHoldDir() (string, error){
  holddir:= "tmp/"
  _, err:= ioutil.ReadDir(holddir)
  if err != nil {
    err= os.MkdirAll(holddir, os.ModeDir)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error: %s\n", err)
      return "", err
    }
  }
  return holddir, nil
}

// ============================================================================

// If in == nil, the source is the contents of the file with the given filename.
func ProcessFile( filename string, in io.Reader, out io.Writer, fileSet *token.FileSet, parserMode parser.Mode) error {
	if in == nil {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	}

	src, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	files, frag, err:= parser.ParseMultiFile(fileSet, filename, src, predefMacros, parserMode)
	if err != nil {
		return err
	}

	if frag != "" {
		holdDir, _:= GetHoldDir()
		holdFile:= holdDir + "macros-run.go"
		_= ioutil.WriteFile(holdFile, []byte(frag), os.ModeExclusive)
		out, err:= exec.Command("go", "run", holdFile, filename).CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s executing %s\n%s", err, holdFile, out)
			return err
		}
		fmt.Fprintf(os.Stderr, "%s", out)

	} else {
		for name, file:= range files {
			ast.SortImports(fileSet, file)

			sw:= StringWriter{""}
			_= format.Node(&sw, fileSet, file)

			parentPath, _:= path.Split(name)
			err := os.MkdirAll(filepath.Dir(filename) + "\\" + parentPath, os.ModeDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "creating directory: %s\n", err)
				return err
			}

			err = ioutil.WriteFile(filepath.Dir(filename) + "\\" + name, []byte(sw.data), 0644)
		}
	}

	return err
}

// ============================================================================

//
func ProcessFileWithMacros(aliases map[rune]macro.StmtMacro, args []string, mode parser.Mode) error {
	filename:= args[0]
	in, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer in.Close()

	src, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	fileSet:= token.NewFileSet()
	files, _, err:= parser.ParseMultiFile(fileSet, filename, src, aliases, mode)
	if err != nil {
		return err
	}

	for name, file:= range files {
		ast.SortImports(fileSet, file)

		sw:= StringWriter{""}
		_= format.Node(&sw, fileSet, file)

		parentPath, _:= path.Split(name)
		err := os.MkdirAll(filepath.Dir(filename) + "\\" + parentPath, os.ModeDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "creating directory: %s\n", err)
			return err
		}

		err = ioutil.WriteFile(filepath.Dir(filename) + "\\" + name, []byte(sw.data), 0644)
	}

	return nil
}

//================================================================================

