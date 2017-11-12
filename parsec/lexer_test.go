// Copyright 2017 The Gro authors. All rights reserved.
// Portions translated from Armando Blancas' Clojure-based Kern parser.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package parsec

import (
	ts "github.com/grolang/gro/assert"
	u8 "github.com/grolang/gro/utf88"
	"testing"
)

//================================================================================
func TestAsciiLexer(t *testing.T) {
	ts.LogAsserts("AsciiLexer", t, func(tt *ts.T) {
		a := NewAsciiLexer(false)
		at := NewAsciiLexer(true)

		tt.AssertEqual(parseStr(a.Letter(), "defg"),
			PState{input: "efg", pos: 1, value: u8.Text("d"), ok: true})
		tt.AssertEqual(parseStr(a.Letter(), ";efg"),
			PState{input: ";efg", empty: true, error: makeUnexpInp(";")})

		tt.AssertEqual(parseStr(a.Lower(), "defg"),
			PState{input: "efg", pos: 1, value: u8.Text("d"), ok: true})
		tt.AssertEqual(parseStr(a.Lower(), "Defg"),
			PState{input: "Defg", empty: true, error: makeUnexpInp("D")})
		tt.AssertEqual(parseStr(a.Lower(), ";efg"),
			PState{input: ";efg", empty: true, error: makeUnexpInp(";")})

		tt.AssertEqual(parseStr(a.Upper(), "Defg"),
			PState{input: "efg", pos: 1, value: u8.Text("D"), ok: true})
		tt.AssertEqual(parseStr(a.Upper(), "defg"),
			PState{input: "defg", empty: true, error: makeUnexpInp("d")})
		tt.AssertEqual(parseStr(a.Upper(), ";efg"),
			PState{input: ";efg", empty: true, error: makeUnexpInp(";")})

		tt.AssertEqual(parseStr(a.Digit(), "6efg"),
			PState{input: "efg", pos: 1, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(a.Digit(), "defg"),
			PState{input: "defg", empty: true, error: makeUnexpInp("d")})

		tt.AssertEqual(parseStr(a.Space(), " efg"),
			PState{input: "efg", pos: 1, value: u8.Text(" "), ok: true})
		tt.AssertEqual(parseStr(a.Space(), "defg"),
			PState{input: "defg", empty: true, error: makeUnexpInp("d")})

		tt.AssertEqual(parseStr(a.Tab(), "\tefg"),
			PState{input: "efg", pos: 1, value: u8.Text("\t"), ok: true})
		tt.AssertEqual(parseStr(a.Tab(), "defg"),
			PState{input: "defg", empty: true, error: makeUnexpInp("d")})

		tt.AssertEqual(parseStr(a.Newline(),
			"\nefg"), PState{input: "efg", pos: 1, value: u8.Text("\n"), ok: true})
		tt.AssertEqual(parseStr(a.Newline(),
			"\r\nefg"), PState{input: "\r\nefg", empty: true, error: makeUnexpInp("\r")})

		tt.AssertEqual(parseStr(at.Newline(),
			"\nefg"), PState{input: "efg", pos: 1, value: u8.Text("\n"), ok: true})
		tt.AssertEqual(parseStr(at.Newline(),
			"\r\nefg"), PState{input: "efg", pos: 2, value: u8.Text("\r\n"), ok: true})

		tt.AssertEqual(parseStr(a.Whitespace(), " efg"),
			PState{input: "efg", pos: 1, value: u8.Text(" "), ok: true})
		tt.AssertEqual(parseStr(a.Whitespace(), "\tefg"),
			PState{input: "efg", pos: 1, value: u8.Text("\t"), ok: true})
		tt.AssertEqual(parseStr(a.Whitespace(), "\refg"),
			PState{input: "efg", pos: 1, value: u8.Text("\r"), ok: true})
		tt.AssertEqual(parseStr(a.Whitespace(), "\nefg"),
			PState{input: "efg", pos: 1, value: u8.Text("\n"), ok: true})
		tt.AssertEqual(parseStr(a.Whitespace(), "\fefg"),
			PState{input: "efg", pos: 1, value: u8.Text("\f"), ok: true})
		tt.AssertEqual(parseStr(a.Whitespace(), "\vefg"),
			PState{input: "efg", pos: 1, value: u8.Text("\v"), ok: true})
		tt.AssertEqual(parseStr(a.Whitespace(), "defg"),
			PState{input: "defg", empty: true, error: makeUnexpInp("d")})

		// Punct() Parser{  return Regexp(`[!"#%&'()*,\-./:;?@[\\\]_{}]`)
		// StartPunct() Parser{  return Regexp(`[([{]`)
		// EndPunct() Parser{  return Regexp(`[)\]}]`)
		// ConnectorPunct() Parser{  return Regexp(`_`)
		// DashPunct() Parser{  return Regexp(`-`)
		// OtherPunct() Parser{  return Regexp(`[!"#%&'*,./:;?@\\]`)
		// Symbol() Parser{  return Regexp(`[$+<=>^`+"`"+`|~]`)
		// CurrencySymbol() Parser{  return Regexp(`$`)
		// MathSymbol() Parser{  return Regexp(`[+<=>|~]`)
		// ModifierSymbol() Parser{  return Regexp(`[^`+"`"+`]`)
		// Control() Parser{  return Regexp(`[\x00-\x1F\x7F]`) //[:cntrl:]
		// Graphical() Parser{  return Regexp(`[\x21-\x7E]`) //[:graph:]
		// Printable() Parser{  return Regexp(`[\x20-\x7E]`) //[:print:]
	})
}

//================================================================================
func TestLexer(t *testing.T) {
	ts.LogAsserts("Lexer", t, func(tt *ts.T) {
		u := NewUnicodeLexer()

		tt.AssertEqual(parseStr(u.Letter(), "defg"),
			PState{input: "efg", pos: 1, value: u8.Text("d"), ok: true})
		tt.AssertEqual(parseStr(u.Letter(), ";efg"),
			PState{input: ";efg", empty: true, error: makeUnexpInp(";")})

		tt.AssertEqual(parseStr(u.Lower(), "defg"),
			PState{input: "efg", pos: 1, value: u8.Text("d"), ok: true})
		tt.AssertEqual(parseStr(u.Lower(), "Defg"),
			PState{input: "Defg", empty: true, error: makeUnexpInp("D")})
		tt.AssertEqual(parseStr(u.Lower(), ";efg"),
			PState{input: ";efg", empty: true, error: makeUnexpInp(";")})

		tt.AssertEqual(parseStr(u.Upper(), "Defg"),
			PState{input: "efg", pos: 1, value: u8.Text("D"), ok: true})
		tt.AssertEqual(parseStr(u.Upper(), "defg"),
			PState{input: "defg", empty: true, error: makeUnexpInp("d")})
		tt.AssertEqual(parseStr(u.Upper(), ";efg"),
			PState{input: ";efg", empty: true, error: makeUnexpInp(";")})

		tt.AssertEqual(parseStr(u.Digit(), "6efg"),
			PState{input: "efg", pos: 1, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(u.Digit(), "defg"),
			PState{input: "defg", empty: true, error: makeUnexpInp("d")})

		tt.AssertEqual(parseStr(u.Whitespace(), " efg"),
			PState{input: "efg", pos: 1, value: u8.Codepoint(' '), ok: true})
		tt.AssertEqual(parseStr(u.Whitespace(), "\tefg"),
			PState{input: "efg", pos: 1, value: u8.Codepoint('\t'), ok: true})
		tt.AssertEqual(parseStr(u.Whitespace(), "\refg"),
			PState{input: "efg", pos: 1, value: u8.Codepoint('\r'), ok: true})
		tt.AssertEqual(parseStr(u.Whitespace(), "\nefg"),
			PState{input: "efg", pos: 1, value: u8.Codepoint('\n'), ok: true})
		tt.AssertEqual(parseStr(u.Whitespace(), "\fefg"),
			PState{input: "efg", pos: 1, value: u8.Codepoint('\f'), ok: true})
		tt.AssertEqual(parseStr(u.Whitespace(), "\vefg"),
			PState{input: "efg", pos: 1, value: u8.Codepoint('\v'), ok: true})
		tt.AssertEqual(parseStr(u.Whitespace(), "defg"),
			PState{input: "defg", empty: true, error: makeUnexpInp("d")})

		tt.AssertEqual(parseStr(u.Space(), " efg"),
			PState{input: "efg", pos: 1, value: u8.Text(" "), ok: true})
		tt.AssertEqual(parseStr(u.Space(), "defg"),
			PState{input: "defg", empty: true, error: makeUnexpInp("d")})

		// TitlecaseLetter() Parser{  return Regexp(`\p{Lt}`)
		// ModifyingLetter() Parser{  return Regexp(`\p{Lm}`)
		// OtherLetter() Parser{  return Regexp(`\p{Lo}`)
		// Number() Parser{  return Regexp(`\pN`)
		// LetterNumber() Parser{  return Regexp(`\p{Nl}`)
		// OtherNumber() Parser{  return Regexp(`\p{No}`)
		// Mark() Parser{  return Regexp(`\pM`)
		// SpacingMark() Parser{  return Regexp(`\p{Mc}`)
		// EnclosingMark() Parser{  return Regexp(`\p{Me}`)
		// NonspacingMark() Parser{  return Regexp(`\p{Mn}`)
		// Punct() Parser{  return Regexp(`\pP`)
		// StartPunct() Parser{  return Regexp(`\p{Ps}`)
		// EndPunct() Parser{  return Regexp(`\p{Pe}`)
		// InitialPunct() Parser{  return Regexp(`\p{Pi}`)
		// FinalPunct() Parser{  return Regexp(`\p{Pf}`)
		// ConnectorPunct() Parser{  return Regexp(`\p{Pc}`)
		// DashPunct() Parser{  return Regexp(`\p{Pd}`)
		// OtherPunct() Parser{  return Regexp(`\p{Po}`)
		// Symbol() Parser{  return Regexp(`\pS`)
		// CurrencySymbol() Parser{  return Regexp(`\p{Sc}`)
		// MathSymbol() Parser{  return Regexp(`\p{Sm}`)
		// ModifierSymbol() Parser{  return Regexp(`\p{Sk}`)
		// OtherSymbol() Parser{  return Regexp(`\p{So}`)
		// Spacing() Parser{  return Regexp(`\pZ`)
		// LineSeparator() Parser{  return Regexp(`\p{Zl}`)
		// ParagraphSeparator() Parser{  return Regexp(`\p{Zp}`)
		// OtherCategory() Parser{  return Regexp(`\pC`)
		// Control() Parser{  return Regexp(`\p{Cc}`)
		// Format() Parser{  return Regexp(`\p{Cf}`)
		// PrivateUse() Parser{  return Regexp(`\p{Co}`)
		// Nonchar() Parser{  return Regexp(`\p{Cn}`)
		// Surrogate() Parser{  return Regexp(`\p{Cs}`)
		// Graphical() Parser{    return len(cc) == 1 && unicode.IsGraphic(cc[0])
		// Printable() Parser{    return len(cc) == 1 && unicode.IsPrint(cc[0])
		// SpecialControl() Parser{    return len(cc) == 1 && unicode.IsControl(cc[0])
		// CommonScript() Parser{    return len(cc) == 1 && unicode.In(cc[0], unicode.Common)
		// InheritedScript() Parser{    return len(cc) == 1 && unicode.In(cc[0], unicode.Inherited)
		// UnicodeIn(ranges ...*unicode.RangeTable) Parser {    return len(cc) == 1 && unicode.In(cc[0], ranges...)
	})
}

//================================================================================
