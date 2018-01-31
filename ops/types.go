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
	"strconv"

	u8 "github.com/grolang/gro/utf88"
)

type (
	Int        = int64
	BigInt     big.Int
	BigRat     big.Rat
	Float      = float64
	BigFloat   big.Float
	Complex    = complex128
	BigComplex struct{ Re, Im big.Float }
	Infinity   struct{} //Riemann-style infinity

	Any       = interface{}
	Void      = struct{}
	Pair      struct{ First, Second Any }
	CharClass string
	Regex     string
	Func      = func(...Any) Any
	Slice     []Any
	Map       struct {
		lkp map[Any]Any
		seq []Any
	}
)

var (
	Inf      = Infinity{}
	UseUtf88 bool
)

//==============================================================================
//used by gro parser: Inf, UseUtf88, InitMap, NewPair, MakeText, Runex

func InitSlice(vs ...Any) Any {
	s := make([]Any, 0)
	for _, v := range vs {
		s = append(s, widen(v))
	}
	return Slice(s)
}

/*
InitMap makes a new Map and initializes it to the supplied pair-values.
*/
func InitMap(es ...MapEntry) Any {
	m := NewMap()
	for _, e := range es {
		m.Add(e.Key(), e.Val())
	}
	return m
}

func NewPair(first, second Any) Pair {
	return Pair{widen(first), widen(second)}
}

/*
MakeText accepts a string as input, and returns Text.
*/
func MakeText(x Any) Any {
	return u8.Desur(x.(string))
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

//String of Map returns a string representation of its value.
func (m *Map) String() string {
	str, first := "{", true
	for _, k := range m.seq {
		v := m.lkp[k]
		if first {
			first = false
		} else {
			str = str + ", "
		}
		str = str + fmt.Sprintf("%v: %v", k, v)
	}
	return str + "}"
}

//==============================================================================

/*
MapEntry represents a key-value pair for a Map.
*/
type MapEntry interface {
	Key() Any
	Val() Any
	fmt.Stringer
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

func (r Pair) String() string {
	return fmt.Sprintf("%v: %v", r.First, r.Second)
}

//Key returns the first field as a map key.
func (r Pair) Key() Any {
	return r.First
}

//Key returns the second field as a map value.
func (r Pair) Val() Any {
	return r.Second
}

/*
From returns the first field as an Int.
It returns 0 if it's negative.
It returns the maximum Int if it's a BigInt greater than that.
*/
func (r Pair) From() Int {
	return toInt(r.First)
}

/*
To returns the second field as an Int.
It returns 0 if it's negative.
It returns the maximum Int if it's infinity or a BigInt greater than that.
*/
func (r Pair) To() Int {
	return toInt(r.Second)
}

/*
IsToInf returns true if the second field is infinity, otherwise false.
*/
func (r Pair) IsToInf() bool {
	return r.Second == Inf
}

func toInt(r Any) Int {
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
		panic(fmt.Sprintf("toInt: invalid type %T", r))
	}
}

//==============================================================================

/*
toBool converts nil, Void, false, zeroes, and empty utf88.Text to false,
and all other values to true.
*/
func toBool(x Any) bool {
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

	case Void:
		return false

	case u8.Text:
		return len(x.(u8.Text)) > 0

	case Slice:
		return len(x.(Slice)) > 0

	case *Map:
		return len(x.(*Map).lkp) > 0

	//TODO: Codepoint, CharClass, Regex, Parsec, Time, Func ???

	default:
		return true
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
//TODO: change output to []Any to cater for multiple results
//TODO: add tests
*/
func Power(x, y Any) Any {
	x, y = widen(x), widen(y)
	switch x.(type) {
	case Int:
		switch y.(type) {
		case bool:
			if y.(bool) {
				return x.(Int)
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
		}

	case BigInt:
		switch y.(type) {
		case bool:
			if y.(bool) {
				return x.(BigInt)
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
		}

	case BigRat:
		switch y.(type) {
		case bool:
			if y.(bool) {
				return x.(BigRat)
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
		}

	case Float:
		switch y.(type) {
		case bool:
			if y.(bool) {
				return x.(Float)
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
		}

	case Complex:
		switch y.(type) {
		case bool:
			if y.(bool) {
				return x.(Complex)
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
		}
	}

	panic(fmt.Sprintf("Power: incompatible types %T and %T", x, y))
}

//==============================================================================

/*
RegexMatch is a struct returned from a regex match to indicate the results
of one of the capturing groups, or of the whole match if there's no groups.
Name is the name of the group in the regex.
Start and End are the starting and ending positions in the regex.
Match is the result of the match.
*/
type RegexMatch struct {
	Name       string
	Start, End int
	Match      string
}

func (x CharClass) unwrap() (string, bool) {
	mid := string(x)[1 : len(x)-1]
	neg := false
	if mid[0] == '^' {
		neg = true
		mid = mid[1:]
	}
	return mid, neg
}

//==============================================================================

/*
ToRegex converts a Codepoint, Text, or CharClass into a Regex.
*/
func ToRegex(x Any) Regex {
	switch x.(type) {
	case u8.Codepoint:
		return Regex(u8.Sur(x.(u8.Codepoint)))
	case u8.Text:
		return Regex(`(?:\Q` + u8.Sur(x.(u8.Text)...) + `\E)`)
	case CharClass:
		return Regex(string(x.(CharClass)))
	case Regex:
		return Regex("(?:" + string(x.(Regex)) + ")")
	}

	panic(fmt.Sprintf("ToRegex: invalid type %T", x))
}

/*
ToParser converts a Codepoint, Text, CharClass, or Regex into a Parser.
*/
func ToParser(x Any) Parsec {
	switch x.(type) {
	case u8.Codepoint:
		f := func(c Any) Any {
			return c.(u8.Text)[0]
		}
		return Apply(f, Token([]u8.Codepoint{x.(u8.Codepoint)}))
	case u8.Text:
		return Token(x.(u8.Text))
	case CharClass:
		f := func(c Any) Any {
			return c.(u8.Text)[0]
		}
		return Apply(f, Regexp(string(x.(CharClass))))
	case Regex:
		return Regexp(string(x.(Regex)))
	case Parsec:
		return x.(Parsec)
	case Func:
		for {
			if z, ok := x.(func(...Any) Any); ok {
				x = z()
			} else {
				break
			}
		}
		return ToParser(x)
	}

	panic(fmt.Sprintf("ToParser: invalid type %T", x))
}

//==============================================================================

/*
Runex accepts either a rune or a string as input,
and returns either a Codepoint, a CharClass, or a Regex.

Runex string input merges the formats for a rune constant and a regexp character class
into a single format. A rune constant is embedded between single quotes in Go,
whereas a regexp character class is embedded between square brackets inside
a regexp literal itself encoded as a Go string between double quotes or backticks.

The input string can take all UTF-88 compatible codepoints and codepoint ranges,
and converts them into the relevant regex.

A Go rune can be any arbitrary Unicode codepoint:
  except newline
  a backslash-escaped character \a \b \e \f \n \r \t \v \' \\
  a hex up to \xff
  an octal up to \377
  a 4-byte unicode point e.g. \u159d
  an 8-byte unicode point e.g. \U13579bdf

A Go rexexp char class can be:
  a single char or sequence of them
  puncs may be backslashed but some must be, i.e. first ^, each ], each \, each - except last
  can include ranges using hyphen e.g. a-zA-Z0-9
  first char can be ^ to negate
  \a \f \n \r \t \v
  an octal up to \377
  a hex up to \xff
  a hex of any length \x{13579bdf}
  Perl char classes \d \D \s \S \w \W
  ASCII char classes [:word:] or [:^word:] or ^[:word:]
  Unicode char classes \pL or \PL
  Unicode char classes \p{Greek} or \P{Greek}

TODO: write example function
*/
func Runex(x Any) Any {
	charClass := `(?:\\[dDsSwW])|` + //Perl char classes \d \D \s \S \w \W
		`(?:\\(?:p|P)(?:[A-Za-z]|{[A-Za-z0-9_]+}))|` + //Unicode char classes \pL or \PL or \p{Greek} or \P{Greek}
		`(?:\[:^?[a-z]+:\])` //ASCII char classes [:word:] or [:^word:]

	runeItem := `(?:\\x[0-9A-Fa-f]{2})|` + //hex_byte_value
		`(?:\\x{[0-9A-Fa-f_]+})|` + //hex byte value between curlies
		`(?:\\u[0-9A-Fa-f]{4})|` + //little_u_value
		`(?:\\U[0-9A-Fa-f]{8})|` + //big_u_value
		`(?:\\[0-7]{3})|` + //octal_byte_value
		`(?:\\[abefnrtv\\\]\-\^'])|` + //escaped_char
		`(?:[^\n\\])` //unicode_char
	runeRangePattern := `(?:(?:(` + runeItem + `)(?:-(` + runeItem + `))?)|(` + charClass + `))`

	switch x.(type) {
	case rune:
		return u8.Codepoint(x.(rune))

	case string:
		s := x.(string)
		if r, e := regexp.Compile(runeRangePattern); e == nil && r.MatchString(s) {
			var m, n string
			ass := r.FindAllStringSubmatchIndex(s, -1) //0,1: whole regexp; 2,3: initial; 4,5: end of range; 6,7: char class

			//TODO: make sure regex indexes cover entire string
			if len(ass) == 1 && ass[0][4] == -1 && ass[0][6] == -1 {
				as := ass[0]
				m = s[as[2]:as[3]]
				cp, _ := convertToCp(m)
				return cp

				//TODO: make sure all regex indexes cover entire string
			} else {
				for _, as := range ass {
					if as[6] == -1 && as[4] == -1 {
						r, _ := convertForCC(s[as[2]:as[3]])
						m += r
					} else if as[6] == -1 {
						r, rr, _ := convertForCC2(s[as[2]:as[3]], s[as[4]:as[5]])
						m += r
						n += rr
					} else {
						m = m + s[as[6]:as[7]]
					}
				}
				if n == "" {
					return CharClass("[" + m + "]")
				} else if m == "" {
					return Regex(n[1:]) //take off initial "|"
				} else {
					return Regex("[" + m + "]" + n)
				}
			}

		} else if len(s) == 1 {
			return u8.Codepoint(rune(s[0]))

		} else {
			panic("Runex: invalid string arg, should evaluate to rune constant or regexp char class; " + e.Error())
		}

	default:
		panic(fmt.Sprintf("Runex: invalid input type %T", x))
	}
}

func convertToCp(m string) (cp u8.Codepoint, ok bool) {
	var controlCharLookup = map[string]rune{
		`\a`: '\a',
		`\b`: '\x08',
		`\e`: '\x1b',
		`\f`: '\f',
		`\n`: '\n',
		`\r`: '\r',
		`\t`: '\t',
		`\v`: '\v',
		`\'`: '\'',
		`\\`: '\\',
		`\]`: ']',
		`\-`: '-',
		`\^`: '^',
	}

	if len(m) >= 5 && m[:3] == "\\x{" {
		raw := ""
		for i, n := range m {
			if i >= 3 && i < len(m)-1 && rune(n) != '_' {
				raw = raw + string(n)
			}
		}
		if cp, err := strconv.ParseInt(raw, 16, 64); err == nil {
			return u8.Codepoint(cp), true
		} else {
			return u8.Codepoint(0), false
		}

	} else if len(m) == 4 && m[:2] == "\\x" ||
		len(m) == 6 && m[:2] == "\\u" ||
		len(m) == 10 && m[:2] == "\\U" {
		if cp, err := strconv.ParseInt(m[2:], 16, 64); err == nil {
			return u8.Codepoint(cp), true
		} else {
			return u8.Codepoint(0), false
		}

	} else if len(m) == 4 && m[0] == '\\' && m[1] >= '0' && m[1] <= '7' {
		cp, err := strconv.ParseInt(m[1:], 8, 64)
		if err == nil {
			return u8.Codepoint(cp), true
		} else {
			return u8.Codepoint(0), false
		}

	} else if len(m) >= 1 && m[0] == '\\' {
		return u8.Codepoint(controlCharLookup[m]), true

	} else {
		return u8.Codepoint(m[0]), true
	}
}

func convertForCC(m string) (cc string, ok bool) {
	if len(m) >= 4 && m[:3] == "\\x{" {
		raw := ""
		for i, n := range m {
			if i >= 3 && i < len(m)-1 && rune(n) != '_' {
				raw = raw + string(n)
			}
		}
		if cp, err := strconv.ParseInt(raw, 16, 64); err != nil {
			return "", false
		} else {
			return fmt.Sprintf("\\x{%x}", cp), true
		}
	} else if len(m) == 2 && m == "\\b" {
		return "\\x{8}", true
	} else if len(m) == 2 && m == "\\e" {
		return "\\x{1b}", true
	} else {
		return m, true
	}
}

func convertForCC2(m, n string) (cc, rg string, ok bool) {
	allOk := true
	cp1, ok := convertToCp(m)
	allOk = allOk && ok
	cp2, ok := convertToCp(n)
	allOk = allOk && ok
	if !allOk {
		return "", "", false

	} else if cp1 > u8.Codepoint(0xFFFF) || cp2 > u8.Codepoint(0xFFFF) {
		rp := []rune(u8.Sur(cp1))
		rq := []rune(u8.Sur(cp2))
		if rp[0] == rq[0] {
			return "", fmt.Sprintf("|(?:\\x{%x}[\\x{%x}-\\x{%x}])", rp[0], rp[1], rq[1]), true
		} else if rp[0]+1 == rq[0] {
			return "", fmt.Sprintf("|(?:\\x{%x}[\\x{%x}-\\x{10ffff}])"+
				"|(?:\\x{%x}[\\x{100000}-\\x{%x}])", rp[0], rp[1], rq[0], rq[1]), true
		} else if rp[0]+2 == rq[0] {
			return "", fmt.Sprintf("|(?:\\x{%x}[\\x{%x}-\\x{10ffff}])"+
				"|(?:\\x{%x}[\\x{100000}-\\x{10ffff}])"+
				"|(?:\\x{%x}[\\x{100000}-\\x{%x}])", rp[0], rp[1], rp[0]+1, rq[0], rq[1]), true
		} else if rp[0] < rq[0] {
			return "", fmt.Sprintf("|(?:\\x{%x}[\\x{%x}-\\x{10ffff}])"+
				"|(?:[\\x{%x}-\\x{%x}][\\x{100000}-\\x{10ffff}])"+
				"|(?:\\x{%x}[\\x{100000}-\\x{%x}])", rp[0], rp[1], rp[0]+1, rq[0]-1, rq[0], rq[1]), true
		} else {
			return "", "", false
		}
	} else {
		allOk := true
		cq1, ok := convertForCC(m)
		allOk = allOk && ok
		cq2, ok := convertForCC(n)
		allOk = allOk && ok
		return cq1 + "-" + cq2, "", allOk
	}
}

//==============================================================================

/*
RegexRepeat returns a Regex matching Codepoint, Text, CharClass, or Regex
specified by parameter a based on values in parameter b.
If b is an Int, regex will specify {b} times.
If b is a PosIntRange, regex suffix is according to this table:
  n.From  n.To    regex suffix
  ------  ------  ------------
  0       Inf     *
  1       Inf     +
  0       1       ?
  x       Inf     {x,}
  x       y       {x,y}

If reluctant is specified with PosIntRange, the reluctant match is used, otherwise greedy.
*/
func RegexRepeat(a Any, n Any, reluctant bool) Any {
	switch n.(type) {
	case Int:
		return times(a, n.(Int), reluctant) //TODO: distinguish between Codepoint/Text and CharClass/Regex
	case Pair:
		from := n.(Pair).From()
		inf := n.(Pair).IsToInf()
		to := n.(Pair).To()
		if from == 0 && inf {
			return many(a, reluctant)
		} else if from == to {
			return times(a, from, reluctant)
		} else if from == 1 && inf {
			return many1(a, reluctant)
		} else if from == 0 && to == 1 {
			return optional(a, reluctant)
		} else if inf {
			return atLeast(a, from, reluctant)
		} else {
			return atLeastButNoMoreThan(a, from, to, reluctant)
		}
	}
	panic(fmt.Sprintf("RegexRepeat: invalid numeric/range type %T", n))
}

func many(a Any, reluctant bool) Any {
	suffix := Regex("")
	if reluctant {
		suffix = "?"
	}
	switch a.(type) {
	case u8.Codepoint, u8.Text, CharClass, Regex:
		return Regex(ToRegex(a) + "*" + suffix)
	}
	panic(fmt.Sprintf("many: invalid type %T", a))
}

func many1(a Any, reluctant bool) Any {
	suffix := Regex("")
	if reluctant {
		suffix = "?"
	}
	switch a.(type) {
	case u8.Codepoint, u8.Text, CharClass, Regex:
		return Regex(ToRegex(a) + "+" + suffix)
	}
	panic(fmt.Sprintf("many1: invalid type %T", a))
}

func optional(a Any, reluctant bool) Any {
	suffix := Regex("")
	if reluctant {
		suffix = "?"
	}
	switch a.(type) {
	case u8.Codepoint, u8.Text, CharClass, Regex:
		return Regex(ToRegex(a) + "?" + suffix)
	}
	panic(fmt.Sprintf("optional: invalid type %T", a))
}

func times(a Any, n Int, reluctant bool) Any {
	suffix := ""
	if reluctant {
		suffix = "?"
	}
	switch a.(type) {
	case u8.Codepoint, u8.Text, CharClass, Regex:
		return Regex(string(ToRegex(a)) + fmt.Sprintf("{%d}"+suffix, n))
	}
	panic(fmt.Sprintf("times: invalid type %T", a))
}

func atLeast(a Any, n Int, reluctant bool) Any {
	suffix := ""
	if reluctant {
		suffix = "?"
	}
	switch a.(type) {
	case u8.Codepoint, u8.Text, CharClass, Regex:
		return Regex(string(ToRegex(a)) + fmt.Sprintf("{%d,}"+suffix, n))
	}
	panic(fmt.Sprintf("atLeast: invalid type %T", a))
}

func atLeastButNoMoreThan(a Any, n, m Int, reluctant bool) Any {
	suffix := ""
	if reluctant {
		suffix = "?"
	}
	switch a.(type) {
	case u8.Codepoint, u8.Text, CharClass, Regex:
		return Regex(string(ToRegex(a)) + fmt.Sprintf("{%d,%d}"+suffix, n, m))
	}
	panic(fmt.Sprintf("atLeastButNoMoreThan: invalid type %T", a))
}

//==============================================================================

/*
Group returns a Regex embedded in a capturing group, with name if supplied.
*/
func Group(x Any, name string) Any {
	var xx string
	if name != "" {
		xx = "?P<" + name + ">"
	}
	switch x.(type) {
	case u8.Codepoint:
		return Regex("(" + xx + string(x.(u8.Codepoint)) + ")")
	case u8.Text:
		return Regex("(" + xx + `\Q` + u8.Sur(x.(u8.Text)...) + `\E)`)
	case CharClass:
		return Regex("(" + xx + string(x.(CharClass)) + ")")
	case Regex:
		return Regex("(" + xx + string(x.(Regex)) + ")")
	}
	panic(fmt.Sprintf("Group: invalid type %T", x))
}

/*
Parenthesize puts regex parentheses around a CharClass or Regex,
and returns everything else unchanged.

//TODO: check this logic
*/
func Parenthesize(x Any) Any {
	switch x.(type) {
	case CharClass:
		return Regex("(?:" + string(x.(CharClass)) + ")")
	case Regex:
		return Regex("(?:" + string(x.(Regex)) + ")")
	default:
		return x
	}
}

//==============================================================================

/*
FindFirst returns a slice of RegexGroupMatch structs.
The 0-index in the slice will contain the full match details.
The subsequent indexes in the slice will contain the matches for each capturing group.

TODO: Write logic; test it; write example function.
*/
func FindFirst(x Any, s string) []RegexMatch {
	panic(fmt.Sprintf("FindFirst: invalid type %T", x))
}

//==============================================================================

//NewMap creates a new Map.
func NewMap() *Map {
	return &Map{lkp: map[Any]Any{}, seq: []Any{}}
}

func (m *Map) Add(key, val Any) *Map {
	if _, ok := m.lkp[key]; ok {
		m.Delete(key)
	}
	m.lkp[key] = val
	m.seq = append(m.seq, key)
	return m
}

func (m *Map) Delete(k Any) *Map {
	if _, ok := m.lkp[k]; ok {
		delete(m.lkp, k)
		idx := 0
		for ; idx < len(m.seq); idx++ {
			if m.seq[idx] == k {
				break
			}
		}
		m.seq = append(m.seq[:idx], m.seq[idx+1:len(m.seq)]...)
	}
	return m
}

func (m *Map) Len() int {
	return len(m.lkp)
}

func (m *Map) Merge(m2 *Map) *Map {
	for _, k := range m2.seq {
		m.Add(k, m2.lkp[k])
	}
	return m
}

func (m *Map) Val(key Any) Any {
	return m.lkp[key]
}

func (m *Map) SetVal(key, val Any) {
	m.Add(key, val)
}

func (m *Map) IsEqual(m2 *Map) bool {
	return reflect.DeepEqual(m.lkp, m2.lkp) //don't care about sequence
}

func (m *Map) Copy() *Map {
	mo := &Map{lkp: map[Any]Any{}, seq: []Any{}}
	for _, k := range m.seq {
		v := m.lkp[k]
		mo.lkp[k] = v
		mo.seq = append(mo.seq, k)
	}
	return mo
}

//==============================================================================

/*
Len returns the length of the supplied Slice or Map.
*/
func Len(s Any) int {
	switch s.(type) {
	case Slice:
		return len([]Any(s.(Slice)))

	case *Map:
		return s.(*Map).Len()
	}

	panic(fmt.Sprintf("Len: invalid type %T", s))
}

/*
Copy copies the values of a Slice or the entries of a Map
to a new one.
*/
func Copy(s Any) Any {
	s = widen(s)
	switch s.(type) {
	case Slice:
		so := make([]Any, len([]Any(s.(Slice))))
		copy(so, s.(Slice))
		return Slice(so)

	case *Map:
		u := s.(*Map)
		mo := *u
		return mo.Copy()
	}

	panic(fmt.Sprintf("Copy: invalid type %T", s))
}

//==============================================================================

/*
StructToMap converts a struct to a map of string to data.

TODO: test it
*/
func StructToMap(f Any) map[string]Any {
	fv := reflect.ValueOf(f)
	ft := reflect.TypeOf(f)
	m := map[string]Any{}
	for i := 0; i < fv.NumField(); i++ {
		m[ft.Field(i).Name] = fv.Field(i).Interface()
	}
	return m
}

/*
ArrayToSlice converts an array to a slice.

TODO: test it
*/
func ArrayToSlice(f Any) []Any {
	fv := reflect.ValueOf(f)
	//ft:= reflect.TypeOf(f) //need to check it's really an array
	vec := []Any{}
	for i := 0; i < fv.Len(); i++ {
		vec = append(vec, fv.Index(i).Interface())
	}
	return vec
}

/*
Unwrap unwraps tuply-returned args into a slice.

TODO: test it
*/
func Unwrap(a ...Any) Any {
	if len(a) == 0 { //never happens
		return nil
	} else if len(a) == 1 {
		return a[0]
	} else {
		return a
	}
}

/*
Assert asserts the arg is true.

TODO: test it
*/
func Assert(b Any) {
	if !b.(bool) {
		fmt.Printf("assert failed.\n....found:%v (%[1]T)\n", b)
	}
}

//==============================================================================

/*
Prf printf's an untyped parameter, accepting any number of additional args
as values.
*/
func Prf(fs Any, is ...Any) {
	fmt.Print(sprf(fs, is...))
}

func Sprf(fs Any, is ...Any) u8.Text {
	return u8.Desur(sprf(fs, is...))
}

func sprf(fs Any, is ...Any) string {
	switch fs.(type) {
	case u8.Text:
		return fmt.Sprintf(u8.Surr(fs.(u8.Text)), is...)
	case string:
		return fmt.Sprintf(fs.(string), is...)
	case nil:
		return fmt.Sprint(is...)
	default:
		as := []Any{fs}
		for _, i := range is {
			as = append(as, i)
		}
		return fmt.Sprint(as...)
	}
}

func Pr(is ...Any) {
	fmt.Print(Spr(is...))
}

func Spr(is ...Any) u8.Text {
	return u8.Desur(fmt.Sprint(is...))
}

//==============================================================================
