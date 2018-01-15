// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package ops_test

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/grolang/gro/ops"
	"github.com/grolang/gro/parsec"
	u8 "github.com/grolang/gro/utf88"
)

func validateRegex(t *testing.T, b interface{}) {
	if nb, isCharClass := b.(ops.CharClass); isCharClass {
		_, e := regexp.Compile(string(nb))
		assert(t, e == nil)
	} else if nb, isRegex := b.(ops.Regex); isRegex {
		_, e := regexp.Compile(string(nb))
		assert(t, e == nil)
	}
}

func assert(t *testing.T, b bool) {
	if !b {
		t.Errorf("assert failed....found:%v (%[1]T)\n", b)
	}
}

func assertEqual(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("assert failed......found:%v (%[1]T)\n"+
			"             ...expected:%v (%[2]T)\n", a, b)
	}
}

//================================================================================
func TestRunex(t *testing.T) {
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
		{`ab`, ops.CharClass(`[ab]`)},
		{`abc`, ops.CharClass(`[abc]`)},

		// * can include ranges using hyphen e.g. a-zA-Z0-9
		{`2a-b`, ops.CharClass(`[2a-b]`)},

		// * first char can be ^ to negate
		{`^2a-c`, ops.CharClass(`[^2a-c]`)},

		// * puncs may be backslashed but some must be, i.e. first ^, each ], each \, each - except last
		{`2a-c-`, ops.CharClass(`[2a-c-]`)},
		{`2a-c\-`, ops.CharClass(`[2a-c\-]`)},
		{`2a\]c`, ops.CharClass(`[2a\]c]`)},
		{`2a\\c`, ops.CharClass(`[2a\\c]`)},
		{`2a\-c`, ops.CharClass(`[2a\-c]`)},
		{`\^2a-c`, ops.CharClass(`[\^2a-c]`)},

		// * \a \f \n \r \t \v ALSO: \b \e
		{`a\ad\fg\nj`, ops.CharClass(`[a\ad\fg\nj]`)},
		{`a\rd\tg\vj`, ops.CharClass(`[a\rd\tg\vj]`)},
		{`a-c\bd-f`, ops.CharClass(`[a-c\x{8}d-f]`)},
		{`a-c\ed-f`, ops.CharClass(`[a-c\x{1b}d-f]`)},

		// * an octal up to \377
		// * a hex up to \xff
		// * a hex of any length \x{13579bdf}
		{`\012\177`, ops.CharClass(`[\012\177]`)},
		{`\x12\x1f`, ops.CharClass(`[\x12\x1f]`)},
		{`\x{123}\x{125}`, ops.CharClass(`[\x{123}\x{125}]`)},
		{`\x{12_3f}\x{1_2ef}`, ops.CharClass(`[\x{123f}\x{12ef}]`)},
		{`\x{12_3f}-\x{1_2ef}`, ops.CharClass(`[\x{123f}-\x{12ef}]`)},

		{`\x{1234_567f}-\x{1234_67ef}`, ops.Regex(`(?:\x{f9234}[\x{10567f}-\x{1067ef}])`)},
		{`\x{1234_567f}-\x{1235_12ef}`, ops.Regex(`(?:\x{f9234}[\x{10567f}-\x{10ffff}])` +
			`|(?:\x{f9235}[\x{100000}-\x{1012ef}])`)},
		{`\x{1234_567f}-\x{1236_12ef}`, ops.Regex(`(?:\x{f9234}[\x{10567f}-\x{10ffff}])` +
			`|(?:\x{f9235}[\x{100000}-\x{10ffff}])` +
			`|(?:\x{f9236}[\x{100000}-\x{1012ef}])`)},
		{`\x{1234_567f}-\x{12ff_67ef}`, ops.Regex(`(?:\x{f9234}[\x{10567f}-\x{10ffff}])` +
			`|(?:[\x{f9235}-\x{f92fe}][\x{100000}-\x{10ffff}])` +
			`|(?:\x{f92ff}[\x{100000}-\x{1067ef}])`)},

		// * Perl char classes \d \D \s \S \w \W
		{`\d`, ops.CharClass(`[\d]`)},
		{`\D`, ops.CharClass(`[\D]`)},
		{`\s`, ops.CharClass(`[\s]`)},
		{`abc\S\w\Wdef`, ops.CharClass(`[abc\S\w\Wdef]`)},

		// * Unicode char classes \pL or \PL
		{`\pL`, ops.CharClass(`[\pL]`)},
		{`\PL`, ops.CharClass(`[\PL]`)},

		// * Unicode char classes \p{Greek} or \P{Greek}
		{`\p{Greek}`, ops.CharClass(`[\p{Greek}]`)},
		{`\P{Greek}`, ops.CharClass(`[\P{Greek}]`)},

		// * ASCII char classes [:word:] or [:^word:] or ^[:word:]

	} {
		assertEqual(t, ops.Runex(n.a), n.b)
		validateRegex(t, n.b)
		//assertEqual(t, fmt.Sprintf("%T", tp.Runex(n.a)), fmt.Sprintf("%T", n.b))
	}
}

//================================================================================
func TestRegexRepeat(t *testing.T) {
	for _, n := range []struct{ a, b interface{} }{
		{ops.RegexRepeat(ops.Runex("a"), ops.NewPosIntRange(ops.Int(0), ops.Inf), false), ops.Regex(`a*`)},
		{ops.RegexRepeat(ops.Runex("a-z"), ops.NewPosIntRange(ops.Int(0), ops.Inf), false), ops.Regex(`[a-z]*`)},
		{ops.RegexRepeat(u8.Text("abc"), ops.NewPosIntRange(ops.Int(0), ops.Inf), false), ops.Regex(`(?:\Qabc\E)*`)},
		{ops.RegexRepeat(ops.Regex("a|b"), ops.NewPosIntRange(ops.Int(0), ops.Inf), false), ops.Regex(`(?:a|b)*`)},

		{ops.RegexRepeat(ops.Runex("a"), ops.NewPosIntRange(ops.Int(0), ops.Inf), true), ops.Regex(`a*?`)},
		{ops.RegexRepeat(ops.Runex("a-z"), ops.NewPosIntRange(ops.Int(0), ops.Inf), true), ops.Regex(`[a-z]*?`)},
		{ops.RegexRepeat(u8.Text("abc"), ops.NewPosIntRange(ops.Int(0), ops.Inf), true), ops.Regex(`(?:\Qabc\E)*?`)},
		{ops.RegexRepeat(ops.Regex("a|b"), ops.NewPosIntRange(ops.Int(0), ops.Inf), true), ops.Regex(`(?:a|b)*?`)},

		//--------------------------------------------------------------------------------
		{ops.RegexRepeat(ops.Runex("a"), ops.NewPosIntRange(ops.Int(1), ops.Inf), false), ops.Regex(`a+`)},
		{ops.RegexRepeat(ops.Runex("a-z"), ops.NewPosIntRange(ops.Int(1), ops.Inf), false), ops.Regex(`[a-z]+`)},
		{ops.RegexRepeat(u8.Text("abc"), ops.NewPosIntRange(ops.Int(1), ops.Inf), false), ops.Regex(`(?:\Qabc\E)+`)},
		{ops.RegexRepeat(ops.Regex("a|b"), ops.NewPosIntRange(ops.Int(1), ops.Inf), false), ops.Regex(`(?:a|b)+`)},

		{ops.RegexRepeat(ops.Runex("a"), ops.NewPosIntRange(ops.Int(1), ops.Inf), true), ops.Regex(`a+?`)},
		{ops.RegexRepeat(ops.Runex("a-z"), ops.NewPosIntRange(ops.Int(1), ops.Inf), true), ops.Regex(`[a-z]+?`)},
		{ops.RegexRepeat(u8.Text("abc"), ops.NewPosIntRange(ops.Int(1), ops.Inf), true), ops.Regex(`(?:\Qabc\E)+?`)},
		{ops.RegexRepeat(ops.Regex("a|b"), ops.NewPosIntRange(ops.Int(1), ops.Inf), true), ops.Regex(`(?:a|b)+?`)},

		//--------------------------------------------------------------------------------
		{ops.RegexRepeat(ops.Runex("a"), ops.NewPosIntRange(ops.Int(0), ops.Int(1)), false), ops.Regex(`a?`)},
		{ops.RegexRepeat(ops.Runex("a-z"), ops.NewPosIntRange(ops.Int(0), ops.Int(1)), false), ops.Regex(`[a-z]?`)},
		{ops.RegexRepeat(u8.Text("abc"), ops.NewPosIntRange(ops.Int(0), ops.Int(1)), false), ops.Regex(`(?:\Qabc\E)?`)},
		{ops.RegexRepeat(ops.Regex("a|b"), ops.NewPosIntRange(ops.Int(0), ops.Int(1)), false), ops.Regex(`(?:a|b)?`)},

		{ops.RegexRepeat(ops.Runex("a"), ops.NewPosIntRange(ops.Int(0), ops.Int(1)), true), ops.Regex(`a??`)},
		{ops.RegexRepeat(ops.Runex("a-z"), ops.NewPosIntRange(ops.Int(0), ops.Int(1)), true), ops.Regex(`[a-z]??`)},
		{ops.RegexRepeat(u8.Text("abc"), ops.NewPosIntRange(ops.Int(0), ops.Int(1)), true), ops.Regex(`(?:\Qabc\E)??`)},
		{ops.RegexRepeat(ops.Regex("a|b"), ops.NewPosIntRange(ops.Int(0), ops.Int(1)), true), ops.Regex(`(?:a|b)??`)},

		//--------------------------------------------------------------------------------
		{ops.RegexRepeat(ops.Runex("a"), ops.Int(2), false), ops.Regex(`a{2}`)},
		{ops.RegexRepeat(ops.Runex("a-z"), ops.Int(2), false), ops.Regex(`[a-z]{2}`)},
		{ops.RegexRepeat(u8.Text("abc"), ops.Int(2), false), ops.Regex(`(?:\Qabc\E){2}`)},
		{ops.RegexRepeat(ops.Regex("a|b"), ops.Int(2), false), ops.Regex(`(?:a|b){2}`)},

		{ops.RegexRepeat(ops.Runex("a"), ops.Int(2), true), ops.Regex(`a{2}?`)},
		{ops.RegexRepeat(ops.Runex("a-z"), ops.Int(2), true), ops.Regex(`[a-z]{2}?`)},
		{ops.RegexRepeat(u8.Text("abc"), ops.Int(2), true), ops.Regex(`(?:\Qabc\E){2}?`)},
		{ops.RegexRepeat(ops.Regex("a|b"), ops.Int(2), true), ops.Regex(`(?:a|b){2}?`)},

		{ops.RegexRepeat(ops.Runex("a"), ops.NewPosIntRange(ops.Int(2), ops.Int(2)), false), ops.Regex(`a{2}`)},
		{ops.RegexRepeat(ops.Runex("a-z"), ops.NewPosIntRange(ops.Int(2), ops.Int(2)), false), ops.Regex(`[a-z]{2}`)},
		{ops.RegexRepeat(u8.Text("abc"), ops.NewPosIntRange(ops.Int(2), ops.Int(2)), false), ops.Regex(`(?:\Qabc\E){2}`)},
		{ops.RegexRepeat(ops.Regex("a|b"), ops.NewPosIntRange(ops.Int(2), ops.Int(2)), false), ops.Regex(`(?:a|b){2}`)},

		{ops.RegexRepeat(ops.Runex("a"), ops.NewPosIntRange(ops.Int(2), ops.Int(2)), true), ops.Regex(`a{2}?`)},
		{ops.RegexRepeat(ops.Runex("a-z"), ops.NewPosIntRange(ops.Int(2), ops.Int(2)), true), ops.Regex(`[a-z]{2}?`)},
		{ops.RegexRepeat(u8.Text("abc"), ops.NewPosIntRange(ops.Int(2), ops.Int(2)), true), ops.Regex(`(?:\Qabc\E){2}?`)},
		{ops.RegexRepeat(ops.Regex("a|b"), ops.NewPosIntRange(ops.Int(2), ops.Int(2)), true), ops.Regex(`(?:a|b){2}?`)},

		//--------------------------------------------------------------------------------
		{ops.RegexRepeat(ops.Runex("a"), ops.NewPosIntRange(ops.Int(2), ops.Inf), false), ops.Regex(`a{2,}`)},
		{ops.RegexRepeat(ops.Runex("a-z"), ops.NewPosIntRange(ops.Int(2), ops.Inf), false), ops.Regex(`[a-z]{2,}`)},
		{ops.RegexRepeat(u8.Text("abc"), ops.NewPosIntRange(ops.Int(2), ops.Inf), false), ops.Regex(`(?:\Qabc\E){2,}`)},
		{ops.RegexRepeat(ops.Regex("a|b"), ops.NewPosIntRange(ops.Int(2), ops.Inf), false), ops.Regex(`(?:a|b){2,}`)},

		{ops.RegexRepeat(ops.Runex("a"), ops.NewPosIntRange(ops.Int(2), ops.Inf), true), ops.Regex(`a{2,}?`)},
		{ops.RegexRepeat(ops.Runex("a-z"), ops.NewPosIntRange(ops.Int(2), ops.Inf), true), ops.Regex(`[a-z]{2,}?`)},
		{ops.RegexRepeat(u8.Text("abc"), ops.NewPosIntRange(ops.Int(2), ops.Inf), true), ops.Regex(`(?:\Qabc\E){2,}?`)},
		{ops.RegexRepeat(ops.Regex("a|b"), ops.NewPosIntRange(ops.Int(2), ops.Inf), true), ops.Regex(`(?:a|b){2,}?`)},

		//--------------------------------------------------------------------------------
		{ops.RegexRepeat(ops.Runex("a"), ops.NewPosIntRange(ops.Int(2), ops.Int(3)), false), ops.Regex(`a{2,3}`)},
		{ops.RegexRepeat(ops.Runex("a-z"), ops.NewPosIntRange(ops.Int(2), ops.Int(3)), false), ops.Regex(`[a-z]{2,3}`)},
		{ops.RegexRepeat(u8.Text("abc"), ops.NewPosIntRange(ops.Int(2), ops.Int(3)), false), ops.Regex(`(?:\Qabc\E){2,3}`)},
		{ops.RegexRepeat(ops.Regex("a|b"), ops.NewPosIntRange(ops.Int(2), ops.Int(3)), false), ops.Regex(`(?:a|b){2,3}`)},

		{ops.RegexRepeat(ops.Runex("a"), ops.NewPosIntRange(ops.Int(2), ops.Int(3)), true), ops.Regex(`a{2,3}?`)},
		{ops.RegexRepeat(ops.Runex("a-z"), ops.NewPosIntRange(ops.Int(2), ops.Int(3)), true), ops.Regex(`[a-z]{2,3}?`)},
		{ops.RegexRepeat(u8.Text("abc"), ops.NewPosIntRange(ops.Int(2), ops.Int(3)), true), ops.Regex(`(?:\Qabc\E){2,3}?`)},
		{ops.RegexRepeat(ops.Regex("a|b"), ops.NewPosIntRange(ops.Int(2), ops.Int(3)), true), ops.Regex(`(?:a|b){2,3}?`)},
	} {
		assertEqual(t, n.a, n.b)
		validateRegex(t, n.b)
	}
}

//================================================================================
func TestRegexNot(t *testing.T) {
	for _, n := range []struct{ a, b interface{} }{
		{ops.Not(ops.Runex("a")), ops.CharClass(`[^a]`)},
		{ops.Not(ops.Runex(`2a-c`)), ops.CharClass(`[^2a-c]`)},
		{ops.Not(ops.Runex(`^2a-c`)), ops.CharClass(`[2a-c]`)},
	} {
		assertEqual(t, n.a, n.b)
		validateRegex(t, n.b)
	}
}

//================================================================================
func TestRegexAlts(t *testing.T) {
	for _, n := range []struct{ a, b interface{} }{
		//Codepoint...
		{ops.Alt(ops.Runex("a"), ops.Runex("b")), ops.CharClass("[ab]")},     //point + point
		{ops.Alt(ops.Runex("a"), ops.Runex("g-j")), ops.CharClass("[ag-j]")}, //point + class
		{ops.Alt(ops.Runex("a"), ops.Runex("^g-j")), ops.Regex("a|[^g-j]")},  //point + negclass
		{ops.Alt(ops.Runex("a"), u8.Text("fg")), ops.Regex(`a|(?:\Qfg\E)`)},  //point + text
		{ops.Alt(ops.Runex("a"), ops.Regex("fg+")), ops.Regex("a|(?:fg+)")},  //point + regexp

		//u8.Text...
		{ops.Alt(u8.Text("ab"), ops.Runex("g")), ops.Regex(`(?:\Qab\E)|g`)},             //text + point
		{ops.Alt(u8.Text("abc"), ops.Runex("g-j")), ops.Regex(`(?:\Qabc\E)|[g-j]`)},     //text + class
		{ops.Alt(u8.Text("abc"), u8.Text("fgh")), ops.Regex(`(?:\Qabc\E)|(?:\Qfgh\E)`)}, //text + text
		{ops.Alt(u8.Text("abc"), ops.Regex("f|g")), ops.Regex(`(?:\Qabc\E)|(?:f|g)`)},   //text + regexp

		//tp.CharClass...
		{ops.Alt(ops.Runex("a-e"), ops.Runex("g")), ops.Runex("a-eg")},              //class + point
		{ops.Alt(ops.Runex("^a-e"), ops.Runex("g")), ops.Regex("[^a-e]|g")},         //negclass + point
		{ops.Alt(ops.Runex("a-e"), ops.Runex("g-j")), ops.CharClass("[a-eg-j]")},    //class + class
		{ops.Alt(ops.Runex("^a-e"), ops.Runex("^g-j")), ops.CharClass(`[^a-eg-j]`)}, //negclass + negclass
		{ops.Alt(ops.Runex("^a-e"), ops.Runex("g-j")), ops.Regex(`[^a-e]|[g-j]`)},   //negclass + class
		{ops.Alt(ops.Runex("a-e"), ops.Runex("^g-j")), ops.Regex(`[a-e]|[^g-j]`)},   //class + negclass
		{ops.Alt(ops.Runex("a-e"), u8.Text("fgh")), ops.Regex(`[a-e]|(?:\Qfgh\E)`)}, //class + text
		{ops.Alt(ops.Runex("a-e"), ops.Regex("fg*")), ops.Regex("[a-e]|(?:fg*)")},   //class + regexp

		//tp.Regex...
		{ops.Alt(ops.Regex("f|g"), ops.Runex("a")), ops.Regex("(?:f|g)|a")},           //regexp + point
		{ops.Alt(ops.Regex("f|g"), ops.Runex("a-e")), ops.Regex("(?:f|g)|[a-e]")},     //regexp + class
		{ops.Alt(ops.Regex("f|g"), u8.Text("abc")), ops.Regex(`(?:f|g)|(?:\Qabc\E)`)}, //regexp + text
		{ops.Alt(ops.Regex("f|g"), ops.Regex("J|k")), ops.Regex("(?:f|g)|(?:J|k)")},   //regexp + regexp

	} {
		assertEqual(t, n.a, n.b)
		validateRegex(t, n.b)
	}
}

//================================================================================
func TestRegexSeqs(t *testing.T) {
	for _, n := range []struct{ a, b interface{} }{
		{ops.Seq(ops.Runex("a"), ops.Runex("b")), u8.Text("ab")},           //point + point
		{ops.Seq(ops.Runex("a"), ops.Runex("g-j")), ops.Regex("a[g-j]")},   //point + class
		{ops.Seq(ops.Runex("a"), u8.Text("fg")), u8.Text("afg")},           //point + string
		{ops.Seq(ops.Runex("a"), ops.Regex("f|g")), ops.Regex("a(?:f|g)")}, //point + regexp

		{ops.Seq(ops.Runex("a-e"), ops.Runex("g")), ops.Regex("[a-e]g")},         //class + point
		{ops.Seq(ops.Runex("a-e"), ops.Runex("g-j")), ops.Regex("[a-e][g-j]")},   //class + class
		{ops.Seq(ops.Runex("a-e"), u8.Text("fgh")), ops.Regex("[a-e]fgh")},       //class + string
		{ops.Seq(ops.Runex("a-e"), ops.Regex("f|g")), ops.Regex("[a-e](?:f|g)")}, //class + regexp

		{ops.Seq(u8.Text("ab"), ops.Runex("g")), u8.Text("abg")},             //string + point
		{ops.Seq(u8.Text("abc"), ops.Runex("g-j")), ops.Regex("abc[g-j]")},   //string + class
		{ops.Seq(u8.Text("abc"), u8.Text("fgh")), u8.Text("abcfgh")},         //string + string
		{ops.Seq(u8.Text("abc"), ops.Regex("f|g")), ops.Regex("abc(?:f|g)")}, //string + regexp

		{ops.Seq(ops.Regex("f|g"), ops.Runex("a")), ops.Regex("(?:f|g)a")},         //regexp + point
		{ops.Seq(ops.Regex("f|g"), ops.Runex("a-e")), ops.Regex("(?:f|g)[a-e]")},   //regexp + class
		{ops.Seq(ops.Regex("f|g"), u8.Text("abc")), ops.Regex("(?:f|g)abc")},       //regexp + string
		{ops.Seq(ops.Regex("f|g"), ops.Regex("J|k")), ops.Regex("(?:f|g)(?:J|k)")}, //regexp + regexp
	} {
		assertEqual(t, n.a, n.b)
		validateRegex(t, n.b)
	}
}

//================================================================================
//TODO: TEST LeftShift - finish
func TestLeftShift(t *testing.T) {
}

//================================================================================
func TestRightShift(t *testing.T) {
	w1 := ops.Seq(ops.Group(ops.Runex("a-z"), "alpha"),
		ops.Group(ops.Seq(ops.Runex("0-9"),
			ops.Regex(ops.ToRegex(ops.Runex("0-9"))+"?")), "digits")) //used tp.ToRegex because kn.Optional removed

	for _, n := range []struct{ a, b interface{} }{
		{ops.RightShift("a", ops.Regex("a|b")).([][]ops.RegexMatch)[0][0].Match, "a"},
		{ops.RightShift("acb", ops.Regex("a|b")).([][]ops.RegexMatch)[1][0].Match, "b"},
		{ops.RightShift("abcdefg", ops.Regex("a|(c)|e")).([][]ops.RegexMatch)[1][1].Match, "c"},

		{ops.RightShift("a", ops.Regex("a|b")), [][]ops.RegexMatch{{{"", 0, 1, "a"}}}},
		{ops.RightShift("a", u8.Text("a|b")), [][]ops.RegexMatch{{{"", 0, 1, "a"}}}},
		{ops.RightShift("acb", ops.Regex("a|b")), [][]ops.RegexMatch{{{"", 0, 1, "a"}},
			{{"", 2, 3, "b"}}}},
		{ops.RightShift("abcdefg", ops.Regex("a|c|e")), [][]ops.RegexMatch{{{"", 0, 1, "a"}},
			{{"", 2, 3, "c"}},
			{{"", 4, 5, "e"}}}},
		{ops.RightShift("abcdefg", ops.Regex("a|(c)|e")), [][]ops.RegexMatch{{{"", 0, 1, "a"}, {"", -1, -1, ""}},
			{{"", 2, 3, "c"}, {"", 2, 3, "c"}},
			{{"", 4, 5, "e"}, {"", -1, -1, ""}}}},

		{w1, ops.Regex(`(?:(?P<alpha>[a-z]))(?:(?P<digits>[0-9](?:[0-9]?)))`)},
		{ops.RightShift("a7, b22, c86, d99, e3", w1), [][]ops.RegexMatch{
			{{"", 0, 2, "a7"}, {"alpha", 0, 1, "a"}, {"digits", 1, 2, "7"}},
			{{"", 4, 7, "b22"}, {"alpha", 4, 5, "b"}, {"digits", 5, 7, "22"}},
			{{"", 9, 12, "c86"}, {"alpha", 9, 10, "c"}, {"digits", 10, 12, "86"}},
			{{"", 14, 17, "d99"}, {"alpha", 14, 15, "d"}, {"digits", 15, 17, "99"}},
			{{"", 19, 21, "e3"}, {"alpha", 19, 20, "e"}, {"digits", 20, 21, "3"}},
		}},
	} {
		assertEqual(t, n.a, n.b)
		validateRegex(t, n.b)
	}
}

//================================================================================
func TestRegexGroups(t *testing.T) {
	for _, n := range []struct{ a, b interface{} }{
		{ops.Group(ops.Runex("a"), ""), ops.Regex(`(a)`)},
		{ops.Group(ops.Runex("a-z"), ""), ops.Regex(`([a-z])`)},
		{ops.Group(u8.Text("abc"), ""), ops.Regex(`(\Qabc\E)`)},
		{ops.Group(ops.Regex("a|b"), ""), ops.Regex(`(a|b)`)},

		{ops.Group(ops.Runex("a"), "aName"), ops.Regex(`(?P<aName>a)`)},
		{ops.Group(ops.Runex("a-z"), "aName"), ops.Regex(`(?P<aName>[a-z])`)},
		{ops.Group(u8.Text("abc"), "aName"), ops.Regex(`(?P<aName>\Qabc\E)`)},
		{ops.Group(ops.Regex("a|b"), "aName"), ops.Regex(`(?P<aName>a|b)`)},
	} {
		assertEqual(t, n.a, n.b)
		validateRegex(t, n.b)
	}
}

//================================================================================
func TestRegexParenthesize(t *testing.T) {
	for _, n := range []struct{ a, b interface{} }{
		{ops.Parenthesize(ops.Runex("a")), ops.Runex("a")},
		{ops.Parenthesize(u8.Text("abc")), u8.Text("abc")},
		{ops.Parenthesize(ops.Runex("a-z")), ops.Regex(`(?:[a-z])`)},
		{ops.Parenthesize(ops.Regex("a|b")), ops.Regex(`(?:a|b)`)},
	} {
		assertEqual(t, n.a, n.b)
		validateRegex(t, n.b)
	}
}

//================================================================================
func TestRegexMisc(t *testing.T) {
	pa := ops.ToParser(ops.Runex("a"))
	_, _ = parsec.ParseItem(pa, u8.Text("abc"))

}

//================================================================================
func TestBooleanExprs(t *testing.T) {
	assert(t, !ops.Not(true).(bool))
	assert(t, ops.Not(false).(bool))

	assert(t, ops.And(true, func() interface{} { return true }))
	assert(t, !ops.And(false, func() interface{} { return true }))
	assert(t, !ops.And(true, func() interface{} { return false }))
	assert(t, !ops.And(false, func() interface{} { return false }))

	assert(t, ops.Or(true, func() interface{} { return true }))
	assert(t, ops.Or(false, func() interface{} { return true }))
	assert(t, ops.Or(true, func() interface{} { return false }))
	assert(t, !ops.Or(false, func() interface{} { return false }))
}

//================================================================================

//TODO: TEST ToRegex
//TODO: TEST ToParser
//TODO: TEST Parenthesize - finish
//TODO: TEST Not - finish
//TODO: TEST Alt - finish
//TODO: TEST Seq - finish
//TODO: TEST Matcher
//TODO: TEST FindFirst
//TODO: TEST Xor
//TODO: TEST SeqXor
//TODO: TEST MultAssign
//TODO: TEST DivideAssign
//TODO: TEST ModAssign
//TODO: TEST LeftShiftAssign
//TODO: TEST RightShiftAssign
//TODO: TEST SeqAssign
//TODO: TEST SeqXorAssign
//TODO: TEST AltAssign
//TODO: TEST XorAssign
//TODO: TEST Incr - finish
//TODO: TEST Decr - finish

//================================================================================
