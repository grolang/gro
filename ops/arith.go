// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package ops

import (
	"math"
	"math/big"
	"math/cmplx"
	"reflect"
	"time"

	u8 "github.com/grolang/gro/utf88"
)

type (
	Int      int64
	BigInt   big.Int
	BigRat   big.Rat
	Float    float64
	BigFloat big.Float
	Complex  complex128
)

var (
	Inf = Complex(cmplx.Inf())
	NaN = Complex(cmplx.NaN())
)

/*
BigComplex consists of real and imaginary components, each of type big.Float
as introduced in the Go 1.5 math/big package.
*/
type BigComplex struct {
	Re, Im big.Float
}

/*
PosIntRange represents a range in the positive integers.
From can be any Int or BigInt from 0 up.
To can be any Int or BigInt from 0 up, or be Float.Inf(1).
*/
type PosIntRange struct {
	From, To interface{}
}

/*
FromVal of PosIntRange returns From as an Int. It returns 0, however, if it's negative,
or returns the maximum Int if it's a BigInt greater than that.

TODO: more tests.
*/
func (r PosIntRange) FromVal() Int {
	switch r.From.(type) {
	case Int:
		if r.From.(Int) < 0 {
			return Int(0)
		} else {
			return r.From.(Int)
		}
	case BigInt:
		ar := big.Int(r.From.(BigInt))
		if big.NewInt(0).Cmp(&ar) == 1 {
			return Int(0)
		} else if big.NewInt(0x7fffffffffffffff).Cmp(&ar) == -1 {
			return Int(0x7fffffffffffffff)
		} else {
			return Int(ar.Int64())
		}
	default:
		panic("PosIntRange.FromVal: invalid input type")
	}
}

/*
ToVal of PosIntRange returns To as an Int. It returns 0, however, if it's negative,
and returns the maximum Int if it's a BigInt greater than that, or infinity.

TODO: more tests.
*/
func (r PosIntRange) ToVal() Int {
	switch r.To.(type) {
	case float64:
		if math.IsInf(r.To.(float64), 1) {
			return Int(0x7fffffffffffffff)
		} else {
			panic("PosIntRange.ToVal: invalid input type")
		}
	case Int:
		if r.To.(Int) < 0 {
			return Int(0)
		} else {
			return r.To.(Int)
		}
	case BigInt:
		ar := big.Int(r.To.(BigInt))
		if big.NewInt(0).Cmp(&ar) == 1 {
			return Int(0)
		} else if big.NewInt(0x7fffffffffffffff).Cmp(&ar) == -1 {
			return Int(0x7fffffffffffffff)
		} else {
			return Int(ar.Int64())
		}
	default:
		panic("PosIntRange.ToVal: invalid input type")
	}
}

/*
IsToInf of PosIntRange returns true if To is infinity, otherwise false.
*/
func (r PosIntRange) IsToInf() bool {
	floatVal, isFloat := r.To.(float64)
	return isFloat && math.IsInf(floatVal, 1)
}

/*
NewBigInt returns a new BigInt seeded by an int64 value.
*/
func NewBigInt(x int64) BigInt {
	return BigInt(*big.NewInt(x))
}

/*
NewBigRat returns a new BigRat seeded by two int64 values.
*/
func NewBigRat(x, y int64) BigRat {
	return BigRat(*big.NewRat(x, y))
}

/*
NewBigFloat returns a new BigFloat seeded by a float64 value,
a big.Float value, or another BigFloat value.
*/
func NewBigFloat(x interface{}) BigFloat {
	switch x.(type) {
	case float64:
		xf := x.(float64)
		if math.IsInf(xf, 1) || math.IsInf(xf, -1) || math.IsNaN(xf) {
			panic("NewBigFloat: Arg can't be infinity or NaN")
		} else {
			return BigFloat(*big.NewFloat(xf))
		}
	case big.Float:
		return BigFloat(x.(big.Float))
	case BigFloat:
		return x.(BigFloat)
	default:
		panic("NewBigFloat: invalid input type")
	}
}

/*
NewBigComplex returns a new BigComplex, the real and imaginary values each
seeded by a float64 value, a big.Float value, or a BigFloat value.
*/
func NewBigComplex(x, y interface{}) BigComplex {
	xf := big.Float(NewBigFloat(x))
	yf := big.Float(NewBigFloat(y))

	if xf.IsInf() || yf.IsInf() {
		panic("NewBigComplex: Neither arg can be infinity")
	} else {
		return BigComplex{xf, yf}
	}
}

/*
String of BigInt returns a string representation of its value.
*/
func (n BigInt) String() string {
	nr := big.Int(n)
	return nr.String()
}

/*
String of BigRat returns a string representation of its value.
*/
func (n BigRat) String() string {
	nr := big.Rat(n)
	return nr.String()
}

/*
String of BigFloat returns a string representation of its value.
*/
func (n BigFloat) String() string {
	x := big.Float(n)
	return x.String()
}

/*
String of BigComplex returns a string representation of its value.
*/
func (n BigComplex) String() string {
	r := big.Float(n.Re)
	i := big.Float(n.Im)
	return r.String() + "+" + i.String() + "i"
}

/*
IsEqual of Float returns true if this equals that.
*/
func (this Float) IsEqual(that interface{}) bool {
	switch that.(type) {
	case Float:
		x := that.(Float)
		if math.IsNaN(float64(x)) && math.IsNaN(float64(this)) {
			return true
		}
		return float64(x) == float64(this)

	default:
		panic("Float.IsEqual: only Floats are valid args for now")
	}
}

/*
IsEqual of BigFloat returns true if this equals that.
*/
func (this BigFloat) IsEqual(that interface{}) bool {
	switch that.(type) {

	case Float:
		x := big.Float(this)
		y := big.Float(NewBigFloat(that.(Float)))
		return x.Cmp(&y) == 0

	case BigFloat:
		x := big.Float(this)
		y := big.Float(that.(BigFloat))
		return x.Cmp(&y) == 0

	default:
		panic("BigFloat.IsEqual: only numerics of lower weighting are valid args")
	}
}

/*
IsEqual of Complex returns true if this equals that.
*/
func (this Complex) IsEqual(that interface{}) bool {
	switch that.(type) {
	case Complex:
		x := that.(Complex)
		if cmplx.IsNaN(complex128(x)) && cmplx.IsNaN(complex128(this)) {
			return true
		}
		return complex128(x) == complex128(this)

	default:
		panic("Complex.IsEqual: only Complexes are valid args for now")
	}
}

/*
IsEqual of BigComplex returns true if this equals that.
*/
func (this BigComplex) IsEqual(that interface{}) bool {
	switch that.(type) {

	case BigFloat:
		x := big.Float(that.(BigFloat))
		y := *big.NewFloat(0)
		return this.Re.Cmp(&x) == 0 && this.Im.Cmp(&y) == 0

	case Complex:
		x := that.(Complex)
		xr := *big.NewFloat(real(complex128(x)))
		xi := *big.NewFloat(imag(complex128(x)))
		return this.Re.Cmp(&xr) == 0 && this.Im.Cmp(&xi) == 0

	case BigComplex:
		x := that.(BigComplex)
		xr := x.Re
		xi := x.Im
		return this.Re.Cmp(&xr) == 0 && this.Im.Cmp(&xi) == 0

	default:
		panic("BigComplex.IsEqual: only numerics of lower weighting are valid args")
	}
}

var typeWeights = map[reflect.Type]int{
	nil: 1,
	reflect.TypeOf(false):                   2,
	reflect.TypeOf(Int(0)):                  3,
	reflect.TypeOf(NewBigInt(0)):            4,
	reflect.TypeOf(NewBigRat(0, 1)):         5,
	reflect.TypeOf(Float(0.0)):              6,
	reflect.TypeOf(NewBigFloat(0.0)):        7,
	reflect.TypeOf(Complex(0 + 0i)):         8,
	reflect.TypeOf(NewBigComplex(0.0, 0.0)): 9,
}

func widen(x interface{}) interface{} {
	switch x.(type) {
	case int:
		return Int(x.(int))
	case int8:
		return Int(x.(int8))
	case int16:
		return Int(x.(int16))
	case int32:
		return Int(x.(int32))
	case int64:
		return Int(x.(int64))

	case uint8:
		return Int(x.(uint8))
	case uint16:
		return Int(x.(uint16))
	case uint32:
		return Int(x.(uint32))

	case uint64:
		if x.(uint64) < (1 << 63) {
			return Int(x.(uint64))
		} else {
			x := big.NewInt(int64(x.(uint64) - (1 << 63)))
			y := big.NewInt(1 << 62)
			return BigInt(*x.Add(x, y).Add(x, y))
		}

	case uint:
		if uint64(x.(uint)) < (1 << 63) {
			return Int(x.(uint))
		} else {
			x := big.NewInt(int64(x.(uint)) - (1 << 62) - (1 << 62))
			y := big.NewInt(1 << 62)
			return BigInt(*x.Add(x, y).Add(x, y))
		}

	case float32:
		return narrow(Float(x.(float32)))
	case float64:
		return narrow(Float(x.(float64)))
	case complex64:
		return narrow(Complex(x.(complex64)))
	case complex128:
		return narrow(Complex(x.(complex128)))

	case big.Int:
		return BigInt(x.(big.Int))
	case big.Rat:
		return narrow(BigRat(x.(big.Rat)))
	case big.Float:
		return BigFloat(x.(big.Float))

	case time.Time:
		d := x.(time.Time)
		return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)

	case string:
		return u8.Desur(x.(string))
		//MAYBE TODO: check if string is a valid Codepoint first ???

	/*case []interface{}: //TODO: extend logic to handle slices of any typed value
	vs := []interface{}{}
	for _, v := range x.([]interface{}) {
		vs = append(vs, widen(v))
	}
	return Slice(vs)
	*/

	/*case map[interface{}]interface{}: //TODO: extend logic to handle maps of any typed key and value
	es := map[interface{}]interface{}{}
	for k, v := range x.(map[interface{}]interface{}) {
		es[widen(k)] = widen(v)
	}
	return es
	*/

	default: //includes nil, bool
		return x
	}
}

func narrow(x interface{}) interface{} {
	switch x.(type) {
	case nil, bool, Int, BigInt:
		return x

	//TODO: narrow BigInt to Int for low BigInt
	//MAYBE TODO: narrow 0.0+0.0i etc to narrower 0

	case BigRat:
		xf := big.Rat(x.(BigRat))
		if reflect.DeepEqual(*xf.Denom(), *big.NewInt(1)) {
			return BigInt(*xf.Num())
		} else {
			return x
		}

	case Float:
		xf := float64(x.(Float))
		if math.IsInf(xf, -1) || math.IsInf(xf, 1) {
			return Inf
		} else if math.IsNaN(xf) {
			return NaN
		} else {
			return x
		}

	case BigFloat:
		xf := big.Float(x.(BigFloat))
		if xf.IsInf() {
			return Inf
		} else {
			return x
		}

	case Complex:
		xc := complex128(x.(Complex))
		if cmplx.IsInf(xc) {
			return Inf
		} else if cmplx.IsNaN(xc) {
			return NaN
		} else {
			return x
		}

	case BigComplex:
		xr := x.(BigComplex).Re
		xi := x.(BigComplex).Im
		if xr.IsInf() || xi.IsInf() {
			return Inf
		} else {
			return x
		}

	default:
		return x
	}
}

/*
Identity converts:
  float32 and float64 to Float (float64), but NaN is converted to Complex NaN
  complex64 and complex128 to Complex (complex128)
  int, int8, int16, int32/rune, int64, uint8, uint16, and uint32 to Int
  uint and uint64 to Int, unless it overflows, in which case it converts to BigInt
  string to u8.Text
  eliminates fractions of days from Date, and switches to UTC
and leaves the nil, bool, BigInt, BigRat, and BigFloat types unchanged.
*/
func Identity(x interface{}) interface{} {
	return widen(x)
}

/*
IsEqual compares the Identity of each argument to see if they are equal,
irrespective of their type.
Numerics and booleans are compared with each other, Dates with each other,
and Text with each other, otherwise, reflect.DeepEqual is called.
*/
func IsEqual(x, y interface{}) bool {
	x, y = widen(x), widen(y)
	if typeWeights[reflect.TypeOf(x)] > typeWeights[reflect.TypeOf(y)] {
		x, y = y, x
	}

	switch x.(type) {
	case nil:
		switch y.(type) {
		case nil:
			return true
		case bool:
			return !x.(bool)
		case Int, BigInt, BigRat, Float, BigFloat, Complex, BigComplex:
			return !ToBool(x)
		default:
			panic("IsEqual: nil and non-numeric invalid")
		}

	case bool:
		switch y.(type) {
		case bool:
			return x.(bool) == y.(bool)
		case Int, BigInt, BigRat, Float, BigFloat, Complex, BigComplex:
			return x.(bool) == ToBool(y)
		default:
			panic("IsEqual: bool and non-numeric invalid")
		}

	case Int:
		switch y.(type) {
		case Int:
			return x.(Int) == y.(Int)
		case BigInt:
			ay := big.Int(y.(BigInt))
			return big.NewInt(int64(x.(Int))).Cmp(&ay) == 0
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return big.NewRat(int64(x.(Int)), 1).Cmp(&ay) == 0
		case Float:
			return Float(x.(Int)) == y.(Float)
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return big.NewFloat(float64(x.(Int))).Cmp(&ay) == 0
		case Complex:
			return Complex(complex(float64(x.(Int)), 0)) == y.(Complex)
		case BigComplex:
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			ayi := big.Float(BigFloat(y.(BigComplex).Im))
			return big.NewFloat(float64(x.(Int))).Cmp(&ayr) == 0 && big.NewFloat(0.0).Cmp(&ayi) == 0
		default:
			panic("IsEqual: Int and non-numeric invalid")
		}

	case BigInt:
		switch y.(type) {
		case BigInt:
			ax := big.Int(x.(BigInt))
			ay := big.Int(y.(BigInt))
			return ax.Cmp(&ay) == 0
		case BigRat:
			ax := big.Int(x.(BigInt))
			ay := big.Rat(y.(BigRat))
			return big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Cmp(&ay) == 0
		case Float:
			ax := big.Int(x.(BigInt))
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return Float(xf) == y.(Float)
		case BigFloat:
			ax := big.Int(x.(BigInt))
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			ay := big.Float(y.(BigFloat))
			return big.NewFloat(xf).Cmp(&ay) == 0
		case Complex:
			ax := big.Int(x.(BigInt))
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return Complex(complex(xf, 0)) == y.(Complex)
		case BigComplex:
			ax := big.Int(x.(BigInt))
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			ayi := big.Float(BigFloat(y.(BigComplex).Im))
			return big.NewFloat(xf).Cmp(&ayr) == 0 && big.NewFloat(0.0).Cmp(&ayi) == 0
		default:
			panic("IsEqual: BigInt and non-numeric invalid")
		}

	case BigRat:
		switch y.(type) {
		case BigRat:
			ax := big.Rat(x.(BigRat))
			ay := big.Rat(y.(BigRat))
			return ax.Cmp(&ay) == 0
		case Float:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			return Float(xf) == y.(Float)
		case BigFloat:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			ay := big.Float(y.(BigFloat))
			return big.NewFloat(xf).Cmp(&ay) == 0
		case Complex:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			return Complex(complex(xf, 0)) == y.(Complex)
		case BigComplex:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			ayi := big.Float(BigFloat(y.(BigComplex).Im))
			return big.NewFloat(xf).Cmp(&ayr) == 0 && big.NewFloat(0.0).Cmp(&ayi) == 0
		default:
			panic("IsEqual: BigRat and non-numeric invalid")
		}

	case Float:
		switch y.(type) {
		case Float:
			return x.(Float) == y.(Float)
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return big.NewFloat(float64(x.(Float))).Cmp(&ay) == 0
		case Complex:
			return Complex(complex(x.(Float), 0)) == y.(Complex)
		case BigComplex:
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			ayi := big.Float(BigFloat(y.(BigComplex).Im))
			return big.NewFloat(float64(x.(Float))).Cmp(&ayr) == 0 && big.NewFloat(0.0).Cmp(&ayi) == 0
		default:
			panic("IsEqual: Float and non-numeric invalid")
		}

	case BigFloat:
		switch y.(type) {
		case BigFloat:
			ax := big.Float(x.(BigFloat))
			ay := big.Float(y.(BigFloat))
			return ax.Cmp(&ay) == 0
		case Complex:
			ax := big.Float(x.(BigFloat))
			ay := big.Float(NewBigFloat(real(y.(Complex))))
			return ax.Cmp(&ay) == 0 && imag(y.(Complex)) == 0
		case BigComplex:
			return y.(BigComplex).IsEqual(x)
		default:
			panic("IsEqual: BigFloat and non-numeric invalid")
		}

	case Complex:
		switch y.(type) {
		case Complex:
			return x.(Complex) == y.(Complex)
		case BigComplex:
			return y.(BigComplex).IsEqual(x)
		default:
			panic("IsEqual: Complex and non-numeric invalid")
		}

	case BigComplex:
		switch y.(type) {
		case BigComplex:
			return x.(BigComplex).IsEqual(y)
		default:
			panic("IsEqual: BigComplex and non-numeric invalid")
		}

	case time.Time:
		switch y.(type) {
		case time.Time:
			return x.(time.Time).Equal(y.(time.Time))
		default:
			panic("IsEqual: Date and non-Date invalid")
		}

	case u8.Text:
		switch y.(type) {
		case u8.Text:
			return string(u8.SurrogatePoints(x.(u8.Text))) == string(u8.SurrogatePoints(y.(u8.Text)))
		default:
			panic("IsEqual: u8.Text and non-u8.Text invalid")
		}

	default:
		return reflect.DeepEqual(x, y)
	}
}

/*
IsLessThan returns true if both arguments are non-complex numeric or Dates,
and x is less than y.
It panics if either arg is nil, bool, Complex, BigComplex, or non-numeric.
*/
func IsLessThan(x, y interface{}) bool {
	x, y = widen(x), widen(y)
	if typeWeights[reflect.TypeOf(x)] > typeWeights[reflect.TypeOf(y)] {
		return !IsLessThan(y, x) && !IsEqual(x, y)
	}

	//MAYBE TODO: add cases comparing Complex/BigComplex with each other using absolute values ???
	//MAYBE TODO: use Absolute Value if at least one arg is Complex or BigComplex ???
	//MAYBE TODO: nil lower than everything ???

	switch x.(type) {
	case Int:
		switch y.(type) {
		case Int:
			return x.(Int) < y.(Int)
		case BigInt:
			ay := big.Int(y.(BigInt))
			return big.NewInt(int64(x.(Int))).Cmp(&ay) == -1
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return big.NewRat(int64(x.(Int)), 1).Cmp(&ay) == -1
		case Float:
			return Float(x.(Int)) < y.(Float)
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return big.NewFloat(float64(x.(Int))).Cmp(&ay) == -1
		default:
			panic("IsLessThan: only Int, BigInt, Rat, Float, and BigFloat args valid")
		}

	case BigInt:
		switch y.(type) {
		case BigInt:
			ax := big.Int(x.(BigInt))
			ay := big.Int(y.(BigInt))
			return ax.Cmp(&ay) == -1
		case BigRat:
			ax := big.Int(x.(BigInt))
			ay := big.Rat(y.(BigRat))
			return big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Cmp(&ay) == -1
		case Float:
			ax := big.Int(x.(BigInt))
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return Float(xf) < y.(Float)
		case BigFloat:
			ax := big.Int(x.(BigInt))
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			ay := big.Float(y.(BigFloat))
			return big.NewFloat(xf).Cmp(&ay) == -1
		default:
			panic("IsLessThan: only Int, BigInt, Rat, Float, and BigFloat args valid")
		}

	case BigRat:
		switch y.(type) {
		case BigRat:
			ax := big.Rat(x.(BigRat))
			ay := big.Rat(y.(BigRat))
			return ax.Cmp(&ay) == -1
		case Float:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			return Float(xf) < y.(Float)
		case BigFloat:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			ay := big.Float(y.(BigFloat))
			return big.NewFloat(xf).Cmp(&ay) == -1
		default:
			panic("IsLessThan: only Int, BigInt, Rat, Float, and BigFloat args valid")
		}

	case Float:
		switch y.(type) {
		case Float:
			return x.(Float) < y.(Float)
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return big.NewFloat(float64(x.(Float))).Cmp(&ay) == -1
		default:
			panic("IsLessThan: only Int, BigInt, Rat, Float, and BigFloat args valid")
		}

	case BigFloat:
		switch y.(type) {
		case BigFloat:
			ax := big.Float(x.(BigFloat))
			ay := big.Float(y.(BigFloat))
			return ax.Cmp(&ay) == -1
		default:
			panic("IsLessThan: only Int, BigInt, Rat, Float, and BigFloat args valid")
		}

	case time.Time:
		switch y.(type) {
		case time.Time:
			return x.(time.Time).Before(y.(time.Time))
		default:
			panic("IsLessThan: Date and non-Date invalid")
		}

	default:
		panic("IsLessThan: only Int, BigInt, Rat, Float, BigFloat, and Date args valid")
	}
}

/*
IsNotEqual returns true if both arguments are non-complex numeric or Dates,
and x is not equal to y.
It panics if either arg is nil, bool, Complex, BigComplex, or non-numeric.

//TODO: TEST IT
*/
func IsNotEqual(x, y interface{}) bool {
	return !IsEqual(x, y)
}

/*
IsLessOrEqual returns true if both arguments are non-complex numeric or Dates,
and x is less than or equal to y.
It panics if either arg is nil, bool, Complex, BigComplex, or non-numeric.

//TODO: TEST IT
*/
func IsLessOrEqual(x, y interface{}) bool {
	return IsLessThan(x, y) || IsEqual(x, y)
}

/*
IsGreaterThan returns true if both arguments are non-complex numeric or Dates,
and x is greater than y.
It panics if either arg is nil, bool, Complex, BigComplex, or non-numeric.

//TODO: TEST IT
*/
func IsGreaterThan(x, y interface{}) bool {
	return !IsLessOrEqual(x, y)
}

/*
IsGreaterOrEqual returns true if both arguments are non-complex numeric or Dates,
and x is greater than or equal to y.
It panics if either arg is nil, bool, Complex, BigComplex, or non-numeric.

//TODO: TEST IT
*/
func IsGreaterOrEqual(x, y interface{}) bool {
	return !IsLessThan(x, y)
}

/*
Plus creates a new element being the sum of the two arguments, for numeric types.
Two Ints are promoted to BigInt if it overflows.
For Date and integer, create new Date being y number of days added to Date x.
For Text and Codepoint, create new Text being concatenation.
For Slice x, return a new slice with element y appended.
*/
func Plus(x, y interface{}) interface{} {
	x, y = widen(x), widen(y)
	if typeWeights[reflect.TypeOf(x)] > typeWeights[reflect.TypeOf(y)] {
		x, y = y, x
	}

	switch x.(type) {
	case nil:
		switch y.(type) {
		case nil, bool, Int, BigInt, BigRat, Float, BigFloat, Complex, BigComplex:
			return y
		default:
			panic("Plus: nil and non-numeric invalid")
		}

	case bool:
		switch y.(type) {
		case bool:
			return x.(bool) || y.(bool)
		case Int:
			if x.(bool) {
				return 1 + y.(Int)
			} else {
				return y.(Int)
			}
		case BigInt:
			if x.(bool) {
				ay := big.Int(y.(BigInt))
				return BigInt(*big.NewInt(0).Add(big.NewInt(1), &ay))
			} else {
				return y.(BigInt)
			}
		case BigRat:
			if x.(bool) {
				ay := big.Rat(y.(BigRat))
				return BigRat(*big.NewRat(0, 1).Add(big.NewRat(1, 1), &ay))
			} else {
				return y.(BigRat)
			}
		case Float:
			if x.(bool) {
				return 1.0 + y.(Float)
			} else {
				return y.(Float)
			}
		case BigFloat:
			if x.(bool) {
				ay := big.Float(y.(BigFloat))
				return BigFloat(*big.NewFloat(0.0).Add(big.NewFloat(1.0), &ay))
			} else {
				return y.(BigFloat)
			}
		case Complex:
			if x.(bool) {
				return 1 + y.(Complex)
			} else {
				return y.(Complex)
			}
		case BigComplex:
			if x.(bool) {
				ayr := big.Float(BigFloat(y.(BigComplex).Re))
				return BigComplex{*big.NewFloat(0.0).Add(big.NewFloat(1.0), &ayr),
					y.(BigComplex).Im}
			} else {
				return y.(BigComplex)
			}
		default:
			panic("Plus: bool and non-numeric invalid")
		}

	case Int:
		switch y.(type) {
		case Int:
			c := x.(Int) + y.(Int)
			if (x.(Int) > 0 && y.(Int) > 0 && c > x.(Int) && c > y.(Int)) ||
				(x.(Int) < 0 && y.(Int) < 0 && c < x.(Int) && c < y.(Int)) ||
				(x.(Int) > 0 && y.(Int) < 0) ||
				(x.(Int) < 0 && y.(Int) > 0) {
				return c
			} else {
				return BigInt(*big.NewInt(0).Add(big.NewInt(int64(x.(Int))), big.NewInt(int64(y.(Int)))))
			}
		case BigInt:
			ay := big.Int(y.(BigInt))
			return BigInt(*big.NewInt(0).Add(big.NewInt(int64(x.(Int))), &ay))
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return narrow(BigRat(*big.NewRat(0, 1).Add(big.NewRat(int64(x.(Int)), 1), &ay)))
		case Float:
			return narrow(Float(x.(Int)) + y.(Float))
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(0.0).Add(big.NewFloat(float64(x.(Int))), &ay)))
		case Complex:
			return narrow(Complex(complex(float64(x.(Int)), 0)) + y.(Complex))
		case BigComplex:
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			return narrow(BigComplex{*big.NewFloat(0.0).Add(big.NewFloat(float64(x.(Int))), &ayr),
				y.(BigComplex).Im})
		default:
			panic("Plus: Int and non-numeric invalid")
		}

	case BigInt:
		switch y.(type) {
		case BigInt:
			ax := big.Int(x.(BigInt))
			ay := big.Int(y.(BigInt))
			return BigInt(*big.NewInt(0).Add(&ax, &ay))
		case BigRat:
			ax := big.Int(x.(BigInt))
			ay := big.Rat(y.(BigRat))
			return narrow(BigRat(*big.NewRat(0, 1).Add(big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)), &ay)))
		case Float:
			ax := big.Int(x.(BigInt))
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return narrow(Float(xf) + y.(Float))
		case BigFloat:
			ax := big.Int(x.(BigInt))
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			axf := big.Float(NewBigFloat(xf))
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(0).Add(&axf, &ay)))
		case Complex:
			ax := big.Int(x.(BigInt))
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return narrow(Complex(complex(xf, 0)) + y.(Complex))
		case BigComplex:
			ax := big.Int(x.(BigInt))
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			axf := big.Float(NewBigFloat(xf))
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			return narrow(BigComplex{*big.NewFloat(0).Add(&axf, &ayr), y.(BigComplex).Im})
		default:
			panic("Plus: BigInt and non-numeric invalid")
		}

	case BigRat:
		switch y.(type) {
		case BigRat:
			ax := big.Rat(x.(BigRat))
			ay := big.Rat(y.(BigRat))
			return narrow(BigRat(*big.NewRat(0, 1).Add(&ax, &ay)))
		case Float:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			return narrow(Float(xf) + y.(Float))
		case BigFloat:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			axf := big.Float(NewBigFloat(xf))
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(0).Add(&axf, &ay)))
		case Complex:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			return narrow(Complex(complex(xf, 0)) + y.(Complex))
		case BigComplex:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			axf := big.Float(NewBigFloat(xf))
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			return narrow(BigComplex{*big.NewFloat(0).Add(&axf, &ayr), y.(BigComplex).Im})
		default:
			panic("Plus: BigRat and non-numeric invalid")
		}

	case Float:
		switch y.(type) {
		case Float:
			return narrow(x.(Float) + y.(Float))
		case BigFloat:
			ax := big.Float(NewBigFloat(float64(x.(Float))))
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(0).Add(&ax, &ay)))
		case Complex:
			return narrow(Complex(complex(x.(Float), 0)) + y.(Complex))
		case BigComplex:
			ax := big.Float(NewBigFloat(float64(x.(Float))))
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			return narrow(BigComplex{*big.NewFloat(0).Add(&ax, &ayr), y.(BigComplex).Im})
		default:
			panic("Plus: Float and non-numeric invalid")
		}

	case BigFloat:
		switch y.(type) {
		case BigFloat:
			ax := big.Float(x.(BigFloat))
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(0).Add(&ax, &ay)))
		case Complex:
			ax := big.Float(x.(BigFloat))
			ayr := big.Float(NewBigFloat(real(complex128(y.(Complex)))))
			return narrow(BigComplex{*big.NewFloat(0).Add(&ax, &ayr),
				big.Float(NewBigFloat(imag(complex128(y.(Complex)))))})
		case BigComplex:
			ax := big.Float(x.(BigFloat))
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			return narrow(BigComplex{*big.NewFloat(0).Add(&ax, &ayr), y.(BigComplex).Im})
		default:
			panic("Plus: BigFloat and non-numeric invalid")
		}

	case Complex:
		switch y.(type) {
		case Complex:
			return narrow(x.(Complex) + y.(Complex))
		case BigComplex:
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			ayi := big.Float(BigFloat(y.(BigComplex).Im))
			axr := big.Float(NewBigFloat(real(complex128(x.(Complex)))))
			axi := big.Float(NewBigFloat(imag(complex128(x.(Complex)))))
			return narrow(BigComplex{*big.NewFloat(0).Add(&axr, &ayr), *big.NewFloat(0).Add(&axi, &ayi)})
		default:
			panic("Plus: Complex and non-numeric invalid")
		}

	case BigComplex:
		switch y.(type) {
		case BigComplex:
			axr := big.Float(BigFloat(x.(BigComplex).Re))
			axi := big.Float(BigFloat(x.(BigComplex).Im))
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			ayi := big.Float(BigFloat(y.(BigComplex).Im))
			return narrow(BigComplex{*big.NewFloat(0).Add(&axr, &ayr), *big.NewFloat(0).Add(&axi, &ayi)})
		default:
			panic("Plus: BigComplex and non-numeric invalid")
		}

	case time.Time:
		switch y.(type) {
		case Int:
			return x.(time.Time).Add(time.Duration(int64(time.Hour) * 24 * int64(y.(Int))))
		default:
			panic("Plus: Date and non-integer invalid")
		}

	//TODO: delete this special case for 2 strings when no longer needed for testing
	/*case string:
	  switch y.(type){
	  case string:
	    return x.(string) + y.(string)
	  default:
	    panic("Plus: string and non-string invalid")
	  }*/

	case u8.Codepoint:
		switch y.(type) {
		case u8.Codepoint:
			return u8.Text{x.(u8.Codepoint), y.(u8.Codepoint)}
		case u8.Text:
			return append(u8.Text{x.(u8.Codepoint)}, y.(u8.Text)...)
		default:
			panic("Plus: u8.Text or u8.Codepoint can only operate with u8.Text or u8.Codepoint")
		}

	case u8.Text:
		switch y.(type) {
		case u8.Codepoint:
			return append(x.(u8.Text), y.(u8.Codepoint))
		case u8.Text:
			return u8.Text(u8.Sur(x.(u8.Text)...) + u8.Sur(y.(u8.Text)...))
		default:
			panic("Plus: u8.Text or u8.Codepoint can only operate with u8.Text or u8.Codepoint")
		}

	//TODO:
	//case Slice:
	//return Slice(append([]interface{}(x.(Slice)), y))

	//TODO: for case Map, return a new map with all entries from old map added

	default:
		panic("Plus: two non-numerics invalid")
	}
}

/*
Negate returns the negative of a number.
An Int is promoted to BigInt if it overflows.
For Codepoint or CharClass, it returns the CharClass that matches the complement.
*/
func Negate(x interface{}) interface{} {
	x = widen(x)
	switch x.(type) {
	case nil:
		return nil
	case bool:
		return !x.(bool) //MAYBE TODO: remove this; throw panic instead ???
	case Int:
		if x == Int(-0x8000000000000000) {
			return BigInt(*big.NewInt(0).Neg(big.NewInt(int64(x.(Int)))))
		} else {
			return -x.(Int)
		}
	case BigInt:
		ax := big.Int(x.(BigInt))
		return BigInt(*big.NewInt(0).Neg(&ax))
	case BigRat:
		ax := big.Rat(x.(BigRat))
		return BigRat(*big.NewRat(0, 1).Neg(&ax))
	case Float:
		return Float(-x.(Float))
	case BigFloat:
		ax := big.Float(x.(BigFloat))
		return BigFloat(*big.NewFloat(0).Neg(&ax))
	case Complex:
		if x.(Complex) == Inf {
			return Inf
		} else {
			return Complex(-x.(Complex))
		}
	case BigComplex:
		axr := big.Float(BigFloat(x.(BigComplex).Re))
		axi := big.Float(BigFloat(x.(BigComplex).Im))
		return BigComplex{*big.NewFloat(0).Neg(&axr), *big.NewFloat(0).Neg(&axi)}

	//TODO: delete this when no longer needed for testing
	/*case u8.Text:
	  return u8.Surr(x.(u8.Text)) //turn Text into a string*/

	case u8.Codepoint:
		return CharClass("[^" + u8.Sur(x.(u8.Codepoint)) + "]")

	case CharClass:
		mid, neg := x.(CharClass).unwrap()
		if neg {
			return CharClass("[" + mid + "]")
		} else {
			return CharClass("[^" + mid + "]")
		}

	//TODO: logic for case Regex

	default:
		panic("Negate: non-numeric invalid")
	}
}

/*
Minus negates the 2nd arg and adds the two together.

//TODO: TEST BIGFLOAT
*/
func Minus(x, y interface{}) interface{} {
	return narrow(Plus(x, Negate(y)))
}

/*
Mult creates a new element being the product of the two arguments, for numeric types.
Two Ints are promoted to BigInt if it overflows.
For one argument of Codepoint or Text type, and the other of numeric type z,
returns Text repeated z times.
For one argument of CharClass or Regex type, and the other of numeric type z,
returns Regex matching repetition of z times.

//TODO: TEST BIGFLOAT
*/
func Mult(x, y interface{}) interface{} {
	x, y = widen(x), widen(y)
	if typeWeights[reflect.TypeOf(x)] > typeWeights[reflect.TypeOf(y)] {
		x, y = y, x
	}

	switch x.(type) {
	case u8.Codepoint, u8.Text, CharClass, Regex: //TODO: make sure this works correctly with typeWeights
		switch y.(type) {
		case PosIntRange:
			return RegexRepeat(x, y.(PosIntRange), true)
		case Int:
			return RegexRepeat(x, y.(Int), true)
		default:
			panic("Mult: Codepoint, Text, CharClass, or Regex multiplied with invalid value")
		}
		//TODO: For type Codepoint or Text, return repeated Text

	case PosIntRange: //TODO: make sure this works correctly with typeWeights
		switch y.(type) {
		case u8.Codepoint, u8.Text, CharClass, Regex:
			return RegexRepeat(y, x, false)
		default:
			panic("Mult: positive integer range multiplied with invalid value")
		}
		//TODO: For type Codepoint or Text, return repeated Text

	case nil:
		switch y.(type) {
		case nil:
			return nil
		case bool:
			return false
		case Int:
			return Int(0)
		case BigInt:
			return BigInt(*big.NewInt(0))
		case BigRat:
			return BigRat(*big.NewRat(0, 1))
		case Float:
			return Float(0)
		case BigFloat:
			return BigFloat(*big.NewFloat(0.0))
		case Complex:
			if cmplx.IsNaN(complex128(y.(Complex))) { //TODO: is this logic correct?
				return NaN
			} else {
				return Complex(0)
			}
		case BigComplex:
			return BigComplex{*big.NewFloat(0.0), *big.NewFloat(0.0)}
		default:
			panic("Mult: nil and non-numeric invalid")
		}

	case bool:
		switch y.(type) {
		case bool:
			return x.(bool) && y.(bool)
		case Int:
			if x.(bool) {
				return y.(Int)
			} else {
				return Int(0)
			}
		case BigInt:
			if x.(bool) {
				return y.(BigInt)
			} else {
				return BigInt(*big.NewInt(0))
			}
		case BigRat:
			if x.(bool) {
				return y.(BigRat)
			} else {
				return BigRat(*big.NewRat(0, 1)) //MAYBE TODO: should be BigInt ???
			}
		case Float:
			if x.(bool) {
				return y.(Float)
			} else {
				return Float(0)
			}
		case BigFloat:
			if x.(bool) {
				return y.(BigFloat)
			} else {
				return BigFloat(*big.NewFloat(0.0))
			}
		case Complex:
			if x.(bool) {
				return y.(Complex)
			} else {
				return Complex(0)
			}
		case BigComplex:
			if x.(bool) {
				return y.(BigComplex)
			} else {
				return BigComplex{*big.NewFloat(0.0), *big.NewFloat(0.0)}
			}
		default:
			panic("Mult: bool and non-numeric invalid")
		}

	case Int:
		switch y.(type) {
		case u8.Codepoint, u8.Text, CharClass, Regex: //TODO: make sure this works correctly with typeWeights
			return RegexRepeat(y, PosIntRange{x.(Int), x.(Int)}, false)

		case Int:
			c := x.(Int) * y.(Int)
			absc := Abs(c).(Int)
			if absc >= Abs(x).(Int) && absc >= Abs(y).(Int) {
				return c
			} else {
				return BigInt(*big.NewInt(0).Mul(big.NewInt(int64(x.(Int))), big.NewInt(int64(y.(Int)))))
			}
		case BigInt:
			ay := big.Int(y.(BigInt))
			return BigInt(*big.NewInt(0).Mul(big.NewInt(int64(x.(Int))), &ay))
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return narrow(BigRat(*big.NewRat(0, 1).Mul(big.NewRat(int64(x.(Int)), 1), &ay)))
		case Float:
			return narrow(Float(x.(Int)) * y.(Float))
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(0).Mul(big.NewFloat(float64(x.(Int))), &ay)))
		case Complex:
			return narrow(Complex(complex(float64(x.(Int)), 0)) * y.(Complex))
		case BigComplex:
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			ayi := big.Float(BigFloat(y.(BigComplex).Im))
			return narrow(BigComplex{*big.NewFloat(0).Mul(big.NewFloat(float64(x.(Int))), &ayr),
				*big.NewFloat(0).Mul(big.NewFloat(float64(x.(Int))), &ayi)})
		default:
			panic("Mult: Int and non-numeric invalid")
		}

	case BigInt:
		switch y.(type) {
		case BigInt:
			ax := big.Int(x.(BigInt))
			ay := big.Int(y.(BigInt))
			return BigInt(*big.NewInt(0).Mul(&ax, &ay))
		case BigRat:
			ax := big.Int(x.(BigInt))
			ay := big.Rat(y.(BigRat))
			return narrow(BigRat(*big.NewRat(0, 1).Mul(big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)), &ay)))
		case Float:
			ax := big.Int(x.(BigInt))
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return narrow(Float(xf) * y.(Float))
		case BigFloat:
			ax := big.Int(x.(BigInt))
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			axf := big.Float(NewBigFloat(xf))
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(0).Mul(&axf, &ay)))
		case Complex:
			ax := big.Int(x.(BigInt))
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return narrow(Complex(complex(xf, 0)) * y.(Complex))
		case BigComplex:
			ax := big.Int(x.(BigInt))
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			axf := big.Float(NewBigFloat(xf))
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			ayi := big.Float(BigFloat(y.(BigComplex).Im))
			return narrow(BigComplex{*big.NewFloat(0).Mul(&axf, &ayr), *big.NewFloat(0).Mul(&axf, &ayi)})
		default:
			panic("Mult: Int and non-numeric invalid")
		}

	case BigRat:
		switch y.(type) {
		case BigRat:
			ax := big.Rat(x.(BigRat))
			ay := big.Rat(y.(BigRat))
			return narrow(BigRat(*big.NewRat(0, 1).Mul(&ax, &ay)))
		case Float:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			return narrow(Float(xf) * y.(Float))
		case BigFloat:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			axf := big.Float(NewBigFloat(xf))
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(1.0).Mul(&axf, &ay)))
		case Complex:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			return narrow(Complex(complex(xf, 0)) * y.(Complex))
		case BigComplex:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			axf := big.Float(NewBigFloat(xf))
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			ayi := big.Float(BigFloat(y.(BigComplex).Im))
			return narrow(BigComplex{*big.NewFloat(1.0).Mul(&axf, &ayr), *big.NewFloat(1.0).Mul(&axf, &ayi)})
		default:
			panic("Mult: Rat and non-numeric invalid")
		}

	case Float:
		switch y.(type) {
		case Float:
			return narrow(x.(Float) * y.(Float))
		case BigFloat:
			ax := big.Float(NewBigFloat(float64(x.(Float))))
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(1.0).Mul(&ax, &ay)))
		case Complex:
			return narrow(Complex(complex(x.(Float), 0)) * y.(Complex))
		case BigComplex:
			ax := big.Float(NewBigFloat(float64(x.(Float))))
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			ayi := big.Float(BigFloat(y.(BigComplex).Im))
			return narrow(BigComplex{*big.NewFloat(1.0).Mul(&ax, &ayr), *big.NewFloat(1.0).Mul(&ax, &ayi)})
		default:
			panic("Mult: Float and non-numeric invalid")
		}

	case BigFloat:
		switch y.(type) {
		case BigFloat:
			ax := big.Float(x.(BigFloat))
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(1.0).Mul(&ax, &ay)))
		case Complex:
			ax := big.Float(x.(BigFloat))
			ayr := big.Float(NewBigFloat(real(complex128(y.(Complex)))))
			ayi := big.Float(NewBigFloat(imag(complex128(y.(Complex)))))
			return narrow(BigComplex{*big.NewFloat(1.0).Mul(&ax, &ayr), *big.NewFloat(1.0).Mul(&ax, &ayi)})
		case BigComplex:
			ax := big.Float(x.(BigFloat))
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			ayi := big.Float(BigFloat(y.(BigComplex).Im))
			return narrow(BigComplex{*big.NewFloat(1.0).Mul(&ax, &ayr),
				*big.NewFloat(1.0).Mul(&ax, &ayi)})
		default:
			panic("Mult: BigFloat and non-numeric invalid")
		}

	case Complex:
		switch y.(type) {
		case Complex:
			return narrow(x.(Complex) * y.(Complex))
		case BigComplex:
			axr := big.Float(NewBigFloat(real(complex128(x.(Complex)))))
			axi := big.Float(NewBigFloat(imag(complex128(x.(Complex)))))
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			ayi := big.Float(BigFloat(y.(BigComplex).Im))
			realPart := *big.NewFloat(1.0).Sub(big.NewFloat(1.0).Mul(&axr, &ayr), big.NewFloat(1.0).Mul(&axi, &ayi))
			imagPart := *big.NewFloat(1.0).Add(big.NewFloat(1.0).Mul(&axi, &ayr), big.NewFloat(1.0).Mul(&axr, &ayi))
			return narrow(BigComplex{realPart, imagPart})
		default:
			panic("Mult: Complex and non-numeric invalid")
		}

	case BigComplex:
		switch y.(type) {
		case BigComplex:
			axr := big.Float(BigFloat(x.(BigComplex).Re))
			axi := big.Float(BigFloat(x.(BigComplex).Im))
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			ayi := big.Float(BigFloat(y.(BigComplex).Im))
			realPart := *big.NewFloat(1.0).Sub(big.NewFloat(1.0).Mul(&axr, &ayr), big.NewFloat(1.0).Mul(&axi, &ayi))
			imagPart := *big.NewFloat(1.0).Add(big.NewFloat(1.0).Mul(&axi, &ayr), big.NewFloat(1.0).Mul(&axr, &ayi))
			return narrow(BigComplex{realPart, imagPart})
		default:
			panic("Mult: BigComplex and non-numeric invalid")
		}

	default:
		panic("Mult: two non-numerics invalid")
	}
}

/*
Invert returns the result of dividing the numeric into 1.
Division by 0 returns infinity.
*/
func Invert(x interface{}) interface{} {
	x = widen(x)
	switch x.(type) {
	case nil:
		return NaN
	case bool:
		if x.(bool) {
			return true
		} else {
			return NaN
		}
	case Int:
		if x.(Int) == 0 { //MAYBE TODO: check for Inf so we can generate 0; also in BigInt, BigRat, Float, and Complex ???
			return Inf
		}
		return BigRat(*big.NewRat(1, int64(x.(Int))))
	case BigInt:
		if reflect.DeepEqual(x.(BigInt), BigInt(*big.NewInt(0))) {
			return Inf
		}
		ax := big.Int(x.(BigInt))
		return BigRat(*big.NewRat(0, 1).SetFrac(big.NewInt(1), &ax)) //FIX: make copy of denom first ???
	case BigRat:
		if reflect.DeepEqual(x.(BigRat), BigRat(*big.NewRat(0, 1))) {
			return Inf
		}
		ax := big.Rat(x.(BigRat))
		return BigRat(*big.NewRat(0, 1).Inv(&ax))
	case Float:
		if x.(Float) == 0 {
			return Inf
		}
		return Float(1 / float64(x.(Float)))
	case BigFloat:
		ax := big.Float(x.(BigFloat))
		if big.NewFloat(0.0).Cmp(&ax) == 0 {
			return Inf
		}
		return BigFloat(*big.NewFloat(1.0).Quo(big.NewFloat(1.0), &ax))
	case Complex:
		if x.(Complex) == 0 {
			return Inf
		}
		return Complex(1 / complex128(x.(Complex)))
	case BigComplex:
		axr := big.Float(BigFloat(x.(BigComplex).Re))
		axi := big.Float(BigFloat(x.(BigComplex).Im))
		if big.NewFloat(0.0).Cmp(&axr) == 0 && big.NewFloat(0.0).Cmp(&axi) == 0 {
			return Inf
		}
		return NewBigComplex(BigFloat(*big.NewFloat(1.0).Quo(big.NewFloat(1.0), &axr)),
			BigFloat(*big.NewFloat(1.0).Quo(big.NewFloat(1.0), &axi)))
	default:
		panic("Invert: non-numeric invalid")
	}
}

/*
Divide inverts the seconnd argument and multiplies the two together.
Division by nil or false gives infinity.
Divide result is of the same type as whichever argument is higher in the numeric hierarchy.
A division by a Int or BigInt promotes to a BigRat.

//TODO: TEST BIGFLOAT
*/
func Divide(x, y interface{}) interface{} {
	//TODO: check for 0/0 so we can generate NaN; ditto Inf/Inf

	return narrow(Mult(x, Invert(y)))
}

/*
Abs returns the absolute value of the numeric argument.
*/
func Abs(x interface{}) interface{} {
	x = widen(x)
	switch x.(type) {
	case nil:
		return nil
	case bool:
		return x.(bool)
	case Int:
		if x.(Int) < 0 {
			return -x.(Int)
		} else {
			return x.(Int)
		}
	case BigInt:
		ax := big.Int(x.(BigInt))
		return big.NewInt(0).Abs(&ax)
	case BigRat:
		ax := big.Rat(x.(BigRat))
		return big.NewRat(0, 1).Abs(&ax)
	case Float:
		return Float(math.Abs(float64(x.(Float))))
	case BigFloat:
		ax := big.Float(x.(BigFloat))
		return BigFloat(*big.NewFloat(0.0).Abs(&ax))
	case Complex:
		return Float(cmplx.Abs(complex128(x.(Complex))))
	case BigComplex:
		axr := big.Float(BigFloat(x.(BigComplex).Re))
		axi := big.Float(BigFloat(x.(BigComplex).Im))
		az, _ := big.NewFloat(0.0).Add(big.NewFloat(0.0).Mul(&axr, &axr), big.NewFloat(0.0).Mul(&axi, &axi)).Float64()
		return Float(math.Sqrt(az))
	default:
		//MAYBE TODO: any logic for arbitrary object ???
		panic("Abs: Invalid numeric type")
	}
}

/*
Power returns a result of the same type as the argument higher in the numeric hierarchy.
Raising to the power of a nil or false generates a panic;
an Int to the power of a Int is a BigInt;
an Int or BigInt to the power of a BigRat is a Float;
a BigRat to the power of an Int, BigInt, or BigRat is a Float.

//WARNING: This function is alpha version only.

//TODO: TEST IT; ADD CODE AND TEST FOR BigFloat AND BigComplex; change output to []interface{} to cater for multiple results
*/
func Power(x, y interface{}) interface{} {
	x, y = widen(x), widen(y)
	switch x.(type) {
	case Int:
		switch y.(type) {
		case bool:
			if y.(bool) {
				return x.(Int)
			} else {
				panic("Power: only Int, Int, Rat, Float, BigFloat, Complex, and BigComplex args valid")
			}
		case Int:
			return BigInt(*big.NewInt(0).Exp(big.NewInt(int64(x.(Int))), big.NewInt(int64(y.(Int))), nil))
		case BigInt:
			ay := big.Int(y.(BigInt))
			return BigInt(*big.NewInt(0).Exp(big.NewInt(int64(x.(Int))), &ay, nil))
		case BigRat:
			xf, _ := big.NewRat(int64(x.(Int)), 1).Float64()
			ay := big.Rat(y.(BigRat))
			yf, _ := ay.Float64()
			return math.Pow(xf, yf)
		case Float:
			return math.Pow(float64(x.(Int)), float64(y.(Float)))
		case Complex:
			return cmplx.Pow(complex(float64(x.(Int)), 0), complex128(y.(Complex)))
		default:
			panic("Power: only Int, Int, Rat, Float, and Complex args valid")
		}

	case BigInt:
		switch y.(type) {
		case bool:
			if y.(bool) {
				return x.(BigInt)
			} else {
				panic("Power: only Int, Int, Rat, Float, and Complex args valid")
			}
		case Int:
			ay := big.Int(y.(BigInt))
			return BigInt(*big.NewInt(0).Exp(&ay, big.NewInt(int64(x.(Int))), nil))
		case BigInt:
			ax := big.Int(x.(BigInt))
			ay := big.Int(y.(BigInt))
			return BigInt(*big.NewInt(0).Exp(&ax, &ay, nil))
		case BigRat:
			ax := big.Int(x.(BigInt))
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			ay := big.Rat(y.(BigRat))
			yf, _ := ay.Float64()
			return math.Pow(xf, yf)
		case Float:
			ax := big.Int(x.(BigInt))
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return math.Pow(xf, float64(y.(Float)))
		case Complex:
			ax := big.Int(x.(BigInt))
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return cmplx.Pow(complex(xf, 0), complex128(y.(Complex)))
		default:
			panic("Power: only Int, Int, Rat, Float, and Complex args valid")
		}

	case BigRat:
		switch y.(type) {
		case bool:
			if y.(bool) {
				return x.(BigRat)
			} else {
				panic("Power: only Int, Int, Rat, Float, and Complex args valid")
			}
		case Int:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			yf, _ := big.NewRat(int64(y.(Int)), 1).Float64()
			return math.Pow(xf, yf)
		case BigInt:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			ay := big.Int(y.(BigInt))
			yf, _ := big.NewRat(0, 1).SetFrac(&ay, big.NewInt(1)).Float64()
			return math.Pow(xf, yf)
		case BigRat:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			ay := big.Rat(y.(BigRat))
			yf, _ := ay.Float64()
			return math.Pow(xf, yf)
		case Float:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			return math.Pow(xf, float64(y.(Float)))
		case Complex:
			ax := big.Rat(x.(BigRat))
			xf, _ := ax.Float64()
			return cmplx.Pow(complex(xf, 0), complex128(y.(Complex)))
		default:
			panic("Power: only Int, Int, Rat, Float, and Complex args valid")
		}

	case Float:
		switch y.(type) {
		case bool:
			if y.(bool) {
				return x.(Float)
			} else {
				panic("Power: only Int, Int, Rat, Float, and Complex args valid")
			}
		case Int:
			return math.Pow(float64(x.(Float)), float64(y.(Int)))
		case BigInt:
			ay := big.Int(y.(BigInt))
			yf, _ := big.NewRat(0, 1).SetFrac(&ay, big.NewInt(1)).Float64()
			return math.Pow(float64(x.(Float)), yf)
		case BigRat:
			ay := big.Rat(y.(BigRat))
			yf, _ := ay.Float64()
			return math.Pow(float64(x.(Float)), yf)
		case Float:
			return math.Pow(float64(x.(Float)), float64(y.(Float)))
		case Complex:
			return cmplx.Pow(complex(x.(Float), 0), complex128(y.(Complex)))
		default:
			panic("Power: only Int, Int, Rat, Float, and Complex args valid")
		}

	case Complex:
		switch y.(type) {
		case bool:
			if y.(bool) {
				return x.(Complex)
			} else {
				panic("Power: only Int, Int, Rat, Float, and Complex args valid")
			}
		case Int:
			return cmplx.Pow(complex128(x.(Complex)), complex(float64(y.(Int)), 0))
		case BigInt:
			ay := big.Int(y.(BigInt))
			yf, _ := big.NewRat(0, 1).SetFrac(&ay, big.NewInt(1)).Float64()
			return cmplx.Pow(complex128(x.(Complex)), complex(yf, 0))
		case BigRat:
			ay := big.Rat(y.(BigRat))
			yf, _ := ay.Float64()
			return cmplx.Pow(complex128(x.(Complex)), complex(yf, 0))
		case Float:
			return cmplx.Pow(complex128(x.(Complex)), complex(y.(Float), 0))
		case Complex:
			return cmplx.Pow(complex128(x.(Complex)), complex128(y.(Complex)))
		default:
			panic("Power: only Int, Int, Rat, Float, and Complex args valid")
		}

	default:
		panic("Power: only Int, Int, Rat, Float, and Complex args valid")
	}
}
