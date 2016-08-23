// Copyright 2016 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package assert_test

import(
  "testing"
  ts "github.com/grolang/gro/assert"
)

type A struct{
  slice []interface{}
}

func (this *A) lengthenByOne() *A {
  o:= len(this.slice)
  this.slice = this.slice[ : o + 1]
  return this
}

//================================================================================
func TestAssertSkeletons(t *testing.T){
  ts.AssertTrue(1 == 1)

  ts.LogAsserts("AssertSkeletons", t, func(tt *ts.T){

      tt.Assert(true)
      tt.AssertEqual(1, 1)
      tt.AssertEqualN('a', 'a', 123)
      //tt.PrintValue("abcdefg")
      tt.AssertPanic(func(){
        if true { panic("X") }
      }, "X")
      tt.AssertAnyPanic(func(){
        if true { panic("X") }
      })
  })
}

//================================================================================
func TestAssert(t *testing.T){
  ts.LogAsserts("Assert", t, func(tt *ts.T){

    a:= &A{ make([]interface{}, 0, 10) }
    tt.Assert(len(a.slice) == 0) //(1)
    b:= a.lengthenByOne()

    //tt.PrintValue(1234567890) //print test

    tt.Assert(len(b.slice) == 1) //(2)
    tt.Assert(len(a.slice) == 1) //(3)

    a.slice = a.slice[ : 3]
    tt.Assert(len(a.slice) == 3) //(4)
    tt.Assert(len(b.slice) == 3) //(5)

    c:= a.lengthenByOne()
    tt.Assert(len(a.slice) == 4) //(6)
    tt.Assert(len(b.slice) == 4) //(7)
    tt.Assert(len(c.slice) == 4) //(8)

    tt.AssertPanic(func(){
      if len(c.slice) == 4 { panic("X") } //(9)
    }, "X")

    tt.AssertAnyPanic(func(){
      if len(c.slice) == 4 { panic("X") } //(10)
    })

    /*tt.AssertPanic(func(){
      if len(c.slice) == 4 { panic("X") } //(11)
    }, "Y")*/

    //tt.Assert(false) //(12) force failure
  })
}

//================================================================================
