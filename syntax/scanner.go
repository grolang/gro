// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements scanner, a lexical tokenizer for
// Go source. After initialization, consecutive calls of
// next advance the scanner one token at a time.
//
// This file, source.go, and tokens.go are self-contained
// (go tool compile scanner.go source.go tokens.go compiles)
// and thus could be made into its own package.

package syntax

import (
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"

	"github.com/grolang/gro/nodes"
)

//--------------------------------------------------------------------------------
func init() {
	// populate keywordMap
	for tok := nodes.BreakT; tok <= nodes.VarT; tok++ {
		h := hash([]byte(nodes.Tokstrings[tok]))
		if keywordMap[h] != 0 {
			panic("imperfect hash")
		}
		keywordMap[h] = tok
	}
}

//--------------------------------------------------------------------------------
// hash is a perfect hash function for keywords.
// It assumes that s has at least length 2.
func hash(s []byte) uint {
	return (uint(s[0])<<4 ^ uint(s[1]) + uint(len(s))) & uint(len(keywordMap)-1)
}

//--------------------------------------------------------------------------------
func isLetter(c rune) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_'
}

//--------------------------------------------------------------------------------
func isDigit(c rune) bool {
	return '0' <= c && c <= '9'
}

//--------------------------------------------------------------------------------
type scanner struct {
	source
	pragh  func(line, col uint, msg string)
	nlsemi bool // if set '\n' and EOF translate to ';'

	// current token, valid after calling next()
	line, col uint
	tok       nodes.Token
	lit       string         // valid if tok is _Name, _Literal, or _Semi ("semicolon", "newline", or "EOF")
	kind      nodes.LitKind  // valid if tok is _Literal
	op        nodes.Operator // valid if tok is _Operator, _AssignOp, or _IncOp
	prec      nodes.Prec     // valid if tok is _Operator, _AssignOp, or _IncOp

	commentGroups [][]string
	comments      []string
	dynamicMode   bool
	dynCharSet    string
}

var keywordMap [1 << 6]nodes.Token // size must be power of two

//--------------------------------------------------------------------------------
func (s *scanner) init(src io.Reader, errh, pragh func(line, col uint, msg string)) {
	s.source.init(src, errh)
	s.pragh = pragh
	s.nlsemi = false
}

//--------------------------------------------------------------------------------
// next advances the scanner by reading the next token.
//
// If a read, source encoding, or lexical error occurs, next
// calls the error handler installed with init. The handler
// must exist.
//
// If a //line or //go: directive is encountered at the start
// of a line, next calls the directive handler pragh installed
// with init, if not nil.
//
// The (line, col) position passed to the error and directive
// handler is always at or after the current source reading
// position.
func (s *scanner) Next() {
	nlsemi := s.nlsemi
	s.nlsemi = false
	s.commentGroups = nil
	s.comments = []string{}

redo:
	// skip white space
	c := s.getr()
	numNl := 0
	for c == ' ' || c == '\t' || c == '\n' && !nlsemi || c == '\r' {
		if c == '\n' {
			numNl++
		}
		c = s.getr()
	}
	if numNl >= 2 { //second newline in a row disqualifies preceding comment as doc-comment
		if s.commentGroups == nil {
			s.commentGroups = [][]string{}
		}
		s.commentGroups = append(s.commentGroups, s.comments)
		s.comments = []string{}
	}

	// token start
	s.line, s.col = s.source.line0, s.source.col0

	if isLetter(c) || c >= utf8.RuneSelf && s.isIdentRune(c, true) {
		s.ident()
		return
	}

	switch c {
	case -1:
		if nlsemi {
			s.lit = "EOF"
			s.tok = nodes.SemiT
			break
		}
		s.tok = nodes.EofT

	case '\n':
		s.lit = "newline"
		s.tok = nodes.SemiT

	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		s.number(c)

	case '"':
		s.stdString()

	case '`':
		s.rawString()

	case '\'':
		s.rune()

	case '(':
		s.tok = nodes.LparenT

	case '[':
		s.tok = nodes.LbrackT

	case '{':
		s.tok = nodes.LbraceT

	case ',':
		s.tok = nodes.CommaT

	case ';':
		s.lit = "semicolon"
		s.tok = nodes.SemiT

	case ')':
		s.nlsemi = true
		s.tok = nodes.RparenT

	case ']':
		s.nlsemi = true
		s.tok = nodes.RbrackT

	case '}':
		s.nlsemi = true
		s.tok = nodes.RbraceT

	case ':':
		if s.getr() == '=' {
			s.tok = nodes.DefineT
			break
		}
		s.ungetr()
		s.tok = nodes.ColonT

	case '.':
		c = s.getr()
		if isDigit(c) {
			s.ungetr2()
			s.number('.')
			break
		}
		if c == '.' {
			c = s.getr()
			if c == '.' {
				s.tok = nodes.DotDotDotT
				break
			}
			s.ungetr2()
		}
		s.ungetr()
		s.tok = nodes.DotT

	case '+':
		s.op, s.prec = nodes.Add, nodes.PrecAdd
		c = s.getr()
		if c != '+' {
			goto assignop
		}
		s.nlsemi = true
		s.tok = nodes.IncOpT

	case '-':
		s.op, s.prec = nodes.Sub, nodes.PrecAdd
		c = s.getr()
		if c != '-' {
			goto assignop
		}
		s.nlsemi = true
		s.tok = nodes.IncOpT

	case '*':
		s.op, s.prec = nodes.Mul, nodes.PrecMul
		// don't goto assignop - want _Star token
		if s.getr() == '=' {
			s.tok = nodes.AssignOpT
			break
		}
		s.ungetr()
		s.tok = nodes.StarT

	case '/':
		c = s.getr()
		if c == '/' {
			s.lineComment()
			goto redo
		}
		if c == '*' {
			s.fullComment()
			if s.source.line > s.line && nlsemi {
				// A multi-line comment acts like a newline;
				// it translates to a ';' if nlsemi is set.
				s.lit = "newline"
				s.tok = nodes.SemiT
				break
			}
			goto redo
		}
		s.op, s.prec = nodes.Div, nodes.PrecMul
		goto assignop

	case '#':
		if s.line != 1 || s.col != 1 {
			s.error("#! not at first position")
		}
		c = s.getr()
		if c != '!' {
			s.error("# not followed by !")
		}
		r := s.getr()
		s.startLit()
		s.skipLine(r)
		s.getr()
		s.stopLit()
		goto redo

	case '%':
		s.op, s.prec = nodes.Rem, nodes.PrecMul
		c = s.getr()
		goto assignop

	case '&':
		c = s.getr()
		if c == '&' {
			s.op, s.prec = nodes.AndAnd, nodes.PrecAndAnd
			s.tok = nodes.OperatorT
			break
		}
		s.op, s.prec = nodes.And, nodes.PrecMul
		if c == '^' {
			s.op = nodes.AndNot
			c = s.getr()
			goto assignop
		}
		if c == '>' {
			s.op = nodes.AndRgt
			c = s.getr()
			goto assignop
		}
		goto assignop

	case '|':
		c = s.getr()
		if c == '|' {
			s.op, s.prec = nodes.OrOr, nodes.PrecOrOr
			s.tok = nodes.OperatorT
			break
		}
		s.op, s.prec = nodes.Or, nodes.PrecAdd
		goto assignop

	case '~':
		s.error("bitwise complement operator is ^")
		fallthrough

	case '^':
		s.op, s.prec = nodes.Xor, nodes.PrecAdd
		c = s.getr()
		goto assignop

	case '<':
		c = s.getr()
		if c == '=' {
			s.op, s.prec = nodes.Leq, nodes.PrecCmp
			s.tok = nodes.OperatorT
			break
		}
		if c == '<' {
			s.op, s.prec = nodes.Shl, nodes.PrecMul
			c = s.getr()
			goto assignop
		}
		if c == '&' {
			s.op, s.prec = nodes.LftAnd, nodes.PrecMul
			c = s.getr()
			goto assignop
		}
		if c == '-' {
			s.tok = nodes.ArrowT
			break
		}
		s.ungetr()
		s.op, s.prec = nodes.Lss, nodes.PrecCmp
		s.tok = nodes.OperatorT

	case '>':
		c = s.getr()
		if c == '=' {
			s.op, s.prec = nodes.Geq, nodes.PrecCmp
			s.tok = nodes.OperatorT
			break
		}
		if c == '>' {
			s.op, s.prec = nodes.Shr, nodes.PrecMul
			c = s.getr()
			goto assignop
		}
		s.ungetr()
		s.op, s.prec = nodes.Gtr, nodes.PrecCmp
		s.tok = nodes.OperatorT

	case '=':
		if s.getr() == '=' {
			s.op, s.prec = nodes.Eql, nodes.PrecCmp
			s.tok = nodes.OperatorT
			break
		}
		s.ungetr()
		s.tok = nodes.AssignT

	case '!':
		if s.getr() == '=' {
			s.op, s.prec = nodes.Neq, nodes.PrecCmp
			s.tok = nodes.OperatorT
			break
		}
		s.ungetr()
		s.op, s.prec = nodes.Not, 0
		s.tok = nodes.OperatorT

	default:
		s.tok = 0
		s.error(fmt.Sprintf("invalid character %#U", c))
		goto redo
	}

	return

assignop:
	if c == '=' {
		s.tok = nodes.AssignOpT
		return
	}
	s.ungetr()
	s.tok = nodes.OperatorT
}

//--------------------------------------------------------------------------------
func (s *scanner) ident() {
	s.startLit()

	// accelerate common case (7bit ASCII)
	c := s.getr()
	for isLetter(c) || isDigit(c) {
		c = s.getr()
	}

	// general case
	if c >= utf8.RuneSelf {
		for s.isIdentRune(c, false) {
			c = s.getr()
		}
	}
	s.ungetr()

	lit := s.stopLit()

	// possibly a keyword
	if len(lit) >= 2 {
		if tok := keywordMap[hash(lit)]; tok != 0 && nodes.Tokstrings[tok] == string(lit) {
			s.nlsemi = nodes.Contains(1<<nodes.BreakT|1<<nodes.ContinueT|1<<nodes.FallthroughT|1<<nodes.ReturnT, tok)
			s.tok = tok
			return
		}
	}

	s.nlsemi = true
	s.lit = string(lit)
	s.tok = nodes.NameT
}

//--------------------------------------------------------------------------------
func (s *scanner) isIdentRune(c rune, first bool) bool {
	switch {
	case unicode.IsLetter(c) || c == '_':
		// ok
	case unicode.IsDigit(c):
		if first {
			s.error(fmt.Sprintf("identifier cannot begin with digit %#U", c))
		}
	case c >= utf8.RuneSelf:
		s.error(fmt.Sprintf("invalid identifier character %#U", c))
	default:
		return false
	}
	return true
}

//--------------------------------------------------------------------------------
func (s *scanner) number(c rune) {
	s.startLit()

	if c != '.' {
		s.kind = nodes.IntLit // until proven otherwise
		if c == '0' {
			c = s.getr()
			if c == 'x' || c == 'X' {
				// hex
				c = s.getr()
				hasDigit := false
				for isDigit(c) || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' || c == '_' {
					c = s.getr()
					hasDigit = true
				}
				if !hasDigit {
					s.error("malformed hex constant")
				}
				goto done
			}

			// decimal 0, octal, or float
			has8or9 := false
			for isDigit(c) || c == '_' {
				if c == '8' || c == '9' {
					has8or9 = true
				}
				c = s.getr()
			}
			if c != '.' && c != 'e' && c != 'E' && c != 'i' {
				// octal
				if has8or9 {
					s.error("malformed octal constant")
				}
				goto done
			}

		} else {
			// decimal or float
			for isDigit(c) || c == '_' {
				c = s.getr()
			}
		}
	}

	// float
	if c == '.' {
		s.kind = nodes.FloatLit
		c = s.getr()
		for isDigit(c) || c == '_' {
			c = s.getr()
		}
	}

	// date
	if c == '.' {
		s.kind = nodes.DateLit
		c = s.getr()
		for isDigit(c) || c == '_' {
			c = s.getr()
		}
	}

	// exponent
	if c == 'e' || c == 'E' {
		s.kind = nodes.FloatLit
		c = s.getr()
		if c == '-' || c == '+' {
			c = s.getr()
		}
		if !isDigit(c) {
			s.error("malformed floating-point constant exponent")
		}
		for isDigit(c) {
			c = s.getr()
		}
	}

	// complex
	if c == 'i' {
		s.kind = nodes.ImagLit
		s.getr()
	}

done:
	s.ungetr()
	s.nlsemi = true
	s.lit = string(s.stopLit())
	s.tok = nodes.LiteralT
}

//--------------------------------------------------------------------------------
func (s *scanner) rune() {
	s.startLit()

	ok := true // only report errors if we're ok so far
	n := 0
	for ; ; n++ {
		r := s.getr()
		if r == '\'' {
			break
		}
		if r == '\\' {
			if !s.escape('\'') {
				ok = false
			}
			continue
		}
		if r == '\n' {
			s.ungetr() // assume newline is not part of literal
			if ok {
				s.error("newline in character literal")
				ok = false
			}
			break
		}
		if r < 0 {
			if ok {
				s.errh(s.line, s.col, "invalid character literal (missing closing ')")
				ok = false
			}
			break
		}
	}

	if ok {
		if n == 0 {
			s.error("empty character literal or unescaped ' in character literal")
			/*} else if n != 1 {
			s.errh(s.line, s.col, "invalid character literal (more than one character)")
			*/
		}
	}

	s.nlsemi = true
	s.lit = string(s.stopLit())
	s.kind = nodes.RuneLit
	s.tok = nodes.LiteralT
}

//--------------------------------------------------------------------------------
func (s *scanner) stdString() {
	s.startLit()

	for {
		r := s.getr()
		if r == '"' {
			break
		}
		if r == '\\' {
			s.escape('"')
			continue
		}
		if r == '\n' {
			s.ungetr() // assume newline is not part of literal
			s.error("newline in string")
			break
		}
		if r < 0 {
			s.errh(s.line, s.col, "string not terminated")
			break
		}
	}

	s.nlsemi = true
	s.lit = string(s.stopLit())
	s.kind = nodes.StringLit
	s.tok = nodes.LiteralT
}

//--------------------------------------------------------------------------------
func (s *scanner) rawString() {
	s.startLit()

	for {
		r := s.getr()
		if r == '`' {
			break
		}
		if r < 0 {
			s.errh(s.line, s.col, "string not terminated")
			break
		}
	}
	// We leave CRs in the string since they are part of the
	// literal (even though they are not part of the literal
	// value).

	s.nlsemi = true
	s.lit = string(s.stopLit())
	s.kind = nodes.StringLit
	s.tok = nodes.LiteralT
}

//--------------------------------------------------------------------------------
func (s *scanner) skipLine(r rune) {
	for r >= 0 {
		if r == '\n' {
			s.ungetr() // don't consume '\n' - needed for nlsemi logic
			break
		}
		r = s.getr()
	}
}

//--------------------------------------------------------------------------------
func (s *scanner) escape(quote rune) bool {
	var n int
	var base, max uint32
	const maxUtf88Point = 0x7fbfffff //TODO: should reference utf88 package

	c := s.getr()
	switch c {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', quote:
		return true
	case '0', '1', '2', '3', '4', '5', '6', '7':
		n, base, max = 3, 8, 255
	case 'x':
		c = s.getr()
		n, base, max = 2, 16, 255
	case 'u':
		c = s.getr()
		n, base, max = 4, 16, unicode.MaxRune
	case 'U':
		c = s.getr()
		n, base = 8, 16
		if s.dynamicMode {
			max = maxUtf88Point
		} else {
			max = unicode.MaxRune
		}
	default:
		if c < 0 {
			return true // complain in caller about EOF
		}
		s.error("unknown escape sequence")
		return false
	}

	var x uint32
	for i := n; i > 0; i-- {
		d := base
		switch {
		case isDigit(c):
			d = uint32(c) - '0'
		case 'a' <= c && c <= 'f':
			d = uint32(c) - ('a' - 10)
		case 'A' <= c && c <= 'F':
			d = uint32(c) - ('A' - 10)
		}
		if d >= base {
			if c < 0 {
				return true // complain in caller about EOF
			}
			kind := "hex"
			if base == 8 {
				kind = "octal"
			}
			s.error(fmt.Sprintf("non-%s character in escape sequence: %c", kind, c))
			s.ungetr()
			return false
		}
		// d < base
		x = x*base + d
		c = s.getr()
	}
	s.ungetr()

	if x > max && base == 8 {
		s.error(fmt.Sprintf("octal escape value > 255: %d", x))
		return false
	}

	if x > max || 0xD800 <= x && x < 0xE000 /* surrogate range */ { //TODO: add utf-88 surrogates
		s.error("escape sequence is invalid Unicode code point")
		return false
	}

	return true
}

//--------------------------------------------------------------------------------
func (s *scanner) lineComment() {
	r := s.getr()
	// directives must start at the beginning of the line (s.col == colbase)
	if s.col != colbase || s.pragh == nil || (r != 'g' && r != 'l') {
		s.startLit()
		s.skipLine(r)
		s.comments = append(s.comments, "//"+string(s.stopLit()))
		return
	}
	// s.col == colbase && s.pragh != nil && (r == 'g' || r == 'l')

	// recognize directives
	prefix := "go:"
	if r == 'l' {
		prefix = "line "
	}
	rs := ""
	for _, m := range prefix {
		if r != m {
			s.startLit()
			s.skipLine(r)
			text := s.stopLit()
			s.comments = append(s.comments, "//"+rs+string(text))
			return
		}
		rs += string(r)
		r = s.getr()
	}

	// directive text without line ending (which may be "\r\n" if Windows),
	s.startLit()
	s.skipLine(r)
	text := s.stopLit()
	if i := len(text) - 1; i >= 0 && text[i] == '\r' {
		text = text[:i]
	}

	s.pragh(s.line, s.col+2, prefix+string(text)) // +2 since directive text starts after //
	s.comments = append(s.comments, "//"+prefix+string(text))
}

//--------------------------------------------------------------------------------
func (s *scanner) fullComment() {
	s.startLit()

	for {
		r := s.getr()
		for r == '*' {
			r = s.getr()
			if r == '/' {
				s.comments = append(s.comments, "/"+string(s.stopLit()))
				return
			}
		}
		if r < 0 {
			s.errh(s.line, s.col, "comment not terminated")
			return
		}
	}
}

//--------------------------------------------------------------------------------
