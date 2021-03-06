// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/grolang/gro/nodes"
	"github.com/grolang/gro/syntax/src"
)

var fast = flag.Bool("fast", false, "parse package files in parallel")
var src_ = flag.String("src", "parser.go", "source file to parse")
var verify = flag.Bool("verify", false, "verify idempotent printing")
var fullprint = flag.Bool("fullprint", false, "full print for output heavy tests")

//================================================================================
/*
The tests of the Gro extensions to Go's syntax are in files
with no gro_xxxx_test.go where xxxx is:

	blacklist - TestBlacklist
	comments - TestComments, TestLineTags, TestNewSyntax
	divisions - TestDivisions, TestMain, TestCurlies, TestShorthandAliases
	generics - TestGenerics
	initwrap - TestInitwrap, TestWithinProc
	macros - TestMacros, TestUseDecls, TestDynamic
*/
type groTestData []struct {
	num int
	fnm string
	src string
	xtr map[string]string
	prt map[string]string //extra field for checking success
	err string            //extra field for checking failure
}

const locnPrefix = "github.com/grolang/gro/syntax"

func groTest(t *testing.T, groTests groTestData) {
	flag.Parse()
	nums := map[int]bool{}
	for _, arg := range flag.Args() {
		num, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			panic("Invalid args to test -- they should be integers.")
		}
		nums[int(num)] = true
	}

	for _, tst := range groTests {
		if len(nums) > 0 && !nums[tst.num] {
			continue
		}
		if tst.prt != nil && tst.err != "" {
			t.Error("Both \"prt\" and \"err\" are defined in testdata but only one should be.")
		}
		getFile := func(filename string) (src string, err error) {
			if tst.xtr == nil {
				return "", errors.New("Extra file map not there.")
			}
			wd, _ := os.Getwd()
			filename = strings.TrimPrefix(strings.TrimPrefix(filename, filepath.ToSlash(wd)), "/")
			xtr, ok := tst.xtr[filename]
			if !ok {
				return "", errors.New("Extra file not in map.")
			}
			return xtr, nil
		}
		asts, err := ParseBytes(tst.fnm, src.NewFileBase(tst.fnm, tst.fnm), []byte(tst.src), nil, nil, 0, getFile)
		if tst.prt != nil {
			if err != nil {
				t.Error(fmt.Sprintf("Test %d: Error received: %s", tst.num, err))
				continue
			}
			if len(asts) != len(tst.prt) {
				rcd := "Received:\n"
				for k, _ := range asts {
					rcd += "\t" + k + "\n"
				}
				t.Error(fmt.Sprintf("Test %d: Expected %d files from ParsePackage but received %d.\n%s", tst.num, len(tst.prt), len(asts), rcd))
			}
			for fn, ast := range asts {
				fn = strings.TrimPrefix(strings.TrimPrefix(filepath.ToSlash(fn), locnPrefix), "/")
				want := tst.prt[fn]
				if got := StringWithLinebreaks(ast); got != want {
					//t.Errorf("Test %d: expected and received source not the same for file %s.\nExpected: %d bytes.\nReceived: %d bytes.\n\n",
					//tst.num, fn, len(want), len(got))
					t.Errorf("Test %d: expected and received source not the same for file %s.\n\n#### Expected:\n%s\n\n#### Received:\n%s\n\n",
						tst.num, fn, want, got)

					/*var outputDir = "C:/Users/gavin/Documents/Golang/"
					t.Errorf("Test %d: Expected and received source not the same.\nFiles have been output.\n", tst.num)
					err:= ioutil.WriteFile(outputDir + "FileExpected_" + fn + ".txt", []byte(want), 0644)
					if err != nil {
						t.Error(err)
					}
					err = ioutil.WriteFile(outputDir + "FileReceived_" + fn + ".txt", []byte(got), 0644)
					if err != nil {
						t.Error(err)
					}*/
				}
			}
		} else {
			if fmt.Sprintf("%s", err) != tst.err {
				t.Error(fmt.Sprintf("Test %d: Expected error: %s;\nbut received: %s", tst.num, tst.err, err))
				continue
			}
			if len(asts) != 0 {
				t.Error(fmt.Sprintf("Test %d: Expected 0 files from ParsePackage but received %d.\n", tst.num, len(asts)))
				for fn, ast := range asts {
					got := StringWithLinebreaks(ast)
					//t.Errorf("Test %d: Received source \"%s\" of length %d bytes.", tst.num, fn, len(got))
					t.Errorf("Test %d: Received source \"%s\" was:\n%s\n\n", tst.num, fn, got)
				}
			}
		}
	}
}

//================================================================================
func TestParse(t *testing.T) {
	ParseFile(*src_, func(err error) { t.Error(err) }, nil, 0, nil)
}

//================================================================================
func TestParseFile(t *testing.T) {
	_, err := ParseFile("", nil, nil, 0, nil)
	if err == nil {
		t.Error("missing io error")
	}

	var first error
	_, err = ParseFile("", func(err error) {
		if first == nil {
			first = err
		}
	}, nil, 0, nil)
	if err == nil || first == nil {
		t.Error("missing io error")
	}
	if err != first {
		t.Errorf("got %v; want first error %v", err, first)
	}
}

//================================================================================
func TestLineDirectives(t *testing.T) {
	for _, test := range []struct {
		src, msg  string
		filename  string
		line, col uint // 0-based
	}{
		// test validity of //line directive
		{"//line foo:123\n\"fmt\".Println(\"abc\")", "", "", 0, 0}, //no error expected
		{"//line 123\n\"fmt\".Println(\"abc\")", "", "", 0, 0},     //no error expected
		{"//line :", "invalid line number: ", "", 0, 8},
		{"//line :x", "invalid line number: x", "", 0, 8},
		{"//line foo :", "invalid line number: ", "", 0, 12},
		{"//line foo:123abc", "invalid line number: 123abc", "", 0, 11},
		//{"/**///line foo:x", "syntax error: gro-file empty", "", 0, 16}, //line directive not at start of line - ignored
		{"/**///line foo:x", "syntax error: gro-file empty", "", 0, 0},
		{"//line foo:0", "invalid line number: 0", "", 0, 11},
		{fmt.Sprintf("//line foo:%d", lineMax+1), fmt.Sprintf("invalid line number: %d", lineMax+1), "", 0, 11},

		// test effect of //line directive on (relative) position information
		{"//line foo:123\n   foo", "syntax error: unexpected name at top-level", "foo", 123 - linebase, 3},
		{"//line foo:123\n//line bar:345\nfoo", "syntax error: unexpected name at top-level", "bar", 345 - linebase, 0},
	} {
		_, err := ParseBytes("dud", nil, []byte(test.src), nil, nil, 0, nil)
		if err == nil && test.msg == "" { //blank message means no error expected
			continue
		}
		if err == nil {
			t.Errorf("%s: no error reported", test.src)
			continue
		}
		perr, ok := err.(Error)
		if !ok {
			t.Errorf("%s: got %v; want parser error", test.src, err)
			continue
		}
		if msg := perr.Msg; msg != test.msg {
			t.Errorf("%s: got msg = %q; want %q", test.src, msg, test.msg)
		}
		if filename := perr.Pos.RelFilename(); filename != test.filename {
			t.Errorf("%s: got filename = %q; want %q", test.src, filename, test.filename)
		}
		if line := perr.Pos.RelLine(); line != test.line+linebase {
			t.Errorf("%s: got line = %d; want %d", test.src, line, test.line+linebase)
		}
		if col := perr.Pos.Col(); col != test.col+colbase {
			t.Errorf("%s: got col = %d; want %d", test.src, col, test.col+colbase)
		}
	}
}

//================================================================================
func TestStdLib(t *testing.T) {
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	start := time.Now()

	type parseResult struct {
		filename string
		lines    uint
		err1     bool
		err2     bool
	}

	results := make(chan parseResult)
	go func() {
		defer close(results)
		for _, dir := range []string{
			filepath.Join(runtime.GOROOT(), "src"),
		} {
			walkDirs(t, dir, func(filename string) {
				if debug {
					fmt.Printf("parsing %s\n", filename)
				}
				asts, err := ParseFile(filename, nil, nil, 0, nil)
				if err != nil {
					t.Error(err)
					return
				}
				if len(asts) != 1 {
					t.Error(fmt.Sprintf("More than one file returned from parse of %s.", filename))
				}
				for _, ast := range asts {
					if *verify {
						results <- parseResult{filename, ast.Lines, verifyPrint1(filename, ast), verifyPrint2(filename, ast)}
					} else {
						results <- parseResult{filename, ast.Lines, false, false}
					}
				}
			})
		}
	}()

	var count, lines, err1, err2 uint
	for res := range results {
		count++
		lines += res.lines
		if *verify {
			if res.err1 {
				err1++
			}
			if res.err2 {
				err2++
			}
		}
		if testing.Verbose() {
			//fmt.Printf("%5d  %s (%d lines)\n", count, res.filename, res.lines)
		}
	}

	dt := time.Since(start)
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	dm := float64(m2.TotalAlloc-m1.TotalAlloc) / 1e6

	t.Logf("parsed %d lines (%d files) in %v (%d lines/s)\n",
		lines, count, dt, int64(float64(lines)/dt.Seconds()))
	t.Logf("allocated %.3fMb (%.3fMb/s)\n", dm, dm/dt.Seconds())
	if *verify {
		t.Logf("Note: %d files in StdLib (out of %d) verified 1st stage Ok (%d%%)\n",
			count-err1, count, (count-err1)*100/count)
		t.Logf("Note: %d files in StdLib (out of %d) verified 2nd stage Ok (%d%%)\n",
			count-err2, count, (count-err2)*100/count)
	}
}

//================================================================================
func walkDirs(t *testing.T, dir string, action func(string)) {
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Error(err)
		return
	}

	var files, dirs []string
	for _, fi := range fis {
		if fi.Mode().IsRegular() {
			if strings.HasSuffix(fi.Name(), ".go") {
				path := filepath.Join(dir, fi.Name())
				files = append(files, path)
			}
		} else if fi.IsDir() && fi.Name() != "testdata" {
			path := filepath.Join(dir, fi.Name())
			if !strings.HasSuffix(path, "/test") {
				dirs = append(dirs, path)
			}
		}
	}

	if *fast {
		var wg sync.WaitGroup
		wg.Add(len(files))
		for _, filename := range files {
			go func(filename string) {
				defer wg.Done()
				action(filename)
			}(filename)
		}
		wg.Wait()
	} else {
		for _, filename := range files {
			action(filename)
		}
	}

	for _, dir := range dirs {
		walkDirs(t, dir, action)
	}
}

//================================================================================
func verifyPrint1(filename string, ast1 *nodes.File) bool {
	text, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	buf0 := *bytes.NewBuffer(text)
	text0, err := format.Source(buf0.Bytes())
	if err != nil {
		panic(err)
	}
	buf0 = *bytes.NewBuffer(text0) //just in case std-lib file isn't formatted

	var buf1 bytes.Buffer
	_, err = Fprint(&buf1, ast1, true)
	if err != nil {
		panic(err)
	}
	text1, err := format.Source(buf1.Bytes())
	if err != nil {
		panic(err)
	}
	buf1 = *bytes.NewBuffer(text1)

	if bytes.Compare(buf0.Bytes(), buf1.Bytes()) != 0 {
		//if strings.HasSuffix(filename, "graphic.go") { //&& len(buf1.Bytes()) < 1000 {
		if false {
			fmt.Printf("len before: %d; len after: %d\n", len(buf0.Bytes()), len(buf1.Bytes()))
			fmt.Printf("BEFORE: --- %s ---\n", filename)
			fmt.Printf("%s\n", buf0.Bytes())
			fmt.Println("--------------------------------------------------------------------------------")
			fmt.Printf("AFTER: --- %s ---\n", filename)
			fmt.Printf("%s\n", buf1.Bytes())
			fmt.Println("================================================================================")
		}
		return true
	} else {
		return false
	}
}

//================================================================================
func verifyPrint2(filename string, ast1 *nodes.File) bool {
	var buf1 bytes.Buffer
	_, err := Fprint(&buf1, ast1, true)
	if err != nil {
		panic(err)
	}

	asts, err := ParseBytes(filename, src.NewFileBase(filename, filename), buf1.Bytes(), nil, nil, 0, nil)
	if err != nil {
		panic(err)
	}
	if len(asts) != 1 {
		panic(fmt.Sprintf("More than one file returned from parse of %s.", filename))
	}
	for _, ast2 := range asts {
		var buf2 bytes.Buffer
		_, err = Fprint(&buf2, ast2, true)
		if err != nil {
			panic(err)
		}
		if bytes.Compare(buf1.Bytes(), buf2.Bytes()) != 0 {
			if false {
				fmt.Printf("BEFORE: --- %s ---\n", filename)
				fmt.Printf("%s\n", buf1.Bytes())
				fmt.Println("--------------------------------------------------------------------------------")
				fmt.Printf("AFTER: --- %s ---\n", filename)
				fmt.Printf("%s\n", buf2.Bytes())
				fmt.Println("################################################################################")
			}
			return true
		} else {
			return false
		}
	}
	panic("should never reach here")
}

//================================================================================
