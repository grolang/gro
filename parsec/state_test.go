// Copyright 2017 The Gro authors. All rights reserved.
// Portions translated from Armando Blancas' Clojure-based Kern parser.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package parsec

import (
	"testing"

	ts "github.com/grolang/gro/assert"
	u8 "github.com/grolang/gro/utf88"
)

//================================================================================
func TestInputAndPosition(t *testing.T) {
	ts.LogAsserts("InputAndPosition", t, func(tt *ts.T) {
		letter := Regexp(`\pL`)
		p1 := Flatten(letter, letter, letter)

		tt.AssertEqual(parseStr(SeqRight(SetInput(u8.Desur("Hi!")), GetInput).(Parser), "abcdefg"),
			PState{input: "Hi!", pos: 0, value: u8.Desur("Hi!"), empty: true, ok: true})
		tt.AssertEqual(parseStr(SeqRight(letter, GetInput).(Parser), "abcdefg"),
			PState{input: "bcdefg", pos: 1, value: u8.Desur("bcdefg"), ok: true})

		tt.AssertEqual(parseStr(SeqRight(SetPosition(100), GetPosition).(Parser), "abcdefg"),
			PState{input: "abcdefg", pos: 100, value: uint64(100), empty: true, ok: true})

		tt.AssertEqual(parseStr(SeqRight(SetInputAndPosition(InputAndPosition{u8.Desur("Hi there!"), 5}), GetInputAndPosition).(Parser), "abcdefg"),
			PState{input: "Hi there!", pos: 5, value: InputAndPosition{u8.Desur("Hi there!"), 5}, empty: true, ok: true})

		tt.AssertEqual(parseStr(Collect(p1, SeqRight(SetInput(u8.Desur("XYZ")), p1).(Parser)), "ABCDEFG"),
			PState{input: "", pos: 6, value: arrText("ABC", "XYZ"), ok: true}) //TODO: what should pos be?
		tt.AssertEqual(parseStr(Collect(p1, SeqRight(SetInput(u8.Desur("XYZ")), p1).(Parser)), "ABC"),
			PState{input: "", pos: 6, value: arrText("ABC", "XYZ"), ok: true}) //TODO: ditto

		tt.AssertEqual(parseStr(Collect(p1), "ABC"),
			PState{input: "", pos: 3, value: arrText("ABC"), ok: true})

		p2 := SeqRight(Skip(SetInput(u8.Text("XY0")), SetPosition(0)), p1)
		tt.AssertEqual(parseStr(Collect(p1, p2.(Parser)), "ABC"),
			PState{input: "0", pos: 2, value: nil, error: makeUnexpInp("0")})

		p3 := SeqRight(Skip(SetInput(u8.Text("WXYZ")), SetPosition(0)), GetPosition)
		tt.AssertEqual(parseStr(SeqRight(p1, p3).(Parser), "ABC"),
			PState{input: "WXYZ", pos: 0, value: uint64(0), ok: true})

		p4 := SeqRight(SetInput(u8.Text("XYZ")), GetInput)
		tt.AssertEqual(parseStr(SeqRight(p1, p4).(Parser), "ABC"),
			PState{input: "XYZ", pos: 3, value: u8.Text("XYZ"), ok: true})

		tt.AssertEqual(parseStr(Collect(p1, SeqRight(SetInputAndPosition(InputAndPosition{u8.Desur("XYZ"), 2}), p1).(Parser)), "ABCDEFG"), //posn changed from 3 to 2...
			PState{input: "", pos: 5, value: arrText("ABC", "XYZ"), ok: true}) //...so posn is 5 instead of 6

	})
}

//================================================================================
func TestState(t *testing.T) {
	ts.LogAsserts("State", t, func(tt *ts.T) {
		incrIfNewline := func(x interface{}) Parser {
			if x.(u8.Codepoint) == u8.Codepoint('\n') {
				return ModifyState(func(n interface{}, _ ...interface{}) interface{} {
					return n.(int) + 1
				})
			} else {
				return Return(nil)
			}
		}

		p2 := SeqRight(PutState(21),
			Bind(AnyChar, incrIfNewline)).(Parser)
		tt.AssertEqual(parseStr(p2, "\naa"),
			PState{input: "aa", pos: 1, value: u8.Codepoint('\n'), ok: true, user: 22})
		tt.AssertEqual(parseStr(p2, "aaa"),
			PState{input: "aa", pos: 1, value: nil, ok: true, user: 21})
		tt.AssertEqual(parseStr(p2, "aa\naa\na"),
			PState{input: "a\naa\na", pos: 1, value: nil, ok: true, user: 21})

		p1 := SeqRight(PutState(21),
			Bind(AnyChar, incrIfNewline),
			GetState).(Parser)
		tt.AssertEqual(parseStr(p1, "\naa"),
			PState{input: "aa", pos: 1, value: 22, ok: true, user: 22})
		tt.AssertEqual(parseStr(p1, "aaa"),
			PState{input: "aa", pos: 1, value: 21, ok: true, user: 21})
		tt.AssertEqual(parseStr(p1, "aa\naa\na"),
			PState{input: "a\naa\na", pos: 1, value: 21, ok: true, user: 21})

		p3 := SeqRight(PutState(21),
			Bind(AnyChar, incrIfNewline),
			AnyChar).(Parser)
		tt.AssertEqual(parseStr(p3, "\naa"),
			PState{input: "a", pos: 2, value: u8.Codepoint('a'), ok: true, user: 22})
		tt.AssertEqual(parseStr(p3, "aaa"),
			PState{input: "a", pos: 2, value: u8.Codepoint('a'), ok: true, user: 21})
		tt.AssertEqual(parseStr(p3, "aa\naa\na"),
			PState{input: "\naa\na", pos: 2, value: u8.Codepoint('a'), ok: true, user: 21})
		tt.AssertEqual(parseStr(p3, "a\naa\na"),
			PState{input: "aa\na", pos: 2, value: u8.Codepoint('\n'), ok: true, user: 21})
	})
}

//================================================================================
func TestStateMap(t *testing.T) {
	ts.LogAsserts("StateMap", t, func(tt *ts.T) {
		init := PutStateMapEntry("counter", 123)

		//--------------------------------------------------------------------------------
		getB := func(x interface{}) Parser {
			if x.(u8.Codepoint) == u8.Codepoint('\n') {
				return GetStateMapEntry("counter")
			} else {
				return Return(nil)
			}
		}

		p1 := SeqRight(init, Bind(AnyChar, getB)).(Parser)
		tt.AssertEqual(parseStr(p1, "\nbc"),
			PState{input: "bc", pos: 1, value: 123, ok: true, user: map[string]interface{}{"counter": 123}})
		tt.AssertEqual(parseStr(p1, "abc"),
			PState{input: "bc", pos: 1, value: nil, ok: true, user: map[string]interface{}{"counter": 123}})

		//--------------------------------------------------------------------------------
		solePutB := func(x interface{}) (p Parser) {
			if x.(u8.Codepoint) == u8.Codepoint('\n') {
				return Return(nil)
			} else {
				return PutStateMapEntry("counter", 777)
			}
		}

		p2 := SeqRight(init, Bind(AnyChar, solePutB)).(Parser)
		tt.AssertEqual(parseStr(p2, "\nbc"),
			PState{input: "bc", pos: 1, value: nil, ok: true, user: map[string]interface{}{"counter": 123}})
		tt.AssertEqual(parseStr(p2, "abc"),
			PState{input: "bc", pos: 1, value: map[string]interface{}{"counter": 123}, ok: true, user: map[string]interface{}{"counter": 777}})

		//--------------------------------------------------------------------------------
		putB := func(x interface{}) (p Parser) {
			if x.(u8.Codepoint) == u8.Codepoint('\n') {
				return GetStateMapEntry("counter")
			} else {
				return Bind(PutStateMapEntry("counter", 777), func(_ interface{}) Parser { return GetStateMapEntry("counter") })
			}
		}

		p3 := SeqRight(init, Bind(AnyChar, putB)).(Parser)
		tt.AssertEqual(parseStr(p3, "\nbc"),
			PState{input: "bc", pos: 1, value: 123, ok: true, user: map[string]interface{}{"counter": 123}})
		tt.AssertEqual(parseStr(p3, "abc"),
			PState{input: "bc", pos: 1, value: 777, ok: true, user: map[string]interface{}{"counter": 777}})

		p4 := Bind(AnyChar, putB)
		tt.AssertEqual(parseStr(p4, "\nbc"),
			PState{input: "bc", pos: 1, value: nil, ok: true})
		tt.AssertEqual(parseStr(p4, "abc"),
			PState{input: "bc", pos: 1, value: 777, ok: true, user: map[string]interface{}{"counter": 777}})

		p5 := SeqRight(Return(nil), Bind(AnyChar, putB)).(Parser)
		tt.AssertEqual(parseStr(p5, "\nbc"),
			PState{input: "bc", pos: 1, value: nil, ok: true})
		tt.AssertEqual(parseStr(p5, "abc"),
			PState{input: "bc", pos: 1, value: 777, ok: true, user: map[string]interface{}{"counter": 777}})

		//--------------------------------------------------------------------------------
		soleIncrIfNewline := func(x interface{}) Parser {
			if x.(u8.Codepoint) == u8.Codepoint('\n') {
				return ModifyStateMapEntry("counter", func(n interface{}, _ ...interface{}) interface{} {
					return n.(int) + 1
				})
			} else {
				return Return(nil)
			}
		}

		p20 := SeqRight(init, Bind(AnyChar, soleIncrIfNewline)).(Parser)
		tt.AssertEqual(parseStr(p20, "\nbc"),
			PState{input: "bc", pos: 1, value: map[string]interface{}{"counter": 123}, ok: true, user: map[string]interface{}{"counter": 124}})
		tt.AssertEqual(parseStr(p20, "abc"),
			PState{input: "bc", pos: 1, value: nil, ok: true, user: map[string]interface{}{"counter": 123}})

		//--------------------------------------------------------------------------------
		incrIfNewline := func(x interface{}) Parser {
			if x.(u8.Codepoint) == u8.Codepoint('\n') {
				return Bind(ModifyStateMapEntry("counter", func(n interface{}, _ ...interface{}) interface{} {
					return n.(int) + 1
				}),
					func(_ interface{}) Parser { return GetStateMapEntry("counter") })
			} else {
				return GetStateMapEntry("counter")
			}
		}

		p21 := SeqRight(init, Bind(AnyChar, incrIfNewline)).(Parser)
		tt.AssertEqual(parseStr(p21, "\nbc"),
			PState{input: "bc", pos: 1, value: 124, ok: true, user: map[string]interface{}{"counter": 124}})
		tt.AssertEqual(parseStr(p21, "abc"),
			PState{input: "bc", pos: 1, value: 123, ok: true, user: map[string]interface{}{"counter": 123}})
	})
}

//================================================================================
func TestStateStack(t *testing.T) {
	ts.LogAsserts("StateStack", t, func(tt *ts.T) {
		p0 := Bind(PushStateStack(555), func(c1 interface{}) Parser {
			return Bind(PushStateStack(556), func(c2 interface{}) Parser {
				return PushStateStack(557)
			})
		})

		p2 := SeqRight(
			SeqRight(p0, Bind(AnyChar, func(x interface{}) (p Parser) {
				if u8.Sur(x.(u8.Codepoint)) == "\n" {
					return PopStateStack()
				} else {
					return Return(nil)
				}
			})),
			GetState).(Parser)

		tt.AssertEqual(parseStr(p2, "\naa"),
			PState{input: "aa", pos: 1, value: []interface{}{555, 556},
				ok: true, user: []interface{}{555, 556}})
		tt.AssertEqual(parseStr(p2, "abba"),
			PState{input: "bba", pos: 1, value: []interface{}{555, 556, 557},
				ok: true, user: []interface{}{555, 556, 557}})

		p3 := SeqRight(p0, Bind(AnyChar, func(x interface{}) (p Parser) {
			if u8.Sur(x.(u8.Codepoint)) == "\n" {
				return AlterTopStateStack(777)
			} else {
				return PeekStateStack()
			}
		})).(Parser)

		tt.AssertEqual(parseStr(p3, "\nbc"),
			PState{input: "bc", pos: 1, ok: true, user: []interface{}{555, 556, 777}})
		tt.AssertEqual(parseStr(p3, "abc"),
			PState{input: "bc", pos: 1, value: 557, ok: true, user: []interface{}{555, 556, 557}})

		p4 := Bind(AnyChar, func(x interface{}) (p Parser) {
			if u8.Sur(x.(u8.Codepoint)) == "\n" {
				return AlterTopStateStack(777)
			} else {
				return PeekStateStack()
			}
		})

		tt.AssertEqual(parseStr(p4, "\naa"),
			PState{input: "aa", pos: 1, error: "AlterTopStateStack doesn't handle zero-sized stacks."})
		tt.AssertEqual(parseStr(p4, "abc"),
			PState{input: "bc", pos: 1, ok: true})
	})
}

//================================================================================
func TestStateStruct(t *testing.T) {
	ts.LogAsserts("StateStruct", t, func(tt *ts.T) {

		type UserState struct {
			nlfound bool
			imports map[string]string
		}

		newUserState := UserState{
			nlfound: true,
			imports: map[string]string{},
		}

		tt.AssertEqual(parseStr(PutState(newUserState), ""),
			PState{ok: true, empty: true, user: UserState{nlfound: true, imports: map[string]string{}}})

		unsetNlFound := func(u UserState) UserState {
			u.nlfound = false
			m := make(map[string]string)
			for a, b := range u.imports {
				m[a] = b
			}
			u.imports = m
			return u
		}

		modifyUserStateP := func(f func(UserState) UserState) Parser {
			return Bind(GetState, func(s interface{}) Parser {
				r := f(s.(UserState))
				return SeqRight(PutState(r), Return(nil)).(Parser)
			})
		}

		tt.AssertEqual(parseStr(SeqRight(PutState(newUserState), modifyUserStateP(unsetNlFound)).(Parser), ""),
			PState{ok: true, empty: true, user: UserState{imports: map[string]string{}}})

	})
}

//================================================================================
