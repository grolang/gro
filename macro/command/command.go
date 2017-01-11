// Copyright 2016 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

//Package for macroized commands such as 准, 执, 跑.
package command

import (
	"fmt"
	"github.com/grolang/gro/ast"
	"github.com/grolang/gro/token"
	"github.com/grolang/gro/macro"
	"os/exec"
)

//--------------------------------------------------------------------------------

func FunPrep (c string) {
  x, _:= exec.Command("gro", "prepare", c).Output()
  fmt.Println(string(x))
}

type Prepare struct{}

func (m Prepare) Init(p macro.Parser) {
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
		Name: &ast.Ident{NamePos: token.NoPos, Name: "command"},
		Path: &ast.BasicLit{ValuePos: token.NoPos, Kind: token.STRING, Value: "github.com/grolang/gro/macro/command"},
	})

	//`FunPrep(` + cmd + `)`
	s:= &ast.ExprStmt{
		X: &ast.CallExpr {
			Fun: &ast.SelectorExpr { X: &ast.Ident { Name: "command" }, Sel: &ast.Ident { Name: "FunPrep" } },
			Args: []ast.Expr { &ast.BasicLit{Kind: token.STRING, Value: cmd} },
		},
	}
	return s
}

//--------------------------------------------------------------------------------

func FunExec (c string) {
  x, _:= exec.Command("gro", "execute", c).Output()
  fmt.Println(string(x))
}

type Execute struct{}

func (m Execute) Init(p macro.Parser) {
}

func (m Execute) Main(p macro.Parser) ast.Stmt {
	p.Next()
	cmd:= ""
	if p.Tok() == token.STRING {
		cmd = p.Lit()
		p.Next()
	} else {
		p.ErrorExpected(p.Pos(), "string")
	}

	p.AppendUnihanImports(&ast.ImportSpec{
		Name: &ast.Ident{NamePos: token.NoPos, Name: "command"},
		Path: &ast.BasicLit{ValuePos: token.NoPos, Kind: token.STRING, Value: "github.com/grolang/gro/macro/command"},
	})

	//`FunExec(` + cmd + `)`
	s:= &ast.ExprStmt{
		X: &ast.CallExpr {
			Fun: &ast.SelectorExpr { X: &ast.Ident { Name: "command" }, Sel: &ast.Ident { Name: "FunExec" } },
			Args: []ast.Expr { &ast.BasicLit{Kind: token.STRING, Value: cmd} },
		},
	}
	return s
}

//--------------------------------------------------------------------------------

func FunRun (c string) {
  x, _:= exec.Command("go", "run", c).Output()
  fmt.Println(string(x))
}

type Run struct{}

func (m Run) Init(p macro.Parser) {
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
		Name: &ast.Ident{NamePos: token.NoPos, Name: "command"},
		Path: &ast.BasicLit{ValuePos: token.NoPos, Kind: token.STRING, Value: "github.com/grolang/gro/macro/command"},
	})

	//`FunRun(` + cmd + `)`
	s:= &ast.ExprStmt{
		X: &ast.CallExpr {
			Fun: &ast.SelectorExpr { X: &ast.Ident { Name: "command" }, Sel: &ast.Ident { Name: "FunRun" } },
			Args: []ast.Expr { &ast.BasicLit{Kind: token.STRING, Value: cmd} },
		},
	}

	return s
}

//--------------------------------------------------------------------------------
//reminder of how much hassle it is when we don't wrap as much code as we can in a function and define the macro as just calling that function...
	/*p.AppendUnihanImports(&ast.ImportSpec{
		Name: &ast.Ident{NamePos: token.NoPos, Name: "exec"},
		Path: &ast.BasicLit{ValuePos: token.NoPos, Kind: token.STRING, Value: "os/exec"},
	})
	p.AppendUnihanImports(&ast.ImportSpec{
		Name: &ast.Ident{NamePos: token.NoPos, Name: "fmt"},
		Path: &ast.BasicLit{ValuePos: token.NoPos, Kind: token.STRING, Value: "fmt"},
	})*/
	/*s:= &ast.ExprStmt{
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
	}*/

//--------------------------------------------------------------------------------

