// Copyright 2009-17 The Go and Gro authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package parsec_test

import (
	"fmt"
	kern "github.com/grolang/gro/parsec"
	"github.com/grolang/gro/utf88"
)

//==============================================================================
// Parser Executors
//------------------------------------------------------------------------------

func ExampleParseText() {
	var r interface{}
	var ok bool
	prt := func() {
		if ok {
			fmt.Printf("Result: %c\n", r)
		} else {
			fmt.Printf("Error: %v\n", r)
		}
	}

	p := kern.Symbol(utf88.Codepoint('a'))

	t := utf88.Text("abc")
	r, ok = kern.ParseText(p, t)
	prt()

	u := utf88.Text("defg")
	r, ok = kern.ParseText(p, u)
	prt()

	// Output:
	// Result: a
	// Error: Unexpected d input.
}

//------------------------------------------------------------------------------

func ExampleParseString() {
	var r interface{}
	var ok bool
	prt := func() {
		if ok {
			fmt.Printf("Result: %c\n", r)
		} else {
			fmt.Printf("Error: %v\n", r)
		}
	}

	p := kern.Symbol(utf88.Codepoint('a'))

	t := "abc"
	r, ok = kern.ParseString(p, t)
	prt()

	u := "defg"
	r, ok = kern.ParseString(p, u)
	prt()

	// Output:
	// Result: a
	// Error: Unexpected d input.
}

//==============================================================================
// Basic Parsers
//------------------------------------------------------------------------------

func ExampleReturn() {
	p := kern.Return(1234567890)

	t := utf88.Text("anything")
	r, ok := kern.ParseText(p, t)
	if ok {
		fmt.Printf("Result: %v\n", r)
	} else {
		fmt.Printf("Error: %v\n", r)
	}

	// Output:
	// Result: 1234567890
}

func ExampleFail() {
	p := kern.Fail(utf88.Text("Some Failure"))

	t := utf88.Text("anything")
	r, ok := kern.ParseText(p, t)
	if ok {
		fmt.Printf("Result: %v\n", r)
	} else {
		fmt.Printf("Error: %v\n", r)
	}

	// Output:
	// Error: Some Failure
}

func ExampleSatisfy() {
	p := kern.Satisfy(func(c utf88.Codepoint) bool {
		return c == 'a'
	})

	t := utf88.Text("abcde")
	r, ok := kern.ParseText(p, t)
	if ok {
		fmt.Printf("Result: %c\n", r)
	} else {
		fmt.Printf("Error: %v\n", r)
	}

	// Output:
	// Result: a
}

//==============================================================================
// Optimized Primitive Parsers
//------------------------------------------------------------------------------

func ExampleSymbol() {
	p := kern.Symbol(utf88.Codepoint('a'))

	t := utf88.Text("abc")
	r, ok := kern.ParseText(p, t)
	if ok {
		fmt.Printf("Result: %c\n", r)
	} else {
		fmt.Printf("Error: %v\n", r)
	}

	// Output:
	// Result: a
}

//------------------------------------------------------------------------------

func ExampleToken() {
	p := kern.Token(utf88.Text("ab"))

	t := utf88.Text("abc")
	r, ok := kern.ParseText(p, t)
	if ok {
		fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
	} else {
		fmt.Printf("Error: %v\n", r)
	}

	// Output:
	// Result: ab
}

//------------------------------------------------------------------------------

func ExampleRegexp() {
	p := kern.Regexp(`z|a|b`)

	t := utf88.Text("abc")
	r, ok := kern.ParseText(p, t)
	if ok {
		fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
	} else {
		fmt.Printf("Error: %v\n", r)
	}

	// Output:
	// Result: a
}

//==============================================================================
// Primitive Parsers
//------------------------------------------------------------------------------

func ExampleOneOf() {
	p := kern.OneOf(utf88.Text("xyzab"))

	t := utf88.Text("abcde")
	r, ok := kern.ParseText(p, t)
	if ok {
		fmt.Printf("Result: %s\n", utf88.Sur(r.(utf88.Codepoint)))
	} else {
		fmt.Printf("Error: %v\n", r)
	}

	// Output:
	// Result: a
}

//------------------------------------------------------------------------------

func ExampleNoneOf() {
	p := kern.NoneOf(utf88.Text("xyz\n\t"))

	t := utf88.Text("abcde")
	r, ok := kern.ParseText(p, t)
	if ok {
		fmt.Printf("Result: %s\n", utf88.Sur(r.(utf88.Codepoint)))
	} else {
		fmt.Printf("Error: %v\n", r)
	}

	// Output:
	// Result: a
}

//==============================================================================
// Combinator Parsers
//------------------------------------------------------------------------------

func ExampleFwd() {
	var expr func() kern.Parser

	paren := func() kern.Parser {
		return kern.SeqRight(kern.Token(utf88.Text(string('('))),
			kern.SeqLeft(kern.Fwd(expr), // Fwd will evaluate passed expression lazily
				kern.Token(utf88.Text(string(')')))))
	}
	expr = func() kern.Parser { // will parse string enclosed in parenthesis pairs to any depth
		return kern.Alt(kern.Token(utf88.Text(string('a'))),
			paren())
	}

	t := utf88.Text("(((a)))")
	r, ok := kern.ParseText(expr(), t)
	if ok {
		fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
	} else {
		fmt.Printf("Error: %v\n", r)
	}

	// Output:
	// Result: a
}

//------------------------------------------------------------------------------

func ExampleFwdWithParams() {
	var expr func(...interface{}) kern.Parser

	paren := func(as ...interface{}) kern.Parser {
		return kern.SeqRight(kern.Token(utf88.Text(string('('))),
			kern.SeqLeft(kern.FwdWithParams(expr, as...), // FwdWithParams will evaluate passed expression lazily
				kern.Token(utf88.Text(string(')')))))
	}
	expr = func(as ...interface{}) kern.Parser { // will parse string enclosed in parenthesis pairs to any depth
		return kern.Alt(kern.Token(utf88.Text(string('a'))),
			paren(as...))
	}

	t := utf88.Text("(((a)))")
	r, ok := kern.ParseText(expr(101, 102), t) // call expr with extra (unused in this e.g.) args
	if ok {
		fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
	} else {
		fmt.Printf("Error: %v\n", r)
	}

	// Output:
	// Result: a
}

//------------------------------------------------------------------------------

func ExampleTry() {
	p := kern.Flatten(kern.Regexp(`\pL`), kern.Regexp(`\pL`))
	q := kern.Flatten(kern.Regexp(`\pL`), kern.Regexp(`\pN`))

	t := utf88.Text("a7bcde")

	var r interface{}
	var ok bool
	prt := func() {
		if ok {
			fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
		} else {
			fmt.Printf("Error: %v\n", r)
		}
	}

	r, ok = kern.ParseText(kern.Alt(p, q), t)
	prt()

	r, ok = kern.ParseText(kern.Alt(kern.Try(p), q), t)
	prt()

	// Output:
	// Error: Unexpected 7 input.
	// Result: a7
}

//------------------------------------------------------------------------------

func ExampleAsk() {
	letter := kern.Regexp(`\pL`)
	digit := kern.Regexp(`\p{Nd}`)

	p := kern.Ask(kern.Collect(digit, letter), utf88.Text("digit,letter"))

	t := utf88.Text(";efg")
	r, ok := kern.ParseText(p, t)
	if ok {
		fmt.Printf("Result: %s\n", utf88.Sur(r.(utf88.Codepoint)))
	} else {
		fmt.Printf("Error: %v\n", r)
	}

	// Output:
	// Error: Unexpected ; input. Expecting digit,letter
}

//------------------------------------------------------------------------------

func ExampleAlt() {
	var r interface{}
	var ok bool
	prt := func() {
		if ok {
			fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
		} else {
			fmt.Printf("Error: %v\n", r)
		}
	}

	lower := kern.Regexp(`\p{Ll}`)
	upper := kern.Regexp(`\p{Lu}`)
	digit := kern.Regexp(`\p{Nd}`)

	p := kern.Alt(lower, digit)
	q := kern.Alt(lower, digit, upper)

	t := utf88.Text("7ef")
	u := utf88.Text(";ef")

	r, ok = kern.ParseText(p, t)
	prt()

	r, ok = kern.ParseText(p, u)
	prt()

	r, ok = kern.ParseText(q, t)
	prt()

	// Output:
	// Result: 7
	// Error: No alternatives selected.
	// Result: 7
}

//------------------------------------------------------------------------------

func ExampleBind() {
	var r interface{}
	var ok bool
	prt := func() {
		if ok {
			fmt.Printf("Result: %s\n", r)
		} else {
			fmt.Printf("Error: %v\n", r)
		}
	}

	lower := kern.Regexp(`\p{Ll}`)
	intNum := kern.Regexp(`\p{Nd}*`)

	p := kern.Bind(lower, func(c interface{}) kern.Parser {
		return kern.Bind(intNum, func(d interface{}) kern.Parser {
			return kern.Bind(lower, func(e interface{}) kern.Parser {
				return kern.Return(utf88.Surr(c.(utf88.Text)) + "," +
					utf88.Surr(d.(utf88.Text)) + "," +
					utf88.Surr(e.(utf88.Text)))
			})
		})
	})

	t := utf88.Text("e789fg")
	r, ok = kern.ParseText(p, t)
	prt()

	// Output:
	// Result: e,789,f
}

//------------------------------------------------------------------------------

func ExampleSeqLeft() {
	digit := kern.Regexp(`\p{Nd}`)
	letter := kern.Regexp(`\pL`)

	p := kern.SeqLeft(digit, letter)

	t := utf88.Text("7efg")
	r, ok := kern.ParseText(p, t)
	if ok {
		fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
	} else {
		fmt.Printf("Error: %v\n", r)
	}

	// Output:
	// Result: 7
}

//------------------------------------------------------------------------------

func ExampleSeqRight() {
	digit := kern.Regexp(`\p{Nd}`)
	letter := kern.Regexp(`\pL`)

	p := kern.SeqRight(digit, letter)

	t := utf88.Text("7efg")
	r, ok := kern.ParseText(p, t)
	if ok {
		fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
	} else {
		fmt.Printf("Error: %v\n", r)
	}

	// Output:
	// Result: e
}

//------------------------------------------------------------------------------

func ExampleApply() {
	letter := kern.Regexp(`\pL`)

	p := kern.Apply(func(c interface{}) interface{} {
		cs := utf88.Surr(c.(utf88.Text))
		return cs + cs + cs
	}, letter)

	t := utf88.Text("abc")
	r, ok := kern.ParseText(p, t)
	if ok {
		fmt.Printf("Result: %s\n", r)
	} else {
		fmt.Printf("Error: %v\n", r)
	}

	// Output:
	// Result: aaa
}

//------------------------------------------------------------------------------

func ExampleCollect() {
	letter := kern.Regexp(`\pL`)
	digit := kern.Regexp(`\p{Nd}`)

	p := kern.Collect(letter, digit, letter)

	t := utf88.Text("a5bc")
	r, ok := kern.ParseText(p, t)
	if ok {
		fmt.Printf("Result: %c\n", r)
	} else {
		fmt.Printf("Error: %v\n", r)
	}

	// Output:
	// Result: [[a] [5] [b]]
}

//------------------------------------------------------------------------------

func ExampleMany() {
	var r interface{}
	var ok bool
	prt := func() {
		if ok {
			fmt.Printf("Result: %c\n", r)
		} else {
			fmt.Printf("Error: %v\n", r)
		}
	}

	letter := kern.Regexp(`\pL`)
	p := kern.Many(letter)

	t := utf88.Text("abc78d")
	u := utf88.Text("789def")

	r, ok = kern.ParseText(p, t)
	prt()
	r, ok = kern.ParseText(p, u)
	prt()

	// Output:
	// Result: [[a] [b] [c]]
	// Result: []
}

//------------------------------------------------------------------------------

func ExampleOptional() {
	var r interface{}
	var ok bool
	prt := func() {
		if ok && r == nil {
			fmt.Println("Nil result")
		} else if ok {
			fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
		} else {
			fmt.Printf("Error: %v\n", r)
		}
	}

	letter := kern.Regexp(`\pL`)
	p := kern.Optional(letter)

	t := utf88.Text("abc")
	u := utf88.Text("789")

	r, ok = kern.ParseText(p, t)
	prt()
	r, ok = kern.ParseText(p, u)
	prt()

	// Output:
	// Result: a
	// Nil result
}

//------------------------------------------------------------------------------

func ExampleOption() {
	var r interface{}
	var ok bool
	prt := func() {
		if ok {
			fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
		} else {
			fmt.Printf("Error: %v\n", r)
		}
	}

	letter := kern.Regexp(`\pL`)
	p := kern.Option(utf88.Text("None."), letter)

	t := utf88.Text("abc")
	u := utf88.Text("789")

	r, ok = kern.ParseText(p, t)
	prt()
	r, ok = kern.ParseText(p, u)
	prt()

	// Output:
	// Result: a
	// Result: None.
}

//------------------------------------------------------------------------------

func ExampleSepBy() {
	var r interface{}
	var ok bool
	prt := func() {
		if ok {
			fmt.Printf("Result: %c\n", r)
		} else {
			fmt.Printf("Error: %v\n", r)
		}
	}

	letter := kern.Regexp(`\pL`)
	punc := kern.Regexp(`\pP`)
	p := kern.SepBy(punc, letter)

	t := utf88.Text("a;b:c789")
	u := utf88.Text(";789")

	r, ok = kern.ParseText(p, t)
	prt()
	r, ok = kern.ParseText(p, u)
	prt()

	// Output:
	// Result: [[a] [b] [c]]
	// Result: []
}

//==============================================================================
