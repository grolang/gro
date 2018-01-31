// Copyright 2016-17 The Go and Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nodes

import (
	"fmt"

	"github.com/grolang/gro/syntax/src"
)

// ----------------------------------------------------------------------------
// Nodes

type Node interface {
	// Pos() returns the position associated with the node as follows:
	// 1) The position of a node representing a terminal syntax production
	//    (Name, BasicLit, etc.) is the position of the respective production
	//    in the source.
	// 2) The position of a node representing a non-terminal production
	//    (IndexExpr, IfStmt, etc.) is the position of a token uniquely
	//    associated with that production; usually the left-most one
	//    ('[' for IndexExpr, 'if' for IfStmt, etc.)
	Pos() src.Pos
	SetPos(src.Pos)
	LineDirect() src.Pos
	TagLineDirect()
	Comments() *Comments
	MakeComments()
	SetAboveComment(string)
	SetRightComment(string)
	AppendAloneComment(string)
	Print(printer)
	Walk(*Walker) Node
	aNode()
}

type node struct {
	pos        src.Pos
	lineDirect src.Pos   // nil means no line-directive to be printed
	comments   *Comments // nil means no comment(s) attached
}

func (n *node) Pos() src.Pos             { return n.pos }
func (n *node) SetPos(pos src.Pos)       { n.pos = pos }
func (n *node) LineDirect() src.Pos      { return n.lineDirect }
func (n *node) TagLineDirect()           { n.lineDirect = n.pos }
func (n *node) Comments() *Comments      { return n.comments }
func (n *node) MakeComments()            { n.comments = &Comments{Alone: []*Comment{}} }
func (n *node) SetAboveComment(s string) { n.comments.Above = &Comment{Text: s} }
func (n *node) SetRightComment(s string) { n.comments.Right = &Comment{Text: s} }
func (n *node) AppendAloneComment(s string) {
	n.comments.Alone = append(n.comments.Alone, &Comment{Text: s})
}
func (*node) aNode() {}

type Comments struct {
	Alone []*Comment
	Above *Comment
	Left  *Comment
	Right *Comment
	Below *Comment
}

type Comment struct {
	Text string
	pos  src.Pos
}

// ----------------------------------------------------------------------------
// Files

type Project struct {
	Name       string // project name
	FileExt    string
	Locn       string // absolute path
	Root       string
	HasKw      bool
	DirStr     string   // dir-string
	Doc        []string // doc-comment
	Pkgs       []*Package
	ArgImports []*ImportDecl
	node
}

type Package struct {
	Name    string
	Dir     string
	Params  []*Name
	Files   []*File
	IdsUsed map[string]bool
	node
}

// package PkgName; DeclList[0], DeclList[1], ...
type File struct {
	PkgName    *Name
	SectName   string
	FileName   string
	OwnerPkg   *Package
	DeclList   []Decl
	InfImports []*ImportDecl     //infered imports based on occurrences of, say, "fmt".Println
	InfImpMap  map[string]string //assoc with InfImports
	Lines      uint
	HasMain    bool // does it have a main() function?
	HasStmts   bool // does it have standalone stmts?
	HeadKw     string
	node
}

func (f *File) Clone() *File {
	decls := []Decl{}
	for _, decl := range f.DeclList {
		decls = append(decls, decl)
	}
	return &File{
		PkgName:  f.PkgName,
		DeclList: decls,
		Lines:    f.Lines,
		SectName: f.SectName,
		FileName: f.FileName,
		OwnerPkg: f.OwnerPkg,
	}
}

func (f *File) Imports() []*ImportDecl {
	ids := []*ImportDecl{}
	for _, decl := range f.DeclList {
		if decl, ok := decl.(*ImportDecl); ok {
			ids = append(ids, decl)
		}
	}
	return ids
}

func (f *File) ImportsString() string {
	impStr := ""
	for _, imp := range f.Imports() {
		impStr += fmt.Sprintf("%p ", imp)
	}
	return impStr
}

func (f *File) String() string {
	return fmt.Sprintf("[%p] pkg:%s; #decls:%d; file:%s; #imps:%d; imps:%s",
		f, f.PkgName.Value, len(f.DeclList), f.FileName, len(f.Imports()), f.ImportsString())
}

// ----------------------------------------------------------------------------
// Declarations

type decl struct{ node }

func (*decl) aDecl() {}

type (
	Decl interface {
		Node
		aDecl()
	}

	// isolated comments among declarations
	CommentDecl struct {
		CommentList []*Comment
		Group       *DeclGroup // nil means not part of a group
		decl
	}

	EmptyDecl struct {
		decl
	}

	//              Path
	// LocalPkgName Path
	ImportDecl struct {
		LocalPkgName *Name // including "."; nil means no rename present
		Path         *BasicLit
		Group        *DeclGroup // nil means not part of a group
		OwnerFile    *File
		Args         []Expr // empty or nil means not an import invoked with args
		Infers       []*ImportDecl
		decl
	}

	// Name Type
	TypeDecl struct {
		Name  *Name
		Alias bool
		Type  Expr
		Group *DeclGroup // nil means not part of a group
		//Pragma Pragma
		decl
	}

	// NameList
	// NameList      = Values
	// NameList Type = Values
	constOrVarDecl struct {
		NameList []*Name
		Type     Expr       // nil means no type
		Values   Expr       // nil means no values
		Group    *DeclGroup // nil means not part of a group
		decl
	}
	VarDecl   struct{ constOrVarDecl }
	ConstDecl struct{ constOrVarDecl }

	// func          Name Type { Body }
	// func          Name Type
	// func Receiver Name Type { Body }
	// func Receiver Name Type
	FuncDecl struct {
		Attr map[string]bool // go:attr map
		Recv *Field          // nil means regular function
		Name *Name
		Type *FuncType
		Body *BlockStmt // nil means no body (forward declaration)
		//Pragma Pragma     // TODO(mdempsky): Cleaner solution.
		decl
	}
)

// All declarations belonging to the same group point to the same DeclGroup node.
type DeclGroup struct {
	//dummy int // not empty so we are guaranteed different DeclGroup instances
	node

	Tok   Token // these 2 fields used by printer.go
	Decls []Decl
}

// ----------------------------------------------------------------------------
// Statements

type (
	Stmt interface {
		Node
		aStmt()
	}

	SimpleStmt interface {
		Stmt
		aSimpleStmt()
	}

	EmptyStmt struct {
		simpleStmt
	}

	// isolated comments among statements
	CommentStmt struct {
		CommentList []*Comment
		stmt
	}

	LabeledStmt struct {
		Label *Name
		Stmt  Stmt
		stmt
	}

	BlockStmt struct {
		List   []Stmt
		Rbrace src.Pos
		stmt
	}

	ExprStmt struct {
		X Expr
		simpleStmt
	}

	SendStmt struct {
		Chan, Value Expr // Chan <- Value
		simpleStmt
	}

	DeclStmt struct {
		DeclList []Decl
		stmt
	}

	AssignStmt struct {
		Op       Operator // 0 means no operation
		Lhs, Rhs Expr     // Rhs == ImplicitOne means Lhs++ (Op == Add) or Lhs-- (Op == Sub)
		simpleStmt
	}

	BranchStmt struct {
		Tok   Token // Break, Continue, Fallthrough, or Goto
		Label *Name
		// Target is the continuation of the control flow after executing
		// the branch; it is computed by the parser if CheckBranches is set.
		// Target is a *LabeledStmt for gotos, and a *SwitchStmt, *SelectStmt,
		// or *ForStmt for breaks and continues, depending on the context of
		// the branch. Target is not set for fallthroughs.
		Target Stmt
		stmt
	}

	CallStmt struct {
		Tok  Token // Go or Defer
		Call *CallExpr
		stmt
	}

	ReturnStmt struct {
		Results Expr // nil means no explicit return values
		stmt
	}

	IfStmt struct {
		Init SimpleStmt
		Cond Expr
		Then *BlockStmt
		Else Stmt // either *IfStmt or *BlockStmt
		stmt
	}

	ForStmt struct {
		Init SimpleStmt // incl. *RangeClause
		Cond Expr
		Post SimpleStmt
		Body *BlockStmt
		stmt
	}

	RangeClause struct {
		Lhs Expr // nil means no Lhs = or Lhs :=
		Def bool // means :=
		X   Expr // range X
		simpleStmt
	}

	SwitchStmt struct {
		Init   SimpleStmt
		Tag    Expr
		Body   []*CaseClause
		Rbrace src.Pos
		stmt
	}

	TypeSwitchGuard struct {
		// TODO(gri) consider using Name{"..."} instead of nil (permits attaching of comments)
		Lhs *Name // nil means no Lhs :=
		X   Expr  // X.(type)
		expr
	}

	CaseClause struct {
		Cases Expr // nil means default clause
		Body  []Stmt
		Colon src.Pos
		Final bool
		node
	}

	SelectStmt struct {
		Body   []*CommClause
		Rbrace src.Pos
		stmt
	}

	CommClause struct {
		Comm  SimpleStmt // send or receive stmt; nil means default clause
		Body  []Stmt
		Colon src.Pos
		Final bool
		node
	}
)

type stmt struct{ node }

func (stmt) aStmt() {}

type simpleStmt struct {
	stmt
}

func (simpleStmt) aSimpleStmt() {}

// ----------------------------------------------------------------------------
// Expressions

type (
	Expr interface {
		Node
		aExpr()
	}

	// Placeholder for an expression that failed to parse
	// correctly and where we can't provide a better node.
	BadExpr struct {
		expr
	}

	// Value
	Name struct {
		Value string
		expr
	}

	// Value
	BasicLit struct {
		Value string
		Kind  LitKind
		expr
	}

	// Type { ElemList[0], ElemList[1], ... }
	CompositeLit struct {
		Type     Expr // nil means no literal type
		ElemList []Expr
		NKeys    int // number of elements with keys
		Rbrace   src.Pos
		expr
	}

	// Key: Value
	KeyValueExpr struct {
		Key, Value Expr
		expr
	}

	// func Type { Body }
	FuncLit struct {
		Type      *FuncType
		Body      *BlockStmt
		ShortForm bool
		expr
	}

	// (X)
	ParenExpr struct {
		X Expr
		expr
	}

	// indicate is the RHS-side of dynamic-mode expression
	RhsExpr struct {
		X Expr
		expr
	}

	// X.Sel
	SelectorExpr struct {
		X   Expr
		Sel *Name
		expr
	}

	// X[Index]
	IndexExpr struct {
		X     Expr
		Index Expr
		expr
	}

	// X[Index[0] : Index[1] : Index[2]]
	SliceExpr struct {
		X     Expr
		Index [3]Expr
		// Full indicates whether this is a simple or full slice expression.
		// In a valid AST, this is equivalent to Index[2] != nil.
		// TODO(mdempsky): This is only needed to report the "3-index
		// slice of string" error when Index[2] is missing.
		Full bool
		expr
	}

	// X.(Type)
	AssertExpr struct {
		X Expr
		// TODO(gri) consider using Name{"..."} instead of nil (permits attaching of comments)
		Type Expr
		expr
	}

	Operation struct {
		Op   Operator
		X, Y Expr // Y == nil means unary expression
		expr
	}

	// Fun(ArgList[0], ArgList[1], ...)
	CallExpr struct {
		Fun     Expr
		ArgList []Expr
		HasDots bool // last argument is followed by ...
		expr
	}

	// ElemList[0], ElemList[1], ...
	ListExpr struct {
		ElemList []Expr
		expr
	}

	// Name Type
	//      Type
	Field struct {
		Name *Name // nil means anonymous field/parameter (structs/parameters), or embedded interface (interfaces)
		Type Expr  // field names declared in a list share the same Type (identical pointers)
		node
	}

	// ----------------------------------------------------------------------------
	// Types

	// [Len]Elem
	ArrayType struct {
		// TODO(gri) consider using Name{"..."} instead of nil (permits attaching of comments)
		Len  Expr // nil means Len is ...
		Elem Expr
		expr
	}

	// []Elem
	SliceType struct {
		Elem Expr
		expr
	}

	// ...Elem
	DotsType struct {
		Elem Expr
		expr
	}

	// struct { FieldList[0] TagList[0]; FieldList[1] TagList[1]; ... }
	StructType struct {
		FieldList []*Field
		TagList   []*BasicLit // i >= len(TagList) || TagList[i] == nil means no tag for field i
		expr
	}

	FuncType struct {
		ParamList  []*Field
		ResultList []*Field
		expr
	}

	// interface { MethodList[0]; MethodList[1]; ... }
	InterfaceType struct {
		MethodList []*Field
		expr
	}

	// map[Key]Value
	MapType struct {
		Key   Expr
		Value Expr
		Dyn   bool
		expr
	}

	//   chan Elem
	// <-chan Elem
	// chan<- Elem
	ChanType struct {
		Dir  ChanDir // 0 means no direction
		Elem Expr
		expr
	}
)

type expr struct{ node }

func (*expr) aExpr() {}

type ChanDir uint

const (
	_ ChanDir = iota
	SendOnly
	RecvOnly
)

// ----------------------------------------------------------------------------
