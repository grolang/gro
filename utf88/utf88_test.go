// Copyright 2009-17 The Go and Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package utf88_test

import (
	u8 "github.com/grolang/gro/utf88"
	"reflect"
	"testing"
)

var surPointTests = []struct {
	in  u8.Codepoint
	out string
}{
	{-1, string(0xfffd)},
	{0, string(0)},
	{1, string(1)},
	{4, string(4)},
	{'a', "a"},
	{'b', "b"},
	{0x7f, "\u007f"},                         //1-byte max
	{0x80, string(0x80)},                     //1-byte max + 1
	{0x7ff, string(0x7ff)},                   //2-byte max
	{0xd7ff, string(0xd7ff)},                 //min utf16 surr - 1
	{0xd800, string(0xd800)},                 //min utf16 surr
	{u8.ParsePoint("ud800"), string(0xfffd)}, //min utf16 surr
	{0xdfff, string(0xdfff)},                 //max utf16 surr
	{0xe000, string(0xe000)},                 //max utf16 surr + 1
	{0xfffd, string(0xfffd)},                 //error rune
	{0xffff, string(0xffff)},                 //3-byte max
	{0xf8000, "\U000f8000"},                  //surrogate base
	{0xf8010, string(0xf8010)},               //min hi surr - 1 -> itself
	{0xf8011, "\ufffd"},                      //min hi surr -> error rune
	{0xfffbf, string(0xfffd)},                //max hi surr -> error rune
	{0xfffc0, string(0xfffc0)},               //max hi surr + 1 -> itself
	{0xfffff, string(0xfffff)},               //min lo surr - 1 -> itself
	{0x100000, string(0x100000)},             //min lo surr -> itself
	{0x10ffff, string(0x10ffff)},             //max lo surr -> itself

	{0x110000, "\U000f8011\U00100000"}, //max lo surr + 1
	{0x11ffff, string([]rune{0xf8011, 0x10ffff})},
	{0x120000, string([]rune{0xf8012, 0x100000})},
	{0x1fffff, string([]rune{0xf801f, 0x10ffff})},   //4-byte max
	{0x3ffffff, string([]rune{0xf83ff, 0x10ffff})},  //5-byte max
	{0x7fbfffff, string([]rune{0xfffbf, 0x10ffff})}, //max encodable rune
	{u8.ParsePoint("U7fbffffff"), string([]rune{0xfffbf, 0x10ffff})},
	{0x7fc00000, string(0xfffd)}, //max encodable rune + 1
	{0x7fffffff, string(0xfffd)}, //max go rune
	{u8.ParsePoint("U7fffffff"), string(0xfffd)},
	{-0x80000000, string(0xfffd)}, //max go rune + 1
}

func TestSurPoints(t *testing.T) {
	for i, tt := range surPointTests {
		ri := u8.Sur(tt.in)
		so := tt.out
		if !reflect.DeepEqual(ri, so) {
			t.Errorf("%d: %#x != %#x", i, ri, so)
		}
	}
}

func TestParsePoint(t *testing.T) {
	for i, tt := range []struct {
		s string
		r u8.Codepoint
	}{
		{"ud800", 0xd800},
		{"U7ffdffff", 0x7ffdffff},
	} {
		s := u8.ParsePoint(tt.s)
		r := tt.r
		if !reflect.DeepEqual(s, r) {
			t.Errorf("%d: %#x != %#x", i, s, r)
		}
	}
}

func TestParsePoints(t *testing.T) {
	for i, tt := range []struct {
		ss string
		rs u8.Text
	}{
		{"ud800", u8.Text{0xd800}},
		{"ud800ud801", u8.Text{0xd800, 0xd801}},
		{"ud800U7fbfffff", u8.Text{0xd800, 0x7fbfffff}},
		{"U7fbfffff", u8.Text{0x7fbfffff}},
	} {
		ss := u8.ParsePoints(tt.ss)
		rs := tt.rs
		if !reflect.DeepEqual(ss, rs) {
			t.Errorf("%d: %#x != %#x", i, ss, rs)
		}
	}
}

var stringTests = []struct {
	in  string
	out string
}{
	//let's see how 'string' cast works...
	{string(-1), string(0xfffd)},
	{"abc" + string(100) + "efg", "abcdefg"},
	{"123" + string(0x10ffff) + "456", "123\U0010ffff456"},
	{"123" + string(0x110000) + "456", "123\ufffd456"},

	{string([]byte("\xed\x9f\xbf")), "\ud7ff"},       //min utf16 surr - 1
	{string([]byte("\xed\x9f\xbf")), "\xed\x9f\xbf"}, //min utf16 surr - 1
	{string(0xd800), "\ufffd"},                       //min utf16 surr
	{string([]byte("\xed\x9f\xc0")), "\xed\x9f\xc0"}, //min utf16 surr
	{string(0xdfff), "\ufffd"},                       //max utf16 surr
	{string("\xee\x80\x7f"), "\xee\x80\x7f"},         //max utf16 surr

	// Decode function...
	{u8.Decode("123\xff456"), "123\ufffd456"},
	{string("123\xff456"), "123\xff456"},

	//multiple args in Sur function...
	{"123" + string(0x7fbfffff) + "456", "123\ufffd456"},
	{"123" + u8.Sur(0x7fbfffff) + "456", "123\U000fffbf\U0010ffff456"},

	{u8.Sur('a', 0xffff, 0x110000, 0x11ffff, 0x120000, 0x7fbf0000),
		"a\uffff\U000f8011\U00100000\U000f8011\U0010ffff\U000f8012\U00100000\U000fffbf\U00100000"},

	{u8.Sur('a', 'b', 'c'), "abc"},

	{u8.Sur(1, 2, 3, 4),
		string([]rune{1, 2, 3, 4})},

	{u8.Sur('a', 'b', 0xd7ff, 0xd800, 0xdfff, 0xe000, 0x110000, -1),
		string([]rune{'a', 'b', 0xd7ff, 0xd800, 0xdfff, 0xe000, 0xf8011, 0x100000, 0xfffd})},

	{u8.Sur(0xf8010, 0xf8011, 0xfffbf, 0xfffc0, 0xfffff, 0x100000, 0x10ffff, 0x7fbfffff, 0x7fc00000, 0x7fffffff, -0x80000000),
		string([]rune{0xf8010, 0xfffd, 0xfffd, 0xfffc0, 0xfffff, 0x100000, 0x10ffff, 0xfffbf, 0x10ffff, 0xfffd, 0xfffd, 0xfffd})},
}

func TestStrings(t *testing.T) {
	for i, tt := range stringTests {
		si := tt.in
		so := tt.out
		if !reflect.DeepEqual(si, so) {
			t.Errorf("%d: %#x != %#x", i, si, so)
		}
	}
}

var desurTests = []struct {
	in  string
	out u8.Text
}{
	{string(-1), u8.Text{0xfffd}},
	{string(0), u8.Text{0}},
	{string(1), u8.Text{1}},
	{"a", u8.Text{'a'}},
	{"\u007f", u8.Text{0x7f}},         //1-byte max
	{"\u0080", u8.Text{0x80}},         //1-byte max + 1
	{"\u07ff", u8.Text{0x7ff}},        //2-byte max
	{"\ud7ff", u8.Text{0xd7ff}},       //min utf16 surr - 1
	{string(0xd800), u8.Text{0xfffd}}, //min utf16 surr
	{string(0xdfff), u8.Text{0xfffd}}, //max utf16 surr
	{"\ue000", u8.Text{0xe000}},       //max utf16 surr + 1
	{"\ufffd", u8.Text{0xfffd}},       //error rune
	{"\uffff", u8.Text{0xffff}},       //3-byte max
	{"\U000f8000", u8.Text{0xf8000}},  //surr base
	{"\U000f8010", u8.Text{0xf8010}},  //min hi surr - 1

	{"\U000f8011\U00100000", u8.Text{0x110000}}, //min hi surr
	{"\U000f8011\U0010ffff", u8.Text{0x11ffff}},
	{"\U000f8012\U00100000", u8.Text{0x120000}},
	{"\U000f8012\U0010ffff", u8.Text{0x12ffff}},
	{"\U000f8fff\U00100000", u8.Text{0xfff0000}},
	{"\U000f8fff\U0010ffff", u8.Text{0xfffffff}},
	{"\U000fffbf\U0010ffff", u8.Text{0x7fbfffff}}, //max hi surr

	{"\U000fffc0", u8.Text{0xfffc0}},
	{"\U000fffff", u8.Text{0xfffff}},
	{"\U00100000", u8.Text{0x100000}}, //min lo surr behaves like normal codepoint
	{"\U0010ffff", u8.Text{0x10ffff}}, //max lo surr behaves like normal codepoint

	{string(0x110000), u8.Text{0xfffd}},
	{string(0x7fbfffff), u8.Text{0xfffd}},
	{string(0x7fffffff), u8.Text{0xfffd}},
	{string(-0x80000000), u8.Text{0xfffd}},

	{"\U000f8011" + "a", u8.Text{0xfffd, 'a'}},         //surr with illegal trailer
	{"\U000f8011\U000fffff", u8.Text{0xfffd, 0xfffff}}, //surr with first valid trailer - 1
	{"\U000f8011\U00100000", u8.Text{0x110000}},        //surr with first valid trailer
	{"\U000f8011\U0010ffff", u8.Text{0x11ffff}},        //surr with last valid trailer

	{string([]rune{0xf8011, 0x110000}), u8.Text{0xfffd, 0xfffd}}, //surr with last valid trailer + 1

	{"\U000f8011", u8.Text{0xfffd}}, //surr with no trailer

	{"\uffff\U000f8011\U00100000\U000f8011\U0010ffff\U000fffbf\U00100000\U000fffbf\U0010ffff\U000fffc0\U000fffff",
		u8.Text{0xffff, 0x110000, 0x11ffff, 0x7fbf0000, 0x7fbfffff, 0xfffc0, 0xfffff}},

	{string([]rune{1, 2, 3, 4}),
		u8.Text{1, 2, 3, 4}},

	{string([]rune{0xf8011, 'a'}),
		u8.Text{0xfffd, 'a'}},

	{string([]rune{0xfffbf}),
		u8.Text{0xfffd}},
}

func TestDesur(t *testing.T) {
	for i, tt := range desurTests {
		ri := u8.Desur(tt.in)
		ro := tt.out
		if !reflect.DeepEqual(ri, ro) {
			t.Errorf("%d: %#x != %#x", i, ri, ro)
		}
	}
}

var desurLastTests = []struct {
	in  []rune
	out u8.Codepoint
	len int
}{
	{[]rune{1, 2, 3}, 3, 1},
	{[]rune("abc"), 'c', 1},
	{[]rune{'a', 0x7f}, 0x7f, 1},       //1-byte max
	{[]rune{'a', 0x80}, 0x80, 1},       //1-byte max + 1
	{[]rune{'a', 0x7ff}, 0x7ff, 1},     //2-byte max
	{[]rune{'a', 0xd7ff}, 0xd7ff, 1},   //min utf16 surr - 1
	{[]rune{'a', 0xd800}, 0xd800, 1},   //min utf16 surr
	{[]rune{'a', 0xdfff}, 0xdfff, 1},   //max utf16 surr
	{[]rune{'a', 0xe000}, 0xe000, 1},   //max utf16 surr + 1
	{[]rune{'a', 0xfffd}, 0xfffd, 1},   //error rune
	{[]rune{'a', 0xffff}, 0xffff, 1},   //3-byte max
	{[]rune{'a', 0xf8000}, 0xf8000, 1}, //surr base
	{[]rune{'a', 0xf8010}, 0xf8010, 1}, //min hi surr - 1
	{[]rune{'a', 0xf8011}, 0xfffd, 1},  //min hi surr
	{[]rune{'a', 0xfffbf}, 0xfffd, 1},  //max hi surr
	{[]rune("z\U000fffc0"), 0xfffc0, 1},
	{[]rune("z\U000fffff"), 0xfffff, 1},

	{[]rune("z\U000f8011\U00100000"), 0x110000, 2}, //min lo surr preceded by min hi surr
	{[]rune("z\U000f8012\U00100000"), 0x120000, 2},
	{[]rune("z\U000f8fff\U00100000"), 0xfff0000, 2},
	{[]rune("z\U000fffbf\U00100000"), 0x7fbf0000, 2}, //min lo surr preceded by max hi surr
	{[]rune("z\U00100000"), 0x100000, 1},             //min lo surr not preceded by hi surr, behaves like normal codepoint

	{[]rune("z\U000f8011\U0010abcd"), 0x11abcd, 2}, //lo surr preceded by min hi surr
	{[]rune("z\U000f8012\U0010abcd"), 0x12abcd, 2},
	{[]rune("z\U000fffbf\U0010abcd"), 0x7fbfabcd, 2}, //lo surr preceded by max hi surr
	{[]rune("z\U0010abcd"), 0x10abcd, 1},             //lo surr not preceded by hi surr

	{[]rune("z\U000f8011\U0010ffff"), 0x11ffff, 2}, //max lo surr preceded by min hi surr
	{[]rune("z\U000f8012\U0010ffff"), 0x12ffff, 2},
	{[]rune("z\U000f8fff\U0010ffff"), 0xfffffff, 2},
	{[]rune("z\U000fffbf\U0010ffff"), 0x7fbfffff, 2}, //max lo surr preceded by max hi surr
	{[]rune("z\U0010ffff"), 0x10ffff, 1},             //max lo surr not preceded by hi surr, behaves like normal codepoint

	{[]rune{'a', 0x110000}, 0xfffd, 1},
	{[]rune{'a', 0x7fbfffff}, 0xfffd, 1},
	{[]rune{'a', 0x7fffffff}, 0xfffd, 1},
	{[]rune{'a', -0x80000000}, 0xfffd, 1},
}

func TestLastDesur(t *testing.T) {
	for i, tt := range desurLastTests {
		ri, ci := u8.DesurrogateLastPoint(tt.in)
		ro := tt.out
		co := tt.len
		if ci != co || ri != ro {
			t.Errorf("%d: sizes %d != %d, or runes %#x != %#x", i, ci, co, ri, ro)
		}
	}
}

var pointCountTests = []struct {
	in    []rune
	count int
}{
	{[]rune{0xfffd}, 1},
	{[]rune{0}, 1},
	{[]rune{1}, 1},
	{[]rune{'a'}, 1},
	{[]rune{0x7f}, 1},              //1-byte max
	{[]rune{0x80}, 1},              //1-byte max + 1
	{[]rune{0x7ff}, 1},             //2-byte max
	{[]rune{0xd7ff}, 1},            //min utf16 surr - 1
	{[]rune{0xd800}, 1},            //min utf16 surr
	{[]rune{0xdfff}, 1},            //max utf16 surr
	{[]rune{0xe000}, 1},            //max utf16 surr + 1
	{[]rune{0xfffd}, 1},            //error rune
	{[]rune{0xffff}, 1},            //3-byte max
	{[]rune{0xf8000}, 1},           //surr base
	{[]rune{0xf8010}, 1},           //min hi surr - 1
	{[]rune{0xf8011, 0x100000}, 1}, //min hi surr
	{[]rune{0xf8011, 0x10ffff}, 1},
	{[]rune{0xf8012, 0x100000}, 1},
	{[]rune{0xf8012, 0x10ffff}, 1},
	{[]rune{0xf8fff, 0x100000}, 1},
	{[]rune{0xf8fff, 0x10ffff}, 1},
	{[]rune{0xfffbf, 0x10ffff}, 1}, //max hi surr

	{[]rune{0xfffc0}, 1},
	{[]rune{0xfffff}, 1},
	{[]rune{0x100000}, 1}, //min lo surr behaves like normal codepoint
	{[]rune{0x10ffff}, 1}, //max lo surr behaves like normal codepoint

	{[]rune{0x110000}, 1},
	{[]rune{0x7fbfffff}, 1},
	{[]rune{0x7fffffff}, 1},
	{[]rune{-0x80000000}, 1},

	{[]rune{0xf8011, 'a'}, 2},      //surr with illegal trailer
	{[]rune{0xf8011, 0xfffff}, 2},  //surr with first valid trailer - 1
	{[]rune{0xf8011, 0x100000}, 1}, //surr with first valid trailer
	{[]rune{0xf8011, 0x10ffff}, 1}, //surr with last valid trailer
	{[]rune{0xf8011, 0x110000}, 2}, //surr with last valid trailer + 1
	{[]rune{0xf8011}, 1},           //surr with no trailer

	{[]rune{0xffff, 0x110000, 0x11ffff, 0x7fbf0000, 0x7fbfffff, 0xfffc0, 0xfffff}, 7},
	{[]rune{1, 2, 3, 4}, 4},
	{[]rune{0xfffd, 'a'}, 2},
	{[]rune{0xfffd}, 1},
}

func TestPointCountOfRunes(t *testing.T) {
	for i, tt := range pointCountTests {
		li := u8.PointCountOfRunes(tt.in)
		lo := tt.count
		if li != lo {
			t.Errorf("%d: %#x != %#x", i, li, lo)
		}
	}
}

var bytesTests = []struct {
	in     []byte
	points int
	valid  bool
}{
	{[]byte(string(0)), 1, true},
	{[]byte("\u0000abc"), 4, true},
	{[]byte("\u007f"), 1, true},             //1-byte max
	{[]byte("\u0080"), 1, true},             //1-byte max + 1
	{[]byte("\u07ff"), 1, true},             //2-byte max
	{[]byte("\u007f\u0080\u07ff"), 3, true}, //example multi-rune string

	{[]byte("\ud7ff"), 1, true}, //min utf16 surr - 1

	//utf16 surrogates unrepresentable in string format, so each byte is converted to a separate rune...
	{[]byte("\xed\x9f\xc0"), 3, false}, //min utf16 surr
	{[]byte("\xee\x80\x7f"), 3, false}, //max utf16 surr

	{[]byte("\ue000"), 1, true},      //max utf16 surr + 1
	{[]byte("\ufffd"), 1, true},      //error rune
	{[]byte("\uffff"), 1, true},      //3-byte max
	{[]byte("\U000f8000"), 1, true},  //surr base
	{[]byte("\U000f8010"), 1, true},  //min hi surr - 1
	{[]byte("\U000f8011"), 1, false}, //min hi surr with no lo surr
	{[]byte("\U000fffbf"), 1, false}, //max hi surr with no lo surr
	{[]byte("\U000fffc0"), 1, true},  //max hi surr + 1
	{[]byte("\U000fffff"), 1, true},
	{[]byte("\U00100000"), 1, true}, //min lo surr
	{[]byte("\U0010ffff"), 1, true}, //max lo surr

	{[]byte(u8.Sur(0x110000)), 1, true},
	{[]byte("\U000f8011\U00100000"), 1, true},
	{[]byte(u8.Sur(0x11ffff)), 1, true},
	{[]byte(u8.Sur(0x120000)), 1, true},
	{[]byte(u8.Sur(0xfff0000)), 1, true},
	{[]byte(u8.Sur(0x7fbfffff)), 1, true}, //max utf88 rune

	{[]byte(u8.Sur(0x7fffffff)), 1, true}, //ValidBytes true because it's already produced a validly embedded ErrorRune

	{[]byte("\U000f8011" + "a"), 2, false}, // surr with illegal trailer
	{[]byte(u8.Sur(0x110000) + u8.Sur(0x7fbfffff) + "a"), 3, true},
}

func TestPointCountOfBytes(t *testing.T) {
	for i, tt := range bytesTests {
		ci := u8.PointCountOfBytes(tt.in)
		co := tt.points
		if ci != co {
			t.Errorf("%d: %#x != %#x", i, ci, co)
		}
	}
}

func TestValidRunes(t *testing.T) {
	for i, tt := range []struct {
		rs []rune
		b  bool
	}{
		{[]rune{0xd7ff}, true},
		{[]rune{0xd7ff, 0x7fbfffff}, false},
	} {
		rs := u8.ValidRunes(tt.rs)
		b := tt.b
		if rs != b {
			t.Errorf("%d: %#x != %#x", i, rs, b)
		}
	}
}

func TestValidBytes(t *testing.T) {
	for i, tt := range bytesTests {
		bi := u8.ValidBytes(tt.in)
		bo := tt.valid
		if bi != bo {
			t.Errorf("%d: %#x != %#x", i, bi, bo)
		}
	}
}

var lenValidTests = []struct {
	in         u8.Codepoint
	len        int
	valid      bool
	lenBytes   int
	validBytes bool
}{
	{-1, 1, false, -1, false},
	{0, 1, true, 1, true},
	{'a', 1, true, 1, true},
	{0x7f, 1, true, 1, true},       //1-byte max
	{0x80, 1, true, 2, true},       //1-byte max + 1
	{0x7ff, 1, true, 2, true},      //2-byte max
	{0xd7ff, 1, true, 3, true},     //min utf16 surr - 1
	{0xd800, 1, true, -1, false},   //min utf16 surr
	{0xdfff, 1, true, -1, false},   //max utf16 surr
	{0xe000, 1, true, 3, true},     //max utf16 surr + 1
	{0xfffd, 1, true, 3, true},     //error rune
	{0xffff, 1, true, 3, true},     //3-byte max
	{0xf8000, 1, true, 4, true},    //surrogate base
	{0xf8010, 1, true, 4, true},    //min hi surr - 1
	{0xf8011, 1, false, -1, false}, //min hi surr
	{0xfffbf, 1, false, -1, false}, //max hi surr
	{0xfffc0, 1, true, 4, true},    //max hi surr + 1
	{0xfffff, 1, true, 4, true},    //min lo surr - 1
	{0x100000, 1, true, 4, true},   //min lo surr
	{0x10ffff, 1, true, 4, true},   //max lo surr
	{0x110000, 2, true, 8, true},   //max lo surr + 1
	{0x11ffff, 2, true, 8, true},
	{0x120000, 2, true, 8, true},
	{0x1fffff, 2, true, 8, true},       //4-byte max
	{0x3ffffff, 2, true, 8, true},      //5-byte max
	{0x7fbfffff, 2, true, 8, true},     //max encodable rune
	{0x7fc00000, 1, false, -1, false},  //max encodable rune + 1
	{0x7fffffff, 1, false, -1, false},  //max go rune
	{-0x80000000, 1, false, -1, false}, //max go rune + 1
}

func TestLenInRunes(t *testing.T) {
	for i, tt := range lenValidTests {
		li := u8.LenInRunes(tt.in)
		lo := tt.len
		if li != lo {
			t.Errorf("%d: %#x != %#x", i, li, lo)
		}
	}
}

func TestLenInBytes(t *testing.T) {
	for i, tt := range lenValidTests {
		li := u8.LenInBytes(tt.in)
		lo := tt.lenBytes
		if li != lo {
			t.Errorf("%d: %#x != %#x", i, li, lo)
		}
	}
}

func TestValidForSurrogation(t *testing.T) {
	for i, tt := range lenValidTests {
		bi := u8.ValidForSurrogation(tt.in)
		bo := tt.valid
		if bi != bo {
			t.Errorf("%d: %#b != %#b", i, bi, bo)
		}
	}
}

func TestValidForEncoding(t *testing.T) {
	for i, tt := range lenValidTests {
		bi := u8.ValidForEncoding(tt.in)
		bo := tt.validBytes
		if bi != bo {
			t.Errorf("%d: %#b != %#b", i, bi, bo)
		}
	}
}
