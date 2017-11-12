// Copyright 2009-17 The Go and Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package utf88_test

import (
	"fmt"
	u8 "github.com/grolang/gro/utf88"
	"reflect"
)

func ExampleSur() {
	fmt.Println(
		//successfully surrogated rune...
		"xyz"+u8.Sur(0x7fbfffff)+"xyz" ==
			"xyz\U000FFFBF\U0010FFFFxyz" &&

			//rune not needing surrogation...
			"xyz"+u8.Sur('v')+"xyz" ==
				"xyzvxyz" &&

			//multiple runes...
			"xyz"+u8.Sur(0x7fbfffff, 'v')+"xyz" ==
				"xyz\U000FFFBF\U0010FFFF\u0076xyz" &&

			//leading surrogate illegal for surrogation...
			"xyz"+u8.Sur(0xf8011)+"xyz" ==
				"xyz\uFFFDxyz" &&

			//runes above maximum illegal for surrogation...
			"xyz"+u8.Sur(0x7fe00000)+"xyz" ==
				"xyz\uFFFDxyz")

	// Output:
	// true
}

func ExampleParsePoints() {
	fmt.Println(reflect.DeepEqual(
		u8.ParsePoints("ud800U7fdfffff"),
		u8.Text{0xd800, 0x7fdfffff}))

	// Output:
	// true
}

func ExampleDecode() {
	fmt.Println(
		u8.Decode("123\xff456") ==
			"123\ufffd456")

	// Output:
	// true
}

func ExampleDesur() {
	fmt.Println(
		reflect.DeepEqual(
			u8.Desur("xyz\U000FFFBF\U0010FFFF\u0076xyz"),
			u8.Text{'x', 'y', 'z', 0x7fbfffff, 'v', 'x', 'y', 'z'}) &&

			reflect.DeepEqual(
				u8.Desur("\uffff"+u8.Sur(0x110000)+"abc"+u8.Sur(0x7fbfffff)+"z"),
				u8.Text{0xffff, 0x110000, 'a', 'b', 'c', 0x7fbfffff, 'z'}))

	// Output:
	// true
}

func ExampleDesurrogateLastPoint() {
	r1, s1 := u8.DesurrogateLastPoint([]rune("abc"))
	fmt.Println(r1 == 'c' && s1 == 1)

	//leading utf88-surrogate gives error rune...
	r2, s2 := u8.DesurrogateLastPoint([]rune("a\U000F8011"))
	fmt.Println(r2 == 0xfffd && s2 == 1)

	//leading utf88-surrogate followed by trailing one...
	r3, s3 := u8.DesurrogateLastPoint([]rune("xyz\U000f8011\U00100000"))
	fmt.Println(r3 == 0x110000 && s3 == 2)

	// Output:
	// true
	// true
	// true
}

func ExamplePointCountOfRunes() {
	fmt.Println(
		u8.PointCountOfRunes([]rune("xyz\U000FFFBF\U0010FFFF\u0076xyz")) == 8 &&
			u8.PointCountOfRunes([]rune("\uffff"+u8.Sur(0x110000)+"abc"+u8.Sur(0x7fbfffff)+"z")) == 7)

	// Output:
	// true
}

func ExamplePointCountOfBytes() {
	fmt.Println(
		//string of valid runes
		u8.PointCountOfBytes([]byte("\u007f\u0080\u07ff")) == 3 &&

			//another string of valid runes
			u8.PointCountOfBytes([]byte(u8.Sur(0x110000)+u8.Sur(0x7fdfffff)+"a")) == 3 &&

			//leading utf88-surrogate with no trailing one gives U+FFFD
			u8.PointCountOfBytes([]byte("\U000f8011")) == 1 &&

			//valid surrogation of leading then trailing utf88-surrogate
			u8.PointCountOfBytes([]byte("\U000f8011\U00100000")) == 1 &&

			//rune too high but produced valid replacement rune U+FFFD
			u8.PointCountOfBytes([]byte(u8.Sur(0x7fffffff))) == 1 &&

			//leading utf88-surrogate with illegal rune following
			u8.PointCountOfBytes([]byte("\U000f8011"+"a")) == 2)

	// Output:
	// true
}

func ExampleValidBytes() {
	fmt.Println(
		//string of valid runes
		u8.ValidBytes([]byte("\u007f\u0080\u07ff")) &&

			//bytes for utf16-surrogate
			!u8.ValidBytes([]byte("\xed\x9f\xc0")) &&

			//leading utf88-surrogate with no trailing one
			!u8.ValidBytes([]byte("\U000f8011")) &&

			//valid surrogation of leading then trailing utf88-surrogate
			u8.ValidBytes([]byte("\U000f8011\U00100000")) &&

			//rune too high but produced valid replacement rune U+FFFD
			u8.ValidBytes([]byte(u8.Sur(0x7fffffff))) &&

			//leading utf88-surrogate with illegal rune following
			!u8.ValidBytes([]byte("\U000f8011"+"a")))

	// Output:
	// true
}

func ExampleValidRunes() {
	fmt.Println(
		u8.ValidRunes([]rune{0xd7ff}) && //valid surrogated rune
			!u8.ValidRunes([]rune{0xd7ff, 0x7fdfffff})) //valid followed by invalid surrogated rune

	// Output:
	// true
}

func ExampleLenInRunes() {
	fmt.Println(
		u8.LenInRunes(0x10ffff) == 1 && //max utf-8 rune
			u8.LenInRunes(0x110000) == 2)

	// Output:
	// true
}

func ExampleLenInBytes() {
	fmt.Println(
		u8.LenInBytes('a') == 1 &&
			u8.LenInBytes('÷') == 2 &&
			u8.LenInBytes('世') == 3 &&
			u8.LenInBytes(0xd800) == -1 && //utf16-surrogate half
			u8.LenInBytes(0xF8011) == -1 && //leading utf88-surrogate
			u8.LenInBytes(0x10FFFF) == 4 &&
			u8.LenInBytes(0x110000) == 8 &&
			u8.LenInBytes(0x7fe00000) == -1) //rune too high

	// Output:
	// true
}

func ExampleValidForSurrogation() {
	fmt.Println(
		u8.ValidForSurrogation('a') &&
			u8.ValidForSurrogation('÷') &&
			u8.ValidForSurrogation('世') &&
			u8.ValidForSurrogation(0xd800) && //utf16-surrogate half valid...
			!u8.ValidForSurrogation(0xF8011) && //...but leading utf88-surrogate not valid
			u8.ValidForSurrogation(0x10FFFF) &&
			u8.ValidForSurrogation(0x110000) &&
			!u8.ValidForSurrogation(0x7fe00000)) //rune too high is not valid

	// Output:
	// true
}

func ExampleValidForEncoding() {
	fmt.Println(
		u8.ValidForEncoding('a') &&
			u8.ValidForEncoding('÷') &&
			u8.ValidForEncoding('世') &&
			!u8.ValidForEncoding(0xd800) && //utf16-surrogate half not valid
			!u8.ValidForEncoding(0xF8011) && //leading utf88-surrogate not valid
			u8.ValidForEncoding(0x10FFFF) &&
			u8.ValidForEncoding(0x110000) &&
			!u8.ValidForEncoding(0x7fe00000)) //rune too high is not valid

	// Output:
	// true
}
