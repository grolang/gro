// Copyright 2016 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package parser_test

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

func TestGroSyntax(t *testing.T) {
	for i, tst:= range groSyntaxTests {
		src:= tst.key
		dst:= tst.val
		dbug:= tst.debug
		fset := token.NewFileSet() // positions are relative to fset
		fm, frag, err := parser.ParseMultiFile(fset, "suppliedName.gro", src, nil, 0)
		if err != nil {
			t.Errorf("parse error in run %d: %q", i, err)
		} else if frag != "" {
			t.Errorf("unexpected fragment returned in run %d:\nreceived:%q\n", i, frag)
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

var groSyntaxTests = map[int]struct{
	key   string
	debug int
	val   map[string]string
} {

// ========== ========== ========== ==========
//test keyword with scoping: 功入
1:{`
package main;入"fmt"
功main(){
  do fmt.Printf("Hi!\n") // comment here
}`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`// +build ignore
package main

import _fmt "fmt"

func _main() {
	_fmt.Printf("Hi!\n")
}
`}},

// ========== ========== ========== ==========
//test keyword with scoping: 功
//test keyword without scoping: 回
//test specids: 度,串,整,整64, 漂32,漂64,复,复64,复128(i.e. all float and complex specids)
2:{`
package main;入"fmt";
功main(){
  do fmt.Printf("Len: %d\n", 度(fs("abcdefg")))
}
功fs(a串)串{回a+"xyz"}
功ff(a漂32)漂64{回漂64(a)}
功fc(a复64)复128{回复128(a)+复(1,1)}
功fi(a整)整64{回整64(a)}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`// +build ignore
package main

import _fmt "fmt"

func _main() {
	_fmt.Printf("Len: %d\n", len(_fs("abcdefg")))
}
func _fs(_a string) string        { return _a + "xyz" }
func _ff(_a float32) float64      { return float64(_a) }
func _fc(_a complex64) complex128 { return complex128(_a) + complex(1, 1) }
func _fi(_a int) int64            { return int64(_a) }
`}},

// ========== ========== ========== ==========
//test keyword with scoping: 变,如,否
//test specids: 真,假
3:{`
package main;入"fmt"
import 吧"fmt"
import 哪_fg"fmt"
入㕤hij"fmt"
入卟"unicode/utf8"
入叨叩kl"unicode/utf8"
var n = 50
变p=70
变string=170
功main(){
  如真{
    fmt.Printf("Len: %d\n", 度("abcdefg") + p)
  }
}
func deputy(){
  if真{
    fmt.Printf("Len: %d\n", 度("abcdefg") + n)
  }
  如假{
    fg.Printf("Len: %d\n", 度("hijk") + p)
  }否{
    hij.Printf("Len: %d\n", 度("hi") + p)
  }
  do fr,_:= _utf8.DecodeRune([]byte("lmnop"))
  do fmt.Printf("1st rune: %s; Len: %d\n", fr, len("lmnop") + n)
  做_,_=kl.DecodeRune([]节("lmnop"))
  㕤Printf("Fifty: %d\n", n)
  哪Printf("Fifty: %d\n", n)
  吧Printf("Fifty: %d\n", n)
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`// +build ignore
package main

import _fmt "fmt"
import "fmt"
import _fg "fmt"
import _hij "fmt"
import _utf8 "unicode/utf8"
import _kl "unicode/utf8"

var n = 50
var _p = 70
var _string = 170

func _main() {
	if true {
		_fmt.Printf("Len: %d\n", len("abcdefg")+_p)
	}
}
func deputy() {
	if true {
		fmt.Printf("Len: %d\n", len("abcdefg")+n)
	}
	if false {
		_fg.Printf("Len: %d\n", len("hijk")+_p)
	} else {
		_hij.Printf("Len: %d\n", len("hi")+_p)
	}
	fr, _ := _utf8.DecodeRune([]byte("lmnop"))
	fmt.Printf("1st rune: %s; Len: %d\n", fr, len("lmnop")+n)
	_, _ = _kl.DecodeRune([]byte("lmnop"))
	_hij.Printf("Fifty: %d\n", n)
	_fg.Printf("Fifty: %d\n", n)
	fmt.Printf("Fifty: %d\n", n)
}
`}},

// ========== ========== ========== ==========
//test keywords with scoping: 种,久,构
//test specids: 整8,整16,整32
4:{`
package main
type _string string
种A struct{a string; b 整8}
种littleB struct{a string; b 整16}
种C构{a string; b 整32}
type D struct{a string; b *整32}
种E构{a串;b整32}
type F构{a串;b整32}
久a=3.1416
const b=2.72
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`// +build ignore
package main

type _string string
type A struct {
	_a _string
	_b int8
}
type _littleB struct {
	_a _string
	_b int16
}
type C struct {
	_a _string
	_b int32
}
type D struct {
	a string
	b *int32
}
type E struct {
	_a string
	_b int32
}
type F struct {
	_a string
	_b int32
}

const _a = 3.1416
const b = 2.72
`}},

// ========== ========== ========== ==========
//test keywords with scoping: 入,图,为,继,破
//test keyword without scoping: 围
//test specids: 节,字
//TODO: fix scoping for 图 and slices/arrays
5:{`
package main;入"fmt"
import "fmt"
type _byte byte
type _rune rune
var (
	_a= 图[字]节{'a': 127, 'b': 0, '7':7}
	_b= byte(7)
	c= 图[字]节{'a': b}
	cc= map[rune]byte{'a': _b}
	d= []节{127, b, 0}
	dd= [3]节{127, b, 0}
	e= 图[byte]rune{byte(b):'a'}
	_f= 7
	g= 构{a串;b整}{"abc",f} //TODO: why does this work, i.e. f become _f ?
	h= 功(a串)串{a="def"; return a}
	i 面{doIt(a串)串}
)
func main(){
	_zx:为i:=0;i<19;i++{
		if i==3 {继}; if i==6{破}
		如 i== 16{破zx}
		如 i== 17{继zx}
		fmt.Print(i," ")
	}
	do fmt.Println("abc")
	为i:=围a{fmt.Print(i," ")}
	for i:= 0; i<28; i++ {
		if i==3 { continue }
		if i==6 { break }
		do fmt.Print(i, " ")
	}
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`// +build ignore
package main

import _fmt "fmt"
import "fmt"

type _byte byte
type _rune rune

var (
	_a = map[rune]byte{'a': 127, 'b': 0, '7': 7}
	_b = byte(7)
	c  = map[rune]byte{'a': _b}
	cc = map[rune]byte{'a': _b}
	d  = []byte{127, _b, 0}
	dd = [3]byte{127, _b, 0}
	e  = map[_byte]_rune{_byte(_b): 'a'}
	_f = 7
	g  = struct {
		_a string
		_b int
	}{"abc", _f}
	h = func(_a string) string { _a = "def"; return _a }
	i interface {
		_doIt(_a string) string
	}
)

func main() {
_zx:
	for _i := 0; _i < 19; _i++ {
		if _i == 3 {
			continue
		}
		if _i == 6 {
			break
		}
		if _i == 16 {
			break _zx
		}
		if _i == 17 {
			continue _zx
		}
		_fmt.Print(_i, " ")
	}
	fmt.Println("abc")
	for _i := range _a {
		_fmt.Print(_i, " ")
	}
	for i := 0; i < 28; i++ {
		if i == 3 {
			continue
		}
		if i == 6 {
			break
		}
		fmt.Print(i, " ")
	}
}
`}},

// ========== ========== ========== ==========
//test unscoped keyword: 掉
//test keywords with scoping: 面,考,事,别
//test specids: 双,空,绝,绝8,绝16,绝32,绝64
6:{`package main;入"fmt"
type _uint16 uint16
type A interface {
  aMeth()绝
}
种B面{
  bMeth()绝8
}
种C interface{
  cMeth(theC uint16)绝16
}
type D面{
  dMeth(theD绝32)绝64
}
func abc()*双{回空}
func main(){
	do _a:=2
	考a{
	事1:
		fmt.Print('a')
	事2:
		fmt.Print('b')
		掉
	事3:
		fmt.Print('c')
	别:
		fmt.Print('d')
	}
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`// +build ignore
package main

import _fmt "fmt"

type _uint16 uint16
type A interface {
	aMeth() uint
}
type B interface {
	_bMeth() uint8
}
type C interface {
	_cMeth(_theC _uint16) uint16
}
type D interface {
	_dMeth(_theD uint32) uint64
}

func abc() *bool { return nil }
func main() {
	_a := 2
	switch _a {
	case 1:
		_fmt.Print('a')
	case 2:
		_fmt.Print('b')
		fallthrough
	case 3:
		_fmt.Print('c')
	default:
		_fmt.Print('d')
	}
}
`}},

// ========== ========== ========== ==========
//test keyword with scoping: 通
//test keyword without scoping: 选,去
//test special name: 正
//test using keyword "range" as id in LHS Unihan-context of "做", and in RHS
7:{`包正;入("math/rand";"sync/atomic")
种readOp构{key整;resp通整}
种writeOp构{key整;val整;resp通双}
功正(){
    变ops整64=0
    reads:=造(通*readOp)
    writes:=造(通*writeOp)
    去功(){
        变state=造(图[整]整)
        为{选{
          事read:=<-reads:read.resp<-state[read.key]
          事write:=<-writes:state[write.key]=write.val;write.resp<-真
          }}}()
    为r:=0;r<100;r++{
        去功(){
            为{read:=&readOp{key:rand.Intn(5),resp:造(通整)}
               reads<-read
               <-read.resp
               atomic.AddInt64(&ops,1)
              }}()}
    为w:=0;w<10;w++{
        去功(){
            为{write:=&writeOp{key:rand.Intn(5),val:rand.Intn(100),resp:造(通双)}
               writes<-write
               <-write.resp
               atomic.AddInt64(&ops,1)
              }}()}
    时Sleep(时Second)
    opsFinal:=atomic.LoadInt64(&ops)
    形Println("ops:",opsFinal)

    做range:="abc" //when used with 做, Go keywords like "range" can be used as identifiers
    做range="abcdefg"
	形Printf("range: %v\n",range)
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go": //FIX: should only have one "// +build ignore" comment
`// +build ignore
// +build ignore
package main

import (
	fmt "fmt"
	_rand "math/rand"
	_atomic "sync/atomic"
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
	var _ops int64 = 0
	_reads := make(chan *_readOp)
	_writes := make(chan *_writeOp)
	go func() {
		var _state = make(map[int]int)
		for {
			select {
			case _read := <-_reads:
				_read._resp <- _state[_read._key]
			case _write := <-_writes:
				_state[_write._key] = _write._val
				_write._resp <- true
			}
		}
	}()
	for _r := 0; _r < 100; _r++ {
		go func() {
			for {
				_read := &_readOp{_key: _rand.Intn(5), _resp: make(chan int)}
				_reads <- _read
				<-_read._resp
				_atomic.AddInt64(&_ops, 1)
			}
		}()
	}
	for _w := 0; _w < 10; _w++ {
		go func() {
			for {
				_write := &_writeOp{_key: _rand.Intn(5), _val: _rand.Intn(100), _resp: make(chan bool)}
				_writes <- _write
				<-_write._resp
				_atomic.AddInt64(&_ops, 1)
			}
		}()
	}
	time.Sleep(time.Second)
	_opsFinal := _atomic.LoadInt64(&_ops)
	fmt.Println("ops:", _opsFinal)
	{

		_range := "abc"
		_range = "abcdefg"
		fmt.Printf("range: %v\n", _range)
	}
}
`}},

// ========== ========== ========== ==========
//test keyword with scoping: 围
//test special name: 做
8:{`package main

import (
	//"fmt"
	"github.com/grolang/gro/parser"
	"github.com/grolang/gro/token"
	"github.com/grolang/gro/format"
	//"github.com/grolang/gro/ast"
	//"os"
)

func main() {
	形Printf("Hi!\n")

	do fset := token.NewFileSet() // positions are relative to fset

	do _f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		do fmt.Println(err)
		return
	}

	for _, s := 围 f.Imports {
		do fmt.Println(s.Path.Value)
	}

	{	do sw:= StringWriter{""}
		do _= format.Node(&sw, fset, _f)
		do fmt.Println(sw.data)
		do fmt.Println(sw.data == dst)
	}
	做{ abc:= "abc"
		_ = abc
	}
}

type StringWriter struct{ data string }
func (sw *StringWriter) Write(b []byte) (int, error) {
	do sw.data += string(b)
	return len(b), nil
}

var src string = "package main"
var dst string = "package main"`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go": //FIX: should only have one "// +build ignore" comment
`// +build ignore
// +build ignore
package main

import (
	fmt "fmt"
	"github.com/grolang/gro/format"
	"github.com/grolang/gro/parser"
	"github.com/grolang/gro/token"
)

func main() {
	fmt.Printf("Hi!\n")

	fset := token.NewFileSet()

	_f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, s := range _f.Imports {
		fmt.Println(s.Path.Value)
	}

	{
		sw := StringWriter{""}
		_ = format.Node(&sw, fset, _f)
		fmt.Println(sw.data)
		fmt.Println(sw.data == dst)
	}
	{
		_abc := "abc"
		_ = _abc
	}
}

type StringWriter struct{ data string }

func (sw *StringWriter) Write(b []byte) (int, error) {
	sw.data += string(b)
	return len(b), nil
}

var src string = "package main"
var dst string = "package main"
`}},

// ========== ========== ========== ==========
//test special names: 做,任,英,引
//test using keyword "func" as id in LHS Unihan-context of "做", and in RHS
//test using "source" as identifier name is OK
9:{`
package main;入"fmt"
import "fmt"
const a=6
功main(){
  做引"b":=7
  做func:=8
  fmt.Printf("Hi, nos.%s and %s!\n", b, func)
  英{
    fmt.Printf("Hi, no. %s.\n", a)
  }
}
func baba(){
  do source:="abc"
  do fmt.Printf("Hi, %s!\n", source)
  变b任= 17
  do fmt.Printf("Hi, no.%s!\n", _b)
}`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`// +build ignore
package main

import _fmt "fmt"
import "fmt"

const a = 6

func _main() {
	{
		_b := 7
		{
			_func := 8
			_fmt.Printf("Hi, nos.%s and %s!\n", _b, _func)
			{
				fmt.Printf("Hi, no. %s.\n", a)
			}
		}
	}
}
func baba() {
	source := "abc"
	fmt.Printf("Hi, %s!\n", source)
	var _b interface{} = 17
	fmt.Printf("Hi, no.%s!\n", _b)
}
`}},

// ========== ========== ========== ==========
//test special name: 做
//test 源 with defaulted option
10:{`
包main;源;入"fmt"
功main(){
  做b:=7
  做func:=8
  fmt.Printf("Hi, nos.%s and %s!\n", b, func) // different comment here
}`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`// +build ignore
package main

import _fmt "fmt"

func _main() {
	{
		_b := 7
		{
			_func := 8
			_fmt.Printf("Hi, nos.%s and %s!\n", _b, _func)
		}
	}
}
`}},

// ========== ========== ========== ==========
//test 源 with specified filename
11:{`package main
源"mySource"
func main() {
	a:= true
	b:= 真
	nil:= true
	iota:= 真
	形Printf("a: %v, b: %v, nil: %v, iota: %v\n", a, b, nil, iota)

	变abc图[串]整;
	做abc[串("def")]=789

	var _z= "abc"
	形Println(串(z))
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"mySource.go":
`// +build ignore
package main

import fmt "fmt"

func main() {
	a := true
	b := true
	nil := true
	iota := true
	fmt.Printf("a: %v, b: %v, nil: %v, iota: %v\n", a, b, nil, iota)

	var _abc map[string]int
	_abc[string("def")] = 789

	var _z = "abc"
	fmt.Println(string(_z))
}
`}},

// ========== ========== ========== ==========
//test special name: 做,引
12:{`
package main
type A int
func main(){
  引"_a":=123
  形Println(整64(a)) //why does this translate as _a when I don't think I've put in the code to detect Unihan used as type converters ???
  var b A
  var c这A
  形Println(b, c)
  做d:='d'
  形Println(_a,_d)
}
`,


// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`// +build ignore
package main

import fmt "fmt"

type A int

func main() {
	_a := 123
	fmt.Println(int64(_a))
	var b A
	var c A
	fmt.Println(b, c)
	{
		_d := 'd'
		fmt.Println(_a, _d)
	}
}
`}},

// ========== ========== ========== ==========
//test example of totally spaceless code
13:{`包正;种A整;功正(){a:=123;形Println(整64(a));变b这A;变c这A;形Println(b,c)}`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`// +build ignore
package main

import fmt "fmt"

type A int

func main() { _a := 123; fmt.Println(int64(_a)); var _b A; var _c A; fmt.Println(_b, _c) }
`}},

// ========== ========== ========== ==========
//test special name: 这,开,尖
//test adding standalone stmts to func main()
14:{`
包正
源
形Println("abcdefg") //added to func main(), with extra newline afterwards
种A整; //TOFIX: semi required here but shouldn't be
功开(){形Println("init running")}
功正(){
	a:=123
	形Println(整64(a))
	变b这A
	变c尖A
	形Println(b,c)
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`// +build ignore
package main

import fmt "fmt"

type A int

func init() { fmt.Println("init running") }
func main() {
	_a := 123
	fmt.Println(int64(_a))
	var _b A
	var _c *A
	fmt.Println(_b, _c)
	fmt.Println("abcdefg")

}
`}},

// ========== ========== ========== ==========
//test special names: 准,跑
//test multi-source use of 源 (with both specified and implied)
15:{`
包正
种A整;
功正(){
	a:=123
	形Println(整64(a))
}
源"beau"
种B整;
功hello(){
	b:=789
	形Println(整64(b))
	准"src/github.com/grolang/qutests/goByEg4.qu"
	跑"src/github.com/grolang/qutests/goByEg4.go"
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`// +build ignore
package main

import fmt "fmt"

type A int

func main() {
	_a := 123
	fmt.Println(int64(_a))
}
`,

// ---------- ----------
"beau.go": //FIX: should only have one "// +build ignore" comment
`// +build ignore
// +build ignore
package main

import (
	fmt "fmt"
	command "github.com/grolang/gro/macro/command"
)

type B int

func _hello() {
	_b := 789
	fmt.Println(int64(_b))
	command.FunPrep("src/github.com/grolang/qutests/goByEg4.qu")
	command.FunRun("src/github.com/grolang/qutests/goByEg4.go")

}
`}},

// ========== ========== ========== ==========
//test single standalone statement
16:{`形Println("Hello, world!")`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{
"suppliedName.go":
`// +build ignore
package main

import fmt "fmt"

func main() {
	fmt.Println("Hello, world!")
}
`}},

// ========== ========== ========== ==========
//test multi-package use of 包 and multi-source use of 源
17:{`
包正 //two explicit sources
源"tiful"
种A整;
功正(){
	a:=123
	形Println(整64(a))
}
源"beau"
种B整;
功hello(){
	b:=789
	形Println(整64(b))
}
包hi //one implied source
功hoorah(){
	c:=666
	形Println(整32(c))
}
包thinking //one implied and one explicit source/s
功mind(){
	e:= -12345
	形Println(整8(e))
}
源"psyche" //one explicit func and one implied init function
gg:=90909
形Println(整64(gg))
功psycheGo(){
	g:= 54321
	形Println(整8(g))
}
包bye //one explicit source
源"booyah"
功cherio(){
	d:= -99
	形Println(整16(d))
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"tiful.go":
`// +build ignore
package main

import fmt "fmt"

type A int

func main() {
	_a := 123
	fmt.Println(int64(_a))
}
`,

// ---------- ----------
"hi.go": //FIX: don't want "// +build ignore" here
`// +build ignore
package hi

import fmt "fmt"

func _hoorah() {
	_c := 666
	fmt.Println(int32(_c))
}
`,

// ---------- ----------
"thinking.go": //FIX: don't want "// +build ignore" here
`// +build ignore
package thinking

import fmt "fmt"

func _mind() {
	_e := -12345
	fmt.Println(int8(_e))
}
`,

// ---------- ----------
"psyche.go": //FIX: don't want "// +build ignore" here
`// +build ignore
package thinking

import fmt "fmt"

func _psycheGo() {
	_g := 54321
	fmt.Println(int8(_g))
}

func main() {
	_gg := 90909
	fmt.Println(int64(_gg))
}
`,

// ---------- ----------
"booyah.go": //FIX: don't want "// +build ignore" here
`// +build ignore
package bye

import fmt "fmt"

func _cherio() {
	_d := -99
	fmt.Println(int16(_d))
}
`,

// ---------- ----------
"beau.go":
`// +build ignore
package main

import fmt "fmt"

type B int

func _hello() {
	_b := 789
	fmt.Println(int64(_b))
}
`}},

// ========== ========== ========== ==========
//further test multi-package use of 包 and multi-source use of 源
18:{`
形Println("Hey, world!") //dangling stmts for pkg:main, dir:<curr>, src:suppliedName.go
包正 //two explicit src-names for pkg:main, dir:<curr>, srcs:tiful.go & beau.go
源"tiful"
种A整;
功正(){
	a:=123
	形Println(整64(a))
}
源"beau"
种B整;
功hello(){
	b:=789
	形Println(整64(b))
}
包"oneLib/hi" //one implied src-name with specified string for pkg:hi, dir:oneLib/hi, src:hi.go
功hoorah(){
	c:=666
	形Println(整32(c))
}
包"someDir"thinking //one implied and one explicit src-name with specified pkg-name and string for pkg:thinking, dir:someDir, srcs:thinking.go & psyche.go
功mind(){
	e:= -12345
	形Println(整8(e))
}
源"psyche"
功psycheGo(){
	g:= 54321
	形Println(整8(g))
}
包"lets/wave"bye //one explicit src-name with specified pkg-name and string for pkg:bye, dir:lets/wave, src:booyah.go
源"booyah"
功cherio(){
	d:= -99
	形Println(整16(d))
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"suppliedName.go":
`// +build ignore
package main

import fmt "fmt"

func main() {
	fmt.Println("Hey, world!")
}
`,

// ---------- ----------
"tiful.go":
`// +build ignore
package main

import fmt "fmt"

type A int

func main() {
	_a := 123
	fmt.Println(int64(_a))
}
`,

// ---------- ----------
"beau.go":
`// +build ignore
package main

import fmt "fmt"

type B int

func _hello() {
	_b := 789
	fmt.Println(int64(_b))
}
`,

// ---------- ----------
"oneLib/hi/hi.go": //FIX: don't want "// +build ignore" here
`// +build ignore
package hi

import fmt "fmt"

func _hoorah() {
	_c := 666
	fmt.Println(int32(_c))
}
`,

// ---------- ----------
"someDir/thinking.go": //FIX: don't want "// +build ignore" here
`// +build ignore
package thinking

import fmt "fmt"

func _mind() {
	_e := -12345
	fmt.Println(int8(_e))
}
`,

// ---------- ----------
"someDir/psyche.go": //FIX: don't want "// +build ignore" here
`// +build ignore
package thinking

import fmt "fmt"

func _psycheGo() {
	_g := 54321
	fmt.Println(int8(_g))
}
`,

// ---------- ----------
"lets/wave/booyah.go": //FIX: don't want "// +build ignore" here
`// +build ignore
package bye

import fmt "fmt"

func _cherio() {
	_d := -99
	fmt.Println(int16(_d))
}
`,

// ---------- ----------
}},

// ========== ========== ========== ==========
//further test special names: 准,跑
19:{`
包正
源"cmds"
功正(){
  准"src/github.com/grolang/qutests/goByEg4.qu"
  跑"src/github.com/grolang/qutests/goByEg4.go"
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"cmds.go":
`// +build ignore
package main

import command "github.com/grolang/gro/macro/command"

func main() {
	command.FunPrep("src/github.com/grolang/qutests/goByEg4.qu")
	command.FunRun("src/github.com/grolang/qutests/goByEg4.go")

}
`}},

// ========== ========== ========== ==========
//further test special names: 准,跑
20:{`
源"cmds"
功正(){
  准"src/github.com/grolang/qutests/goByEg4.qu"
  跑"src/github.com/grolang/qutests/goByEg4.go"
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"cmds.go":
`// +build ignore
package main

import command "github.com/grolang/gro/macro/command"

func main() {
	command.FunPrep("src/github.com/grolang/qutests/goByEg4.qu")
	command.FunRun("src/github.com/grolang/qutests/goByEg4.go")

}
`}},

// ========== ========== ========== ==========
//further test special names: 准,跑
21:{`
功正(){
  准"src/github.com/grolang/qutests/goByEg4.qu"
  跑"src/github.com/grolang/qutests/goByEg4.go"
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"suppliedName.go":
`// +build ignore
package main

import command "github.com/grolang/gro/macro/command"

func main() {
	command.FunPrep("src/github.com/grolang/qutests/goByEg4.qu")
	command.FunRun("src/github.com/grolang/qutests/goByEg4.go")

}
`}},

// ========== ========== ========== ==========
//further test special names: 准,跑
22:{`
准"src/github.com/grolang/qutests/goByEg4.qu"
跑"src/github.com/grolang/qutests/goByEg4.go"
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"suppliedName.go":
`// +build ignore
package main

import command "github.com/grolang/gro/macro/command"

func main() {
	command.FunPrep("src/github.com/grolang/qutests/goByEg4.qu")
	command.FunRun("src/github.com/grolang/qutests/goByEg4.go")
}
`}},

// ========== ========== ========== ==========
//further test special names: 准,跑
23:{`
//import "fmt"
功正(){
  准"src/github.com/grolang/qutests/goByEg4.qu"
  跑"src/github.com/grolang/qutests/goByEg4.go"
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"suppliedName.go":
`// +build ignore
package main

import command "github.com/grolang/gro/macro/command"

func main() {
	command.FunPrep("src/github.com/grolang/qutests/goByEg4.qu")
	command.FunRun("src/github.com/grolang/qutests/goByEg4.go")

}
`}},

// ========== ========== ========== ==========
//test special name: 叫
24:{`
a:="Hello, "+叫+"world!"
形Println(a)
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"suppliedName.go":
`// +build ignore
package main

import fmt "fmt"

func main() {
	a := "Hello, " + "" + "world!"
	fmt.Println(a)
}
`}},

// ========== ========== ========== ==========
//test Unihan punctuation “”‘’（）《》【】！；：，。

25:{`
入“time“
a：=“Hello, ”+叫+“world!”；b：=10
形Println（a，‘z’，1《2，2》1，！false，a【0】，b）
做time。Sleep(time。Second)
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"suppliedName.go": //FIX: should only have one "// +build ignore" comment
`// +build ignore
// +build ignore
package main

import (
	fmt "fmt"
	_time "time"
)

func main() {
	a := "Hello, " + "" + "world!"
	b := 10
	fmt.Println(a, 'z', 1 < 2, 2 > 1, !false, a[0], b)
	_time.Sleep(_time.Second)
}
`}},

// ========== ========== ========== ==========
//test "source" kw with specified filename
//test "do" kw with "do" ident
//test "do" label
26:{`package main
source "mySource"
func main() {
	a:= true
	b:= 真
	nil:= true
	iota:= 真
	形Printf("a: %v, b: %v, nil: %v, iota: %v\n", a, b, nil, iota)

	变abc图[串]整;
	做abc[串("def")]=789

	var _z= "abc"
	do do:="789"
	形Println(串(z)+do)

	do: for true {
		break do
	}
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"mySource.go":
`// +build ignore
package main

import fmt "fmt"

func main() {
	a := true
	b := true
	nil := true
	iota := true
	fmt.Printf("a: %v, b: %v, nil: %v, iota: %v\n", a, b, nil, iota)

	var _abc map[string]int
	_abc[string("def")] = 789

	var _z = "abc"
	do := "789"
	fmt.Println(string(_z) + do)

do:
	for true {
		break do
	}
}
`}},

// ========== ========== ========== ==========
//test using keyword "var" as id in 做 stmt
27:{`
func main(){
做var:="sass"
形Println(_var)
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"suppliedName.go":
`// +build ignore
package main

import fmt "fmt"

func main() {
	{
		_var := "sass"
		fmt.Println(_var)
	}
}
`}},

// ========== ========== ========== ==========
//test using keyword "var" as id in 变 stmt
28:{`
func main(){
变var="sass"
形Println(_var)
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"suppliedName.go":
`// +build ignore
package main

import fmt "fmt"

func main() {
	var _var = "sass"
	fmt.Println(_var)
}
`}},

// ========== ========== ========== ==========
//test using keyword "var" as id in 做 stmt, outside of explicit main fn
998:{`
做v:="sass"
形Println(v)
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"suppliedName.go":
`// +build ignore
package main

import "fmt"

func main() {
}
`}},

// ========== ========== ========== ==========
999:{`
package main
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{
}},

// ========== ========== ========== ==========
}

//================================================================================================================================================================
func TestGroSyntaxParseError(t *testing.T) {
	for i, tst:= range groSyntaxParseErrorTests {
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

		_, _, err := parser.ParseMultiFile(fset, "suppliedName.gro", src, nil, 0)
		if fmt.Sprintf("%v", err) != dse {
			t.Errorf("unexpected parse error in %d: received error: %q; expected error: %q", i, err, dse)
		}
	}
}

var groSyntaxParseErrorTests = map[int]struct{key string; val string} {

// ========== ========== ========== ==========
//test prohibited Unihan 假 use on LHS
1001:{`
package main
func main() {
	//a:= true
	//b:= 真
	//nil:= true
	//iota:= 真
	假:= true //this generates parse error "Unihan special identifier on left hand side"
}
`,

// ---------- ---------- ---------- ----------
`suppliedName.gro:8:5: Unihan special identifier 假 on left hand side (and 1 more errors)`},

// ========== ========== ========== ==========
//test prohibited Unihan 串 use on LHS
1002:{`
package main
func main() {
	串["def"]=789
}
`,

// ---------- ---------- ---------- ----------
`suppliedName.gro:4:12: Unihan special identifier 串 on left hand side (and 1 more errors)`},

// ========== ========== ========== ==========
1999:{`
package
`,

// ---------- ---------- ---------- ----------
`suppliedName.gro:2:9: expected ';', found 'EOF'`},

// ========== ========== ========== ==========
}

//================================================================================================================================================================
