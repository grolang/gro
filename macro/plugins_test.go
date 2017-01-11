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
type StringWriter struct{ data string }
func (sw *StringWriter) Write(b []byte) (int, error) {
	sw.data += string(b)
	return len(b), nil
}

func TestUnihan(t *testing.T) {
	for i, tst:= range unihanTests {
		src:= tst.key
		dst:= tst.val
		dbug:= tst.debug
		expFrag:= tst.frag
		fset := token.NewFileSet() // positions are relative to fset
		fm, frag, err := parser.ParseMultiFile(fset, "suppliedName", src, nil, 0)
		if err != nil {
			t.Errorf("parse error in run %d: %q", i, err)
		} else if frag != "" {
			if frag != expFrag {
				t.Errorf("unexpected fragment returned in run %d:\nreceived:%q\nexpected:%q\n", i, frag, expFrag)
			}
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

const (
	Normal = iota
	Debug
	Tree
)

var unihanTests = map[int]struct{
	key   string
	debug int
	val   map[string]string
	frag  string
} {

// ========== ========== ========== ==========
24:{`
用㗢"github.com/grolang/groo/macro"
功正(){
	形Println('a')
	㗢{
		形Println('a')
	}
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

import (
	fmt "fmt"
	ops "github.com/grolang/groo/ops"
)

func main() {
	fmt.Println('a')
	{
		fmt.Println(ops.Runex("a"))
	}
}
`},

`package main
import (
	"flag"
	"github.com/grolang/gro/macro"
	"github.com/grolang/gro/parser"
	"github.com/grolang/gro/sys"
	_macro "github.com/grolang/groo/macro"
)
var aliases = map[rune]macro.StmtMacro {
	'㗢': _macro.Struct{},
}
func main () {
	flag.Parse()
	args := flag.Args()
	err := sys.ProcessFileWithMacros(aliases, args, parser.IgnoreUsesClause)
	if err != nil {
		sys.Report(err)
	}
}
`},

// ========== ========== ========== ==========
25:{`
用㗢"github.com/grolang/dyn/macro"
功正(){
	形Println('a')
	形Println(+6)
	形Println(-7)
	形Println(13+7)
	形Println(13-7)
	形Println(13*7)
	形Println(13/7)
	形Println(14/7)
	㗢{
		形Println('a')
		形Println('ab')
		形Println(!'ab')
		形Println(+6)
		形Println(-7)
		形Println(13+7)
		形Println(13-7)
		形Println(13*7)
		形Println(13/7)
		形Println(14/7)
	}
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

import (
	fmt "fmt"
	ops "github.com/grolang/dyn/ops"
)

func main() {
	fmt.Println('a')
	fmt.Println(+6)
	fmt.Println(-7)
	fmt.Println(13 + 7)
	fmt.Println(13 - 7)
	fmt.Println(13 * 7)
	fmt.Println(13 / 7)
	fmt.Println(14 / 7)
	{
		fmt.Println(ops.Runex("a"))
		fmt.Println(ops.Runex("ab"))
		fmt.Println(ops.Not(ops.Runex("ab")))
		fmt.Println(ops.Identity(6))
		fmt.Println(ops.Negate(7))
		fmt.Println(ops.Plus(13, 7))
		fmt.Println(ops.Minus(13, 7))
		fmt.Println(ops.Mult(13, 7))
		fmt.Println(ops.Divide(13, 7))
		fmt.Println(ops.Divide(14, 7))
	}
}
`},
`package main
import (
	"flag"
	"github.com/grolang/gro/macro"
	"github.com/grolang/gro/parser"
	"github.com/grolang/gro/sys"
	_macro "github.com/grolang/dyn/macro"
)
var aliases = map[rune]macro.StmtMacro {
	'㗢': _macro.Struct{},
}
func main () {
	flag.Parse()
	args := flag.Args()
	err := sys.ProcessFileWithMacros(aliases, args, parser.IgnoreUsesClause)
	if err != nil {
		sys.Report(err)
	}
}
`},

// ========== ========== ========== ==========
//cribbed from #7 to test "用"
27:{`用㕨"github.com/grolang/samples/moremacs"
包正
种readOp构{key整;resp通整}
种writeOp构{key整;val整;resp通双}
功正(){
    时Sleep(时Second)
    让range:="abc"
    让range="abcdefg"
	形Printf("range: %v\n",range)
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

import (
	fmt "fmt"
	time "time"
)

type _readOp struct {
	_key  int
	_resp chan int
}
type _writeOp struct {
	_key  int
	_val  int
	_resp chan bool
}

func main() {
	time.Sleep(time.Second)
	_range := "abc"
	_range = "abcdefg"
	fmt.Printf("range: %v\n", _range)
}
`},

//usesFrag:
`package main
import (
	"flag"
	"github.com/grolang/gro/macro"
	"github.com/grolang/gro/parser"
	"github.com/grolang/gro/sys"
	_moremacs "github.com/grolang/samples/moremacs"
)
var aliases = map[rune]macro.StmtMacro {
	'㕨': _moremacs.Struct{},
}
func main () {
	flag.Parse()
	args := flag.Args()
	err := sys.ProcessFileWithMacros(aliases, args, parser.IgnoreUsesClause)
	if err != nil {
		sys.Report(err)
	}
}
`},

// ========== ========== ========== ==========
28:{`
用㗢"github.com/grolang/gro/macro/whitelist"
源"whitelist.go"
func main() {
	回
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{},
`package main
import (
	"flag"
	"github.com/grolang/gro/macro"
	"github.com/grolang/gro/parser"
	"github.com/grolang/gro/sys"
	_whitelist "github.com/grolang/gro/macro/whitelist"
)
var aliases = map[rune]macro.StmtMacro {
	'㗢': _whitelist.Struct{},
}
func main () {
	flag.Parse()
	args := flag.Args()
	err := sys.ProcessFileWithMacros(aliases, args, parser.IgnoreUsesClause)
	if err != nil {
		sys.Report(err)
	}
}
`,},

// ========== ========== ========== ==========
999:{`
package main
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{
},""},

// ========== ========== ========== ==========
}

//================================================================================================================================================================
func TestUnihanParseError(t *testing.T) {
	for i, tst:= range unihanParseErrorTests {
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

		_, _, err := parser.ParseMultiFile(fset, "suppliedName", src, nil, 0)
		if fmt.Sprintf("%v", err) != dse {
			t.Errorf("unexpected parse error in %d: received error: %q; expected error: %q", i, err, dse)
		}
	}
}

var unihanParseErrorTests = map[int]struct{key string; val string} {

// ========== ========== ========== ==========
1999:{`
package
`,

// ---------- ---------- ---------- ----------
`suppliedName:2:9: expected ';', found 'EOF'`},

// ========== ========== ========== ==========
}

//================================================================================================================================================================
