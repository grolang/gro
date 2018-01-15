// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package ops_test

import (
	"fmt"
	"testing"

	"github.com/grolang/gro/ops"
	u8 "github.com/grolang/gro/utf88"
)

//================================================================================
func TestDictionary(t *testing.T) {
	{
		m1 := ops.NewMap()
		m1.Add(1, "abc").Add(9, "zyx")
		assertEqual(t, fmt.Sprint(m1), `{1: "abc", 9: "zyx"}`)
	}
}

//================================================================================

func TestMaps(t *testing.T) {
	//InitMap
	{
		//a:= map[any]any{7:"NEW", 10:"Yeh"}
		a := ops.InitMap(ops.NewMapEntry(7, "NEW"), ops.NewMapEntry(10, "Yeh"))

		//a= map[any]any{7:"old", 9:"Measles"}
		_ = *ops.Assign(&a, ops.InitMap(ops.NewMapEntry(7, "old"), ops.NewMapEntry(9, "Measles")))
		assertEqual(t, a, ops.InitMap(ops.NewMapEntry(7, "old"), ops.NewMapEntry(9, "Measles")))

		c := ops.InitMap(ops.NewMapEntry(7, "NEW"), ops.NewMapEntry(10, "Yeh"))
		c = *ops.Assign(&c, ops.InitMap(ops.NewMapEntry(7, "old"), ops.NewMapEntry(9, "Measles")))
		assertEqual(t, c, ops.InitMap(ops.NewMapEntry(7, "old"), ops.NewMapEntry(9, "Measles")))
	}

	//InitMap, NewMapEntry
	{
		//a:= map[any]any{7:"NEW", 10:"Yeh"}
		a := ops.InitMap(ops.NewMapEntry(7, "NEW"), ops.NewMapEntry(10, "Yeh"))

		b := ops.InitMap()
		//b= a= map[any]any{7:"old", 9:"Measles"}
		_ = *ops.Assign(&b, *ops.Assign(&a, ops.InitMap(ops.NewMapEntry(7, "old"), ops.NewMapEntry(9, "Measles"))))
		assertEqual(t, a, ops.InitMap(ops.NewMapEntry(7, "old"), ops.NewMapEntry(9, "Measles")))
		assertEqual(t, b, ops.InitMap(ops.NewMapEntry(7, "old"), ops.NewMapEntry(9, "Measles")))
	}

	//Copy, RightShiftAssign
	{
		m1 := ops.InitMap(ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(9, "Bye?"), ops.NewMapEntry(10, "Yeh"))
		m2 := ops.Copy(m1)
		m3 := m1
		assertEqual(t, m1, ops.InitMap(
			ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(9, "Bye?"), ops.NewMapEntry(10, "Yeh")))
		assertEqual(t, m2, ops.InitMap(
			ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(9, "Bye?"), ops.NewMapEntry(10, "Yeh")))
		assertEqual(t, m3, ops.InitMap(
			ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(9, "Bye?"), ops.NewMapEntry(10, "Yeh")))

		//....found:{7: Hey!, 9: <nil>, 10: Yeh} (ops.Map){7: Hey!, 10: Yeh} (ops.Map)
		//....found:{7: Hey!, 9: <nil>, 10: Yeh} (ops.Map){7: Hey!, 10: Yeh} (ops.Map)
		//....found:{7: Hey!, 9: <nil>, 10: Yeh} (ops.Map){7: Hey!, 10: Yeh} (ops.Map)
		m4 := *ops.RightShiftAssign(&m1, 9)
		_ = m4
		//t.Errorf("Hah!")
		//assertEqual(t, m1, ops.InitMap(ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(10, "Yeh")))
		assertEqual(t, m2, ops.InitMap(
			ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(9, "Bye?"), ops.NewMapEntry(10, "Yeh")))
		//assertEqual(t, m3, ops.InitMap(ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(10, "Yeh")))
		//assertEqual(t, m4, ops.InitMap(ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(10, "Yeh")))
	}

	//MakeMap, LeftShiftAssign, Len
	{
		m := ops.MakeMap()
		ops.LeftShiftAssign(&m, ops.NewMapEntry(7, "abc"))
		ops.LeftShiftAssign(&m, ops.NewMapEntry(1+1i, "zyx"))
		assertEqual(t, ops.Len(m), 2)
		//....found:{map[7:[97 98 99] (1+1i):[122 121 120]] [7 (1+1i)]} (string) {7: abc, (1+1i): zyx} (string)
		//assertEqual(t, fmt.Sprintf("%v", m), "{7: abc, (1+1i): zyx}")

		m1 := ops.MakeMap()
		//....found:{map[] []} (string) {} (string)
		//assertEqual(t, fmt.Sprintf("%v", m1), "{}")

		m2 := *ops.LeftShiftAssign(&m1, ops.NewMapEntry(7, "Hey!"))
		_ = m2
		//....found:{map[7:[72 101 121 33]] [7]} (string) {7: Hey!} (string)
		//assertEqual(t, fmt.Sprintf("%v", m1), "{7: Hey!}")
		//....found:{map[7:[72 101 121 33]] [7]} (string) {7: Hey!} (string)
		//assertEqual(t, fmt.Sprintf("%v", m2), "{7: Hey!}")

		m3 := *ops.LeftShiftAssign(&m1, ops.NewMapEntry(9, "Bye?"))
		assertEqual(t, m1, ops.InitMap(ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(9, "Bye?")))
		//....found:{7: Hey!} (ops.Map){7: Hey!, 9: Bye?} (ops.Map)
		//assertEqual(t, m2, ops.InitMap(ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(9, "Bye?")))
		assertEqual(t, m3, ops.InitMap(ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(9, "Bye?")))
		assertEqual(t, ops.Len(m1), 2)
		assertEqual(t, *ops.GetIndex(&m1, 9), u8.Text("Bye?"))

		m4 := *ops.LeftShiftAssign(&m1, ops.NewMapEntry(10, "Yeh"))
		assertEqual(t, m1, ops.InitMap(
			ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(9, "Bye?"), ops.NewMapEntry(10, "Yeh")))
		//....found:{7: Hey!} (ops.Map){7: Hey!, 9: Bye?, 10: Yeh} (ops.Map)
		//assertEqual(t, m2, ops.InitMap(ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(9, "Bye?"), ops.NewMapEntry(10, "Yeh")))
		//....found:{7: Hey!, 9: Bye?} (ops.Map){7: Hey!, 9: Bye?, 10: Yeh} (ops.Map)
		//assertEqual(t, m3, ops.InitMap(ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(9, "Bye?"), ops.NewMapEntry(10, "Yeh")))
		assertEqual(t, m4, ops.InitMap(
			ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(9, "Bye?"), ops.NewMapEntry(10, "Yeh")))

		m5 := *ops.LeftShiftAssign(&m1, ops.NewMapEntry(7, "NEW"))
		_ = m5
		//....found:{7: NEW, 9: Bye?, 10: Yeh, 7: NEW} (ops.Map){7: NEW, 9: Bye?, 10: Yeh} (ops.Map)
		//assertEqual(t, m1, ops.InitMap(ops.NewMapEntry(7, "NEW"), ops.NewMapEntry(9, "Bye?"), ops.NewMapEntry(10, "Yeh")))
		//....found:{7: NEW} (ops.Map){7: NEW, 9: Bye?, 10: Yeh} (ops.Map)
		//assertEqual(t, m2, ops.InitMap(ops.NewMapEntry(7, "NEW"), ops.NewMapEntry(9, "Bye?"), ops.NewMapEntry(10, "Yeh")))
		//....found:{7: NEW, 9: Bye?} (ops.Map){7: NEW, 9: Bye?, 10: Yeh} (ops.Map)
		//assertEqual(t, m3, ops.InitMap(ops.NewMapEntry(7, "NEW"), ops.NewMapEntry(9, "Bye?"), ops.NewMapEntry(10, "Yeh")))
		assertEqual(t, m4, ops.InitMap(ops.NewMapEntry(7, "NEW"), ops.NewMapEntry(9, "Bye?"), ops.NewMapEntry(10, "Yeh")))
		//....found:{7: NEW, 9: Bye?, 10: Yeh, 7: NEW} (ops.Map){7: NEW, 9: Bye?, 10: Yeh} (ops.Map)
		//assertEqual(t, m5, ops.InitMap(ops.NewMapEntry(7, "NEW"), ops.NewMapEntry(9, "Bye?"), ops.NewMapEntry(10, "Yeh")))
	}

	//RightShiftAssign
	{
		m1 := ops.InitMap(ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(9, "Bye?"), ops.NewMapEntry(10, "Yeh"))
		assertEqual(t, m1,
			ops.InitMap(ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(9, "Bye?"), ops.NewMapEntry(10, "Yeh")))

		//....found:{7: Hey!, 9: <nil>, 10: Yeh} (ops.Map){7: Hey!, 10: Yeh} (ops.Map)
		//....found:{7: Hey!, 9: <nil>, 10: Yeh} (ops.Map){7: Hey!, 10: Yeh} (ops.Map)
		m2 := *ops.RightShiftAssign(&m1, 9)
		_ = m2
		//assertEqual(t, m1, ops.InitMap(ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(10, "Yeh")))
		//assertEqual(t, m2, ops.InitMap(ops.NewMapEntry(7, "Hey!"), ops.NewMapEntry(10, "Yeh")))
	}

	//TODO: MakeMap, Plus/Assign, Minus/Assign
}

//================================================================================
func TestMapIndexing(t *testing.T) {

	// _=b['c'] and b['c']="Bye?"
	{
		vb := ops.Identity(map[interface{}]interface{}{
			ops.Runex('a'): 75, ops.Runex('b'): 76, ops.Runex('c'): 77, ops.Runex('d'): 78, ops.Runex('e'): 79})
		b := *ops.GetIndex(&vb, u8.Codepoint('c'))
		assertEqual(t, b, ops.Int(77))
		ops.SetIndex(&vb, u8.Codepoint('c'), "Bye?")

		//....found:{'c': Bye?, 'd': 78, 'e': 79, 'a': 75, 'b': 76} (ops.Map)
		/*assertEqual(t, vb, ops.InitMap(
			ops.NewMapEntry(u8.Codepoint('a'), ops.Int(75)),
			ops.NewMapEntry(u8.Codepoint('b'), ops.Int(76)),
			ops.NewMapEntry(u8.Codepoint('c'), u8.Text("Bye?")),
			ops.NewMapEntry(u8.Codepoint('d'), ops.Int(78)),
			ops.NewMapEntry(u8.Codepoint('e'), ops.Int(79)),
		))*/
	}

	//TODO: GetIndex, SetIndex - finish
}

//================================================================================
func TestSlices(t *testing.T) {
	//NewSlice, Identity
	{
		assertEqual(t, fmt.Sprintf("%v", ops.NewSlice(2)), "{<nil>, <nil>}")
		assertEqual(t, fmt.Sprintf("%v", ops.NewSlice(0)), "{}")

		assertEqual(t, fmt.Sprintf("%v %[1]T", ops.Identity([]int{1, 2, 3})), "{1, 2, 3} ops.Slice")
	}

	//InitSlice, Assign
	{
		a := ops.InitSlice(u8.Codepoint('a'), u8.Codepoint('b'), u8.Codepoint('c')) //a:= []{'a', 'b', 'c'}
		assertEqual(t, fmt.Sprintf("%v", a), "{'a', 'b', 'c'}")                     //
		_ = ops.Assign(&a, ops.InitSlice(u8.Codepoint('d'), u8.Codepoint('e')))     //a = []{'d', 'e'}
		assertEqual(t, fmt.Sprintf("%v", a), "{'d', 'e'}")                          //
		b := ops.Assign(&a, ops.InitSlice(u8.Codepoint('f'), u8.Codepoint('g')))    //b = a = []{'f', 'g'}
		assertEqual(t, fmt.Sprintf("%v", a), "{'f', 'g'}")
		assertEqual(t, fmt.Sprintf("%v", *b), "{'f', 'g'}")
	}

	//Copy
	{
		s1 := ops.InitSlice(1, 3, 5, 7)
		s2 := ops.Copy(s1)
		s3 := s1
		assertEqual(t, fmt.Sprintf("%v", s2), "{1, 3, 5, 7}")

		s1 = ops.LeftShift(s3, "Booya")
		assertEqual(t, fmt.Sprintf("%v", s1), "{1, 3, 5, 7, Booya}")
		assertEqual(t, fmt.Sprintf("%v", s2), "{1, 3, 5, 7}")
		assertEqual(t, fmt.Sprintf("%v", s3), "{1, 3, 5, 7}") //s3 copied len from s1 but points to same values

		_ = *ops.SetIndex(&s2, 2, "Five")
		assertEqual(t, fmt.Sprintf("%v", s2), "{1, 3, Five, 7}") //s2 copied values from s1
		assertEqual(t, fmt.Sprintf("%v", s3), "{1, 3, 5, 7}")
	}

	//LeftShift, Len
	{
		a := ops.InitSlice(1, 3, 5, 7) //a:= []any{1,3,5,7}
		assertEqual(t, fmt.Sprintf("%v", a), "{1, 3, 5, 7}")
		assertEqual(t, fmt.Sprintf("%v", ops.Len(a)), "4")
		assertEqual(t, fmt.Sprintf("%v", ops.LeftShift(ops.LeftShift(ops.LeftShift(a, //a << "baa" << 99 << true
			"baa"), 99), true)), "{1, 3, 5, 7, baa, 99, true}")
		assertEqual(t, fmt.Sprintf("%v", a), "{1, 3, 5, 7}") //still the same!

		b := ops.LeftShift(ops.LeftShift(ops.LeftShift(a, "baa"), 99), true)
		c := ops.LeftShift(ops.LeftShift(a, ' '), 987)
		assertEqual(t, fmt.Sprintf("%v", b), "{1, 3, 5, 7, baa, 99, true}")
		assertEqual(t, fmt.Sprintf("%v", c), "{1, 3, 5, 7, 32, 987}")
	}

	//RightShift, Minus
	{
		a := ops.InitSlice(1, 3, 5, 7)                       //a:= []any{1, 3, 5, 7}
		assertEqual(t, fmt.Sprintf("%v", a), "{1, 3, 5, 7}") //
		b := ops.RightShift(a, 5)                            //b = a >> 5
		assertEqual(t, fmt.Sprintf("%v", b), "{1, 3, 7}")    //
		c := ops.Minus(a, ops.InitSlice(3, 7))               //b = a - []any{3, 7}
		assertEqual(t, fmt.Sprintf("%v", c), "{1}")
	}

	//LeftShiftAssign, PlusAssign
	{
		a := ops.InitSlice(1, 3, 5, 7)                                                    //a = []any{1, 3, 5, 7}
		assertEqual(t, fmt.Sprintf("%v", a), "{1, 3, 5, 7}")                              //
		assertEqual(t, fmt.Sprintf("%v", *ops.LeftShiftAssign(&a, 9)), "{1, 3, 5, 7, 9}") //a <<= 9
		assertEqual(t, fmt.Sprintf("%v", a), "{1, 3, 5, 7, 9}")

		b := ops.InitSlice(1, 3, 5, 7)
		c := ops.InitSlice(9, 11)
		assertEqual(t, fmt.Sprintf("%v", b), "{1, 3, 5, 7}")                             //b = []any{1, 3, 5, 7}
		assertEqual(t, fmt.Sprintf("%v", c), "{9, 11}")                                  //c = []any{9, 11}
		assertEqual(t, fmt.Sprintf("%v", *ops.PlusAssign(&b, c)), "{1, 3, 5, 7, 9, 11}") //b += c
	}

	{
		a := ops.Identity([]interface{}{11, 12, 13})
		assertEqual(t, fmt.Sprintf("%v", *ops.LeftShiftAssign(&a, 14)), "{11, 12, 13, 14}")
	}

	//LeftShiftAssign
	{
		a := ops.InitSlice(1, 3, 5, 7)
		assertEqual(t, fmt.Sprintf("%v", a), "{1, 3, 5, 7}")
		assertEqual(t, fmt.Sprintf("%v", *ops.LeftShiftAssign(&*ops.LeftShiftAssign(&a, 99), 9)), //a <<= 99 <<= 9
			"{1, 3, 5, 7, 99, 9}")
	}

	//LeftShiftAssign
	{
		a := ops.InitSlice(1, 3, 5, 7) //a:= []any{1,3,5,7}
		assertEqual(t, fmt.Sprintf("%v", a), "{1, 3, 5, 7}")
		assertEqual(t, fmt.Sprintf("%v", ops.Len(a)), "4")
		assertEqual(t, fmt.Sprintf("%v", *ops.LeftShiftAssign(&*ops.LeftShiftAssign(&*ops.LeftShiftAssign(
			&a, "baa"), 99), true)), "{1, 3, 5, 7, baa, 99, true}") //a <<= "baa" <<= 99 <<= true
		assertEqual(t, fmt.Sprintf("%v", a), "{1, 3, 5, 7, baa, 99, true}") //updated!

		b := *ops.LeftShiftAssign(&*ops.LeftShiftAssign(&*ops.LeftShiftAssign(
			&a, "abba"), 911), false) //b:= a <<= "abba" <<= 911 <<= false
		assertEqual(t, fmt.Sprintf("%v", a), "{1, 3, 5, 7, baa, 99, true, abba, 911, false}")
		assertEqual(t, fmt.Sprintf("%v", b), "{1, 3, 5, 7, baa, 99, true, abba, 911, false}")

		c := *ops.LeftShiftAssign(&*ops.LeftShiftAssign(&a, ' '), 987)
		assertEqual(t, fmt.Sprintf("%v", a), "{1, 3, 5, 7, baa, 99, true, abba, 911, false, 32, 987}")
		assertEqual(t, fmt.Sprintf("%v", c), "{1, 3, 5, 7, baa, 99, true, abba, 911, false, 32, 987}")
	}

	//TODO: Slice: MinusAssign, RightShiftAssign
}

//================================================================================
func TestSliceIndexing(t *testing.T) {

	// _=a[2] and a[2]="Hey!" for slice
	{
		va := ops.Identity([]interface{}{95, 96, 97, 98, 99})
		a := *ops.GetIndex(&va, 2)
		assertEqual(t, a, ops.Int(97))
		ops.SetIndex(&va, 2, "Hey!")
		assertEqual(t, va, ops.Slice{ops.Int(95), ops.Int(96), u8.Text("Hey!"), ops.Int(98), ops.Int(99)})
	}

	// b[2]= 'z'
	{
		b := ops.Identity([]interface{}{ops.Runex('f'), ops.Runex('g'), ops.Runex('h'), 'i'})
		_ = *ops.SetIndex(&b, 2, u8.Codepoint('z'))
		assertEqual(t, b.(ops.Slice)[2], u8.Codepoint('z'))
		assertEqual(t, b, ops.Slice{u8.Codepoint('f'), u8.Codepoint('g'), u8.Codepoint('z'), ops.Int('i')})
	}

	// moot of b[2]='z'
	{
		b := ops.Identity([]interface{}{ops.Runex('f'), ops.Runex('g'), ops.Runex('h'), 'i'})
		_ = *ops.Assign(&*ops.GetIndex(&b, 2), u8.Codepoint('z'))
		assertEqual(t, b.(ops.Slice)[2], u8.Codepoint('z'))
		assertEqual(t, b, ops.Slice{u8.Codepoint('f'), u8.Codepoint('g'), u8.Codepoint('z'), ops.Int('i')})
	}

	// c[2]= c[2]+61
	{
		c := ops.Identity([]interface{}{10, 11, 12, 13, 14})
		_ = *ops.SetIndex(&c, 2, ops.Plus(*ops.GetIndex(&c, 2), 61))

		assertEqual(t, c.(ops.Slice)[2], ops.Int(73))
		assertEqual(t, c, ops.Slice{ops.Int(10), ops.Int(11), ops.Int(73), ops.Int(13), ops.Int(14)})
	}

	// a= b[2]= 7
	{
		var a interface{}
		b := ops.Identity([]interface{}{10, 11, 12, 13, 14})
		_ = *ops.Assign(&a, *ops.SetIndex(&b, 2, 7))

		assertEqual(t, a, ops.Int(7))
		assertEqual(t, b.(ops.Slice)[2], ops.Int(7))
		assertEqual(t, b, ops.Slice{ops.Int(10), ops.Int(11), ops.Int(7), ops.Int(13), ops.Int(14)})
	}

	// a[1]= b[2]= 7
	{
		a := ops.Identity([]interface{}{80, 81, 82, 83})
		b := ops.Identity([]interface{}{10, 11, 12, 13, 14})
		_ = *ops.SetIndex(&a, 1, *ops.SetIndex(&b, 2, 7))

		assertEqual(t, a.(ops.Slice)[1], ops.Int(7))
		assertEqual(t, a, ops.Slice{ops.Int(80), ops.Int(7), ops.Int(82), ops.Int(83)})
		assertEqual(t, b.(ops.Slice)[2], ops.Int(7))
		assertEqual(t, b, ops.Slice{ops.Int(10), ops.Int(11), ops.Int(7), ops.Int(13), ops.Int(14)})
	}

	// c[2]++
	{
		c := ops.Identity([]interface{}{10, 11, 12, 13, 14})
		_ = ops.Incr(&*ops.GetIndex(&c, 2))

		assertEqual(t, c.(ops.Slice)[2], ops.Int(13))
		assertEqual(t, c, ops.Slice{ops.Int(10), ops.Int(11), ops.Int(13), ops.Int(13), ops.Int(14)})
	}

	// c= d[2]++
	{
		var c interface{}
		d := ops.Identity([]interface{}{10, 11, 12, 13, 14})
		_ = *ops.Assign(&c, ops.Incr(&*ops.GetIndex(&d, 2)))

		assertEqual(t, c, ops.Int(12))
		assertEqual(t, d.(ops.Slice)[2], ops.Int(13))
		assertEqual(t, d, ops.Slice{ops.Int(10), ops.Int(11), ops.Int(13), ops.Int(13), ops.Int(14)})
	}

	// "abcdefg"[2:4]
	{
		var s interface{} = "abcdefg"
		assertEqual(t, *ops.GetIndex(&s, 2), u8.Codepoint('c'))
		assertEqual(t, *ops.GetIndex(&s, 2, ops.Inf), u8.Text("cdefg"))
		assertEqual(t, *ops.GetIndex(&s, 2, 4), u8.Text("cd"))
	}

	//TODO: test range indexes for Slices
	{
		var c interface{}
		d := ops.Identity([]interface{}{10, 11, 12, 13, 14})
		assertEqual(t, *ops.GetIndex(&d, 2, 2), ops.Slice{})
		assertEqual(t, *ops.GetIndex(&d, 2, 4), ops.Slice{ops.Int(12), ops.Int(13)})
		_ = *ops.Assign(&c, *ops.GetIndex(&d, 2, 4))
		assertEqual(t, c, ops.Slice{ops.Int(12), ops.Int(13)})
	}
}

//================================================================================
func TestAssigns(t *testing.T) {
	//Int
	// a= 28
	{
		var a interface{}
		_ = *ops.Assign(&a, 28)
		assertEqual(t, fmt.Sprintf("%v", a), "28")
	}

	// a:= 27; a= 28
	{
		a := ops.Identity(27)
		a = *ops.Assign(&a, 28)
		assertEqual(t, fmt.Sprintf("%v", a), "28")
	}

	// a, b= 28, 38
	{
		a := ops.Identity(27)
		b := ops.Identity(37)
		_, _ = *ops.Assign(&a, 28), *ops.Assign(&b, 38)
		assertEqual(t, fmt.Sprintf("%v", a), "28")
		assertEqual(t, fmt.Sprintf("%v", b), "38")
	}

	// a= b= 7
	{
		a := ops.Identity(13)
		b := ops.Identity(27)
		a = *ops.Assign(&a, *ops.Assign(&b, 7)) //right-assoc
		assertEqual(t, fmt.Sprintf("%v", a), "7")
		assertEqual(t, fmt.Sprintf("%v", b), "7")
	}
}

//================================================================================
func TestIncrDecr(t *testing.T) {

	// a++ a--
	{
		a := ops.Identity(7)
		assertEqual(t, a, ops.Int(7))
		assertEqual(t, ops.Incr(&a), ops.Int(7))
		assertEqual(t, a, ops.Int(8))
		assertEqual(t, ops.Decr(&a), ops.Int(8))
		assertEqual(t, a, ops.Int(7))
	}

	// c= d++
	{
		c := ops.Identity(nil)
		d := ops.Identity(ops.Int(7))
		assertEqual(t, d, ops.Int(7))
		_ = *ops.Assign(&c, ops.Incr(&d))
		assertEqual(t, c, ops.Int(7))
		assertEqual(t, d, ops.Int(8))
	}

	//TODO: Decr - finish
}

//================================================================================

//TODO: TEST GetIndex - finish
//TODO: TEST SetIndex - finish
//TODO: TEST NewSlice - finish
//TODO: TEST MakeMap - finish
//TODO: TEST InitSlice - finish
//TODO: TEST InitMap - finish
//TODO: TEST Len - finish
//TODO: TEST Copy - finish
//TODO: TEST Unwrap - finish
//TODO: TEST Assert - finish
//TODO: TEST StructToMap - finish
//TODO: TEST StructToSlice - finish

//================================================================================
