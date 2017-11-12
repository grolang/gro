// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

/*
contents:
func TestPosIntRange(t *testing.T){
func TestIdentity(t *testing.T){
func TestEqual(t *testing.T){
func TestLessThan(t *testing.T){
func TestPlus(t *testing.T){
func TestNegate(t *testing.T){
func TestMinus(t *testing.T){
func TestMult(t *testing.T){
func TestInvert(t *testing.T){
func TestDivide(t *testing.T){
func TestAbs(t *testing.T){
func TestInfNanNil(t *testing.T){
func TestDate(t *testing.T){
*/
package ops_test

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"testing"
	"time"

	ts "github.com/grolang/gro/assert"
	tp "github.com/grolang/gro/ops"
)

var (
	bigInt_minInt64        = *big.NewInt(0).Mul(big.NewInt(-0x80000000), big.NewInt(0x100000000)) // -0x8000_0000_0000_0000
	bigInt_minInt64plusOne = *big.NewInt(0).Add(&bigInt_minInt64, big.NewInt(1))                  // -0x7fff_ffff_ffff_ffff

	bigInt_maxInt64        = *big.NewInt(0).Sub(&bigInt_maxInt64plusOne, big.NewInt(1))          // +0x7fff_ffff_ffff_ffff
	bigInt_maxInt64plusOne = *big.NewInt(0).Mul(big.NewInt(0x80000000), big.NewInt(0x100000000)) // +0x8000_0000_0000_0000
	bigInt_maxInt64plusTwo = *big.NewInt(0).Add(&bigInt_maxInt64plusOne, big.NewInt(1))          // +0x8000_0000_0000_0001

	bigInt_maxUint64        = *big.NewInt(0).Sub(&bigInt_maxUint64plusOne, big.NewInt(1))          // +0xffff_ffff_ffff_ffff
	bigInt_maxUint64plusOne = *big.NewInt(0).Mul(big.NewInt(0x100000000), big.NewInt(0x100000000)) // +0x1_0000_0000_0000_0000
)

//================================================================================
func TestPosIntRange(t *testing.T) {
	ts.LogAsserts("PosIntRange", t, func(tt *ts.T) {

		for _, n := range []struct {
			r       tp.PosIntRange
			a, b, c interface{}
		}{
			{tp.PosIntRange{1, 3}, 1, 3, false},
			{tp.PosIntRange{1, 1}, 1, 1, false},
			{tp.PosIntRange{17, tp.BigInt(bigInt_maxUint64)}, 17, tp.BigInt(bigInt_maxUint64), false},
			{tp.PosIntRange{7, math.Inf(1)}, 7, math.Inf(1), true},
			//{tp.PosIntRange{7,  tp.Inf},                   7,  tp.Inf,                   true}, //put these infinities into PosIntRange
		} {
			tt.AssertEqual(n.r.From, n.a)
			tt.AssertEqual(n.r.To, n.b)
			tt.AssertEqual(n.r.IsToInf(), n.c)
		}

		tt.AssertEqual(tp.PosIntRange{789, math.Inf(1)}.IsToInf(), true)
	})
}

//================================================================================
func TestIdentity(t *testing.T) {
	ts.LogAsserts("Identity", t, func(tt *ts.T) {

		for _, n := range []struct{ a, b interface{} }{

			// Ints...
			{0, tp.Int(0)},

			{int(-0x80), tp.Int(-0x80)},
			{uint(0xff), tp.Int(0xff)},

			{int8(-0x80), tp.Int(-0x80)},
			{int8(-0x7f), tp.Int(-0x7f)},
			{int8(0), tp.Int(0)},
			{int8(0x7f), tp.Int(0x7f)},

			{uint8(0), tp.Int(0)},
			{uint8(0xff), tp.Int(0xff)},

			{int16(-0x8000), tp.Int(-0x8000)},
			{int16(-0x7fff), tp.Int(-0x7fff)},
			{int16(0), tp.Int(0)},
			{int16(0x7fff), tp.Int(0x7fff)},

			{uint16(0), tp.Int(0)},
			{uint16(0xffff), tp.Int(0xffff)},

			{int32(-0x80000000), tp.Int(-0x80000000)},
			{int32(-0x7fffffff), tp.Int(-0x7fffffff)},
			{int32(0), tp.Int(0)},
			{int32(0x7fffffff), tp.Int(0x7fffffff)},

			{uint32(0), tp.Int(0)},
			{uint32(0xffffffff), tp.Int(0xffffffff)},

			{int64(-0x8000000000000000), tp.Int(-0x8000000000000000)},
			{int64(-0x7fffffffffffffff), tp.Int(-0x7fffffffffffffff)},
			{int64(0), tp.Int(0)},
			{int64(0x7fffffffffffffff), tp.Int(0x7fffffffffffffff)},

			{uint64(0), tp.Int(0)},
			{uint64(0x7fffffffffffff), tp.Int(0x7fffffffffffff)},
			{uint64(0x7ffffffffffffff), tp.Int(0x7ffffffffffffff)},
			{uint64(0x7fffffffffffffff), tp.Int(0x7fffffffffffffff)},

			// BigInts...
			{uint64(0x8000000000000000), tp.BigInt(bigInt_maxInt64plusOne)},
			{uint64(0x8000000000000001), tp.BigInt(bigInt_maxInt64plusTwo)},
			{uint64(0xffffffffffffffff), tp.BigInt(bigInt_maxUint64)},

			{bigInt_minInt64plusOne, tp.BigInt(bigInt_minInt64plusOne)}, // -0x7fff_ffff_ffff_ffff
			{*big.NewInt(-0x8000000000000000), tp.NewBigInt(-0x8000000000000000)},
			{*big.NewInt(-0x7fffffffffffffff), tp.BigInt(*big.NewInt(-0x7fffffffffffffff))},
			{*big.NewInt(-1), tp.BigInt(*big.NewInt(-1))},
			{*big.NewInt(0), tp.BigInt(*big.NewInt(0))},
			{*big.NewInt(1), tp.BigInt(*big.NewInt(1))},
			{*big.NewInt(0x7fffffffffffffff), tp.BigInt(*big.NewInt(0x7fffffffffffffff))},
			{bigInt_maxInt64plusOne, tp.BigInt(bigInt_maxInt64plusOne)}, // 0x8000_0000_0000_0000
			{bigInt_maxInt64plusTwo, tp.BigInt(bigInt_maxInt64plusTwo)}, // 0x8000_0000_0000_0001
			{bigInt_maxUint64, tp.BigInt(bigInt_maxUint64)},             // 0xffff_ffff_ffff_ffff

			// BigRats...
			{*big.NewRat(-4, 16), tp.BigRat(*big.NewRat(-1, 4))},

			// Floats...
			{0.0, tp.Float(0)},
			{1.234, tp.Float(1.234)},

			{float32(0), tp.Float(0)},
			{float32(1.234), tp.Float(1.2339999675750732)},

			{float64(1.234), tp.Float(1.234)},

			{float64(1000000000000000), tp.Float(1e+15)},     //quadrillion, i.e. billiard
			{float64(1000000000000000000), tp.Float(1e+18)},  //quintillion
			{float64(10000000000000000000), tp.Float(1e+19)}, //10 quintillion
			{float64(1000000000000000000000000000000000000000000000000000000000), tp.Float(1e+57)},
			{float64(1000000000000000700000000000000000000000000000000000000000), tp.Float(1.0000000000000007e+57)}, //7 sensed
			{float64(1000000000000000070000000000000000000000000000000000000000), tp.Float(1e+57)},                  //7 ignored

			// BigFloats...
			{*big.NewFloat(1e+57), tp.NewBigFloat(1e+57)},
			{*big.NewFloat(1000000000000000700000000000000000000000000000000000000000), tp.NewBigFloat(1.0000000000000007e+57)}, //7 sensed
			{*big.NewFloat(1000000000000000070000000000000000000000000000000000000000), tp.NewBigFloat(1e+57)},                  //7 ignored

			// Complexes...
			{0i, tp.Complex(0)},
			{1.234 + 0i, tp.Complex(1.234)},
			{1.234i, tp.Complex(1.234i)},
			{1.234 + 1.234i, tp.Complex(1.234 + 1.234i)},

			{complex64(0), tp.Complex(0)},
			{complex64(0i), tp.Complex(0)},
			{complex64(1.234), tp.Complex(1.2339999675750732)},
			{complex64(1.234i), tp.Complex(1.2339999675750732i)},
			{complex64(1.234 + 1.234i), tp.Complex(1.2339999675750732 + 1.2339999675750732i)},

			{complex128(1.234), tp.Complex(1.234)},
			{complex128(1.234i), tp.Complex(1.234i)},

			// BigComplexes...
			{tp.NewBigComplex(*big.NewFloat(1e+57), *big.NewFloat(1.0)),
				tp.NewBigComplex(*big.NewFloat(1e+57), *big.NewFloat(1.0))},
		} {
			w := tp.Identity(n.a)
			tt.AssertEqual(w, n.b)
			tt.AssertEqual(fmt.Sprintf("%T", w), fmt.Sprintf("%T", n.b))
		}
	})
}

//================================================================================
func TestEqual(t *testing.T) {
	ts.LogAsserts("Equal", t, func(tt *ts.T) {

		//equal...
		for _, n := range []struct{ a, b interface{} }{
			{true, true},
			{false, false},
			{0, 0.0}, //int, float32
			{tp.Int(0), tp.Float(0)},
			{tp.Int(7), tp.NewBigInt(7)},
			{tp.NewBigInt(8), tp.NewBigRat(8, 1)},
			{tp.NewBigRat(8, 1), tp.NewBigInt(8)},
			{tp.Float(1.1), tp.NewBigRat(11, 10)},
			{tp.NewBigFloat(1.1), tp.Float(1.10)},
			{tp.Complex(1.23 + 0i), tp.Float(1.23)},
			{tp.Complex(1.23 + 4.56i), 1.23 + 4.56i},
			{tp.BigComplex{*big.NewFloat(1.23), *big.NewFloat(4.56)},
				tp.Complex(1.23 + 4.56i)},
		} {
			w := tp.IsEqual(n.a, n.b)
			tt.Assert(w)
		}

		//----------------------------------------------------------------------------
		//not equal...
		for _, n := range []struct{ a, b interface{} }{
			{true, false},
		} {
			w := !tp.IsEqual(n.a, n.b)
			tt.Assert(w)
		}
	})
}

//================================================================================
func TestLessThan(t *testing.T) {
	ts.LogAsserts("LessThan", t, func(tt *ts.T) {

		for _, n := range []struct{ a, b interface{} }{
			{-1, 0.0},
			{tp.Int(0), tp.Float(1.23)},
			{tp.Int(-7), tp.NewBigInt(7)},
			{tp.NewBigInt(-8), tp.NewBigRat(8, 1)},
			{tp.NewBigRat(8, -1), tp.NewBigInt(8)},
			{tp.Float(1.1), tp.NewBigRat(12, 10)},
			{tp.NewBigFloat(1.23), tp.NewBigRat(5, 4)},
			{float64(1.23), tp.Float(1.25)},
		} {
			w := tp.IsLessThan(n.a, n.b)
			tt.Assert(w)
		}
	})
}

//================================================================================
func TestPlus(t *testing.T) {
	ts.LogAsserts("Plus", t, func(tt *ts.T) {

		for _, n := range []struct{ a, b, c interface{} }{

			//bool...
			{false, false, false}, // + acts like || but without lazy eval
			{true, false, true},
			{false, true, true},
			{true, true, true},
			{int64(0x15a1), false, tp.Int(0x15a1)}, // false is 0; true is 1
			{int64(0x15a1), true, tp.Int(0x15a2)},
			{0x15a1 + 789i, true, tp.Complex(0x15a2 + 789i)},

			//Int...
			{int64(0x1f1f), int64(0x1010), tp.Int(0x2f2f)},                              // 2 Quints without overflow remain a Int
			{int64(0x1), uint64(0x8000000000000000), tp.BigInt(bigInt_maxInt64plusTwo)}, // 2 Quints (i.e. int64 + uint64) promoted to BigInt if oflo
			{uint64(0x8000000000000000), int64(0x1), tp.BigInt(bigInt_maxInt64plusTwo)}, // uint64 + int64 -> BigInt
			{bigInt_maxInt64plusOne, int64(0x1), tp.BigInt(bigInt_maxInt64plusTwo)},     // BigInt + Int -> BigInt
			{int64(0x1), bigInt_maxInt64plusOne, tp.BigInt(bigInt_maxInt64plusTwo)},     // Int + BigInt -> BigInt

			//BigInt...
			{tp.NewBigInt(-0x80000000), tp.NewBigInt(0x100000000), tp.NewBigInt(0x80000000)},    // BigInt + BigInt -> BigInt
			{tp.NewBigInt(-0x80000000), tp.NewBigRat(0x100000000, 1), tp.NewBigInt(0x80000000)}, // BigInt + BigRat -> BigRat, unless narrowed to BigInt
			{tp.NewBigInt(-0x80000000), 1.0, tp.Float(-0x7fffffff)},                             // BigInt + float32 -> Float
			{tp.NewBigInt(-0x80000000), 1 + 0i, tp.Complex(-0x7fffffff + 0i)},                   // BigInt + complex64 -> Complex

			//BigRat...
			{tp.NewBigRat(-0x80000000, 1), tp.NewBigRat(0x100000000, 1), tp.NewBigInt(0x80000000)}, // BigRat + BigRat -> BigRat, unless narrowed to BigInt
			{tp.NewBigRat(-0x80000000, 1), 1.0, tp.Float(-0x7fffffff)},                             // BigRat + float32 -> Float
			{tp.NewBigRat(-0x80000000, 1), 1 + 0i, tp.Complex(-0x7fffffff + 0i)},                   // BigRat + complex64 -> Complex

			//Float...
			{1.1, 0 + 1i, tp.Complex(1.1 + 1i)}, // float32 + complex64 -> Complex

			//BigFloat...
			{tp.NewBigFloat(1.1), tp.Float(1.95), tp.NewBigFloat(3.05)}, // BigFloat + Float -> BigFloat

			//Complex...
			{151 + 789i, 0 + 10i, tp.Complex(151 + 799i)}, // complex64 + complex64 -> Complex

			//BigComplex...
			{tp.BigComplex{*big.NewFloat(151.0), *big.NewFloat(789.0)},
				tp.BigComplex{*big.NewFloat(0.0), *big.NewFloat(10.0)},
				tp.BigComplex{*big.NewFloat(151.0), *big.NewFloat(799.0)}}, // BigComplex + BigComplex -> BigComplex

			{tp.NewBigComplex(151.0, 789.0),
				0.0 + 10.0i,
				tp.BigComplex{*big.NewFloat(151.0), *big.NewFloat(799.0)}}, // BigComplex + Complex -> BigComplex

		} {
			w := tp.Plus(n.a, n.b)
			wf, isBigFloat := w.(tp.BigFloat)
			wc, isBigComplex := w.(tp.BigComplex)
			tt.Assert(isBigFloat && wf.IsEqual(n.c) ||
				isBigComplex && wc.IsEqual(n.c) ||
				reflect.DeepEqual(w, n.c))
			tt.AssertEqual(fmt.Sprintf("%T", w), fmt.Sprintf("%T", n.c))
		}

	})
}

//================================================================================
func TestNegate(t *testing.T) {
	ts.LogAsserts("Negate", t, func(tt *ts.T) {

		for _, n := range []struct{ a, b interface{} }{
			{true, false},
			{false, true},
			{tp.Int(4), tp.Int(-4)},
			{tp.Int(0), tp.Int(0)},
			{tp.NewBigInt(0), tp.NewBigInt(0)},
			{tp.NewBigInt(-7), tp.NewBigInt(7)},
			{tp.NewBigRat(0, 1), tp.NewBigRat(0, 1)},
			{tp.NewBigRat(1, 7), tp.NewBigRat(1, -7)},
			{tp.Float(0), tp.Float(0)},
			{tp.Float(-7.0), tp.Float(7.0)},
			{tp.NewBigFloat(1.23), tp.NewBigFloat(-1.23)},
			{tp.Complex(0 + 0i), tp.Complex(0 + 0i)},
			{tp.Complex(2 + 3i), tp.Complex(-2 - 3i)},
			{tp.NewBigComplex(151.0, 789.0), tp.NewBigComplex(-151.0, -789.0)},
		} {
			tt.AssertEqual(tp.Negate(n.a), n.b)
		}

	})
}

//================================================================================
func TestMinus(t *testing.T) {
	ts.LogAsserts("Minus", t, func(tt *ts.T) {

		for _, n := range []struct{ a, b, c interface{} }{
			//bool...
			{false, false, true}, // a-b is short for a+(-b), which acts like (a || !b) but without lazy eval
			{true, false, true},
			{false, true, false},
			{true, true, true},

			//TODO ???: eliminate this from the code...
			{int64(0x15a1), false, tp.Int(0x15a2)},
			{int64(0x15a1), true, tp.Int(0x15a1)},
			{0x15a1 + 789i, true, tp.Complex(0x15a1 + 789i)},

			//Int...
			{uint64(0x1), int64(0xF), tp.Int(-0xE)},                                      // uint64 - int64 -> Int
			{int64(0x1f1f), int64(0x1010), tp.Int(0x0f0f)},                               // 2 Quints without overflow remain a Int
			{int64(0x1), uint64(0x8000000000000000), tp.BigInt(bigInt_minInt64plusOne)},  // 2 Quints (i.e. int64 - uint64) promoted to BigInt when oflo
			{uint64(0x1), int64(-0x7FFFFFFFFFFFFFFF), tp.BigInt(bigInt_maxInt64plusOne)}, // uint64 - int64 -> BigInt when overflow
			{bigInt_maxInt64plusOne, int64(0x1), tp.BigInt(bigInt_maxInt64)},             // BigInt - Int -> BigInt
			{int64(0x1), bigInt_maxInt64plusOne, tp.BigInt(bigInt_minInt64plusOne)},      // Int  - BigInt -> BigInt

			//BigInt...
			{tp.NewBigInt(-0x80000000), tp.NewBigInt(0x100000000), tp.NewBigInt(-0x180000000)},    // BigInt - BigInt -> BigInt
			{tp.NewBigInt(-0x80000000), tp.NewBigRat(0x100000000, 1), tp.NewBigInt(-0x180000000)}, // BigInt - BigRat -> BigRat, unless narrowed to BigInt
			{tp.NewBigInt(-0x80000000), 1.0, tp.Float(-0x80000001)},                               // BigInt - float32 -> Float
			{tp.NewBigInt(-0x80000000), 1 + 0i, tp.Complex(-0x80000001 + 0i)},                     // BigInt - complex64 -> Complex

			//BigRat...
			{tp.NewBigRat(-0x80000000, 1), tp.NewBigRat(0x100000000, 1), tp.NewBigInt(-0x180000000)}, // BigRat - BigRat -> BigRat, unless narrowed to BigInt
			{tp.NewBigRat(-0x80000000, 1), 1.0, tp.Float(-0x80000001)},                               // BigRat - float32 -> Float
			{tp.NewBigRat(-0x80000000, 1), 1 + 0i, tp.Complex(-0x80000001 + 0i)},                     // BigRat - complex64 -> Complex

			//Float...
			{1.1, 0 + 1i, tp.Complex(1.1 - 1i)}, // float32 - complex64 -> Complex

			//BigFloat...
			//{tp.NewBigFloat(3.1),        tp.Float(1.95),               tp.NewBigFloat(1.15)},    // BigFloat - Float -> BigFloat

			//Complex...
			{151 + 789i, 0 + 10i, tp.Complex(151 + 779i)}, // complex64 - complex64 -> Complex

		} {
			w := tp.Minus(n.a, n.b)

			tt.AssertEqual(w, n.c)
			wf, isBigFloat := w.(tp.BigFloat)
			wc, isBigComplex := w.(tp.BigComplex)
			tt.Assert(isBigFloat && wf.IsEqual(n.c) ||
				isBigComplex && wc.IsEqual(n.c) ||
				reflect.DeepEqual(w, n.c))
			tt.AssertEqual(fmt.Sprintf("%T", w), fmt.Sprintf("%T", n.c))

		}
	})
}

//================================================================================
func TestMult(t *testing.T) {
	ts.LogAsserts("Mult", t, func(tt *ts.T) {

		for _, n := range []struct{ a, b, c interface{} }{
			//bool...
			{false, false, false}, // * acts like && but without lazy eval
			{true, false, false},
			{false, true, false},
			{true, true, true},
			{int64(0x15a1), true, tp.Int(0x15a1)},
			{0x15a1 + 789i, true, tp.Complex(0x15a1 + 789i)},

			{int64(0x15a1), false, tp.Int(0)}, // false is 0; true is 1
			{int64(0x15a1), true, tp.Int(0x15a1)},
			{0x15a1 + 789i, false, tp.Complex(0 + 0i)},
			{0x15a1 + 789i, true, tp.Complex(0x15a1 + 789i)},

			//Int...
			{int64(0x1f1f), int64(0x1010), tp.Int(0x1f3e1f0)},                           // 2 Quints without overflow remain a Int
			{int64(0x1), uint64(0x8000000000000000), tp.BigInt(bigInt_maxInt64plusOne)}, // 2 Quints (i.e. int64 * uint64) promoted to BigInt when overflow
			{uint64(0x8000000000000000), int64(0x1), tp.BigInt(bigInt_maxInt64plusOne)}, // uint64 * int64 -> BigInt
			{bigInt_maxInt64plusOne, int64(0x1), tp.BigInt(bigInt_maxInt64plusOne)},     // BigInt * Int -> BigInt
			{int64(0x1), bigInt_maxInt64plusOne, tp.BigInt(bigInt_maxInt64plusOne)},     // Int * BigInt -> BigInt

			//BigInt...
			{tp.NewBigInt(0x80000000), tp.NewBigInt(1), tp.NewBigInt(0x80000000)},    // BigInt * BigInt -> BigInt
			{tp.NewBigInt(0x80000000), tp.NewBigRat(1, 1), tp.NewBigInt(0x80000000)}, // BigInt * BigRat -> BigRat, unless narrowed to BigInt
			{tp.NewBigInt(-0x80000000), 1.0, tp.Float(-0x80000000)},                  // BigInt * float32 -> Float
			{tp.NewBigInt(-0x80000000), 1 + 0i, tp.Complex(-0x80000000 + 0i)},        // BigInt * complex64 -> Complex

			//BigRat...
			{tp.NewBigRat(-0x80000000, 1), tp.NewBigRat(1, 1), tp.NewBigInt(-0x80000000)}, // BigRat * BigRat -> BigRat, unless narrowed to BigInt
			{tp.NewBigRat(-0x80000000, 1), 1.0, tp.Float(-0x80000000)},                    // BigRat * float32 -> Float
			{tp.NewBigRat(-0x80000000, 1), 1 + 0i, tp.Complex(-0x80000000 + 0i)},          // BigRat * complex64 -> Complex

			//Float...
			{1.1, 0 + 1i, tp.Complex(1.1i)}, // float32 * complex64 -> Complex

			//Complex...
			{151 + 789i, 0 + 1i, tp.Complex(-789 + 151i)}, // complex64 * complex64 -> Complex

		} {
			w := tp.Mult(n.a, n.b)
			tt.AssertEqual(w, n.c)
			tt.AssertEqual(fmt.Sprintf("%T", w), fmt.Sprintf("%T", n.c))
		}
	})
}

//================================================================================
func TestInvert(t *testing.T) {
	ts.LogAsserts("Invert", t, func(tt *ts.T) {

		for _, n := range []struct{ a, b interface{} }{
			{true, true},
			{tp.Int(4), tp.NewBigRat(1, 4)},
			{tp.Float(4), tp.Float(0.25)},
			{tp.Complex(4i), tp.Complex(-0.25i)},

			{tp.NewBigFloat(-5.0), tp.NewBigFloat(-0.2)},

			{tp.NewBigComplex(-5.0, 0.25), tp.NewBigComplex(-0.2, 4.0)},
		} {
			wf, isBigFloat := tp.Invert(n.a).(tp.BigFloat)
			wc, isBigComplex := tp.Invert(n.a).(tp.BigComplex)
			tt.Assert(isBigFloat && wf.IsEqual(n.b) ||
				isBigComplex && wc.IsEqual(n.b) ||
				reflect.DeepEqual(tp.Invert(n.a), n.b))
		}

	})
}

//================================================================================
func TestDivide(t *testing.T) {
	ts.LogAsserts("Divide", t, func(tt *ts.T) {

		for _, n := range []struct{ a, b, c interface{} }{

			//bool...
			//{false,                      false,                      false}, // division by false invalid
			//{true,                       false,                      false}, // division by false invalid
			{false, true, false},
			{true, true, true},
			{int64(0x15a1), true, tp.Int(0x15a1)},
			{0x15a1 + 789i, true, tp.Complex(0x15a1 + 789i)},

			//{int64(0x15a1),              false,                      tp.Int(0)}, // division by false invalid
			{int64(0x15a1), true, tp.Int(0x15a1)},
			//{0x15a1+789i,                false,                      tp.Complex(0+0i)}, // division by false invalid
			{0x15a1 + 789i, true, tp.Complex(0x15a1 + 789i)},

			//Int...
			{int64(123), int64(3), tp.NewBigInt(41)},
			{uint64(0x8000000000000000), int64(0x1), tp.BigInt(bigInt_maxInt64plusOne)}, // uint64 / int64  -> BigInt
			{int64(0x7fffffffffffffff), uint64(0x1), tp.BigInt(bigInt_maxInt64)},        // int64  / uint64 -> BigInt
			{bigInt_maxInt64plusOne, int64(0x1), tp.BigInt(bigInt_maxInt64plusOne)},     // BigInt / Int  -> BigInt

			//BigInt...
			{tp.NewBigInt(0x80000000), tp.NewBigInt(1), tp.NewBigInt(0x80000000)},    // BigInt / BigInt -> BigInt
			{tp.NewBigInt(0x80000000), tp.NewBigRat(1, 1), tp.NewBigInt(0x80000000)}, // BigInt / BigRat -> BigInt if no integral
			{tp.NewBigInt(-0x80000000), 1.0, tp.Float(-0x80000000)},                  // BigInt / float32 -> Float
			{tp.NewBigInt(-0x80000000), 1 + 0i, tp.Complex(-0x80000000 + 0i)},        // BigInt / complex64 -> Complex

			//BigRat...
			{tp.NewBigRat(-0x80000000, 1), tp.NewBigRat(1, 1), tp.NewBigInt(-0x80000000)}, // BigRat / BigRat -> BigInt if no integral
			{tp.NewBigRat(-0x80000000, 1), 1.0, tp.Float(-0x80000000)},                    // BigRat / float32 -> Float
			{tp.NewBigRat(-0x80000000, 1), 1 + 0i, tp.Complex(-0x80000000 + 0i)},          // BigRat / complex64 -> Complex

			//Float...
			{1.1, 0 + 1i, tp.Complex(-1.1i)}, // float32 * complex64 -> Complex

			//Complex...
			{151 + 789i, 0 + 1i, tp.Complex(789 - 151i)}, // complex64 * complex64 -> Complex

			{nil, nil, tp.NaN},
		} {
			w := tp.Divide(n.a, n.b)
			wc, isComplex := w.(tp.Complex)
			tt.Assert(isComplex && wc.IsEqual(n.c) ||
				reflect.DeepEqual(w, n.c))
			tt.AssertEqual(fmt.Sprintf("%T", w), fmt.Sprintf("%T", n.c))
		}

	})
}

//================================================================================
func TestAbs(t *testing.T) {
	ts.LogAsserts("Abs", t, func(tt *ts.T) {

		for _, n := range []struct{ a, b interface{} }{} {
			tt.AssertEqual(tp.Abs(n.a), n.b)
		}

		for _, n := range []struct{ a, b interface{} }{
			{true, true},
			{tp.Int(4), tp.Int(4)},
			{tp.Int(-4), tp.Int(4)},

			{tp.Float(4.0), tp.Float(4.0)},
			{tp.Float(-4), tp.Float(4.0)},

			{tp.NewBigFloat(-5.0), tp.NewBigFloat(5.0)},

			{tp.Complex(4i), tp.Float(4)},

			{tp.NewBigComplex(-5.0, 0.0), tp.Float(5.0)},
		} {
			wf, isBigFloat := tp.Abs(n.a).(tp.BigFloat)
			wc, isBigComplex := tp.Abs(n.a).(tp.BigComplex)
			tt.Assert(isBigFloat && wf.IsEqual(n.b) ||
				isBigComplex && wc.IsEqual(n.b) ||
				reflect.DeepEqual(tp.Abs(n.a), n.b))
		}

	})
}

//================================================================================
func TestInfNanNil(t *testing.T) {
	ts.LogAsserts("InfNanNil", t, func(tt *ts.T) {

		//unary ops...
		for _, n := range []struct{ a, fa, b interface{} }{

			{nil, tp.Identity, nil},
			{tp.Inf, tp.Identity, tp.Inf},
			{tp.NaN, tp.Identity, tp.NaN},

			{nil, tp.Negate, nil},
			{tp.Inf, tp.Negate, tp.Inf},
			{tp.NaN, tp.Negate, tp.NaN},

			{nil, tp.Invert, tp.NaN},

			{tp.Int(0), tp.Invert, tp.Inf}, // z / 0 = Inf, for z<>0
			{tp.NewBigInt(0), tp.Invert, tp.Inf},
			{tp.NewBigRat(0, 1), tp.Invert, tp.Inf},
			{0.0, tp.Invert, tp.Inf},
			{-0.0, tp.Invert, tp.Inf},
			{tp.Float(0), tp.Invert, tp.Inf},
			{tp.NewBigFloat(0.0), tp.Invert, tp.Inf},
			{tp.Complex(0 + 0i), tp.Invert, tp.Inf},
			{tp.NewBigComplex(0.0, 0.0),
				tp.Invert, tp.Inf},

			{tp.Inf, tp.Invert, tp.Complex(0 + 0i)}, // z / Inf = 0, for z<>0

			{tp.NaN, tp.Invert, tp.NaN}, //TODO ???: should be nil

		} {
			fn := n.fa.(func(a interface{}) interface{})
			w := fn(n.a)
			wc, isComplex := w.(tp.Complex)
			tt.Assert(isComplex && wc.IsEqual(n.b) ||
				reflect.DeepEqual(w, n.b))
			tt.AssertEqual(fmt.Sprintf("%T", w), fmt.Sprintf("%T", n.b))
		}

		//----------------------------------------------------------------------------
		//binary ops...
		for _, n := range []struct{ a, b, fab, c interface{} }{

			//TODO: IsEqual, IsLessThan

			{nil, nil, tp.Plus, nil},
			{nil, 123, tp.Plus, tp.Int(123)}, // nil is 0

			{1, tp.Inf, tp.Plus, tp.Inf}, // z + Inf = Inf, for z<>Inf
			{0, tp.Inf, tp.Plus, tp.Inf},
			{tp.Inf, tp.Inf, tp.Plus, tp.Inf}, //TODO: should be tp.NaN // Inf + Inf = NaN

			/*  math.IsInf(math.Inf(1) + math.Inf(1), 1) &&
			    math.IsInf(math.Inf(-1) + math.Inf(-1), -1) &&
			    math.IsNaN(math.Inf(1) + math.Inf(-1)) &&
			    math.IsNaN(math.Inf(-1) + math.Inf(1)) && */

			{tp.Inf, tp.Inf, tp.Minus, tp.Inf}, //TODO: should be tp.NaN // Inf - Inf = NaN
			{nil, nil, tp.Minus, nil},
			{nil, 123, tp.Minus, tp.Int(-123)}, // nil is 0

			/*  math.IsInf(math.Inf(1) - math.Inf(-1), 1) &&
			    math.IsInf(math.Inf(-1) - math.Inf(1), -1) &&
			    math.IsNaN(math.Inf(1) - math.Inf(1)) &&
			    math.IsNaN(math.Inf(-1) - math.Inf(-1)) && */

			{nil, nil, tp.Mult, nil}, // nil is 0
			{nil, 123, tp.Mult, tp.Int(0)},
			{0, tp.Inf, tp.Mult, tp.NaN}, // 0 * Inf = NaN
			{1, tp.Inf, tp.Mult, tp.NaN}, //TODO: should be Inf // z * Inf = Inf, for z<>0 and z<>Inf
			{-1, tp.Inf, tp.Mult, tp.NaN},
			{tp.Inf, tp.Inf, tp.Mult, tp.Inf}, // Inf * Inf = Inf

			{nil, nil, tp.Divide, tp.NaN},
			{nil, 123, tp.Divide, tp.NewBigInt(0)},

			{0, 0, tp.Divide, tp.NaN},      // 0 / 0 = NaN
			{1, 0, tp.Divide, tp.NaN},      //TODO: should be Inf // z / 0 = Inf, for z<>0
			{-1, 0, tp.Divide, tp.NaN},     //TODO: should also be Inf
			{tp.Inf, 0, tp.Divide, tp.Inf}, // Inf / 0 = Inf

			{0, tp.Inf, tp.Divide, tp.Complex(0)},      // 0 / Inf = 0
			{1, tp.Inf, tp.Divide, tp.Complex(0 + 0i)}, // z / Inf = 0, for z<>0
			{-1, tp.Inf, tp.Divide, tp.Complex(0 + 0i)},
			{tp.Inf, tp.Inf, tp.Divide, tp.NaN}, // Inf / Inf = NaN

		} {
			fn := n.fab.(func(a, b interface{}) interface{})
			w := fn(n.a, n.b)
			wc, isComplex := w.(tp.Complex)
			tt.Assert(isComplex && wc.IsEqual(n.c) ||
				reflect.DeepEqual(w, n.c))
			tt.AssertEqual(fmt.Sprintf("%T", w), fmt.Sprintf("%T", n.c))
		}

	})
}

//================================================================================
func TestDate(t *testing.T) {
	ts.LogAsserts("Date", t, func(tt *ts.T) {

		//unary ops...
		for _, n := range []struct{ a, fa, b interface{} }{

			{time.Date(2009, time.November, 10, 23, 22, 21, 20, time.UTC), tp.Identity, time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC)},
		} {
			fn := n.fa.(func(a interface{}) interface{})
			w := fn(n.a)
			tt.AssertEqual(w, n.b)
			tt.AssertEqual(fmt.Sprintf("%T", w), fmt.Sprintf("%T", n.b))
		}

		//----------------------------------------------------------------------------
		//binary ops...
		for _, n := range []struct{ a, b, fab, c interface{} }{

			{time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC), 1, tp.Plus,
				time.Date(2009, time.November, 11, 0, 0, 0, 0, time.UTC)}, //gro: 2009.11.10 + 1

			{time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC), 3, tp.Minus,
				time.Date(2009, time.November, 7, 0, 0, 0, 0, time.UTC)}, //gro: 2009.11.10 - 3

			{1, time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC), tp.Plus,
				time.Date(2009, time.November, 11, 0, 0, 0, 0, time.UTC)}, //gro: 1 + 2009.11.10

		} {
			fn := n.fab.(func(a, b interface{}) interface{})
			w := fn(n.a, n.b)
			tt.AssertEqual(w, n.c)
			tt.AssertEqual(fmt.Sprintf("%T", w), fmt.Sprintf("%T", n.c))
		}

		//----------------------------------------------------------------------------
		//predicate binary ops...
		for _, n := range []struct{ a, b, fab, c interface{} }{

			{time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC),
				time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC), tp.IsEqual, true}, //gro: 2009.11.10 == 2009.11.10
			{time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC),
				time.Date(2009, time.November, 11, 0, 0, 0, 0, time.UTC), tp.IsEqual, false}, //gro: 2009.11.10 == 2009.11.11

			{time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC),
				time.Date(2009, time.November, 9, 0, 0, 0, 0, time.UTC), tp.IsLessThan, false}, //gro: 2009.11.10 < 2009.11.9
			{time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC),
				time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC), tp.IsLessThan, false}, //gro: 2009.11.10 < 2009.11.10
			{time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC),
				time.Date(2009, time.November, 11, 0, 0, 0, 0, time.UTC), tp.IsLessThan, true}, //gro: 2009.11.10 < 2009.11.11

		} {
			fn := n.fab.(func(a, b interface{}) bool)
			w := fn(n.a, n.b)
			tt.AssertEqual(w, n.c)
			tt.AssertEqual(fmt.Sprintf("%T", w), fmt.Sprintf("%T", n.c))
		}

	})
}

//================================================================================
