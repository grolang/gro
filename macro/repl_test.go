// Copyright 2016 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package macro_test

import (
	"bytes"
	"fmt"
	"github.com/grolang/gro/parser"
	"github.com/grolang/gro/ast"
	"github.com/grolang/gro/format"
	"github.com/grolang/gro/importer"
	"github.com/grolang/gro/printer"
	"github.com/grolang/gro/token"
	"github.com/grolang/gro/types"
	"testing"
)

//================================================================================================================================================================
func TestRepl(t *testing.T) {
	for i, tst:= range replTests {
		src:= tst.key
		dst:= tst.val
		dbug:= tst.debug
		fset := token.NewFileSet() // positions are relative to fset
		fm, _, err := parser.ParseRepl(fset, src, 0)
		if err != nil {
			t.Errorf("parse error in run %d: %q", i, err)
		} else {
			for k, f:= range fm {
				var conf = types.Config{
					Importer: importer.Default(),
				}
				info := types.Info{
					Types: make(map[ast.Expr]types.TypeAndValue),
					Defs:  make(map[*ast.Ident]types.Object),
					Uses:  make(map[*ast.Ident]types.Object),
				}
				_, err = conf.Check("testing", fset, []*ast.File{f}, &info)
				if err != nil {
					t.Errorf("type check error in run %d: %q", i, err)
				}
				sw:= StringWriter{""}
				_= format.Node(&sw, fset, f)

				if dbug == Debug {
					var buf bytes.Buffer
					printer.Fprint(&buf, fset, f)
					fmt.Printf("%s", &buf)
				} else if dbug == Tree {
					ast.Print(fset, f)
				}

				if sw.data != dst[k] {
					t.Errorf("unexpected Go source in run %d: for filename %q, received source: %q; expected source: %q", i, k, sw.data, dst[k])
				}
			}
			if len(fm) != len(dst) {
				t.Errorf("unexpected Go sources in run %d: received %d sources; expected %d sources", i, len(fm), len(dst))
			}
		}
	}
}

var replTests = map[int]struct{
	key   string
	debug int
	val   map[string]string
} {

// ========== ========== ========== ==========
1:{`
入"fmt"
出
功正(){
  fmt.Printf("Hi!\n") // comment here
}`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

import _fmt "fmt"

func main() {
	_fmt.Printf("Hi!\n")
}
`}},

// ========== ========== ========== ==========
999:{`
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{
}},

// ========== ========== ========== ==========
}

//================================================================================================================================================================
func TestReplParseError(t *testing.T) {
	for i, tst:= range replParseErrorTests {
		src:= tst.key
		dse:= tst.val
		fset := token.NewFileSet()
		/*defer func(){
			if x:= recover(); fmt.Sprintf("%v", x) != msg {
				tt.Errorf("assert %d failed.\n" +
					"....found recover:%v\n" +
					"...expected panic:%v\n", i, fmt.Sprintf("%v", x), msg)
			}
		}*/

		_, _, err := parser.ParseRepl(fset, src, 0)
		if fmt.Sprintf("%v", err) != dse {
			t.Errorf("unexpected parse error in %d: received error: %q; expected error: %q", i, err, dse)
		}
	}
}

var replParseErrorTests = map[int]struct{key string; val string} {

// ========== ========== ========== ==========
1001:{`
func main() {
	假:= true //this generates parse error "Unihan special identifier on left hand side"
}
`,

// ---------- ---------- ---------- ----------
`3:5: Unihan special identifier 假 on left hand side (and 1 more errors)`},

// ========== ========== ========== ==========
1999:{`
`,

// ---------- ---------- ---------- ----------
`2:9: expected ';', found 'EOF'`},

// ========== ========== ========== ==========
}

//================================================================================================================================================================
