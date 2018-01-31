// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nodes

import "fmt"

//================================================================================
type Token uint

const (
	_ Token = iota
	EofT

	CommentT // isolated comment

	// names and literals
	NameT
	LiteralT

	// operators and operations
	OperatorT // excluding '*' (StarT)
	AssignOpT
	IncOpT
	AssignT
	DefineT
	ArrowT
	StarT

	// delimiters
	LparenT
	LbrackT
	LbraceT
	RparenT
	RbrackT
	RbraceT
	CommaT
	SemiT
	ColonT
	DotT
	DotDotDotT

	// keywords
	keyword_start

	BreakT
	CaseT
	ChanT
	ConstT
	ContinueT
	DefaultT
	DeferT
	ElseT
	FallthroughT
	ForT
	FuncT
	GoT
	GotoT
	IfT
	ImportT
	InterfaceT
	MapT
	PackageT
	RangeT
	ReturnT
	SelectT
	StructT
	SwitchT
	TypeT
	VarT

	keyword_end

	tokenCount
)

// Make sure we have at most 64 tokens so we can use them in a set.
const _ uint64 = 1 << (tokenCount - 1)

//--------------------------------------------------------------------------------
func (tok Token) IsKeyword() bool {
	return tok > keyword_start && tok < keyword_end
}

//--------------------------------------------------------------------------------
func (tok Token) String() string {
	var s string
	if 0 <= tok && int(tok) < len(Tokstrings) {
		s = Tokstrings[tok]
	}
	if s == "" && tok != CommentT {
		s = fmt.Sprintf("<tok-%d>", tok)
	}
	return s
}

//--------------------------------------------------------------------------------
var Tokstrings = [...]string{
	// source control
	EofT:     "EOF",
	CommentT: "",

	// names and literals
	NameT:    "name",
	LiteralT: "literal",

	// operators and operations
	OperatorT: "op",
	AssignOpT: "op=",
	IncOpT:    "opop",
	AssignT:   "=",
	DefineT:   ":=",
	ArrowT:    "<-",
	StarT:     "*",

	// delimiters
	LparenT:    "(",
	LbrackT:    "[",
	LbraceT:    "{",
	RparenT:    ")",
	RbrackT:    "]",
	RbraceT:    "}",
	CommaT:     ",",
	SemiT:      ";",
	ColonT:     ":",
	DotT:       ".",
	DotDotDotT: "...",

	// keywords
	BreakT:       "break",
	CaseT:        "case",
	ChanT:        "chan",
	ConstT:       "const",
	ContinueT:    "continue",
	DefaultT:     "default",
	DeferT:       "defer",
	ElseT:        "else",
	FallthroughT: "fallthrough",
	ForT:         "for",
	FuncT:        "func",
	GoT:          "go",
	GotoT:        "goto",
	IfT:          "if",
	ImportT:      "import",
	InterfaceT:   "interface",
	MapT:         "map",
	PackageT:     "package",
	RangeT:       "range",
	ReturnT:      "return",
	SelectT:      "select",
	StructT:      "struct",
	SwitchT:      "switch",
	TypeT:        "type",
	VarT:         "var",
}

//--------------------------------------------------------------------------------
type LitKind uint

const (
	IntLit LitKind = iota
	FloatLit
	ImagLit
	RuneLit
	StringLit
	DateLit
)

//--------------------------------------------------------------------------------
// contains reports whether tok is in tokset.
func Contains(tokset uint64, tok Token) bool {
	return tokset&(1<<tok) != 0
}

//================================================================================
type Operator uint

const (
	_    Operator = iota
	Def           // :=
	Not           // !
	Recv          // <-

	// precOrOr
	OrOr // ||

	// precAndAnd
	AndAnd // &&

	// precCmp
	Eql // ==
	Neq // !=
	Lss // <
	Leq // <=
	Gtr // >
	Geq // >=

	// precAdd
	Add // +
	Sub // -
	Or  // |
	Xor // ^

	// precMul
	Mul    // *
	Div    // /
	Rem    // %
	And    // &
	AndNot // &^
	Shl    // <<
	Shr    // >>
	AndRgt // &>
	LftAnd // <&
)

//--------------------------------------------------------------------------------
func (op Operator) String() string {
	var s string
	if 0 <= op && int(op) < len(opstrings) {
		s = opstrings[op]
	}
	if s == "" {
		s = fmt.Sprintf("<op-%d>", op)
	}
	return s
}

//--------------------------------------------------------------------------------
var opstrings = [...]string{
	// prec == 0
	Def:  ":", // : in :=
	Not:  "!",
	Recv: "<-",

	// precOrOr
	OrOr: "||",

	// precAndAnd
	AndAnd: "&&",

	// precCmp
	Eql: "==",
	Neq: "!=",
	Lss: "<",
	Leq: "<=",
	Gtr: ">",
	Geq: ">=",

	// precAdd
	Add: "+",
	Sub: "-",
	Or:  "|",
	Xor: "^",

	// precMul
	Mul:    "*",
	Div:    "/",
	Rem:    "%",
	And:    "&",
	AndNot: "&^",
	Shl:    "<<",
	Shr:    ">>",
	AndRgt: "&>",
	LftAnd: "<&",
}

//--------------------------------------------------------------------------------
// Operator precedences
type Prec int

const (
	_ Prec = iota
	PrecOrOr
	PrecAndAnd
	PrecCmp
	PrecAdd
	PrecMul
)

//================================================================================
type Symbol int

const (
	_ Symbol = iota
	BlankSym
	NewlineSym
	IndentSym
	OutdentSym
)

//================================================================================
