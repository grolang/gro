// Copyright 2016 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

//Package for macroized commands such as 准, 跑.
package command

import (
	"go/ast"
	"go/token"
	"github.com/grolang/gro/macro"
)

type Prepare struct{}

func (m Prepare) Init(p interface{}) {
}

func (m Prepare) Main(p macro.Parser) ast.Stmt {

	p.Next()
	cmd:= ""
	if p.Tok() == token.STRING {
		cmd = p.Lit()
		p.Next()
	} else {
		p.ErrorExpected(p.Pos(), "string")
	}
	p.AppendUnihanImports(&ast.ImportSpec{
		Name: &ast.Ident{NamePos: token.NoPos, Name: "exec"},
		Path: &ast.BasicLit{ValuePos: token.NoPos, Kind: token.STRING, Value: "os/exec"},
	})
	p.AppendUnihanImports(&ast.ImportSpec{
		Name: &ast.Ident{NamePos: token.NoPos, Name: "fmt"},
		Path: &ast.BasicLit{ValuePos: token.NoPos, Kind: token.STRING, Value: "fmt"},
	})

	s:= &ast.ExprStmt{
		//`func(){
		//  x, _:= exec.Command("gro", ` + cmd + `).Output()
		//  fmt.Println(string(x))
		//}()`
		X: &ast.CallExpr {
			Fun: &ast.FuncLit {
				Type: &ast.FuncType { Params: &ast.FieldList {} },
				Body: &ast.BlockStmt {
					List: []ast.Stmt {
						&ast.AssignStmt {
							Lhs: []ast.Expr { &ast.Ident { Name: "x" }, &ast.Ident { Name: "_" } },
							Tok: token.DEFINE,
							Rhs: []ast.Expr {
								&ast.CallExpr {
									Fun: &ast.SelectorExpr {
										X: &ast.CallExpr {
											Fun: &ast.SelectorExpr { X: &ast.Ident { Name: "exec" }, Sel: &ast.Ident { Name: "Command" } },
											Args: []ast.Expr {
												&ast.BasicLit { Kind: token.STRING, Value: "\"gro\""},
												&ast.BasicLit { Kind: token.STRING, Value: "\"prepare\""},
												&ast.BasicLit { Kind: token.STRING, Value: cmd},
											},
										},
										Sel: &ast.Ident { Name: "Output" },
									},
								},
							},
						},
						&ast.ExprStmt {
							X: &ast.CallExpr {
								Fun: &ast.SelectorExpr { X: &ast.Ident { Name: "fmt" }, Sel: &ast.Ident { Name: "Println" } },
								Args: []ast.Expr {
									&ast.CallExpr {
										Fun: &ast.Ident { Name: "string" },
										Args: []ast.Expr { &ast.Ident { Name: "x" } },
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return s
}

type Run struct{}

func (m Run) Init(p interface{}) {
}

func (m Run) Main(p macro.Parser) ast.Stmt {

	p.Next()
	cmd:= ""
	if p.Tok() == token.STRING {
		cmd = p.Lit()
		p.Next()
	} else {
		p.ErrorExpected(p.Pos(), "string")
	}
	p.AppendUnihanImports(&ast.ImportSpec{
		Name: &ast.Ident{NamePos: token.NoPos, Name: "exec"},
		Path: &ast.BasicLit{ValuePos: token.NoPos, Kind: token.STRING, Value: "os/exec"},
	})
	p.AppendUnihanImports(&ast.ImportSpec{
		Name: &ast.Ident{NamePos: token.NoPos, Name: "fmt"},
		Path: &ast.BasicLit{ValuePos: token.NoPos, Kind: token.STRING, Value: "fmt"},
	})

	s:= &ast.ExprStmt{
		//`func(){
		//  x, _:= exec.Command("go", "run", ` + cmd + `).Output()
		//  fmt.Println(string(x))
		//}()`
		X: &ast.CallExpr {
			Fun: &ast.FuncLit {
				Type: &ast.FuncType { Params: &ast.FieldList {} },
				Body: &ast.BlockStmt {
					List: []ast.Stmt {
						&ast.AssignStmt {
							Lhs: []ast.Expr { &ast.Ident { Name: "x" }, &ast.Ident { Name: "_" } },
							Tok: token.DEFINE,
							Rhs: []ast.Expr {
								&ast.CallExpr {
									Fun: &ast.SelectorExpr {
										X: &ast.CallExpr {
											Fun: &ast.SelectorExpr { X: &ast.Ident { Name: "exec" }, Sel: &ast.Ident { Name: "Command" } },
											Args: []ast.Expr {
												&ast.BasicLit { Kind: token.STRING, Value: "\"go\""},
												&ast.BasicLit { Kind: token.STRING, Value: "\"run\""},
												&ast.BasicLit { Kind: token.STRING, Value: cmd},
											},
										},
										Sel: &ast.Ident { Name: "Output" },
									},
								},
							},
						},
						&ast.ExprStmt {
							X: &ast.CallExpr {
								Fun: &ast.SelectorExpr { X: &ast.Ident { Name: "fmt" }, Sel: &ast.Ident { Name: "Println" } },
								Args: []ast.Expr {
									&ast.CallExpr {
										Fun: &ast.Ident { Name: "string" },
										Args: []ast.Expr { &ast.Ident { Name: "x" } },
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return s
}

