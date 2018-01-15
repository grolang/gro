// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nodes_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/grolang/gro/nodes"
	"github.com/grolang/gro/syntax"
)

const colbase = 1 //taken from source.go

// A test is a source code snippet of a particular node type.
// In the snippet, a '@' indicates the position recorded by
// the parser when creating the respective node.
type test struct {
	nodetyp string
	snippet string
}

var decls = []test{
	// The position of declarations is always the
	// position of the first token of an individual
	// declaration, independent of grouping.
	{"*nodes.ImportDecl", `import @"math"`},
	{"*nodes.ImportDecl", `import @mymath "math"`},
	{"*nodes.ImportDecl", `import @. "math"`},
	{"*nodes.ImportDecl", `import (@"math")`},
	{"*nodes.ImportDecl", `import (@mymath "math")`},
	{"*nodes.ImportDecl", `import (@. "math")`},

	{"*nodes.ConstDecl", `const @x`},
	{"*nodes.ConstDecl", `const @x = 0`},
	{"*nodes.ConstDecl", `const @x, y, z = 0, 1, 2`},
	{"*nodes.ConstDecl", `const (@x)`},
	{"*nodes.ConstDecl", `const (@x = 0)`},
	{"*nodes.ConstDecl", `const (@x, y, z = 0, 1, 2)`},

	{"*nodes.TypeDecl", `type @T int`},
	{"*nodes.TypeDecl", `type @T = int`},
	{"*nodes.TypeDecl", `type (@T int)`},
	{"*nodes.TypeDecl", `type (@T = int)`},

	{"*nodes.VarDecl", `var @x int`},
	{"*nodes.VarDecl", `var @x, y, z int`},
	{"*nodes.VarDecl", `var @x int = 0`},
	{"*nodes.VarDecl", `var @x, y, z int = 1, 2, 3`},
	{"*nodes.VarDecl", `var @x = 0`},
	{"*nodes.VarDecl", `var @x, y, z = 1, 2, 3`},
	{"*nodes.VarDecl", `var (@x int)`},
	{"*nodes.VarDecl", `var (@x, y, z int)`},
	{"*nodes.VarDecl", `var (@x int = 0)`},
	{"*nodes.VarDecl", `var (@x, y, z int = 1, 2, 3)`},
	{"*nodes.VarDecl", `var (@x = 0)`},
	{"*nodes.VarDecl", `var (@x, y, z = 1, 2, 3)`},

	{"*nodes.FuncDecl", `func @f() {}`},
	{"*nodes.FuncDecl", `func @(T) f() {}`},
	{"*nodes.FuncDecl", `func @(x T) f() {}`},
}

var exprs = []test{
	// The position of an expression is the position
	// of the left-most token that identifies the
	// kind of expression.
	{"*nodes.Name", `@x`},

	{"*nodes.BasicLit", `@0`},
	{"*nodes.BasicLit", `@0x123`},
	{"*nodes.BasicLit", `@3.1415`},
	{"*nodes.BasicLit", `@.2718`},
	{"*nodes.BasicLit", `@1i`},
	{"*nodes.BasicLit", `@'a'`},
	{"*nodes.BasicLit", `@"abc"`},
	{"*nodes.BasicLit", "@`abc`"},

	{"*nodes.CompositeLit", `@{}`},
	{"*nodes.CompositeLit", `T@{}`},
	{"*nodes.CompositeLit", `struct{x, y int}@{}`},

	{"*nodes.KeyValueExpr", `"foo"@: true`},
	{"*nodes.KeyValueExpr", `"a"@: b`},

	{"*nodes.FuncLit", `@func (){}`},
	{"*nodes.ParenExpr", `@(x)`},
	{"*nodes.SelectorExpr", `a@.b`},
	{"*nodes.IndexExpr", `a@[i]`},

	{"*nodes.SliceExpr", `a@[:]`},
	{"*nodes.SliceExpr", `a@[i:]`},
	{"*nodes.SliceExpr", `a@[:j]`},
	{"*nodes.SliceExpr", `a@[i:j]`},
	{"*nodes.SliceExpr", `a@[i:j:k]`},

	{"*nodes.AssertExpr", `x@.(T)`},

	{"*nodes.Operation", `@*b`},
	{"*nodes.Operation", `@+b`},
	{"*nodes.Operation", `@-b`},
	{"*nodes.Operation", `@!b`},
	{"*nodes.Operation", `@^b`},
	{"*nodes.Operation", `@&b`},
	{"*nodes.Operation", `@<-b`},

	{"*nodes.Operation", `a @|| b`},
	{"*nodes.Operation", `a @&& b`},
	{"*nodes.Operation", `a @== b`},
	{"*nodes.Operation", `a @+ b`},
	{"*nodes.Operation", `a @* b`},

	{"*nodes.CallExpr", `f@()`},
	{"*nodes.CallExpr", `f@(x, y, z)`},
	{"*nodes.CallExpr", `obj.f@(1, 2, 3)`},
	{"*nodes.CallExpr", `func(x int) int { return x + 1 }@(y)`},

	// ListExpr: tested via multi-value const/var declarations
}

var types = []test{
	{"*nodes.Operation", `@*T`},
	{"*nodes.Operation", `@*struct{}`},

	{"*nodes.ArrayType", `@[10]T`},
	{"*nodes.ArrayType", `@[...]T`},

	{"*nodes.SliceType", `@[]T`},
	{"*nodes.DotsType", `@...T`},
	{"*nodes.StructType", `@struct{}`},
	{"*nodes.InterfaceType", `@interface{}`},
	{"*nodes.FuncType", `func@()`},
	{"*nodes.MapType", `@map[T]T`},

	{"*nodes.ChanType", `@chan T`},
	{"*nodes.ChanType", `@chan<- T`},
	{"*nodes.ChanType", `@<-chan T`},
}

var fields = []test{
	{"*nodes.Field", `@T`},
	{"*nodes.Field", `@(T)`},
	{"*nodes.Field", `@x T`},
	{"*nodes.Field", `@x *(T)`},
	{"*nodes.Field", `@x, y, z T`},
	{"*nodes.Field", `@x, y, z (*T)`},
}

var stmts = []test{
	{"*nodes.EmptyStmt", `@`},

	{"*nodes.LabeledStmt", `L@:`},
	{"*nodes.LabeledStmt", `L@: ;`},
	{"*nodes.LabeledStmt", `L@: f()`},

	{"*nodes.BlockStmt", `@{}`},

	// The position of an ExprStmt is the position of the expression.
	{"*nodes.ExprStmt", `@<-ch`},
	{"*nodes.ExprStmt", `f@()`},
	{"*nodes.ExprStmt", `append@(s, 1, 2, 3)`},

	{"*nodes.SendStmt", `ch @<- x`},

	{"*nodes.DeclStmt", `@const x = 0`},
	{"*nodes.DeclStmt", `@const (x = 0)`},
	{"*nodes.DeclStmt", `@type T int`},
	{"*nodes.DeclStmt", `@type T = int`},
	{"*nodes.DeclStmt", `@type (T1 = int; T2 = float32)`},
	{"*nodes.DeclStmt", `@var x = 0`},
	{"*nodes.DeclStmt", `@var x, y, z int`},
	{"*nodes.DeclStmt", `@var (a, b = 1, 2)`},

	{"*nodes.AssignStmt", `x @= y`},
	{"*nodes.AssignStmt", `a, b, x @= 1, 2, 3`},
	{"*nodes.AssignStmt", `x @+= y`},
	{"*nodes.AssignStmt", `x @:= y`},
	{"*nodes.AssignStmt", `x, ok @:= f()`},
	{"*nodes.AssignStmt", `x@++`},
	{"*nodes.AssignStmt", `a[i]@--`},

	{"*nodes.BranchStmt", `@break`},
	{"*nodes.BranchStmt", `@break L`},
	{"*nodes.BranchStmt", `@continue`},
	{"*nodes.BranchStmt", `@continue L`},
	{"*nodes.BranchStmt", `@fallthrough`},
	{"*nodes.BranchStmt", `@goto L`},

	{"*nodes.CallStmt", `@defer f()`},
	{"*nodes.CallStmt", `@go f()`},

	{"*nodes.ReturnStmt", `@return`},
	{"*nodes.ReturnStmt", `@return x`},
	{"*nodes.ReturnStmt", `@return a, b, a + b*f(1, 2, 3)`},

	{"*nodes.IfStmt", `@if cond {}`},
	{"*nodes.IfStmt", `@if cond { f() } else {}`},
	{"*nodes.IfStmt", `@if cond { f() } else { g(); h() }`},
	{"*nodes.ForStmt", `@for {}`},
	{"*nodes.ForStmt", `@for { f() }`},
	{"*nodes.SwitchStmt", `@switch {}`},
	{"*nodes.SwitchStmt", `@switch { default: }`},
	{"*nodes.SwitchStmt", `@switch { default: x++ }`},
	{"*nodes.SelectStmt", `@select {}`},
	{"*nodes.SelectStmt", `@select { default: }`},
	{"*nodes.SelectStmt", `@select { default: ch <- false }`},
}

var ranges = []test{
	{"*nodes.RangeClause", `@range s`},
	{"*nodes.RangeClause", `i = @range s`},
	{"*nodes.RangeClause", `i := @range s`},
	{"*nodes.RangeClause", `_, x = @range s`},
	{"*nodes.RangeClause", `i, x = @range s`},
	{"*nodes.RangeClause", `_, x := @range s.f`},
	{"*nodes.RangeClause", `i, x := @range f(i)`},
}

var guards = []test{
	{"*nodes.TypeSwitchGuard", `x@.(type)`},
	{"*nodes.TypeSwitchGuard", `x := x@.(type)`},
}

var cases = []test{
	{"*nodes.CaseClause", `@case x:`},
	{"*nodes.CaseClause", `@case x, y, z:`},
	{"*nodes.CaseClause", `@case x == 1, y == 2:`},
	{"*nodes.CaseClause", `@default:`},
}

var comms = []test{
	{"*nodes.CommClause", `@case <-ch:`},
	{"*nodes.CommClause", `@case x <- ch:`},
	{"*nodes.CommClause", `@case x = <-ch:`},
	{"*nodes.CommClause", `@case x := <-ch:`},
	{"*nodes.CommClause", `@case x, ok = <-ch: f(1, 2, 3)`},
	{"*nodes.CommClause", `@case x, ok := <-ch: x++`},
	{"*nodes.CommClause", `@default:`},
	{"*nodes.CommClause", `@default: ch <- true`},
}

func TestPos(t *testing.T) {
	// TODO(gri) Once we have a general tree walker, we can use that to find
	// the first occurrence of the respective node and we don't need to hand-
	// extract the node for each specific kind of construct.

	testPos(t, decls, "package p; ", "",
		func(f *File) Node { return f.DeclList[0] },
	)

	// embed expressions in a composite literal so we can test key:value and naked composite literals
	testPos(t, exprs, "package p; var _ = T{ ", " }",
		func(f *File) Node { return f.DeclList[0].(*VarDecl).Values.(*CompositeLit).ElemList[0] },
	)

	// embed types in a function  signature so we can test ... types
	testPos(t, types, "package p; func f(", ")",
		func(f *File) Node { return f.DeclList[0].(*FuncDecl).Type.ParamList[0].Type },
	)

	testPos(t, fields, "package p; func f(", ")",
		func(f *File) Node { return f.DeclList[0].(*FuncDecl).Type.ParamList[0] },
	)

	testPos(t, stmts, "package p; func _() { ", "; }",
		func(f *File) Node { return f.DeclList[0].(*FuncDecl).Body.List[0] },
	)

	testPos(t, ranges, "package p; func _() { for ", " {} }",
		func(f *File) Node { return f.DeclList[0].(*FuncDecl).Body.List[0].(*ForStmt).Init.(*RangeClause) },
	)

	testPos(t, guards, "package p; func _() { switch ", " {} }",
		func(f *File) Node { return f.DeclList[0].(*FuncDecl).Body.List[0].(*SwitchStmt).Tag.(*TypeSwitchGuard) },
	)

	testPos(t, cases, "package p; func _() { switch { ", " } }",
		func(f *File) Node { return f.DeclList[0].(*FuncDecl).Body.List[0].(*SwitchStmt).Body[0] },
	)

	testPos(t, comms, "package p; func _() { select { ", " } }",
		func(f *File) Node { return f.DeclList[0].(*FuncDecl).Body.List[0].(*SelectStmt).Body[0] },
	)
}

func testPos(t *testing.T, list []test, prefix, suffix string, extract func(*File) Node) {
	for _, test := range list {
		// complete source, compute @ position, and strip @ from source
		src, index := stripAt(prefix + test.snippet + suffix)
		if index < 0 {
			t.Errorf("missing @: %s (%s)", src, test.nodetyp)
			continue
		}

		// build syntax tree
		asts, err := syntax.ParseBytes("dud", nil, []byte(src), nil, nil, 0, nil)
		if err != nil {
			t.Errorf("parse error: %s: %v (%s)", src, err, test.nodetyp)
			continue
		}
		if len(asts) != 1 {
			t.Error("More than one file returned from parse.")
		}
		for _, file := range asts {
			// extract desired node
			node := extract(file)
			if typ := typeOf(node); typ != test.nodetyp {
				t.Errorf("type error: %s: type = %s, want %s", src, typ, test.nodetyp)
				continue
			}

			// verify node position with expected position as indicated by @
			if pos := int(node.Pos().Col()); pos != index+colbase {
				t.Errorf("pos error: %s: pos = %d, want %d (%s)", src, pos, index+colbase, test.nodetyp)
				continue
			}
		}
	}
}

func stripAt(s string) (string, int) {
	if i := strings.Index(s, "@"); i >= 0 {
		return s[:i] + s[i+1:], i
	}
	return s, -1
}

func typeOf(n Node) string {
	const prefix = "*syntax."
	k := fmt.Sprintf("%T", n)
	if strings.HasPrefix(k, prefix) {
		return k[len(prefix):]
	}
	return k
}
