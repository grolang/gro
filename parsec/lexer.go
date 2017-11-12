// Copyright 2017 The Gro authors. All rights reserved.
// Portions translated from Armando Blancas' Clojure-based Kern parser.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package parsec

import (
	u8 "github.com/grolang/gro/utf88"
	"unicode"
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
func (a AsciiLexer) Letter() Parser {
	return Regexp(`[A-Za-z]`) //[:alpha:]
}

// Digit provides a parser of Ascii digits.
func (a AsciiLexer) Digit() Parser {
	return Regexp(`[0-9]`) //[:digit:]
}

// Lower provides a parser of lowercase Ascii letters.
func (a AsciiLexer) Lower() Parser {
	return Regexp(`[a-z]`) //[:lower:]
}

// Upper provides a parser of uppercase Ascii letters.
func (a AsciiLexer) Upper() Parser {
	return Regexp(`[A-Z]`) //[:upper:]
}

// Lower provides a parser of the Ascii space character.
func (a AsciiLexer) Space() Parser {
	return Regexp(` `)
}

// Tab provides a parser of the Ascii tab character.
func (a AsciiLexer) Tab() Parser {
	return Regexp("\t")
}

// Punct provides a parser of Ascii punctuation characters.
func (a AsciiLexer) Punct() Parser {
	return Regexp(`[!"#%&'()*,\-./:;?@[\\\]_{}]`)
}

// StartPunct provides a parser of Ascii punctuation characters
// used for starting a pair, i.e. the open paren, bracket, and curly.
func (a AsciiLexer) StartPunct() Parser {
	return Regexp(`[([{]`)
}

// EndPunct provides a parser of Ascii punctuation characters
// used for ending a pair, i.e. the close paren, bracket, and curly.
func (a AsciiLexer) EndPunct() Parser {
	return Regexp(`[)\]}]`)
}

// ConnectorPunct provides a parser of the Ascii connector punctuation, i.e. the underscore.
func (a AsciiLexer) ConnectorPunct() Parser {
	return Regexp(`_`)
}

// DashPunct provides a parser of the Ascii dash punctuation, i.e. the hyphen.
func (a AsciiLexer) DashPunct() Parser {
	return Regexp(`-`)
}

// OtherPunct provides a parser of Ascii punctuation that isn't starting, ending,
// connector, or dash.
func (a AsciiLexer) OtherPunct() Parser {
	return Regexp(`[!"#%&'*,./:;?@\\]`)
}

// Symbol provides a parser of Ascii symbols.
func (a AsciiLexer) Symbol() Parser {
	return Regexp(`[$+<=>^` + "`" + `|~]`)
}

// CurrencySymbol provides a parser of the Ascii currency symbols, i.e. the dollar sign.
func (a AsciiLexer) CurrencySymbol() Parser {
	return Regexp(`$`)
}

// MathSymbol provides a parser of the Ascii math symbols.
func (a AsciiLexer) MathSymbol() Parser {
	return Regexp(`[+<=>|~]`)
}

// ModifierSymbol provides a parser of the Ascii modifier symbols.
func (a AsciiLexer) ModifierSymbol() Parser {
	return Regexp(`[^` + "`" + `]`)
}

// Newline provides a parser of the Ascii newline. If lexer was created
// with rn is true, \r\n will lex as a single newline character.
func (a AsciiLexer) Newline() Parser {
	if a.joinCrNl {
		//return Regexp(`\n|\r\n?`) //TODO: allow for \r by itself
		return Regexp(`\n|\r\n`)
	} else {
		return Regexp(`\n`)
	}
}

// Whitespace provides a parser of the Ascii whitespace characters,
// i.e. one of "\t\n\v\f\r", and the space.
func (a AsciiLexer) Whitespace() Parser {
	return Regexp(`[ \t\n\v\f\r]`) //[:space:]
}

// Control provides a parser of the Ascii control characters,
// i.e. 0x00 to 0x1F, and 0x7F.
func (a AsciiLexer) Control() Parser {
	return Regexp(`[\x00-\x1F\x7F]`) //[:cntrl:]
}

// Graphical provides a parser of the Ascii graphical characters,
// i.e. 0x21 to 0x7E.
func (a AsciiLexer) Graphical() Parser {
	return Regexp(`[\x21-\x7E]`) //[:graph:]
}

// Printable provides a parser of the Ascii printable characters,
// i.e. 0x20 to 0x7E.
func (a AsciiLexer) Printable() Parser {
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
func (u UnicodeLexer) UnicodeIn(ranges ...*unicode.RangeTable) Parser {
	return Satisfy(func(c u8.Codepoint) bool {
		cc := u8.SurrogatePoint(c)
		return len(cc) == 1 && unicode.In(cc[0], ranges...)
	})
}

// Letter provides a parser of category L, i.e. letters.
func (u UnicodeLexer) Letter() Parser {
	return Regexp(`\pL`)
}

// Lower provides a parser of category Ll, i.e. lowercase letters.
func (u UnicodeLexer) Lower() Parser {
	return Regexp(`\p{Ll}`)
}

// Upper provides a parser of category Lu, i.e. uppercase letters.
func (u UnicodeLexer) Upper() Parser {
	return Regexp(`\p{Lu}`)
}

// TitlecaseLetter provides a parser of category Lt, i.e. titlecase letters.
func (u UnicodeLexer) TitlecaseLetter() Parser {
	return Regexp(`\p{Lt}`)
}

// ModifyingLetter provides a parser of category Lm, i.e. modifying letters.
func (u UnicodeLexer) ModifyingLetter() Parser {
	return Regexp(`\p{Lm}`)
}

// OtherLetter provides a parser of category Lo, i.e. other letters.
func (u UnicodeLexer) OtherLetter() Parser {
	return Regexp(`\p{Lo}`)
}

// Number provides a parser of category N, i.e. numbers.
func (u UnicodeLexer) Number() Parser {
	return Regexp(`\pN`)
}

// Digit provides a parser of category Nd, i.e. digits.
func (u UnicodeLexer) Digit() Parser {
	return Regexp(`\p{Nd}`)
}

// LetterNumber provides a parser of category Nl, i.e. letter numbers.
func (u UnicodeLexer) LetterNumber() Parser {
	return Regexp(`\p{Nl}`)
}

// OtherNumber provides a parser of category No, i.e. other numbers.
func (u UnicodeLexer) OtherNumber() Parser {
	return Regexp(`\p{No}`)
}

// Mark provides a parser of category M, i.e. marks.
func (u UnicodeLexer) Mark() Parser {
	return Regexp(`\pM`)
}

// SpacingMark provides a parser of category Mc, i.e. spacing marks.
func (u UnicodeLexer) SpacingMark() Parser {
	return Regexp(`\p{Mc}`)
}

// EnclosingMark provides a parser of category Me, i.e. enclosing marks.
func (u UnicodeLexer) EnclosingMark() Parser {
	return Regexp(`\p{Me}`)
}

// NonspacingMark provides a parser of category Mn, i.e. nonspacing marks.
func (u UnicodeLexer) NonspacingMark() Parser {
	return Regexp(`\p{Mn}`)
}

// Punct provides a parser of category P, i.e. punctuation.
func (u UnicodeLexer) Punct() Parser {
	return Regexp(`\pP`)
}

// StartPunct provides a parser of category Ps, i.e. starting punctuation.
func (u UnicodeLexer) StartPunct() Parser {
	return Regexp(`\p{Ps}`)
}

// EndPunct provides a parser of category Pe, i.e. ending punctuation.
func (u UnicodeLexer) EndPunct() Parser {
	return Regexp(`\p{Pe}`)
}

// InitialPunct provides a parser of category Pi, i.e. initial punctuation.
func (u UnicodeLexer) InitialPunct() Parser {
	return Regexp(`\p{Pi}`)
}

// FinalPunct provides a parser of category Pf, i.e. final punctuation.
func (u UnicodeLexer) FinalPunct() Parser {
	return Regexp(`\p{Pf}`)
}

// ConnectorPunct provides a parser of category Pc, i.e. connector punctuation.
func (u UnicodeLexer) ConnectorPunct() Parser {
	return Regexp(`\p{Pc}`)
}

// DashPunct provides a parser of category Pd, i.e. dash punctuation.
func (u UnicodeLexer) DashPunct() Parser {
	return Regexp(`\p{Pd}`)
}

// OtherPunct provides a parser of category Po, i.e. other punctuation.
func (u UnicodeLexer) OtherPunct() Parser {
	return Regexp(`\p{Po}`)
}

// Symbol provides a parser of category S, i.e. symbols.
func (u UnicodeLexer) Symbol() Parser {
	return Regexp(`\pS`)
}

// CurrencySymbol provides a parser of category Sc, i.e. currency symbols.
func (u UnicodeLexer) CurrencySymbol() Parser {
	return Regexp(`\p{Sc}`)
}

// MathSymbol provides a parser of category Sm, i.e. math symbols.
func (u UnicodeLexer) MathSymbol() Parser {
	return Regexp(`\p{Sm}`)
}

// ModifierSymbol provides a parser of category Sk, i.e. modifier symbols.
func (u UnicodeLexer) ModifierSymbol() Parser {
	return Regexp(`\p{Sk}`)
}

// OtherSymbol provides a parser of category So, i.e. other symbols.
func (u UnicodeLexer) OtherSymbol() Parser {
	return Regexp(`\p{So}`)
}

// UnicodeWhitespace succeeds if the utf88.Codepoint is defined by Unicode's White_Space property;
// in the Latin-1 space this is '\t', '\n', '\v', '\f', '\r', ' ', U+0085 (NEL), U+00A0 (NBSP).
func (u UnicodeLexer) Whitespace() Parser {
	return Satisfy(func(c u8.Codepoint) bool {
		cc := u8.SurrogatePoint(c)
		return len(cc) == 1 && unicode.IsSpace(cc[0])
	})
}

// Spacing provides a parser of category Z, i.e. spacing characters.
func (u UnicodeLexer) Spacing() Parser {
	return Regexp(`\pZ`)
}

// Space provides a parser of category Zs, i.e. spaces.
func (u UnicodeLexer) Space() Parser {
	return Regexp(`\p{Zs}`)
}

// LineSeparator provides a parser of category Zl, i.e. the line separator.
func (u UnicodeLexer) LineSeparator() Parser {
	return Regexp(`\p{Zl}`)
}

// ParagraphSeparator provides a parser of category Zp, i.e. the paragraph separator.
func (u UnicodeLexer) ParagraphSeparator() Parser {
	return Regexp(`\p{Zp}`)
}

// Graphical provides a parser that succeeds if the Codepoint surrogates to itself
// and is a Graphic, i.e. from categories L, M, N, P, S, Zs.
func (u UnicodeLexer) Graphical() Parser {
	return Satisfy(func(c u8.Codepoint) bool {
		cc := u8.SurrogatePoint(c)
		return len(cc) == 1 && unicode.IsGraphic(cc[0])
	})
}

// Printable provides a parser that succeeds if the Codepoint surrogates to itself
// and is printable by Go, i.e. from categories L, M, N, P, S, or U+0020.
func (u UnicodeLexer) Printable() Parser {
	return Satisfy(func(c u8.Codepoint) bool {
		cc := u8.SurrogatePoint(c)
		return len(cc) == 1 && unicode.IsPrint(cc[0])
	})
}

// SpecialControl provides a parser that succeeds if the Codepoint surrogates to itself
// and is a control character, excluding some such as utf-16 surrogates.
func (u UnicodeLexer) SpecialControl() Parser {
	return Satisfy(func(c u8.Codepoint) bool {
		cc := u8.SurrogatePoint(c)
		return len(cc) == 1 && unicode.IsControl(cc[0])
	})
}

// CommonScript provides a parser that succeeds if the Codepoint surrogates to itself
// and is in script Common.
func (u UnicodeLexer) CommonScript() Parser {
	return Satisfy(func(c u8.Codepoint) bool {
		cc := u8.SurrogatePoint(c)
		return len(cc) == 1 && unicode.In(cc[0], unicode.Common)
	})
}

// InheritedScript provides a parser that succeeds if the Codepoint surrogates to itself
// and is in script Inherited.
func (u UnicodeLexer) InheritedScript() Parser {
	return Satisfy(func(c u8.Codepoint) bool {
		cc := u8.SurrogatePoint(c)
		return len(cc) == 1 && unicode.In(cc[0], unicode.Inherited)
	})
}

// OtherCategory provides a parser of category C.
func (u UnicodeLexer) OtherCategory() Parser {
	return Regexp(`\pC`)
}

// Control provides a parser of category Cc, i.e. control characters.
func (u UnicodeLexer) Control() Parser {
	return Regexp(`\p{Cc}`)
}

// Format provides a parser of category Cf, i.e. format characters.
func (u UnicodeLexer) Format() Parser {
	return Regexp(`\p{Cf}`)
}

// PrivateUse provides a parser of category Co, i.e. private use characters.
func (u UnicodeLexer) PrivateUse() Parser {
	return Regexp(`\p{Co}`)
}

// Nonchar provides a parser of category Cn, i.e. nonchar characters.
func (u UnicodeLexer) Nonchar() Parser {
	return Regexp(`\p{Cn}`)
}

// Surrogate provides a parser of category Cs, i.e. surrogates.
func (u UnicodeLexer) Surrogate() Parser {
	return Regexp(`\p{Cs}`)
}

//================================================================================
