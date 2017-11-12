// Copyright 2009-17 The Go and Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package utf88

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type (
	Codepoint int32       //TODO: remove, use rune instead
	Text      []Codepoint //TODO: change to []rune
)

const (
	rune1ByteMax = 0x7f
	rune2ByteMax = 0x7ff
	gapMin       = 0xd800 // utf-16 surrogates not valid
	gapMax       = 0xdfff
	runeError    = 0xfffd // the "error" Rune or replacement character
	rune3ByteMax = 0xffff
	surrBase     = 0xf8000
	leadSurrMin  = 0xf8011
	leadSurrMax  = 0xfffbf
	trailSurrMin = 0x100000
	trailSurrMax = 0x10ffff // maximum valid Unicode code point encodable without using utf-88 surrogates
	surrSelf     = 0x110000

	MaxPoint = 0x7fbfffff // maximum valid Unicode code point encodable with utf-88 surrogates
)

var _C = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0x0001, 0x001f, 1},
		{0x007f, 0x009f, 1},
		{0x00ad, 0x0600, 1363},
		{0x0601, 0x0604, 1},
		{0x061c, 0x06dd, 193},
		{0x070f, 0x180e, 4351},
		{0x200b, 0x200f, 1},
		{0x202a, 0x202e, 1},
		{0x2060, 0x2064, 1},
		{0x2066, 0x206f, 1},
		{0xd800, 0xf8ff, 1},
		{0xfeff, 0xfff9, 250},
		{0xfffa, 0xfffb, 1},
	},
	R32: []unicode.Range32{
		{0x110bd, 0x1d173, 49334},
		{0x1d174, 0x1d17a, 1},
		{0xe0001, 0xe0020, 31},
		{0xe0021, 0xe007f, 1},
		{0xf0000, 0xf7fff, 1},   //changed from {0xf0000, 0xffffd, 1},
		{0xf8011, 0xfffbf, 1},   //added
		{0x100000, 0x1fffff, 1}, //changed from {0x100000, 0x10fffd, 1},
	},
	LatinOffset: 2,
}

var _Co = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0xe000, 0xf8ff, 1},
	},
	R32: []unicode.Range32{
		{0xf0000, 0xf7fff, 1},   //changed from {0xf0000, 0xffffd, 1},
		{0x100000, 0x1fffff, 1}, //changed from {0x100000, 0x10fffd, 1},
	},
}

var _Cs = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0xd800, 0xdfff, 1},
	},
	R32: []unicode.Range32{ //added
		{0xf8011, 0xfffbf, 1},
	},
}

var (
	Cc     = unicode.Cc
	Cf     = unicode.Cf
	Co     = _Co // Private Use codepoints have top half of plane 0xF removed, and 15 planes from 0x11 to 0x1F added
	Cs     = _Cs // Surrogate codepoints have most of top half of plane 0xF added
	Other  = _C
	C      = _C
	Zs     = unicode.Zs
	Zl     = unicode.Zl
	Zp     = unicode.Zp
	Space  = unicode.Z
	Z      = unicode.Z
	Mc     = unicode.Mc
	Me     = unicode.Me
	Mn     = unicode.Mn
	Mark   = unicode.M
	M      = unicode.M
	Pc     = unicode.Pc
	Pd     = unicode.Pd
	Ps     = unicode.Ps
	Pe     = unicode.Pe
	Pi     = unicode.Pi
	Pf     = unicode.Pf
	Po     = unicode.Po
	Punct  = unicode.P
	P      = unicode.P
	Sc     = unicode.Sc
	Sk     = unicode.Sk
	Sm     = unicode.Sm
	So     = unicode.So
	Symbol = unicode.S
	S      = unicode.S
	Digit  = unicode.Nd
	Nd     = unicode.Nd
	Nl     = unicode.Nl
	No     = unicode.No
	Number = unicode.N
	N      = unicode.N
	Lower  = unicode.Ll
	Ll     = unicode.Ll
	Upper  = unicode.Lu
	Lu     = unicode.Lu
	Title  = unicode.Lt
	Lt     = unicode.Lt
	Lm     = unicode.Lm
	Lo     = unicode.Lo
	Letter = unicode.L
	L      = unicode.L
)

var _Noncharacter_Code_Point = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0xfdd0, 0xfdef, 1},
		{0xfffe, 0xffff, 1},
	},
	R32: []unicode.Range32{
		{0x1fffe, 0x1ffff, 1},
		{0x2fffe, 0x2ffff, 1},
		{0x3fffe, 0x3ffff, 1},
		{0x4fffe, 0x4ffff, 1},
		{0x5fffe, 0x5ffff, 1},
		{0x6fffe, 0x6ffff, 1},
		{0x7fffe, 0x7ffff, 1},
		{0x8fffe, 0x8ffff, 1},
		{0x9fffe, 0x9ffff, 1},
		{0xafffe, 0xaffff, 1},
		{0xbfffe, 0xbffff, 1},
		{0xcfffe, 0xcffff, 1},
		{0xdfffe, 0xdffff, 1},
		{0xefffe, 0xeffff, 1},
		{0xffffe, 0xfffff, 1},
		//removed {0x10fffe, 0x10ffff, 1},
	},
}

var (
	ASCII_Hex_Digit                    = unicode.ASCII_Hex_Digit
	Bidi_Control                       = unicode.Bidi_Control
	Dash                               = unicode.Dash
	Deprecated                         = unicode.Deprecated
	Diacritic                          = unicode.Diacritic
	Extender                           = unicode.Extender
	Hex_Digit                          = unicode.Hex_Digit
	Hyphen                             = unicode.Hyphen
	IDS_Binary_Operator                = unicode.IDS_Binary_Operator
	IDS_Trinary_Operator               = unicode.IDS_Trinary_Operator
	Ideographic                        = unicode.Ideographic
	Join_Control                       = unicode.Join_Control
	Logical_Order_Exception            = unicode.Logical_Order_Exception
	Noncharacter_Code_Point            = _Noncharacter_Code_Point
	Other_Alphabetic                   = unicode.Other_Alphabetic
	Other_Default_Ignorable_Code_Point = unicode.Other_Default_Ignorable_Code_Point
	Other_Grapheme_Extend              = unicode.Other_Grapheme_Extend
	Other_ID_Continue                  = unicode.Other_ID_Continue
	Other_ID_Start                     = unicode.Other_ID_Start
	Other_Lowercase                    = unicode.Other_Lowercase
	Other_Math                         = unicode.Other_Math
	Other_Uppercase                    = unicode.Other_Uppercase
	Pattern_Syntax                     = unicode.Pattern_Syntax
	Pattern_White_Space                = unicode.Pattern_White_Space
	Quotation_Mark                     = unicode.Quotation_Mark
	Radical                            = unicode.Radical
	STerm                              = unicode.STerm
	Soft_Dotted                        = unicode.Soft_Dotted
	Terminal_Punctuation               = unicode.Terminal_Punctuation
	Unified_Ideograph                  = unicode.Unified_Ideograph
	Variation_Selector                 = unicode.Variation_Selector
	White_Space                        = unicode.White_Space
)

// String returns a string representation of the Codepoint between single quotes.
func (r Codepoint) String() string {
	return "'" + string(r) + "'"
}

// String returns a string representation of the Text.
func (t Text) String() string {
	return Surr(t)
}

// SurrogatePoint returns the surrogate-based encoding for the given codepoint.
// If the point does not need encoding, returns an array containing the given point as a rune.
// If the point is not valid UTF-88, returns the array containing error rune U+FFFD.
func SurrogatePoint(r Codepoint) []rune {
	if r < 0 || leadSurrMin <= r && r <= leadSurrMax || r > MaxPoint {
		return []rune{runeError}
	}
	if r <= trailSurrMax {
		return []rune{rune(r)}
	}
	return []rune{surrBase + (rune(r)>>16)&0xffff, trailSurrMin + rune(r)&0xffff}
}

// SurrogatePoints returns the surrogate-based encoding of the codepoint sequence s in
// an array. If any point does not need encoding, returns the given point as a rune in the
// array. If any point is not valid UTF-88, returns error rune U+FFFD in the array.
func SurrogatePoints(s Text) []rune {
	a := make([]rune, 2*len(s)) //doesn't call SurrogatePoint for time efficiency reasons
	n := 0
	for _, r := range s {
		switch {
		case r < 0, leadSurrMin <= r && r <= leadSurrMax, r > MaxPoint:
			r = runeError
			fallthrough
		case r <= trailSurrMax:
			a[n] = rune(r)
			n++
		default:
			a[n], a[n+1] = surrBase+(rune(r)>>16)&0xffff, trailSurrMin+rune(r)&0xffff
			n += 2
		}
	}
	return a[0:n]
}

// Surr returns the surrogate-based string encoding for a sequence of the given Text.
// If any point does not need encoding, returns the given point as a rune in the string.
// If any point is not valid UTF-88, returns error rune U+FFFD.
func Surr(s Text) string {
	return string(SurrogatePoints(s))
}

// Sur returns the surrogate-based string encoding for a sequence of the
// given codepoints. If any point does not need encoding, returns the given point as a rune in
// the string. If any point is not valid UTF-88, returns error rune U+FFFD.
func Sur(rs ...Codepoint) string {
	if len(rs) == 1 {
		return string(SurrogatePoint(rs[0]))
	} else if len(rs) > 1 {
		return string(SurrogatePoints(rs))
	} else {
		return ""
	}
}

// ParsePoint parses a string with a codepoint encoded in a similar manner to Go's
// native syntax for runes, but without the backslash. Runes invalid in Go such as UTF-16
// surrogate halves and those above U+10FFFF are accepted as valid codepoints, e.g. "ud800",
// "U7fdfffff".
func ParsePoint(s string) Codepoint {
	r := regexp.MustCompile(`U[0-9A-Fa-f]{8}|u[0-9A-Fa-f]{4}`)
	m := r.FindString(s)
	if m != "" {
		z, _ := strconv.ParseInt(m[1:], 16, 32)
		return Codepoint(z)
	} else {
		panic("ParsePoint: arg must encode valid codepoint in string, either Uxxxxxxxx or uxxxx")
	}
}

// ParsePoints parses a string with one or more codepoints encoded in a similar manner
// to Go's native syntax for runes, but without the backslashes. Runes invalid in Go such
// as UTF-16 surrogate halves and those above U+10FFFF are accepted as valid codepoints, e.g.
// "ud800U7fdfffff".
func ParsePoints(s string) Text {
	r := regexp.MustCompile(`U[0-9A-Fa-f]{8}|u[0-9A-Fa-f]{4}`)
	m := r.FindAllString(s, -1)
	if m != nil {
		a := make(Text, len(m))
		for i, o := range m {
			z, _ := strconv.ParseInt(o[1:], 16, 32)
			a[i] = Codepoint(z)
		}
		return a
	} else {
		panic("ParsePoints: arg must encode a sequence of valid codepoints in string, either Uxxxxxxxx or uxxxx")
	}
}

// Decode returns valid UTF-8 decoding of a string. It substitutes U+FFFD for
// any incorrect UTF-8 such as a rune out of range or an encoding that is not
// the shortest possible for a rune.
func Decode(s string) string {
	p := make([]rune, 4*len(s))
	n := 0
	for _, r := range s {
		p[n] = r
		n++
	}
	return string(p[:n])
}

// DesurrogateFirstPoint desurrogates the first UTF-88 surrogated rune in s, returning
// the desurrogated codepoint and its surrogated width in runes. If the first rune is
// out of range (U+0 to U+10ffff), or the first rune is a leading surrogate but
// the second rune is not a trailing one, returns (U+FFFD, 1).
func DesurrogateFirstPoint(s []rune) (Codepoint, int) {
	r := s[0]
	if leadSurrMin <= r && r <= leadSurrMax && len(s) >= 2 && trailSurrMin <= s[1] && s[1] <= trailSurrMax {
		return (Codepoint(r)-surrBase)<<16 | (Codepoint(s[1]) - trailSurrMin), 2
	}
	if leadSurrMin <= r && r <= leadSurrMax || r < 0 || r >= surrSelf {
		return Codepoint(runeError), 1
	}
	return Codepoint(r), 1
}

// Desur returns the desurrogated codepoints represented by the surrogated string s.
func Desur(s string) Text {
	p := []rune(s)
	a := make(Text, len(p))
	n := 0
	for i := 0; i < len(p); {
		r, size := DesurrogateFirstPoint(p[i:])
		a[n] = r
		n++
		i += size
	}
	return a[0:n]
}

// Concat concatenates the disjoint string or surrogated Text items together, then
// returns the desurrogated resulting string as Text.
func Concat(ps ...interface{}) Text {
	s := ""
	for _, p := range ps {
		switch p.(type) {
		case Text:
			s += Sur(p.(Text)...)
		case string:
			s += p.(string)
		default:
			panic("Concat: invalid input type")
		}
	}
	return Desur(s)
}

// Join concatenates the array of Texts with the joiner between each.
func Join(ts []Text, j Text) Text {
	ss := []string{}
	for _, t := range ts {
		ss = append(ss, Sur(t...))
	}
	return Desur(strings.Join(ss, Sur(j...)))
}

// DesurrogateLastRune desurrogates the last UTF-88 surrogated rune in s,
// returning the desurrogated rune and its surrogated width in runes. If the last
// rune is out of range (U+0 to U+10ffff), returns (U+FFFD, 1).
func DesurrogateLastPoint(s []rune) (Codepoint, int) {
	r := s[len(s)-1]
	if trailSurrMin <= r && r <= trailSurrMax && len(s) >= 2 && leadSurrMin <= s[len(s)-2] && s[len(s)-2] <= leadSurrMax {
		return (Codepoint(s[len(s)-2])-surrBase)<<16 | (Codepoint(r) - trailSurrMin), 2
	}
	if leadSurrMin <= r && r <= leadSurrMax || r < 0 || r >= surrSelf {
		return Codepoint(runeError), 1
	}
	return Codepoint(r), 1
}

// PointCountOfRunes returns the number of unsurrogated codepoints in surrogated
// rune array p. Lead surrogates not followed by a trailing one and runes out of
// range are treated as single unsurrogated codepoints.
func PointCountOfRunes(p []rune) int {
	i := 0
	var n int
	for n = 0; i < len(p); n++ {
		r := p[i]
		if leadSurrMin <= r && r <= leadSurrMax && len(p) >= 2 && trailSurrMin <= p[i+1] && p[i+1] <= trailSurrMax {
			i += 2
		} else {
			i++
		}
	}
	return n
}

// PointCountOfBytes returns the number of unsurrogated codepoints in UTF-8 bytes s.
// Erroneous and short encodings are treated as single unsurrogated codepoints.
// Bytes which decode to lead surrogates not followed by a trailing one and runes
// out of range are treated as single unsurrogated codepoints.
func PointCountOfBytes(s []byte) int {
	return PointCountOfRunes([]rune(string(s)))
}

// ValidRunes reports whether runes consist entirely of valid UTF-88 surrogated runes.
func ValidRunes(p []rune) bool {
	i := 0
	for i < len(p) {
		r := p[i]
		if leadSurrMin <= r && r <= leadSurrMax && len(p) >= 2 && trailSurrMin <= p[1] && p[1] <= trailSurrMax {
			i += 2
		} else if leadSurrMin <= r && r <= leadSurrMax || r < 0 || r >= surrSelf {
			return false
		} else {
			i++
		}
	}
	return true
}

// ValidBytes report whether bytes consist entirely of valid UTF-8 encoded runes
// which decode only to valid UTF-88 surrogated runes.
func ValidBytes(s []byte) bool {
	var p []rune
	ss := string(s)
	p = make([]rune, len(s))
	n := 0
	for i, r := range ss {
		//if runeError is error sentinel and not itself encoded properly
		if r == runeError && !(i+3 <= len(ss) && ss[i:i+3] == "\xef\xbf\xbd") {
			return false
		}
		p[n] = r
		n++
	}
	p = p[:n]
	return ValidRunes(p)
}

// LenInRunes returns the number of runes required to surrogate the codepoint,
// or 1 for an invalid codepoint.
func LenInRunes(r Codepoint) int {
	if r < 0 || leadSurrMin <= r && r <= leadSurrMax || r > MaxPoint || r <= trailSurrMax {
		return 1
	}
	return 2
}

// LenInBytes returns the number of bytes required to surrogate and encode the codepoint.
// It returns -1 if the codepoint is not a valid value to surrogate in UTF-88 then encode in UTF-8.
func LenInBytes(r Codepoint) int {
	switch {
	case r < 0:
		return -1
	case r <= rune1ByteMax:
		return 1
	case r <= rune2ByteMax:
		return 2
	case gapMin <= r && r <= gapMax:
		return -1
	case r <= rune3ByteMax:
		return 3
	case leadSurrMin <= r && r <= leadSurrMax:
		return -1
	case r <= trailSurrMax:
		return 4
	case r <= MaxPoint:
		return 8
	default:
		return -1
	}
}

// ValidForSurrogation reports whether codepoint r can be legally surrogated as UTF-88.
// Codepoints out of range and leading utf-88 surrogates are illegal.
func ValidForSurrogation(r Codepoint) bool {
	if r < 0 || leadSurrMin <= r && r <= leadSurrMax || r > MaxPoint {
		return false
	}
	return true
}

// ValidForEncoding reports whether codepoint r can be legally encoded as UTF-8.
// Codepoints out of range, UTF-16 surrogate halves, and leading UTF-88 surrogates are illegal.
func ValidForEncoding(r Codepoint) bool {
	switch {
	case r < 0:
		return false
	case gapMin <= r && r <= gapMax:
		return false
	case leadSurrMin <= r && r <= leadSurrMax:
		return false
	case r > MaxPoint:
		return false
	}
	return true
}
