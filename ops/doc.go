// Copyright 2017 The Gro Authors. All rights reserved.
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
*/
package ops
