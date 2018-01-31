// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package ops_test

import (
	"fmt"
	"math/big"
	"math/cmplx"
	"reflect"
	"testing"
	"time"

	"github.com/grolang/gro/ops"
)

var (
	// -0x8000_0000_0000_0000
	minInt64 = *big.NewInt(0).Mul(big.NewInt(-0x80000000), big.NewInt(0x100000000))

	// -0x7fff_ffff_ffff_ffff
	minInt64plusOne = *big.NewInt(0).Add(&minInt64, big.NewInt(1))

	// +0x7fff_ffff_ffff_ffff
	maxInt64 = *big.NewInt(0).Sub(&maxInt64plusOne, big.NewInt(1))

	// +0x8000_0000_0000_0000
	maxInt64plusOne = *big.NewInt(0).Mul(big.NewInt(0x80000000), big.NewInt(0x100000000))

	// +0x8000_0000_0000_0001
	maxInt64plusTwo = *big.NewInt(0).Add(&maxInt64plusOne, big.NewInt(1))

	// +0xffff_ffff_ffff_ffff
	maxUint64 = *big.NewInt(0).Sub(&maxUint64plusOne, big.NewInt(1))

	// +0x1_0000_0000_0000_0000
	maxUint64plusOne = *big.NewInt(0).Mul(big.NewInt(0x100000000), big.NewInt(0x100000000))
)

func newBigInt(x int) ops.BigInt {
	return ops.BigInt(*big.NewInt(int64(x)))
}

func newBigRat(x, y int) ops.BigRat {
	return ops.BigRat(*big.NewRat(int64(x), int64(y)))
}

func newBigFloat(x float64) ops.BigFloat {
	return ops.BigFloat(*big.NewFloat(x))
}

func newBigComplex(x, y float64) ops.BigComplex {
	return ops.BigComplex{*big.NewFloat(x), *big.NewFloat(y)}
}

func bigFloatIsEqual(this ops.BigFloat, that ops.BigFloat) bool {
	x := big.Float(this)
	y := big.Float(that)
	return x.Cmp(&y) == 0
}

func complexIsEqual(this ops.Complex, that interface{}) bool {
	switch that.(type) {
	case ops.Complex:
		x := that.(ops.Complex)
		if cmplx.IsNaN(complex128(x)) && cmplx.IsNaN(complex128(this)) {
			return true
		}
		return complex128(x) == complex128(this)

	default:
		panic("complexIsEqual: only Complexes are valid 2nd args")
	}
}

func bigComplexIsEqual(this ops.BigComplex, that interface{}) bool {
	switch that.(type) {

	case ops.BigFloat:
		x := big.Float(that.(ops.BigFloat))
		y := *big.NewFloat(0)
		return this.Re.Cmp(&x) == 0 && this.Im.Cmp(&y) == 0

	case ops.Complex:
		x := that.(ops.Complex)
		xr := *big.NewFloat(real(complex128(x)))
		xi := *big.NewFloat(imag(complex128(x)))
		return this.Re.Cmp(&xr) == 0 && this.Im.Cmp(&xi) == 0

	case ops.BigComplex:
		x := that.(ops.BigComplex)
		xr := x.Re
		xi := x.Im
		return this.Re.Cmp(&xr) == 0 && this.Im.Cmp(&xi) == 0

	default:
		panic("bigComplexIsEqual: only BigFloats, Complexes and BigComplexes are 2nd valid args")
	}
}

func checkEqual(t *testing.T, w, z interface{}) {
	if !reflect.DeepEqual(w, z) {
		t.Errorf("Failure in values: %v not equal to %v", w, z)
	}
	if fmt.Sprintf("%T", w) != fmt.Sprintf("%T", z) {
		t.Errorf("Failure in types: %T not the same as %T", w, z)
	}
}

//================================================================================

//TODO: TEST ToBool

func TestIdentity(t *testing.T) {
	for _, n := range []struct{ a, b interface{} }{

		// Ints...
		{0, ops.Int(0)},

		{int(-0x80), ops.Int(-0x80)},
		{uint(0xff), ops.Int(0xff)},

		{int8(-0x80), ops.Int(-0x80)},
		{int8(-0x7f), ops.Int(-0x7f)},
		{int8(0), ops.Int(0)},
		{int8(0x7f), ops.Int(0x7f)},

		{uint8(0), ops.Int(0)},
		{uint8(0xff), ops.Int(0xff)},

		{int16(-0x8000), ops.Int(-0x8000)},
		{int16(-0x7fff), ops.Int(-0x7fff)},
		{int16(0), ops.Int(0)},
		{int16(0x7fff), ops.Int(0x7fff)},

		{uint16(0), ops.Int(0)},
		{uint16(0xffff), ops.Int(0xffff)},

		{int32(-0x80000000), ops.Int(-0x80000000)},
		{int32(-0x7fffffff), ops.Int(-0x7fffffff)},
		{int32(0), ops.Int(0)},
		{int32(0x7fffffff), ops.Int(0x7fffffff)},

		{uint32(0), ops.Int(0)},
		{uint32(0xffffffff), ops.Int(0xffffffff)},

		{int64(-0x8000000000000000), ops.Int(-0x8000000000000000)},
		{int64(-0x7fffffffffffffff), ops.Int(-0x7fffffffffffffff)},
		{int64(0), ops.Int(0)},
		{int64(0x7fffffffffffffff), ops.Int(0x7fffffffffffffff)},

		{uint64(0), ops.Int(0)},
		{uint64(0x7fffffffffffff), ops.Int(0x7fffffffffffff)},
		{uint64(0x7ffffffffffffff), ops.Int(0x7ffffffffffffff)},
		{uint64(0x7fffffffffffffff), ops.Int(0x7fffffffffffffff)},

		// BigInts...
		{uint64(0x8000000000000000), ops.BigInt(maxInt64plusOne)},
		{uint64(0x8000000000000001), ops.BigInt(maxInt64plusTwo)},
		{uint64(0xffffffffffffffff), ops.BigInt(maxUint64)},

		{minInt64plusOne, ops.BigInt(minInt64plusOne)}, // -0x7fff_ffff_ffff_ffff
		{*big.NewInt(-0x8000000000000000), newBigInt(-0x8000000000000000)},
		{*big.NewInt(-0x7fffffffffffffff), ops.BigInt(*big.NewInt(-0x7fffffffffffffff))},
		{*big.NewInt(-1), ops.BigInt(*big.NewInt(-1))},
		{*big.NewInt(0), ops.BigInt(*big.NewInt(0))},
		{*big.NewInt(1), ops.BigInt(*big.NewInt(1))},
		{*big.NewInt(0x7fffffffffffffff), ops.BigInt(*big.NewInt(0x7fffffffffffffff))},
		{maxInt64plusOne, ops.BigInt(maxInt64plusOne)}, // 0x8000_0000_0000_0000
		{maxInt64plusTwo, ops.BigInt(maxInt64plusTwo)}, // 0x8000_0000_0000_0001
		{maxUint64, ops.BigInt(maxUint64)},             // 0xffff_ffff_ffff_ffff

		// BigRats...
		{*big.NewRat(-4, 16), ops.BigRat(*big.NewRat(-1, 4))},

		// Floats...
		{0.0, ops.Float(0)},
		{1.234, ops.Float(1.234)},

		{float32(0), ops.Float(0)},
		{float32(1.234), ops.Float(1.2339999675750732)},

		{float64(1.234), ops.Float(1.234)},

		{float64(1000000000000000), ops.Float(1e+15)},     //quadrillion
		{float64(1000000000000000000), ops.Float(1e+18)},  //quintillion
		{float64(10000000000000000000), ops.Float(1e+19)}, //10 quintillion
		{float64(1000000000000000000000000000000000000000000000000000000000),
			ops.Float(1e+57)},
		{float64(1000000000000000700000000000000000000000000000000000000000),
			ops.Float(1.0000000000000007e+57)}, //7 sensed
		{float64(1000000000000000070000000000000000000000000000000000000000),
			ops.Float(1e+57)}, //7 ignored

		// BigFloats...
		{*big.NewFloat(1e+57), newBigFloat(1e+57)},
		{*big.NewFloat(1000000000000000700000000000000000000000000000000000000000),
			newBigFloat(1.0000000000000007e+57)}, //7 sensed
		{*big.NewFloat(1000000000000000070000000000000000000000000000000000000000),
			newBigFloat(1e+57)}, //7 ignored

		// Complexes...
		{0i, ops.Complex(0)},
		{1.234 + 0i, ops.Complex(1.234)},
		{1.234i, ops.Complex(1.234i)},
		{1.234 + 1.234i, ops.Complex(1.234 + 1.234i)},

		{complex64(0), ops.Complex(0)},
		{complex64(0i), ops.Complex(0)},
		{complex64(1.234), ops.Complex(1.2339999675750732)},
		{complex64(1.234i), ops.Complex(1.2339999675750732i)},
		{complex64(1.234 + 1.234i), ops.Complex(1.2339999675750732 + 1.2339999675750732i)},

		{complex128(1.234), ops.Complex(1.234)},
		{complex128(1.234i), ops.Complex(1.234i)},

		// BigComplexes...
		{newBigComplex(1e+57, 1.0), newBigComplex(1e+57, 1.0)},
	} {
		w := ops.Identity(n.a)
		checkEqual(t, w, n.b)
	}
	//})
}

//================================================================================
func TestEqual(t *testing.T) {
	//equal...
	for _, n := range []struct{ a, b interface{} }{
		{true, true},
		{false, false},
		{0, 0.0}, //int, float32
		{ops.Int(0), ops.Float(0)},
		{ops.Int(7), newBigInt(7)},
		{newBigInt(8), newBigRat(8, 1)},
		{newBigRat(8, 1), newBigInt(8)},
		{ops.Float(1.1), newBigRat(11, 10)},
		{newBigFloat(1.1), ops.Float(1.10)},
		{ops.Complex(1.23 + 0i), ops.Float(1.23)},
		{ops.Complex(1.23 + 4.56i), 1.23 + 4.56i},
		{ops.BigComplex{*big.NewFloat(1.23), *big.NewFloat(4.56)},
			ops.Complex(1.23 + 4.56i)},
	} {
		if !ops.IsEqual(n.a, n.b) {
			t.Errorf("Failure: %v not equal to %v", n.a, n.b)
		}
	}

	//----------------------------------------------------------------------------
	//not equal...
	for _, n := range []struct{ a, b interface{} }{
		{true, false},
	} {
		if ops.IsEqual(n.a, n.b) {
			t.Errorf("Failure: %v equal to %v, but shouldn't be", n.a, n.b)
		}
	}
}

//================================================================================
func TestLessThan(t *testing.T) {
	for _, n := range []struct{ a, b interface{} }{
		{-1, 0.0},
		{ops.Int(0), ops.Float(1.23)},
		{ops.Int(-7), newBigInt(7)},
		{newBigInt(-8), newBigRat(8, 1)},
		{newBigRat(8, -1), newBigInt(8)},
		{ops.Float(1.1), newBigRat(12, 10)},
		{newBigFloat(1.23), newBigRat(5, 4)},
		{float64(1.23), ops.Float(1.25)},
	} {
		if !ops.IsLessThan(n.a, n.b) {
			t.Errorf("Failure: %v not less than %v, but should be", n.a, n.b)
		}
	}
}

//TODO: TEST IsNotEqual

//TODO: TEST IsLessOrEqual

//TODO: TEST IsGreaterThan

//TODO: TEST IsGreaterOrEqual

//================================================================================
func TestPlus(t *testing.T) {
	for _, n := range []struct{ a, b, c interface{} }{

		//bool...
		{false, false, false}, // + acts like || but without lazy eval
		{true, false, true},
		{false, true, true},
		{true, true, true},
		{int64(0x15a1), false, ops.Int(0x15a1)}, // false is 0; true is 1
		{int64(0x15a1), true, ops.Int(0x15a2)},
		{0x15a1 + 789i, true, ops.Complex(0x15a2 + 789i)},

		//Int...
		{int64(0x1f1f), int64(0x1010),
			ops.Int(0x2f2f)}, // 2 Ints without overflow remain a Int
		{int64(0x1), uint64(0x8000000000000000),
			ops.BigInt(maxInt64plusTwo)}, // 2 Ints (i.e. int64 + uint64) promoted to BigInt if oflo
		{uint64(0x8000000000000000), int64(0x1),
			ops.BigInt(maxInt64plusTwo)}, // uint64 + int64 -> BigInt
		{maxInt64plusOne, int64(0x1),
			ops.BigInt(maxInt64plusTwo)}, // BigInt + Int -> BigInt
		{int64(0x1), maxInt64plusOne,
			ops.BigInt(maxInt64plusTwo)}, // Int + BigInt -> BigInt

		//BigInt...
		{newBigInt(-0x80000000), newBigInt(0x100000000),
			newBigInt(0x80000000)}, // BigInt + BigInt -> BigInt
		{newBigInt(-0x80000000), newBigRat(0x100000000, 1),
			newBigInt(0x80000000)}, // BigInt + BigRat -> BigRat, unless narrowed to BigInt
		{newBigInt(-0x80000000), 1.0,
			ops.Float(-0x7fffffff)}, // BigInt + float32 -> Float
		{newBigInt(-0x80000000), 1 + 0i,
			ops.Complex(-0x7fffffff + 0i)}, // BigInt + complex64 -> Complex

		//BigRat...
		{newBigRat(-0x80000000, 1), newBigRat(0x100000000, 1),
			newBigInt(0x80000000)}, // BigRat + BigRat -> BigRat, unless narrowed to BigInt
		{newBigRat(-0x80000000, 1), 1.0,
			ops.Float(-0x7fffffff)}, // BigRat + float32 -> Float
		{newBigRat(-0x80000000, 1), 1 + 0i,
			ops.Complex(-0x7fffffff + 0i)}, // BigRat + complex64 -> Complex

		//Float...
		{1.1, 0 + 1i, ops.Complex(1.1 + 1i)}, // float32 + complex64 -> Complex

		//BigFloat...
		{newBigFloat(1.1), ops.Float(1.95), newBigFloat(3.05)}, // BigFloat + Float -> BigFloat

		//Complex...
		{151 + 789i, 0 + 10i, ops.Complex(151 + 799i)}, // complex64 + complex64 -> Complex

		//BigComplex...
		{ops.BigComplex{*big.NewFloat(151.0), *big.NewFloat(789.0)},
			ops.BigComplex{*big.NewFloat(0.0), *big.NewFloat(10.0)},
			ops.BigComplex{*big.NewFloat(151.0), *big.NewFloat(799.0)}}, // BigComplex + BigComplex -> BigComplex

		{newBigComplex(151.0, 789.0),
			0.0 + 10.0i,
			ops.BigComplex{*big.NewFloat(151.0), *big.NewFloat(799.0)}}, // BigComplex + Complex -> BigComplex

	} {
		w := ops.Plus(n.a, n.b)
		wf, wIsBigFloat := w.(ops.BigFloat)
		ncf, ncIsBigFloat := n.c.(ops.BigFloat)
		wc, isBigComplex := w.(ops.BigComplex)
		oneOf := wIsBigFloat && ncIsBigFloat && bigFloatIsEqual(wf, ncf) ||
			isBigComplex && bigComplexIsEqual(wc, n.c) ||
			reflect.DeepEqual(w, n.c)
		if !oneOf {
			t.Errorf("Failure in values: %v plus %v not equal to %v", n.a, n.b, n.c)
		}
		if fmt.Sprintf("%T", w) != fmt.Sprintf("%T", n.c) {
			t.Errorf("Failure in types: %T not the same as %T", w, n.c)
		}
	}
}

//================================================================================
func TestNegate(t *testing.T) {
	for _, n := range []struct{ a, b interface{} }{
		{true, false},
		{false, true},
		{ops.Int(4), ops.Int(-4)},
		{ops.Int(0), ops.Int(0)},
		{newBigInt(0), newBigInt(0)},
		{newBigInt(-7), newBigInt(7)},
		{newBigRat(0, 1), newBigInt(0)},
		{newBigRat(1, 7), newBigRat(1, -7)},
		{ops.Float(0), ops.Float(0)},
		{ops.Float(-7.0), ops.Float(7.0)},
		{newBigFloat(1.23), newBigFloat(-1.23)},
		{ops.Complex(0 + 0i), ops.Complex(0 + 0i)},
		{ops.Complex(2 + 3i), ops.Complex(-2 - 3i)},
		{newBigComplex(151.0, 789.0), newBigComplex(-151.0, -789.0)},
	} {
		w := ops.Negate(n.a)
		checkEqual(t, w, n.b)
	}
}

//================================================================================
func TestMinus(t *testing.T) {
	for _, n := range []struct{ a, b, c interface{} }{
		//bool...
		{false, false, true}, // a-b is short for a+(-b), which acts like (a || !b) but without lazy eval
		{true, false, true},
		{false, true, false},
		{true, true, true},

		//TODO ???: eliminate this from the code...
		{int64(0x15a1), false, ops.Int(0x15a2)},
		{int64(0x15a1), true, ops.Int(0x15a1)},
		{0x15a1 + 789i, true, ops.Complex(0x15a1 + 789i)},

		//Int...
		{uint64(0x1), int64(0xF),
			ops.Int(-0xE)}, // uint64 - int64 -> Int
		{int64(0x1f1f), int64(0x1010),
			ops.Int(0x0f0f)}, // 2 Ints without overflow remain a Int
		{int64(0x1), uint64(0x8000000000000000),
			ops.BigInt(minInt64plusOne)}, // 2 Ints (i.e. int64 - uint64) promoted to BigInt when oflo
		{uint64(0x1), int64(-0x7FFFFFFFFFFFFFFF),
			ops.BigInt(maxInt64plusOne)}, // uint64 - int64 -> BigInt when overflow
		{maxInt64plusOne, int64(0x1),
			ops.BigInt(maxInt64)}, // BigInt - Int -> BigInt
		{int64(0x1), maxInt64plusOne,
			ops.BigInt(minInt64plusOne)}, // Int  - BigInt -> BigInt

		//BigInt...
		{newBigInt(-0x80000000), newBigInt(0x100000000),
			newBigInt(-0x180000000)}, // BigInt - BigInt -> BigInt
		{newBigInt(-0x80000000), newBigRat(0x100000000, 1),
			newBigInt(-0x180000000)}, // BigInt - BigRat -> BigRat, unless narrowed to BigInt
		{newBigInt(-0x80000000), 1.0,
			ops.Float(-0x80000001)}, // BigInt - float32 -> Float
		{newBigInt(-0x80000000), 1 + 0i,
			ops.Complex(-0x80000001 + 0i)}, // BigInt - complex64 -> Complex

		//BigRat...
		{newBigRat(-0x80000000, 1), newBigRat(0x100000000, 1),
			newBigInt(-0x180000000)}, // BigRat - BigRat -> BigRat, unless narrowed to BigInt
		{newBigRat(-0x80000000, 1), 1.0,
			ops.Float(-0x80000001)}, // BigRat - float32 -> Float
		{newBigRat(-0x80000000, 1), 1 + 0i,
			ops.Complex(-0x80000001 + 0i)}, // BigRat - complex64 -> Complex

		//Float...
		{1.1, 0 + 1i, ops.Complex(1.1 - 1i)}, // float32 - complex64 -> Complex

		//BigFloat...
		//{newBigFloat(3.1), ops.Float(1.95), newBigFloat(1.15)}, // BigFloat - Float -> BigFloat
		// TODO: FIX: //Failure in values: 3.1 minus 1.95 equals 1.15, instead of 1.15

		//Complex...
		{151 + 789i, 0 + 10i, ops.Complex(151 + 779i)}, // complex64 - complex64 -> Complex

	} {
		w := ops.Minus(n.a, n.b)
		wf, wIsBigFloat := w.(ops.BigFloat)
		ncf, ncIsBigFloat := n.c.(ops.BigFloat)
		wc, isBigComplex := w.(ops.BigComplex)
		oneOf := wIsBigFloat && ncIsBigFloat && bigFloatIsEqual(wf, ncf) ||
			isBigComplex && bigComplexIsEqual(wc, n.c) ||
			reflect.DeepEqual(w, n.c)
		if !oneOf {
			t.Errorf("Failure in values: %v minus %v equals %v, instead of %v", n.a, n.b, w, n.c)
		}
		if fmt.Sprintf("%T", w) != fmt.Sprintf("%T", n.c) {
			t.Errorf("Failure in types: %T not the same as %T", w, n.c)
		}
	}
}

//================================================================================
func TestMult(t *testing.T) {
	for _, n := range []struct{ a, b, c interface{} }{
		//bool...
		{false, false, false}, // * acts like && but without lazy eval
		{true, false, false},
		{false, true, false},
		{true, true, true},
		{int64(0x15a1), true, ops.Int(0x15a1)},
		{0x15a1 + 789i, true, ops.Complex(0x15a1 + 789i)},

		{int64(0x15a1), false, ops.Int(0)}, // false is 0; true is 1
		{int64(0x15a1), true, ops.Int(0x15a1)},
		{0x15a1 + 789i, false, ops.Complex(0 + 0i)},
		{0x15a1 + 789i, true, ops.Complex(0x15a1 + 789i)},

		//Int...
		{int64(0x1f1f), int64(0x1010),
			ops.Int(0x1f3e1f0)}, // 2 Ints without overflow remain a Int
		{int64(0x1), uint64(0x8000000000000000),
			ops.BigInt(maxInt64plusOne)}, // 2 Ints (i.e. int64 * uint64) promoted to BigInt when overflow
		{uint64(0x8000000000000000), int64(0x1),
			ops.BigInt(maxInt64plusOne)}, // uint64 * int64 -> BigInt
		{maxInt64plusOne, int64(0x1),
			ops.BigInt(maxInt64plusOne)}, // BigInt * Int -> BigInt
		{int64(0x1), maxInt64plusOne,
			ops.BigInt(maxInt64plusOne)}, // Int * BigInt -> BigInt

		//BigInt...
		{newBigInt(0x80000000), newBigInt(1),
			newBigInt(0x80000000)}, // BigInt * BigInt -> BigInt
		{newBigInt(0x80000000), newBigRat(1, 1),
			newBigInt(0x80000000)}, // BigInt * BigRat -> BigRat, unless narrowed to BigInt
		{newBigInt(-0x80000000), 1.0,
			ops.Float(-0x80000000)}, // BigInt * float32 -> Float
		{newBigInt(-0x80000000), 1 + 0i,
			ops.Complex(-0x80000000 + 0i)}, // BigInt * complex64 -> Complex

		//BigRat...
		{newBigRat(-0x80000000, 1), newBigRat(1, 1),
			newBigInt(-0x80000000)}, // BigRat * BigRat -> BigRat, unless narrowed to BigInt
		{newBigRat(-0x80000000, 1), 1.0,
			ops.Float(-0x80000000)}, // BigRat * float32 -> Float
		{newBigRat(-0x80000000, 1), 1 + 0i,
			ops.Complex(-0x80000000 + 0i)}, // BigRat * complex64 -> Complex

		//Float...
		{1.1, 0 + 1i, ops.Complex(1.1i)}, // float32 * complex64 -> Complex

		//TODO: TEST BigFloat

		//Complex...
		{151 + 789i, 0 + 1i, ops.Complex(-789 + 151i)}, // complex64 * complex64 -> Complex

	} {
		w := ops.Mult(n.a, n.b)
		if !reflect.DeepEqual(w, n.c) {
			t.Errorf("Failure in values: %v not the same as %v", w, n.c)
		}
		if fmt.Sprintf("%T", w) != fmt.Sprintf("%T", n.c) {
			t.Errorf("Failure in types: %T not the same as %T", w, n.c)
		}
	}
}

//================================================================================
func TestInvert(t *testing.T) {
	for _, n := range []struct{ a, b interface{} }{
		{nil, nil},
		{true, false},
		{false, true},
		{ops.Int(4), newBigRat(1, 4)},
		{ops.Float(4), ops.Float(0.25)},
		{ops.Complex(4i), ops.Complex(-0.25i)},
		{newBigFloat(-5.0), newBigFloat(-0.2)},

		//{newBigComplex(-5.0, 0.25), newBigComplex(-0.1995012469, -0.009975062344)},
		//TODO: FIX: Failure in values: -5+0.25i inverted equals -0.1995012469+-0.009975062344i,
		//               instead of -0.1995012469+-0.009975062344i
	} {
		w := ops.Invert(n.a)
		wf, wIsBigFloat := w.(ops.BigFloat)
		nbf, nbIsBigFloat := n.b.(ops.BigFloat)
		wc, isBigComplex := w.(ops.BigComplex)
		oneOf := wIsBigFloat && nbIsBigFloat && bigFloatIsEqual(wf, nbf) ||
			isBigComplex && bigComplexIsEqual(wc, n.b) ||
			reflect.DeepEqual(w, n.b)
		if !oneOf {
			t.Errorf("Failure in values: %v inverted equals %v, instead of %v", n.a, w, n.b)
		}
		if fmt.Sprintf("%T", w) != fmt.Sprintf("%T", n.b) {
			t.Errorf("Failure in types: %T not the same as %T", w, n.b)
		}
	}
}

//================================================================================
func TestDivide(t *testing.T) {
	for _, n := range []struct{ a, b, c interface{} }{

		//bool...
		{false, false, false},
		{true, false, true},
		{false, true, false},
		{true, true, false},

		//TODO: do we really want these six values ???...
		{int64(0x15a1), true, ops.Int(0)},          //do we want ops.Int(0x15a1) ???
		{0x15a1 + 789i, true, ops.Complex(0 + 0i)}, //do we want ops.Complex(0x15a1 + 789i) ???
		{int64(0x15a1), true, ops.Int(0)},          // division by false invalid
		{int64(0x15a1), false, ops.Int(0x15a1)},
		{0x15a1 + 789i, true, ops.Complex(0 + 0i)}, // division by false invalid
		{0x15a1 + 789i, false, ops.Complex(0x15a1 + 789i)},

		//Int...
		{int64(123), int64(3), newBigInt(41)},
		{uint64(0x8000000000000000), int64(0x1),
			ops.BigInt(maxInt64plusOne)}, // uint64 / int64  -> BigInt
		{int64(0x7fffffffffffffff), uint64(0x1),
			ops.BigInt(maxInt64)}, // int64  / uint64 -> BigInt
		{maxInt64plusOne, int64(0x1),
			ops.BigInt(maxInt64plusOne)}, // BigInt / Int  -> BigInt

		//BigInt...
		{newBigInt(0x80000000), newBigInt(1),
			newBigInt(0x80000000)}, // BigInt / BigInt -> BigInt
		{newBigInt(0x80000000), newBigRat(1, 1),
			newBigInt(0x80000000)}, // BigInt / BigRat -> BigInt if no integral
		{newBigInt(-0x80000000), 1.0,
			ops.Float(-0x80000000)}, // BigInt / float32 -> Float
		{newBigInt(-0x80000000), 1 + 0i,
			ops.Complex(-0x80000000 + 0i)}, // BigInt / complex64 -> Complex

		//BigRat...
		{newBigRat(-0x80000000, 1), newBigRat(1, 1),
			newBigInt(-0x80000000)}, // BigRat / BigRat -> BigInt if no integral
		{newBigRat(-0x80000000, 1), 1.0,
			ops.Float(-0x80000000)}, // BigRat / float32 -> Float
		{newBigRat(-0x80000000, 1), 1 + 0i,
			ops.Complex(-0x80000000 + 0i)}, // BigRat / complex64 -> Complex

		//Float...
		{1.1, 0 + 1i, ops.Complex(-1.1i)}, // float32 * complex64 -> Complex

		//TODO: TEST for BigFloat

		//Complex...
		{151 + 789i, 0 + 1i, ops.Complex(789 - 151i)}, // complex64 * complex64 -> Complex

		{nil, nil, nil},
	} {
		w := ops.Divide(n.a, n.b)
		wc, wIsComplex := w.(ops.Complex)
		ncc, ncIsComplex := n.c.(ops.Complex)
		oneOf := wIsComplex && ncIsComplex && complexIsEqual(wc, ncc) ||
			reflect.DeepEqual(w, n.c)
		if !oneOf {
			t.Errorf("Failure in values: %v not the same as %v", w, n.c)
		}
		if fmt.Sprintf("%T", w) != fmt.Sprintf("%T", n.c) {
			t.Errorf("Failure in types: %T not the same as %T", w, n.c)
		}
	}
}

//================================================================================
//TODO: TEST Mod

//================================================================================
func TestNilInf(t *testing.T) {
	//unary ops...
	for _, n := range []struct{ a, fa, b interface{} }{

		{nil, ops.Identity, nil},
		{ops.Inf, ops.Identity, ops.Inf},

		{nil, ops.Negate, nil},
		{ops.Inf, ops.Negate, ops.Inf},

		{nil, ops.Invert, nil},

		// 1/0 is inf
		{ops.Int(0), ops.Invert, ops.Inf},
		{newBigInt(0), ops.Invert, ops.Inf},
		{newBigRat(0, 1), ops.Invert, ops.Inf},
		{0.0, ops.Invert, ops.Inf},
		{-0.0, ops.Invert, ops.Inf}, //only one Inf in Gro, both +ve and -ve
		{ops.Float(0), ops.Invert, ops.Inf},
		{newBigFloat(0.0), ops.Invert, ops.Inf},
		{ops.Complex(0 + 0i), ops.Invert, ops.Inf},
		{newBigComplex(0.0, 0.0), ops.Invert, ops.Inf},

		// 1/inf is 0
		{ops.Inf, ops.Invert, 0},

		// 1/1 is 1
		{ops.Int(1), ops.Invert, newBigInt(1)}, //should be ops.Int(1)
		{newBigInt(1), ops.Invert, newBigInt(1)},
		{newBigRat(1, 1), ops.Invert, newBigInt(1)}, //should be ops.Int(1)
		{1.0, ops.Invert, 1.0},
		{-1.0, ops.Invert, -1.0},
		{ops.Float(1), ops.Invert, ops.Float(1)},
		{newBigFloat(1.0), ops.Invert, newBigFloat(1.0)},
		{ops.Complex(1 + 0i), ops.Invert, ops.Complex(1 + 0i)},
		//{newBigComplex(1.0, 0.0), ops.Invert, newBigComplex(1.0, -0.0)}, //TODO: need to change all -0's to +0
		//Failure in values: 1+-0i not the same as 1+0i
	} {
		fn := n.fa.(func(a interface{}) interface{})
		w := fn(n.a)
		wc, wIsComplex := w.(ops.Complex)
		oneOf := wIsComplex && complexIsEqual(wc, n.b) || reflect.DeepEqual(w, n.b)
		if !oneOf {
			t.Errorf("Failure in values: %v not the same as %v", w, n.b)
		}
		if fmt.Sprintf("%T", w) != fmt.Sprintf("%T", n.b) {
			t.Errorf("Failure in types: %T not the same as %T", w, n.b)
		}
	}

	//----------------------------------------------------------------------------
	//binary ops...
	for _, n := range []struct{ a, b, fab, c interface{} }{

		//TODO: IsEqual, IsLessThan

		{nil, nil, ops.Plus, nil},
		{nil, 123, ops.Plus, nil},
		{456, nil, ops.Plus, nil},

		{1, ops.Inf, ops.Plus, ops.Inf}, // z + Inf = Inf, for z<>Inf
		{0, ops.Inf, ops.Plus, ops.Inf},

		//in Go:
		//  math.IsNaN(math.Inf(-1) + math.Inf(1))
		//  math.IsNaN(math.Inf(1) + math.Inf(-1))
		//  math.IsNaN(math.Inf(1) - math.Inf(1))
		//  math.IsNaN(math.Inf(-1) - math.Inf(-1))
		//but in Go:
		//  math.IsInf(math.Inf(1)+math.Inf(1), 1)
		//  math.IsInf(math.Inf(-1)+math.Inf(-1), -1)
		//  math.IsInf(math.Inf(1)-math.Inf(-1), 1)
		//  math.IsInf(math.Inf(-1)-math.Inf(1), -1)
		//however in Gro, we define Inf + Inf = NaN and Inf - Inf = NaN because of mathematics
		{ops.Inf, ops.Inf, ops.Plus, nil},
		{ops.Inf, ops.Inf, ops.Minus, nil},

		{nil, nil, ops.Minus, nil},
		{nil, 123, ops.Minus, nil},

		{nil, nil, ops.Mult, nil},
		{nil, 123, ops.Mult, nil},

		{0, ops.Inf, ops.Mult, nil},     // 0 * Inf = NaN
		{1, ops.Inf, ops.Mult, ops.Inf}, // z * Inf = Inf, for z<>0
		{-1, ops.Inf, ops.Mult, ops.Inf},
		{789, ops.Inf, ops.Mult, ops.Inf},

		{ops.Inf, ops.Inf, ops.Mult, ops.Inf}, // Inf * Inf = Inf

		{nil, nil, ops.Divide, nil},
		{nil, 123, ops.Divide, nil},
		{123, nil, ops.Divide, nil},

		{0, 0, ops.Divide, nil},     // 0 / 0 = NaN
		{1, 0, ops.Divide, ops.Inf}, // z / 0 = Inf, for z<>0
		{-1, 0, ops.Divide, ops.Inf},
		{789, 0, ops.Divide, ops.Inf},

		{ops.Inf, 0, ops.Divide, ops.Inf},      // Inf / 0 = Inf
		{0, ops.Inf, ops.Divide, ops.Int(0)},   // 0 / Inf = 0
		{1, ops.Inf, ops.Divide, newBigInt(0)}, // z / Inf = 0, for z<>0
		{-1, ops.Inf, ops.Divide, newBigInt(0)},
		{ops.Inf, ops.Inf, ops.Divide, nil}, // Inf / Inf = NaN

	} {
		fn := n.fab.(func(a, b interface{}) interface{})
		w := fn(n.a, n.b)
		wc, wIsComplex := w.(ops.Complex)
		oneOf := wIsComplex && complexIsEqual(wc, n.c) || reflect.DeepEqual(w, n.c)
		if !oneOf {
			t.Errorf("Failure in values: %v not the same as %v", w, n.c)
		}
		if fmt.Sprintf("%T", w) != fmt.Sprintf("%T", n.c) {
			t.Errorf("Failure in types: %T not the same as %T", w, n.c)
		}
	}
}

//================================================================================
func TestPosIntRange(t *testing.T) {
	for _, n := range []struct {
		r       ops.PosIntRange
		a, b, c interface{}
	}{
		{newPosIntRange(1, 3), ops.Int(1), ops.Int(3), false},
		{newPosIntRange(1, 1), ops.Int(1), ops.Int(1), false},
		{newPosIntRange(17, ops.BigInt(maxUint64)),
			ops.Int(17), ops.Int(0x7fffffffffffffff /*maxInt64*/), false},
		{newPosIntRange(7, ops.Inf), ops.Int(7), ops.Int(0x7fffffffffffffff), true},
	} {
		//TODO: more tests
		if !reflect.DeepEqual(n.r.From(), n.a) {
			t.Errorf("Failure in From values: %v not the same as %v", n.r.From(), n.a)
		}
		if !reflect.DeepEqual(n.r.To(), n.b) {
			t.Errorf("Failure in To values: %v not the same as %v", n.r.To(), n.b)
		}
		if !reflect.DeepEqual(n.r.IsToInf(), n.c) {
			t.Errorf("Failure in IsToInf predicate: %v not the same as %v", n.r.IsToInf(), n.c)
		}
	}

	if !(newPosIntRange(789, ops.Inf).IsToInf()) {
		t.Error("Failure in one-off test")
	}
}

//================================================================================
func TestDate(t *testing.T) {
	//unary ops...
	for _, n := range []struct{ a, fa, b interface{} }{

		{time.Date(2009, time.November, 10, 23, 22, 21, 20, time.UTC), ops.Identity,
			time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC)},
	} {
		fn := n.fa.(func(a interface{}) interface{})
		w := fn(n.a)
		if !reflect.DeepEqual(w, n.b) {
			t.Errorf("Failure in values: %v not equal to %v", w, n.b)
		}
		if fmt.Sprintf("%T", w) != fmt.Sprintf("%T", n.b) {
			t.Errorf("Failure in types: %T not the same as %T", w, n.b)
		}
	}

	//----------------------------------------------------------------------------
	//binary ops...
	for _, n := range []struct{ a, b, fab, c interface{} }{

		{time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC), 1, ops.Plus,
			time.Date(2009, time.November, 11, 0, 0, 0, 0, time.UTC)}, //gro: 2009.11.10 + 1

		{time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC), 3, ops.Minus,
			time.Date(2009, time.November, 7, 0, 0, 0, 0, time.UTC)}, //gro: 2009.11.10 - 3

		{1, time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC), ops.Plus,
			time.Date(2009, time.November, 11, 0, 0, 0, 0, time.UTC)}, //gro: 1 + 2009.11.10

	} {
		fn := n.fab.(func(a, b interface{}) interface{})
		w := fn(n.a, n.b)
		if !reflect.DeepEqual(w, n.c) {
			t.Errorf("Failure in values: %v not equal to %v", w, n.c)
		}
		if fmt.Sprintf("%T", w) != fmt.Sprintf("%T", n.c) {
			t.Errorf("Failure in types: %T not the same as %T", w, n.c)
		}
	}

	//----------------------------------------------------------------------------
	//predicate binary ops...
	for _, n := range []struct{ a, b, fab, c interface{} }{

		//gro: 2009.11.10 == 2009.11.10
		{time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC),
			time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC), ops.IsEqual, true},

		//gro: 2009.11.10 == 2009.11.11
		{time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC),
			time.Date(2009, time.November, 11, 0, 0, 0, 0, time.UTC), ops.IsEqual, false},

		//gro: 2009.11.10 < 2009.11.9
		{time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC),
			time.Date(2009, time.November, 9, 0, 0, 0, 0, time.UTC), ops.IsLessThan, false},

		//gro: 2009.11.10 < 2009.11.10
		{time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC),
			time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC), ops.IsLessThan, false},

		//gro: 2009.11.10 < 2009.11.11
		{time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC),
			time.Date(2009, time.November, 11, 0, 0, 0, 0, time.UTC), ops.IsLessThan, true},
	} {
		fn := n.fab.(func(a, b interface{}) bool)
		w := fn(n.a, n.b)
		if !reflect.DeepEqual(w, n.c) {
			t.Errorf("Failure in values: %v not equal to %v", w, n.c)
		}
		if fmt.Sprintf("%T", w) != fmt.Sprintf("%T", n.c) {
			t.Errorf("Failure in types: %T not the same as %T", w, n.c)
		}
	}
}

//================================================================================
