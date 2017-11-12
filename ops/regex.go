// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package ops

import (
	"fmt"
	"math/big"
	//"reflect"
	"regexp"
	"strconv"

	kn "github.com/grolang/gro/parsec"
	u8 "github.com/grolang/gro/utf88"
)

type (
	CharClass string
	Regex     string
	//Slice     []interface{}
	//Map       map[interface{}]interface{}
	//MapEntry  struct{ Key, Val interface{} }
)

func (x CharClass) unwrap() (string, bool) {
	mid := string(x)[1 : len(x)-1]
	neg := false
	if mid[0] == '^' {
		neg = true
		mid = mid[1:]
	}
	return mid, neg
}

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
func Runex(x interface{}) interface{} {
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
		panic("Runex: invalid input type")
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

/*
RegexGroupMatch is a struct returned from a regex match using capturing groups
to indicate the results of one of the capturing groups.
Name is the name of the group in the regex.
Start and End are the starting and ending positions in the regex.
Match is the result of the match.
*/
type RegexGroupMatch struct {
	Name       string
	Start, End int
	Match      string
}

/*
FindAll returns a slice of a slice of RegexGroupMatch structs.
The outer slice contains an inner slice for each match.
The 0-index in each inner slice will contain the full match details.
The subsequent indexes in each inner slice will contain the matches for each capturing group.

TODO: Write example function.
*/
func FindAll(x interface{}, s string) [][]RegexGroupMatch {
	var p string
	switch x.(type) {
	case u8.Codepoint:
		p = string(x.(u8.Codepoint))
	case u8.Text:
		p = `\Q` + u8.Sur(x.(u8.Text)...) + `\E`
	case CharClass:
		p = string(x.(CharClass))
	case Regex:
		p = string(x.(Regex))
	default:
		panic("FindAll: invalid input type, must be codepoint, string, char class, or regex")
	}

	re := regexp.MustCompile(p)
	arr := re.FindAllStringSubmatchIndex(s, -1)
	names := re.SubexpNames()
	numGrTot := re.NumSubexp() + 1
	rgm := make([][]RegexGroupMatch, len(arr))
	for n, _ := range arr {
		rgm[n] = make([]RegexGroupMatch, numGrTot)
		for m := 0; m < numGrTot; m++ {
			x, y := arr[n][2*m], arr[n][2*m+1]
			var ss string
			if x != -1 && y != -1 {
				ss = s[x:y]
			}
			rgm[n][m] = RegexGroupMatch{names[m], x, y, ss}
		}
	}
	return rgm
}

/*
FindFirst returns a slice of RegexGroupMatch structs.
The 0-index in the slice will contain the full match details.
The subsequent indexes in the slice will contain the matches for each capturing group.

TODO: Write logic; test it; write example function.
*/
func FindFirst(x interface{}, s string) []RegexGroupMatch {
	panic("FindFirst: invalid first arg")
}

/*
RegexRepeat returns a Regex matching Codepoint, Text, CharClass, or Regex specified by parameter a
based on values in parameter b. If b is an Int, regex will specify {b} times.
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
func RegexRepeat(a interface{}, n interface{}, reluctant bool) interface{} {
	switch n.(type) {
	case Int:
		return times(a, n.(Int), reluctant) //TODO: distinguish between Codepoint/Text and CharClass/Regex
	case PosIntRange:
		from := n.(PosIntRange).FromVal()
		inf := n.(PosIntRange).IsToInf()
		to := n.(PosIntRange).ToVal()
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
	default:
		panic("RegexRepeat: invalid input type")
	}
}

func many(a interface{}, reluctant bool) interface{} {
	suffix := Regex("")
	if reluctant {
		suffix = "?"
	}
	switch a.(type) {
	case u8.Codepoint, u8.Text, CharClass, Regex:
		return Regex(ToRegex(a) + "*" + suffix)
	default:
		panic("Many: invalid input type")
	}
}

func many1(a interface{}, reluctant bool) interface{} {
	suffix := Regex("")
	if reluctant {
		suffix = "?"
	}
	switch a.(type) {
	case u8.Codepoint, u8.Text, CharClass, Regex:
		return Regex(ToRegex(a) + "+" + suffix)
	default:
		panic("Many1: invalid input type")
	}
}

func optional(a interface{}, reluctant bool) interface{} {
	suffix := Regex("")
	if reluctant {
		suffix = "?"
	}
	switch a.(type) {
	case u8.Codepoint, u8.Text, CharClass, Regex:
		return Regex(ToRegex(a) + "?" + suffix)
	default:
		panic("Optional: invalid input type")
	}
}

func times(a interface{}, n Int, reluctant bool) interface{} {
	suffix := ""
	if reluctant {
		suffix = "?"
	}
	switch a.(type) {
	case u8.Codepoint, u8.Text, CharClass, Regex:
		return Regex(string(ToRegex(a)) + fmt.Sprintf("{%d}"+suffix, n))
	default:
		panic("Times: invalid input type")
	}
}

func atLeast(a interface{}, n Int, reluctant bool) interface{} {
	suffix := ""
	if reluctant {
		suffix = "?"
	}
	switch a.(type) {
	case u8.Codepoint, u8.Text, CharClass, Regex:
		return Regex(string(ToRegex(a)) + fmt.Sprintf("{%d,}"+suffix, n))
	default:
		panic("AtLeast: invalid input type")
	}
}

func atLeastButNoMoreThan(a interface{}, n, m Int, reluctant bool) interface{} {
	suffix := ""
	if reluctant {
		suffix = "?"
	}
	switch a.(type) {
	case u8.Codepoint, u8.Text, CharClass, Regex:
		return Regex(string(ToRegex(a)) + fmt.Sprintf("{%d,%d}"+suffix, n, m))
	default:
		panic("AtLeastButNoMoreThan: invalid input type")
	}
}

/*
Group returns a Regex embedded in a capturing group, with name if supplied.
*/
func Group(x interface{}, name string) interface{} {
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
	default:
		panic("Group: invalid input type")
	}
}

/*
ToRegex converts a Codepoint, Text, or CharClass into a Regex.

//TODO: TEST IT
*/
func ToRegex(x interface{}) Regex {
	switch x.(type) {
	case u8.Codepoint:
		return Regex(u8.Sur(x.(u8.Codepoint)))
	case u8.Text:
		return Regex(`(?:\Q` + u8.Sur(x.(u8.Text)...) + `\E)`)
	case CharClass:
		return Regex(string(x.(CharClass)))
	case Regex:
		return Regex("(?:" + string(x.(Regex)) + ")")
	default:
		panic("ToRegex: invalid input type")
	}
}

/*
ToParser converts a Codepoint, Text, CharClass, or Regex into a Parser.

//TODO: TEST IT
*/
func ToParser(x interface{}) kn.Parser {
	switch x.(type) {
	case rune:
		return codepointParser(string(x.(rune)))
	case string:
		return kn.Token(u8.Desur(x.(string)))

	case u8.Codepoint:
		return codepointParser(u8.Sur(x.(u8.Codepoint)))
	case u8.Text:
		return kn.Token(x.(u8.Text))
	case CharClass:
		return runexParser(x.(CharClass))
	case Regex:
		return kn.Regexp(string(x.(Regex)))
	case kn.Parser:
		return x.(kn.Parser)
	default:
		panic(fmt.Sprintf("ToParser: invalid input type: %T", x))
	}
}

func codepointParser(r string) kn.Parser {
	f := func(c interface{}) interface{} {
		return []u8.Codepoint(c.(u8.Text))[0]
	}
	return kn.Apply(f, kn.Token(u8.Desur(r)))
}

func runexParser(r CharClass) kn.Parser {
	f := func(c interface{}) interface{} {
		return []u8.Codepoint(c.(u8.Text))[0]
	}
	return kn.Apply(f, kn.Regexp(string(r)))
}

/*
ToBool converts nil, false, zeroes, the empty string or utf88.Text to false,
and all other values to true.

//TODO: TEST IT
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

/*
Parenthesize puts regex parentheses around a CharClass or Regex,
and returns everything else unchanged.

//TODO: CHECK AND FINISH TESTING
*/
func Parenthesize(x interface{}) interface{} {
	switch x.(type) {
	case CharClass:
		return Regex("(?:" + string(x.(CharClass)) + ")")
	case Regex:
		return Regex("(?:" + string(x.(Regex)) + ")")
	default:
		return x
	}
}

//TODO: get rid of this when no longer required by testing
/*func UnaryConvert (x interface{}) interface{}{ //converts Text into a Regex
  switch x.(type){
  case u8.Text:
    return Regex(u8.Surr(x.(u8.Text)))
  default:
    return x
  }
}*/

/*
Not returns the boolean-not of the parameter converted to a boolean.
For Codepoint or CharClass, it returns the CharClass that matches the complement.

//TODO: FINISH TESTING
*/
func Not(x interface{}) interface{} {
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
		return !ToBool(x)
	}
}

/*
And returns the boolean-and of parameter a and the value returned from calling parameter b
as a function. Parameter b is only called if a is true.
*/
func And(x interface{}, y func() interface{}) bool {
	return ToBool(x) && ToBool(y())
}

/*
Or returns the boolean-or of parameter a and the value returned from calling parameter b
as a function. Parameter b is only called if a is false.
*/
func Or(x interface{}, y func() interface{}) bool {
	return ToBool(x) || ToBool(y())
}

/*
Alt accepts a Codepoint, Text, CharClass, Regex, or Parser for each parameter,
and returns a CharClass, Regex, or Parser using ordered alternation on the inputs.

The args convert such that:
  2 Codepoints become a CharClass
  a CharClass or Codepoint with a CharClass may remain a CharClass, but might become a Regex
  anything with a Parser becomes a Parser
  anything else with Text or a Regex becomes a Regex

//TODO: FINISH TESTING
*/
func Alt(x, y interface{}) interface{} {
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
		case kn.Parser:
			return kn.Alt(ToParser(x), y.(kn.Parser))
		default:
			panic("Alt: invalid input type")
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
		case kn.Parser:
			return kn.Alt(ToParser(x), y.(kn.Parser))
		default:
			panic("Alt: invalid input type")
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
		case kn.Parser:
			return kn.Alt(ToParser(x), y.(kn.Parser))
		default:
			panic("Alt: invalid input type")
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
		case kn.Parser:
			return kn.Alt(ToParser(x), y.(kn.Parser))
		default:
			panic("Alt: invalid input type")
		}

	case kn.Parser:
		switch y.(type) {
		case u8.Codepoint, u8.Text, CharClass, Regex:
			return kn.Alt(x.(kn.Parser), ToParser(y))
		case kn.Parser:
			return kn.Alt(x.(kn.Parser), y.(kn.Parser))
		default:
			panic("Alt: invalid input type")
		}

	default:
		panic("Alt: invalid arg/s")
	}
}

/*
Seq accepts a Codepoint, Text, CharClass, or Regex for each parameter,
and returns Text concatenating the inputs, or a Regex matching the inputs in sequence.

The args convert such that:
  2 Codepoints become a Text, as does a Codepoint combined with a Text
  2 Texts remain a Text
  2 CharClasses become a Regex, as does a Codepoint or Text combined with a CharClass
  2 Regexes remain a Regex, as does a Codepoint, Text, or CharClass combined with a Regex

//TODO: FINISH TESTING
*/
func Seq(x, y interface{}) interface{} {
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
		default:
			panic("Seq: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
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
		default:
			panic("Seq: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
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
		default:
			panic("Seq: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
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
		default:
			panic("Seq: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
		}

	default:
		panic("Seq: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
	}
}

/*
LeftShift calls kern.LeftShift with the 2 parsers supplied as args.

//TODO: TEST IT
*/
func LeftShift(x, y interface{}) interface{} {
	switch x.(type) {
	case kn.Parser:
		switch y.(type) {
		case kn.Parser:
			return kn.SeqLeft(x.(kn.Parser), y.(kn.Parser))
		default:
			panic("LeftShift: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
		}
	default:
		panic("LeftShift: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
	}
}

/*
RightShift calls kern.RightShift with the 2 parsers supplied as args.

//TODO: TEST IT
*/
func RightShift(x, y interface{}) interface{} {
	switch x.(type) {
	case kn.Parser:
		switch y.(type) {
		case kn.Parser:
			return kn.SeqRight(x.(kn.Parser), y.(kn.Parser))
		default:
			panic("RightShift: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
		}
	default:
		panic("RightShift: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
	}
}

/*
Matcher takes a Parser, Regex, CharClass, Text, or Codepoint as first arg,
and calls kern.ParseText with it on the Text supplied as the second arg.
In Gro, operator ==~ will be used.

//TODO: TEST IT
*/
func Matcher(x, y interface{}) interface{} {
	y = widen(y)
	switch x.(type) {
	case u8.Codepoint, u8.Text, CharClass, Regex, kn.Parser:
		switch y.(type) {
		case u8.Text:
			r, _ := kn.ParseText(ToParser(x), y.(u8.Text))
			return r
		default:
			panic("Matcher: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
		}
	default:
		panic("Matcher: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
	}
}

/*
Matcher takes a Regex, CharClass, Text, or Codepoint as first arg,
and calls FindAll with it on the Text supplied as the second arg.
In Gro, operator =~ will be used.

//TODO: TEST IT
*/
func Finder(x, y interface{}) interface{} {
	y = widen(y)
	switch x.(type) {
	case u8.Codepoint, u8.Text, CharClass, Regex: //MAYBE TODO: also use Parser, calling kern.Search ???
		switch y.(type) {
		case u8.Text:
			return FindAll(x, u8.Surr(y.(u8.Text)))
		default:
			panic("Finder: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
		}

	default:
		panic("Finder: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
	}
}

/*
Mod takes two Ints and calls Go's % operator on them.

//TODO: TEST IT
*/
func Mod(x, y interface{}) interface{} {
	x, y = widen(x), widen(y)
	switch x.(type) {
	case Int:
		switch y.(type) {
		case Int:
			return Int(int64(x.(Int)) % int64(y.(Int)))
		default:
			panic("Mod: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
		}
	default:
		panic("Mod: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
	}
}

/*
Xor takes two Ints and calls Go's ^ operator on them.

//TODO: TEST IT
*/
func Xor(x, y interface{}) interface{} {
	x, y = widen(x), widen(y)
	switch x.(type) {
	case Int:
		switch y.(type) {
		case Int:
			return Int(int64(x.(Int)) ^ int64(y.(Int)))
		default:
			panic("Xor: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
		}
	default:
		panic("Xor: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
	}
}

/*
SeqXor takes two Ints and calls Go's &^ operator on them.

//TODO: TEST IT
*/
func SeqXor(x, y interface{}) interface{} {
	x, y = widen(x), widen(y)
	switch x.(type) {
	case Int:
		switch y.(type) {
		case Int:
			return Int(int64(x.(Int)) &^ int64(y.(Int)))
		default:
			panic("SeqXor: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
		}
	default:
		panic("SeqXor: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
	}
}

/*
Assign makes the first arg point to the second arg.
*/
func Assign(pp *interface{}, k interface{}) *interface{} {
	switch k.(type) {
	case nil, bool, Int, BigInt, BigRat, Float, Complex,
		string:
		k = widen(k)
	default:
		k = widen(k)
	}
	*pp = k
	return pp
}

//MAYBE TODO: MultipleAssign ???

/*
PlusAssign adds the value/s to a vector or the key-value pair/s to a map,
otherwise calls Plus and Assign on the args.
*/
func PlusAssign(pp *interface{}, it ...interface{}) *interface{} {
	//TODO: widen pp
	switch (*pp).(type) {
	/*case Slice:
		v := (*pp).(Slice)
		v = append(v, it...)
		*pp = v
		return pp

	case Map:
		m := (*pp).(Map)
		for _, u := range it {
			m[u.(MapEntry).Key] = u.(MapEntry).Val
		}
		return pp
	*/

	default:
		for _, n := range it {
			nn := widen(n)

			switch nn.(type) {
			case nil, bool, Int, BigInt, BigRat, Float, Complex,
				string:
				*pp = Plus(*pp, nn)

			default:
				//TODO: any logic for arbitrary object?
				panic(fmt.Sprintf("PlusAssign: invalid parameter/s: %T", nn))
			}
		}
		return pp
	}
}

/*
MinusAssign deletes an entry from a Map, otherwise calls Minus and Assign on the args.
*/
func MinusAssign(pp *interface{}, k interface{}) *interface{} { // TODO: add Slice functionality
	//TODO: widen pp
	/*if m, ok := (*pp).(Map); ok {
		delete(m, k)
		return pp
	} else {*/

	nn := widen(k)

	switch nn.(type) {
	case nil, bool, Int, BigInt, BigRat, Float, Complex,
		string:
		*pp = Minus(*pp, nn)
		return pp

	default:
		//TODO: any logic for arbitrary object?
		panic(fmt.Sprintf("MinusAssign: invalid parameter/s: %T", nn))
	}

	//}
}

/*
MultAssign calls Mult and Assign on the args.

//TODO: TEST IT
*/
func MultAssign(pp *interface{}, it ...interface{}) *interface{} {
	//TODO: widen pp
	for _, n := range it {
		nn := widen(n)
		switch nn.(type) {
		case nil, bool, Int, BigInt, BigRat, Float, Complex,
			string:
			*pp = Mult(*pp, nn)

		default:
			//TODO: any logic for arbitrary object?
			panic(fmt.Sprintf("MultAssign: invalid parameter/s: %T", nn))
		}
	}
	return pp
}

/*
DivideAssign calls Divide and Assign on the args.

//TODO: TEST IT
*/
func DivideAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	nn := widen(k)
	switch nn.(type) {
	case nil, bool, Int, BigInt, BigRat, Float, Complex,
		string:
		*pp = Divide(*pp, nn)

	default:
		//TODO: any logic for arbitrary object?
		panic(fmt.Sprintf("DivideAssign: invalid parameter/s: %T", nn))
	}
	return pp
}

/*
ModAssign calls Mod and Assign on the args.

//TODO: TEST IT
*/
func ModAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	nn := widen(k)
	switch nn.(type) {
	case nil, bool, Int, BigInt, BigRat, Float, Complex,
		string:
		*pp = Mod(*pp, nn)

	default:
		//TODO: any logic for arbitrary object?
		panic(fmt.Sprintf("ModAssign: invalid parameter/s: %T", nn))
	}
	return pp
}

/*
LeftShiftAssign calls LeftShift and Assign on the args.

//TODO: TEST IT
*/
func LeftShiftAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	nn := widen(k)
	switch nn.(type) {
	case nil, bool, Int, BigInt, BigRat, Float, Complex,
		string:
		*pp = LeftShift(*pp, nn)

	default:
		//TODO: any logic for arbitrary object?
		panic(fmt.Sprintf("LeftShiftAssign: invalid parameter/s: %T", nn))
	}
	return pp
}

/*
RightShiftAssign calls RightShift and Assign on the args.

//TODO: TEST IT
*/
func RightShiftAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	nn := widen(k)
	switch nn.(type) {
	case nil, bool, Int, BigInt, BigRat, Float, Complex,
		string:
		*pp = RightShift(*pp, nn)

	default:
		//TODO: any logic for arbitrary object?
		panic(fmt.Sprintf("RightShiftAssign: invalid parameter/s: %T", nn))
	}
	return pp
}

/*
SeqAssign calls Seq and Assign on the args.

//TODO: TEST IT
*/
func SeqAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	nn := widen(k)
	switch nn.(type) {
	case nil, bool, Int, BigInt, BigRat, Float, Complex,
		string:
		*pp = Seq(*pp, nn)

	default:
		//TODO: any logic for arbitrary object?
		panic(fmt.Sprintf("SeqAssign: invalid parameter/s: %T", nn))
	}
	return pp
}

/*
SeqXorAssign calls Xor and Assign on the args.

//TODO: TEST IT
*/
func SeqXorAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	nn := widen(k)
	switch nn.(type) {
	case nil, bool, Int, BigInt, BigRat, Float, Complex,
		string:
		*pp = SeqXor(*pp, nn)

	default:
		//TODO: any logic for arbitrary object?
		panic(fmt.Sprintf("SeqXorAssign: invalid parameter/s: %T", nn))
	}
	return pp
}

/*
AltAssign calls Alt and Assign on the args.

//TODO: TEST IT
*/
func AltAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	nn := widen(k)
	switch nn.(type) {
	case nil, bool, Int, BigInt, BigRat, Float, Complex,
		string:
		*pp = Alt(*pp, nn)

	default:
		//TODO: any logic for arbitrary object?
		panic(fmt.Sprintf("AltAssign: invalid parameter/s: %T", nn))
	}
	return pp
}

/*
XorAssign calls Xor and Assign on the args.

//TODO: TEST IT
*/
func XorAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	nn := widen(k)
	switch nn.(type) {
	case nil, bool, Int, BigInt, BigRat, Float, Complex,
		string:
		*pp = Xor(*pp, nn)

	default:
		//TODO: any logic for arbitrary object?
		panic(fmt.Sprintf("XorAssign: invalid parameter/s: %T", nn))
	}
	return pp
}

/*
Incr increments the arg in-place.

//TODO: FINISH TESTING
*/
func Incr(pn *interface{}) interface{} {
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

//TODO: FINISH TESTING
*/
func Decr(pn *interface{}) interface{} {
	temp := *pn
	switch (*pn).(type) {
	case Int, BigInt:
		*pn = Minus(*pn, 1)
	default:
		panic("Decr: non-integral arg")
	}
	return temp
}

/*
GetIndex gets the index specified by the second arg of the first arg.

//TODO: FINISH TESTING
*/
/*func GetIndex(pa *interface{}, n interface{}) *interface{} {
	switch (*pa).(type) {
	case []interface{}:
		return &(*pa).([]interface{})[n.(int)]
	case map[interface{}]interface{}:
		an := (*pa).(map[interface{}]interface{})[n]
		return &an
	case Slice:
		return &[]interface{}((*pa).(Slice))[n.(int)]
	case Map:
		an := (*pa).(Map)[n]
		return &an
	default:
		panic(fmt.Sprintf("GetIndex: invalid aggregate type: %T", *pa))
	}
}

//TODO: MERGE INTO GetIndex
func substring(s string, b, e int) string {
	if b > 0 {
		if e > 0 {
			return s[b:e]
		} else {
			return s[b : len(s)+e]
		}
	} else {
		if e > 0 {
			return s[len(s)+b : e]
		} else {
			return s[len(s)+b : len(s)+e]
		}
	}
}*/

/*
SetIndex sets the index specified by the second arg of the first arg to the third arg.

//TODO: FINISH TESTING
*/
/*func SetIndex(pa *interface{}, n, u interface{}) *interface{} {
	v := widen(u)
	switch (*pa).(type) {
	case []interface{}:
		(*pa).([]interface{})[n.(int)] = v
		an := (*pa).([]interface{})[n.(int)]
		return &an
	case map[interface{}]interface{}:
		(*pa).(map[interface{}]interface{})[n] = v
		an := (*pa).(map[interface{}]interface{})[n]
		return &an
	case Slice:
		[]interface{}((*pa).(Slice))[n.(int)] = v
		an := []interface{}((*pa).(Slice))[n.(int)]
		return &an
	case Map:
		map[interface{}]interface{}((*pa).(Map))[n] = v
		an := map[interface{}]interface{}((*pa).(Map))[n]
		return &an
	default:
		panic("SetIndex: invalid type")
	}
}*/

/*
NewSlice creates a new Slice.

//TODO: FINISH TESTING
*/
/*func NewSlice(n interface{}) interface{} {
	return Slice(make([]interface{}, n.(int)))
}*/

/*
NewMap creates a new Map.

//TODO: FINISH TESTING
*/
/*func NewMap() interface{} {
	return Map(make(map[interface{}]interface{}))
}*/

/*
InitSlice creates a new Slice and initializes it to the supplied values.

//TODO: FINISH TESTING
*/
/*func InitSlice(vs ...interface{}) interface{} {
	s := make([]interface{}, 0)
	s = append(s, vs...)
	return Slice(s)
}*/

//TODO: write method String for Slice

/*
InitMap creates a new Map and initializes it to the supplied pair-values.

//TODO: FINISH TESTING
*/
/*func InitMap(es ...interface{}) interface{} {
	m := make(map[interface{}]interface{})
	for _, e := range es {
		ee := []interface{}(e.(Slice))
		if len(ee) == 2 {
			m[ee[0]] = ee[1]
		} else {
			panic("InitMap: each arg must be a Slice of size 2.")
		}
	}
	return Map(m)
}*/

//TODO: write method String for Map

/*
Len returns the length of the supplied Slice or Map.

//TODO: FINISH TESTING
*/
/*func Len(s interface{}) int {
	switch s.(type) {
	case Slice:
		return len([]interface{}(s.(Slice)))
	case Map:
		return len(s.(Map))
	default:
		panic("Len: unknown type")
	}
}*/

/*
Copy copies the values of a Slice or the entries of a Map
to a new one.

//TODO: FINISH TESTING
*/
/*func Copy(s interface{}) interface{} {
	switch s.(type) {
	case Slice:
		so := make([]interface{}, len([]interface{}(s.(Slice))))
		copy(so, s.(Slice))
		return Slice(so)
	case Map:
		mo := make(map[interface{}]interface{})
		for k, v := range s.(Map) {
			mo[k] = v
		}
		return Map(mo)
	default:
		panic("Copy: unknown type")
	}
}*/

/*
Unwrap unwraps tuply-returned args into a slice.

//TODO: TEST IT
*/
/*func Unwrap(a ...interface{}) interface{} {
	if len(a) == 0 { //never happens
		return nil
	} else if len(a) == 1 {
		return a[0]
	} else {
		return a
	}
}*/

/*
Assert asserts the arg is true.

//TODO: TEST IT
*/
/*func Assert(b interface{}) {
	if !b.(bool) {
		fmt.Printf("assert failed.\n....found:%v (%[1]T)\n", b)
	}
}*/

/*
StructToMap converts a struct to a map of string to data.

//TODO: TEST IT
*/
/*func StructToMap(f interface{}) map[string]interface{} {
	fv := reflect.ValueOf(f)
	ft := reflect.TypeOf(f)
	m := map[string]interface{}{}
	for i := 0; i < fv.NumField(); i++ {
		m[ft.Field(i).Name] = fv.Field(i).Interface()
	}
	return m
}*/

/*
ArrayToSlice converts an array to a slice.

//TODO: TEST IT
*/
/*func ArrayToSlice(f interface{}) []interface{} {
	fv := reflect.ValueOf(f)
	//ft:= reflect.TypeOf(f) //need to check it's really an array
	vec := []interface{}{}
	for i := 0; i < fv.Len(); i++ {
		vec = append(vec, fv.Index(i).Interface())
	}
	return vec
}*/

/*
Prf printf's an untyped parameter, accepting any number of additional args
as values.

//TODO: TEST IT
*/
func Prf(fs interface{}, is ...interface{}) {
	fmt.Print(Sprf(fs, is...))
}

func Sprf(fs interface{}, is ...interface{}) string {
	switch fs.(type) {
	case u8.Text:
		return fmt.Sprintf(u8.Surr(fs.(u8.Text)), is...)
	case string:
		return fmt.Sprintf(fs.(string), is...)
	case nil:
		return fmt.Sprintln(is...)
	default:
		as := []interface{}{fs}
		for _, i := range is {
			as = append(as, i)
		}
		return fmt.Sprintln(as...)
	}
}
