// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package macros

import (
	"fmt"

	"github.com/grolang/gro/nodes"
)

var (
	assertLib = "\"github.com/grolang/gro/assert\""
	sysLib    = "\"github.com/grolang/gro/sys\""
)

//--------------------------------------------------------------------------------
func InitGenerics(p nodes.GeneralParser, rets []string, args []string) {
	if len(rets) > 1 {
		p.SyntaxError("use \"generics\" has too many return values")
		p.Advance(nodes.SemiT, nodes.RbraceT)
		return
	}
	if len(args) > 1 {
		p.SyntaxError("use \"generics\" has too many arguments")
		p.Advance(nodes.SemiT, nodes.RbraceT)
		return
	}
	p.SetPermit("genericCall")
	p.SetPermit("genericDef")
}

//--------------------------------------------------------------------------------
func InitDynamic(p nodes.GeneralParser, rets []string, args []string) {
	if len(rets) > 1 {
		p.SyntaxError("use \"dynamic\" has too many return values")
		p.Advance(nodes.SemiT, nodes.RbraceT)
		return
	}
	if len(args) > 1 {
		p.SyntaxError("use \"dynamic\" has too many arguments")
		p.Advance(nodes.SemiT, nodes.RbraceT)
		return
	}
	p.SetDynamicMode(true)
	p.SetDynCharSet("utf8")
	for _, arg := range args {
		switch arg {
		case "utf88":
			p.SetDynCharSet("utf88")
		default:
			p.SyntaxError("\"dynamic\" doesn't support argument \"" + args[0] + "\"")
			p.Advance(nodes.SemiT, nodes.RbraceT)
			return
		}
	}
	if len(rets) == 1 {
		ret := rets[0]
		if ret == "_" {
			p.SetDynamicBlock("groo")
		} else {
			p.SetStmtRegistry(ret, func(p nodes.GeneralParser, _ ...interface{}) nodes.Stmt {
				dynBlock := p.DynamicBlock()
				p.SetDynamicBlock(ret)
				bs := p.BlockStmt("", p.TlStmt)
				p.SetDynamicBlock(dynBlock)
				return bs
			})
		}
	}
}

//--------------------------------------------------------------------------------
func GroSystemCmd(p nodes.GeneralParser, s string) nodes.Stmt {
	if !p.IsPermit(s) {
		p.SyntaxError(fmt.Sprintf("\"%s\" command disabled but is present", s))
		return nil
	}
	switch s {
	case "prepare":
		s = "Prepare"
	case "execute":
		s = "Execute"
	case "run":
		s = "Run"
	case "test":
		s = "Test"
	}
	fn := p.OLiteral()
	if fn == nil || fn.Kind != nodes.StringLit {
		p.SyntaxError("missing filename for " + s)
		p.Advance(nodes.SemiT, nodes.RparenT)
		return nil
	} else {
		a := p.ProcImportAlias(&nodes.BasicLit{Value: sysLib, Kind: nodes.StringLit}, "")
		es := &nodes.ExprStmt{
			X: &nodes.CallExpr{
				Fun: &nodes.SelectorExpr{
					X:   &nodes.Name{Value: a},
					Sel: &nodes.Name{Value: s},
				},
				ArgList: []nodes.Expr{fn},
			},
		}
		return es
	}
}

//--------------------------------------------------------------------------------
func Assert(p nodes.GeneralParser) nodes.Stmt {
	if !p.IsPermit("assert") {
		p.SyntaxError("\"assert\" macro disabled but is present")
		return nil
	}

	if p.Tok() == nodes.LparenT {
		fl := p.NewBlankFuncLit()
		p.List(nodes.LparenT, nodes.SemiT, nodes.RparenT, func() bool {
			e := p.Expr()
			fes := &nodes.ExprStmt{
				X: &nodes.CallExpr{
					Fun: &nodes.SelectorExpr{
						X:   &nodes.Name{Value: "assert"},
						Sel: &nodes.Name{Value: "AssertTrue"},
					},
					ArgList: []nodes.Expr{e},
				},
			}
			fl.Body.List = append(fl.Body.List, fes)
			return false
		})
		p.ProcImportAlias(&nodes.BasicLit{Value: assertLib, Kind: nodes.StringLit}, "assert")
		es := &nodes.ExprStmt{
			X: &nodes.CallExpr{
				Fun:     fl,
				ArgList: []nodes.Expr{},
			},
		}
		return es
	} else {
		e := p.Expr()
		p.ProcImportAlias(&nodes.BasicLit{Value: assertLib, Kind: nodes.StringLit}, "assert")
		es := &nodes.ExprStmt{
			X: &nodes.CallExpr{
				Fun: &nodes.SelectorExpr{
					X:   &nodes.Name{Value: "assert"},
					Sel: &nodes.Name{Value: "AssertTrue"},
				},
				ArgList: []nodes.Expr{e},
			},
		}
		return es
	}
}

//--------------------------------------------------------------------------------
func Let(p nodes.GeneralParser, stmt func() nodes.Stmt) nodes.Stmt {
	if !p.IsPermit("let") {
		p.SyntaxError("\"let\" macro disabled but is present")
		return nil
	}
	pos := p.Pos()
	lhs := p.ExprList(false)
	op := nodes.Def
	if nameNode, isName := lhs.(*nodes.Name); isName && nameNode.Value == "_" {
		op = 0
	}
	if p.Tok() == nodes.AssignT {
		// expr_list '=' expr_list
		p.Next()
		l := []nodes.Stmt{
			p.NewAssignStmt(pos, op, lhs, p.ExprList(true)),
		}
		p.Got(nodes.SemiT)
		l = append(l, p.TlStmtList(stmt)...)
		return &nodes.BlockStmt{
			List: l,
		}
	}
	p.SyntaxError("invalid syntax in \"let\" statement")
	return nil
}

//--------------------------------------------------------------------------------
func InitLineDirectives(p nodes.GeneralParser, rets, args []string) {
	if len(rets) != 0 {
		p.SyntaxError("use \"linedirectives\" shouldn't return any names but does")
		return
	}
	if len(args) != 0 {
		p.SyntaxError("use \"linedirectives\" shouldn't take any arguments but does")
		return
	}
	p.SetLineDirectives(true)
}

//--------------------------------------------------------------------------------
func InitBlacklist(p nodes.GeneralParser, rets, args []string) {
	if len(rets) != 0 {
		p.SyntaxError("use \"blacklist\" shouldn't return any names but does")
		return
	}
	for _, s := range args {
		switch s {
		default:
			p.UnsetPermit(s)
		case "package":
			p.UnsetPermit("package")
			p.UnsetPermit("internal")
		case "section":
			p.UnsetPermit("section")
			p.UnsetPermit("main")
			p.UnsetPermit("testcode")
		case "if":
			p.UnsetPermit("if")
			p.UnsetPermit("else")
		case "switch":
			p.UnsetPermit("switch")
			p.UnsetPermit("fallthrough")
			if !p.IsPermit("select") {
				p.UnsetPermit("case")
				p.UnsetPermit("default")
			}
			if !p.IsPermit("for") && !p.IsPermit("select") {
				p.UnsetPermit("break")
			}
		case "select":
			p.UnsetPermit("select")
			if !p.IsPermit("switch") {
				p.UnsetPermit("case")
				p.UnsetPermit("default")
			}
			if !p.IsPermit("for") && !p.IsPermit("switch") {
				p.UnsetPermit("break")
			}
		case "for":
			p.UnsetPermit("for")
			p.UnsetPermit("range")
			p.UnsetPermit("continue")
			if !p.IsPermit("switch") && !p.IsPermit("select") {
				p.UnsetPermit("break")
			}
		}
	}
}

//--------------------------------------------------------------------------------
func Propertied(p nodes.GeneralParser) nodes.Expr {
	if !p.IsPermit("propertied") {
		p.SyntaxError("\"propertied\" macro disabled but is present")
		return nil
	}
	return nil //TODO: put logic here
}

//--------------------------------------------------------------------------------
