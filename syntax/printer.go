// Copyright 2016-17 The Go and Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements printing of syntax trees in source format.

package syntax

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/grolang/gro/nodes"
	"github.com/grolang/gro/syntax/src"
)

//================================================================================
// public interface
//--------------------------------------------------------------------------------
// TODO(gri) Consider removing the linebreaks flag from this signature.
// Its likely rarely used in common cases.

func Fprint(w io.Writer, x nodes.Node, linebreaks bool) (n int, err error) {
	p := printer{
		output:     w,
		linebreaks: linebreaks,
	}

	defer func() {
		n = p.written
		if e := recover(); e != nil {
			err = e.(localError).err // re-panics if it's not a localError
		}
	}()

	p.Print(x)
	p.flush(nodes.EofT)

	return
}

//--------------------------------------------------------------------------------

func String(n nodes.Node) string {
	var buf bytes.Buffer
	_, err := Fprint(&buf, n, false)
	if err != nil {
		panic(err) // TODO(gri) print something sensible into buf instead
	}
	return buf.String()
}

//--------------------------------------------------------------------------------

func StringWithLinebreaks(n nodes.Node) string {
	var buf bytes.Buffer
	_, err := Fprint(&buf, n, true)
	if err != nil {
		panic(err) // TODO(gri) print something sensible into buf instead
	}
	return buf.String()
}

//================================================================================
// private types and consts
//--------------------------------------------------------------------------------
type printer struct {
	output     io.Writer
	written    int  // number of bytes written
	linebreaks bool // print linebreaks instead of semis

	indent        int // current indentation level
	nlcount       int // number of consecutive newlines
	nlsince       int // total number of newlines since last line directive
	lastlinedirno uint

	pending []whitespace // pending whitespace
	lastTok nodes.Token  // last token (after any pending semi) processed by print
}

//--------------------------------------------------------------------------------
type whitespace struct {
	last nodes.Token
	kind ctrlSymbol
	sym  nodes.Symbol // valid if kind == symbol
	text string       // comment text (possibly ""); valid if kind == comment
}

type ctrlSymbol int

const (
	none ctrlSymbol = iota
	semi
	comment
	eolComment
	symbol
)

//================================================================================
// public methods of printer
//--------------------------------------------------------------------------------
func (p printer) Linebreaks() bool {
	return p.linebreaks
}

//--------------------------------------------------------------------------------
func (p *printer) Print(args ...interface{}) {
	for i := 0; i < len(args); i++ {
		switch x := args[i].(type) {
		case nil:
			// we should not reach here but don't crash

		case nodes.Node:
			p.PrintNode(x)

		case *nodes.Comment: // comments are not Nodes
			p.addWhitespace(comment, 0, x.Text)
			p.flush(nodes.CommentT)

		case nodes.Token:
			// NameT implies an immediately following string argument
			// which is the actual value to print.
			var s string
			if x == nodes.NameT {
				i++
				if i >= len(args) {
					panic("missing string argument after NameT")
				}
				s = args[i].(string)
			} else {
				s = x.String()
			}

			// TODO(gri) This check seems at the wrong place since it doesn't
			//           take into account pending white space.
			if len(s) > 0 && mayCombine(p.lastTok, s[0]) {
				panic("adjacent tokens combine without whitespace")
			}

			if x == nodes.SemiT {
				// delay printing of semi
				p.addWhitespace(semi, 0, "")
			} else {
				p.flush(x)
				p.writeString(s)
				p.nlcount = 0
				p.lastTok = x
			}

		case nodes.Operator:
			if x != 0 {
				p.flush(nodes.OperatorT)
				p.writeString(x.String())
			}

		case nodes.Symbol:
			if x == nodes.NewlineSym && !p.linebreaks {
				x = nodes.BlankSym
			}
			p.addWhitespace(symbol, x, "")
			// TODO(gri) need to handle mandatory newlines after a //-style comment

		default: //incl case ctrlSymbol
			panic(fmt.Sprintf("unexpected argument %v (%T)", x, x))
		}
	}
}

//--------------------------------------------------------------------------------
func (p *printer) PrintNode(n nodes.Node) {
	// TODO(gri) in general we cannot make assumptions about whether
	// a comment is a /*- or a //-style comment since the syntax
	// tree may have been manipulated. Need to make sure the correct
	// whitespace is emitted.
	ncom := n.Comments()
	if ncom != nil && len(ncom.Alone) > 0 {
		for _, c := range ncom.Alone {
			p.Print(c, nodes.NewlineSym, nodes.NewlineSym)
		}
	}

	if n.LineDirect() != src.NoPos {
		lineNo := n.LineDirect().Line()
		if uint(p.nlsince+1) != lineNo-p.lastlinedirno {
			//TODO: perhaps we should do the following before the above if-stmt
			if ncom != nil && ncom.Above != nil {
				lineNo -= uint(strings.Count(ncom.Above.Text, "\n") + 1)
			}
			t := "//line " + n.LineDirect().Filename() + ":" + strconv.FormatUint(uint64(lineNo), 10)
			p.Print(&nodes.Comment{Text: t}, nodes.NewlineSym)
			p.lastlinedirno = lineNo
		}
	}

	if ncom != nil {
		if c := ncom.Above; c != nil {
			if c.Text == "" {
				panic("unexpected empty line")
			}
			p.Print(c, nodes.NewlineSym)
		}
		if c := ncom.Left; c != nil {
			if c.Text == "" || lineComment(c.Text) {
				panic("unexpected empty line or //-style 'left' comment")
			}
			p.Print(c, nodes.BlankSym)
		}
	}

	switch n.(type) {
	case nil:
		// we should not reach here but don't crash
	case nodes.Node:
		n.Print(p)
	default:
		panic(fmt.Sprintf("syntax.PrintNode: unexpected node type %T", n))
	}

	if ncom != nil && ncom.Right != nil {
		if c := ncom.Right; c != nil {
			if c.Text == "" {
				panic("unexpected empty line")
			}
			p.Print(nodes.BlankSym, c)
		}
	}
	if ncom != nil && ncom.Below != nil {
		if c := ncom.Below; c != nil {
			if c.Text == "" {
				panic("unexpected empty line")
			}
			p.Print(nodes.NewlineSym, c)
		}
	}
}

//================================================================================
// private methods of printer
//--------------------------------------------------------------------------------
// write is a thin wrapper around p.output.Write
// that takes care of accounting and error handling.
func (p *printer) write(data []byte) {
	n, err := p.output.Write(data)
	p.written += n
	if err != nil {
		panic(localError{err})
	}
}

//--------------------------------------------------------------------------------
func (p *printer) writeBytes(data []byte) {
	/*if len(data) == 0 {
		panic("expected non-empty []byte")
	}*/
	lineDir := strings.HasPrefix(string(data), "//line ")
	if p.nlcount > 0 && p.indent > 0 && !lineDir {
		// write indentation
		n := p.indent
		for n > len(tabBytes) {
			p.write(tabBytes)
			n -= len(tabBytes)
		}
		p.write(tabBytes[:n])
	}
	p.write(data)
	p.nlcount = 0
	if lineDir {
		p.nlsince = 0
	}
}

var (
	tabBytes    = []byte("\t\t\t\t\t\t\t\t")
	newlineByte = []byte("\n")
	blankByte   = []byte(" ")
)

//--------------------------------------------------------------------------------
func (p *printer) writeString(s string) {
	p.writeBytes([]byte(s))
}

//--------------------------------------------------------------------------------
func (p *printer) addWhitespace(kind ctrlSymbol, sym nodes.Symbol, text string) {
	p.pending = append(p.pending, whitespace{p.lastTok, kind, sym, text})
	switch kind {
	case semi:
		p.lastTok = nodes.SemiT
	case symbol:
		if sym == nodes.NewlineSym {
			p.lastTok = 0
		}
		// TODO(gri) do we need to handle /*-style comments containing newlines here?
	}
}

//--------------------------------------------------------------------------------
func (p *printer) flush(next nodes.Token) {
	// eliminate semis and redundant whitespace
	sawNewline := next == nodes.EofT
	sawParen := next == nodes.RparenT || next == nodes.RbraceT
	for i := len(p.pending) - 1; i >= 0; i-- {
		switch p.pending[i].kind {
		case semi:
			k := semi
			if sawParen {
				sawParen = false
				k = none // eliminate semi
			} else if sawNewline && impliesSemi(p.pending[i].last) {
				sawNewline = false
				k = none // eliminate semi
			}
			p.pending[i].kind = k
			if k == none {
				p.pending[i].sym = 0
			}
		case comment:
			// A multi-line comment acts like a newline; and a ""
			// comment implies by definition at least one newline.
			if text := p.pending[i].text; strings.HasPrefix(text, "/*") && strings.ContainsRune(text, '\n') {
				sawNewline = true
			}
		case eolComment:
			// TODO(gri) act depending on sawNewline
		case symbol:
			if p.pending[i].sym == nodes.NewlineSym {
				sawNewline = true
			}
		default:
			panic("unreachable")
		}
	}

	// print pending
	var prevSym nodes.Symbol = 0
	for i := range p.pending {
		switch p.pending[i].kind {
		case none:
			// nothing to do
		case semi:
			p.writeString(";")
			p.nlcount = 0
			prevSym = 0
		case comment:
			if text := p.pending[i].text; text != "" {
				p.writeString(text)
				p.nlcount = 0
				prevSym = 0
			}
			// TODO(gri) should check that line comments are always followed by newline
		case symbol:
			switch p.pending[i].sym {
			case nodes.BlankSym:
				if prevSym != nodes.BlankSym {
					// at most one blank
					p.writeBytes(blankByte)
					p.nlcount = 0
					prevSym = nodes.BlankSym
				}
			case nodes.NewlineSym:
				const maxEmptyLines = 1
				if p.nlcount <= maxEmptyLines {
					p.write(newlineByte)
					p.nlcount++
					prevSym = nodes.NewlineSym
				}
				p.nlsince++
			case nodes.IndentSym:
				p.indent++
			case nodes.OutdentSym:
				p.indent--
				if p.indent < 0 {
					panic("negative indentation")
				}
			}
		default:
			panic("unreachable")
		}
	}
	//_ = prev
	if next == nodes.EofT && p.written > 0 {
		p.write(newlineByte)
	}
	p.pending = p.pending[:0] // re-use underlying array
}

//================================================================================
// private functions
//--------------------------------------------------------------------------------
// If impliesSemi returns true for a non-blank line's final token tok,
// a semicolon is automatically inserted. Vice versa, a semicolon may
// be omitted in those cases.
func impliesSemi(tok nodes.Token) bool {
	switch tok {
	case nodes.NameT,
		nodes.BreakT, nodes.ContinueT, nodes.FallthroughT, nodes.ReturnT, nodes.CommentT,
		/*_Inc, _Dec,*/ nodes.RparenT, nodes.RbrackT, nodes.RbraceT: // TODO(gri) fix this
		return true
	}
	return false
}

//--------------------------------------------------------------------------------
// TODO(gri) provide table of []byte values for all tokens to avoid repeated string conversion
func lineComment(text string) bool {
	return strings.HasPrefix(text, "//")
}

//--------------------------------------------------------------------------------
func mayCombine(prev nodes.Token, next byte) (b bool) {
	return // for now
	// switch prev {
	// case lexical.Int:
	// 	b = next == '.' // 1.
	// case lexical.Add:
	// 	b = next == '+' // ++
	// case lexical.Sub:
	// 	b = next == '-' // --
	// case lexical.Quo:
	// 	b = next == '*' // /*
	// case lexical.Lss:
	// 	b = next == '-' || next == '<' // <- or <<
	// case lexical.And:
	// 	b = next == '&' || next == '^' // && or &^
	// }
	// return
}

//--------------------------------------------------------------------------------
