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
	"regexp"
	"time"

	u8 "github.com/grolang/gro/utf88"
)

//--------------------------------------------------------------------------------
//nil, bool, Int, BigInt, BigRat, Float, BigFloat, Complex, BigComplex, Infinity
//Any, Void, Pair, Codepoint, Text, CharClass, Regex, Parsec, Time, Slice, Map, Func
//--------------------------------------------------------------------------------

//==============================================================================

// widen converts a value to a Groo type, then rewidths it if possible.
func widen(x Any) Any {
	switch x.(type) {
	case nil, bool, Infinity:
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
			return rewidth(Float(x.(float32)))
		}
	case float64:
		if math.IsInf(x.(float64), 0) {
			return Inf
		} else {
			return rewidth(Float(x.(float64)))
		}

	case complex64:
		if cmplx.IsInf(complex128(x.(complex64))) {
			return Inf
		} else {
			return rewidth(Complex(x.(complex64)))
		}
	case complex128:
		if cmplx.IsInf(x.(complex128)) {
			return Inf
		} else {
			return rewidth(Complex(x.(complex128)))
		}

	case big.Int:
		return rewidth(BigInt(x.(big.Int)))
	case big.Rat:
		return rewidth(BigRat(x.(big.Rat)))
	case big.Float:
		return rewidth(BigFloat(x.(big.Float)))

	case BigInt:
		return rewidth(x.(BigInt))
	case BigRat:
		return rewidth(x.(BigRat))
	case BigFloat:
		return rewidth(x.(BigFloat))
	case BigComplex:
		return rewidth(x.(BigComplex))

	case time.Time:
		d := x.(time.Time)
		return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)

	case string:
		return u8.Desur(x.(string))
		//TODO: if string is a valid Codepoint, return Codepoint ???

	case Void, u8.Codepoint, u8.Text, CharClass, Regex, Parsec, Slice, *Map, Func:
		return x

	default:
		switch reflect.TypeOf(x).Kind() {
		case reflect.Slice:
			vs := []Any{}
			if x != nil {
				s := reflect.ValueOf(x)
				for i := 0; i < s.Len(); i++ {
					vs = append(vs, widen(s.Index(i).Interface()))
				}
			}
			return Slice(vs)

		case reflect.Map:
			m := NewMap()
			if x != nil {
				s := reflect.ValueOf(x)
				for _, k := range s.MapKeys() {
					kw := widen(k.Interface())
					vw := widen(s.MapIndex(k).Interface())
					m.Add(kw, vw)
				}
			}
			return m

		default:
			return x
		}
	}
}

/*
rewidth returns Inf if x is infinite, nil if x is NaN, otherwise converts
the type of x to its simplest possible numeric Groo type.
x is returned unchanged if it's not a numeric Groo type.
*/
func rewidth(x Any) Any {
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
narrow converts values to types outside the Groo hierarchy.
*/
func narrow(x Any) Any {
	switch x.(type) {

	case nil, bool:
		return x

	case Int:
		return int64(x.(Int))

	case BigInt:
		return big.Int(x.(BigInt))

	case BigRat:
		return big.Rat(x.(BigRat))

	case Float:
		return float64(x.(Float))

	case BigFloat:
		return big.Float(x.(BigFloat))

	case Complex:
		return complex128(x.(Complex))

	case BigComplex:
		return x //TODO

	case Infinity:
		return x //TODO

	case u8.Codepoint:
		return x //TODO

	case u8.Text:
		return u8.Surr(x.(u8.Text))

	case CharClass:
		return x //TODO

	case Regex:
		return x //TODO

	case Parsec:
		return x //TODO

	case Slice:
		return []Any(x.(Slice))

	case *Map:
		return x.(*Map).lkp

	case Void, time.Time, Func:
		return x

	default:
		return x
	}
}

//==============================================================================

/*
Identity converts:
  float32 and float64 to Float (float64)
  complex64 and complex128 to Complex (complex128)
  but NaN converted to nil and Inf converted to inf
  int, int8, int16, int32/rune, int64, uint8, uint16, and uint32 to Int
  uint and uint64 to Int, unless it overflows, in which case it converts to BigInt
  string to u8.Text
  eliminates fractions of days from Date, and switches to UTC
and leaves the nil, bool, BigInt, BigRat, and BigFloat types unchanged.
*/
func Identity(x Any) Any {
	return widen(x)
}

//==============================================================================

var numericWeights = map[reflect.Type]int{
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

func numericTypes(x, y Any) bool {
	return numericWeights[reflect.TypeOf(x)] != 0 && numericWeights[reflect.TypeOf(y)] != 0 &&
		numericWeights[reflect.TypeOf(x)] > numericWeights[reflect.TypeOf(y)]
}

var comparingWeights = map[reflect.Type]int{
	reflect.TypeOf(nil):             1,
	reflect.TypeOf(Void{}):          2,
	reflect.TypeOf(false):           3,
	reflect.TypeOf(u8.Codepoint(0)): 4,
	reflect.TypeOf(u8.Desur("")):    5,
	reflect.TypeOf(Slice{}):         6,
	reflect.TypeOf(NewMap()):        7,
}

func comparingTypes(x, y Any) bool {
	return comparingWeights[reflect.TypeOf(x)] != 0 && comparingWeights[reflect.TypeOf(y)] != 0 &&
		comparingWeights[reflect.TypeOf(x)] > comparingWeights[reflect.TypeOf(y)]
}

/*
IsEqual compares the Identity of each argument to see if they are equal,
irrespective of their type.
Numerics and booleans are compared with each other, Dates with each other,
and Text with each other, otherwise, reflect.DeepEqual is called.
*/
func IsEqual(x, y Any) bool {
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
			return !toBool(x)
		case Infinity:
			return false
		}

	case bool:
		switch y.(type) {
		case bool:
			return x.(bool) == y.(bool)
		case Int, BigInt, BigRat, Float, BigFloat, Complex, BigComplex:
			return x.(bool) == toBool(y)
		case Infinity:
			return x.(bool)
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
		}

	case Infinity:
		switch y.(type) {
		case Infinity:
			return true
		}
	}

	//- - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	if comparingTypes(x, y) {
		x, y = y, x
	}
	switch x.(type) {
	case nil:
		switch y.(type) {
		//case nil in numeric switch
		case Void:
			return true
		//case bool in numeric switch
		//TODO: u8.Codepoint, CharClass, Regex, Parsec, time.Time, Func ???
		case u8.Text, Slice, *Map:
			return !toBool(x)
		}

	case Void:
		switch y.(type) {
		case Void:
			return true
		case bool:
			return !y.(bool)
			//TODO: u8.Codepoint, CharClass, Regex, Parsec, time.Time, Func ???
			//TODO: u8.Text, Slice, *Map
		}

	case bool:
		switch y.(type) {
		//case bool in numeric switch
		//TODO: u8.Codepoint, CharClass, Regex, Parsec, time.Time, Func ???
		case u8.Text, Slice, *Map:
			return x.(bool) == toBool(y)
		}

	//TODO: u8.Codepoint, CharClass, Regex, Parsec, Func ???

	case Pair:
		switch y.(type) {
		case Pair:
			ax := x.(Pair)
			ay := y.(Pair)
			return ax.First == ay.First && ax.Second == ay.Second
		}

	case u8.Text:
		switch y.(type) {
		case u8.Text:
			return string(u8.SurrogatePoints(x.(u8.Text))) == string(u8.SurrogatePoints(y.(u8.Text)))
		}

	case Slice:
		switch y.(type) {
		case Slice:
			return reflect.DeepEqual(x.(Slice), y.(Slice))
		}

	case *Map:
		switch y.(type) {
		case *Map:
			return x.(*Map).IsEqual(y.(*Map))
		}

	case time.Time:
		switch y.(type) {
		case time.Time:
			return x.(time.Time).Equal(y.(time.Time))
		}

	default:
		return reflect.DeepEqual(x, y)
	}

	panic(fmt.Sprintf("IsEqual: incompatible types %T and %T", x, y))
}

/*
IsLessThan returns true if both arguments are non-complex numeric or both Dates,
and x is less than y; otherwise it returns false.
It panics if either arg is some other type.
*/
func IsLessThan(x, y Any) bool {
	x, y = widen(x), widen(y)
	if numericTypes(x, y) {
		return !IsLessThan(y, x) && !IsEqual(x, y)
	}

	//TODO: compare nil, Void, boolean, and Infinity ???

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
		}

	case Float:
		ax := x.(Float)
		switch y.(type) {
		case Float:
			return ax < y.(Float)
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return big.NewFloat(float64(ax)).Cmp(&ay) == -1
		}

	case BigFloat:
		switch y.(type) {
		case BigFloat:
			ax := big.Float(x.(BigFloat))
			ay := big.Float(y.(BigFloat))
			return ax.Cmp(&ay) == -1
		}

	case time.Time:
		switch y.(type) {
		case time.Time:
			return x.(time.Time).Before(y.(time.Time))
		}

		//TODO: u8.Codepoint and u8.Text cases

	}

	panic(fmt.Sprintf("IsLessThan: incompatible types %T and %T", x, y))
}

/*
IsNotEqual returns true if both arguments are non-complex numeric or Dates,
and x is not equal to y.
It panics if either arg is nil, bool, Complex, BigComplex, or non-numeric.
*/
func IsNotEqual(x, y Any) bool {
	return !IsEqual(x, y)
}

/*
IsLessOrEqual returns true if both arguments are non-complex numeric or Dates,
and x is less than or equal to y.
It panics if either arg is nil, bool, Complex, BigComplex, or non-numeric.
*/
func IsLessOrEqual(x, y Any) bool {
	return IsLessThan(x, y) || IsEqual(x, y)
}

/*
IsGreaterThan returns true if both arguments are non-complex numeric or Dates,
and x is greater than y.
It panics if either arg is nil, bool, Complex, BigComplex, or non-numeric.
*/
func IsGreaterThan(x, y Any) bool {
	return !IsLessOrEqual(x, y)
}

/*
IsGreaterOrEqual returns true if both arguments are non-complex numeric or Dates,
and x is greater than or equal to y.
It panics if either arg is nil, bool, Complex, BigComplex, or non-numeric.
*/
func IsGreaterOrEqual(x, y Any) bool {
	return !IsLessThan(x, y)
}

//==============================================================================

/*
Plus creates a new element being the sum of the two arguments, for numeric types.
Two Ints are promoted to BigInt if it overflows.
For Date and integer, create new Date being y number of days added to Date x.
For Text and Codepoint, create new Text being concatenation.
For Slice x, return a new slice with element y appended.
*/
func Plus(x, y Any) Any {
	x, y = widen(x), widen(y)

	if numericTypes(x, y) {
		x, y = y, x
	}
	switch x.(type) {
	case nil:
		switch y.(type) {
		case nil, bool, Int, BigInt, BigRat, Float, BigFloat, Complex, BigComplex, Infinity:
			return nil
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
			return rewidth(BigRat(*big.NewRat(0, 1).Add(big.NewRat(int64(ax), 1), &ay)))
		case Float:
			return rewidth(Float(ax) + y.(Float))
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return rewidth(BigFloat(*big.NewFloat(0.0).Add(big.NewFloat(float64(ax)), &ay)))
		case Complex:
			return rewidth(Complex(complex(float64(ax), 0)) + y.(Complex))
		case BigComplex:
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			sr := *big.NewFloat(0.0).Add(big.NewFloat(float64(ax)), &yr)
			return rewidth(BigComplex{sr, y.(BigComplex).Im})
		case Infinity:
			return y
		case time.Time:
			return y.(time.Time).Add(time.Duration(int64(time.Hour) * 24 * int64(ax)))
		}

	case BigInt:
		ax := big.Int(x.(BigInt))
		switch y.(type) {
		case BigInt:
			ay := big.Int(y.(BigInt))
			return BigInt(*big.NewInt(0).Add(&ax, &ay))
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return rewidth(BigRat(*big.NewRat(0, 1).Add(big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)), &ay)))
		case Float:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return rewidth(Float(xf) + y.(Float))
		case BigFloat:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			axf := *big.NewFloat(xf)
			ay := big.Float(y.(BigFloat))
			return rewidth(BigFloat(*big.NewFloat(0).Add(&axf, &ay)))
		case Complex:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return rewidth(Complex(complex(xf, 0)) + y.(Complex))
		case BigComplex:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			xr := *big.NewFloat(xf)
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			return rewidth(BigComplex{*big.NewFloat(0).Add(&xr, &yr), y.(BigComplex).Im})
		case Infinity:
			return y
		}

	case BigRat:
		ax := big.Rat(x.(BigRat))
		switch y.(type) {
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return rewidth(BigRat(*big.NewRat(0, 1).Add(&ax, &ay)))
		case Float:
			xf, _ := ax.Float64()
			return rewidth(Float(xf) + y.(Float))
		case BigFloat:
			xf, _ := ax.Float64()
			xr := *big.NewFloat(xf)
			ay := big.Float(y.(BigFloat))
			return rewidth(BigFloat(*big.NewFloat(0).Add(&xr, &ay)))
		case Complex:
			xf, _ := ax.Float64()
			return rewidth(Complex(complex(xf, 0)) + y.(Complex))
		case BigComplex:
			xf, _ := ax.Float64()
			xr := *big.NewFloat(xf)
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			return rewidth(BigComplex{*big.NewFloat(0).Add(&xr, &yr), y.(BigComplex).Im})
		case Infinity:
			return y
		}

	case Float:
		ax := x.(Float)
		switch y.(type) {
		case Float:
			return rewidth(ax + y.(Float))
		case BigFloat:
			xr := *big.NewFloat(float64(ax))
			ay := big.Float(y.(BigFloat))
			return rewidth(BigFloat(*big.NewFloat(0).Add(&xr, &ay)))
		case Complex:
			return rewidth(Complex(complex(ax, 0)) + y.(Complex))
		case BigComplex:
			xr := *big.NewFloat(float64(ax))
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			return rewidth(BigComplex{*big.NewFloat(0).Add(&xr, &yr), y.(BigComplex).Im})
		case Infinity:
			return y
		}

	case BigFloat:
		ax := big.Float(x.(BigFloat))
		switch y.(type) {
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return rewidth(BigFloat(*big.NewFloat(0).Add(&ax, &ay)))
		case Complex:
			ayr := *big.NewFloat(real(complex128(y.(Complex))))
			return rewidth(BigComplex{*big.NewFloat(0).Add(&ax, &ayr),
				*big.NewFloat(imag(complex128(y.(Complex))))})
		case BigComplex:
			ayr := big.Float(BigFloat(y.(BigComplex).Re))
			return rewidth(BigComplex{*big.NewFloat(0).Add(&ax, &ayr), y.(BigComplex).Im})
		case Infinity:
			return y
		}

	case Complex:
		ax := x.(Complex)
		switch y.(type) {
		case Complex:
			return rewidth(ax + y.(Complex))
		case BigComplex:
			xr := *big.NewFloat(real(complex128(ax)))
			xi := *big.NewFloat(imag(complex128(ax)))
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			return rewidth(BigComplex{*big.NewFloat(0).Add(&xr, &yr), *big.NewFloat(0).Add(&xi, &yi)})
		case Infinity:
			return y
		}

	case BigComplex:
		switch y.(type) {
		case BigComplex:
			xr := big.Float(BigFloat(x.(BigComplex).Re))
			xi := big.Float(BigFloat(x.(BigComplex).Im))
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			return rewidth(BigComplex{*big.NewFloat(0).Add(&xr, &yr), *big.NewFloat(0).Add(&xi, &yi)})
		case Infinity:
			return y
		}

	case Infinity:
		switch y.(type) {
		case Infinity:
			return nil //Inf + Inf = NaN
		}
	}

	//- - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	switch x.(type) {
	case Int:
		switch y.(type) {
		case time.Time:
			return y.(time.Time).Add(time.Duration(int64(time.Hour) * 24 * int64(x.(Int))))
		}

	//TODO: BigInt with Time

	case time.Time:
		switch y.(type) {
		case Int:
			return x.(time.Time).Add(time.Duration(int64(time.Hour) * 24 * int64(y.(Int))))

			//TODO: BigInt
		}

	case u8.Codepoint:
		switch y.(type) {
		case u8.Codepoint:
			return u8.Text{x.(u8.Codepoint), y.(u8.Codepoint)}
		case u8.Text:
			return append(u8.Text{x.(u8.Codepoint)}, y.(u8.Text)...)
		}

	case u8.Text:
		switch y.(type) {
		case u8.Codepoint:
			return append(x.(u8.Text), y.(u8.Codepoint))
		case u8.Text:
			return u8.Text(u8.Sur(x.(u8.Text)...) + u8.Sur(y.(u8.Text)...))
		}

	//TODO: CharClass, Regex, Parsec, Void, Func

	case Slice:
		switch y.(type) {
		case Slice:
			return Slice(append([]Any(x.(Slice)), []Any(y.(Slice))...))
		}

	case *Map:
		switch y.(type) {
		case *Map:
			ax, ay := x.(*Map), y.(*Map)
			az := ax.Merge(ay)
			return az
		}
	}

	panic(fmt.Sprintf("Plus: incompatible types %T and %T", x, y))
}

/*
Negate returns the negative of a number.
An Int is promoted to BigInt if it overflows.
For Codepoint or CharClass, it returns the CharClass that matches the complement.
*/
func Negate(x Any) Any {
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

	//TODO: Regex, Parsec

	//TODO: Text, Void, Time, Slice, Map, Func ???

	default:
		panic(fmt.Sprintf("Negate: invalid type %T", x))
	}
}

/*
Minus negates the 2nd arg and adds the two together.
*/
func Minus(x, y Any) Any {
	x, y = widen(x), widen(y)

	switch x.(type) {
	case Slice:
		switch y.(type) {
		case Slice:
			ax := x.(Slice)
			ay := y.(Slice)
			ns := InitSlice()
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
			panic(fmt.Sprintf("Minus: incompatible types %T and %T", x, y))
		}

	//TODO: case Map with Map

	default:
		return rewidth(Plus(x, Negate(y)))
	}
}

/*
Complement returns the bitwise complement of a number.
*/
func Complement(x Any) Any {
	x = widen(x)
	switch x.(type) {
	case nil:
		return nil
	case bool:
		return !x.(bool)
	case Int:
		return ^x.(Int)

	//TODO: any others ???

	default:
		panic(fmt.Sprintf("Complement: invalid type %T", x))
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
func Mult(x, y Any) Any {
	x, y = widen(x), widen(y)
	if numericTypes(x, y) {
		x, y = y, x
	}

	switch x.(type) {
	case nil:
		switch y.(type) {
		case nil, bool, Int, BigInt, BigRat, Float, BigFloat, Complex, BigComplex, Infinity:
			return nil
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
		}

	case Int:
		ax := x.(Int)
		switch y.(type) {
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
			return rewidth(BigRat(*big.NewRat(0, 1).Mul(big.NewRat(int64(ax), 1), &ay)))
		case Float:
			return rewidth(Float(ax) * y.(Float))
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return rewidth(BigFloat(*big.NewFloat(0).Mul(big.NewFloat(float64(ax)), &ay)))
		case Complex:
			return rewidth(Complex(complex(float64(ax), 0)) * y.(Complex))
		case BigComplex:
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			pr := *big.NewFloat(0).Mul(big.NewFloat(float64(ax)), &yr)
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			pi := *big.NewFloat(0).Mul(big.NewFloat(float64(ax)), &yi)
			return rewidth(BigComplex{pr, pi})
		case Infinity:
			if ax == 0 {
				return nil
			} else {
				return y
			}
		}

	case BigInt:
		ax := big.Int(x.(BigInt))
		switch y.(type) {
		case BigInt:
			ay := big.Int(y.(BigInt))
			return BigInt(*big.NewInt(0).Mul(&ax, &ay))
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return rewidth(BigRat(*big.NewRat(0, 1).Mul(big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)), &ay)))
		case Float:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return rewidth(Float(xf) * y.(Float))
		case BigFloat:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			xr := *big.NewFloat(xf)
			ay := big.Float(y.(BigFloat))
			return rewidth(BigFloat(*big.NewFloat(0).Mul(&xr, &ay)))
		case Complex:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			return rewidth(Complex(complex(xf, 0)) * y.(Complex))
		case BigComplex:
			xf, _ := big.NewRat(0, 1).SetFrac(&ax, big.NewInt(1)).Float64()
			xr := *big.NewFloat(xf)
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			return rewidth(BigComplex{*big.NewFloat(0).Mul(&xr, &yr), *big.NewFloat(0).Mul(&xr, &yi)})
		case Infinity:
			if big.NewInt(0).Cmp(&ax) == 0 {
				return nil
			} else {
				return y
			}
		}

	case BigRat:
		ax := big.Rat(x.(BigRat))
		switch y.(type) {
		case BigRat:
			ay := big.Rat(y.(BigRat))
			return rewidth(BigRat(*big.NewRat(0, 1).Mul(&ax, &ay)))
		case Float:
			xf, _ := ax.Float64()
			return rewidth(Float(xf) * y.(Float))
		case BigFloat:
			xf, _ := ax.Float64()
			xr := *big.NewFloat(xf)
			ay := big.Float(y.(BigFloat))
			return rewidth(BigFloat(*big.NewFloat(1.0).Mul(&xr, &ay)))
		case Complex:
			xf, _ := ax.Float64()
			return rewidth(Complex(complex(xf, 0)) * y.(Complex))
		case BigComplex:
			xf, _ := ax.Float64()
			xr := *big.NewFloat(xf)
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			return rewidth(BigComplex{*big.NewFloat(1.0).Mul(&xr, &yr), *big.NewFloat(1.0).Mul(&xr, &yi)})
		case Infinity:
			if big.NewRat(0, 1).Cmp(&ax) == 0 {
				return nil
			} else {
				return y
			}
		}

	case Float:
		ax := x.(Float)
		switch y.(type) {
		case Float:
			return rewidth(ax * y.(Float))
		case BigFloat:
			xf := *big.NewFloat(float64(ax))
			yf := big.Float(y.(BigFloat))
			return rewidth(BigFloat(*big.NewFloat(1.0).Mul(&xf, &yf)))
		case Complex:
			return rewidth(Complex(complex(ax, 0)) * y.(Complex))
		case BigComplex:
			xf := *big.NewFloat(float64(ax))
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			return rewidth(BigComplex{*big.NewFloat(1.0).Mul(&xf, &yr), *big.NewFloat(1.0).Mul(&xf, &yi)})
		case Infinity:
			if x.(Float) == 0.0 {
				return nil
			} else {
				return y
			}
		}

	case BigFloat:
		ax := big.Float(x.(BigFloat))
		switch y.(type) {
		case BigFloat:
			ay := big.Float(y.(BigFloat))
			return rewidth(BigFloat(*big.NewFloat(1.0).Mul(&ax, &ay)))
		case Complex:
			yr := *big.NewFloat(real(complex128(y.(Complex))))
			yi := *big.NewFloat(imag(complex128(y.(Complex))))
			return rewidth(BigComplex{*big.NewFloat(1.0).Mul(&ax, &yr), *big.NewFloat(1.0).Mul(&ax, &yi)})
		case BigComplex:
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			return rewidth(BigComplex{*big.NewFloat(1.0).Mul(&ax, &yr), *big.NewFloat(1.0).Mul(&ax, &yi)})
		case Infinity:
			if big.NewFloat(0.0).Cmp(&ax) == 0 {
				return nil
			} else {
				return y
			}
		}

	case Complex:
		ax := x.(Complex)
		switch y.(type) {
		case Complex:
			return rewidth(ax * y.(Complex))
		case BigComplex:
			xr := *big.NewFloat(real(complex128(ax)))
			xi := *big.NewFloat(imag(complex128(ax)))
			yr := big.Float(BigFloat(y.(BigComplex).Re))
			yi := big.Float(BigFloat(y.(BigComplex).Im))
			realPart := *big.NewFloat(1.0).Sub(big.NewFloat(1.0).Mul(&xr, &yr), big.NewFloat(1.0).Mul(&xi, &yi))
			imagPart := *big.NewFloat(1.0).Add(big.NewFloat(1.0).Mul(&xi, &yr), big.NewFloat(1.0).Mul(&xr, &yi))
			return rewidth(BigComplex{realPart, imagPart})
		case Infinity:
			if ax == 0.0+0.0i {
				return nil
			} else {
				return y
			}
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
			return rewidth(BigComplex{realPart, imagPart})
		case Infinity:
			xr := big.Float(BigFloat(ax.Re))
			xi := big.Float(BigFloat(ax.Im))
			if big.NewFloat(0.0).Cmp(&xr) == 0 && big.NewFloat(0.0).Cmp(&xi) == 0 {
				return nil
			} else {
				return y
			}
		}

	case Infinity:
		switch y.(type) {
		case Infinity:
			return y
		}
	}

	//- - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	switch x.(type) {
	case Int:
		ax := x.(Int)
		switch y.(type) {
		case u8.Codepoint, u8.Text:
			return RegexRepeat(y, Pair{ax, ax}, false) //FIX: return repeated Text
		case CharClass, Regex:
			return RegexRepeat(y, Pair{ax, ax}, false)
		case Parsec:
			return ParserRepeat(y, x)
		case Func:
			ay := y.(func(...Any) Any)
			return ParserRepeat(Fwd(ay), x)
		}

	case Pair:
		switch y.(type) {
		case u8.Codepoint, u8.Text, CharClass, Regex:
			return RegexRepeat(y, x, false)
		case Parsec:
			return ParserRepeat(y, x)
		case Func:
			ay := y.(func(...Any) Any)
			return ParserRepeat(Fwd(ay), x)
		}

	case u8.Codepoint, u8.Text:
		switch y.(type) {
		case Int:
			return RegexRepeat(x, y.(Int), true) //FIX: return repeated Text
		case Pair:
			return RegexRepeat(x, y.(Pair), true)
		}

	case CharClass, Regex:
		switch y.(type) {
		case Int:
			return RegexRepeat(x, y.(Int), true)
		case Pair:
			return RegexRepeat(x, y.(Pair), true)
		}

	case Parsec:
		switch y.(type) {
		case Int:
		case Parsec:
			return ParserRepeat(x, y)
		case Pair:
			return ParserRepeat(x, y)
		}

	case Func:
		ax := x.(func(...Any) Any)
		switch y.(type) {
		case Int:
			return ParserRepeat(Fwd(ax), y)
		case Pair:
			return ParserRepeat(Fwd(ax), y)
		}
	}

	panic(fmt.Sprintf("Mult: incompatible types %T and %T", x, y))
}

/*
Invert returns the result of dividing the numeric into 1.
Division by 0 returns infinity.
*/
func Invert(x Any) Any {
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
		return rewidth(BigRat(*big.NewRat(1, int64(x.(Int)))))
	case BigInt:
		if reflect.DeepEqual(x.(BigInt), BigInt(*big.NewInt(0))) {
			return Inf
		}
		ax := big.Int(x.(BigInt))
		return rewidth(BigRat(*big.NewRat(0, 1).SetFrac(big.NewInt(1), &ax))) //FIX: make copy of denom first ???
	case BigRat:
		if reflect.DeepEqual(x.(BigRat), BigRat(*big.NewRat(0, 1))) {
			return Inf
		}
		ax := big.Rat(x.(BigRat))
		return rewidth(BigRat(*big.NewRat(0, 1).Inv(&ax)))
	case Float:
		if x.(Float) == 0 {
			return Inf
		}
		return rewidth(Float(1 / float64(x.(Float))))
	case BigFloat:
		ax := big.Float(x.(BigFloat))
		if big.NewFloat(0.0).Cmp(&ax) == 0 {
			return Inf
		}
		return rewidth(BigFloat(*big.NewFloat(1.0).Quo(big.NewFloat(1.0), &ax)))
	case Complex:
		if x.(Complex) == 0 {
			return Inf
		}
		return rewidth(Complex(1 / complex128(x.(Complex))))
	case BigComplex:
		xr := big.Float(BigFloat(x.(BigComplex).Re))
		xi := big.Float(BigFloat(x.(BigComplex).Im))
		if big.NewFloat(0.0).Cmp(&xr) == 0 && big.NewFloat(0.0).Cmp(&xi) == 0 {
			return Inf
		}
		mod := big.NewFloat(0.0).Add(big.NewFloat(0.0).Mul(&xr, &xr), big.NewFloat(0.0).Mul(&xi, &xi))
		cr := *big.NewFloat(1.0).Quo(&xr, mod)
		ci := *big.NewFloat(1.0).Quo(big.NewFloat(1.0).Neg(&xi), mod)
		return rewidth(BigComplex{cr, ci})
	case Infinity:
		return 0
	default:
		panic(fmt.Sprintf("Invert: invalid type %T", x))
	}
}

/*
Divide inverts the second argument and multiplies the two together.
Division by nil or false gives infinity.
Divide result is of the same type as whichever argument is higher in the numeric hierarchy.
A division by a Int or BigInt promotes to a BigRat.
*/
func Divide(x, y Any) Any {
	return rewidth(Mult(x, Invert(y)))
}

/*
Mod takes two Int, BigInt, or Infinite args,
and returns the modulus.
Anything mod 0 is infinity;
anything mod infinity is unchanged.

//TODO: both 0%0 and inf%inf should return nil
//TODO: put in logic for true and false
*/
func Mod(x, y Any) Any {
	x, y = widen(x), widen(y)
	switch x.(type) {
	case Int:
		ax := x.(Int)
		switch y.(type) {
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
		}

	case BigInt:
		ax := big.Int(x.(BigInt))
		switch y.(type) {
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
		}
	case Infinity:
		switch y.(type) {
		case Int, BigInt, Infinity:
			return x
		}
	}

	//- - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	switch y.(type) {
	case u8.Text:
		ay := u8.Surr(y.(u8.Text))
		return fmt.Sprintf("%"+ay, x)
	}

	panic(fmt.Sprintf("Mod: incompatible types %T and %T", x, y))
}

//==============================================================================

/*
Not returns the boolean-not of the parameter converted to a boolean.
For Codepoint or CharClass, it returns the CharClass that matches the complement.
*/
func Not(x Any) Any {
	x = widen(x)
	switch x.(type) {
	case u8.Codepoint:
		return CharClass("[^" + u8.Sur(x.(u8.Codepoint)) + "]")
	case CharClass:
		mid, neg := x.(CharClass).unwrap()
		if neg {
			return CharClass("[" + mid + "]")
		} else {
			return CharClass("[^" + mid + "]")
		}
	default:
		return !toBool(x)
	}
}

/*
And returns the boolean-and of parameter a and the value returned from calling parameter b
as a function. Parameter b is only called if a is true.
*/
func And(x Any, y func() Any) bool {
	return toBool(x) && toBool(y())
}

/*
Or returns the boolean-or of parameter a and the value returned from calling parameter b
as a function. Parameter b is only called if a is false.
*/
func Or(x Any, y func() Any) bool {
	return toBool(x) || toBool(y())
}

//==============================================================================

/*
Alt accepts a Codepoint, Text, CharClass, Regex, or Parser for each parameter,
and returns a CharClass, Regex, or Parser using ordered alternation on the inputs.

The args convert such that:
  2 Codepoints become a CharClass
  a CharClass or Codepoint with a CharClass may remain a CharClass, but might become a Regex
  anything with a Parser becomes a Parser
  anything else with Text or a Regex becomes a Regex
*/
func Alt(x, y Any) Any {
	switch x.(type) {

	case u8.Codepoint:
		switch y.(type) {
		case u8.Codepoint:
			return CharClass("[" + u8.Sur(x.(u8.Codepoint)) + u8.Sur(y.(u8.Codepoint)) + "]")
		case u8.Text:
			return Regex(u8.Sur(x.(u8.Codepoint)) + `|(?:\Q` + u8.Sur(y.(u8.Text)...) + `\E)`)
		case CharClass:
			ymid, yneg := y.(CharClass).unwrap()
			if yneg {
				return Regex(u8.Sur(x.(u8.Codepoint)) + "|" + string(y.(CharClass)))
			} else {
				return CharClass("[" + u8.Sur(x.(u8.Codepoint)) + ymid + "]")
			}
		case Regex:
			return Regex(u8.Sur(x.(u8.Codepoint)) + "|(?:" + string(y.(Regex)) + ")")
		case Parsec:
			return AltParse(ToParser(x), y.(Parsec))
		case Func:
			ay := y.(func(...Any) Any)
			return AltParse(ToParser(x), Fwd(ay))
		}

	case u8.Text:
		switch y.(type) {
		case u8.Codepoint:
			return Regex(`(?:\Q` + u8.Sur(x.(u8.Text)...) + `\E)|` + u8.Sur(y.(u8.Codepoint)))
		case u8.Text:
			return Regex(`(?:\Q` + u8.Sur(x.(u8.Text)...) + `\E)|(?:\Q` + u8.Sur(y.(u8.Text)...) + `\E)`)
		case CharClass:
			return Regex(`(?:\Q` + u8.Sur(x.(u8.Text)...) + `\E)|` + string(y.(CharClass)))
		case Regex:
			return Regex(`(?:\Q` + u8.Sur(x.(u8.Text)...) + `\E)|(?:` + string(y.(Regex)) + ")")
		case Parsec:
			return AltParse(ToParser(x), ToParser(y))
		case Func:
			ay := y.(func(...Any) Any)
			return AltParse(ToParser(x), Fwd(ay))
		}

	case CharClass:
		switch y.(type) {
		case u8.Codepoint:
			xmid, xneg := x.(CharClass).unwrap()
			if xneg {
				return Regex(string(x.(CharClass)) + "|" + u8.Sur(y.(u8.Codepoint)))
			} else {
				return CharClass("[" + xmid + u8.Sur(y.(u8.Codepoint)) + "]")
			}
		case u8.Text:
			return Regex(string(x.(CharClass)) + `|(?:\Q` + u8.Sur(y.(u8.Text)...) + `\E)`)
		case CharClass:
			xmid, xneg := x.(CharClass).unwrap()
			ymid, yneg := y.(CharClass).unwrap()
			if xneg == yneg {
				if xneg {
					return CharClass("[^" + xmid + ymid + "]")
				} else {
					return CharClass("[" + xmid + ymid + "]")
				}
			} else {
				return Regex(string(x.(CharClass)) + "|" + string(y.(CharClass)))
			}
		case Regex:
			return Regex(string(x.(CharClass)) + "|(?:" + string(y.(Regex)) + ")")
		case Parsec:
			return AltParse(ToParser(x), ToParser(y))
		case Func:
			ay := y.(func(...Any) Any)
			return AltParse(ToParser(x), Fwd(ay))
		}

	case Regex:
		switch y.(type) {
		case u8.Codepoint:
			return Regex("(?:" + string(x.(Regex)) + ")|" + u8.Sur(y.(u8.Codepoint)))
		case u8.Text:
			return Regex("(?:" + string(x.(Regex)) + `)|(?:\Q` + u8.Sur(y.(u8.Text)...) + `\E)`)
		case CharClass:
			return Regex("(?:" + string(x.(Regex)) + ")|" + string(y.(CharClass)))
		case Regex:
			return Regex("(?:" + string(x.(Regex)) + ")|(?:" + string(y.(Regex)) + ")")
		case Parsec:
			return AltParse(ToParser(x), y.(Parsec))
		case Func:
			ay := y.(func(...Any) Any)
			return AltParse(ToParser(x), Fwd(ay))
		}

	case Parsec:
		switch y.(type) {
		case u8.Codepoint, u8.Text, CharClass, Regex:
			return AltParse(x.(Parsec), ToParser(y))
		case Parsec:
			return AltParse(x.(Parsec), y.(Parsec))
		case Func:
			ay := y.(func(...Any) Any)
			return AltParse(x.(Parsec), Fwd(ay))
		}

	case Func:
		ax := x.(func(...Any) Any)
		xfwd := Fwd(ax)
		switch y.(type) {
		case u8.Codepoint:
			return AltParse(xfwd, Symbol(y.(u8.Codepoint)))
		case u8.Text:
			return AltParse(xfwd, Token(y.(u8.Text)))
		case Parsec:
			return AltParse(xfwd, y.(Parsec))
		case Func:
			ay := y.(func(...Any) Any)
			return AltParse(xfwd, Fwd(ay))
		}
	}

	panic(fmt.Sprintf("Alt: incompatible types %T and %T", x, y))
}

//==============================================================================

/*
Seq accepts a Codepoint, Text, CharClass, or Regex for each parameter,
and returns Text concatenating the inputs, or a Regex matching the inputs in sequence.

The args convert such that:
  2 Codepoints become a Text, as does a Codepoint combined with a Text
  2 Texts remain a Text
  2 CharClasses become a Regex, as does a Codepoint or Text combined with a CharClass
  2 Regexes remain a Regex, as does a Codepoint, Text, or CharClass combined with a Regex
*/
func Seq(x, y Any) Any {
	switch x.(type) {
	case u8.Codepoint:
		switch y.(type) {
		case u8.Codepoint:
			return u8.Text{x.(u8.Codepoint), y.(u8.Codepoint)}
		case u8.Text:
			return append(u8.Text{x.(u8.Codepoint)}, y.(u8.Text)...)
		case CharClass:
			return Regex(u8.Sur(x.(u8.Codepoint)) + string(y.(CharClass)))
		case Regex:
			return Regex(u8.Sur(x.(u8.Codepoint)) + "(?:" + string(y.(Regex)) + ")")
		case Parsec, Func:
			return RightSeq(x, y)
		}

	case u8.Text:
		switch y.(type) {
		case u8.Codepoint:
			return append(x.(u8.Text), y.(u8.Codepoint))
		case u8.Text:
			return u8.Text(u8.Sur(x.(u8.Text)...) + u8.Sur(y.(u8.Text)...))
		case CharClass:
			return Regex(u8.Sur(x.(u8.Text)...) + string(y.(CharClass)))
		case Regex:
			return Regex(u8.Sur(x.(u8.Text)...) + "(?:" + string(y.(Regex)) + ")")
		case Parsec, Func:
			return RightSeq(x, y)
		}

	case CharClass:
		switch y.(type) {
		case u8.Codepoint:
			return Regex(string(x.(CharClass)) + u8.Sur(y.(u8.Codepoint)))
		case u8.Text:
			return Regex(string(x.(CharClass)) + u8.Sur(y.(u8.Text)...))
		case CharClass:
			return Regex(string(x.(CharClass)) + string(y.(CharClass)))
		case Regex:
			return Regex(string(x.(CharClass)) + "(?:" + string(y.(Regex)) + ")")
		case Parsec, Func:
			return RightSeq(x, y)
		}

	case Regex:
		xx := "(?:" + string(x.(Regex)) + ")"
		switch y.(type) {
		case u8.Codepoint:
			return Regex(xx + u8.Sur(y.(u8.Codepoint)))
		case u8.Text:
			return Regex(xx + u8.Sur(y.(u8.Text)...))
		case CharClass:
			return Regex(xx + string(y.(CharClass)))
		case Regex:
			return Regex(xx + "(?:" + string(y.(Regex)) + ")")
		case Parsec, Func:
			return RightSeq(x, y)
		}

	case Parsec, Func:
		return RightSeq(x, y)
	}

	panic(fmt.Sprintf("Seq: incompatible types %T and %T", x, y))
}

//==============================================================================

/*
LeftSeq calls SeqLeft with the 2 parsers supplied as args.

TODO: Write example function.
*/
func LeftSeq(x, y Any) Any {
	return directionalAnd(SeqLeft, "SeqLeft", x, y)
}

/*
RightSeq calls SeqRight with the 2 parsers supplied as args.

TODO: Write example function.
*/
func RightSeq(x, y Any) Any {
	return directionalAnd(SeqRight, "SeqRight", x, y)
}

func directionalAnd(
	function func(...Any) Any,
	text string,
	x, y Any,
) Any {
	x, y = widen(x), widen(y)
	switch x.(type) {

	case u8.Codepoint:
		ax := x.(u8.Codepoint)
		switch y.(type) {
		case u8.Codepoint:
			return function(Symbol(ax), Symbol(y.(u8.Codepoint)))
		case u8.Text:
			return function(Symbol(ax), Token(y.(u8.Text)))
		case CharClass, Regex:
			return function(Symbol(ax), ToParser(y))
		case Parsec:
			return function(Symbol(ax), y.(Parsec))
		case Func:
			ay := y.(func(...Any) Any)
			return function(Symbol(ax), Fwd(ay))
		}

	case u8.Text:
		ax := x.(u8.Text)
		switch y.(type) {
		case u8.Codepoint:
			return function(Token(ax), Symbol(y.(u8.Codepoint)))
		case u8.Text:
			return function(Token(ax), Token(y.(u8.Text)))
		case CharClass, Regex:
			return function(Token(ax), ToParser(y))
		case Parsec:
			return function(Token(ax), y.(Parsec))
		case Func:
			ay := y.(func(...Any) Any)
			return function(Token(ax), Fwd(ay))
		}

	case CharClass, Regex:
		ax := ToParser(x)
		switch y.(type) {
		case u8.Codepoint:
			return function(ax, Symbol(y.(u8.Codepoint)))
		case u8.Text:
			return function(ax, Token(y.(u8.Text)))
		case CharClass, Regex:
			return function(ax, ToParser(y))
		case Parsec:
			return function(ax, y.(Parsec))
		case Func:
			ay := y.(func(...Any) Any)
			return function(ax, Fwd(ay))
		}

	case Parsec:
		ax := x.(Parsec)
		switch y.(type) {
		case u8.Codepoint:
			return function(ax, Symbol(y.(u8.Codepoint)))
		case u8.Text:
			return function(ax, Token(y.(u8.Text)))
		case CharClass, Regex:
			return function(ax, ToParser(y))
		case Parsec:
			return function(ax, y.(Parsec))
		case Func:
			ay := y.(func(...Any) Any)
			return function(ax, Fwd(ay))
		}

	case Func:
		ax := x.(func(...Any) Any)
		xfwd := Fwd(ax)
		switch y.(type) {
		case u8.Codepoint:
			return function(xfwd, Symbol(y.(u8.Codepoint)))
		case u8.Text:
			return function(xfwd, Token(y.(u8.Text)))
		case CharClass, Regex:
			return function(xfwd, ToParser(y))
		case Parsec:
			return function(xfwd, y.(Parsec))
		case Func:
			ay := y.(func(...Any) Any)
			return function(xfwd, Fwd(ay))
		}
	}

	panic(fmt.Sprintf(text+": incompatible types %T and %T", x, y))
}

//==============================================================================

/*
Xor takes two Ints and calls Go's ^ operator on them.
*/
func Xor(x, y Any) Any {
	x, y = widen(x), widen(y)
	switch x.(type) {
	case Int:
		switch y.(type) {
		case Int:
			return Int(int64(x.(Int)) ^ int64(y.(Int)))
		}
	}

	panic(fmt.Sprintf("Xor: incompatible types %T and %T", x, y))
}

/*
SeqXor takes two Ints and calls Go's &^ operator on them.
*/
func SeqXor(x, y Any) Any {
	x, y = widen(x), widen(y)
	switch x.(type) {
	case Int:
		switch y.(type) {
		case Int:
			return Int(int64(x.(Int)) &^ int64(y.(Int)))
		}
	}

	panic(fmt.Sprintf("SeqXor: incompatible types %T and %T", x, y))
}

//==============================================================================

/*
LeftShift calls SeqLeft with the 2 parsers supplied as args.
*/
func LeftShift(x, y Any) Any {
	x, y = widen(x), widen(y)
	switch x.(type) {
	case Slice:
		return Slice(append([]Any(x.(Slice)), y))

	case *Map:
		ax, ay := x.(*Map), y.(MapEntry)
		ax.Add(ay.Key(), ay.Val())
		return ax
	}

	panic(fmt.Sprintf("LeftShift: incompatible types %T and %T", x, y))
}

/*
The the 1st arg must be some Text.
If the 2nd arg is a Regex, CharClass, Text, or Codepoint,
it is used to match the Text.
It will return a slice of a slice of RegexGroupMatch structs.
The outer slice contains an inner slice for each match.
The 0-index in each inner slice will contain the full match details.
The subsequent indexes in each inner slice will contain the matches for each capturing group.

If the 2nd arg is a Parser, calls ParseText with it on the Text.

TODO: Write example function.
*/
func RightShift(x, y Any) Any {
	x, y = widen(x), widen(y)
	switch x.(type) {
	case u8.Text:
		var p string
		switch y := y.(type) {
		case u8.Codepoint:
			p = string(y)
		case u8.Text:
			//p = `\Q` + u8.Sur(y.(u8.Text)...) + `\E`
			p = u8.Sur(y...)
		case CharClass:
			p = string(y)
		case Regex:
			p = string(y)
		case Parsec:
			r, _ := ParseItem(y, x.(u8.Text))
			return r
		case Func:
			r, _ := ParseItem(Fwd(y).(Parsec), x.(u8.Text))
			return r
		}

		ax := u8.Surr(x.(u8.Text))
		re := regexp.MustCompile(p)
		arr := re.FindAllStringSubmatchIndex(ax, -1)
		names := re.SubexpNames()
		numGrTot := re.NumSubexp() + 1
		rgm := make([][]RegexMatch, len(arr))
		for n, _ := range arr {
			rgm[n] = make([]RegexMatch, numGrTot)
			for m := 0; m < numGrTot; m++ {
				a, b := arr[n][2*m], arr[n][2*m+1]
				var ss string
				if a != -1 && b != -1 {
					ss = ax[a:b]
				}
				rgm[n][m] = RegexMatch{names[m], a, b, ss}
			}
		}
		return rgm

	//TODO: case *Map

	case Slice:
		ax := x.(Slice)
		ns := InitSlice()
		for _, a := range ax {
			if !IsEqual(a, y) {
				_ = LeftShiftAssign(&ns, a)
			}
		}
		return ns.(Slice)
	}

	panic(fmt.Sprintf("RightShift: incompatible types %T and %T", x, y))
}

//==============================================================================

/*
Assign makes the first arg point to the second arg.
*/
func Assign(pp *Any, k Any) *Any {
	k = widen(k)
	*pp = k
	return pp
}

//TODO: MultipleAssign ???

//==============================================================================

/*
PlusAssign adds the value/s to a vector or the key-value pair/s to a map,
otherwise calls Plus and Assign on the args.
*/
func PlusAssign(pp *Any, k Any) *Any {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	case Slice:
		*pp = Plus(*pp, k)
		return pp
	case Map:
		*pp = Plus(*pp, k)
		return pp

	default:
		*pp = Plus(*pp, k)
		return pp
	}

	//panic(fmt.Sprintf("Assign: invalid type %T", k))
}

/*
MinusAssign deletes an entry from a Map, otherwise calls Minus and Assign on the args.
*/
func MinusAssign(pp *Any, k Any) *Any {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {

	// TODO: add Slice functionality

	default:
		*pp = Minus(*pp, k)
		return pp
	}

	//panic(fmt.Sprintf("MinusAssign: invalid type %T", k))
}

/*
MultAssign calls Mult and Assign on the args.
*/
func MultAssign(pp *Any, k Any) *Any {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	default:
		*pp = Mult(*pp, k)
		return pp
	}

	//panic(fmt.Sprintf("MultAssign: invalid type %T", k))
}

/*
DivideAssign calls Divide and Assign on the args.
*/
func DivideAssign(pp *Any, k Any) *Any {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	default:
		*pp = Divide(*pp, k)
		return pp
	}

	//panic(fmt.Sprintf("DivideAssign: invalid type %T", k))
}

/*
ModAssign calls Mod and Assign on the args.
*/
func ModAssign(pp *Any, k Any) *Any {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	default:
		*pp = Mod(*pp, k)
		return pp
	}

	//panic(fmt.Sprintf("ModAssign: invalid type %T", k))
}

/*
SeqAssign calls Seq and Assign on the args.
*/
func SeqAssign(pp *Any, k Any) *Any {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	default:
		*pp = Seq(*pp, k)
		return pp
	}

	//panic(fmt.Sprintf("SeqAssign: invalid type %T", k))
}

/*
AltAssign calls Alt and Assign on the args.
*/
func AltAssign(pp *Any, k Any) *Any {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	default:
		*pp = Alt(*pp, k)
		return pp
	}

	//panic(fmt.Sprintf("AltAssign: invalid type %T", k))
}

/*
XorAssign calls Xor and Assign on the args.
*/
func XorAssign(pp *Any, k Any) *Any {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	default:
		*pp = Xor(*pp, k)
		return pp
	}

	//panic(fmt.Sprintf("XorAssign: invalid type %T", k))
}

/*
SeqXorAssign calls SeqXor and Assign on the args.
*/
func SeqXorAssign(pp *Any, k Any) *Any {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	default:
		*pp = SeqXor(*pp, k)
		return pp
	}

	//panic(fmt.Sprintf("SeqXorAssign: invalid type %T", k))
}

/*
LeftSeqAssign calls LeftSeq and Assign on the args.
*/
func LeftSeqAssign(pp *Any, k Any) *Any {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	default:
		*pp = LeftSeq(*pp, k)
		return pp
	}

	//panic(fmt.Sprintf("LeftSeqAssign: invalid type %T", k))
}

/*
RightSeqAssign calls RightSeq and Assign on the args.
*/
func RightSeqAssign(pp *Any, k Any) *Any {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	default:
		*pp = RightSeq(*pp, k)
		return pp
	}

	//panic(fmt.Sprintf("RightSeqAssign: invalid type %T", k))
}

/*
LeftShiftAssign calls LeftShift and Assign on the args.
*/
func LeftShiftAssign(pp *Any, k Any) *Any {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	case Slice, *Map:
		*pp = LeftShift(*pp, k)
		return pp

	default:
		*pp = LeftShift(*pp, k)
		return pp
	}

	//panic(fmt.Sprintf("LeftShiftAssign: invalid type %T", k))
}

/*
RightShiftAssign calls RightShift and Assign on the args.
*/
func RightShiftAssign(pp *Any, k Any) *Any {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	case *Map:
		m := (*pp).(*Map)
		m.Delete(k)
		ma := Any(m)
		return &ma

	default:
		*pp = RightShift(*pp, k)
		return pp
	}

	//panic(fmt.Sprintf("RightShiftAssign: invalid type %T", k))
}

//==============================================================================

/*
GetIndex gets the index of the 1st arg as specified by the 2nd,
and maybe also 3rd, arg/s.
*/
func GetIndex(pa *Any, ns ...Any) *Any {
	var ab, ae Any
	if len(ns) > 0 {
		ab = widen(ns[0])
		switch ab.(type) {
		case Int:
			ab = ab.(Int)
		case BigInt:
			ab = ab.(BigInt)
		case Infinity:
			ab = Inf
		default:
			ab = ab
		}
	} else {
		panic(fmt.Sprintf("GetIndex: 1st index must be supplied"))
	}
	if len(ns) > 1 {
		ae = widen(ns[1])
		switch ae.(type) {
		case Int:
			ae = ae.(Int)
		case BigInt:
			ae = ae.(BigInt)
		case Infinity:
			ae = Inf
		default:
			panic("GetIndex: invalid type of 2nd index arg")
		}
	}

	wpa := widen(*pa)
	pa = &wpa

	switch (*pa).(type) {
	case u8.Text:
		b := toInt(ab)
		var e Int
		if ae != nil {
			e = toInt(ae)
		} //else 2nd index not supplied

		s := (*pa).(u8.Text)
		lenS := Int(len(s))
		var is Any
		if ae == nil {
			is = s[b]
		} else if ae == Inf {
			is = s[b:lenS]
		} else { // 0 <= e < maxInt {
			is = s[b:e]
		}
		return &is

	case Slice:
		b := toInt(ab)
		var e Int
		if ae != nil {
			e = toInt(ae)
		} //else 2nd index not supplied

		s := (*pa).(Slice)
		lenS := Int(len(s))
		if ae == nil {
			return &s[b]
		} else if ae == Inf {
			is := Any(s[b:lenS])
			return &is
		} else { // 0 <= e < maxInt {
			is := Any(s[b:e])
			return &is
		}

	case *Map:
		//TODO: ensure only one key given for Maps
		t := (*pa).(*Map)
		n := widen(ns[0])
		an := t.Val(n)
		return &an
	}

	panic(fmt.Sprintf("GetIndex: invalid type %T", *pa))
}

/*
SetIndex sets the index specified by the second arg of the first arg to the third arg.
*/
func SetIndex(pa *Any, n, v Any) *Any {
	n = widen(n)
	v = widen(v)

	switch (*pa).(type) {
	case u8.Text:
		panic("SetIndex: Texts are immutable")

	case Slice:
		switch n.(type) {
		case Int:
			n = n.(Int)
		case BigInt:
			n = n.(BigInt)
		case Infinity:
			n = Inf
		default:
			n = n
		}
		nn := toInt(n)
		[]Any((*pa).(Slice))[nn] = v
		an := []Any((*pa).(Slice))[nn]
		return &an

	case *Map: //TODO: cater for any type of index -- see Slice above
		t := (*pa).(*Map)
		nw := widen(n)
		t.SetVal(nw, v)
		return pa
	}

	panic(fmt.Sprintf("SetIndex: invalid type %T", *pa))
}

//==============================================================================

/*
Incr increments the arg in-place.
*/
func Incr(pn *Any) Any {
	temp := *pn
	switch (*pn).(type) {
	case Int, BigInt:
		*pn = Plus(*pn, 1)
	default:
		panic("Incr: non-integral arg")
	}
	return temp
}

/*
Decr decrements the arg in-place.
*/
func Decr(pn *Any) Any {
	temp := *pn
	switch (*pn).(type) {
	case Int, BigInt:
		*pn = Minus(*pn, 1)
	default:
		panic("Decr: non-integral arg")
	}
	return temp
}

//==============================================================================

/*
Reflect returns reflection object of interface value, or vice versa.
*/
func Reflect(x Any) Any {
	x = widen(x)
	switch x.(type) {
	case reflect.Value:
		return x.(reflect.Value).Interface()
	default:
		return reflect.ValueOf(x)
	}
}

//==============================================================================
