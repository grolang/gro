// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file exercises the import parser but also checks that
// some low-level packages do not have new dependencies added.

package pkgs

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

//================================================================================
// copied from stdlib's go/build/deps_test.go
//--------------------------------------------------------------------------------

// isMacro reports whether p is a package dependency macro
// (uppercase name).
func isMacro(p string) bool {
	return 'A' <= p[0] && p[0] <= 'Z'
}

func allowed(pkg string) map[string]bool {
	m := map[string]bool{}
	var allow func(string)
	allow = func(p string) {
		if m[p] {
			return
		}
		m[p] = true // set even for macros, to avoid loop on cycle

		// Upper-case names are macro-expanded.
		if isMacro(p) {
			for _, pp := range PkgDeps[p] {
				allow(pp)
			}
		}
	}
	for _, pp := range PkgDeps[pkg] {
		allow(pp)
	}
	return m
}

// listStdPkgs returns the same list of packages as "go list std".
func listStdPkgs(goroot string) ([]string, error) {
	// Based on cmd/go's matchPackages function.
	var pkgs []string

	src := filepath.Join(goroot, "src") + string(filepath.Separator)
	walkFn := func(path string, fi os.FileInfo, err error) error {
		if err != nil || !fi.IsDir() || path == src {
			return nil
		}

		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") || strings.HasPrefix(base, "_") || base == "testdata" {
			return filepath.SkipDir
		}

		name := filepath.ToSlash(path[len(src):])
		if name == "builtin" || name == "cmd" || strings.Contains(name, "golang_org") {
			return filepath.SkipDir
		}

		pkgs = append(pkgs, name)
		return nil
	}
	if err := filepath.Walk(src, walkFn); err != nil {
		return nil, err
	}
	return pkgs, nil
}

func TestDependencies(t *testing.T) {
	iOS := runtime.GOOS == "darwin" && (runtime.GOARCH == "arm" || runtime.GOARCH == "arm64")
	if runtime.GOOS == "nacl" || iOS {
		// Tests run in a limited file system and we do not
		// provide access to every source file.
		t.Skipf("skipping on %s/%s, missing full GOROOT", runtime.GOOS, runtime.GOARCH)
	}

	all, err := listStdPkgs(build.Default.GOROOT)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(all)

	for _, pkg := range all {
		imports, err := findImports(pkg)
		if err != nil {
			t.Error(err)
			continue
		}
		ok := allowed(pkg)
		var bad []string
		for _, imp := range imports {
			if !ok[imp] {
				bad = append(bad, imp)
			}
		}
		if bad != nil {
			t.Errorf("unexpected dependency: %s imports %v", pkg, bad)
		}
	}
}

var buildIgnore = []byte("\n// +build ignore")

func findImports(pkg string) ([]string, error) {
	dir := filepath.Join(build.Default.GOROOT, "src", pkg)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var imports []string
	var haveImport = map[string]bool{}
	for _, file := range files {
		name := file.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		var imp []string
		data, err := readImports(f, false, &imp)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("reading %v: %v", name, err)
		}
		if bytes.Contains(data, buildIgnore) {
			continue
		}
		for _, quoted := range imp {
			path, err := strconv.Unquote(quoted)
			if err != nil {
				continue
			}
			if !haveImport[path] {
				haveImport[path] = true
				imports = append(imports, path)
			}
		}
	}
	sort.Strings(imports)
	return imports, nil
}

//================================================================================
// copied from stdlib's go/build/read.go
//--------------------------------------------------------------------------------
type importReader struct {
	b    *bufio.Reader
	buf  []byte
	peek byte
	err  error
	eof  bool
	nerr int
}

func isIdent(c byte) bool {
	return 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' || c == '_' || c >= utf8.RuneSelf
}

var (
	errSyntax = errors.New("syntax error")
	errNUL    = errors.New("unexpected NUL in input")
)

// syntaxError records a syntax error, but only if an I/O error has not already been recorded.
func (r *importReader) syntaxError() {
	if r.err == nil {
		r.err = errSyntax
	}
}

// readByte reads the next byte from the input, saves it in buf, and returns it.
// If an error occurs, readByte records the error in r.err and returns 0.
func (r *importReader) readByte() byte {
	c, err := r.b.ReadByte()
	if err == nil {
		r.buf = append(r.buf, c)
		if c == 0 {
			err = errNUL
		}
	}
	if err != nil {
		if err == io.EOF {
			r.eof = true
		} else if r.err == nil {
			r.err = err
		}
		c = 0
	}
	return c
}

// peekByte returns the next byte from the input reader but does not advance beyond it.
// If skipSpace is set, peekByte skips leading spaces and comments.
func (r *importReader) peekByte(skipSpace bool) byte {
	if r.err != nil {
		if r.nerr++; r.nerr > 10000 {
			panic("go/build: import reader looping")
		}
		return 0
	}

	// Use r.peek as first input byte.
	// Don't just return r.peek here: it might have been left by peekByte(false)
	// and this might be peekByte(true).
	c := r.peek
	if c == 0 {
		c = r.readByte()
	}
	for r.err == nil && !r.eof {
		if skipSpace {
			// For the purposes of this reader, semicolons are never necessary to
			// understand the input and are treated as spaces.
			switch c {
			case ' ', '\f', '\t', '\r', '\n', ';':
				c = r.readByte()
				continue

			case '/':
				c = r.readByte()
				if c == '/' {
					for c != '\n' && r.err == nil && !r.eof {
						c = r.readByte()
					}
				} else if c == '*' {
					var c1 byte
					for (c != '*' || c1 != '/') && r.err == nil {
						if r.eof {
							r.syntaxError()
						}
						c, c1 = c1, r.readByte()
					}
				} else {
					r.syntaxError()
				}
				c = r.readByte()
				continue
			}
		}
		break
	}
	r.peek = c
	return r.peek
}

// nextByte is like peekByte but advances beyond the returned byte.
func (r *importReader) nextByte(skipSpace bool) byte {
	c := r.peekByte(skipSpace)
	r.peek = 0
	return c
}

// readKeyword reads the given keyword from the input.
// If the keyword is not present, readKeyword records a syntax error.
func (r *importReader) readKeyword(kw string) {
	r.peekByte(true)
	for i := 0; i < len(kw); i++ {
		if r.nextByte(false) != kw[i] {
			r.syntaxError()
			return
		}
	}
	if isIdent(r.peekByte(false)) {
		r.syntaxError()
	}
}

// readIdent reads an identifier from the input.
// If an identifier is not present, readIdent records a syntax error.
func (r *importReader) readIdent() {
	c := r.peekByte(true)
	if !isIdent(c) {
		r.syntaxError()
		return
	}
	for isIdent(r.peekByte(false)) {
		r.peek = 0
	}
}

// readString reads a quoted string literal from the input.
// If an identifier is not present, readString records a syntax error.
func (r *importReader) readString(save *[]string) {
	switch r.nextByte(true) {
	case '`':
		start := len(r.buf) - 1
		for r.err == nil {
			if r.nextByte(false) == '`' {
				if save != nil {
					*save = append(*save, string(r.buf[start:]))
				}
				break
			}
			if r.eof {
				r.syntaxError()
			}
		}
	case '"':
		start := len(r.buf) - 1
		for r.err == nil {
			c := r.nextByte(false)
			if c == '"' {
				if save != nil {
					*save = append(*save, string(r.buf[start:]))
				}
				break
			}
			if r.eof || c == '\n' {
				r.syntaxError()
			}
			if c == '\\' {
				r.nextByte(false)
			}
		}
	default:
		r.syntaxError()
	}
}

// readImport reads an import clause - optional identifier followed by quoted string -
// from the input.
func (r *importReader) readImport(imports *[]string) {
	c := r.peekByte(true)
	if c == '.' {
		r.peek = 0
	} else if isIdent(c) {
		r.readIdent()
	}
	r.readString(imports)
}

// readComments is like ioutil.ReadAll, except that it only reads the leading
// block of comments in the file.
func readComments(f io.Reader) ([]byte, error) {
	r := &importReader{b: bufio.NewReader(f)}
	r.peekByte(true)
	if r.err == nil && !r.eof {
		// Didn't reach EOF, so must have found a non-space byte. Remove it.
		r.buf = r.buf[:len(r.buf)-1]
	}
	return r.buf, r.err
}

// readImports is like ioutil.ReadAll, except that it expects a Go file as input
// and stops reading the input once the imports have completed.
func readImports(f io.Reader, reportSyntaxError bool, imports *[]string) ([]byte, error) {
	r := &importReader{b: bufio.NewReader(f)}

	r.readKeyword("package")
	r.readIdent()
	for r.peekByte(true) == 'i' {
		r.readKeyword("import")
		if r.peekByte(true) == '(' {
			r.nextByte(false)
			for r.peekByte(true) != ')' && r.err == nil {
				r.readImport(imports)
			}
			r.nextByte(false)
		} else {
			r.readImport(imports)
		}
	}

	// If we stopped successfully before EOF, we read a byte that told us we were done.
	// Return all but that last byte, which would cause a syntax error if we let it through.
	if r.err == nil && !r.eof {
		return r.buf[:len(r.buf)-1], nil
	}

	// If we stopped for a syntax error, consume the whole file so that
	// we are sure we don't change the errors that go/parser returns.
	if r.err == errSyntax && !reportSyntaxError {
		r.err = nil
		for r.err == nil && !r.eof {
			r.readByte()
		}
	}

	return r.buf, r.err
}

//================================================================================
func TestStdLib(t *testing.T) {
	type parseResult struct {
		filename string
		lines    uint
	}
	processed:= map[string]bool{}
	for _, dir := range []string{
		filepath.Join(runtime.GOROOT(), "src"),
	} {
		walkDirs(t, dir, func(filename string) {
			pth, _:= filepath.Split(filename)
			pth = strings.Trim(pth[len(dir):], "\\")
			pth = strings.Replace(pth, "\\", "/", -1)
			if _, ok:= PkgDeps[pth]; !ok {
				if _, ok:= processed[pth]; !ok && !strings.HasPrefix(pth, "vendor") && !strings.HasPrefix(pth, "cmd") {
					t.Errorf("not in PkgDeps: %s\n", pth)
					processed[pth] = true
				}
			}
		})
	}
}

// same function as in "gro/syntax" package
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
	for _, filename := range files {
		action(filename)
	}
	for _, dir := range dirs {
		walkDirs(t, dir, action)
	}
}

//================================================================================
