// Copyright 2017 The Gro authors. All rights reserved.
// Portions translated from Armando Blancas' Clojure-based Kern parser.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package parsec

import (
	ts "github.com/grolang/gro/assert"
	u8 "github.com/grolang/gro/utf88"
	"testing"
	"unicode"
)

//================================================================================
func arrText(ss ...string) []interface{} {
	at := make([]interface{}, 0, 0)
	for _, s := range ss {
		at = append(at, u8.Text(s))
	}
	return at
}

func arrCp(rs ...rune) []interface{} {
	ac := make([]interface{}, 0, 0)
	for _, r := range rs {
		ac = append(ac, u8.Codepoint(r))
	}
	return ac
}

//================================================================================
//Return, Fail, Satisfy
func TestBasics(t *testing.T) {
	ts.LogAsserts("Basics", t, func(tt *ts.T) {
		tt.AssertEqual(parseStr(Return(1234567890), "abc"),
			PState{input: "abc", value: 1234567890, ok: true, empty: true})
		tt.AssertEqual(parseStr(Fail(u8.Desur("I fail!")), "abc"),
			PState{input: "abc", empty: true, error: "I fail!"})

		letterA := Satisfy(func(c u8.Codepoint) bool { return c == 'a' })
		tt.AssertEqual(parseStr(letterA, "abc"),
			PState{input: "bc", pos: 1, value: u8.Codepoint('a'), ok: true})
		tt.AssertEqual(parseStr(letterA, "defg"),
			PState{input: "defg", empty: true, error: makeUnexpInp("d")})
	})
}

//================================================================================
//Symbol, Token, Regexp
func TestOptimizedPrimitives(t *testing.T) {
	ts.LogAsserts("OptimizedPrimitives", t, func(tt *ts.T) {
		tt.AssertEqual(parseStr(Symbol(u8.Codepoint('a')), "abcdefg"),
			PState{input: "bcdefg", pos: 1, value: u8.Codepoint('a'), ok: true})
		tt.AssertEqual(parseStr(Symbol(u8.Codepoint('a')), "cdefg"),
			PState{input: "cdefg", empty: true, error: makeUnexpInp("c")})

		tt.AssertEqual(parseStr(Token(u8.Text("abc")), "abcdefg"),
			PState{input: "defg", pos: 3, value: u8.Text("abc"), ok: true})
		tt.AssertEqual(parseStr(Token(u8.Text("abc")), "cdefg"),
			PState{input: "cdefg", empty: true, error: makeUnexpInp("cde")})
		tt.AssertEqual(parseStr(Token(u8.Text(string('a'))), "abcdefg"),
			PState{input: "bcdefg", pos: 1, value: u8.Text("a"), ok: true})
		tt.AssertEqual(parseStr(Token(u8.Text(string('a'))), "cdefg"),
			PState{input: "cdefg", empty: true, error: makeUnexpInp("c")})

		tt.AssertEqual(parseStr(Regexp(`a.c`), "abcdefg"),
			PState{input: "defg", pos: 3, value: u8.Text("abc"), ok: true})
		tt.AssertEqual(parseStr(Regexp(`a[bc]*?d`), "abcdefg"),
			PState{input: "efg", pos: 4, value: u8.Text("abcd"), ok: true})
		tt.AssertEqual(parseStr(Regexp(`abc`), "cdefg"),
			PState{input: "cdefg", empty: true, error: makeUnexpInp("c")})

		tt.AssertEqual(parseStr(Regexp(`[abc]`), "abcdefg"),
			PState{input: "bcdefg", pos: 1, value: u8.Text("a"), ok: true})
		tt.AssertEqual(parseStr(Regexp(`[abc]`), "cdefg"),
			PState{input: "defg", pos: 1, value: u8.Text("c"), ok: true})
		tt.AssertEqual(parseStr(Regexp(`[^xyz]`), "abcdefg"),
			PState{input: "bcdefg", pos: 1, value: u8.Text("a"), ok: true})
		tt.AssertEqual(parseStr(Regexp(`[ab]`), "cdefg"),
			PState{input: "cdefg", empty: true, error: makeUnexpInp("c")})
	})
}

//================================================================================
//OneOf, NoneOf, AnyChar
func TestPrimitives(t *testing.T) {
	ts.LogAsserts("Primitives", t, func(tt *ts.T) {

		tt.AssertEqual(parseStr(OneOf(u8.Desur("abcd")), "defg"),
			PState{input: "efg", pos: 1, value: u8.Codepoint('d'), ok: true})
		tt.AssertEqual(parseStr(OneOf(u8.Desur("abc")), "defg"),
			PState{input: "defg", empty: true, error: makeUnexpInp("d")})

		tt.AssertEqual(parseStr(OneOf(u8.Desur(" \t\r\v\f")), "abcdefg"),
			PState{input: "abcdefg", empty: true, error: makeUnexpInp("a")})
		tt.AssertEqual(parseStr(OneOf(u8.Desur(" \t\r\v\f")), " abc"),
			PState{input: "abc", value: u8.Codepoint(' '), pos: 1, ok: true})
		tt.AssertEqual(parseStr(OneOf(u8.Desur(" \t\r\v\f")), "\tabc"),
			PState{input: "abc", value: u8.Codepoint('\t'), pos: 1, ok: true})

		tt.AssertEqual(parseStr(NoneOf(u8.Desur("abc")), "defg"),
			PState{input: "efg", pos: 1, value: u8.Codepoint('d'), ok: true})
		tt.AssertEqual(parseStr(NoneOf(u8.Desur("abcd")), "defg"),
			PState{input: "defg", empty: true, error: makeUnexpInp("d")})

		tt.AssertEqual(parseStr(AnyChar, "defg"),
			PState{input: "efg", pos: 1, value: u8.Codepoint('d'), ok: true})
		tt.AssertEqual(parseStr(AnyChar, ""),
			PState{input: "", empty: true, error: errUnexpEof})

		tt.AssertEqual(parseStr(Tokens(u8.Text("xyz"), u8.Text("abc"), u8.Text("def")), "abcdefg"),
			PState{input: "defg", pos: 3, value: u8.Text("abc"), ok: true})
		tt.AssertEqual(parseStr(Tokens(u8.Text("xyz"), u8.Text("abc"), u8.Text("de")), "cfghij"),
			PState{input: "cfghij", empty: true, error: errNoAlts})

		tt.AssertEqual(parseStr(Field(u8.Desur("xyz")), "deyfg"),
			PState{input: "yfg", pos: 2, value: arrCp('d', 'e'), ok: true})
		tt.AssertEqual(parseStr(Field(u8.Desur("xyz")), "xabc"),
			PState{input: "xabc", value: arrCp(), ok: true, empty: true})
	})
}

//================================================================================
func globalParen() Parser {
	return SeqRight(Token(u8.Text(string('('))), SeqLeft(Fwd(globalExpr), Token(u8.Text(string(')')))))
}
func globalExpr() Parser {
	return Alt(Token(u8.Text(string('a'))), globalParen())
}

//--------------------------------------------------------------------------------
//Fwd
func TestRecursive(t *testing.T) {
	ts.LogAsserts("Recursive", t, func(tt *ts.T) {
		var expr func() Parser
		paren := func() Parser {
			return SeqRight(Token(u8.Text(string('('))), SeqLeft(Fwd(expr), Token(u8.Text(string(')')))))
		}
		expr = func() Parser {
			return Alt(Token(u8.Text(string('a'))), paren())
		}

		tt.AssertEqual(parseStr(expr(), "a"),
			PState{input: "", pos: 1, value: u8.Text("a"), ok: true})
		tt.AssertEqual(parseStr(expr(), "(a)"),
			PState{input: "", pos: 3, value: u8.Text("a"), ok: true})
		tt.AssertEqual(parseStr(expr(), "((a))"),
			PState{input: "", pos: 5, value: u8.Text("a"), ok: true})

		tt.AssertEqual(parseStr(globalExpr(), "a"),
			PState{input: "", pos: 1, value: u8.Text("a"), ok: true})
		tt.AssertEqual(parseStr(globalExpr(), "(a)"),
			PState{input: "", pos: 3, value: u8.Text("a"), ok: true})
		tt.AssertEqual(parseStr(globalExpr(), "((a))"),
			PState{input: "", pos: 5, value: u8.Text("a"), ok: true})
	})
}

//================================================================================
func globalParenWithParams(ps ...interface{}) Parser {
	return SeqRight(Token(u8.Text(string('('))), SeqLeft(FwdWithParams(globalExprWithParams, ps...), Token(u8.Text(string(')')))))
}
func globalExprWithParams(ps ...interface{}) Parser {
	return Alt(Token(u8.Text(string('a'))), globalParenWithParams(ps...))
}

//--------------------------------------------------------------------------------
//FwdWithParams
func TestRecursiveWithParams(t *testing.T) {
	ts.LogAsserts("RecursiveWithParams", t, func(tt *ts.T) {
		var expr func(ps ...interface{}) Parser
		paren := func(as ...interface{}) Parser {
			return SeqRight(Token(u8.Text(string('('))), SeqLeft(FwdWithParams(expr, as...), Token(u8.Text(string(')')))))
		}
		expr = func(ps ...interface{}) Parser {
			return Alt(Token(u8.Text(string('a'))), paren(ps...))
		}

		tt.AssertEqual(parseStr(expr(101, 102), "a"),
			PState{input: "", pos: 1, value: u8.Text("a"), ok: true})
		tt.AssertEqual(parseStr(expr(101, 102), "(a)"),
			PState{input: "", pos: 3, value: u8.Text("a"), ok: true})
		tt.AssertEqual(parseStr(expr(101, 102), "((a))"),
			PState{input: "", pos: 5, value: u8.Text("a"), ok: true})

		tt.AssertEqual(parseStr(globalExprWithParams(101, 102), "a"),
			PState{input: "", pos: 1, value: u8.Text("a"), ok: true})
		tt.AssertEqual(parseStr(globalExprWithParams(101, 102), "(a)"),
			PState{input: "", pos: 3, value: u8.Text("a"), ok: true})
		tt.AssertEqual(parseStr(globalExprWithParams(101, 102), "((a))"),
			PState{input: "", pos: 5, value: u8.Text("a"), ok: true})
	})
}

//================================================================================
//Try, NotFollowedBy, Eof
func TestBacktracking(t *testing.T) {
	ts.LogAsserts("Backtracking", t, func(tt *ts.T) {
		letter := Regexp(`\pL`)
		digit := Regexp(`\p{Nd}`)
		lower := Regexp(`\p{Ll}`)

		tt.AssertEqual(parseStr(Try(letter), "defg"),
			PState{input: "efg", pos: 1, value: u8.Text("d"), ok: true})
		tt.AssertEqual(parseStr(Try(letter), ";efg"),
			PState{input: ";efg", empty: true, error: makeUnexpInp(";")})

		p2 := Bind(digit, func(c1 interface{}) Parser {
			return Bind(lower, func(c2 interface{}) Parser {
				return Return([]interface{}{c1, c2})
			})
		})
		tt.AssertEqual(parseStr(Try(p2), "67fg"),
			PState{input: "67fg", empty: true, error: makeUnexpInp("7")})

		tt.AssertEqual(parseStr(Try(SeqRight(digit, lower)), "66efg"),
			PState{input: "66efg", empty: true, error: makeUnexpInp("6")})

		tt.AssertEqual(parseStr(SeqLeft(digit, NotFollowedBy(letter)), "66efg"),
			PState{input: "6efg", pos: 1, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(SeqRight(digit, NotFollowedBy(letter)), "66efg"),
			PState{input: "6efg", pos: 1, ok: true})
		tt.AssertEqual(parseStr(SeqLeft(digit, NotFollowedBy(digit)), "66efg"),
			PState{input: "6efg", pos: 1, error: errParsFail})
		tt.AssertEqual(parseStr(Try(SeqLeft(digit, NotFollowedBy(digit))), "66efg"),
			PState{input: "66efg", empty: true, error: errParsFail})

		tt.AssertEqual(parseStr(Eof, ""),
			PState{input: "", ok: true, empty: true})
		tt.AssertEqual(parseStr(Eof, "abc"),
			PState{input: "abc", empty: true, error: errParsFail})
	})
}

//================================================================================
//Ask, Expect
func TestAskExpect(t *testing.T) {
	ts.LogAsserts("AskExpect", t, func(tt *ts.T) {
		letter := Regexp(`\pL`)
		digit := Regexp(`\p{Nd}`)

		tt.AssertEqual(parseStr(Ask(Collect(digit, letter), u8.Text("digit,letter")), ";efg"),
			PState{input: ";efg", empty: true, error: "Unexpected ; input. Expecting digit,letter"})
		tt.AssertEqual(parseStr(Ask(Collect(digit, letter), u8.Text("digit,letter")), "01"),
			PState{input: "1", pos: 1, error: "Unexpected 1 input."})
		tt.AssertEqual(parseStr(Ask(Collect(Ask(digit, u8.Text("digit")), Ask(letter, u8.Text("letter"))), u8.Text("digit,letter")), "01"),
			PState{input: "1", pos: 1, error: "Unexpected 1 input. Expecting letter"})
		tt.AssertEqual(parseStr(Ask(Collect(digit, letter), u8.Text("digit,letter")), "0"),
			PState{input: "", pos: 1, error: "Unexpected end-of-file reached."})
		tt.AssertEqual(parseStr(Ask(Collect(Ask(digit, u8.Text("digit")), Ask(letter, u8.Text("letter"))), u8.Text("digit,letter")), "0"),
			PState{input: "", pos: 1, error: "Unexpected end-of-file reached. Expecting letter"})

		tt.AssertEqual(parseStr(Expect(Collect(letter, digit), u8.Text("number two")), "U2"),
			PState{input: "", pos: 2, value: arrText("U", "2"), ok: true})
		tt.AssertEqual(parseStr(Expect(Collect(letter, digit), u8.Text("number two")), "UX"),
			PState{input: "X", pos: 1, error: "Unexpected X input. Expecting number two"})
		tt.AssertEqual(parseStr(Expect(Collect(letter, digit), u8.Text("number two")), "007"),
			PState{input: "007", empty: true, error: "Unexpected 0 input. Expecting number two"})
	})
}

//================================================================================
//Alt
func TestAlternation(t *testing.T) {
	ts.LogAsserts("Alternation", t, func(tt *ts.T) {
		letter := Regexp(`\pL`)
		digit := Regexp(`\p{Nd}`)
		lower := Regexp(`\p{Ll}`)
		upper := Regexp(`\p{Lu}`)

		tt.AssertEqual(parseStr(Alt(digit, letter), "defg"),
			PState{input: "efg", pos: 1, value: u8.Text("d"), ok: true})
		tt.AssertEqual(parseStr(Alt(digit, letter), "6efg"),
			PState{input: "efg", pos: 1, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(Alt(digit, letter), ";efg"),
			PState{input: ";efg", empty: true, error: errNoAlts})

		tt.AssertEqual(parseStr(Alt(letter), "defg"),
			PState{input: "efg", pos: 1, value: u8.Text("d"), ok: true})
		tt.AssertEqual(parseStr(Alt(letter), ";efg"),
			PState{input: ";efg", empty: true, error: makeUnexpInp(";")})

		tt.AssertEqual(parseStr(Alt(), "defg"),
			PState{input: "defg", empty: true, error: errNoParser})

		tt.AssertEqual(parseStr(Alt(digit, upper, lower), "defg"),
			PState{input: "efg", pos: 1, value: u8.Text("d"), ok: true})
		tt.AssertEqual(parseStr(Alt(digit, upper, lower), "6efg"),
			PState{input: "efg", pos: 1, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(Alt(digit, upper, lower), "Defg"),
			PState{input: "efg", pos: 1, value: u8.Text("D"), ok: true})
		tt.AssertEqual(parseStr(Alt(digit, upper, lower), ";efg"),
			PState{input: ";efg", empty: true, error: errNoAlts})
	})
}

//================================================================================
//Bind, SeqLeft, SeqRight
func TestSequencing(t *testing.T) {
	ts.LogAsserts("Sequencing", t, func(tt *ts.T) {
		digit := Regexp(`\p{Nd}`)
		intNum := Regexp(`\p{Nd}*`)
		lower := Regexp(`\p{Ll}`)

		singleLower := Satisfy(func(c u8.Codepoint) bool {
			cc := u8.SurrogatePoint(c)
			return len(cc) == 1 && unicode.IsLower(cc[0])
		})
		singleDigit := Satisfy(func(c u8.Codepoint) bool {
			cc := u8.SurrogatePoint(c)
			return len(cc) == 1 && unicode.IsDigit(cc[0]) //same as p{Nd} ???
		})
		singleUpper := Satisfy(func(c u8.Codepoint) bool {
			cc := u8.SurrogatePoint(c)
			return len(cc) == 1 && unicode.IsUpper(cc[0])
		})

		p1 := Bind(digit, func(c interface{}) Parser {
			return lower
		})
		tt.AssertEqual(parseStr(p1, "6efg"), PState{input: "fg", pos: 2, value: u8.Text("e"), ok: true})

		p2 := Bind(singleDigit, func(c1 interface{}) Parser {
			return Bind(singleLower, func(c2 interface{}) Parser {
				return Return([]interface{}{c1, c2})
			})
		})
		tt.AssertEqual(parseStr(p2, "6efg"), PState{input: "fg", pos: 2, value: arrCp('6', 'e'), ok: true})
		tt.AssertEqual(parseStr(p2, "67fg"), PState{input: "7fg", pos: 1, error: makeUnexpInp("7")})

		p3 := Bind(singleDigit, func(c1 interface{}) Parser {
			return Bind(singleLower, func(c2 interface{}) Parser {
				return Bind(singleUpper, func(c3 interface{}) Parser {
					return Return([]interface{}{c1, c2, c3})
				})
			})
		})
		tt.AssertEqual(parseStr(p3, "6eFghij"),
			PState{input: "ghij", pos: 3, value: arrCp('6', 'e', 'F'), ok: true})

		tt.AssertEqual(parseStr(SeqLeft(), "6efg"),
			PState{input: "6efg", empty: true, error: errNoParser})
		tt.AssertEqual(parseStr(SeqLeft(digit), "efg"),
			PState{input: "efg", empty: true, error: makeUnexpInp("e")})
		tt.AssertEqual(parseStr(SeqLeft(digit), "6efg"),
			PState{input: "efg", pos: 1, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(SeqLeft(digit, lower), "efg"),
			PState{input: "efg", empty: true, error: makeUnexpInp("e")})
		tt.AssertEqual(parseStr(SeqLeft(digit, lower), "66efg"),
			PState{input: "6efg", pos: 1, error: makeUnexpInp("6")})
		tt.AssertEqual(parseStr(SeqLeft(digit, lower), "6efg"),
			PState{input: "fg", pos: 2, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(SeqLeft(intNum, lower), "678efg"),
			PState{input: "fg", pos: 4, value: u8.Text("678"), ok: true})
		tt.AssertEqual(parseStr(SeqLeft(digit, lower, lower), "6efg"),
			PState{input: "g", pos: 3, value: u8.Text("6"), ok: true})

		tt.AssertEqual(parseStr(SeqRight(), "6efg"),
			PState{input: "6efg", empty: true, error: errNoParser})
		tt.AssertEqual(parseStr(SeqRight(digit), "efg"),
			PState{input: "efg", empty: true, error: makeUnexpInp("e")})
		tt.AssertEqual(parseStr(SeqRight(digit), "6efg"),
			PState{input: "efg", pos: 1, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(SeqRight(digit, lower), "efg"),
			PState{input: "efg", empty: true, error: makeUnexpInp("e")})
		tt.AssertEqual(parseStr(SeqRight(digit, lower), "66efg"),
			PState{input: "6efg", pos: 1, error: makeUnexpInp("6")})
		tt.AssertEqual(parseStr(SeqRight(digit, lower), "6efg"),
			PState{input: "fg", pos: 2, value: u8.Text("e"), ok: true})
		tt.AssertEqual(parseStr(SeqRight(digit, lower, lower), "6efg"),
			PState{input: "g", pos: 3, value: u8.Text("f"), ok: true})
	})
}

//================================================================================
//Ask
func TestAsk(t *testing.T) {
	ts.LogAsserts("Ask", t, func(tt *ts.T) {
		letter := Ask(Regexp(`\pL`), u8.Desur("letter"))
		digit := Ask(Regexp(`\p{Nd}`), u8.Desur("digit"))
		lower := Ask(Regexp(`\p{Ll}`), u8.Desur("lower"))
		upper := Ask(Regexp(`\p{Lu}`), u8.Desur("upper"))
		intNum := Ask(Regexp(`\p{Nd}*`), u8.Desur("integer number"))
		singleLower := Ask(Satisfy(func(c u8.Codepoint) bool {
			cc := u8.SurrogatePoint(c)
			return len(cc) == 1 && unicode.IsLower(cc[0])
		}), u8.Desur("single lower"))
		singleDigit := Ask(Satisfy(func(c u8.Codepoint) bool {
			cc := u8.SurrogatePoint(c)
			return len(cc) == 1 && unicode.IsDigit(cc[0])
		}), u8.Desur("single digit"))
		singleUpper := Ask(Satisfy(func(c u8.Codepoint) bool {
			cc := u8.SurrogatePoint(c)
			return len(cc) == 1 && unicode.IsUpper(cc[0])
		}), u8.Desur("single upper"))

		//Try...
		tt.AssertEqual(parseStr(Try(letter), ";efg"),
			PState{input: ";efg", empty: true, error: "Unexpected ; input. Expecting letter"})
		pp2 := Bind(digit, func(c1 interface{}) Parser {
			return Bind(lower, func(c2 interface{}) Parser {
				return Return([]interface{}{c1, c2})
			})
		})
		tt.AssertEqual(parseStr(Try(pp2), "67fg"),
			PState{input: "67fg", empty: true, error: "Unexpected 7 input. Expecting lower"})
		tt.AssertEqual(parseStr(Try(SeqRight(digit, lower)), "66efg"),
			PState{input: "66efg", empty: true, error: "Unexpected 6 input. Expecting lower"})

		//NotFollowedBy...
		tt.AssertEqual(parseStr(SeqLeft(digit, NotFollowedBy(letter)), "66efg"),
			PState{input: "6efg", pos: 1, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(SeqRight(digit, NotFollowedBy(letter)), "66efg"),
			PState{input: "6efg", pos: 1, ok: true})
		tt.AssertEqual(parseStr(SeqLeft(digit, NotFollowedBy(digit)), "66efg"),
			PState{input: "6efg", pos: 1, error: errParsFail})
		tt.AssertEqual(parseStr(Try(SeqLeft(digit, NotFollowedBy(digit))), "66efg"),
			PState{input: "66efg", empty: true, error: errParsFail})

		//Alt...
		tt.AssertEqual(parseStr(Alt(digit, letter), "defg"),
			PState{input: "efg", pos: 1, value: u8.Text("d"), ok: true})
		tt.AssertEqual(parseStr(Alt(digit, letter), "6efg"),
			PState{input: "efg", pos: 1, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(Alt(digit, letter), ";efg"),
			PState{input: ";efg", empty: true, error: errNoAlts})

		tt.AssertEqual(parseStr(Alt(letter), "defg"),
			PState{input: "efg", pos: 1, value: u8.Text("d"), ok: true})
		tt.AssertEqual(parseStr(Alt(letter), ";efg"),
			PState{input: ";efg", empty: true, error: "Unexpected ; input. Expecting letter"})

		tt.AssertEqual(parseStr(Alt(), "defg"),
			PState{input: "defg", empty: true, error: errNoParser})

		tt.AssertEqual(parseStr(Alt(digit, upper, lower), "defg"),
			PState{input: "efg", pos: 1, value: u8.Text("d"), ok: true})
		tt.AssertEqual(parseStr(Alt(digit, upper, lower), "6efg"),
			PState{input: "efg", pos: 1, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(Alt(digit, upper, lower), "Defg"),
			PState{input: "efg", pos: 1, value: u8.Text("D"), ok: true})
		tt.AssertEqual(parseStr(Alt(digit, upper, lower), ";efg"),
			PState{input: ";efg", pos: 0, empty: true, error: errNoAlts})

		//Bind, SeqLeft, SeqRight... (#18)
		p1 := Bind(digit, func(c interface{}) Parser {
			return lower
		})
		tt.AssertEqual(parseStr(p1, "6efg"), PState{input: "fg", pos: 2, value: u8.Text("e"), ok: true})

		p2 := Bind(singleDigit, func(c1 interface{}) Parser {
			return Bind(singleLower, func(c2 interface{}) Parser {
				return Return([]interface{}{c1, c2})
			})
		})
		tt.AssertEqual(parseStr(p2, "6efg"), PState{input: "fg", pos: 2, value: arrCp('6', 'e'), ok: true})
		tt.AssertEqual(parseStr(p2, "67fg"), PState{input: "7fg", pos: 1, error: "Unexpected 7 input. Expecting single lower"})

		p3 := Bind(singleDigit, func(c1 interface{}) Parser {
			return Bind(singleLower, func(c2 interface{}) Parser {
				return Bind(singleUpper, func(c3 interface{}) Parser {
					return Return([]interface{}{c1, c2, c3})
				})
			})
		})
		tt.AssertEqual(parseStr(p3, "6eFghij"),
			PState{input: "ghij", pos: 3, value: arrCp('6', 'e', 'F'), ok: true})

		tt.AssertEqual(parseStr(SeqLeft(), "6efg"),
			PState{input: "6efg", empty: true, error: errNoParser})
		tt.AssertEqual(parseStr(SeqLeft(digit), "efg"),
			PState{input: "efg", empty: true, error: "Unexpected e input. Expecting digit"})
		tt.AssertEqual(parseStr(SeqLeft(digit), "6efg"),
			PState{input: "efg", pos: 1, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(SeqLeft(digit, lower), "efg"),
			PState{input: "efg", empty: true, error: "Unexpected e input. Expecting digit"})
		tt.AssertEqual(parseStr(SeqLeft(digit, lower), "66efg"),
			PState{input: "6efg", pos: 1, error: "Unexpected 6 input. Expecting lower"})
		tt.AssertEqual(parseStr(SeqLeft(digit, lower), "6efg"),
			PState{input: "fg", pos: 2, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(SeqLeft(intNum, lower), "678efg"),
			PState{input: "fg", pos: 4, value: u8.Text("678"), ok: true})
		tt.AssertEqual(parseStr(SeqLeft(digit, lower, lower), "6efg"),
			PState{input: "g", pos: 3, value: u8.Text("6"), ok: true})

		tt.AssertEqual(parseStr(SeqRight(), "6efg"),
			PState{input: "6efg", empty: true, error: errNoParser})
		tt.AssertEqual(parseStr(SeqRight(digit), "efg"),
			PState{input: "efg", empty: true, error: "Unexpected e input. Expecting digit"})
		tt.AssertEqual(parseStr(SeqRight(digit), "6efg"),
			PState{input: "efg", pos: 1, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(SeqRight(digit, lower), "efg"),
			PState{input: "efg", empty: true, error: "Unexpected e input. Expecting digit"})
		tt.AssertEqual(parseStr(SeqRight(digit, lower), "66efg"),
			PState{input: "6efg", pos: 1, error: "Unexpected 6 input. Expecting lower"})
		tt.AssertEqual(parseStr(SeqRight(digit, lower), "6efg"),
			PState{input: "fg", pos: 2, value: u8.Text("e"), ok: true})
		tt.AssertEqual(parseStr(SeqRight(digit, lower, lower), "6efg"),
			PState{input: "g", pos: 3, value: u8.Text("f"), ok: true})
	})
}

//================================================================================
//Apply, Between
func TestApply(t *testing.T) {
	ts.LogAsserts("Apply", t, func(tt *ts.T) {
		letter := Regexp(`\pL`)
		digit := Regexp(`\p{Nd}`)

		p1 := Apply(func(c interface{}) interface{} {
			cs := u8.Surr(c.(u8.Text))
			return cs + cs + cs
		}, letter)
		tt.AssertEqual(parseStr(p1, "abc"),
			PState{input: "bc", pos: 1, value: "aaa", ok: true})

		tt.AssertEqual(parseStr(Between(letter, letter, digit), "a6zz7"),
			PState{input: "z7", pos: 3, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(Between(letter, letter, digit), "a67yx"),
			PState{input: "a67yx", empty: true, error: makeUnexpInp("7")})
		tt.AssertEqual(parseStr(Between(letter, digit, Return(nil)), "a6zz7"),
			PState{input: "zz7", pos: 2, ok: true})
	})
}

//================================================================================
//Collect, Flatten
func TestCollectEtc(t *testing.T) {
	ts.LogAsserts("CollectEtc", t, func(tt *ts.T) {
		letter := Regexp(`\pL`)
		digit := Regexp(`\p{Nd}`)

		tt.AssertEqual(parseStr(Collect(letter, digit, letter), "a6cde"),
			PState{input: "de", pos: 3, value: arrText("a", "6", "c"), ok: true})
		tt.AssertEqual(parseStr(Collect(letter, Token(u8.Text("bb")), letter), "abbacde"),
			PState{input: "cde", pos: 4, value: arrText("a", "bb", "a"), ok: true})
		tt.AssertEqual(parseStr(Collect(letter), "abcde"),
			PState{input: "bcde", pos: 1, value: arrText("a"), ok: true})
		tt.AssertEqual(parseStr(Collect(), "abcde"),
			PState{input: "abcde", value: []interface{}{}, ok: true, empty: true})

		tt.AssertEqual(parseStr(Collect(Token(u8.Text("-")), digit), "-1"),
			PState{input: "", pos: 2, value: arrText("-", "1"), ok: true})
		tt.AssertEqual(parseStr(Collect(Token(u8.Text("-")), digit, Token(u8.Text(";"))), "-1;"),
			PState{input: "", pos: 3, value: arrText("-", "1", ";"), ok: true})

		tt.AssertEqual(parseStr(Collect(Collect(letter, letter, letter), Collect(digit, digit, digit)), "ABC012"),
			PState{input: "", pos: 6, value: []interface{}{arrText("A", "B", "C"), arrText("0", "1", "2")}, ok: true})
		tt.AssertEqual(parseStr(Collect(letter, Collect(digit, digit, digit)), "A012"),
			PState{input: "", pos: 4, value: []interface{}{u8.Text("A"), arrText("0", "1", "2")}, ok: true})
		tt.AssertEqual(parseStr(Collect(Collect(letter, letter, letter), digit), "ABC2"),
			PState{input: "", pos: 4, value: []interface{}{arrText("A", "B", "C"), u8.Text("2")}, ok: true})

		o1 := Collect(Token(u8.Text("A")), Token(u8.Text("A")), Token(u8.Text("A")))
		tt.AssertEqual(parseStr(Collect(o1, o1), "AAAAAAA"),
			PState{input: "A", pos: 6, value: []interface{}{arrText("A", "A", "A"), arrText("A", "A", "A")}, ok: true})

		p1 := Collect(letter, letter, letter)
		tt.AssertEqual(parseStr(Collect(p1, p1), "ABCDEFG"),
			PState{input: "G", pos: 6, value: []interface{}{arrText("A", "B", "C"), arrText("D", "E", "F")}, ok: true})

		tt.AssertEqual(parseStr(Flatten(Token(u8.Text("-")), digit), "-1"),
			PState{input: "", pos: 2, value: u8.Text("-1"), ok: true})
		tt.AssertEqual(parseStr(Flatten(Token(u8.Text("-")), digit, Token(u8.Text(";"))), "-1;"),
			PState{input: "", pos: 3, value: u8.Text("-1;"), ok: true})
		tt.AssertEqual(parseStr(Flatten(letter, SeqRight(Token(u8.Text("|")), letter), SeqRight(Token(u8.Text("|")), digit)), "X|Y|9"),
			PState{input: "", pos: 5, value: u8.Text("XY9"), ok: true})
		tt.AssertEqual(parseStr(Flatten(Collect(letter, letter, letter), Collect(digit, digit, digit)), "ABC012"),
			PState{input: "", pos: 6, value: u8.Text("ABC012"), ok: true})
		tt.AssertEqual(parseStr(Flatten(letter, digit), "*"),
			PState{input: "*", pos: 0, value: nil, empty: true, error: makeUnexpInp("*")})
		tt.AssertEqual(parseStr(Flatten(letter, digit), "A*"),
			PState{input: "*", pos: 1, value: nil, error: makeUnexpInp("*")})
		tt.AssertEqual(parseStr(Flatten(letter, Token(u8.Text("\t")), digit), "A\t*"),
			PState{input: "*", pos: 2, value: nil, error: makeUnexpInp("*")})
		tt.AssertEqual(parseStr(Flatten(letter, Token(u8.Text("\t")), Token(u8.Text("\t")), Token(u8.Text("\t")), digit), "A\t\t\t*"),
			PState{input: "*", pos: 4, value: nil, error: makeUnexpInp("*")})

		tt.AssertEqual(parseStr(Flatten(letter, letter, letter), "ABCDEFG"),
			PState{input: "DEFG", pos: 3, value: u8.Text("ABC"), ok: true})

		q1 := Flatten(letter, letter, letter)
		q2 := Flatten(letter, letter, letter)
		tt.AssertEqual(parseStr(Collect(q1, q2), "ABCDEFG"),
			PState{input: "G", pos: 6, value: arrText("ABC", "DEF"), ok: true})
	})
}

//================================================================================
//Many, Many1, Optional, Option
func TestMany(t *testing.T) {
	ts.LogAsserts("Many", t, func(tt *ts.T) {
		letter := Regexp(`\pL`)
		singleLetter := Satisfy(func(c u8.Codepoint) bool {
			cc := u8.SurrogatePoint(c)
			return len(cc) == 1 && unicode.IsLetter(cc[0])
		})

		tt.AssertEqual(parseStr(Many(singleLetter), "def;ghij"),
			PState{input: ";ghij", pos: 3, value: arrCp('d', 'e', 'f'), ok: true})
		tt.AssertEqual(parseStr(Many(singleLetter), "d6ef;ghij"),
			PState{input: "6ef;ghij", pos: 1, value: arrCp('d'), ok: true})
		tt.AssertEqual(parseStr(Many(letter), "6def;ghij"),
			PState{input: "6def;ghij", value: []interface{}{}, ok: true, empty: true})

		tt.AssertEqual(parseStr(Many1(singleLetter), "def;ghij"),
			PState{input: ";ghij", pos: 3, value: arrCp('d', 'e', 'f'), ok: true})
		tt.AssertEqual(parseStr(Many1(singleLetter), "d6ef;ghij"),
			PState{input: "6ef;ghij", pos: 1, value: arrCp('d'), ok: true})
		tt.AssertEqual(parseStr(Many1(letter), "6def;ghij"),
			PState{input: "6def;ghij", empty: true, error: makeUnexpInp("6")})

		tt.AssertEqual(parseStr(Optional(letter), "def;ghij"),
			PState{input: "ef;ghij", pos: 1, value: u8.Text("d"), ok: true})
		tt.AssertEqual(parseStr(Optional(letter), "6def;ghij"),
			PState{input: "6def;ghij", ok: true, empty: true})

		tt.AssertEqual(parseStr(Option('z', letter), "def;ghij"),
			PState{input: "ef;ghij", pos: 1, value: u8.Text("d"), ok: true})
		tt.AssertEqual(parseStr(Option('z', letter), "6def;ghij"),
			PState{input: "6def;ghij", value: 'z', ok: true, empty: true})
	})
}

//================================================================================
//Times
func TestTimes(t *testing.T) {
	ts.LogAsserts("Times", t, func(tt *ts.T) {
		letter := Regexp(`\pL`)
		digit := Regexp(`\p{Nd}`)

		tt.AssertEqual(parseStr(Times(0, digit), "0*"),
			PState{input: "0*", pos: 0, value: []interface{}{}, ok: true, empty: true})
		tt.AssertEqual(parseStr(Times(3, SeqRight(AnyChar, letter)), "*a@b$c"),
			PState{input: "", pos: 6, value: arrText("a", "b", "c"), ok: true, empty: false})
		tt.AssertEqual(parseStr(Times(3, SeqRight(AnyChar, letter)), "*a@b$c"),
			PState{input: "", pos: 6, value: arrText("a", "b", "c"), ok: true, empty: false})
		tt.AssertEqual(parseStr(Times(3, SeqRight(AnyChar, letter)), "*a@b$$"),
			PState{input: "$", pos: 5, value: nil, ok: false, empty: false, error: "Unexpected $ input."})
	})
}

//================================================================================
//ManyTill
func TestManyTill(t *testing.T) {
	ts.LogAsserts("ManyTill", t, func(tt *ts.T) {
		letter := Regexp(`\pL`)
		digit := Regexp(`\p{Nd}`)
		upper := Regexp(`\p{Lu}`)

		tt.AssertEqual(parseStr(ManyTill(digit, letter), "1234A"),
			PState{input: "", pos: 5, value: arrText("1", "2", "3", "4"), ok: true})
		tt.AssertEqual(parseStr(ManyTill(digit, letter), "A*"),
			PState{input: "*", pos: 1, value: []interface{}{}, ok: true})
		tt.AssertEqual(parseStr(SeqRight(Token(u8.Text("<!--")), ManyTill(AnyChar, Try(Token(u8.Text("-->"))))), "<!-- -->"),
			PState{input: "", pos: 8, value: arrCp(' '), ok: true})
		tt.AssertEqual(parseStr(SeqRight(Token(u8.Text("<!--")), ManyTill(AnyChar, Try(Token(u8.Text("-->"))))), "<!--foo-->"),
			PState{input: "", pos: 10, value: arrCp('f', 'o', 'o'), ok: true})
		tt.AssertEqual(parseStr(ManyTill(digit, letter), "*A"),
			PState{input: "*A", pos: 0, value: nil, empty: true, error: errNoAlts})
		tt.AssertEqual(parseStr(ManyTill(digit, letter), "123*A"),
			PState{input: "*A", pos: 3, value: nil, error: errNoAlts})
		tt.AssertEqual(parseStr(ManyTill(digit, SeqRight(upper, Token(u8.Text("X")))), "123A*"),
			PState{input: "*", pos: 4, value: nil, error: errNoAlts})
	})
}

//================================================================================
//SepBy, SepBy1, EndBy, EndBy1, SepEndBy, SepEndBy1
func TestSeparators(t *testing.T) {
	ts.LogAsserts("Separators", t, func(tt *ts.T) {
		digit := Regexp(`\p{Nd}`)
		letterA := Satisfy(func(c u8.Codepoint) bool { return c == 'a' })

		tt.AssertEqual(parseStr(SepBy(digit, letterA), "bc"),
			PState{input: "bc", value: []interface{}{}, ok: true, empty: true})
		tt.AssertEqual(parseStr(SepBy(digit, letterA), "abc"),
			PState{input: "bc", pos: 1, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(SepBy(digit, letterA), "a6bc"),
			PState{input: "6bc", pos: 1, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(SepBy(digit, letterA), "a6abc"),
			PState{input: "bc", pos: 3, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(SepBy(digit, letterA), "a6a6bc"),
			PState{input: "6bc", pos: 3, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(SepBy(digit, letterA), "a6a6abc"),
			PState{input: "bc", pos: 5, value: arrCp('a', 'a', 'a'), ok: true})

		tt.AssertEqual(parseStr(SepBy1(digit, letterA), "bc"),
			PState{input: "bc", empty: true, error: makeUnexpInp("b")})
		tt.AssertEqual(parseStr(SepBy1(digit, letterA), "abc"),
			PState{input: "bc", pos: 1, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(SepBy1(digit, letterA), "a6bc"),
			PState{input: "6bc", pos: 1, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(SepBy1(digit, letterA), "a6abc"),
			PState{input: "bc", pos: 3, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(SepBy1(digit, letterA), "a6a6bc"),
			PState{input: "6bc", pos: 3, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(SepBy1(digit, letterA), "a6a6abc"),
			PState{input: "bc", pos: 5, value: arrCp('a', 'a', 'a'), ok: true})

		tt.AssertEqual(parseStr(EndBy(digit, letterA), "bc"),
			PState{input: "bc", value: []interface{}{}, ok: true, empty: true})
		tt.AssertEqual(parseStr(EndBy(digit, letterA), "abc"),
			PState{input: "abc", value: []interface{}{}, ok: true, empty: true})
		tt.AssertEqual(parseStr(EndBy(digit, letterA), "a6bc"),
			PState{input: "bc", pos: 2, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(EndBy(digit, letterA), "a6abc"),
			PState{input: "abc", pos: 2, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(EndBy(digit, letterA), "a6a6bc"),
			PState{input: "bc", pos: 4, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(EndBy(digit, letterA), "a6a6abc"),
			PState{input: "abc", pos: 4, value: arrCp('a', 'a'), ok: true})

		tt.AssertEqual(parseStr(EndBy1(digit, letterA), "bc"),
			PState{input: "bc", empty: true, error: makeUnexpInp("b")})
		tt.AssertEqual(parseStr(EndBy1(digit, letterA), "abc"),
			PState{input: "abc", empty: true, error: makeUnexpInp("b")})
		tt.AssertEqual(parseStr(EndBy1(digit, letterA), "a6bc"),
			PState{input: "bc", pos: 2, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(EndBy1(digit, letterA), "a6abc"),
			PState{input: "abc", pos: 2, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(EndBy1(digit, letterA), "a6a6bc"),
			PState{input: "bc", pos: 4, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(EndBy1(digit, letterA), "a6a6abc"),
			PState{input: "abc", pos: 4, value: arrCp('a', 'a'), ok: true})

		tt.AssertEqual(parseStr(SepEndBy(digit, letterA), "bc"),
			PState{input: "bc", value: []interface{}{}, ok: true, empty: true})
		tt.AssertEqual(parseStr(SepEndBy(digit, letterA), "abc"),
			PState{input: "bc", pos: 1, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(digit, letterA), "a6bc"),
			PState{input: "bc", pos: 2, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(digit, letterA), "a6abc"),
			PState{input: "bc", pos: 3, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(digit, letterA), "a6a6bc"),
			PState{input: "bc", pos: 4, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(digit, letterA), "a6a6abc"),
			PState{input: "bc", pos: 5, value: arrCp('a', 'a', 'a'), ok: true})

		tt.AssertEqual(parseStr(SepEndBy1(digit, letterA), "bc"),
			PState{input: "bc", empty: true, error: makeUnexpInp("b")})
		tt.AssertEqual(parseStr(SepEndBy1(digit, letterA), "abc"),
			PState{input: "bc", pos: 1, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy1(digit, letterA), "a6bc"),
			PState{input: "bc", pos: 2, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy1(digit, letterA), "a6abc"),
			PState{input: "bc", pos: 3, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy1(digit, letterA), "a6a6bc"),
			PState{input: "bc", pos: 4, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy1(digit, letterA), "a6a6abc"),
			PState{input: "bc", pos: 5, value: arrCp('a', 'a', 'a'), ok: true})

		tt.AssertEqual(parseStr(SepEndBy(Many(digit), letterA), "bc"),
			PState{input: "bc", value: []interface{}{}, ok: true, empty: true})
		tt.AssertEqual(parseStr(SepEndBy(Many(digit), letterA), "abc"),
			PState{input: "bc", pos: 1, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(Many(digit), letterA), "a6bc"),
			PState{input: "bc", pos: 2, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(Many(digit), letterA), "a666bc"),
			PState{input: "bc", pos: 4, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(Many(digit), letterA), "aabc"),
			PState{input: "bc", pos: 2, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(Many(digit), letterA), "a6abc"),
			PState{input: "bc", pos: 3, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(Many(digit), letterA), "a6a6bc"),
			PState{input: "bc", pos: 4, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(Many(digit), letterA), "a6a666bc"),
			PState{input: "bc", pos: 6, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(Many(digit), letterA), "a6a6abc"),
			PState{input: "bc", pos: 5, value: arrCp('a', 'a', 'a'), ok: true})

		tt.AssertEqual(parseStr(SepEndBy(Many1(digit), letterA), "bc"),
			PState{input: "bc", value: []interface{}{}, ok: true, empty: true})
		tt.AssertEqual(parseStr(SepEndBy(Many1(digit), letterA), "abc"),
			PState{input: "bc", pos: 1, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(Many1(digit), letterA), "a6bc"),
			PState{input: "bc", pos: 2, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(Many1(digit), letterA), "a666bc"),
			PState{input: "bc", pos: 4, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(Many1(digit), letterA), "aabc"),
			PState{input: "abc", pos: 1, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(Many1(digit), letterA), "a6abc"),
			PState{input: "bc", pos: 3, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(Many1(digit), letterA), "a6a6bc"),
			PState{input: "bc", pos: 4, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(Many1(digit), letterA), "a6a666bc"),
			PState{input: "bc", pos: 6, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(SepEndBy(Many1(digit), letterA), "a6a6abc"),
			PState{input: "bc", pos: 5, value: arrCp('a', 'a', 'a'), ok: true})
	})
}

//================================================================================
//BegSepBy, BegSepBy1
func TestBegSeparators(t *testing.T) {
	ts.LogAsserts("BegSeparators", t, func(tt *ts.T) {
		digit := Regexp(`\p{Nd}`)
		letterA := Satisfy(func(c u8.Codepoint) bool { return c == 'a' })

		tt.AssertEqual(parseStr(BegSepBy(digit, letterA), "bc"),
			PState{input: "bc", value: []interface{}{}, ok: true, empty: true})
		tt.AssertEqual(parseStr(BegSepBy(digit, letterA), "abc"),
			PState{input: "abc", value: []interface{}{}, ok: true, empty: true})
		tt.AssertEqual(parseStr(BegSepBy(digit, letterA), "a6bc"),
			PState{input: "a6bc", value: []interface{}{}, ok: true, empty: true})
		tt.AssertEqual(parseStr(BegSepBy(digit, letterA), "6bc"),
			PState{input: "bc", pos: 1, value: []interface{}{}, ok: true})
		tt.AssertEqual(parseStr(BegSepBy(digit, letterA), "6abc"),
			PState{input: "bc", pos: 2, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(BegSepBy(digit, letterA), "6a6bc"),
			PState{input: "6bc", pos: 2, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(BegSepBy(digit, letterA), "6a6abc"),
			PState{input: "bc", pos: 4, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(BegSepBy(digit, letterA), "6a6a6bc"),
			PState{input: "6bc", pos: 4, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(BegSepBy(digit, letterA), "6a6a6abc"),
			PState{input: "bc", pos: 6, value: arrCp('a', 'a', 'a'), ok: true})

		tt.AssertEqual(parseStr(BegSepBy1(digit, letterA), "bc"),
			PState{input: "bc", value: []interface{}{}, ok: true, empty: true})
		tt.AssertEqual(parseStr(BegSepBy1(digit, letterA), "abc"),
			PState{input: "abc", value: []interface{}{}, ok: true, empty: true})
		tt.AssertEqual(parseStr(BegSepBy1(digit, letterA), "a6bc"),
			PState{input: "a6bc", value: []interface{}{}, ok: true, empty: true})
		tt.AssertEqual(parseStr(BegSepBy1(digit, letterA), "6bc"),
			PState{input: "6bc", value: []interface{}{}, ok: true, empty: true})
		tt.AssertEqual(parseStr(BegSepBy1(digit, letterA), "6abc"),
			PState{input: "bc", pos: 2, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(BegSepBy1(digit, letterA), "6a6bc"),
			PState{input: "6bc", pos: 2, value: arrCp('a'), ok: true})
		tt.AssertEqual(parseStr(BegSepBy1(digit, letterA), "6a6abc"),
			PState{input: "bc", pos: 4, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(BegSepBy1(digit, letterA), "6a6a6bc"),
			PState{input: "6bc", pos: 4, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(BegSepBy1(digit, letterA), "6a6a6abc"),
			PState{input: "bc", pos: 6, value: arrCp('a', 'a', 'a'), ok: true})
	})
}

//================================================================================
//Skip
func TestSkip(t *testing.T) {
	ts.LogAsserts("BegSkip", t, func(tt *ts.T) {
		letter := Regexp(`\pL`)
		digit := Regexp(`\p{Nd}`)
		lower := Regexp(`\p{Ll}`)

		tt.AssertEqual(parseStr(Skip(Token(u8.Desur("*"))), "*"),
			PState{input: "", pos: 1, value: nil, ok: true})
		tt.AssertEqual(parseStr(Skip(letter, digit), "U2"),
			PState{input: "", pos: 2, value: nil, ok: true})
		tt.AssertEqual(parseStr(Skip(Token(u8.Desur("*")), letter, digit, Token(u8.Desur("*"))), "*U2*"),
			PState{input: "", pos: 4, value: nil, ok: true})

		tt.AssertEqual(parseStr(SkipMany(letter), "*"),
			PState{input: "*", pos: 0, value: nil, ok: true, empty: true})
		tt.AssertEqual(parseStr(SkipMany(Collect(letter, digit)), "A*"),
			PState{input: "*", pos: 1, value: nil, error: makeUnexpInp("*")})
		tt.AssertEqual(parseStr(SkipMany(Collect(digit, lower)), "0x*"),
			PState{input: "*", pos: 2, value: nil, ok: true})
		tt.AssertEqual(parseStr(SkipMany(letter), "abcdefghijk*"),
			PState{input: "*", pos: 11, value: nil, ok: true})
		tt.AssertEqual(parseStr(SkipMany(Collect(digit, lower)), "0x1y2z*"),
			PState{input: "*", pos: 6, value: nil, ok: true})
		tt.AssertEqual(parseStr(SeqRight(SkipMany(Optional(digit)), Token(u8.Desur("*"))), "0123456789*"),
			PState{input: "", pos: 11, value: u8.Text("*"), ok: true})

		tt.AssertEqual(parseStr(SkipMany1(letter), "*"),
			PState{input: "*", pos: 0, value: nil, empty: true, error: makeUnexpInp("*")})
		tt.AssertEqual(parseStr(SkipMany1(Collect(letter, digit)), "A*"),
			PState{input: "*", pos: 1, value: nil, error: makeUnexpInp("*")})
		tt.AssertEqual(parseStr(SkipMany1(Collect(digit, lower)), "0x*"),
			PState{input: "*", pos: 2, value: nil, ok: true})
		tt.AssertEqual(parseStr(SkipMany1(letter), "abcdefghijk*"),
			PState{input: "*", pos: 11, value: nil, ok: true})
		tt.AssertEqual(parseStr(SkipMany1(Collect(digit, lower)), "0x1y2z*"),
			PState{input: "*", pos: 6, value: nil, ok: true})
		tt.AssertEqual(parseStr(SeqRight(SkipMany1(Optional(digit)), Token(u8.Desur("*"))), "0123456789*"),
			PState{input: "", pos: 11, value: u8.Text("*"), ok: true})
	})
}

//================================================================================
//LookAhead, Predict
func TestLookAheads(t *testing.T) {
	ts.LogAsserts("LookAheads", t, func(tt *ts.T) {
		letter := Regexp(`\pL`)
		digit := Regexp(`\p{Nd}`)

		tt.AssertEqual(parseStr(LookAhead(Many(digit)), "123"),
			PState{input: "123", pos: 0, value: arrText("1", "2", "3"), ok: true, empty: true})
		tt.AssertEqual(parseStr(LookAhead(Many(digit)), "YYZ"),
			PState{input: "YYZ", pos: 0, value: []interface{}{}, ok: true, empty: true})
		tt.AssertEqual(parseStr(LookAhead(digit), "*"),
			PState{input: "*", pos: 0, value: nil, ok: true, empty: true, error: ""})
		tt.AssertEqual(parseStr(LookAhead(SeqRight(letter, digit)), "A*"),
			PState{input: "A*", pos: 0, value: nil, ok: true, empty: true, error: ""})

		tt.AssertEqual(parseStr(Predict(Token(u8.Text("="))), "=10"),
			PState{input: "=10", pos: 0, value: u8.Text("="), ok: true})
		tt.AssertEqual(parseStr(Predict(Token(u8.Text("="))), "<10"),
			PState{input: "<10", pos: 0, empty: true, error: makeUnexpInp("<")})
	})
}

//================================================================================
//Search
func TestSearch(t *testing.T) {
	ts.LogAsserts("Search", t, func(tt *ts.T) {
		digit := Regexp(`\p{Nd}`)
		intNum := Regexp(`\p{Nd}+`)

		tt.AssertEqual(parseStr(Search(digit), "Now I have 2 dollars"),
			PState{input: " dollars", pos: 12, value: u8.Text("2"), ok: true})
		tt.AssertEqual(parseStr(Search(intNum), "Now I have 20 dollars"),
			PState{input: " dollars", pos: 13, value: u8.Text("20"), ok: true})
		tt.AssertEqual(parseStr(Many(Search(intNum)), "Now I have 20 dollars, or 2 tens."),
			PState{input: "", pos: 33, value: arrText("20", "2"), ok: true})
		tt.AssertEqual(parseStr(Many(Search(Alt(intNum, Token(u8.Text("dollars"))))), "Now I have 20 dollars"),
			PState{input: "", pos: 21, value: arrText("20", "dollars"), ok: true})
	})
}

//================================================================================
func TestBindIdentity(t *testing.T) {
	ts.LogAsserts("BindIdentity", t, func(tt *ts.T) {
		letter := Regexp(`\pL`)
		digit := Regexp(`\p{Nd}`)
		lower := Regexp(`\p{Ll}`)
		upper := Regexp(`\p{Lu}`)
		singleLetter := Satisfy(func(c u8.Codepoint) bool {
			cc := u8.SurrogatePoint(c)
			return len(cc) == 1 && unicode.IsLetter(cc[0])
		})

		var bindIdentity = func(q Parser) Parser {
			return Bind(q, func(s interface{}) Parser {
				return Return(s)
			})
		}

		//Basics
		tt.AssertEqual(parseStr(bindIdentity(Return(1234567890)), "abc"),
			PState{input: "abc", value: 1234567890, ok: true, empty: true})
		tt.AssertEqual(parseStr(bindIdentity(Fail(u8.Desur("I fail!"))), "abc"),
			PState{input: "abc", empty: true, error: "I fail!"})

		//Primitives
		tt.AssertEqual(parseStr(bindIdentity(Regexp(`a.c`)), "abcdefg"),
			PState{input: "defg", pos: 3, value: u8.Text("abc"), ok: true})
		tt.AssertEqual(parseStr(bindIdentity(Regexp(`a[bc]*?d`)), "abcdefg"),
			PState{input: "efg", pos: 4, value: u8.Text("abcd"), ok: true})

		//Lexicals
		tt.AssertEqual(parseStr(bindIdentity(letter), "defg"),
			PState{input: "efg", pos: 1, value: u8.Text("d"), ok: true})
		tt.AssertEqual(parseStr(bindIdentity(letter), ";efg"),
			PState{input: ";efg", empty: true, error: makeUnexpInp(";")})

		//Alts
		tt.AssertEqual(parseStr(bindIdentity(Alt(digit, upper, lower)), "6efg"),
			PState{input: "efg", pos: 1, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(bindIdentity(Alt(digit, upper, lower)), "Defg"),
			PState{input: "efg", pos: 1, value: u8.Text("D"), ok: true})
		tt.AssertEqual(parseStr(bindIdentity(Alt(digit, upper, lower)), ";efg"),
			PState{input: ";efg", empty: true, error: errNoAlts})

		//Seqs
		p3 := Bind(digit, func(c1 interface{}) Parser {
			return Bind(lower, func(c2 interface{}) Parser {
				return Bind(upper, func(c3 interface{}) Parser {
					return Return([]interface{}{c1, c2, c3})
				})
			})
		})
		tt.AssertEqual(parseStr(bindIdentity(p3), "6eFghij"),
			PState{input: "ghij", pos: 3, value: arrText("6", "e", "F"), ok: true})
		tt.AssertEqual(parseStr(bindIdentity(SeqRight(digit, lower, lower)), "6efg"),
			PState{input: "g", pos: 3, value: u8.Text("f"), ok: true})
		tt.AssertEqual(parseStr(bindIdentity(SeqLeft(digit, lower, lower)), "6efg"),
			PState{input: "g", pos: 3, value: u8.Text("6"), ok: true})

		//Try
		p2 := Bind(digit, func(c1 interface{}) Parser {
			return Bind(lower, func(c2 interface{}) Parser {
				return Return([]interface{}{c1, c2})
			})
		})
		tt.AssertEqual(parseStr(bindIdentity(Try(p2)), "67fg"),
			PState{input: "67fg", empty: true, error: makeUnexpInp("7")})
		tt.AssertEqual(parseStr(bindIdentity(Try(SeqRight(digit, lower))), "67efg"),
			PState{input: "67efg", empty: true, error: makeUnexpInp("7")}) //TODO: pos and error don't sync up

		//NotFollowedBy
		tt.AssertEqual(parseStr(bindIdentity(SeqLeft(digit, NotFollowedBy(letter))), "66efg"),
			PState{input: "6efg", pos: 1, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(bindIdentity(SeqRight(digit, NotFollowedBy(letter))), "66efg"),
			PState{input: "6efg", pos: 1, ok: true})
		tt.AssertEqual(parseStr(bindIdentity(SeqLeft(digit, NotFollowedBy(digit))), "66efg"),
			PState{input: "6efg", pos: 1, error: errParsFail})
		tt.AssertEqual(parseStr(bindIdentity(Try(SeqLeft(digit, NotFollowedBy(digit)))), "66efg"),
			PState{input: "66efg", empty: true, error: errParsFail})

		//Eof
		tt.AssertEqual(parseStr(bindIdentity(Eof), ""),
			PState{input: "", ok: true, empty: true})
		tt.AssertEqual(parseStr(bindIdentity(Eof), "abc"),
			PState{input: "abc", empty: true, error: errParsFail})

		//Apply
		p1 := Apply(func(c interface{}) interface{} { cs := u8.Surr(c.(u8.Text)); return cs + cs + cs }, letter)
		tt.AssertEqual(parseStr(bindIdentity(p1), "abc"),
			PState{input: "bc", pos: 1, value: "aaa", ok: true})

		//Between
		tt.AssertEqual(parseStr(bindIdentity(Between(letter, letter, digit)), "a6zz7"),
			PState{input: "z7", pos: 3, value: u8.Text("6"), ok: true})
		tt.AssertEqual(parseStr(bindIdentity(Between(letter, digit, Return(nil))), "a6zz7"),
			PState{input: "zz7", pos: 2, ok: true})

		//Collect
		tt.AssertEqual(parseStr(bindIdentity(Collect(letter, Token(u8.Text("bb")), letter)), "abbacde"),
			PState{input: "cde", pos: 4, value: arrText("a", "bb", "a"), ok: true})

		//Many
		tt.AssertEqual(parseStr(bindIdentity(Many(singleLetter)), "def;ghij"),
			PState{input: ";ghij", pos: 3, value: arrCp('d', 'e', 'f'), ok: true})
		tt.AssertEqual(parseStr(bindIdentity(Many1(singleLetter)), "def;ghij"),
			PState{input: ";ghij", pos: 3, value: arrCp('d', 'e', 'f'), ok: true})
		tt.AssertEqual(parseStr(bindIdentity(Optional(letter)), "def;ghij"),
			PState{input: "ef;ghij", pos: 1, value: u8.Text("d"), ok: true})
		tt.AssertEqual(parseStr(bindIdentity(Option('z', letter)), "def;ghij"),
			PState{input: "ef;ghij", pos: 1, value: u8.Text("d"), ok: true})

		//Separators
		letterA := Satisfy(func(c u8.Codepoint) bool { return c == 'a' })
		tt.AssertEqual(parseStr(bindIdentity(SepBy(digit, letterA)), "a6a6abc"),
			PState{input: "bc", pos: 5, value: arrCp('a', 'a', 'a'), ok: true})
		tt.AssertEqual(parseStr(bindIdentity(SepBy1(digit, letterA)), "a6a6abc"),
			PState{input: "bc", pos: 5, value: arrCp('a', 'a', 'a'), ok: true})
		tt.AssertEqual(parseStr(bindIdentity(EndBy(digit, letterA)), "a6a6abc"),
			PState{input: "abc", pos: 4, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(bindIdentity(EndBy1(digit, letterA)), "a6a6abc"),
			PState{input: "abc", pos: 4, value: arrCp('a', 'a'), ok: true})
		tt.AssertEqual(parseStr(bindIdentity(SepEndBy(digit, letterA)), "a6a6abc"),
			PState{input: "bc", pos: 5, value: arrCp('a', 'a', 'a'), ok: true})
		tt.AssertEqual(parseStr(bindIdentity(SepEndBy1(digit, letterA)), "a6a6abc"),
			PState{input: "bc", pos: 5, value: arrCp('a', 'a', 'a'), ok: true})

		//State
		increment := func(n interface{}, _ ...interface{}) interface{} { return n.(int) + 1 }
		p4 := SeqRight(
			SeqRight(PutState(21),
				Bind(AnyChar, func(x interface{}) (p Parser) {
					if x.(u8.Codepoint) == u8.Codepoint('\n') {
						return ModifyState(increment)
					} else {
						return Return(nil)
					}
				})),
			GetState)
		tt.AssertEqual(parseStr(bindIdentity(p4), "\naa"),
			PState{input: "aa", pos: 1, value: 22, ok: true, user: 22})
		tt.AssertEqual(parseStr(bindIdentity(p4), "aa\naa\na"),
			PState{input: "a\naa\na", pos: 1, value: 21, ok: true, user: 21})
		tt.AssertEqual(parseStr(bindIdentity(SeqRight(p4, AnyChar)), "\naa"),
			PState{input: "a", pos: 2, value: u8.Codepoint('a'), user: 22, ok: true})
	})
}

//================================================================================
func TestFlagging(t *testing.T) {
	ts.LogAsserts("Flagging", t, func(tt *ts.T) {
		//tt.PrintValue("Activate Flagging Printing...")

		//NOTE: flagging not yet implemented
	})
}

//================================================================================
func TestUpperVolumes(t *testing.T) {
	ts.LogAsserts("UpperVolumes", t, func(tt *ts.T) {

		tt.AssertEqual(parseStr(Symbol(u8.Codepoint(0x110000)), "\U000f8011\U00100000abc"),
			PState{input: "abc", pos: 8, value: u8.Codepoint(0x110000), ok: true}) //pos 8 is length of surrogation

		tt.AssertEqual(parseStr(Token(u8.Text{0x110000}), "\U000f8011\U00100000abc"),
			PState{input: "abc", pos: 8, value: u8.Text{0x110000}, ok: true})

		point110000 := Satisfy(func(c u8.Codepoint) bool { return c == 0x110000 })
		tt.AssertEqual(parseStr(point110000, "\U000f8011\U00100000abc"),
			PState{input: "abc", pos: 8, value: u8.Codepoint(0x110000), ok: true})

		tt.AssertEqual(parseStr(AnyChar, "\U000f8011\U00100000abc"),
			PState{input: "abc", pos: 8, value: u8.Codepoint(0x110000), ok: true})
	})
}

//================================================================================
