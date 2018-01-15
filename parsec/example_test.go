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
	var err error
	prt := func() {
		if err == nil {
			fmt.Printf("Result: %c\n", r)
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}

	p := kern.Symbol(utf88.Codepoint('a'))

	t := utf88.Text("abc")
	r, err = kern.ParseItem(p, t)
	prt()

	u := utf88.Text("defg")
	r, err = kern.ParseItem(p, u)
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
	r, err := kern.ParseItem(p, t)
	if err == nil {
		fmt.Printf("Result: %v\n", r)
	} else {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// Result: 1234567890
}

func ExampleFail() {
	p := kern.Fail(utf88.Text("Some Failure"))

	t := utf88.Text("anything")
	r, err := kern.ParseItem(p, t)
	if err == nil {
		fmt.Printf("Result: %v\n", r)
	} else {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// Error: Some Failure
}

func ExampleSatisfy() {
	p := kern.Satisfy(func(c utf88.Codepoint) bool {
		return c == 'a'
	})

	t := utf88.Text("abcde")
	r, err := kern.ParseItem(p, t)
	if err == nil {
		fmt.Printf("Result: %c\n", r)
	} else {
		fmt.Printf("Error: %v\n", err)
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
	r, err := kern.ParseItem(p, t)
	if err == nil {
		fmt.Printf("Result: %c\n", r)
	} else {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// Result: a
}

//------------------------------------------------------------------------------

func ExampleToken() {
	p := kern.Token(utf88.Text("ab"))

	t := utf88.Text("abc")
	r, err := kern.ParseItem(p, t)
	if err == nil {
		fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
	} else {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// Result: ab
}

//------------------------------------------------------------------------------

func ExampleRegexp() {
	p := kern.Regexp(`z|a|b`)

	t := utf88.Text("abc")
	r, err := kern.ParseItem(p, t)
	if err == nil {
		fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
	} else {
		fmt.Printf("Error: %v\n", err)
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
	r, err := kern.ParseItem(p, t)
	if err == nil {
		fmt.Printf("Result: %s\n", utf88.Sur(r.(utf88.Codepoint)))
	} else {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// Result: a
}

//------------------------------------------------------------------------------

func ExampleNoneOf() {
	p := kern.NoneOf(utf88.Text("xyz\n\t"))

	t := utf88.Text("abcde")
	r, err := kern.ParseItem(p, t)
	if err == nil {
		fmt.Printf("Result: %s\n", utf88.Sur(r.(utf88.Codepoint)))
	} else {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// Result: a
}

//==============================================================================
// Combinator Parsers
//------------------------------------------------------------------------------

func ExampleFwd() {
	var expr func(...interface{}) interface{}

	paren := func(...interface{}) interface{} {
		return kern.SeqRight(kern.Token(utf88.Text(string('('))),
			kern.SeqLeft(kern.Fwd(expr), // Fwd will evaluate passed expression lazily
				kern.Token(utf88.Text(string(')')))))
	}
	expr = func(...interface{}) interface{} { // will parse string enclosed in parenthesis pairs to any depth
		return kern.Alt(kern.Token(utf88.Text(string('a'))), paren().(kern.Parser))
	}

	t := utf88.Text("(((a)))")
	r, err := kern.ParseItem(expr().(kern.Parser), t)
	if err == nil {
		fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
	} else {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// Result: a
}

//------------------------------------------------------------------------------
func ExampleFwdWithParams() {
	var expr func(...interface{}) interface{}

	paren := func(as ...interface{}) interface{} {
		return kern.SeqRight(kern.Token(utf88.Text(string('('))),
			kern.SeqLeft(kern.Fwd(append([]interface{}{expr}, as...)...), // Fwd will evaluate passed expression lazily
				kern.Token(utf88.Text(string(')')))))
	}
	expr = func(as ...interface{}) interface{} { // will parse string enclosed in parenthesis pairs to any depth
		return kern.Alt(kern.Token(utf88.Text(string('a'))), paren(as...).(kern.Parser))
	}

	t := utf88.Text("(((a)))")
	r, err := kern.ParseItem(expr(101, 102).(kern.Parser), t) // call expr with extra (unused in this e.g.) args
	if err == nil {
		fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
	} else {
		fmt.Printf("Error: %v\n", err)
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
	var err error
	prt := func() {
		if err == nil {
			fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}

	r, err = kern.ParseItem(kern.Alt(p, q).(kern.Parser), t)
	prt()

	r, err = kern.ParseItem(kern.Alt(kern.Try(p), q).(kern.Parser), t)
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
	r, err := kern.ParseItem(p, t)
	if err == nil {
		fmt.Printf("Result: %s\n", utf88.Sur(r.(utf88.Codepoint)))
	} else {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// Error: Unexpected ; input. Expecting digit,letter
}

//------------------------------------------------------------------------------

func ExampleAlt() {
	var r interface{}
	var err error
	prt := func() {
		if err == nil {
			fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}

	lower := kern.Regexp(`\p{Ll}`)
	upper := kern.Regexp(`\p{Lu}`)
	digit := kern.Regexp(`\p{Nd}`)

	p := kern.Alt(lower, digit).(kern.Parser)
	q := kern.Alt(lower, digit, upper).(kern.Parser)

	t := utf88.Text("7ef")
	u := utf88.Text(";ef")

	r, err = kern.ParseItem(p, t)
	prt()

	r, err = kern.ParseItem(p, u)
	prt()

	r, err = kern.ParseItem(q, t)
	prt()

	// Output:
	// Result: 7
	// Error: No alternatives selected.
	// Result: 7
}

//------------------------------------------------------------------------------

func ExampleBind() {
	var r interface{}
	var err error
	prt := func() {
		if err == nil {
			fmt.Printf("Result: %s\n", r)
		} else {
			fmt.Printf("Error: %v\n", err)
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
	r, err = kern.ParseItem(p, t)
	prt()

	// Output:
	// Result: e,789,f
}

//------------------------------------------------------------------------------

func ExampleSeqLeft() {
	digit := kern.Regexp(`\p{Nd}`)
	letter := kern.Regexp(`\pL`)

	p := kern.SeqLeft(digit, letter).(kern.Parser)

	t := utf88.Text("7efg")
	r, err := kern.ParseItem(p, t)
	if err == nil {
		fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
	} else {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// Result: 7
}

//------------------------------------------------------------------------------

func ExampleSeqRight() {
	digit := kern.Regexp(`\p{Nd}`)
	letter := kern.Regexp(`\pL`)

	p := kern.SeqRight(digit, letter).(kern.Parser)

	t := utf88.Text("7efg")
	r, err := kern.ParseItem(p, t)
	if err == nil {
		fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
	} else {
		fmt.Printf("Error: %v\n", err)
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
	r, err := kern.ParseItem(p, t)
	if err == nil {
		fmt.Printf("Result: %s\n", r)
	} else {
		fmt.Printf("Error: %v\n", err)
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
	r, err := kern.ParseItem(p, t)
	if err == nil {
		fmt.Printf("Result: %c\n", r)
	} else {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// Result: [[a] [5] [b]]
}

//------------------------------------------------------------------------------

func ExampleMany() {
	var r interface{}
	var err error
	prt := func() {
		if err == nil {
			fmt.Printf("Result: %c\n", r)
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}

	letter := kern.Regexp(`\pL`)
	p := kern.Many(letter)

	t := utf88.Text("abc78d")
	u := utf88.Text("789def")

	r, err = kern.ParseItem(p, t)
	prt()
	r, err = kern.ParseItem(p, u)
	prt()

	// Output:
	// Result: [[a] [b] [c]]
	// Result: []
}

//------------------------------------------------------------------------------

func ExampleOptional() {
	var r interface{}
	var err error
	prt := func() {
		if err == nil && r == nil {
			fmt.Println("Nil result")
		} else if err == nil {
			fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}

	letter := kern.Regexp(`\pL`)
	p := kern.Optional(letter)

	t := utf88.Text("abc")
	u := utf88.Text("789")

	r, err = kern.ParseItem(p, t)
	prt()
	r, err = kern.ParseItem(p, u)
	prt()

	// Output:
	// Result: a
	// Nil result
}

//------------------------------------------------------------------------------

func ExampleOption() {
	var r interface{}
	var err error
	prt := func() {
		if err == nil {
			fmt.Printf("Result: %s\n", utf88.Surr(r.(utf88.Text)))
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}

	letter := kern.Regexp(`\pL`)
	p := kern.Option(utf88.Text("None."), letter)

	t := utf88.Text("abc")
	u := utf88.Text("789")

	r, err = kern.ParseItem(p, t)
	prt()
	r, err = kern.ParseItem(p, u)
	prt()

	// Output:
	// Result: a
	// Result: None.
}

//------------------------------------------------------------------------------

func ExampleSepBy() {
	var r interface{}
	var err error
	prt := func() {
		if err == nil {
			fmt.Printf("Result: %c\n", r)
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}

	letter := kern.Regexp(`\pL`)
	punc := kern.Regexp(`\pP`)
	p := kern.SepBy(punc, letter)

	t := utf88.Text("a;b:c789")
	u := utf88.Text(";789")

	r, err = kern.ParseItem(p, t)
	prt()
	r, err = kern.ParseItem(p, u)
	prt()

	// Output:
	// Result: [[a] [b] [c]]
	// Result: []
}

//==============================================================================
