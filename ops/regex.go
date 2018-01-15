// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package ops

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"

	"github.com/grolang/gro/parsec"
	u8 "github.com/grolang/gro/utf88"
)

var UseUtf88 bool

type (
	CharClass string
	Regex     string
)

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
MakeText accepts a string as input, and returns Text.
*/
func MakeText(x interface{}) interface{} {
	return u8.Desur(x.(string))
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

//==============================================================================

/*
ParserRepeat returns a Parser matching Codepoint, Text, CharClass, or Parser
specified by parameter a based on values in parameter n.
If n is an Int, parser.Times(n) is called.
If n is a PosIntRange, parser is run according to this table:
  n.From  n.To    parser func
  ------  ------  -----------
  0       Inf     Many
  1       Inf     Many1
  0       1       Optional
TODO:
  x       Inf     Times(x), then Many
  x       y       Times(x), then no more than y-x
*/
func ParserRepeat(a interface{}, n interface{}) interface{} {
	switch n.(type) {
	case Int:
		return parsec.Times(n.(Int), a.(parsec.Parser))
	case multiusePair:
		from := n.(multiusePair).From()
		inf := n.(multiusePair).IsToInf()
		to := n.(multiusePair).To()
		if from == 0 && inf {
			return parsec.Many(a.(parsec.Parser))
		} else if from == to {
			return parsec.Times(from, a.(parsec.Parser))
		} else if from == 1 && inf {
			return parsec.Many1(a.(parsec.Parser))
		} else if from == 0 && to == 1 {
			return parsec.Optional(a.(parsec.Parser))
		} else if inf {
			panic("ParserRepeat: this combo of nums not yet implemented")
		} else {
			panic("ParserRepeat: this combo of nums not implemented")
		}
	default:
		panic("ParserRepeat: invalid input type")
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
func RegexRepeat(a interface{}, n interface{}, reluctant bool) interface{} {
	switch n.(type) {
	case Int:
		return times(a, n.(Int), reluctant) //TODO: distinguish between Codepoint/Text and CharClass/Regex
	case multiusePair:
		from := n.(multiusePair).From()
		inf := n.(multiusePair).IsToInf()
		to := n.(multiusePair).To()
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

//==============================================================================

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
Parenthesize puts regex parentheses around a CharClass or Regex,
and returns everything else unchanged.

//TODO: check this logic
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

//==============================================================================

/*
ToRegex converts a Codepoint, Text, or CharClass into a Regex.
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
*/
func ToParser(x interface{}) parsec.Parser {
	switch x.(type) {
	case rune:
		return codepointParser(string(x.(rune)))
	case string:
		return parsec.Token(u8.Desur(x.(string)))

	case u8.Codepoint:
		return codepointParser(u8.Sur(x.(u8.Codepoint)))
	case u8.Text:
		return parsec.Token(x.(u8.Text))
	case CharClass:
		return runexParser(x.(CharClass))
	case Regex:
		return parsec.Regexp(string(x.(Regex)))
	case parsec.Parser:
		return x.(parsec.Parser)
	default:
		panic(fmt.Sprintf("ToParser: invalid input type: %T", x))
	}
}

func codepointParser(r string) parsec.Parser {
	f := func(c interface{}) interface{} {
		return []u8.Codepoint(c.(u8.Text))[0]
	}
	return parsec.Apply(f, parsec.Token(u8.Desur(r)))
}

func runexParser(r CharClass) parsec.Parser {
	f := func(c interface{}) interface{} {
		return []u8.Codepoint(c.(u8.Text))[0]
	}
	return parsec.Apply(f, parsec.Regexp(string(r)))
}

//==============================================================================

/*
Reflect returns reflection object of interface value, or vice versa.
*/
func Reflect(x interface{}) interface{} {
	x = widen(x)
	switch x.(type) {
	case reflect.Value:
		return x.(reflect.Value).Interface()
	default:
		return reflect.ValueOf(x)
	}
}

//==============================================================================

/*
Not returns the boolean-not of the parameter converted to a boolean.
For Codepoint or CharClass, it returns the CharClass that matches the complement.
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
		case parsec.Parser:
			return parsec.Alt(ToParser(x), y.(parsec.Parser))
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
		case parsec.Parser:
			return parsec.Alt(ToParser(x), y.(parsec.Parser))
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
		case parsec.Parser:
			return parsec.Alt(ToParser(x), y.(parsec.Parser))
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
		case parsec.Parser:
			return parsec.Alt(ToParser(x), y.(parsec.Parser))
		default:
			panic("Alt: invalid input type")
		}

	case parsec.Parser:
		switch y.(type) {
		case u8.Codepoint, u8.Text, CharClass, Regex:
			return parsec.Alt(x.(parsec.Parser), ToParser(y))
		case parsec.Parser:
			return parsec.Alt(x.(parsec.Parser), y.(parsec.Parser))
		default:
			panic("Alt: invalid input type")
		}

	default:
		panic("Alt: invalid arg/s")
	}
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

//==============================================================================

/*
LeftAnd calls parsec.SeqLeft with the 2 parsers supplied as args.

TODO: Write example function.
*/
func LeftAnd(x, y interface{}) interface{} {
	return directionalAnd(parsec.SeqLeft, "SeqLeft", x, y)
}

/*
RightAnd calls parsec.SeqRight with the 2 parsers supplied as args.

TODO: Write example function.
*/
func RightAnd(x, y interface{}) interface{} {
	return directionalAnd(parsec.SeqRight, "SeqRight", x, y)
}

func directionalAnd(
	function func(...interface{}) interface{},
	text string,
	x, y interface{},
) interface{} {
	x, y = widen(x), widen(y)
	switch x.(type) {

	case u8.Codepoint:
		ax := x.(u8.Codepoint)
		switch y.(type) {
		case u8.Codepoint:
			return function(parsec.Symbol(ax), parsec.Symbol(y.(u8.Codepoint)))
		case u8.Text:
			return function(parsec.Symbol(ax), parsec.Token(y.(u8.Text)))
		case parsec.Parser:
			return function(parsec.Symbol(ax), y.(parsec.Parser))
		case func(...interface{}) interface{}:
			ay := y.(func(...interface{}) interface{})
			//yfwd := parsec.Fwd(ay().(func(...interface{}) interface{})).(parsec.Parser)
			yfwd := unforwardFunc(ay)
			return function(parsec.Symbol(ax), yfwd)
		default:
			panic(text + ": invalid input type" + fmt.Sprintf("; %T; %T", x, y))
		}

	case u8.Text:
		ax := x.(u8.Text)
		switch y.(type) {
		case u8.Codepoint:
			return function(parsec.Token(ax), parsec.Symbol(y.(u8.Codepoint)))
		case u8.Text:
			return function(parsec.Token(ax), parsec.Token(y.(u8.Text)))
		case parsec.Parser:
			return function(parsec.Token(ax), y.(parsec.Parser))
		case func(...interface{}) interface{}:
			ay := y.(func(...interface{}) interface{})
			//yfwd := parsec.Fwd(ay().(func(...interface{}) interface{})).(parsec.Parser)
			yfwd := unforwardFunc(ay)
			return function(parsec.Token(ax), yfwd)
		default:
			panic(text + ": invalid input type" + fmt.Sprintf("; %T; %T", x, y))
		}

	case parsec.Parser:
		ax := x.(parsec.Parser)
		switch y.(type) {
		case u8.Codepoint:
			return function(ax, parsec.Symbol(y.(u8.Codepoint)))
		case u8.Text:
			return function(ax, parsec.Token(y.(u8.Text)))
		case parsec.Parser:
			return function(ax, y.(parsec.Parser))
		case func(...interface{}) interface{}:
			ay := y.(func(...interface{}) interface{})
			//yfwd := parsec.Fwd(ay().(func(...interface{}) interface{})).(parsec.Parser)
			yfwd := unforwardFunc(ay)
			return function(ax, yfwd)
		default:
			panic(text + ": invalid input type" + fmt.Sprintf("; %T; %T", x, y))
		}

	case func(...interface{}) interface{}:
		ax := x.(func(...interface{}) interface{})
		//xfwd := parsec.Fwd(ax().(func(...interface{}) interface{})).(parsec.Parser)
		xfwd := unforwardFunc(ax)
		switch y.(type) {
		case u8.Codepoint:
			return function(xfwd, parsec.Symbol(y.(u8.Codepoint)))
		case u8.Text:
			return function(xfwd, parsec.Token(y.(u8.Text)))
		case parsec.Parser:
			return function(xfwd, y.(parsec.Parser))
		case func(...interface{}) interface{}:
			ay := y.(func(...interface{}) interface{})
			//yfwd := parsec.Fwd(ay().(func(...interface{}) interface{})).(parsec.Parser)
			yfwd := unforwardFunc(ay)
			return function(xfwd, yfwd)
		default:
			panic(text + ": invalid input type" + fmt.Sprintf("; %T; %T", x, y))
		}

	default:
		panic(text + ": invalid input type" + fmt.Sprintf("; %T; %T", x, y))
	}
}

//==============================================================================

/*
Xor takes two Ints and calls Go's ^ operator on them.
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

//==============================================================================

/*func unforwardFunc(ay func(...interface{}) interface{}) interface{} {
	yfwd := parsec.Fwd(ay().(func(...interface{}) interface{})).(parsec.Parser)
	return yfwd
}*/

func unforwardFunc(ay func(...interface{}) interface{}) interface{} {
	aa := ay()
	yfwd := parsec.Fwd(aa.(func(...interface{}) interface{})).(parsec.Parser)
	return yfwd
}

func unwrapFunc(y func(...interface{}) interface{}) interface{} {
	ret := y()
	for {
		if z, ok := ret.(func(...interface{}) interface{}); ok {
			ret = z()
		} else {
			break
		}
	}
	return ret
}

/*
LeftShift calls parsec.SeqLeft with the 2 parsers supplied as args.
*/
func LeftShift(x, y interface{}) interface{} {
	x, y = widen(x), widen(y)
	switch x.(type) {
	case Slice:
		return Slice(append([]interface{}(x.(Slice)), y))

	case Map: //TODO: is this correct ??? Test it!
		ax, ay := x.(Map), y.(MapEntry)
		ax.lkp[ay.Key()] = ay.Val()
		ax.seq = append(ax.seq, ay.Key()) //####
		return ax

	default:
		panic("LeftShift: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
	}
}

/*
The the 1st arg must be some Text.
If the 2nd arg is a Regex, CharClass, Text, or Codepoint,
it is used to match the Text.
It will return a slice of a slice of RegexGroupMatch structs.
The outer slice contains an inner slice for each match.
The 0-index in each inner slice will contain the full match details.
The subsequent indexes in each inner slice will contain the matches for each capturing group.

If the 2nd arg is a Parser, calls parsec.ParseText with it on the Text.

TODO: Write example function.
*/
func RightShift(x, y interface{}) interface{} {
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
		case parsec.Parser:
			r, _ := parsec.ParseItem(y, x.(u8.Text))
			return r
		case func(...interface{}) interface{}:
			r, _ := parsec.ParseItem(ToParser(unwrapFunc(y)), x.(u8.Text))
			return r
		default:
			//MAYBE TODO: also use Parser, calling parsec.Search ???
			panic("RightShift: invalid input type" + fmt.Sprintf("; %T; %T", x, y) +
				"; must be codepoint, string, char class, or regex")
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

	//TODO: case Map

	case Slice:
		ax := x.(Slice)
		ns := NewSlice(0)
		for _, a := range ax {
			if !IsEqual(a, y) {
				_ = LeftShiftAssign(&ns, a)
			}
		}
		return ns.(Slice)

	default:
		panic("RightShift: invalid input type" + fmt.Sprintf("; %T; %T", x, y))
	}
}

//==============================================================================

/*
FindFirst returns a slice of RegexGroupMatch structs.
The 0-index in the slice will contain the full match details.
The subsequent indexes in the slice will contain the matches for each capturing group.

TODO: Write logic; test it; write example function.
*/
func FindFirst(x interface{}, s string) []RegexMatch {
	panic("FindFirst: invalid first arg")
}

//==============================================================================
