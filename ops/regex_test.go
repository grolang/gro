// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

/*
contents:
func TestRunex(t *testing.T){
func TestFindAll(t *testing.T){
func TestRegexRepeat(t *testing.T){
func TestRegexGroups(t *testing.T){
func TestRegexParenthesize(t *testing.T){
func TestRegexNot(t *testing.T){
func TestBooleanExprs(t *testing.T){
func TestRegexAlts(t *testing.T){
func TestRegexSeqs(t *testing.T){
func TestRegexMisc(t *testing.T){
*/
package ops_test

import (
	"math"
	"regexp"
	"testing"

	ts "github.com/grolang/gro/assert"
	tp "github.com/grolang/gro/ops"
	kn "github.com/grolang/gro/parsec"
	u8 "github.com/grolang/gro/utf88"
)

func validateRegex(tt *ts.T, b interface{}) {
	if nb, isCharClass := b.(tp.CharClass); isCharClass {
		_, e := regexp.Compile(string(nb))
		tt.Assert(e == nil)
	} else if nb, isRegex := b.(tp.Regex); isRegex {
		_, e := regexp.Compile(string(nb))
		tt.Assert(e == nil)
	}
}

//================================================================================
func TestRunex(t *testing.T) {
	ts.LogAsserts("Runex", t, func(tt *ts.T) {
		//t.Errorf("%v", 0)
		for _, n := range []struct {
			a string
			b interface{}
		}{
			// A Go rune can be any arbitrary Unicode codepoint:
			// * except newline
			{`a`, u8.Codepoint('a')},
			{`]`, u8.Codepoint(']')},
			{`-`, u8.Codepoint('-')},
			{`^`, u8.Codepoint('^')},
			{`"`, u8.Codepoint('"')},

			// * a backslash-escaped character \a \b \e \f \n \r \t \v \' \\
			{`\a`, u8.Codepoint('\a')},
			{`\b`, u8.Codepoint('\b')},
			{`\e`, u8.Codepoint('\x1b')},
			{`\f`, u8.Codepoint('\f')},
			{`\n`, u8.Codepoint('\n')},
			{`\r`, u8.Codepoint('\r')},
			{`\t`, u8.Codepoint('\t')},
			{`\v`, u8.Codepoint('\v')},
			{`\'`, u8.Codepoint('\'')},
			{`\\`, u8.Codepoint('\\')},
			{`\]`, u8.Codepoint(']')},
			{`\-`, u8.Codepoint('-')},
			{`\^`, u8.Codepoint('^')},

			// * a hex up to \xff
			{`\x20`, u8.Codepoint(' ')},

			// * an octal up to \377
			{`\377`, u8.Codepoint('\377')},
			{`\012`, u8.Codepoint('\012')},
			{`\007`, u8.Codepoint('\007')},

			// * a 4-byte unicode point e.g. \u159d
			{`\u123f`, u8.Codepoint('\u123f')},
			{`\u012f`, u8.Codepoint('\u012f')},
			{`\u001f`, u8.Codepoint('\u001f')},
			{`\u000f`, u8.Codepoint('\u000f')},

			// * an 8-byte unicode point e.g. \U13579bdf
			{`\U0010567f`, u8.Codepoint('\U0010567f')},
			{`\U13579bdf`, u8.Codepoint(0x13579bdf)},

			{`\x{7}`, u8.Codepoint('\u0007')},
			{`\x{f}`, u8.Codepoint('\u000f')},
			{`\x{1f}`, u8.Codepoint('\u001f')},
			{`\x{12F}`, u8.Codepoint('\u012f')},
			{`\x{123f}`, u8.Codepoint('\u123f')},
			{`\x{10_123f}`, u8.Codepoint('\U0010123f')},
			{`\x{10_12_3f}`, u8.Codepoint('\U0010123f')},
			{`\x{10567f}`, u8.Codepoint('\U0010567f')},
			{`\x{1357_9bdf}`, u8.Codepoint(0x13579bdf)},

			//--------------------------------------------------------------------------
			// A Go rexexp char class can be:
			// * a single char or sequence of them
			{`ab`, tp.CharClass(`[ab]`)},
			{`abc`, tp.CharClass(`[abc]`)},

			// * can include ranges using hyphen e.g. a-zA-Z0-9
			{`2a-b`, tp.CharClass(`[2a-b]`)},

			// * first char can be ^ to negate
			{`^2a-c`, tp.CharClass(`[^2a-c]`)},

			// * puncs may be backslashed but some must be, i.e. first ^, each ], each \, each - except last
			{`2a-c-`, tp.CharClass(`[2a-c-]`)},
			{`2a-c\-`, tp.CharClass(`[2a-c\-]`)},
			{`2a\]c`, tp.CharClass(`[2a\]c]`)},
			{`2a\\c`, tp.CharClass(`[2a\\c]`)},
			{`2a\-c`, tp.CharClass(`[2a\-c]`)},
			{`\^2a-c`, tp.CharClass(`[\^2a-c]`)},

			// * \a \f \n \r \t \v ALSO: \b \e
			{`a\ad\fg\nj`, tp.CharClass(`[a\ad\fg\nj]`)},
			{`a\rd\tg\vj`, tp.CharClass(`[a\rd\tg\vj]`)},
			{`a-c\bd-f`, tp.CharClass(`[a-c\x{8}d-f]`)},
			{`a-c\ed-f`, tp.CharClass(`[a-c\x{1b}d-f]`)},

			// * an octal up to \377
			// * a hex up to \xff
			// * a hex of any length \x{13579bdf}
			{`\012\177`, tp.CharClass(`[\012\177]`)},
			{`\x12\x1f`, tp.CharClass(`[\x12\x1f]`)},
			{`\x{123}\x{125}`, tp.CharClass(`[\x{123}\x{125}]`)},
			{`\x{12_3f}\x{1_2ef}`, tp.CharClass(`[\x{123f}\x{12ef}]`)},
			{`\x{12_3f}-\x{1_2ef}`, tp.CharClass(`[\x{123f}-\x{12ef}]`)},

			{`\x{1234_567f}-\x{1234_67ef}`, tp.Regex(`(?:\x{f9234}[\x{10567f}-\x{1067ef}])`)},
			{`\x{1234_567f}-\x{1235_12ef}`, tp.Regex(`(?:\x{f9234}[\x{10567f}-\x{10ffff}])` +
				`|(?:\x{f9235}[\x{100000}-\x{1012ef}])`)},
			{`\x{1234_567f}-\x{1236_12ef}`, tp.Regex(`(?:\x{f9234}[\x{10567f}-\x{10ffff}])` +
				`|(?:\x{f9235}[\x{100000}-\x{10ffff}])` +
				`|(?:\x{f9236}[\x{100000}-\x{1012ef}])`)},
			{`\x{1234_567f}-\x{12ff_67ef}`, tp.Regex(`(?:\x{f9234}[\x{10567f}-\x{10ffff}])` +
				`|(?:[\x{f9235}-\x{f92fe}][\x{100000}-\x{10ffff}])` +
				`|(?:\x{f92ff}[\x{100000}-\x{1067ef}])`)},

			// * Perl char classes \d \D \s \S \w \W
			{`\d`, tp.CharClass(`[\d]`)},
			{`\D`, tp.CharClass(`[\D]`)},
			{`\s`, tp.CharClass(`[\s]`)},
			{`abc\S\w\Wdef`, tp.CharClass(`[abc\S\w\Wdef]`)},

			// * Unicode char classes \pL or \PL
			{`\pL`, tp.CharClass(`[\pL]`)},
			{`\PL`, tp.CharClass(`[\PL]`)},

			// * Unicode char classes \p{Greek} or \P{Greek}
			{`\p{Greek}`, tp.CharClass(`[\p{Greek}]`)},
			{`\P{Greek}`, tp.CharClass(`[\P{Greek}]`)},

			// * ASCII char classes [:word:] or [:^word:] or ^[:word:]

		} {
			tt.AssertEqual(tp.Runex(n.a), n.b)
			validateRegex(tt, n.b)
			//tt.AssertEqual(fmt.Sprintf("%T", tp.Runex(n.a)), fmt.Sprintf("%T", n.b))
		}
	})
}

//================================================================================
func TestFindAll(t *testing.T) {
	ts.LogAsserts("FindAll", t, func(tt *ts.T) {
		w1 := tp.Seq(tp.Group(tp.Runex("a-z"), "alpha"),
			tp.Group(tp.Seq(tp.Runex("0-9"),
				tp.Regex(tp.ToRegex(tp.Runex("0-9"))+"?")), "digits")) //used tp.ToRegex because kn.Optional removed

		for _, n := range []struct{ a, b interface{} }{
			{tp.FindAll(tp.Regex("a|b"), "a")[0][0].Match, "a"},
			{tp.FindAll(tp.Regex("a|b"), "acb")[1][0].Match, "b"},
			{tp.FindAll(tp.Regex("a|(c)|e"), "abcdefg")[1][1].Match, "c"},

			{tp.FindAll(tp.Regex("a|b"), "a"), [][]tp.RegexGroupMatch{{{"", 0, 1, "a"}}}},
			{tp.FindAll(tp.Regex("a|b"), "acb"), [][]tp.RegexGroupMatch{{{"", 0, 1, "a"}},
				{{"", 2, 3, "b"}}}},
			{tp.FindAll(tp.Regex("a|c|e"), "abcdefg"), [][]tp.RegexGroupMatch{{{"", 0, 1, "a"}},
				{{"", 2, 3, "c"}},
				{{"", 4, 5, "e"}}}},
			{tp.FindAll(tp.Regex("a|(c)|e"), "abcdefg"), [][]tp.RegexGroupMatch{{{"", 0, 1, "a"}, {"", -1, -1, ""}},
				{{"", 2, 3, "c"}, {"", 2, 3, "c"}},
				{{"", 4, 5, "e"}, {"", -1, -1, ""}}}},

			{w1, tp.Regex(`(?:(?P<alpha>[a-z]))(?:(?P<digits>[0-9](?:[0-9]?)))`)},
			{tp.FindAll(w1, "a7, b22, c86, d99, e3"), [][]tp.RegexGroupMatch{
				{{"", 0, 2, "a7"}, {"alpha", 0, 1, "a"}, {"digits", 1, 2, "7"}},
				{{"", 4, 7, "b22"}, {"alpha", 4, 5, "b"}, {"digits", 5, 7, "22"}},
				{{"", 9, 12, "c86"}, {"alpha", 9, 10, "c"}, {"digits", 10, 12, "86"}},
				{{"", 14, 17, "d99"}, {"alpha", 14, 15, "d"}, {"digits", 15, 17, "99"}},
				{{"", 19, 21, "e3"}, {"alpha", 19, 20, "e"}, {"digits", 20, 21, "3"}},
			}},
		} {
			tt.AssertEqual(n.a, n.b)
			validateRegex(tt, n.b)
		}
	})
}

//================================================================================
func TestRegexRepeat(t *testing.T) {
	ts.LogAsserts("RegexRepeat", t, func(tt *ts.T) {
		for _, n := range []struct{ a, b interface{} }{
			{tp.RegexRepeat(tp.Runex("a"), tp.PosIntRange{tp.Int(0), math.Inf(1)}, false), tp.Regex(`a*`)},
			{tp.RegexRepeat(tp.Runex("a-z"), tp.PosIntRange{tp.Int(0), math.Inf(1)}, false), tp.Regex(`[a-z]*`)},
			{tp.RegexRepeat(u8.Text("abc"), tp.PosIntRange{tp.Int(0), math.Inf(1)}, false), tp.Regex(`(?:\Qabc\E)*`)},
			{tp.RegexRepeat(tp.Regex("a|b"), tp.PosIntRange{tp.Int(0), math.Inf(1)}, false), tp.Regex(`(?:a|b)*`)},

			{tp.RegexRepeat(tp.Runex("a"), tp.PosIntRange{tp.Int(0), math.Inf(1)}, true), tp.Regex(`a*?`)},
			{tp.RegexRepeat(tp.Runex("a-z"), tp.PosIntRange{tp.Int(0), math.Inf(1)}, true), tp.Regex(`[a-z]*?`)},
			{tp.RegexRepeat(u8.Text("abc"), tp.PosIntRange{tp.Int(0), math.Inf(1)}, true), tp.Regex(`(?:\Qabc\E)*?`)},
			{tp.RegexRepeat(tp.Regex("a|b"), tp.PosIntRange{tp.Int(0), math.Inf(1)}, true), tp.Regex(`(?:a|b)*?`)},

			//--------------------------------------------------------------------------------
			{tp.RegexRepeat(tp.Runex("a"), tp.PosIntRange{tp.Int(1), math.Inf(1)}, false), tp.Regex(`a+`)},
			{tp.RegexRepeat(tp.Runex("a-z"), tp.PosIntRange{tp.Int(1), math.Inf(1)}, false), tp.Regex(`[a-z]+`)},
			{tp.RegexRepeat(u8.Text("abc"), tp.PosIntRange{tp.Int(1), math.Inf(1)}, false), tp.Regex(`(?:\Qabc\E)+`)},
			{tp.RegexRepeat(tp.Regex("a|b"), tp.PosIntRange{tp.Int(1), math.Inf(1)}, false), tp.Regex(`(?:a|b)+`)},

			{tp.RegexRepeat(tp.Runex("a"), tp.PosIntRange{tp.Int(1), math.Inf(1)}, true), tp.Regex(`a+?`)},
			{tp.RegexRepeat(tp.Runex("a-z"), tp.PosIntRange{tp.Int(1), math.Inf(1)}, true), tp.Regex(`[a-z]+?`)},
			{tp.RegexRepeat(u8.Text("abc"), tp.PosIntRange{tp.Int(1), math.Inf(1)}, true), tp.Regex(`(?:\Qabc\E)+?`)},
			{tp.RegexRepeat(tp.Regex("a|b"), tp.PosIntRange{tp.Int(1), math.Inf(1)}, true), tp.Regex(`(?:a|b)+?`)},

			//--------------------------------------------------------------------------------
			{tp.RegexRepeat(tp.Runex("a"), tp.PosIntRange{tp.Int(0), tp.Int(1)}, false), tp.Regex(`a?`)},
			{tp.RegexRepeat(tp.Runex("a-z"), tp.PosIntRange{tp.Int(0), tp.Int(1)}, false), tp.Regex(`[a-z]?`)},
			{tp.RegexRepeat(u8.Text("abc"), tp.PosIntRange{tp.Int(0), tp.Int(1)}, false), tp.Regex(`(?:\Qabc\E)?`)},
			{tp.RegexRepeat(tp.Regex("a|b"), tp.PosIntRange{tp.Int(0), tp.Int(1)}, false), tp.Regex(`(?:a|b)?`)},

			{tp.RegexRepeat(tp.Runex("a"), tp.PosIntRange{tp.Int(0), tp.Int(1)}, true), tp.Regex(`a??`)},
			{tp.RegexRepeat(tp.Runex("a-z"), tp.PosIntRange{tp.Int(0), tp.Int(1)}, true), tp.Regex(`[a-z]??`)},
			{tp.RegexRepeat(u8.Text("abc"), tp.PosIntRange{tp.Int(0), tp.Int(1)}, true), tp.Regex(`(?:\Qabc\E)??`)},
			{tp.RegexRepeat(tp.Regex("a|b"), tp.PosIntRange{tp.Int(0), tp.Int(1)}, true), tp.Regex(`(?:a|b)??`)},

			//--------------------------------------------------------------------------------
			{tp.RegexRepeat(tp.Runex("a"), tp.Int(2), false), tp.Regex(`a{2}`)},
			{tp.RegexRepeat(tp.Runex("a-z"), tp.Int(2), false), tp.Regex(`[a-z]{2}`)},
			{tp.RegexRepeat(u8.Text("abc"), tp.Int(2), false), tp.Regex(`(?:\Qabc\E){2}`)},
			{tp.RegexRepeat(tp.Regex("a|b"), tp.Int(2), false), tp.Regex(`(?:a|b){2}`)},

			{tp.RegexRepeat(tp.Runex("a"), tp.Int(2), true), tp.Regex(`a{2}?`)},
			{tp.RegexRepeat(tp.Runex("a-z"), tp.Int(2), true), tp.Regex(`[a-z]{2}?`)},
			{tp.RegexRepeat(u8.Text("abc"), tp.Int(2), true), tp.Regex(`(?:\Qabc\E){2}?`)},
			{tp.RegexRepeat(tp.Regex("a|b"), tp.Int(2), true), tp.Regex(`(?:a|b){2}?`)},

			{tp.RegexRepeat(tp.Runex("a"), tp.PosIntRange{tp.Int(2), tp.Int(2)}, false), tp.Regex(`a{2}`)},
			{tp.RegexRepeat(tp.Runex("a-z"), tp.PosIntRange{tp.Int(2), tp.Int(2)}, false), tp.Regex(`[a-z]{2}`)},
			{tp.RegexRepeat(u8.Text("abc"), tp.PosIntRange{tp.Int(2), tp.Int(2)}, false), tp.Regex(`(?:\Qabc\E){2}`)},
			{tp.RegexRepeat(tp.Regex("a|b"), tp.PosIntRange{tp.Int(2), tp.Int(2)}, false), tp.Regex(`(?:a|b){2}`)},

			{tp.RegexRepeat(tp.Runex("a"), tp.PosIntRange{tp.Int(2), tp.Int(2)}, true), tp.Regex(`a{2}?`)},
			{tp.RegexRepeat(tp.Runex("a-z"), tp.PosIntRange{tp.Int(2), tp.Int(2)}, true), tp.Regex(`[a-z]{2}?`)},
			{tp.RegexRepeat(u8.Text("abc"), tp.PosIntRange{tp.Int(2), tp.Int(2)}, true), tp.Regex(`(?:\Qabc\E){2}?`)},
			{tp.RegexRepeat(tp.Regex("a|b"), tp.PosIntRange{tp.Int(2), tp.Int(2)}, true), tp.Regex(`(?:a|b){2}?`)},

			//--------------------------------------------------------------------------------
			{tp.RegexRepeat(tp.Runex("a"), tp.PosIntRange{tp.Int(2), math.Inf(1)}, false), tp.Regex(`a{2,}`)},
			{tp.RegexRepeat(tp.Runex("a-z"), tp.PosIntRange{tp.Int(2), math.Inf(1)}, false), tp.Regex(`[a-z]{2,}`)},
			{tp.RegexRepeat(u8.Text("abc"), tp.PosIntRange{tp.Int(2), math.Inf(1)}, false), tp.Regex(`(?:\Qabc\E){2,}`)},
			{tp.RegexRepeat(tp.Regex("a|b"), tp.PosIntRange{tp.Int(2), math.Inf(1)}, false), tp.Regex(`(?:a|b){2,}`)},

			{tp.RegexRepeat(tp.Runex("a"), tp.PosIntRange{tp.Int(2), math.Inf(1)}, true), tp.Regex(`a{2,}?`)},
			{tp.RegexRepeat(tp.Runex("a-z"), tp.PosIntRange{tp.Int(2), math.Inf(1)}, true), tp.Regex(`[a-z]{2,}?`)},
			{tp.RegexRepeat(u8.Text("abc"), tp.PosIntRange{tp.Int(2), math.Inf(1)}, true), tp.Regex(`(?:\Qabc\E){2,}?`)},
			{tp.RegexRepeat(tp.Regex("a|b"), tp.PosIntRange{tp.Int(2), math.Inf(1)}, true), tp.Regex(`(?:a|b){2,}?`)},

			//--------------------------------------------------------------------------------
			{tp.RegexRepeat(tp.Runex("a"), tp.PosIntRange{tp.Int(2), tp.Int(3)}, false), tp.Regex(`a{2,3}`)},
			{tp.RegexRepeat(tp.Runex("a-z"), tp.PosIntRange{tp.Int(2), tp.Int(3)}, false), tp.Regex(`[a-z]{2,3}`)},
			{tp.RegexRepeat(u8.Text("abc"), tp.PosIntRange{tp.Int(2), tp.Int(3)}, false), tp.Regex(`(?:\Qabc\E){2,3}`)},
			{tp.RegexRepeat(tp.Regex("a|b"), tp.PosIntRange{tp.Int(2), tp.Int(3)}, false), tp.Regex(`(?:a|b){2,3}`)},

			{tp.RegexRepeat(tp.Runex("a"), tp.PosIntRange{tp.Int(2), tp.Int(3)}, true), tp.Regex(`a{2,3}?`)},
			{tp.RegexRepeat(tp.Runex("a-z"), tp.PosIntRange{tp.Int(2), tp.Int(3)}, true), tp.Regex(`[a-z]{2,3}?`)},
			{tp.RegexRepeat(u8.Text("abc"), tp.PosIntRange{tp.Int(2), tp.Int(3)}, true), tp.Regex(`(?:\Qabc\E){2,3}?`)},
			{tp.RegexRepeat(tp.Regex("a|b"), tp.PosIntRange{tp.Int(2), tp.Int(3)}, true), tp.Regex(`(?:a|b){2,3}?`)},
		} {
			tt.AssertEqual(n.a, n.b)
			validateRegex(tt, n.b)
		}
	})
}

//================================================================================
func TestRegexGroups(t *testing.T) {
	ts.LogAsserts("RegexGroups", t, func(tt *ts.T) {
		for _, n := range []struct{ a, b interface{} }{
			{tp.Group(tp.Runex("a"), ""), tp.Regex(`(a)`)},
			{tp.Group(tp.Runex("a-z"), ""), tp.Regex(`([a-z])`)},
			{tp.Group(u8.Text("abc"), ""), tp.Regex(`(\Qabc\E)`)},
			{tp.Group(tp.Regex("a|b"), ""), tp.Regex(`(a|b)`)},

			{tp.Group(tp.Runex("a"), "aName"), tp.Regex(`(?P<aName>a)`)},
			{tp.Group(tp.Runex("a-z"), "aName"), tp.Regex(`(?P<aName>[a-z])`)},
			{tp.Group(u8.Text("abc"), "aName"), tp.Regex(`(?P<aName>\Qabc\E)`)},
			{tp.Group(tp.Regex("a|b"), "aName"), tp.Regex(`(?P<aName>a|b)`)},
		} {
			tt.AssertEqual(n.a, n.b)
			validateRegex(tt, n.b)
		}
	})
}

//================================================================================
func TestRegexParenthesize(t *testing.T) {
	ts.LogAsserts("RegexParenthesize", t, func(tt *ts.T) {
		for _, n := range []struct{ a, b interface{} }{
			{tp.Parenthesize(tp.Runex("a")), tp.Runex("a")},
			{tp.Parenthesize(u8.Text("abc")), u8.Text("abc")},
			{tp.Parenthesize(tp.Runex("a-z")), tp.Regex(`(?:[a-z])`)},
			{tp.Parenthesize(tp.Regex("a|b")), tp.Regex(`(?:a|b)`)},
		} {
			tt.AssertEqual(n.a, n.b)
			validateRegex(tt, n.b)
		}
	})
}

//================================================================================
func TestRegexNot(t *testing.T) {
	ts.LogAsserts("RegexNot", t, func(tt *ts.T) {
		for _, n := range []struct{ a, b interface{} }{
			{tp.Not(tp.Runex("a")), tp.CharClass(`[^a]`)},
			{tp.Not(tp.Runex(`2a-c`)), tp.CharClass(`[^2a-c]`)},
			{tp.Not(tp.Runex(`^2a-c`)), tp.CharClass(`[2a-c]`)},
		} {
			tt.AssertEqual(n.a, n.b)
			validateRegex(tt, n.b)
		}
	})
}

//================================================================================
func TestBooleanExprs(t *testing.T) {
	ts.LogAsserts("BooleanExprs", t, func(tt *ts.T) {

		tt.Assert(!tp.Not(true).(bool))
		tt.Assert(tp.Not(false).(bool))

		tt.Assert(tp.And(true, func() interface{} { return true }))
		tt.Assert(!tp.And(false, func() interface{} { return true }))
		tt.Assert(!tp.And(true, func() interface{} { return false }))
		tt.Assert(!tp.And(false, func() interface{} { return false }))

		tt.Assert(tp.Or(true, func() interface{} { return true }))
		tt.Assert(tp.Or(false, func() interface{} { return true }))
		tt.Assert(tp.Or(true, func() interface{} { return false }))
		tt.Assert(!tp.Or(false, func() interface{} { return false }))

	})
}

//================================================================================
func TestRegexAlts(t *testing.T) {
	ts.LogAsserts("RegexAlts", t, func(tt *ts.T) {
		for _, n := range []struct{ a, b interface{} }{
			//Codepoint...
			{tp.Alt(tp.Runex("a"), tp.Runex("b")), tp.CharClass("[ab]")},     //point + point
			{tp.Alt(tp.Runex("a"), tp.Runex("g-j")), tp.CharClass("[ag-j]")}, //point + class
			{tp.Alt(tp.Runex("a"), tp.Runex("^g-j")), tp.Regex("a|[^g-j]")},  //point + negclass
			{tp.Alt(tp.Runex("a"), u8.Text("fg")), tp.Regex(`a|(?:\Qfg\E)`)}, //point + text
			{tp.Alt(tp.Runex("a"), tp.Regex("fg+")), tp.Regex("a|(?:fg+)")},  //point + regexp

			//u8.Text...
			{tp.Alt(u8.Text("ab"), tp.Runex("g")), tp.Regex(`(?:\Qab\E)|g`)},              //text + point
			{tp.Alt(u8.Text("abc"), tp.Runex("g-j")), tp.Regex(`(?:\Qabc\E)|[g-j]`)},      //text + class
			{tp.Alt(u8.Text("abc"), u8.Text("fgh")), tp.Regex(`(?:\Qabc\E)|(?:\Qfgh\E)`)}, //text + text
			{tp.Alt(u8.Text("abc"), tp.Regex("f|g")), tp.Regex(`(?:\Qabc\E)|(?:f|g)`)},    //text + regexp

			//tp.CharClass...
			{tp.Alt(tp.Runex("a-e"), tp.Runex("g")), tp.Runex("a-eg")},               //class + point
			{tp.Alt(tp.Runex("^a-e"), tp.Runex("g")), tp.Regex("[^a-e]|g")},          //negclass + point
			{tp.Alt(tp.Runex("a-e"), tp.Runex("g-j")), tp.CharClass("[a-eg-j]")},     //class + class
			{tp.Alt(tp.Runex("^a-e"), tp.Runex("^g-j")), tp.CharClass(`[^a-eg-j]`)},  //negclass + negclass
			{tp.Alt(tp.Runex("^a-e"), tp.Runex("g-j")), tp.Regex(`[^a-e]|[g-j]`)},    //negclass + class
			{tp.Alt(tp.Runex("a-e"), tp.Runex("^g-j")), tp.Regex(`[a-e]|[^g-j]`)},    //class + negclass
			{tp.Alt(tp.Runex("a-e"), u8.Text("fgh")), tp.Regex(`[a-e]|(?:\Qfgh\E)`)}, //class + text
			{tp.Alt(tp.Runex("a-e"), tp.Regex("fg*")), tp.Regex("[a-e]|(?:fg*)")},    //class + regexp

			//tp.Regex...
			{tp.Alt(tp.Regex("f|g"), tp.Runex("a")), tp.Regex("(?:f|g)|a")},            //regexp + point
			{tp.Alt(tp.Regex("f|g"), tp.Runex("a-e")), tp.Regex("(?:f|g)|[a-e]")},      //regexp + class
			{tp.Alt(tp.Regex("f|g"), u8.Text("abc")), tp.Regex(`(?:f|g)|(?:\Qabc\E)`)}, //regexp + text
			{tp.Alt(tp.Regex("f|g"), tp.Regex("J|k")), tp.Regex("(?:f|g)|(?:J|k)")},    //regexp + regexp

		} {
			tt.AssertEqual(n.a, n.b)
			validateRegex(tt, n.b)
		}
	})
}

//================================================================================
func TestRegexSeqs(t *testing.T) {
	ts.LogAsserts("RegexSeqs", t, func(tt *ts.T) {
		for _, n := range []struct{ a, b interface{} }{
			{tp.Seq(tp.Runex("a"), tp.Runex("b")), u8.Text("ab")},          //point + point
			{tp.Seq(tp.Runex("a"), tp.Runex("g-j")), tp.Regex("a[g-j]")},   //point + class
			{tp.Seq(tp.Runex("a"), u8.Text("fg")), u8.Text("afg")},         //point + string
			{tp.Seq(tp.Runex("a"), tp.Regex("f|g")), tp.Regex("a(?:f|g)")}, //point + regexp

			{tp.Seq(tp.Runex("a-e"), tp.Runex("g")), tp.Regex("[a-e]g")},         //class + point
			{tp.Seq(tp.Runex("a-e"), tp.Runex("g-j")), tp.Regex("[a-e][g-j]")},   //class + class
			{tp.Seq(tp.Runex("a-e"), u8.Text("fgh")), tp.Regex("[a-e]fgh")},      //class + string
			{tp.Seq(tp.Runex("a-e"), tp.Regex("f|g")), tp.Regex("[a-e](?:f|g)")}, //class + regexp

			{tp.Seq(u8.Text("ab"), tp.Runex("g")), u8.Text("abg")},            //string + point
			{tp.Seq(u8.Text("abc"), tp.Runex("g-j")), tp.Regex("abc[g-j]")},   //string + class
			{tp.Seq(u8.Text("abc"), u8.Text("fgh")), u8.Text("abcfgh")},       //string + string
			{tp.Seq(u8.Text("abc"), tp.Regex("f|g")), tp.Regex("abc(?:f|g)")}, //string + regexp

			{tp.Seq(tp.Regex("f|g"), tp.Runex("a")), tp.Regex("(?:f|g)a")},         //regexp + point
			{tp.Seq(tp.Regex("f|g"), tp.Runex("a-e")), tp.Regex("(?:f|g)[a-e]")},   //regexp + class
			{tp.Seq(tp.Regex("f|g"), u8.Text("abc")), tp.Regex("(?:f|g)abc")},      //regexp + string
			{tp.Seq(tp.Regex("f|g"), tp.Regex("J|k")), tp.Regex("(?:f|g)(?:J|k)")}, //regexp + regexp
		} {
			tt.AssertEqual(n.a, n.b)
			validateRegex(tt, n.b)
		}
	})
}

//================================================================================
func TestRegexMisc(t *testing.T) {
	ts.LogAsserts("RegexMisc", t, func(tt *ts.T) {
		pa := tp.ToParser(tp.Runex("a"))
		_, _ = kn.ParseText(pa, u8.Text("abc"))

	})
}

//================================================================================
