// Copyright 2017 The Gro Authors. All rights reserved.
// Portions translated from Armando Blancas's Clojure-based Kern parser.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

/*
Package ops implements various functions to be called when triggered by infix operators
in a dynamic language, with particular emphasis on handling UTF-88 correctly.

The 3 structs in package big (i.e. Int, Rat, and Float) are aliased so a String method is accessable when the value rather than the pointer is passed into printf.
The 3 basic types (int64, float64, and complex128) are aliased to have a better name, without a number in it.

Types Codepoint and Text from package utf88 are used extensively, being UTF-88 friendly replacements
for rune and string respectively.
This package introduces types CharClass and Regex, both redefinitions of string.
Regex is used for representing regexes, and CharClass for a character class within a regex.

Function Runex accepts a string as input, and returns either a Codepoint or a CharClass.
The string accepted merges the formats for a rune constant and a regexp character class into a single format,
and Runex analyses the formats, in a UTF-88 friendly manner.

Operator associativity in Gro will be:
  =         //right-assoc
  += -= etc //left-assoc
  []        //postfix path
  ++ --     //postfix path

So for example:
  a= b = c += 3 += 5 = d
becomes:
  a= (b = (((c += 3) += 5) = d))
because = is right-assoc and += is left-assoc.

The syntax for Regex repetition in Gro will be according to this table:
     greedy        reluctant      other forms             private function called
     ------        ---------      -----------             -----------------------
      x * a   ==   a * x          [x:x]  [x]              times
    [:] * a        a * [:]        [0:]  [:1/0]  [0:1/0]   many
   [1:] * a        a * [1:]       [1:1/0]                 many1
   [:1] * a        a * [:1]       [0:1]                   optional
   [x:] * a        a * [x:]       [x:1/0]                 atLeast
  [x:y] * a        a * [x:y]      [:y]  [0:y]             atLeastButNoMoreThan

//TODO: for regexes...
  (?flags:re)      set flags during re; non-capturing; syntax: xy-z (set xy, unset z)
  flag i           case-insensitive (default false)
  //flag s         let . match \n (default false) //DO WE WANT THIS ???
  ^ $ and flag m   at beginning/end of text, or even line when flag m=true (default false)
  \b \B            at/notAt ASCII word boundary (\w on one side and \W, \A, or \z on the other)

  func QuoteMeta(s string) string
  func MustCompilePOSIX(str string) *Regexp
  func (re *Regexp) Longest()
  func (re *Regexp) ReplaceAllString(src, repl string) string
  func (re *Regexp) ReplaceAllLiteralString(src, repl string) string
  func (re *Regexp) ReplaceAllStringFunc(src string, repl func(string) string) string


The parser combinators use the utf88 encoding format for Unicode.
They are useful for implementing recursive-descent parsers based on predictive LL(1) grammars with on-demand, unlimited look-ahead LL(*).

The main inspiration for Parsec comes from the Clojure-based Kern,
located at http://github.com/blancas/kern and written by Armando Blancas in 2013,
and itself inspired by Parsec, a Haskell library written by Daan Leijen, as well as work by Graham Hutton, Erik Meijer, and William Burge.

Daan Leijen, Parsec, a fast combinator parser, 2001
http://legacy.cs.uu.nl/daan/download/parsec/parsec.pdf

Graham Hutton and Erik Meijer
Monadic Parser Combinators, 1996
http://eprints.nottingham.ac.uk/237/1/monparsing.pdf

William H. Burge
Recursive Programming Techniques
Addison-Wesley, 1975

The parser combinators are primarily composed using Codepoint and Text parameters.
The main exception is the Regexp function which accepts a string instead.

Either Text or string can be passed in when the parsers are executed.
The executor functions are ParseText and ParseString.
Each returns two values: if the second ok value is true, the first value is the parsing result, otherwise it's the error message.

The three basic parsers are Satisfy, Fail, and Return.
Three additional basic parsers are provided for optimization reasons: Symbol, Token, and Regexp.

The usual combinator parsers are provided, e.g. Alt, SeqLeft, SeqRight, Apply, Collect, Many, Optional, Try, Ask, Lookahead, SepBy.
Parsers called Fwd and FwdWithParams enable recursion to be handled in the data being parsed.

The separator parsers behave like so, parsing the sequences shown, as well as the larger ones following the same pattern:
	SepBy      <nil>  p   p:p
	SepBy1            p   p:p
	EndBy      <nil>  p:  p:p:
	EndBy1            p:  p:p:
	SepEndBy   <nil>  p   p:  p:p  p:p:
	SepEndBy1         p   p:  p:p  p:p:
	BegSepBy   <nil>   :   :p  :p:p
	BegSepBy1  <nil>       :p  :p:p

The function known as >>= in Clojure's kern is herein called Bind.
The macro known as bind in Clojure and do in Haskell is not represented here, but will be available as an optional macro in the Gro language grammar of which this parsec package will be an integral component.
Such bind/do can be simulated like so:
	digit:=  Regexp(`\p{Nd}`)
	letter:= Regexp(`\pL`)
	lower:=  Regexp(`\p{Ll}`)
	p:= Bind(digit, func(c1 interface{}) Parser {
		return Bind(letter, func(_ interface{}) Parser {
			return Bind(lower, func(c2 interface{}) Parser {
				return Return([]interface{}{c1, c2})
			})
		})
	})

State parsers are provided, as well as those related to the input data being parsed and current parsing position within it.
The logging facility prints an indented record of the parsers called.
*/
package ops
