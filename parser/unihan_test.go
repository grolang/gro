// Copyright 2016 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package parser_test

import (
	"bytes"
	"fmt"
	"github.com/grolang/gro/parser"
	"go/ast"
	"go/format"
	"go/importer"
	"github.com/grolang/gro/printer"
	"go/token"
	"go/types"
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
		fm, frag, err := parser.ParseMultiFile(fset, "", src, nil, 0)
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
//test keywords: 功
//test keyword scoping: 入
1:{`
package main;入"fmt"
功main(){
  fmt.Printf("Hi!\n") // comment here
}`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

import _fmt "fmt"

func _main() {
	_fmt.Printf("Hi!\n")
}
`,
"dud":`
/*2:1*/package /*2:9*/main

/*2:14*/import /*2:17*/_fmt /*2:17*/"fmt"

/*3:1*/func /*3:4*/_main/*3:8*/(/*3:9*/) /*3:10*/{
        /*4:3*/_fmt/*4:7*/./*4:7*/Printf/*4:13*/(/*4:14*/"Hi!\n"/*4:21*/)
/*5:1*/}
`},""},

// ========== ========== ========== ==========
//test keyword: 回
//test keyword scoping: 功
//test specids: 度,串,整,整64,漂32,漂64,复,复64,复128
2:{`
package main;入"fmt";
功main(){
  fmt.Printf("Len: %d\n", 度(fs("abcdefg")))
}
功fs(a串)串{回a+"xyz"}
功ff(a漂32)漂64{回漂64(a)}
功fc(a复64)复128{回复128(a)+复(1,1)}
功fi(a整)整64{回整64(a)}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

import _fmt "fmt"

func _main() {
	_fmt.Printf("Len: %d\n", len(_fs("abcdefg")))
}
func _fs(_a string) string        { return _a + "xyz" }
func _ff(_a float32) float64      { return float64(_a) }
func _fc(_a complex64) complex128 { return complex128(_a) + complex(1, 1) }
func _fi(_a int) int64            { return int64(_a) }
`},""},

// ========== ========== ========== ==========
//test keyword scoping: 变,如,否
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
  fr,_:= _utf8.DecodeRune([]byte("lmnop"))
  fmt.Printf("1st rune: %s; Len: %d\n", fr, len("lmnop") + n)
  让_,_=kl.DecodeRune([]节("lmnop"))
  㕤Printf("Fifty: %d\n", n)
  哪Printf("Fifty: %d\n", n)
  吧Printf("Fifty: %d\n", n)
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

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
`},""},

// ========== ========== ========== ==========
//test keyword: 构
//test keyword scoping: 种,久
//test specids: 整8,整16,整32
4:{`
package main
type _string string
种A struct{a string; b 整8}
种B struct{a string; b 整16}
种C构{a string; b 整32}
type D struct{a string; b 整32}
种E构{a串;b整32}
久a=3.1416
const b=2.72
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

type _string string
type A struct {
	_a _string
	_b int8
}
type B struct {
	_a _string
	_b int16
}
type C struct {
	_a _string
	_b int32
}
type D struct {
	a string
	b int32
}
type E struct {
	_a string
	_b int32
}

const _a = 3.1416
const b = 2.72
`},""},

// ========== ========== ========== ==========
//test keywords: 围,为,继,破
//test keyword scoping: 入,图
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
	fmt.Println("abc")
	为i:=围a{fmt.Print(i," ")}
	for i:= 0; i<28; i++ {
		if i==3 { continue }
		if i==6 { break }
		fmt.Print(i, " ")
	}
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

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
`},""},

// ========== ========== ========== ==========
//test keywords: 掉
//test keyword scoping: 考,事,别,面
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
	_a:=2
	考a{
	事1:
		fmt.Print('a');
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
`package main

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
`},""},

// ========== ========== ========== ==========
//test keyword scoping: 选,去,通
//test special identifier: 正
7:{`包正;入("math/rand";"sync/atomic")
种readOp构{key整;resp通整}
种writeOp构{key整;val整;resp通双}
功正(){
    变ops整64=0
    reads:=造(通*readOp)
    writes:=造(通*writeOp)
    去功(){
        变state=造(图[整]整)
        为{选{事read:=<-reads:read.resp<-state[read.key]
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

    让range:="abc" //when used with 让, Go keywords like "range" can be used as identifiers
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

	_range := "abc"
	_range = "abcdefg"
	fmt.Printf("range: %v\n", _range)
}
`},""},

// ========== ========== ========== ==========
//test keyword scoping: 围
//test special keyword: 做
8:{`package main

import (
	//"fmt"
	"github.com/grolang/gro/parser"
	"go/token"
	"go/format"
	//"go/ast"
	//"os"
)

func main() {
	形Printf("Hi!\n")

	fset := token.NewFileSet() // positions are relative to fset

	_f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, s := 围 f.Imports {
		fmt.Println(s.Path.Value)
	}

	{	sw:= StringWriter{""}
		_= format.Node(&sw, fset, _f)
		fmt.Println(sw.data)
		fmt.Println(sw.data == dst)
	}
	做{ abc:= "abc"
		_ = abc
	}
}

type StringWriter struct{ data string }
func (sw *StringWriter) Write(b []byte) (int, error) {
	sw.data += string(b)
	return len(b), nil
}

var src string = "package main"
var dst string = "package main"`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

import (
	fmt "fmt"
	"github.com/grolang/gro/parser"
	"go/format"
	"go/token"
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
`},""},

// ========== ========== ========== ==========
//test keyword scoping: 包
//test special keywords: 让,任
//test using keyword as id in Unihan-context (both LHS and RHS)
9:{`
package main;入"fmt"
import "fmt"
const a=6
功main(){
  让b:=7
  让func:=8
  fmt.Printf("Hi, nos.%s and %s!\n", b, func)
  英{
    fmt.Printf("Hi, no. %s.\n", a)
  }
}
func baba(){
  变b任= 17
  fmt.Printf("Hi, no.%s!\n", _b)
}`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

import _fmt "fmt"
import "fmt"

const a = 6

func _main() {
	_b := 7
	_func := 8
	_fmt.Printf("Hi, nos.%s and %s!\n", _b, _func)
	{
		fmt.Printf("Hi, no. %s.\n", a)
	}
}
func baba() {
	var _b interface{} = 17
	fmt.Printf("Hi, no.%s!\n", _b)
}
`},""},

// ========== ========== ========== ==========
10:{`
包main;源;入"fmt"
功main(){
  让b:=7
  让func:=8
  fmt.Printf("Hi, nos.%s and %s!\n", b, func) // different comment here
}`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

import _fmt "fmt"

func _main() {
	_b := 7
	_func := 8
	_fmt.Printf("Hi, nos.%s and %s!\n", _b, _func)
}
`},""},

// ========== ========== ========== ==========
//test prohibited Unihan use on LHS
11:{`package main
源"mySource.go"
func main() {
	a:= true
	b:= 真
	nil:= true
	iota:= 真
	//假:= true //parse error
	形Printf("a: %v, b: %v, nil: %v, iota: %v\n", a, b, nil, iota)

	变abc图[串]整;
	让abc[串("def")]=789

	var _z= "abc"
	形Println(串(z))
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"mySource.go":
`package main

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
`},""},

// ========== ========== ========== ==========
12:{`
package main
type A int
func main(){
  _a:=123
  形Println(整64(a))
  var b A
  var c这A
  形Println(b, c)
  鲜d:='d'
  形Println(a,d)
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

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
`},""},

// ========== ========== ========== ==========
13:{`包正;种A整;功正(){a:=123;形Println(整64(a));变b这A;变c这A;形Println(b,c)}`, //example of totally spaceless code

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

import fmt "fmt"

type A int

func main() { _a := 123; fmt.Println(int64(_a)); var _b A; var _c A; fmt.Println(_b, _c) }
`},""},

// ========== ========== ========== ==========
14:{`
包正
源
形Println("abcdefg") //added to func main(), with extra newline afterwards
种A整; //TOFIX: semi required here but shouldn't be
功正(){
	a:=123
	形Println(整64(a))
	变b这A
	变c这A
	形Println(b,c)
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

import fmt "fmt"

type A int

func main() {
	_a := 123
	fmt.Println(int64(_a))
	var _b A
	var _c A
	fmt.Println(_b, _c)
	fmt.Println("abcdefg")

}
`},""},

// ========== ========== ========== ==========
15:{`
包正
种A整;
功正(){
	a:=123
	形Println(整64(a))
}
源"beau.go"
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
`package main

import fmt "fmt"

type A int

func main() {
	_a := 123
	fmt.Println(int64(_a))
}
`,

// ---------- ----------
"beau.go":
`package main

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
`},""},

// ========== ========== ========== ==========
16:{`形Println("Hello, world!")`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{
"main.go":
`package main

import fmt "fmt"

func main() {
	fmt.Println("Hello, world!")
}
`},""},

// ========== ========== ========== ==========
17:{`
包正 //two explicit sources
源"tiful.go"
种A整;
功正(){
	a:=123
	形Println(整64(a))
}
源"beau.go"
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
源"psyche.go"
功psycheGo(){
	g:= 54321
	形Println(整8(g))
}
包bye //one explicit source
源"booyah.go"
功cherio(){
	d:= -99
	形Println(整16(d))
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"tiful.go":
`package main

import fmt "fmt"

type A int

func main() {
	_a := 123
	fmt.Println(int64(_a))
}
`,

// ---------- ----------
"hi.go":
"package hi\n\nimport fmt \"fmt\"\n\nfunc _hoorah() {\n\t_c := 666\n\tfmt.Println(int32(_c))\n}\n",

// ---------- ----------
"thinking.go":
"package thinking\n\nimport fmt \"fmt\"\n\nfunc _mind() {\n\t_e := -12345\n\tfmt.Println(int8(_e))\n}\n",

// ---------- ----------
"psyche.go":
"package thinking\n\nimport fmt \"fmt\"\n\nfunc _psycheGo() {\n\t_g := 54321\n\tfmt.Println(int8(_g))\n}\n",

// ---------- ----------
"booyah.go":
"package bye\n\nimport fmt \"fmt\"\n\nfunc _cherio() {\n\t_d := -99\n\tfmt.Println(int16(_d))\n}\n",

// ---------- ----------
"beau.go":
`package main

import fmt "fmt"

type B int

func _hello() {
	_b := 789
	fmt.Println(int64(_b))
}
`},""},

// ========== ========== ========== ==========
18:{`
形Println("Hey, world!") //dangling stmts for pkg:main, dir:<curr>, src:main.go
包正 //two explicit src-names for pkg:main, dir:<curr>, srcs:tiful.go & beau.go
源"tiful.go"
种A整;
功正(){
	a:=123
	形Println(整64(a))
}
源"beau.go"
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
源"psyche.go"
功psycheGo(){
	g:= 54321
	形Println(整8(g))
}
包"lets/wave"bye //one explicit src-name with specified pkg-name and string for pkg:bye, dir:lets/wave, src:booyah.go
源"booyah.go"
功cherio(){
	d:= -99
	形Println(整16(d))
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

import fmt "fmt"

func main() {
	fmt.Println("Hey, world!")
}
`,

// ---------- ----------
"tiful.go":
`package main

import fmt "fmt"

type A int

func main() {
	_a := 123
	fmt.Println(int64(_a))
}
`,

// ---------- ----------
"beau.go":
`package main

import fmt "fmt"

type B int

func _hello() {
	_b := 789
	fmt.Println(int64(_b))
}
`,

// ---------- ----------
"oneLib/hi/hi.go":
"package hi\n\nimport fmt \"fmt\"\n\nfunc _hoorah() {\n\t_c := 666\n\tfmt.Println(int32(_c))\n}\n",

// ---------- ----------
"someDir/thinking.go":
"package thinking\n\nimport fmt \"fmt\"\n\nfunc _mind() {\n\t_e := -12345\n\tfmt.Println(int8(_e))\n}\n",

// ---------- ----------
"someDir/psyche.go":
"package thinking\n\nimport fmt \"fmt\"\n\nfunc _psycheGo() {\n\t_g := 54321\n\tfmt.Println(int8(_g))\n}\n",

// ---------- ----------
"lets/wave/booyah.go":
"package bye\n\nimport fmt \"fmt\"\n\nfunc _cherio() {\n\t_d := -99\n\tfmt.Println(int16(_d))\n}\n",

// ---------- ----------
},""},

// ========== ========== ========== ==========
19:{`
包正
源"cmds.go"
功正(){
  准"src/github.com/grolang/qutests/goByEg4.qu"
  跑"src/github.com/grolang/qutests/goByEg4.go"
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"cmds.go":
`package main

import command "github.com/grolang/gro/macro/command"

func main() {
	command.FunPrep("src/github.com/grolang/qutests/goByEg4.qu")
	command.FunRun("src/github.com/grolang/qutests/goByEg4.go")

}
`},""},

// ========== ========== ========== ==========
20:{`
源"cmds.go"
功正(){
  准"src/github.com/grolang/qutests/goByEg4.qu"
  跑"src/github.com/grolang/qutests/goByEg4.go"
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"cmds.go":
`package main

import command "github.com/grolang/gro/macro/command"

func main() {
	command.FunPrep("src/github.com/grolang/qutests/goByEg4.qu")
	command.FunRun("src/github.com/grolang/qutests/goByEg4.go")

}
`},""},

// ========== ========== ========== ==========
21:{`
功正(){
  准"src/github.com/grolang/qutests/goByEg4.qu"
  跑"src/github.com/grolang/qutests/goByEg4.go"
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

import command "github.com/grolang/gro/macro/command"

func main() {
	command.FunPrep("src/github.com/grolang/qutests/goByEg4.qu")
	command.FunRun("src/github.com/grolang/qutests/goByEg4.go")

}
`},""},

// ========== ========== ========== ==========
22:{`
准"src/github.com/grolang/qutests/goByEg4.qu"
跑"src/github.com/grolang/qutests/goByEg4.go"
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

import command "github.com/grolang/gro/macro/command"

func main() {
	command.FunPrep("src/github.com/grolang/qutests/goByEg4.qu")
	command.FunRun("src/github.com/grolang/qutests/goByEg4.go")
}
`,

"dud":`/*1:1*/package /*1:8*/main

/*1:1*/import /*1:1*/(
        /*1:1*/exec /*1:1*/"os/exec"
        /*1:1*/fmt /*1:1*/"fmt"
/*1:1*/)

/*1:2*/func /*1:6*/main/*1:10*/(/*1:11*/) /*1:12*/{
        /*1:13*/func/*1:17*/(/*1:18*/) /*1:19*/{
                /*1:20*/x/*1:21*/, /*1:22*/_ /*1:23*/:= /*1:25*/exec/*1:29*/./*1:30*/Command/*1:37*/(/*1:38*/"gro"/*1:43*/, /*1:44*/"src/github.com/grolang/qutests/goByEg4.qu"/*1:87*/)/*1:88*/./*1:89*/Output/*1:95*/(/*1:96*/)
                /*1:97*/fmt/*1:100*/./*1:101*/Println/*1:108*/(/*1:109*/string/*1:115*/(/*1:116*/x/*1:117*/)/*1:118*/)
        /*1:119*/}/*1:120*/(/*1:121*/)
        /*1:122*/func/*1:126*/(/*1:127*/) /*1:128*/{
                /*1:129*/x/*1:130*/, /*1:131*/_ /*1:132*/:= /*1:134*/exec/*1:138*/./*1:139*/Command/*1:146*/(/*1:147*/"go"/*1:151*/, /*1:152*/"run"/*1:157*/, /*1:158*/"src/github.com/grolang/qutests/goByEg4.go"/*1:201*/)/*1:202*/./*1:203*/Output/*1:209*/(/*1:210*/)
                /*1:211*/fmt/*1:214*/./*1:215*/Println/*1:222*/(/*1:223*/string/*1:229*/(/*1:230*/x/*1:231*/)/*1:232*/)
        /*1:233*/}/*1:234*/(/*1:235*/)
/*1:236*/}
`},""},

// ========== ========== ========== ==========
23:{`
//import "fmt"
功正(){
  准"src/github.com/grolang/qutests/goByEg4.qu"
  跑"src/github.com/grolang/qutests/goByEg4.go"
}
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

import command "github.com/grolang/gro/macro/command"

func main() {
	command.FunPrep("src/github.com/grolang/qutests/goByEg4.qu")
	command.FunRun("src/github.com/grolang/qutests/goByEg4.go")

}
`,

"dud":`
/*1:1*/package /*1:8*/main

/*2:1*/import /*2:8*/(
        /*2:8*/"fmt"
        /*2:8*/exec /*2:8*/"os/exec"
/*2:8*/)

/*3:1*/func /*3:4*/main/*3:7*/(/*3:8*/) /*3:9*/{
        /*3:10*/func/*3:14*/(/*3:15*/) /*3:16*/{
                /*3:17*/x/*3:18*/, /*3:19*/_ /*3:20*/:= /*3:22*/exec/*3:26*/./*3:27*/Command/*3:34*/(/*3:35*/"gro"/*3:40*/, /*3:41*/"src/github.com/grolang/qutests/goByEg4.qu"/*3:84*/)/*3:85*/./*3:86*/Output/*3:92*/(/*3:93*/)
                /*3:94*/fmt/*3:97*/./*3:98*/Println/*3:105*/(/*3:106*/string/*3:112*/(/*3:113*/x/*3:114*/)/*3:115*/)
        /*3:116*/}/*3:117*/(/*3:118*/)
        /*3:119*/func/*3:123*/(/*3:124*/) /*3:125*/{
                /*3:126*/x/*3:127*/, /*3:128*/_ /*3:129*/:= /*3:131*/exec/*3:135*/./*3:136*/Command/*3:143*/(/*3:144*/"go"/*3:148*/, /*3:149*/"run"/*3:154*/, /*3:155*/"src/github.com/grolang/qutests/goByEg4.go"/*3:198*/)/*3:199*/./*3:200*/Output/*3:206*/(/*3:207*/)
                /*3:208*/fmt/*3:211*/./*3:212*/Println/*3:219*/(/*3:220*/string/*3:226*/(/*3:227*/x/*3:228*/)/*3:229*/)
        /*3:230*/}/*3:231*/(/*3:232*/)

/*6:1*/}
`},""},

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
26:{`
a:="Hello, "+叫+"world!"
形Println(a)
`,

// ---------- ---------- ---------- ----------
Normal,
map[string]string{"main.go":
`package main

import fmt "fmt"

func main() {
	a := "Hello, " + "" + "world!"
	fmt.Println(a)
}
`},""},

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

		_, _, err := parser.ParseMultiFile(fset, "", src, nil, 0)
		if fmt.Sprintf("%v", err) != dse {
			t.Errorf("unexpected parse error in %d: received error: %q; expected error: %q", i, err, dse)
		}
	}
}

var unihanParseErrorTests = map[int]struct{key string; val string} {

// ========== ========== ========== ==========
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
`8:5: Unihan special identifier 假 on left hand side (and 1 more errors)`},

// ========== ========== ========== ==========
1002:{`
package main
func main() {
	串["def"]=789
}
`,

// ---------- ---------- ---------- ----------
`4:12: Unihan special identifier 串 on left hand side (and 1 more errors)`},

// ========== ========== ========== ==========
1999:{`
package
`,

// ---------- ---------- ---------- ----------
`2:9: expected ';', found 'EOF'`},

// ========== ========== ========== ==========
}

//================================================================================================================================================================
