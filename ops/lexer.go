// Copyright 2017 The Gro authors. All rights reserved.
// Portions translated from Armando Blancas's Clojure-based Kern parser.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package ops

import (
	"unicode"

	u8 "github.com/grolang/gro/utf88"
)

//================================================================================

// AsciiLexer provides some functions for lexing Ascii-only text.
type AsciiLexer struct {
	joinCrNl bool // indicate whether combined \r\n will parse as one newline character.
}

// NewAsciiLexer creates an AsciiLexer. If rn is true, it will lex \r\n as
// a single newline character.
func NewAsciiLexer(rn bool) *AsciiLexer {
	pa := new(AsciiLexer)
	pa.joinCrNl = rn
	return pa
}

// Letter provides a parser of Ascii letters, either lower- or uppercase.
func (a AsciiLexer) Letter() Parsec {
	return Regexp(`[A-Za-z]`) //[:alpha:]
}

// Digit provides a parser of Ascii digits.
func (a AsciiLexer) Digit() Parsec {
	return Regexp(`[0-9]`) //[:digit:]
}

// Lower provides a parser of lowercase Ascii letters.
func (a AsciiLexer) Lower() Parsec {
	return Regexp(`[a-z]`) //[:lower:]
}

// Upper provides a parser of uppercase Ascii letters.
func (a AsciiLexer) Upper() Parsec {
	return Regexp(`[A-Z]`) //[:upper:]
}

// Lower provides a parser of the Ascii space character.
func (a AsciiLexer) Space() Parsec {
	return Regexp(` `)
}

// Tab provides a parser of the Ascii tab character.
func (a AsciiLexer) Tab() Parsec {
	return Regexp("\t")
}

// Punct provides a parser of Ascii punctuation characters.
func (a AsciiLexer) Punct() Parsec {
	return Regexp(`[!"#%&'()*,\-./:;?@[\\\]_{}]`)
}

// StartPunct provides a parser of Ascii punctuation characters
// used for starting a pair, i.e. the open paren, bracket, and curly.
func (a AsciiLexer) StartPunct() Parsec {
	return Regexp(`[([{]`)
}

// EndPunct provides a parser of Ascii punctuation characters
// used for ending a pair, i.e. the close paren, bracket, and curly.
func (a AsciiLexer) EndPunct() Parsec {
	return Regexp(`[)\]}]`)
}

// ConnectorPunct provides a parser of the Ascii connector punctuation, i.e. the underscore.
func (a AsciiLexer) ConnectorPunct() Parsec {
	return Regexp(`_`)
}

// DashPunct provides a parser of the Ascii dash punctuation, i.e. the hyphen.
func (a AsciiLexer) DashPunct() Parsec {
	return Regexp(`-`)
}

// OtherPunct provides a parser of Ascii punctuation that isn't starting, ending,
// connector, or dash.
func (a AsciiLexer) OtherPunct() Parsec {
	return Regexp(`[!"#%&'*,./:;?@\\]`)
}

// Symbol provides a parser of Ascii symbols.
func (a AsciiLexer) Symbol() Parsec {
	return Regexp(`[$+<=>^` + "`" + `|~]`)
}

// CurrencySymbol provides a parser of the Ascii currency symbols, i.e. the dollar sign.
func (a AsciiLexer) CurrencySymbol() Parsec {
	return Regexp(`$`)
}

// MathSymbol provides a parser of the Ascii math symbols.
func (a AsciiLexer) MathSymbol() Parsec {
	return Regexp(`[+<=>|~]`)
}

// ModifierSymbol provides a parser of the Ascii modifier symbols.
func (a AsciiLexer) ModifierSymbol() Parsec {
	return Regexp(`[^` + "`" + `]`)
}

// Newline provides a parser of the Ascii newline. If lexer was created
// with rn is true, \r\n will lex as a single newline character.
func (a AsciiLexer) Newline() Parsec {
	if a.joinCrNl {
		//return Regexp(`\n|\r\n?`) //TODO: allow for \r by itself
		return Regexp(`\n|\r\n`)
	} else {
		return Regexp(`\n`)
	}
}

// Whitespace provides a parser of the Ascii whitespace characters,
// i.e. one of "\t\n\v\f\r", and the space.
func (a AsciiLexer) Whitespace() Parsec {
	return Regexp(`[ \t\n\v\f\r]`) //[:space:]
}

// Control provides a parser of the Ascii control characters,
// i.e. 0x00 to 0x1F, and 0x7F.
func (a AsciiLexer) Control() Parsec {
	return Regexp(`[\x00-\x1F\x7F]`) //[:cntrl:]
}

// Graphical provides a parser of the Ascii graphical characters,
// i.e. 0x21 to 0x7E.
func (a AsciiLexer) Graphical() Parsec {
	return Regexp(`[\x21-\x7E]`) //[:graph:]
}

// Printable provides a parser of the Ascii printable characters,
// i.e. 0x20 to 0x7E.
func (a AsciiLexer) Printable() Parsec {
	return Regexp(`[\x20-\x7E]`) //[:print:]
}

//================================================================================

// UnicodeLexer provides some functions for lexing Unicode text.
type UnicodeLexer struct {
}

// NewUnicodeLexer creates an UnicodeLexer.
func NewUnicodeLexer() *UnicodeLexer {
	pu := new(UnicodeLexer)
	return pu
}

// UnicodeIn provides a parser that succeeds if the Codepoint surrogates
// to itself, and is a member of one of the ranges.
func (u UnicodeLexer) UnicodeIn(ranges ...*unicode.RangeTable) Parsec {
	return Satisfy(func(c u8.Codepoint) bool {
		cc := u8.SurrogatePoint(c)
		return len(cc) == 1 && unicode.In(cc[0], ranges...)
	})
}

// Letter provides a parser of category L, i.e. letters.
func (u UnicodeLexer) Letter() Parsec {
	return Regexp(`\pL`)
}

// Lower provides a parser of category Ll, i.e. lowercase letters.
func (u UnicodeLexer) Lower() Parsec {
	return Regexp(`\p{Ll}`)
}

// Upper provides a parser of category Lu, i.e. uppercase letters.
func (u UnicodeLexer) Upper() Parsec {
	return Regexp(`\p{Lu}`)
}

// TitlecaseLetter provides a parser of category Lt, i.e. titlecase letters.
func (u UnicodeLexer) TitlecaseLetter() Parsec {
	return Regexp(`\p{Lt}`)
}

// ModifyingLetter provides a parser of category Lm, i.e. modifying letters.
func (u UnicodeLexer) ModifyingLetter() Parsec {
	return Regexp(`\p{Lm}`)
}

// OtherLetter provides a parser of category Lo, i.e. other letters.
func (u UnicodeLexer) OtherLetter() Parsec {
	return Regexp(`\p{Lo}`)
}

// Number provides a parser of category N, i.e. numbers.
func (u UnicodeLexer) Number() Parsec {
	return Regexp(`\pN`)
}

// Digit provides a parser of category Nd, i.e. digits.
func (u UnicodeLexer) Digit() Parsec {
	return Regexp(`\p{Nd}`)
}

// LetterNumber provides a parser of category Nl, i.e. letter numbers.
func (u UnicodeLexer) LetterNumber() Parsec {
	return Regexp(`\p{Nl}`)
}

// OtherNumber provides a parser of category No, i.e. other numbers.
func (u UnicodeLexer) OtherNumber() Parsec {
	return Regexp(`\p{No}`)
}

// Mark provides a parser of category M, i.e. marks.
func (u UnicodeLexer) Mark() Parsec {
	return Regexp(`\pM`)
}

// SpacingMark provides a parser of category Mc, i.e. spacing marks.
func (u UnicodeLexer) SpacingMark() Parsec {
	return Regexp(`\p{Mc}`)
}

// EnclosingMark provides a parser of category Me, i.e. enclosing marks.
func (u UnicodeLexer) EnclosingMark() Parsec {
	return Regexp(`\p{Me}`)
}

// NonspacingMark provides a parser of category Mn, i.e. nonspacing marks.
func (u UnicodeLexer) NonspacingMark() Parsec {
	return Regexp(`\p{Mn}`)
}

// Punct provides a parser of category P, i.e. punctuation.
func (u UnicodeLexer) Punct() Parsec {
	return Regexp(`\pP`)
}

// StartPunct provides a parser of category Ps, i.e. starting punctuation.
func (u UnicodeLexer) StartPunct() Parsec {
	return Regexp(`\p{Ps}`)
}

// EndPunct provides a parser of category Pe, i.e. ending punctuation.
func (u UnicodeLexer) EndPunct() Parsec {
	return Regexp(`\p{Pe}`)
}

// InitialPunct provides a parser of category Pi, i.e. initial punctuation.
func (u UnicodeLexer) InitialPunct() Parsec {
	return Regexp(`\p{Pi}`)
}

// FinalPunct provides a parser of category Pf, i.e. final punctuation.
func (u UnicodeLexer) FinalPunct() Parsec {
	return Regexp(`\p{Pf}`)
}

// ConnectorPunct provides a parser of category Pc, i.e. connector punctuation.
func (u UnicodeLexer) ConnectorPunct() Parsec {
	return Regexp(`\p{Pc}`)
}

// DashPunct provides a parser of category Pd, i.e. dash punctuation.
func (u UnicodeLexer) DashPunct() Parsec {
	return Regexp(`\p{Pd}`)
}

// OtherPunct provides a parser of category Po, i.e. other punctuation.
func (u UnicodeLexer) OtherPunct() Parsec {
	return Regexp(`\p{Po}`)
}

// Symbol provides a parser of category S, i.e. symbols.
func (u UnicodeLexer) Symbol() Parsec {
	return Regexp(`\pS`)
}

// CurrencySymbol provides a parser of category Sc, i.e. currency symbols.
func (u UnicodeLexer) CurrencySymbol() Parsec {
	return Regexp(`\p{Sc}`)
}

// MathSymbol provides a parser of category Sm, i.e. math symbols.
func (u UnicodeLexer) MathSymbol() Parsec {
	return Regexp(`\p{Sm}`)
}

// ModifierSymbol provides a parser of category Sk, i.e. modifier symbols.
func (u UnicodeLexer) ModifierSymbol() Parsec {
	return Regexp(`\p{Sk}`)
}

// OtherSymbol provides a parser of category So, i.e. other symbols.
func (u UnicodeLexer) OtherSymbol() Parsec {
	return Regexp(`\p{So}`)
}

// UnicodeWhitespace succeeds if the utf88.Codepoint is defined by Unicode's White_Space property;
// in the Latin-1 space this is '\t', '\n', '\v', '\f', '\r', ' ', U+0085 (NEL), U+00A0 (NBSP).
func (u UnicodeLexer) Whitespace() Parsec {
	return Satisfy(func(c u8.Codepoint) bool {
		cc := u8.SurrogatePoint(c)
		return len(cc) == 1 && unicode.IsSpace(cc[0])
	})
}

// Spacing provides a parser of category Z, i.e. spacing characters.
func (u UnicodeLexer) Spacing() Parsec {
	return Regexp(`\pZ`)
}

// Space provides a parser of category Zs, i.e. spaces.
func (u UnicodeLexer) Space() Parsec {
	return Regexp(`\p{Zs}`)
}

// LineSeparator provides a parser of category Zl, i.e. the line separator.
func (u UnicodeLexer) LineSeparator() Parsec {
	return Regexp(`\p{Zl}`)
}

// ParagraphSeparator provides a parser of category Zp, i.e. the paragraph separator.
func (u UnicodeLexer) ParagraphSeparator() Parsec {
	return Regexp(`\p{Zp}`)
}

// Graphical provides a parser that succeeds if the Codepoint surrogates to itself
// and is a Graphic, i.e. from categories L, M, N, P, S, Zs.
func (u UnicodeLexer) Graphical() Parsec {
	return Satisfy(func(c u8.Codepoint) bool {
		cc := u8.SurrogatePoint(c)
		return len(cc) == 1 && unicode.IsGraphic(cc[0])
	})
}

// Printable provides a parser that succeeds if the Codepoint surrogates to itself
// and is printable by Go, i.e. from categories L, M, N, P, S, or U+0020.
func (u UnicodeLexer) Printable() Parsec {
	return Satisfy(func(c u8.Codepoint) bool {
		cc := u8.SurrogatePoint(c)
		return len(cc) == 1 && unicode.IsPrint(cc[0])
	})
}

// SpecialControl provides a parser that succeeds if the Codepoint surrogates to itself
// and is a control character, excluding some such as utf-16 surrogates.
func (u UnicodeLexer) SpecialControl() Parsec {
	return Satisfy(func(c u8.Codepoint) bool {
		cc := u8.SurrogatePoint(c)
		return len(cc) == 1 && unicode.IsControl(cc[0])
	})
}

// CommonScript provides a parser that succeeds if the Codepoint surrogates to itself
// and is in script Common.
func (u UnicodeLexer) CommonScript() Parsec {
	return Satisfy(func(c u8.Codepoint) bool {
		cc := u8.SurrogatePoint(c)
		return len(cc) == 1 && unicode.In(cc[0], unicode.Common)
	})
}

// InheritedScript provides a parser that succeeds if the Codepoint surrogates to itself
// and is in script Inherited.
func (u UnicodeLexer) InheritedScript() Parsec {
	return Satisfy(func(c u8.Codepoint) bool {
		cc := u8.SurrogatePoint(c)
		return len(cc) == 1 && unicode.In(cc[0], unicode.Inherited)
	})
}

// OtherCategory provides a parser of category C.
func (u UnicodeLexer) OtherCategory() Parsec {
	return Regexp(`\pC`)
}

// Control provides a parser of category Cc, i.e. control characters.
func (u UnicodeLexer) Control() Parsec {
	return Regexp(`\p{Cc}`)
}

// Format provides a parser of category Cf, i.e. format characters.
func (u UnicodeLexer) Format() Parsec {
	return Regexp(`\p{Cf}`)
}

// PrivateUse provides a parser of category Co, i.e. private use characters.
func (u UnicodeLexer) PrivateUse() Parsec {
	return Regexp(`\p{Co}`)
}

// Nonchar provides a parser of category Cn, i.e. nonchar characters.
func (u UnicodeLexer) Nonchar() Parsec {
	return Regexp(`\p{Cn}`)
}

// Surrogate provides a parser of category Cs, i.e. surrogates.
func (u UnicodeLexer) Surrogate() Parsec {
	return Regexp(`\p{Cs}`)
}

//================================================================================
