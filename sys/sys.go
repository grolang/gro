// Copyright 2016-17 The Go and Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"github.com/grolang/gro/syntax"
)

var (
	// default input and output -- can be changed by test suite in gro/cmd/gro
	Stderr io.Writer = os.Stderr
	Stdout io.Writer = os.Stdout
	Stdin  io.Reader = os.Stdin

	ProgName = "gro"
	WantMsgs bool
	ExitStatus = 0
)

const Suffix = "gro"

var exitMu sync.Mutex

func setExitStatus(n int) {
	exitMu.Lock()
	if ExitStatus < n {
		ExitStatus = n
	}
	exitMu.Unlock()
}

//================================================================================
func Prepare(args ...string) {
	if len(args) < 1 {
		fmt.Fprintf(Stderr, "%s: usage: gro prepare path\nNot enough arguments given.\n", ProgName)
		setExitStatus(2)
		return
	}
	for i := 0; i < len(args); i++ {
		pth := args[i]
		switch dir, err := os.Stat(pth); {
		case err != nil:
			fmt.Fprintf(Stderr, "%s: %s\n", ProgName, err)
			setExitStatus(2)
			return
		case dir.IsDir():
			filepath.Walk(pth, func(pth string, f os.FileInfo, err error) error {
				name := f.Name()
				pth = filepath.ToSlash(pth)
				if err == nil && !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, "."+Suffix) {
					if WantMsgs {
						fmt.Fprintf(Stderr, "%s: preparing %s\n", ProgName, pth)
					}
					err = processFile(pth, nil, Stdout)
				}
				if err != nil {
					fmt.Fprintf(Stderr, "%s: %s\n", ProgName, err)
					setExitStatus(2)
				}
				return nil
			})
		default:
			if WantMsgs {
				fmt.Fprintf(Stderr, "%s: preparing %s\n", ProgName, pth)
			}
			if err := processFile(pth, nil, Stdout); err != nil {
				fmt.Fprintf(Stderr, "%s: %s\n", ProgName, err)
				setExitStatus(2)
				return
			}
		}
	}
}

//--------------------------------------------------------------------------------
func GetFile(filename string) (src string, err error) {
	if WantMsgs {
		fmt.Fprintf(Stderr, "%s: Parsing extra file %s.\n", ProgName, filename)
	}
	f, err := os.Open(filepath.Join("src", filename))
	if err != nil {
		return "", err
	}
	defer f.Close()
	s, err:= ioutil.ReadAll(f)
	return string(s), err
}

//--------------------------------------------------------------------------------
// If in == nil, the source is the contents of the file with the given filename.
// TODO: not called anywhere with non-nil 'in' arg -- needs test
func processFile(filename string, in io.Reader, out io.Writer) error {
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

	asts, err := syntax.ParseBytes(filename, nil, src, nil, nil, 0, GetFile)
	if err != nil {
		fmt.Fprintf(Stderr, "%s: Error received: %s", ProgName, err)
		return err
	}
	if len(asts) != 0 && WantMsgs {
		fmt.Fprintf(Stderr, "%s: Received %d files from ParsePackage.\n", ProgName, len(asts))
	}
	for name, ast:= range asts {
		file:= syntax.StringWithLinebreaks(ast)
		parentPath, _:= filepath.Split(name)
		err := os.MkdirAll(filepath.Join(filepath.Dir(filename), parentPath), os.ModeDir)
		if err != nil {
			fmt.Fprintf(Stderr, "%s: Error creating directory: %s\n", ProgName, err)
			return err
		}
		err = ioutil.WriteFile(filepath.Join(filepath.Dir(filename), name), []byte(file), 0644)
	}

	return err
}

//================================================================================
func Execute(args ...string) {
	if len(args) < 1 {
		fmt.Fprintf(Stderr, "%s: usage: gro execute path\nNot enough arguments given.\n", ProgName)
		setExitStatus(2)
		return
	}
	if len(args) > 1 {
		fmt.Fprintf(Stderr, "%s: usage: gro execute path\nToo many arguments given.\n", ProgName)
		setExitStatus(2)
		return
	}
	Prepare(args...)
	if ExitStatus > 0 {
		return
	}

	extLen:= len(filepath.Ext(args[0]))
	outfile:= args[0][:len(args[0])-extLen] + ".go"

	if WantMsgs {
		fmt.Fprintf(Stderr, "%s: running %s\n", ProgName, outfile)
	}

	c:= exec.Command("go", "run", outfile)
	c.Stdin =  Stdin
	c.Stdout = Stdout
	c.Stderr = Stderr
	err:= c.Run()
	if err != nil {
		fmt.Fprintf(Stderr, "%s: Error: %s executing %s\n", ProgName, err, outfile)
		setExitStatus(2)
		return
	}
}

//================================================================================
