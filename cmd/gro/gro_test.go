// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main_test

import (
	"bytes"
	"fmt"
	"testing"
	"github.com/grolang/gro/cmd/gro"
	"github.com/grolang/gro/sys"
)

const usageStr = `
gro is a tool for managing Gro scripts.

Usage:

	gro command [arguments]

The commands are:

	prepare     generate the go files
	execute     generate the go files then run the main func
	version     print Gro version

Use "gro help [command]" for more information about a command.

Additional help topics:

	flags       flags used in Gro

Use "gro help [topic]" for more information about that topic.

`

func TestMain(t *testing.T) {
	var fn string
	var u, w *bytes.Buffer

	//--------------------------------------------------------------------------------
	//calling Gro without any args
	w = new(bytes.Buffer)
	sys.Stderr = w
	main.Main([]string{})
	if fmt.Sprintf("%s", w) != usageStr {
		t.Errorf("wrong text received from Stderr for calling gro without args:\n%s\n", w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro help'
	w = new(bytes.Buffer)
	sys.Stderr = w
	main.Main([]string{"help"})
	if fmt.Sprintf("%s", w) != usageStr {
		t.Errorf("wrong text received from Stderr for calling gro help:\n%s\n", w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro help prepare'
	prepareStr:= `
usage: gro prepare [flags] [path ...]

Prepare generates formatted Go programs from Gro scripts.
It uses the same whitespace as gofmt.

Without an explicit path, it processes the standard input and prints the reformatted source to the standard output.
Given a file, it operates on that file; given a directory, it operates on all .gro files in that directory, recursively.
(Files starting with a period are ignored.)
It then prints the generated Go source to the output files as determined by the Gro source.

`
	w = new(bytes.Buffer)
	sys.Stderr = w
	main.Main([]string{"help", "prepare"})
	if fmt.Sprintf("%s", w) != prepareStr {
		t.Errorf("wrong text received from Stderr for calling gro help with prepare as arg:\n%s\n", w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro startrek' i.e. unknown subcommand
	w = new(bytes.Buffer)
	sys.Stderr = w
	main.Main([]string{"startrek"})
	if fmt.Sprintf("%s", w) != "gro: unknown subcommand \"startrek\"\n" +
			"Run 'gro help' for usage.\n" {
		t.Errorf("wrong text received from Stderr for unknown subcommand:\n%s\n", w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro version'
	w = new(bytes.Buffer)
	sys.Stderr = w
	main.Main([]string{"version"})
	if fmt.Sprintf("%s", w) != "Gro version 0.7.2, running on Go version go1.9.2 for OS:windows and arch:amd64\n" {
		t.Errorf("wrong version text received from Stderr:\n%s\n", w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro version extra_arg'
	w = new(bytes.Buffer)
	sys.Stderr = w
	main.Main([]string{"version", "extra_arg"})
	if fmt.Sprintf("%s", w) != "usage: gro version\n\nToo many arguments given.\n" {
		t.Errorf("wrong text received from Stderr for version with superfluous args:\n%s\n", w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro prepare' i.e. not enough args
	w = new(bytes.Buffer)
	sys.Stderr = w
	main.Main([]string{"prepare"})
	if fmt.Sprintf("%s", w) != "gro: usage: gro prepare path\n" +
			"Not enough arguments given.\n" {
		t.Errorf("wrong text received from Stderr for prepare with no args:\n%s\n", w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro prepare somefile.gro'
	w = new(bytes.Buffer)
	sys.Stderr = w
	fn = "testdata/sayhi.gro" // we don't put "src/github.com/grolang/gro/cmd/gro/" in front
	main.Main([]string{"prepare", fn})
	if fmt.Sprintf("%s", w) != "" {
		t.Errorf("wrong text received from Stderr for prepare with file %s as arg:\n%s\n", fn, w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro prepare somefile.gro' with message flag
	w = new(bytes.Buffer)
	sys.Stderr = w
	fn = "testdata/sayhi.gro" // we don't put "src/github.com/grolang/gro/cmd/gro/" in front
	main.Main([]string{"prepare", "-v", fn})
	if fmt.Sprintf("%s", w) != "gro: preparing testdata/sayhi.gro\n" +
			"gro: Received 1 files from ParsePackage.\n" {
		t.Errorf("wrong text received from Stderr for prepare with file %s as arg:\n%s\n", fn, w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro prepare somedir'
	w = new(bytes.Buffer)
	sys.Stderr = w
	fn = "testdata/grodir"
	main.Main([]string{"prepare", fn})
	if fmt.Sprintf("%s", w) != "" {
		t.Errorf("wrong text received from Stderr for prepare with file %s as arg:\n%s\n", fn, w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro prepare somedir' with message flag
	w = new(bytes.Buffer)
	sys.Stderr = w
	fn = "testdata/grodir"
	main.Main([]string{"prepare", "-v", fn})
	if fmt.Sprintf("%s", w) != "gro: preparing testdata/grodir/saycat.gro\n" +
			"gro: Received 1 files from ParsePackage.\n" +
			"gro: preparing testdata/grodir/saydog.gro\n" +
			"gro: Received 1 files from ParsePackage.\n" {
		t.Errorf("wrong text received from Stderr for prepare with file %s as arg:\n%s\n", fn, w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro prepare firstfile.gro secondfile.gro'
	w = new(bytes.Buffer)
	sys.Stderr = w
	main.Main([]string{"prepare", "testdata/sayhi.gro", "testdata/saybye.gro"})
	if fmt.Sprintf("%s", w) != "" {
		t.Errorf("wrong text received from Stderr for prepare with files as arg:\n%s\n", w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro prepare firstfile.gro secondfile.gro' with message flag
	w = new(bytes.Buffer)
	sys.Stderr = w
	main.Main([]string{"prepare", "-v", "testdata/sayhi.gro", "testdata/saybye.gro"})
	if fmt.Sprintf("%s", w) != "gro: preparing testdata/sayhi.gro\n" +
			"gro: Received 1 files from ParsePackage.\n" +
			"gro: preparing testdata/saybye.gro\n" +
			"gro: Received 2 files from ParsePackage.\n" {
		t.Errorf("wrong text received from Stderr for prepare with files as arg:\n%s\n", w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro prepare nonexistent_file.gro'
	w = new(bytes.Buffer)
	sys.Stderr = w
	fn = "testdata/nonexistent_file.gro"
	main.Main([]string{"prepare", fn})
	if fmt.Sprintf("%s", w) != "gro: CreateFile testdata/nonexistent_file.gro: The system cannot find the file specified.\n" {
		t.Errorf("wrong text received from Stderr for prepare with file %s as arg:\n%s\n", fn, w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro execute' i.e. not enough args
	w = new(bytes.Buffer)
	sys.Stderr = w
	main.Main([]string{"execute"})
	if fmt.Sprintf("%s", w) != "gro: usage: gro execute path\n" +
			"Not enough arguments given.\n" {
		t.Errorf("wrong text received from Stderr for execute with no args:\n%s\n", w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro execute firstfile.gro secondfile.gro' i.e. too many args
	w = new(bytes.Buffer)
	sys.Stderr = w
	main.Main([]string{"execute", "file_one.gro", "file_two.gro"})
	if fmt.Sprintf("%s", w) != "gro: usage: gro execute path\n" +
			"Too many arguments given.\n" {
		t.Errorf("wrong text received from Stderr for execute with no args:\n%s\n", w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro execute somefile.gro'
	u = new(bytes.Buffer)
	sys.Stdout = u
	w = new(bytes.Buffer)
	sys.Stderr = w
	fn = "testdata/sayhi.gro"
	main.Main([]string{"execute", fn})
	if fmt.Sprintf("%s", u) != "Hello, world!\n" {
		t.Errorf("wrong text received from Stdout for execute with file %s as arg:\n%s\n", fn, u)
	}
	if fmt.Sprintf("%s", w) != "" {
		t.Errorf("wrong text received from Stderr for execute with file %s as arg:\n%s\n", fn, w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro execute somefile.gro' with message flag
	u = new(bytes.Buffer)
	sys.Stdout = u
	w = new(bytes.Buffer)
	sys.Stderr = w
	fn = "testdata/sayhi.gro"
	main.Main([]string{"execute", "-v", fn})
	if fmt.Sprintf("%s", u) != "Hello, world!\n" {
		t.Errorf("wrong text received from Stdout for execute with file %s as arg:\n%s\n", fn, u)
	}
	if fmt.Sprintf("%s", w) != "gro: preparing testdata/sayhi.gro\n" +
			"gro: Received 1 files from ParsePackage.\n" +
			"gro: running testdata/sayhi.go\n" {
		t.Errorf("wrong text received from Stderr for execute with file %s as arg:\n%s\n", fn, w)
	}

	//--------------------------------------------------------------------------------
	//calling 'gro execute executeInside.gro' with message flag
	u = new(bytes.Buffer)
	sys.Stdout = u
	w = new(bytes.Buffer)
	sys.Stderr = w
	fn = "testdata/executeInside.gro"
	main.Main([]string{"execute", "-v", fn})
	if fmt.Sprintf("%s", u) != "'Hello, world!' from executeInside.gro\n" +
			"Hello, world!\n" {
		t.Errorf("wrong text received from Stdout for execute with file %s as arg:\n%s\n", fn, u)
	}
	if fmt.Sprintf("%s", w) != "gro: preparing testdata/executeInside.gro\n" +
			"gro: Received 1 files from ParsePackage.\n" +
			"gro: running testdata/executeInside.go\n" { //only executeInside.gro is run in verbose mode, not sayhi.gro
		t.Errorf("wrong text received from Stderr for execute with file %s as arg:\n%s\n", fn, w)
	}

//--------------------------------------------------------------------------------
}

