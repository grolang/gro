// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

/*
Package ops implements various functions to be called when triggered by infix operators
in a dynamic language, with particular emphasis on handling UTF-88 correctly.

Types Codepoint and Text from package utf88 are used extensively, being UTF-88 friendly replacements
for rune and string respectively.
This package introduces types CharClass and Regex, both redefinitions of string.
Regex is used for representing regexes, and CharClass for a character class within a regex.

Function Runex accepts a string as input, and returns either a Codepoint or a CharClass.
The string accepted merges the formats for a rune constant and a regexp character class into a single format,
and Runex analyses the formats, in a UTF-88 friendly manner.

This package also uses 9 numeric types, arranged in a hierarchy:
  nil
  bool
  Int (which is int64, a "quintillion-bounded" integer)
  BigInt (which is math/big.Int, an unbounded integer)
  BigRat (which is math/big.Rat, an unbounded rational)
  Float (which is float64)
  BigFloat (which is math/big.Float, a float which can have greater precision than float64)
  Complex (which is complex128)
  BigComplex (which is composed of two math/big.Float)

The only valid infinity and not-a-number values are:
  Inf (Riemann infinity, which is positive real and imaginary infinities in complex128)
  NaN (not-a-number in complex128)

All other possible infinity and not-a-number representations will be converted to one of those two.

When two numbers of different types in the numeric hierarchy are passed as args into:
  Plus(x, y interface{})interface{}
  Minus(x, y interface{})interface{}
  Mult(x, y interface{})interface{}
  Divide(x, y interface{})interface{}
  Power(x, y interface{})interface{}
the result will generally be of the same type as whichever argument is higher in the numeric hierarchy,
e.g. for a BigInt plus a Float, the result is a Float.

The exceptions to this rule are:
  two Ints added, subtracted, or multiplied together is promoted to BigInt if it overflows
  division by nil or false generates Inf (+ or -)
  division by a Int or BigInt promotes to a BigRat
  raising to the power of a nil or false generates a panic //TODO: change to Complex.NaN ???
  a Int to the power of a Int is a BigInt
  a Int or BigInt to the power of a BigRat is a Float
  a BigRat to the power of a Int, BigInt, or BigRat is a Float

The 3 structs in package big (i.e. Int, Rat, and Float) are aliased so a String method is accessable when the value rather than the pointer is passed into printf.
The 3 basic types (int64, float64, and complex128) are aliased to have a better name, without a number in it.

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
*/
package ops
