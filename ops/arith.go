// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package ops

import (
	"fmt"
	"math"
	"math/big"
	"math/cmplx"
	"reflect"
	"time"

	"github.com/grolang/gro/parsec"
	u8 "github.com/grolang/gro/utf88"
)

/*
We need to define BigRat as a separate type so we can define
a (BigRat)String() method on the contents of big.Rat,
because only the method on the pointer (*big.Rat)String() exists.
Ditto, big.Int and big.Float.
*/
type (
	Int        = int64
	BigInt     big.Int
	BigRat     big.Rat
	Float      = float64
	BigFloat   big.Float
	Complex    = complex128
	BigComplex struct{ Re, Im big.Float }
	Infinity   struct{} //Riemann-style infinity
)

var Inf = Infinity{}

var typeWeights = map[reflect.Type]int{
	reflect.TypeOf(nil):                                                1,
	reflect.TypeOf(false):                                              2,
	reflect.TypeOf(Int(0)):                                             3,
	reflect.TypeOf(BigInt(*big.NewInt(0))):                             4,
	reflect.TypeOf(BigRat(*big.NewRat(0, 1))):                          5,
	reflect.TypeOf(Float(0.0)):                                         6,
	reflect.TypeOf(BigFloat(*big.NewFloat(0.0))):                       7,
	reflect.TypeOf(Complex(0 + 0i)):                                    8,
	reflect.TypeOf(BigComplex{*big.NewFloat(0.0), *big.NewFloat(0.0)}): 9,
	reflect.TypeOf(Inf):                                                10,
}

func numericTypes(x, y interface{}) bool {
	return typeWeights[reflect.TypeOf(x)] != 0 && typeWeights[reflect.TypeOf(y)] != 0 &&
		typeWeights[reflect.TypeOf(x)] > typeWeights[reflect.TypeOf(y)]
}

//==============================================================================

//String of BigInt returns a string representation of its value.
func (n BigInt) String() string {
	nr := big.Int(n)
	return nr.String()
}

//String of BigRat returns a string representation of its value.
func (n BigRat) String() string {
	nr := big.Rat(n)
	return nr.String()
}

//String of BigFloat returns a string representation of its value.
func (n BigFloat) String() string {
	x := big.Float(n)
	return x.String()
}

//String of BigComplex returns a string representation of its value.
func (n BigComplex) String() string {
	r := big.Float(n.Re)
	i := big.Float(n.Im)
	return r.String() + "+" + i.String() + "i"
}

//String of Infinity returns a string representation of its value.
func (n Infinity) String() string {
	return "inf"
}

//String of Slice returns a string representation of its value.
func (s Slice) String() string {
	str := "{"
	for i, v := range s {
		if i > 0 {
			str = str + ", "
		}
		str = str + fmt.Sprintf("%v", v)
	}
	return str + "}"
}

//==============================================================================
/*
narrow returns Inf if x is infinite, nil if x is NaN, otherwise converts
the type of x to its simplest possible type in the numeric hierarchy.
x is returned unchanged if it's not in the hierarchy.
*/
func narrow(x interface{}) interface{} {
	switch x.(type) {

	case nil, bool, Int, Infinity:
		return x

	case BigInt:
		//TODO: narrow BigInt to Int for low BigInt
		return x

	case BigRat:
		xr := big.Rat(x.(BigRat))
		if reflect.DeepEqual(*xr.Denom(), *big.NewInt(1)) {
			return BigInt(*xr.Num())
		} else {
			return x
		}

	case Float:
		//TODO: narrow Float to Int if nothing after decimal point
		xf := float64(x.(Float))
		if math.IsInf(xf, 0) {
			return Inf
		} else if math.IsNaN(xf) {
			return nil
		} else {
			return x
		}

	case BigFloat:
		//TODO: narrow BigFloat to BigInt if nothing after decimal point
		xf := big.Float(x.(BigFloat))
		if xf.IsInf() {
			return Inf
		} else {
			return x
		}

	case Complex:
		//TODO: narrow re+0.0i etc to narrower 0
		xc := complex128(x.(Complex))
		if cmplx.IsInf(xc) {
			return Inf
		} else if cmplx.IsNaN(xc) {
			return nil
		} else {
			return x
		}

	case BigComplex:
		//TODO: narrow BigComplex re+0.0i etc to narrower 0
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
widen converts values to types in the numeric hierarchy,
then narrows the type if possible.
*/
func widen(x interface{}) interface{} {
	switch x.(type) {
	case nil, bool:
		return x

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
			a := big.NewInt(int64(x.(uint64) - (1 << 63)))
			b := big.NewInt(1 << 62)
			return BigInt(*a.Add(a, b).Add(a, b))
		}

	case uint:
		if uint64(x.(uint)) < (1 << 63) {
			return Int(x.(uint))
		} else {
			a := big.NewInt(int64(x.(uint)) - (1 << 62) - (1 << 62))
			b := big.NewInt(1 << 62)
			return BigInt(*a.Add(a, b).Add(a, b))
		}

	case float32:
		if math.IsInf(float64(x.(float32)), 0) {
			return Inf
		} else {
			return narrow(Float(x.(float32)))
		}
	case float64:
		if math.IsInf(x.(float64), 0) {
			return Inf
		} else {
			return narrow(Float(x.(float64)))
		}

	case complex64:
		if cmplx.IsInf(complex128(x.(complex64))) {
			return Inf
		} else {
			return narrow(Complex(x.(complex64)))
		}
	case complex128:
		if cmplx.IsInf(x.(complex128)) {
			return Inf
		} else {
			return narrow(Complex(x.(complex128)))
		}

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

	case u8.Text, Slice, Map:
		return x

	default:
		switch reflect.TypeOf(x).Kind() {
		case reflect.Slice:
			vs := []interface{}{}
			s := reflect.ValueOf(x)
			for i := 0; i < s.Len(); i++ {
				vs = append(vs, widen(s.Index(i).Interface()))
			}
			return Slice(vs)

		case reflect.Map:
			es := map[interface{}]interface{}{}
			ss := []interface{}{}
			s := reflect.ValueOf(x)
			for _, k := range s.MapKeys() {
				kk := widen(k.Interface())
				vv := widen(s.MapIndex(k).Interface())
				es[kk] = vv
				ss = append(ss, kk) //####
			}
			return Map{lkp: es, seq: ss}

		default:
			return x //should panic ???
		}
		return x //should panic ???
	}
}

//==============================================================================

/*
ToBool converts nil, false, zeroes, the empty string or utf88.Text to false,
and all other values to true.
*/
func ToBool(x interface{}) bool {
	x = widen(x)
	switch x.(type) {
	case nil:
		return false
	case bool:
		return x.(bool)
	case Int:
		return x.(Int) != 0
	case BigInt:
		ax := big.Int(x.(BigInt))
		return big.NewInt(0).Cmp(&ax) != 0
	case BigRat:
		ax := big.Rat(x.(BigRat))
		return big.NewRat(0, 1).Cmp(&ax) != 0
	case Float:
		return x.(Float) != 0
	case BigFloat:
		ax := big.Float(x.(BigFloat))
		return big.NewFloat(0.0).Cmp(&ax) != 0
	case Complex:
		return x.(Complex) != 0
	case BigComplex:
		axx := big.Float(BigFloat(x.(BigComplex).Re))
		axy := big.Float(BigFloat(x.(BigComplex).Im))
		return big.NewFloat(0.0).Cmp(&axx) != 0 || big.NewFloat(0.0).Cmp(&axy) != 0
	case Infinity:
		return true

	case u8.Text:
		if len(x.(u8.Text)) == 0 {
			return false
		} else {
			return true
		}

	//MAYBE TODO: Codepoint, CharClass, Regex ???

	default:
		return true
	}
}

func ToInt(r interface{}) Int {
	switch r.(type) {
	case Int:
		if r.(Int) < 0 {
			return Int(0)
		} else {
			return r.(Int)
		}
	case BigInt:
		ar := big.Int(r.(BigInt))
		if big.NewInt(0).Cmp(&ar) == 1 {
			return Int(0)
		} else if big.NewInt(0x7fffffffffffffff).Cmp(&ar) == -1 {
			return Int(0x7fffffffffffffff)
		} else {
			return Int(ar.Int64())
		}
	case Infinity:
		return Int(0x7fffffffffffffff)
	default:
		panic(fmt.Sprintf("toInt: invalid input type %T", r))
	}
}

//==============================================================================

/*
Identity converts:
  float32 and float64 to Float (float64), but NaN is converted to nil
  complex64 and complex128 to Complex (complex128), but NaN converted to nil
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
	if numericTypes(x, y) {
		x, y = y, x
	}

	switch x.(type) {
	case nil:
		switch y.(type) {
		case nil:
			return true
		case bool:
			return !y.(bool)
		case Int, BigInt, BigRat, Float, BigFloat, Complex, BigComplex:
			return !ToBool(x)
		case Infinity:
			return false
		default:
			panic("IsEqual: nil and non-numeric invalid")
		}

	case bool:
		switch y.(type) {
		case bool:
			return x.(bool) == y.(bool)
		case Int, BigInt, BigRat, Float, BigFloat, Complex, BigComplex:
			return x.(bool) == ToBool(y)
		case Infinity:
			return x.(bool)
		default:
			panic("IsEqual: bool and non-numeric invalid")
		}

	case Int:
		ax := x.(Int)
		switch y.(type) {
		case Int:
			return ax == y.(Int)
		case BigInt:
			ay := big.Int(y.(BigInt))
			return big.NewInt(int64(ax)).Cmp(&ay) == 0
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return big.NewRat(int64(ax), 1).Cmp(&ay) == 0
		case Float:
			return Float(ax) == y.(Float)
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return big.NewFloat(float64(ax)).Cmp(&ay) == 0
		case Complex:
			return Complex(complex(float64(ax), 0)) == y.(Complex)
		case BigComplex:
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			return big.NewFloat(float64(ax)).Cmp(&yr) == 0 && big.NewFloat(0.0).Cmp(&yi) == 0
		case Infinity:
			return false
		default:
			panic("IsEqual: Int and non-numeric invalid")
		}

	case BigInt:
		ax := big.Int(x.(BigInt))
		switch y.(type) {
		case BigInt:
			ay := big.Int(y.(BigInt))
			return ax.Cmp(&ay) == 0
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Cmp(&ay) == 0
		case Float:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return Float(xf) == y.(Float)
		case BigFloat:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			ay := big.Float(y.(BigFloat))
			return big.NewFloat(xf).Cmp(&ay) == 0
		case Complex:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return Complex(complex(xf, 0)) == y.(Complex)
		case BigComplex:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			return big.NewFloat(xf).Cmp(&yr) == 0 && big.NewFloat(0.0).Cmp(&yi) == 0
		case Infinity:
			return false
		default:
			panic("IsEqual: BigInt and non-numeric invalid")
		}

	case BigRat:
		ax := big.Rat(x.(BigRat))
		switch y.(type) {
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return ax.Cmp(&ay) == 0
		case Float:
			xf, _ := ax.Float64()
			return Float(xf) == y.(Float)
		case BigFloat:
			xf, _ := ax.Float64()
			ay := big.Float(y.(BigFloat))
			return big.NewFloat(xf).Cmp(&ay) == 0
		case Complex:
			xf, _ := ax.Float64()
			return Complex(complex(xf, 0)) == y.(Complex)
		case BigComplex:
			xf, _ := ax.Float64()
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			return big.NewFloat(xf).Cmp(&yr) == 0 && big.NewFloat(0.0).Cmp(&yi) == 0
		case Infinity:
			return false
		default:
			panic("IsEqual: BigRat and non-numeric invalid")
		}

	case Float:
		ax := x.(Float)
		switch y.(type) {
		case Float:
			return ax == y.(Float)
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return big.NewFloat(float64(ax)).Cmp(&ay) == 0
		case Complex:
			return Complex(complex(ax, 0)) == y.(Complex)
		case BigComplex:
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			return big.NewFloat(float64(ax)).Cmp(&yr) == 0 && big.NewFloat(0.0).Cmp(&yi) == 0
		case Infinity:
			return false
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
			ay := *big.NewFloat(real(y.(Complex)))
			return ax.Cmp(&ay) == 0 && imag(y.(Complex)) == 0
		case BigComplex:
			ax := big.Float(x.(BigFloat))
			ay := *big.NewFloat(0)
			yr := y.(BigComplex).Re
			yi := y.(BigComplex).Im
			return (&yr).Cmp(&ax) == 0 && (&yi).Cmp(&ay) == 0
		case Infinity:
			return false
		default:
			panic("IsEqual: BigFloat and non-numeric invalid")
		}

	case Complex:
		ax := x.(Complex)
		switch y.(type) {
		case Complex:
			if cmplx.IsNaN(complex128(ax)) && cmplx.IsNaN(complex128(y.(Complex))) {
				return true
			}
			return ax == y.(Complex)
		case BigComplex:
			xr := *big.NewFloat(real(complex128(ax)))
			xi := *big.NewFloat(imag(complex128(ax)))
			yr := y.(BigComplex).Re
			yi := y.(BigComplex).Im
			return (&yr).Cmp(&xr) == 0 && (&yi).Cmp(&xi) == 0
		case Infinity:
			return false
		default:
			panic("IsEqual: Complex and non-numeric invalid")
		}

	case BigComplex:
		switch y.(type) {
		case BigComplex:
			xr := x.(BigComplex).Re
			xi := x.(BigComplex).Im
			yr := y.(BigComplex).Re
			yi := y.(BigComplex).Im
			return (&yr).Cmp(&xr) == 0 && (&yi).Cmp(&xi) == 0
		case Infinity:
			return false
		default:
			panic("IsEqual: BigComplex and non-numeric invalid")
		}

	case Infinity:
		switch y.(type) {
		case Infinity:
			return true
		default:
			panic("IsEqual: Infinity and non-numeric invalid")
		}

	case time.Time:
		switch y.(type) {
		case time.Time:
			return x.(time.Time).Equal(y.(time.Time))
		default:
			panic("IsEqual: Date and non-Date invalid")
		}

	//TODO: add u8.Codepoint cases

	case u8.Text:
		switch y.(type) {
		case u8.Text:
			return string(u8.SurrogatePoints(x.(u8.Text))) == string(u8.SurrogatePoints(y.(u8.Text)))
		default:
			panic("IsEqual: u8.Text can only operate with u8.Text")
		}

	case Map:
		switch y.(type) {
		case Map:
			return reflect.DeepEqual(x.(Map).lkp, y.(Map).lkp)
		default:
			panic("IsEqual: Map can only operate with Map")
		}

	default:
		return reflect.DeepEqual(x, y)
	}
}

/*
IsLessThan returns true if both arguments are non-complex numeric or Dates,
and x is less than y.
It panics if either arg is nil, bool, Complex, BigComplex, Infinity, or non-numeric.
*/
func IsLessThan(x, y interface{}) bool {
	x, y = widen(x), widen(y)
	if numericTypes(x, y) {
		return !IsLessThan(y, x) && !IsEqual(x, y)
	}

	//MAYBE TODO: nil and boolean and Infinity ???

	switch x.(type) {
	case Int:
		ax := x.(Int)
		switch y.(type) {
		case Int:
			return ax < y.(Int)
		case BigInt:
			ay := big.Int(y.(BigInt))
			return big.NewInt(int64(ax)).Cmp(&ay) == -1
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return big.NewRat(int64(ax), 1).Cmp(&ay) == -1
		case Float:
			return Float(ax) < y.(Float)
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return big.NewFloat(float64(ax)).Cmp(&ay) == -1
		default:
			panic("IsLessThan: only Int, BigInt, Rat, Float, and BigFloat args valid")
		}

	case BigInt:
		ax := big.Int(x.(BigInt))
		switch y.(type) {
		case BigInt:
			ay := big.Int(y.(BigInt))
			return ax.Cmp(&ay) == -1
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Cmp(&ay) == -1
		case Float:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return Float(xf) < y.(Float)
		case BigFloat:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			ay := big.Float(y.(BigFloat))
			return big.NewFloat(xf).Cmp(&ay) == -1
		default:
			panic("IsLessThan: only Int, BigInt, Rat, Float, and BigFloat args valid")
		}

	case BigRat:
		ax := big.Rat(x.(BigRat))
		switch y.(type) {
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return ax.Cmp(&ay) == -1
		case Float:
			xf, _ := ax.Float64()
			return Float(xf) < y.(Float)
		case BigFloat:
			xf, _ := ax.Float64()
			ay := big.Float(y.(BigFloat))
			return big.NewFloat(xf).Cmp(&ay) == -1
		default:
			panic("IsLessThan: only Int, BigInt, Rat, Float, and BigFloat args valid")
		}

	case Float:
		ax := x.(Float)
		switch y.(type) {
		case Float:
			return ax < y.(Float)
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return big.NewFloat(float64(ax)).Cmp(&ay) == -1
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

	//TODO: add u8.Codepoint and u8.Text cases

	default:
		panic("IsLessThan: only Int, BigInt, Rat, Float, BigFloat, and Date args valid")
	}
}

/*
IsNotEqual returns true if both arguments are non-complex numeric or Dates,
and x is not equal to y.
It panics if either arg is nil, bool, Complex, BigComplex, or non-numeric.
*/
func IsNotEqual(x, y interface{}) bool {
	return !IsEqual(x, y)
}

/*
IsLessOrEqual returns true if both arguments are non-complex numeric or Dates,
and x is less than or equal to y.
It panics if either arg is nil, bool, Complex, BigComplex, or non-numeric.
*/
func IsLessOrEqual(x, y interface{}) bool {
	return IsLessThan(x, y) || IsEqual(x, y)
}

/*
IsGreaterThan returns true if both arguments are non-complex numeric or Dates,
and x is greater than y.
It panics if either arg is nil, bool, Complex, BigComplex, or non-numeric.
*/
func IsGreaterThan(x, y interface{}) bool {
	return !IsLessOrEqual(x, y)
}

/*
IsGreaterOrEqual returns true if both arguments are non-complex numeric or Dates,
and x is greater than or equal to y.
It panics if either arg is nil, bool, Complex, BigComplex, or non-numeric.
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
	if numericTypes(x, y) {
		x, y = y, x
	}

	switch x.(type) {
	case nil:
		switch y.(type) {
		case nil, bool, Int, BigInt, BigRat, Float, BigFloat, Complex, BigComplex, Infinity:
			return nil
		default:
			panic("Plus: nil and non-numeric invalid")
		}

	case bool: //false is 0, true is 1
		ax := x.(bool)
		switch y.(type) {
		case bool:
			return ax || y.(bool)
		case Int:
			if ax {
				return 1 + y.(Int)
			} else {
				return y.(Int)
			}
		case BigInt:
			if ax {
				ay := big.Int(y.(BigInt))
				return BigInt(*big.NewInt(0).Add(big.NewInt(1), &ay))
			} else {
				return y.(BigInt)
			}
		case BigRat:
			if ax {
				ay := big.Rat(y.(BigRat))
				return BigRat(*big.NewRat(0, 1).Add(big.NewRat(1, 1), &ay))
			} else {
				return y.(BigRat)
			}
		case Float:
			if ax {
				return 1.0 + y.(Float)
			} else {
				return y.(Float)
			}
		case BigFloat:
			if ax {
				ay := big.Float(y.(BigFloat))
				return BigFloat(*big.NewFloat(0.0).Add(big.NewFloat(1.0), &ay))
			} else {
				return y.(BigFloat)
			}
		case Complex:
			if ax {
				return 1 + y.(Complex)
			} else {
				return y.(Complex)
			}
		case BigComplex:
			if ax {
				yr := big.Float(BigFloat(y.(BigComplex).Re))
				yrplus1 := *big.NewFloat(0.0).Add(big.NewFloat(1.0), &yr)
				return BigComplex{yrplus1, y.(BigComplex).Im}
			} else {
				return y.(BigComplex)
			}
		case Infinity:
			return y
		default:
			panic("Plus: bool and non-numeric invalid")
		}

	case Int:
		ax := x.(Int)
		switch y.(type) {
		case Int:
			ay := y.(Int)
			c := ax + ay
			//check for overflow...
			if (ax > 0 && ay > 0 && c > ax && c > ay) ||
				(ax < 0 && ay < 0 && c < ax && c < ay) ||
				(ax > 0 && ay < 0) ||
				(ax < 0 && ay > 0) {
				return c
			} else {
				return BigInt(*big.NewInt(0).Add(big.NewInt(int64(ax)), big.NewInt(int64(ay))))
			}
		case BigInt:
			ay := big.Int(y.(BigInt))
			return BigInt(*big.NewInt(0).Add(big.NewInt(int64(ax)), &ay))
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return narrow(BigRat(*big.NewRat(0, 1).Add(big.NewRat(int64(ax), 1), &ay)))
		case Float:
			return narrow(Float(ax) + y.(Float))
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(0.0).Add(big.NewFloat(float64(ax)), &ay)))
		case Complex:
			return narrow(Complex(complex(float64(ax), 0)) + y.(Complex))
		case BigComplex:
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			sr := *big.NewFloat(0.0).Add(big.NewFloat(float64(ax)), &yr)
			return narrow(BigComplex{sr, y.(BigComplex).Im})
		case Infinity:
			return y
		case time.Time:
			return y.(time.Time).Add(time.Duration(int64(time.Hour) * 24 * int64(ax)))
		default:
			panic("Plus: Int and non-numeric invalid")
		}

	case BigInt:
		ax := big.Int(x.(BigInt))
		switch y.(type) {
		case BigInt:
			ay := big.Int(y.(BigInt))
			return BigInt(*big.NewInt(0).Add(&ax, &ay))
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return narrow(BigRat(*big.NewRat(0, 1).Add(big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)), &ay)))
		case Float:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return narrow(Float(xf) + y.(Float))
		case BigFloat:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			axf := *big.NewFloat(xf)
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(0).Add(&axf, &ay)))
		case Complex:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return narrow(Complex(complex(xf, 0)) + y.(Complex))
		case BigComplex:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			xr := *big.NewFloat(xf)
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			return narrow(BigComplex{*big.NewFloat(0).Add(&xr, &yr), y.(BigComplex).Im})
		case Infinity:
			return y
		default:
			panic("Plus: BigInt and non-numeric invalid")
		}

	case BigRat:
		ax := big.Rat(x.(BigRat))
		switch y.(type) {
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return narrow(BigRat(*big.NewRat(0, 1).Add(&ax, &ay)))
		case Float:
			xf, _ := ax.Float64()
			return narrow(Float(xf) + y.(Float))
		case BigFloat:
			xf, _ := ax.Float64()
			xr := *big.NewFloat(xf)
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(0).Add(&xr, &ay)))
		case Complex:
			xf, _ := ax.Float64()
			return narrow(Complex(complex(xf, 0)) + y.(Complex))
		case BigComplex:
			xf, _ := ax.Float64()
			xr := *big.NewFloat(xf)
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			return narrow(BigComplex{*big.NewFloat(0).Add(&xr, &yr), y.(BigComplex).Im})
		case Infinity:
			return y
		default:
			panic("Plus: BigRat and non-numeric invalid")
		}

	case Float:
		ax := x.(Float)
		switch y.(type) {
		case Float:
			return narrow(ax + y.(Float))
		case BigFloat:
			xr := *big.NewFloat(float64(ax))
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(0).Add(&xr, &ay)))
		case Complex:
			return narrow(Complex(complex(ax, 0)) + y.(Complex))
		case BigComplex:
			xr := *big.NewFloat(float64(ax))
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			return narrow(BigComplex{*big.NewFloat(0).Add(&xr, &yr), y.(BigComplex).Im})
		case Infinity:
			return y
		default:
			panic("Plus: Float and non-numeric invalid")
		}

	case BigFloat:
		ax := big.Float(x.(BigFloat))
		switch y.(type) {
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(0).Add(&ax, &ay)))
		case Complex:
			ayr := *big.NewFloat(real(complex128(y.(Complex))))
			return narrow(BigComplex{*big.NewFloat(0).Add(&ax, &ayr),
				*big.NewFloat(imag(complex128(y.(Complex))))})
		case BigComplex:
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			return narrow(BigComplex{*big.NewFloat(0).Add(&ax, &ayr), y.(BigComplex).Im})
		case Infinity:
			return y
		default:
			panic("Plus: BigFloat and non-numeric invalid")
		}

	case Complex:
		ax := x.(Complex)
		switch y.(type) {
		case Complex:
			return narrow(ax + y.(Complex))
		case BigComplex:
			xr := *big.NewFloat(real(complex128(ax)))
			xi := *big.NewFloat(imag(complex128(ax)))
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			return narrow(BigComplex{*big.NewFloat(0).Add(&xr, &yr), *big.NewFloat(0).Add(&xi, &yi)})
		case Infinity:
			return y
		default:
			panic("Plus: Complex and non-numeric invalid")
		}

	case BigComplex:
		switch y.(type) {
		case BigComplex:
			xr := big.Float(BigFloat(x.(BigComplex).Re))
			xi := big.Float(BigFloat(x.(BigComplex).Im))
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			return narrow(BigComplex{*big.NewFloat(0).Add(&xr, &yr), *big.NewFloat(0).Add(&xi, &yi)})
		case Infinity:
			return y
		default:
			panic("Plus: BigComplex and non-numeric invalid")
		}

	case Infinity:
		switch y.(type) {
		case Infinity:
			return nil //Inf + Inf = NaN
		default:
			panic("Plus: Infinity and non-numeric invalid")
		}

	case time.Time:
		switch y.(type) {
		case Int:
			return x.(time.Time).Add(time.Duration(int64(time.Hour) * 24 * int64(y.(Int))))
		default:
			panic("Plus: Date and non-Int invalid")
		}

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

	case Slice:
		switch y.(type) {
		case Slice:
			return Slice(append([]interface{}(x.(Slice)), []interface{}(y.(Slice))...))
		default:
			panic("Plus: []any can only operate with []any")
		}

	case Map:
		switch y.(type) {
		case Map:
			//TODO: for case Map, return a new map with all entries from old map added
			/*m := (*pp).(Map)
			for _, u := range it {
				m[u.(MapEntry).Key] = u.(MapEntry).Val
				//####
			}
			return pp
			*/
			return nil
		default:
			panic("Plus: ops.Map can only operate with ops.Map")
		}

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
		return !x.(bool)
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
		//TODO: do we check for NaN here ???
		return Float(-x.(Float))
	case BigFloat:
		ax := big.Float(x.(BigFloat))
		return BigFloat(*big.NewFloat(0).Neg(&ax))
	case Complex:
		//TODO: do we check for NaN here ???
		return Complex(-x.(Complex))
	case BigComplex:
		xr := big.Float(BigFloat(x.(BigComplex).Re))
		xi := big.Float(BigFloat(x.(BigComplex).Im))
		return BigComplex{*big.NewFloat(0).Neg(&xr), *big.NewFloat(0).Neg(&xi)}
	case Infinity:
		return x

	case u8.Codepoint:
		return CharClass("[^" + u8.Sur(x.(u8.Codepoint)) + "]")

	case CharClass:
		mid, neg := x.(CharClass).unwrap()
		if neg {
			return CharClass("[" + mid + "]")
		} else {
			return CharClass("[^" + mid + "]")
		}

	//TODO: logic for cases Text and Regex

	default:
		panic("Negate: non-numeric invalid")
	}
}

/*
Minus negates the 2nd arg and adds the two together.
*/
func Minus(x, y interface{}) interface{} {
	x, y = widen(x), widen(y)

	switch x.(type) {
	//TODO: case Map
	case Slice:
		switch y.(type) {
		case Slice:
			ax := x.(Slice)
			ay := y.(Slice)
			ns := NewSlice(0)
		oldouter:
			for ia, a := range ax {
				for ib, b := range ay {
					if ib > ia {
						if IsEqual(a, b) {
							continue oldouter
						}
						_ = LeftShiftAssign(&ns, a)
					}
				}
			}
			return ns.(Slice)
		default:
			panic("Minus: []any requires []any arg")
		}
	default:
		return narrow(Plus(x, Negate(y)))
	}
}

/*
Complement returns the bitwise complement of a number.
*/
func Complement(x interface{}) interface{} {
	x = widen(x)
	switch x.(type) {
	case nil:
		return nil
	case bool:
		return !x.(bool)
	case Int:
		return ^x.(Int)
	case u8.Text:
		return ToRegex(x.(u8.Text))
	default:
		panic("Complement: invalid type")
	}
}

/*
Mult creates a new element being the product of the two arguments, for numeric types.
Two Ints are promoted to BigInt if it overflows.
For one argument of Codepoint or Text type, and the other of numeric type z,
returns Text repeated z times.
For one argument of CharClass or Regex type, and the other of numeric type z,
returns Regex matching repetition of z times.
*/
func Mult(x, y interface{}) interface{} {
	x, y = widen(x), widen(y)
	if numericTypes(x, y) {
		x, y = y, x
	}

	switch x.(type) {
	case u8.Codepoint, u8.Text, CharClass, Regex:
		switch y.(type) {
		case multiusePair:
			return RegexRepeat(x, y.(multiusePair), true)
		case Int:
			return RegexRepeat(x, y.(Int), true)
		default:
			panic("Mult: Codepoint, Text, CharClass, or Regex multiplied with invalid value")
		}
		//TODO: For type Codepoint or Text, return repeated Text

	case multiusePair: //TODO: make sure this works correctly with typeWeights
		switch y.(type) {
		case u8.Codepoint, u8.Text, CharClass, Regex:
			return RegexRepeat(y, x, false)
		case parsec.Parser:
			return ParserRepeat(y, x)
		case func(...interface{}) interface{}:
			ay := y.(func(...interface{}) interface{})
			yfwd := parsec.Fwd(ay().(func(...interface{}) interface{})).(parsec.Parser)
			return ParserRepeat(yfwd, x)
		default:
			panic("Mult: positive integer range multiplied with invalid value")
		}
		//TODO: For type Codepoint or Text, return repeated Text

	case nil:
		switch y.(type) {
		case nil, bool, Int, BigInt, BigRat, Float, BigFloat, Complex, BigComplex, Infinity:
			return nil
		default:
			panic("Mult: nil and non-numeric invalid")
		}

	case bool:
		ax := x.(bool)
		switch y.(type) {
		case bool:
			return ax && y.(bool)
		case Int:
			if ax {
				return y.(Int)
			} else {
				return Int(0)
			}
		case BigInt:
			if ax {
				return y.(BigInt)
			} else {
				return BigInt(*big.NewInt(0))
			}
		case BigRat:
			if ax {
				return y.(BigRat)
			} else {
				return BigInt(*big.NewInt(0))
			}
		case Float:
			if ax {
				return y.(Float)
			} else {
				return Float(0)
			}
		case BigFloat:
			if ax {
				return y.(BigFloat)
			} else {
				return BigFloat(*big.NewFloat(0.0))
			}
		case Complex:
			if ax {
				return y.(Complex)
			} else {
				return Complex(0)
			}
		case BigComplex:
			if ax {
				return y.(BigComplex)
			} else {
				return BigComplex{*big.NewFloat(0.0), *big.NewFloat(0.0)}
			}
		case Infinity:
			if ax {
				return y
			} else {
				return nil
			}
		default:
			panic("Mult: bool and non-numeric invalid")
		}

	case Int:
		ax := x.(Int)
		switch y.(type) {
		case u8.Codepoint, u8.Text, CharClass, Regex: //TODO: make sure this works correctly with typeWeights
			return RegexRepeat(y, multiusePair{ax, ax}, false)
		//TODO: For type Codepoint or Text, return repeated Text

		case Int:
			ay := y.(Int)
			abs := func(x Int) Int {
				if x < 0 {
					return -x
				} else {
					return x
				}
			}
			c := ax * ay
			absc := abs(c)
			if absc >= abs(ax) && absc >= abs(ay) { //check for overflow
				return c
			} else {
				return BigInt(*big.NewInt(0).Mul(big.NewInt(int64(ax)), big.NewInt(int64(ay))))
			}
		case BigInt:
			ay := big.Int(y.(BigInt))
			return BigInt(*big.NewInt(0).Mul(big.NewInt(int64(ax)), &ay))
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return narrow(BigRat(*big.NewRat(0, 1).Mul(big.NewRat(int64(ax), 1), &ay)))
		case Float:
			return narrow(Float(ax) * y.(Float))
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(0).Mul(big.NewFloat(float64(ax)), &ay)))
		case Complex:
			return narrow(Complex(complex(float64(ax), 0)) * y.(Complex))
		case BigComplex:
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			pr := *big.NewFloat(0).Mul(big.NewFloat(float64(ax)), &yr)
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			pi := *big.NewFloat(0).Mul(big.NewFloat(float64(ax)), &yi)
			return narrow(BigComplex{pr, pi})
		case Infinity:
			if ax == 0 {
				return nil
			} else {
				return y
			}
		default:
			panic("Mult: Int and non-numeric invalid")
		}

	case BigInt:
		ax := big.Int(x.(BigInt))
		switch y.(type) {
		case BigInt:
			ay := big.Int(y.(BigInt))
			return BigInt(*big.NewInt(0).Mul(&ax, &ay))
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return narrow(BigRat(*big.NewRat(0, 1).Mul(big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)), &ay)))
		case Float:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return narrow(Float(xf) * y.(Float))
		case BigFloat:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			xr := *big.NewFloat(xf)
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(0).Mul(&xr, &ay)))
		case Complex:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return narrow(Complex(complex(xf, 0)) * y.(Complex))
		case BigComplex:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			xr := *big.NewFloat(xf)
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			return narrow(BigComplex{*big.NewFloat(0).Mul(&xr, &yr), *big.NewFloat(0).Mul(&xr, &yi)})
		case Infinity:
			if big.NewInt(0).Cmp(&ax) == 0 {
				return nil
			} else {
				return y
			}
		default:
			panic("Mult: Int and non-numeric invalid")
		}

	case BigRat:
		ax := big.Rat(x.(BigRat))
		switch y.(type) {
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return narrow(BigRat(*big.NewRat(0, 1).Mul(&ax, &ay)))
		case Float:
			xf, _ := ax.Float64()
			return narrow(Float(xf) * y.(Float))
		case BigFloat:
			xf, _ := ax.Float64()
			xr := *big.NewFloat(xf)
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(1.0).Mul(&xr, &ay)))
		case Complex:
			xf, _ := ax.Float64()
			return narrow(Complex(complex(xf, 0)) * y.(Complex))
		case BigComplex:
			xf, _ := ax.Float64()
			xr := *big.NewFloat(xf)
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			return narrow(BigComplex{*big.NewFloat(1.0).Mul(&xr, &yr), *big.NewFloat(1.0).Mul(&xr, &yi)})
		case Infinity:
			if big.NewRat(0, 1).Cmp(&ax) == 0 {
				return nil
			} else {
				return y
			}
		default:
			panic("Mult: Rat and non-numeric invalid")
		}

	case Float:
		ax := x.(Float)
		switch y.(type) {
		case Float:
			return narrow(ax * y.(Float))
		case BigFloat:
			xf := *big.NewFloat(float64(ax))
			yf := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(1.0).Mul(&xf, &yf)))
		case Complex:
			return narrow(Complex(complex(ax, 0)) * y.(Complex))
		case BigComplex:
			xf := *big.NewFloat(float64(ax))
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			return narrow(BigComplex{*big.NewFloat(1.0).Mul(&xf, &yr), *big.NewFloat(1.0).Mul(&xf, &yi)})
		case Infinity:
			if x.(Float) == 0.0 {
				return nil
			} else {
				return y
			}
		default:
			panic("Mult: Float and non-numeric invalid")
		}

	case BigFloat:
		ax := big.Float(x.(BigFloat))
		switch y.(type) {
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return narrow(BigFloat(*big.NewFloat(1.0).Mul(&ax, &ay)))
		case Complex:
			yr := *big.NewFloat(real(complex128(y.(Complex))))
			yi := *big.NewFloat(imag(complex128(y.(Complex))))
			return narrow(BigComplex{*big.NewFloat(1.0).Mul(&ax, &yr), *big.NewFloat(1.0).Mul(&ax, &yi)})
		case BigComplex:
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			return narrow(BigComplex{*big.NewFloat(1.0).Mul(&ax, &yr), *big.NewFloat(1.0).Mul(&ax, &yi)})
		case Infinity:
			if big.NewFloat(0.0).Cmp(&ax) == 0 {
				return nil
			} else {
				return y
			}
		default:
			panic("Mult: BigFloat and non-numeric invalid")
		}

	case Complex:
		ax := x.(Complex)
		switch y.(type) {
		case Complex:
			return narrow(ax * y.(Complex))
		case BigComplex:
			xr := *big.NewFloat(real(complex128(ax)))
			xi := *big.NewFloat(imag(complex128(ax)))
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			realPart := *big.NewFloat(1.0).Sub(big.NewFloat(1.0).Mul(&xr, &yr), big.NewFloat(1.0).Mul(&xi, &yi))
			imagPart := *big.NewFloat(1.0).Add(big.NewFloat(1.0).Mul(&xi, &yr), big.NewFloat(1.0).Mul(&xr, &yi))
			return narrow(BigComplex{realPart, imagPart})
		case Infinity:
			if ax == 0.0+0.0i {
				return nil
			} else {
				return y
			}
		default:
			panic("Mult: Complex and non-numeric invalid")
		}

	case BigComplex:
		ax := x.(BigComplex)
		switch y.(type) {
		case BigComplex:
			xr := big.Float(BigFloat(ax.Re))
			xi := big.Float(BigFloat(ax.Im))
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			realPart := *big.NewFloat(1.0).Sub(big.NewFloat(1.0).Mul(&xr, &yr), big.NewFloat(1.0).Mul(&xi, &yi))
			imagPart := *big.NewFloat(1.0).Add(big.NewFloat(1.0).Mul(&xi, &yr), big.NewFloat(1.0).Mul(&xr, &yi))
			return narrow(BigComplex{realPart, imagPart})
		case Infinity:
			xr := big.Float(BigFloat(ax.Re))
			xi := big.Float(BigFloat(ax.Im))
			if big.NewFloat(0.0).Cmp(&xr) == 0 && big.NewFloat(0.0).Cmp(&xi) == 0 {
				return nil
			} else {
				return y
			}
		default:
			panic("Mult: BigComplex and non-numeric invalid")
		}

	case Infinity:
		switch y.(type) {
		case Infinity:
			return y
		default:
			panic("Plus: Infinity and non-numeric invalid")
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
		return nil
	case bool:
		return !x.(bool) //TODO: is this correct logic ???
	case Int:
		if x.(Int) == 0 {
			return Inf
		}
		return narrow(BigRat(*big.NewRat(1, int64(x.(Int)))))
	case BigInt:
		if reflect.DeepEqual(x.(BigInt), BigInt(*big.NewInt(0))) {
			return Inf
		}
		ax := big.Int(x.(BigInt))
		return narrow(BigRat(*big.NewRat(0, 1).SetFrac(big.NewInt(1), &ax))) //FIX: make copy of denom first ???
	case BigRat:
		if reflect.DeepEqual(x.(BigRat), BigRat(*big.NewRat(0, 1))) {
			return Inf
		}
		ax := big.Rat(x.(BigRat))
		return narrow(BigRat(*big.NewRat(0, 1).Inv(&ax)))
	case Float:
		if x.(Float) == 0 {
			return Inf
		}
		return narrow(Float(1 / float64(x.(Float))))
	case BigFloat:
		ax := big.Float(x.(BigFloat))
		if big.NewFloat(0.0).Cmp(&ax) == 0 {
			return Inf
		}
		return narrow(BigFloat(*big.NewFloat(1.0).Quo(big.NewFloat(1.0), &ax)))
	case Complex:
		if x.(Complex) == 0 {
			return Inf
		}
		return narrow(Complex(1 / complex128(x.(Complex))))
	case BigComplex:
		xr := big.Float(BigFloat(x.(BigComplex).Re))
		xi := big.Float(BigFloat(x.(BigComplex).Im))
		if big.NewFloat(0.0).Cmp(&xr) == 0 && big.NewFloat(0.0).Cmp(&xi) == 0 {
			return Inf
		}
		mod := big.NewFloat(0.0).Add(big.NewFloat(0.0).Mul(&xr, &xr), big.NewFloat(0.0).Mul(&xi, &xi))
		cr := *big.NewFloat(1.0).Quo(&xr, mod)
		ci := *big.NewFloat(1.0).Quo(big.NewFloat(1.0).Neg(&xi), mod)
		return narrow(BigComplex{cr, ci})
	case Infinity:
		return 0
	default:
		panic("Invert: non-numeric invalid")
	}
}

/*
Divide inverts the second argument and multiplies the two together.
Division by nil or false gives infinity.
Divide result is of the same type as whichever argument is higher in the numeric hierarchy.
A division by a Int or BigInt promotes to a BigRat.
*/
func Divide(x, y interface{}) interface{} {
	return narrow(Mult(x, Invert(y)))
}

/*
Mod takes two Int, BigInt, or Infinite args,
and returns the modulus.
Anything mod 0 is infinity;
anything mod infinity is unchanged.

//TODO: both 0%0 and inf%inf should return nil
//TODO: put in logic for true and false
*/
func Mod(x, y interface{}) interface{} {
	x, y = widen(x), widen(y)
	switch x.(type) {
	case u8.Codepoint:
		ax := x.(u8.Codepoint)
		switch y.(type) {
		case u8.Text:
			ay := u8.Surr(y.(u8.Text))
			return fmt.Sprintf("%"+ay, ax)
		default:
			panic("Mod: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
		}
	case Int:
		ax := x.(Int)
		switch y.(type) {
		case u8.Text:
			ay := u8.Surr(y.(u8.Text))
			return fmt.Sprintf("%"+ay, ax)
		case Int:
			ay := y.(Int)
			if ay == 0 {
				return Inf
			}
			return Int(int64(ax) % int64(y.(Int)))
		case BigInt:
			ay := big.Int(y.(BigInt))
			if big.NewInt(0.0).Cmp(&ay) == 0 {
				return Inf
			}
			return BigInt(*big.NewInt(0).Mod(big.NewInt(int64(ax)), &ay))
		case Infinity:
			return x
		default:
			panic("Mod: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
		}
	case BigInt:
		ax := big.Int(x.(BigInt))
		switch y.(type) {
		case u8.Text:
			ay := u8.Surr(y.(u8.Text))
			return fmt.Sprintf("%"+ay, ax)
		case Int:
			ay := y.(Int)
			if ay == 0 {
				return Inf
			}
			return BigInt(*big.NewInt(0).Mod(&ax, big.NewInt(int64(ay))))
		case BigInt:
			ay := big.Int(y.(BigInt))
			if big.NewInt(0.0).Cmp(&ay) == 0 {
				return Inf
			}
			return BigInt(*big.NewInt(0).Mod(&ax, &ay))
		case Infinity:
			return x
		default:
			panic("Mod: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
		}
	case Infinity:
		switch y.(type) {
		case u8.Text:
			ay := u8.Surr(y.(u8.Text))
			return fmt.Sprintf("%"+ay, x)
		default:
			return x
		}
	default:
		switch y.(type) {
		case u8.Text:
			ay := u8.Surr(y.(u8.Text))
			return fmt.Sprintf("%"+ay, x)
		default:
			panic("Mod: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
		}
	}
}

//==============================================================================

/*
Power returns a result of the same type as the argument higher in the numeric hierarchy.
Raising to the power of a nil or false generates a panic;
an Int to the power of a Int is a BigInt;
an Int or BigInt to the power of a BigRat is a Float;
a BigRat to the power of an Int, BigInt, or BigRat is a Float.

//WARNING: This function is alpha version only.

//TODO: add code for BigFloat and BigComplex ???
//TODO: change output to []interface{} to cater for multiple results
//TODO: add tests
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
			panic("Power: only Int, BigInt, BigRat, Float, and Complex args valid")
		}

	case BigInt:
		switch y.(type) {
		case bool:
			if y.(bool) {
				return x.(BigInt)
			} else {
				panic("Power: only Int, BigInt, BigRat, Float, and Complex args valid")
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
			panic("Power: only Int, BigInt, BigRat, Float, and Complex args valid")
		}

	case BigRat:
		switch y.(type) {
		case bool:
			if y.(bool) {
				return x.(BigRat)
			} else {
				panic("Power: only Int, BigInt, BigRat, Float, and Complex args valid")
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
			panic("Power: only Int, BigInt, BigRat, Float, and Complex args valid")
		}

	case Float:
		switch y.(type) {
		case bool:
			if y.(bool) {
				return x.(Float)
			} else {
				panic("Power: only Int, BigInt, BigRat, Float, and Complex args valid")
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
			panic("Power: only Int, BigInt, BigRat, Float, and Complex args valid")
		}

	case Complex:
		switch y.(type) {
		case bool:
			if y.(bool) {
				return x.(Complex)
			} else {
				panic("Power: only Int, BigInt, BigRat, Float, and Complex args valid")
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
			panic("Power: only Int, BigInt, BigRat, Float, and Complex args valid")
		}

	default:
		panic("Power: only Int, BigInt, BigRat, Float, and Complex args valid")
	}
}

//==============================================================================

type multiusePair struct {
	first, second interface{}
}

type MapEntryOrPosIntRange interface {
	Key() interface{}
	Val() interface{}
	From() Int
	To() Int
	IsToInf() bool
	fmt.Stringer //would duplicate if we composed interface{MapEntry; PosIntRange}
}

func NewMapEntryOrPosIntRange(first, second interface{}) MapEntryOrPosIntRange {
	return multiusePair{widen(first), widen(second)}
}

/*
MapEntry represents a key-value pair for a Map.
*/
type MapEntry interface {
	Key() interface{}
	Val() interface{}
	fmt.Stringer
}

func NewMapEntry(key, val interface{}) MapEntry {
	return multiusePair{widen(key), widen(val)}
}

/*
PosIntRange represents a range in the positive integers.
The first field can be any Int or BigInt.
The second field can be any Int or BigInt, or Inf.
*/
type PosIntRange interface {
	From() Int
	To() Int
	IsToInf() bool
	fmt.Stringer
}

func NewPosIntRange(from, to interface{}) PosIntRange {
	return multiusePair{widen(from), widen(to)}
}

func (r multiusePair) String() string {
	return fmt.Sprintf("%v: %v", r.first, r.second)
}

//Key returns the first field as a map key.
func (r multiusePair) Key() interface{} {
	return r.first
}

//Key returns the second field as a map value.
func (r multiusePair) Val() interface{} {
	return r.second
}

/*
From returns the first field as an Int.
It returns 0 if it's negative.
It returns the maximum Int if it's a BigInt greater than that.
*/
func (r multiusePair) From() Int {
	return ToInt(r.first)
}

/*
To returns the second field as an Int.
It returns 0 if it's negative.
It returns the maximum Int if it's infinity or a BigInt greater than that.
*/
func (r multiusePair) To() Int {
	return ToInt(r.second)
}

/*
IsToInf returns true if the second field is infinity, otherwise false.
*/
func (r multiusePair) IsToInf() bool {
	return r.second == Inf
}

//==============================================================================
