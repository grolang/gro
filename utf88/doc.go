// Copyright 2009-17 The Go and Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

/*
Package utf88 implements UTF-88, which extends Unicode UTF-8 to the 2.1 billion codepoints as originally specified by Pike and Thompson
by using a UTF-16 style surrogate system to encode the points U+110000 and above.

This package implements functions to support text encoded in UTF-88.

UTF-88 encodes the codepoints U+110000 and above by first mapping them to pairs of surrogates
defined in the Private Use planes U+Fxxxx and U+10xxxx.
Note: Reference to "surrogates" in this documentation specifically refers to these;
when the equivalent artifacts for UTF-16 are referenced, they are called "utf16-surrogates".

The leading surrogates are from U+F8011 to U+FFFBF, being presently categorized as Private Use.
The trailing surrogates are from U+100000 to U+10FFFF, presently categorized as
Private Use (U+100000 to U+10FFFD) and Non-Character (U+10FFFE and U+10FFFF).

The formula for surrogation is:
  (leading - 0x8000) * 0x10000 + (trailing - 0x100000)

This gives codepoints from U+110000 to U+7FBFFFFF.

Note: A surrogated UTF-88 codepoint is called a "rune" and aliased by Go's rune type,
whereas we use "codepoint" specifically to refer to an unsurrogated UTF-88 codepoint which has its own alias.

The leading surrogates will become illegal codepoints,
however the trailing surrogates will have the dual function of representing the codepoints in plane U+10xxxx.

Runes will have the restriction that every leading surrogate must be followed by a trailing one,
but a trailing surrogate can exist in isolation not preceded by a leading one.

In unsurrogated code, leading surrogates (U+F8011 to U+FFFBF) and codepoints above U+7FDFFFFF
are substituted with replacement rune U+FFFD when surrogating.

Use

The main interface to the utf88 package are the Sur, Surr, and Desur functions.

Whereas casting a rune between U+110000 and U+7FBFFFFF to a string will result in the Error Rune U+FFFD,
using function Sur on it will surrogate it into two constituent runes that make up a legal Go string:

  //error rune...
  "xyz" + string(0x7fbfffff) + "xyz" == "xyz\uFFFDxyz"

  //will not even compile...
  "xyz\U7FBFFFFFxyz"

  //successfully surrogated rune...
  "xyz" + utf88.Sur(0x7fbfffff) + "xyz"    == "xyz\U000FFFBF\U0010FFFFxyz"

If the rune doesn't need surrogating, Sur will return the rune unchanged:

  //rune not needing surrogation...
  "xyz" + utf88.Sur('v') + "xyz"    == "xyzvxyz"

A leading surrogate (between U+F8011 and U+FFFBF) will return the Error Rune U+FFFD when surrogated,
as will runes below U+0 and those U+7FC00000 and above:

  //leading surrogate illegal for surrogation...
  "xyz" + utf88.Sur(0xf8011) + "xyz"    == "xyz\uFFFDxyz"

  //runes above maximum illegal for surrogation...
  "xyz" + utf88.Sur(0x7fc00000) + "xyz"    == "xyz\uFFFDxyz"

Sur can take any number of runes as arguments.

  //multiple runes...
  "xyz" + utf88.Sur(0x7fbfffff, 'v') + "xyz"    == "xyz\U000FFFBF\U0010FFFF\u0076xyz"

Desur will take a string of surrogated runes and return an array of unsurrogated codepoints:

  //unsurrogating a string...
  reflect.DeepEqual(
    utf88.Desur("xyz\U000FFFBF\U0010FFFF\u0076xyz"),
    []utf88.Codepoint{'x', 'y', 'z', 0x7fbfffff, 'v', 'x', 'y', 'z'}
  )

Other functions are modelled after those in the unicode/utf8 package:
  SurrogatePoint is modelled after utf8.EncodeRune
  DesurrogateFirstPoint is modelled after utf8.DecodeRune
  DesurrogateLastPoint is modelled after utf8.DecodeLastRune
  PointCountOfRunes and PointCountOfBytes are modelled after utf8.RuneCount
  LenInRunes and LenInBytes are modelled after utf8.RuneLen
  ValidRunes and ValidBytes are modelled after utf8.Valid
  ValidForSurrogation and ValidForEncoding are modelled after utf8.ValidRune

Rationale

UTF-88 has these purposes:
  to immediately extend the present 137000 Private Use codepoint limit to over 1 million
  to experiment with management of the 2.1 billion codepoints specified in pre-2003 UTF-8
  to introduce a surrogation scheme that UTF-16 could use someday to encode all points in pre-2003 UTF-8 (and UTF-32)

The general categories of the repurposed and new codepoints are:
  U+F8000 to U+F8010        changed from Private Use to Reserved
  U+F8011 to U+FFFBF        changed from Private Use to Surrogate (used as leading surrogates)
  U+FFFC0 to U+FFFFD        changed from Private Use to Reserved
  U+FFFFE and U+FFFFF       still Non-Character
  U+100000 to U+10FFFD      still Private Use (but has dual use as trailing surrogates)
  U+10FFFE and U+10FFFF     changed from Non-Character to Private Use
  U+110000 to U+1FFFFF      new codepoints with category Private Use
  U+200000 to U+7FBFFFFF    new codepoints with category Reserved
  U+7FC00000 to U+7FFFFFFF  new codepoints with category Surrogate (but use is illegal)

UTF-88 enables users to use Private Use plane U+10xxxx before they make a decision on whether to use UTF-8 or UTF-88 for encoding.
By removing half a Private Use plane, UTF-88 adds back another 15 Private Use planes and 32672 Reserved planes.
The topmost 64 planes are categorized as Surrogate to allow a further surrogation scheme to increase the codepoint count to 4.4 trillion
after the pre-2003 version of UTF-8 is reinstated.

UTF-88 introduces new terminology, that of a "volume" defined as 16 planes of codepoints from (modulo 0x100000 addresses) U+0 to U+FFFFF.
Hence volume V+0 is codepoints U+0 to U+FFFFF, volume V+1 is U+100000 to U+1FFFFF, and volume V+7FF is U+7FF00000 to U+7FFFFFFF.

The management and pre-2003 UTF-8 encoding length of these 2048 volumes is:
  V+0              volume managed by Unicode Consortium, various lengths from 1 to 4 bytes
  V+1              Private Use volume, length 4 bytes
  V+2 to V+3F      Reserved 62 volumes, length 5 bytes
  V+40 to V+7FB    Reserved 1980 volumes, length 6 bytes
  V+7FC to V+7FF   Surrogate 4 volumes, illegal use, length 6 bytes

Although the UTF-88 encoding length of the upper 15 planes in V+1 is 8 bytes,
it will decrease back down to 4 bytes someday when the pre-2003 version of UTF-8 is reinstated.
*/
package utf88
