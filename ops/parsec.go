// Copyright 2017 The Gro authors. All rights reserved.
// Portions translated from Armando Blancas's Clojure-based Kern parser.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package ops

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	u8 "github.com/grolang/gro/utf88"
)

//==============================================================================
// Parser Structure
//------------------------------------------------------------------------------

// Parsec is the primary structure in groo, a lazily-called function passed between various combinators,
// enabling complex top-down parsing structures to be built.
type Parsec = func(PState) PState

// PState is a structure of privately-used data used by the Parser.
type PState struct {
	value Any        // the value of the parsed input
	input string     // the input sequence
	pos   uint64     // the position into the input
	ok    bool       // whether the parser terminated without error
	empty bool       // whether the parser consumed nothing from the input
	user  Any        // an object stored by the client code
	error string     // the last error collected during parsing
	depth int        // the current indent depth for logging messages
	log   []bool     // a stack of booleans indicating whether logging is on or off
	flags []logFlags // not yet used
}

// InputAndPosition is a structure returned when GetInputAndPosition parser called,
// and structure used when SetInputAndPosition parser called.
type InputAndPosition struct {
	inp u8.Text // the input sequence
	pos uint64  // the position into the input
}

//------------------------------------------------------------------------------

// String returns the privately-used data within the PState structure as a formatted string.
func (ps PState) String() string {
	so := ps.format(0)

	if ps.depth != 0 {
		so = so + fmt.Sprintf(", depth:%d", ps.depth)
	}
	if ps.log != nil && len(ps.log) != 0 {
		so = so + fmt.Sprintf(", log:%v", ps.log)
	}
	if ps.flags != nil && len(ps.flags) != 0 {
		so = so + fmt.Sprintf(", flags:%v", ps.flags)
	}

	return so
}

//------------------------------------------------------------------------------
const (
	errEmptyStr = "Empty string given."
	errInvalStr = "Invalid string given."
	errNoParser = "No parser given."
	errNoAlts   = "No alternatives selected."
	errUnexpEof = "Unexpected end-of-file reached."
	errParsFail = "Parser failed."
)

func makeUnexpInp(s string) string {
	return "Unexpected " + s + " input."
}

//------------------------------------------------------------------------------
func (ps PState) format(indent int) (so string) {
	const maxlen = 120
	so = fmt.Sprintf("%q/%d", ps.input, ps.pos)
	tally := indent + len(so)

	addField := func(field string) {
		if tally+2+len(field) > maxlen-1 {
			so += "\n" + strings.Repeat(" ", indent) + field
			tally = indent + len(field)
		} else {
			so += ", " + field
			tally += 2 + len(field)
		}
	}

	if ps.value != nil {
		addField(fmt.Sprintf("value:%q", ps.value)) //OR IS IT: "value:%v(%[1]T)"
	}
	if ps.ok {
		addField("ok")
	}
	if ps.empty {
		addField("empty")
	}
	if ps.user != nil && len(fmt.Sprintf("%v", ps.user)) > 0 {
		addField(fmt.Sprintf("user:%v(%[1]T)", ps.user))
	}

	if ps.error == errNoAlts {
		addField("errNoAlts")
	} else if ps.error == errEmptyStr {
		addField("errEmpStr")
	} else if ps.error == errNoParser {
		addField("errNoParsr")
	} else if ps.error == errUnexpEof {
		addField("errUnexpEof")
	} else if ps.error == errParsFail {
		addField("errParsFail")
	} else if ps.error != "" {
		addField("error:" + ps.error)
	}

	return
}

//------------------------------------------------------------------------------
func (ps PState) clone(ff func(*PState)) PState {
	so := PState{input: ps.input, pos: ps.pos,
		value: ps.value, ok: ps.ok, empty: ps.empty,
		user: ps.user, error: ps.error,
		depth: ps.depth, log: ps.log, flags: ps.flags}
	ff(&so)
	return so
}

//==============================================================================
// Logging Utilities
//------------------------------------------------------------------------------
var log = map[string]bool{}

// not yet used...
type logFlags struct {
}

const loggingEnabled = true

func logFnBA(tag string, si, so *PState) {
	if si.input == so.input && si.pos == so.pos {
		fmt.Printf("%s= %s, INP/POS UNCHANGED: %v\n",
			strings.Repeat(". ", si.depth),
			tag,
			so.format(so.depth*2+len(tag)+23))
	} else {
		fmt.Printf("%s= %s, BEFORE: %v;\n%sAFTER: %v\n",
			strings.Repeat(". ", si.depth),
			tag,
			si.format(si.depth*2+len(tag)+12),
			strings.Repeat(" ", si.depth*2+len(tag)+5),
			so.format(so.depth*2+len(tag)+12))
	}
}

func logFnEq(tag string, so *PState) {
	fmt.Printf("%s= %s %v\n",
		strings.Repeat(". ", so.depth),
		tag,
		so.format(so.depth*2+len(tag)+2))
}

func logFnIn(tag string, si *PState) {
	fmt.Printf("%s> %s %v\n",
		strings.Repeat(". ", si.depth),
		tag,
		si.format(si.depth*2+len(tag)+3))
	si.depth++
}

func logFnDe(tag string, so *PState) {
	so.depth--
	fmt.Printf("%s< %s %v\n",
		strings.Repeat(". ", so.depth),
		tag,
		so.format(so.depth*2+len(tag)+4))
}

//==============================================================================
// Parser Executor
//------------------------------------------------------------------------------

// ParseItem accepts a Parsec and some text, utf88-surrogates it to a string,
// and applies the parser to it. It returns the value if it succeeds,
// and returns the error if it fails.
func ParseItem(p Parsec, y Any) (result Any, err error) {
	switch y.(type) {
	case u8.Text:
		tx := y.(u8.Text)
		s := p(PState{input: u8.Surr(tx), ok: true, empty: true})
		if s.ok {
			return s.value, nil
		} else {
			return nil, errors.New(s.error)
		}
	default:
		panic(fmt.Sprintf("ParseText: illegal type input %T", y))
	}
}

//==============================================================================
// Basic Parsers
//------------------------------------------------------------------------------

// Return succeeds without consuming any input, and returns its argument v as the resulting value.
// Any carried errors are removed.
func Return(v Any) Parsec {
	return func(st PState) (so PState) {
		if log["Return"] || (len(st.log) > 0 && st.log[len(st.log)-1] && loggingEnabled) {
			defer logFnEq("Return", &so)
		}
		so = st.clone(func(e *PState) {
			e.value = v
			e.ok = true
			//QUERY: what about e.empty= false ?
			e.error = ""
		})
		return
	}
}

// Fail fails without consuming any input, having a single error message msg encoded as utf88 text.
func Fail(msg u8.Text) Parsec {
	return func(st PState) (so PState) {
		if log["Fail"] || (len(st.log) > 0 && st.log[len(st.log)-1] && loggingEnabled) {
			defer logFnEq("Fail", &so)
		}
		so = st.clone(func(e *PState) {
			e.value = nil
			e.ok = false
			e.empty = true
			e.error = u8.Surr(msg)
		})
		return
	}
}

// Satisfy succeeds if the next character satisfies the predicate pred, in which case
// advances the position of the input stream. It may fail on an unexpected end of input.
func Satisfy(pred func(u8.Codepoint) bool) Parsec {
	return func(st PState) (so PState) {
		if log["Satisfy"] || (len(st.log) > 0 && st.log[len(st.log)-1] && loggingEnabled) {
			defer logFnBA("Satisfy", &st, &so)
		}
		if st.input == "" {
			so = st.clone(func(e *PState) {
				e.value = nil
				e.ok = false
				e.empty = true
				e.error = errUnexpEof
			})
			return
		}
		c, _ := u8.DesurrogateFirstPoint([]rune(st.input))
		if pred(c) {
			so = st.clone(func(e *PState) {
				e.input = st.input[u8.LenInBytes(c):]
				e.pos = st.pos + uint64(u8.LenInBytes(c))
				e.value = c
				e.ok = true
				e.empty = false
				e.error = ""
			})
			return
		} else {
			so = st.clone(func(e *PState) {
				e.value = nil
				e.ok = false
				e.empty = true
				e.error = makeUnexpInp(u8.Sur(c))
			})
			return
		}
	}
}

//==============================================================================
// Optimized Primitive Parsers
//------------------------------------------------------------------------------

// Symbol succeeds if the next Codepoint equals the given Codepoint, in which case it
// increments the position of the input stream. It may fail on an unexpected end of input.
func Symbol(t u8.Codepoint) Parsec {
	sym := u8.Sur(t)
	return func(st PState) (so PState) {
		if log["Symbol"] || (len(st.log) > 0 && st.log[len(st.log)-1] && loggingEnabled) {
			defer logFnBA("Symbol", &st, &so)
		}
		if len(sym) == 0 {
			so = st.clone(func(e *PState) {
				e.value = nil
				e.ok = false
				e.empty = true
				e.error = errEmptyStr
			})
			return
		} else if len(sym) > 1 && st.input[:len(sym)] == sym {
			so = st.clone(func(e *PState) {
				e.input = st.input[len(sym):]
				e.pos = st.pos + uint64(len(sym))
				e.value = t
				e.ok = true
				e.empty = false
				e.error = ""
			})
			return
		} else if len(sym) > 1 {
			so = st.clone(func(e *PState) {
				e.value = nil
				e.ok = false
				e.empty = true
				e.error = errInvalStr
			})
			return
		} else if st.input == "" {
			so = st.clone(func(e *PState) {
				e.value = nil
				e.ok = false
				e.empty = true
				e.error = errUnexpEof
			})
			return
		} else if st.input[:1] == sym {
			so = st.clone(func(e *PState) {
				e.input = st.input[1:]
				e.pos = st.pos + 1
				e.value = t
				e.ok = true
				e.empty = false
				e.error = ""
			})
			return
		} else {
			so = st.clone(func(e *PState) {
				e.value = nil
				e.ok = false
				e.empty = true
				e.error = makeUnexpInp(st.input[:1])
			})
			return
		}
	}
}

// Token succeeds if the next Codepoint/s equals the given Text, in which case it
// advances the position of the input stream. It may fail on an unexpected end of input.
func Token(t u8.Text) Parsec {
	tok := u8.Surr(t)
	return func(st PState) (so PState) {
		if log["Token"] || (len(st.log) > 0 && st.log[len(st.log)-1] && loggingEnabled) {
			defer logFnBA("Token", &st, &so)
		}
		if len(tok) == 0 {
			so = st.clone(func(e *PState) {
				e.value = nil
				e.ok = false
				e.empty = true
				e.error = errEmptyStr
			})
			return
		} else if st.input == "" || len(st.input) < len(tok) {
			so = st.clone(func(e *PState) {
				e.value = nil
				e.ok = false
				e.empty = true
				e.error = errUnexpEof
			})
			return
		} else if st.input[:len(tok)] == tok {
			so = st.clone(func(e *PState) {
				e.input = st.input[len(tok):]
				e.pos = st.pos + uint64(len(tok))
				e.value = t
				e.ok = true
				e.empty = false
				e.error = ""
			})
			return
		} else {
			so = st.clone(func(e *PState) {
				e.value = nil
				e.ok = false
				e.empty = true
				e.error = makeUnexpInp(st.input[:len(tok)])
			})
			return
		}
	}
}

// Regexp succeeds if the next Character/s mathes the given regex string, in which case it
// advances the position of the input stream. It may fail on an unexpected end of input.
func Regexp(tok string) Parsec {
	return func(st PState) (so PState) {
		if log["Regexp"] || (len(st.log) > 0 && st.log[len(st.log)-1] && loggingEnabled) {
			defer logFnBA("Regexp", &st, &so)
		}

		if len(tok) == 0 {
			so = st.clone(func(e *PState) {
				e.value = nil
				e.ok = false
				e.empty = true
				e.error = errEmptyStr
			})
			return
		} else if st.input == "" {
			so = st.clone(func(e *PState) {
				e.value = nil
				e.ok = false
				e.empty = true
				e.error = errUnexpEof
			})
			return
		} else {
			r := regexp.MustCompile("^(?:" + tok + ")") //QUERY: change to \A ?
			loc := r.FindStringIndex(st.input)
			if loc != nil && loc[0] == 0 {
				so = st.clone(func(e *PState) {
					e.input = st.input[loc[1]:]
					e.pos = st.pos + uint64(loc[1])
					e.value = u8.Desur(st.input[0:loc[1]])
					e.ok = true
					e.empty = false
					e.error = ""
				})
				return
			} else {
				so = st.clone(func(e *PState) {
					e.value = nil
					e.ok = false
					e.empty = true
					e.error = makeUnexpInp(st.input[:1])
				})
				return
			}
		}
	}
}

//==============================================================================
// Primitive Parsers
//------------------------------------------------------------------------------

// AnyChar succeeds with any character.
var AnyChar = Satisfy(func(c u8.Codepoint) bool {
	return true
})

// OneOf succeeds if the next character is in the supplied text.
func OneOf(s u8.Text) Parsec {
	return Satisfy(func(c u8.Codepoint) bool {
		for _, n := range s {
			if n == c {
				return true
			}
		}
		return false
	})
}

// NoneOf succeeds if the next character is not in the supplied text.
func NoneOf(s u8.Text) Parsec {
	return Satisfy(func(c u8.Codepoint) bool {
		for _, n := range s {
			if n == c {
				return false
			}
		}
		return true
	})
}

// Tokens is like Token but accepts more than one Text. It will try each such choice in turn.
func Tokens(ts ...u8.Text) Parsec {
	switch len(ts) {
	case 0:
		return Fail(u8.Desur(errNoAlts))
	case 1:
		return Token(ts[0])
	default:
		return AltParse(Try(Token(ts[0])), Tokens(ts[1:]...)).(Parsec)
	}
}

// Field parses an unquoted text field terminated by any character in cs.
func Field(cs u8.Text) Parsec {
	return Many(NoneOf(cs))
}

//==============================================================================
// Directly used by Groo syntax
//------------------------------------------------------------------------------

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
func ParserRepeat(a Any, n Any) Any {
	switch n.(type) {
	case Int:
		return Times(n.(Int), a.(Parsec))
	case Pair:
		from := n.(Pair).From()
		inf := n.(Pair).IsToInf()
		to := n.(Pair).To()
		if from == 0 && inf {
			return Many(a.(Parsec))
		} else if from == to {
			return Times(from, a.(Parsec))
		} else if from == 1 && inf {
			return Many1(a.(Parsec))
		} else if from == 0 && to == 1 {
			return Optional(a.(Parsec))
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
// Parser Combinators
//------------------------------------------------------------------------------

// Fwd delays the evaluation of a parser that was forward-declared but
// defined recursively. For use in defs of parsers taking arguments.
func Fwd(fps ...Any) Any {
	if len(fps) <= 0 {
		panic("Fwd: not enough args.")
	}
	fp, ok := fps[0].(func(...Any) Any)
	if !ok {
		panic(fmt.Sprintf("Fwd: first arg wrong type: %T", fps[0]))
	}
	return func(s PState) (so PState) {
		if log["Fwd"] || (len(s.log) > 0 && s.log[len(s.log)-1] && loggingEnabled) {
			logFnIn("Fwd", &s)
			defer logFnDe("Fwd", &so)
		}
		so = ToParser(fp(fps[1:]))(s)
		return
	}
}

//------------------------------------------------------------------------------

// Try parses p; on failure it pretends it did not consume any input.
func Try(p Parsec) Parsec {
	return func(s PState) (so PState) {
		if log["Try"] || (len(s.log) > 0 && s.log[len(s.log)-1] && loggingEnabled) {
			logFnIn("Try", &s)
			defer logFnDe("Try", &so)
		}
		st := p(s)
		if st.ok {
			so = st
			return
		} else {
			so = st.clone(func(e *PState) {
				e.input = s.input
				e.pos = s.pos
				//QUERY: what about e.ok ?
				e.empty = true
			})
			return
		}
	}
}

// NotFollowedBy succeeds only if p fails; consumes no input.
func NotFollowedBy(p Parsec) Parsec {
	return Try(AltParse(Bind(Try(p), func(x Any) Parsec {
		return Fail(u8.Desur(errParsFail))
	}),
		Return(nil)).(Parsec))
}

// Eof succeeds on end of input.
var Eof = NotFollowedBy(AnyChar)

//------------------------------------------------------------------------------

// Ask parses p and if it fails consuming no input, will replace the error
// with the message "Expecting" m. This helps to produce more abstract and
// accurate error messages.
func Ask(p Parsec, m u8.Text) Parsec {
	return func(s PState) PState {
		st := p(s)
		if !st.ok && st.empty {
			return st.clone(func(e *PState) {
				e.error = st.error + " Expecting " + u8.Sur(m...)
			})
		} else {
			return st
		}
	}
}

// Expect parses p and if it fails (regardless of input consumed) will replace
// the error with the message "Expecting" m. This helps to produce more abstract
// and accurate error messages.
func Expect(p Parsec, m u8.Text) Parsec {
	return func(s PState) PState {
		st := p(s)
		if st.ok {
			return st
		} else {
			return st.clone(func(e *PState) {
				e.error = st.error + " Expecting " + u8.Sur(m...)
			})
		}
	}
}

//------------------------------------------------------------------------------

// AltParse tries each parser in ps; if any fails without consuming any input, it tries the
// next one. It will stop and succeed if a parser succeeds; it will stop and fail
// if a parser fails consuming input; or it will try the next one if a parser fails
// without consuming input.
func AltParse(ps ...Any) Any {
	switch len(ps) {
	case 0:
		return Fail(u8.Desur(errNoParser))
	case 1:
		return ps[0]
	case 2:
		p := ps[0].(Parsec)
		q := ps[1].(Parsec)
		return func(st PState) (so PState) {
			if log["AltParse"] || (len(st.log) > 0 && st.log[len(st.log)-1] && loggingEnabled) {
				logFnIn("AltParse", &st)
				defer logFnDe("AltParse", &so)
			}
			s2 := p(st)
			if !s2.ok && s2.empty {
				s3 := q(st)
				if s3.ok {
					so = s3
					return
				} else {
					so = s3.clone(func(e *PState) {
						//QUERY: what about e.ok , etc?
						e.error = errNoAlts
					})
					return
				}
			} else {
				so = s2
				return
			}
		}

	default:
		return AltParse(ps[0], AltParse(ps[1:]...))
	}
}

//------------------------------------------------------------------------------

// Bind binds parser p to function f which gets p's value and returns a new parser.
// Function p must define a single parameter. The argument it receives is the value
// parsed by p, not ps' return value, which is a parser state record.
func Bind(p Parsec, f func(Any) Parsec) Parsec {
	return func(s PState) (so PState) {
		if log["Bind"] || (len(s.log) > 0 && s.log[len(s.log)-1] && loggingEnabled) {
			logFnIn("Bind: ", &s)
			defer logFnDe("Bind: ", &so)
		}
		s1 := p(s)
		if s1.ok {
			s2 := (f(s1.value))(s1)
			s3 := s2.clone(func(e *PState) {
				e.empty = s1.empty && s2.empty
			})
			if s3.ok {
				so = s3
				return
			} else {
				so = s3.clone(func(e *PState) {})
				return
			}
		} else {
			so = s1
			return
		}
	}
}

// SeqLeft parses each parser of ps in sequence; it keeps the first result and
// skips the rest.
func SeqLeft(ps ...Any) Any {
	switch len(ps) {
	case 0:
		return Fail(u8.Desur(errNoParser))
	case 1:
		return ps[0]
	case 2:
		return Bind(ps[0].(Parsec), func(x Any) Parsec {
			return SeqRight(ps[1], Return(x)).(Parsec)
		})
	default:
		return SeqLeft(ps[0], SeqLeft(ps[1:]...))
	}
}

// SeqRight parses each parser of ps in sequence; it skips all but last and keeps
// the result of the last.
func SeqRight(ps ...Any) Any {
	switch len(ps) {
	case 0:
		return Fail(u8.Desur(errNoParser))
	case 1:
		return ps[0]
	case 2:
		return Bind(ps[0].(Parsec), func(_ Any) Parsec {
			return ps[1].(Parsec)
		})
	default:
		return Bind(ps[0].(Parsec), func(_ Any) Parsec {
			return SeqRight(ps[1:]...).(Parsec)
		})
	}
}

//------------------------------------------------------------------------------

// Apply parses parser p; if successful, it applies f to the value parsed
// by p.
func Apply(f func(Any) Any, p Parsec) Parsec {
	return Bind(p, func(x Any) Parsec {
		return Return(f(x))
	})
}

// Between applies open, p, close; returns the value of p.
func Between(open, close, p Parsec) Parsec {
	return Try(SeqLeft(SeqRight(open, p), close).(Parsec))
}

//------------------------------------------------------------------------------

// Collect parses each parser of ps in sequence; collects the results in a slice,
// including nil values. If any parser fails, it stops immediately and fails.
func Collect(ps ...Parsec) Parsec {
	return func(s PState) (so PState) {
		if log["Collect"] || (len(s.log) > 0 && s.log[len(s.log)-1] && loggingEnabled) {
			logFnIn("Collect", &s)
			defer logFnDe("Collect", &so)
		}
		qs := ps
		so = s.clone(func(e *PState) {
			e.value = []Any{}
			e.empty = true
		})
		for so.ok && len(qs) > 0 {
			if st := qs[0](so); st.ok {
				so = st.clone(func(e *PState) {
					e.value = append(so.value.([]Any), st.value)
				})
			} else {
				so = st.clone(func(e *PState) {
					e.empty = so.empty && st.empty
				})
			}
			qs = qs[1:]
		}
		return
	}
}

// Flatten applies one or more parsers; flattens the result and
// converts it to text.
func Flatten(ps ...Parsec) Parsec {
	var f func(c Any) Any
	f = func(c Any) Any {
		s := ""
		for _, n := range c.([]Any) {
			if _, isText := n.(u8.Text); !isText {
				n = f(n)
			}
			s = s + u8.Surr(n.(u8.Text))
		}
		return u8.Desur(s)
	}
	switch len(ps) {
	case 0:
		return Fail(u8.Desur(errNoParser))
	case 1:
		return Apply(f, ps[0])
	default:
		return Apply(f, Collect(ps...))
	}
}

//------------------------------------------------------------------------------

// Many parses p zero or more times; returns the result(s) in a slice. It stops
// when p fails, but this parser succeeds.
func Many(p Parsec) Parsec {
	return func(s PState) (so PState) {
		if log["Many"] || (len(s.log) > 0 && s.log[len(s.log)-1] && loggingEnabled) {
			logFnIn("Many", &s)
			defer logFnDe("Many", &so)
		}
		st := p(s)
		vs := make([]Any, 0)
		em := true
		for st.ok && !st.empty {
			vs = append(vs, st.value)
			em = em && st.empty
			st = p(st)
		}
		if st.empty {
			so = st.clone(func(e *PState) {
				e.value = vs
				e.ok = true
				e.empty = em
				e.error = ""
			})
			return
		} else {
			so = st
			return
		}
	}
}

// Many0 is like Many but it won't set the empty flag in the state record.
// Use instead of Many if it comes last to avoid overriding non-empty parsing.
func Many0(p Parsec) Parsec {
	return func(s PState) PState {
		st := Many(p)(s)
		return st.clone(func(e *PState) {
			e.empty = false
		})
	}
}

// Many1 parses p one or more times and returns the result(s) in a slice. It stops
// when p fails, but this parser succeeds.
func Many1(p Parsec) Parsec {
	return Bind(p, func(x Any) Parsec {
		return Bind(Many(p), func(y Any) Parsec {
			return Return(append([]Any{x}, y.([]Any)...))
		})
	})
}

// Optional succeeds if p succeeds or if p fails without consuming input.
func Optional(p Parsec) Parsec {
	return func(s PState) (so PState) {
		if log["Optional"] || (len(s.log) > 0 && s.log[len(s.log)-1] && loggingEnabled) {
			logFnIn("Optional", &s)
			defer logFnDe("Optional", &so)
		}
		st := p(s)
		if st.ok || st.empty {
			so = st.clone(func(e *PState) {
				e.ok = true
				e.error = ""
			})
			return
		} else {
			so = st
			return
		}
	}
}

// Option applies p; if it fails without consuming input, it returns a value x
// as default.
func Option(x Any, p Parsec) Parsec {
	return func(s PState) (so PState) {
		if log["Option"] || (len(s.log) > 0 && s.log[len(s.log)-1] && loggingEnabled) {
			logFnIn("Option", &s)
			defer logFnDe("Option", &so)
		}
		st := p(s)
		if !st.ok && st.empty {
			so = st.clone(func(e *PState) {
				e.value = x
				e.ok = true
				e.error = ""
			})
			return
		} else {
			so = st
			return
		}
	}
}

//------------------------------------------------------------------------------

// Times applies p n times; collects the results in a slice.
//func Times(n int, p Parser) Parser {
func Times(n int64, p Parsec) Parsec {
	if n > 0 {
		as := make([]Parsec, n, n)
		for i := 0; int64(i) < n; i++ {
			as[i] = p
		}
		return Collect(as...)
	} else {
		return Return([]Any{})
	}
}

//------------------------------------------------------------------------------

// ManyTill parses zero or more p while trying end, until end succeeds.
// Returns the results in a slice.
func ManyTill(p, end Parsec) Parsec {
	var scan Parsec
	scan = AltParse(
		SeqRight(end, Return([]Any{})).(Parsec),
		Bind(p, func(x Any) Parsec {
			as := make([]Any, 0)
			as = append(as, x)
			return Bind(scan, func(y Any) Parsec {
				for _, n := range y.([]Any) {
					as = append(as, n)
				}
				return Return(as)
			})
		})).(Parsec)
	return scan
}

//------------------------------------------------------------------------------

// SepBy parses p zero or more times while parsing sep in between; collects the results
// of p in a slice.
func SepBy(sep, p Parsec) Parsec {
	return AltParse(SepBy1(sep, p), Return([]Any{})).(Parsec)
}

// SepBy1 parses p one or more times while parsing sep in between; collects the results
// of p in a slice.
func SepBy1(sep, p Parsec) Parsec {
	return Bind(p, func(x Any) Parsec {
		return Bind(Many(Try(SeqRight(sep, p).(Parsec))), func(y Any) Parsec {
			return Return(append([]Any{x}, y.([]Any)...))
		})
	})
}

// EndBy parses p zero or more times, separated and ended by applications of sep;
// returns the results of p in a slice.
func EndBy(sep, p Parsec) Parsec {
	return Many(Try(SeqLeft(p, sep).(Parsec)))
}

// EndBy1 parses p one or more times, separated and ended by applications of sep;
// returns the results of p in a slice.
func EndBy1(sep, p Parsec) Parsec {
	return Many1(Try(SeqLeft(p, sep).(Parsec)))
}

// SepEndBy parses p one or more times separated, and optionally ended, by sep;
// collects the results in a slice.
func SepEndBy(sep, p Parsec) Parsec {
	//return AltParse(SepEndBy1(sep, Try(p)), Return([]Any{})) //QUERY: should p be "try"ed?
	return AltParse(SepEndBy1(sep, p), Return([]Any{})).(Parsec)
}

// SepEndBy1 parses p zero or more times separated, and optionally ended, by sep;
// collects the results in a slice.
func SepEndBy1(sep, p Parsec) Parsec {
	return Bind(p, func(x Any) Parsec {
		return AltParse(Bind(SeqRight(sep, SepEndBy(sep, Try(p))).(Parsec), func(y Any) Parsec {
			return Return(append([]Any{x}, y.([]Any)...))
		}),
			Return([]Any{x})).(Parsec)
	})
}

// BegSepBy parses p zero or more times separated, but mandatorily begun, by sep;
// collects the results in a slice.
func BegSepBy(sep, p Parsec) Parsec {
	return Option([]Any{},
		Try(SeqRight(sep, SepBy(sep, p)).(Parsec)))
}

// BegSepBy1 parses p one or more times separated, but mandatorily begun, by sep;
// collects the results in a slice.
func BegSepBy1(sep, p Parsec) Parsec {
	return Option([]Any{},
		Try(SeqRight(sep, SepBy1(sep, p)).(Parsec)))
}

//------------------------------------------------------------------------------

// Skip applies one or more parsers and skips the result. That is, it returns
// a parser state record with a value of nil.
func Skip(ps ...Parsec) Parsec {
	switch len(ps) {
	case 0:
		return Fail(u8.Desur(errNoParser))
	case 1:
		return SeqRight(ps[0], Return(nil)).(Parsec)
	case 2:
		return Bind(ps[0], func(_ Any) Parsec {
			return Skip(ps[1])
		})
	default:
		return SeqRight(ps[0], Skip(ps[1:]...)).(Parsec)
	}
}

// SkipMany parses p zero or more times and skips the results. This is
// like skip but it can apply p zero, one, or many times.
func SkipMany(p Parsec) Parsec {
	return func(s PState) PState {
		st := p(s)
		em := true
		for st.ok && !st.empty {
			em = em && st.empty
			st = p(st)
		}
		if st.empty {
			return st.clone(func(e *PState) {
				e.value = nil
				e.ok = true
				e.empty = em
				e.error = ""
			})
		} else {
			return st.clone(func(e *PState) {})
		}
	}
}

// SkipMany1 parses p one or more times and skips the results.
func SkipMany1(p Parsec) Parsec {
	return SeqRight(p, SkipMany(p)).(Parsec)
}

//------------------------------------------------------------------------------

// LookAhead applies p and returns the result; it consumes no input.
func LookAhead(p Parsec) Parsec {
	return func(s PState) PState {
		st := p(s)
		return s.clone(func(e *PState) {
			e.value = st.value
		})
	}
}

// Predict applies p; if it succeeds it consumes no input.
func Predict(p Parsec) Parsec {
	return func(s PState) PState {
		st := p(s)
		if !(st.ok || st.empty) {
			return st.clone(func(e *PState) {})
		} else {
			return st.clone(func(e *PState) {
				e.input = s.input
				e.pos = s.pos
			})
		}
	}
}

//------------------------------------------------------------------------------

// Search applies a parser p, traversing the input as necessary,
// until it succeeds or it reaches the end of input.
func Search(p Parsec) Parsec {
	var f func(PState) PState
	f = func(s PState) PState {
		st := p(s)
		if st.ok || len(st.input) == 0 {
			return st
		} else {
			s3 := st.clone(func(e *PState) {
				e.input = st.input[1:]
				e.pos = st.pos + 1
				e.error = ""
			})
			return f(s3)
		}
	}
	return f
}

//------------------------------------------------------------------------------

// Print prints the parser state, along with the supplied message,
// to the standard output before executing the enclosed parser.
// It can be used for diagnostics.
func Print(msg u8.Text, p Parsec) Parsec {
	return func(s PState) (so PState) {
		fmt.Printf(u8.Surr(msg)+" %v\n", s)
		so = p(s)
		return
	}
}

//==============================================================================
// State Parsers
//------------------------------------------------------------------------------

// GetInput gets the input stream from a parser state.
func GetInput(s PState) PState {
	return s.clone(func(e *PState) {
		e.value = u8.Desur(s.input)
		e.ok = true
		e.empty = true
		e.error = ""
	})
}

// SetInput sets the input stream in a parser state.
func SetInput(in u8.Text) Parsec {
	return func(s PState) PState {
		return s.clone(func(e *PState) {
			e.input = u8.Surr(in)
			e.ok = true
			e.empty = true
			e.error = ""
		})
	}
}

// GetPosition gets the position in the input stream from a parser state.
func GetPosition(s PState) PState {
	return s.clone(func(e *PState) {
		e.value = s.pos
		e.ok = true
		e.empty = true
		e.error = ""
	})
}

// SetPosition sets the position in the input stream in a parser state.
func SetPosition(p uint64) Parsec {
	return func(s PState) PState {
		return s.clone(func(e *PState) {
			e.pos = p
			e.ok = true
			e.empty = true
			e.error = ""
		})
	}
}

// GetInputAndPosition gets the input stream and current position from a parser state.
func GetInputAndPosition(s PState) PState {
	return s.clone(func(e *PState) {
		e.value = InputAndPosition{inp: u8.Desur(s.input), pos: s.pos}
		e.ok = true
		e.empty = true
		e.error = ""
	})
}

// SetInputAndPosition sets the input stream and current position in a parser state.
func SetInputAndPosition(iap InputAndPosition) Parsec {
	return func(s PState) PState {
		return s.clone(func(e *PState) {
			e.input = u8.Surr(iap.inp)
			e.pos = iap.pos
			e.ok = true
			e.empty = true
			e.error = ""
		})
	}
}

//------------------------------------------------------------------------------

// GetState gets the user state from the parser state record.
func GetState(s PState) (so PState) {
	so = s.clone(func(e *PState) {
		e.value = s.user
		e.ok = true
		e.empty = true
		e.error = ""
	})
	return
}

// PutState puts u as the new value for user state in the parser state record.
func PutState(u Any) Parsec {
	return func(s PState) (so PState) {
		if log["PutState"] || (len(s.log) > 0 && s.log[len(s.log)-1] && loggingEnabled) {
			logFnIn("PutState", &s)
			defer logFnDe("PutState", &so)
		}
		so = s.clone(func(e *PState) {
			e.ok = true
			e.empty = true
			e.user = u
			e.error = ""
		})
		return
	}
}

// ModifyState modifies the user state with the result of f, which takes the old
// user state followed by any additional arguments in more.
func ModifyState(f func(Any, ...Any) Any, more ...Any) Parsec {
	return func(s PState) (so PState) {
		if log["ModifyState"] || (len(s.log) > 0 && s.log[len(s.log)-1] && loggingEnabled) {
			logFnIn("ModifyState", &s)
			defer logFnDe("ModifyState", &so)
		}
		u := f(s.user, more...)
		so = s.clone(func(e *PState) {
			e.ok = true
			e.empty = true
			e.user = u
			e.error = ""
		})
		return
	}
}

// PrintState prints the parser state, along with the supplied message,
// to the standard output.
func PrintState(msg u8.Text) Parsec {
	return func(s PState) (so PState) {
		fmt.Printf(u8.Surr(msg)+" %v\n", s)
		so = s.clone(func(e *PState) {})
		return
	}
}

//------------------------------------------------------------------------------

// PushStateStack assumes the user state in the parser state record
// is a slice, and appends a value.
func PushStateStack(k Any) Parsec {
	return Bind(GetState, func(s Any) Parsec {
		var vec []Any
		if s == nil {
			vec = []Any{}
		} else {
			vec = make([]Any, 0)
			for _, n := range s.([]Any) {
				vec = append(vec, n)
			}
		}
		vec = append(vec, k)
		return PutState(vec)
	})
}

// PushStateStack assumes the user state in the parser state record
// is a slice, and removes and returns the final value.
func PopStateStack() Parsec {
	return Bind(GetState, func(s Any) Parsec {
		var vec []Any
		if s == nil {
			vec = []Any{}
		} else { //OR: vec= s[:len(s)-1] ???
			vec = make([]Any, 0)
			for _, n := range s.([]Any)[:len(s.([]Any))-1] {
				vec = append(vec, n)
			}
		}
		return PutState(vec)
	})
}

// PeekStateStack assumes the user state in the parser state record
// is a slice, and returns the final value.
func PeekStateStack() Parsec {
	return Bind(GetState, func(s Any) Parsec {
		if s == nil {
			s = []Any{}
			return Return(nil)
		} else if len(s.([]Any)) == 0 {
			return Return(nil)
		} else {
			return Return(s.([]Any)[len(s.([]Any))-1])
		}
	})
}

// AlterTopStateStack assumes the user state in the parser state record
// is a slice, and alters the final value.
func AlterTopStateStack(k Any) Parsec {
	return Bind(GetState, func(s Any) Parsec {
		var vec []Any
		tx := u8.Text("AlterTopStateStack doesn't handle zero-sized stacks.")
		if s == nil {
			vec = []Any{}
			return Fail(tx)
		} else if len(s.([]Any)) == 0 {
			return Fail(tx)
		} else {
			vec = make([]Any, 0)
			for _, n := range s.([]Any)[:len(s.([]Any))-1] {
				vec = append(vec, n)
			}
			vec = append(vec, k)
		}
		return SeqRight(PutState(vec), Return(nil)).(Parsec)
	})
}

//------------------------------------------------------------------------------

// GetStateMapEntry assumes the user state in the parser state record
// is a map keyed on strings, and gets an entry for k.
func GetStateMapEntry(k string) Parsec {
	return Bind(GetState, func(s Any) Parsec {
		if s == nil {
			s = map[string]Any{}
		}
		return Return(s.(map[string]Any)[k])
	})
}

// PutStateMapEntry assumes the user state in the parser state record
// is a map keyed on strings, and puts u as the new value for key k.
// It returns the old parser state record.
func PutStateMapEntry(k string, v Any) Parsec {
	return Bind(GetState, func(s Any) Parsec {
		var m map[string]Any
		if s == nil {
			m = map[string]Any{}
		} else {
			m = make(map[string]Any)
			for a, b := range s.(map[string]Any) {
				m[a] = b
			}
		}
		m[k] = v
		return PutState(m)
	})
}

// ModifyStateMapEntry assumes the user state in the parser state record
// is a map keyed on strings, and modifies the entry at key k
// with the result of f, which takes the old entry followed by any
// additional arguments in more. It returns the old parser state record.
func ModifyStateMapEntry(k string, f func(Any, ...Any) Any, more ...Any) Parsec {
	return Bind(GetState, func(s Any) Parsec {
		var m map[string]Any
		v := f(s.(map[string]Any)[k], more...)
		if s == nil {
			m = map[string]Any{}
		} else {
			m = make(map[string]Any)
			for a, b := range s.(map[string]Any) {
				m[a] = b
			}
		}
		m[k] = v
		return PutState(m)
	})
}

//==============================================================================
// Logging Parsers
//------------------------------------------------------------------------------

// LogActivate activates the logging facility for all enveloped parsers.
func LogActivate(p Parsec) Parsec {
	return func(s PState) (so PState) {
		if log["Activate"] || loggingEnabled {
			fmt.Println(strings.Repeat("-", 40))
			logFnIn("LogActivate", &s)
			defer logFnDe("LogActivate", &so)
		}
		s.log = append(s.log, true)
		so = p(s)
		so.log = so.log[:len(so.log)-1]
		if len(so.log) == 0 {
			so.log = nil
		}
		return so
	}
}

// LogSuspends the logging facility, previously activated with LogActivate,
// for all enveloped parsers.
func LogSuspend(p Parsec) Parsec {
	return func(s PState) (so PState) {
		if len(s.log) > 0 && s.log[len(s.log)-1] && loggingEnabled {
			logFnIn("LogSuspend", &s)
			defer logFnDe("LogSuspend", &so)
		}
		s.log = append(s.log, false)
		so = p(s)
		so.log = so.log[:len(so.log)-1]
		if len(so.log) == 0 {
			so.log = nil
		}
		return so
	}
}

// LogResume the logging facility, previously suspended with LogSuspend,
// for all enveloped parsers.
func LogResume(p Parsec) Parsec {
	return func(s PState) (so PState) {
		var wasSuspended = false
		if len(s.log) > 0 && !s.log[len(s.log)-1] {
			wasSuspended = true
			s.log = s.log[:len(s.log)-1]
		}
		so = p(s)
		if wasSuspended {
			so.log = append(so.log, false)
		}
		return so
	}
}

// LogFlagging is not yet used in this version of gro/parsec.
func LogFlagging(fl logFlags, p Parsec) Parsec {
	return func(s PState) (so PState) {
		if log["Activate"] || loggingEnabled {
			logFnIn("LogFlags", &s)
			defer logFnDe("LogFlags", &so)
		}
		var oldFlags logFlags
		if len(s.flags) > 0 {
			oldFlags = s.flags[len(s.flags)-1]
			s.flags = s.flags[:len(s.flags)-1]
		}
		so = p(s)
		so.flags = append(so.flags, oldFlags)
		return so
	}
}

//==============================================================================
// Experimental Code
//------------------------------------------------------------------------------
/*
type Combinator interface {
  Build (ps ...Any) Parser
  Monad (ps ...Any) Parser
}

// Symbol succeeds if the next Codepoint equals the given Codepoint, in which case it
// increments the position of the input stream. It may fail on an unexpected end of input.
type SymbolC struct{}
func (c SymbolC) Build (ps ...Any) Parser {
  return Symbol(ps[0].(u8.Codepoint))
}
func (c SymbolC) Monad (ps ...Any) Parser {
  return Symbol(ps[0].(u8.Codepoint))
}

//TODO ???: Token, Regexp, Fwd, FwdWithParams, Try, AltParse, Bind, Collect, Many, Optional, Option
//ALSO: Many0, SkipMany, Times, LookAhead, Predict, Search (also put in log messages for these)
*/

/*
type Combinator func (ps ...Any) Parser
func myComb (ps ...Any) Parser {
  return Symbol(ps[0].(u8.Codepoint))
}
func (c Combinator) Monad (ps ...Any) Parser {
  return Symbol(ps[0].(u8.Codepoint))
}

func TestExperimental(t *testing.T){
ts.LogAsserts("Experimental", t, func(tt *ts.T){

  tt.AssertEqual(ParseItem(myComb(u8.Codepoint('a')), "abc"),
                 PState{input:"bc", pos:1, value:u8.Codepoint('a'), ok:true})
  tt.AssertEqual(ParseItem(Combinator(myComb).Monad(u8.Codepoint('a')), "abc"),
                 PState{input:"bc", pos:1, value:u8.Codepoint('a'), ok:true})
})}
*/

//==============================================================================
