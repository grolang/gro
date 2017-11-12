// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

/*
Package assert adds asserts to the go testing package.

This package must be used in conjunction with the go testing package.
It provides an automatically incremented counter and some assertion functions.
It is called like so:

  import(
    "testing"
    ts "github.com/grolang/gro/assert"
  )

  func TestAssert(t *testing.T){
    ts.LogAsserts("Assert", t, func(tt *ts.T){

      //calls to various tester functions here, such as:
      //some that update the counter...
      tt.Assert(true)
      tt.AssertEqual(1, 1)
      tt.AssertPanic(func(){
        if true { panic("X") }
      }, "X")
      tt.AssertAnyPanic(func(){
        if true { panic("X") }
      })

      //plus some that don't update the counter...
      tt.AssertEqualN('a', 'a', 123)
      tt.PrintValue("abcdefg")

    })
  }

More assert functions will be added to the package in later versions if needed.
*/
package assert

// The example in the package doc above can't be demonstrated by an Example test file.

import (
	"fmt"
	"reflect"
	"testing"
)

//================================================================================

// T is a struct extending the one from standard library package testing with assertion capabilities.
type T struct {
	*testing.T
	AssertPrintSwitch bool
	AssertCounter     int
}

//--------------------------------------------------------------------------------

func AssertTrue(a interface{}) {
	if !reflect.DeepEqual(a, true) {
		panic(fmt.Sprintf("assert.AssertTrue: assertion of %x is false", a))
	}
}

//--------------------------------------------------------------------------------

func isEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

//--------------------------------------------------------------------------------

// LogAsserts is the wrapper function within which calls to the various tester utility functions must be embedded.
// LogAsserts itself must be embedded within a function that will be called by the go testing package.
func LogAsserts(s string, t *testing.T, runner func(*T)) {
	tt := T{t, false, 0}
	runner(&tt)
	if tt.AssertPrintSwitch {
		tt.Fail()
	}
	fmt.Println(tt.AssertCounter, "asserts run for", s)
}

//--------------------------------------------------------------------------------

// Assert will increment the counter, and throw a testing error if the assertion is false.
func (tt *T) Assert(b bool) {
	tt.AssertCounter += 1
	if !b {
		tt.Errorf("assert %d failed.\n"+
			"....found:%v (%[2]T)\n", tt.AssertCounter, b)
	}
}

// AssertEqual will increment the counter, and throw a testing error if the deep equality is false.
func (tt *T) AssertEqual(a, b interface{}) {
	tt.AssertCounter += 1
	if !isEqual(a, b) {
		tt.Errorf("assert %d failed.\n"+
			"......found:%v (%[2]T)\n"+
			"...expected:%v (%[3]T)\n", tt.AssertCounter, a, b)
	}
}

// AssertPanic will increment the counter, and throw a testing error if the supplied function doesn't throw the specified panic.
func (tt *T) AssertPanic(g func(), msg string) {
	tt.AssertCounter += 1
	defer func() {
		if x := recover(); fmt.Sprintf("%v", x) != msg {
			tt.Errorf("assert %d failed.\n"+
				"....found recover:%v\n"+
				"...expected panic:%v\n", tt.AssertCounter, fmt.Sprintf("%v", x), msg)
		}
	}()
	g()
}

// AssertAnyPanic will increment the counter, and throw a testing error if the supplied function doesn't throw a panic.
func (tt *T) AssertAnyPanic(g func()) {
	tt.AssertCounter += 1
	defer func() {
		if x := recover(); x == nil {
			tt.Errorf("assert %d failed. %v\n"+
				"....found nil recover\n", tt.AssertCounter, x)
		}
	}()
	g()
}

// AssertEqualN will throw a testing error if the deep equality is false, using the supplied integer as the error number.
func (tt *T) AssertEqualN(a, b interface{}, n int) {
	if !isEqual(a, b) {
		tt.Errorf("assert %d failed.\n"+
			"......found:%v (%[2]T)\n"+
			"...expected:%v (%[3]T)\n", n, a, b)
	}
}

// PrintValue will print the supplied value by throwing a testing failure.
func (tt *T) PrintValue(a interface{}) {
	if !tt.AssertPrintSwitch {
		tt.AssertPrintSwitch = true
	}
	fmt.Printf("%v(%[1]T)\n", a)
}

//================================================================================
