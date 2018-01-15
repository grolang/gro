// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license as Go that can be found in the LICENSE file.

package ops

import (
	"fmt"
	"reflect"

	u8 "github.com/grolang/gro/utf88"
)

type (
	Any   = interface{}
	Slice []Any
	Map   struct {
		lkp map[Any]Any
		seq []Any
	}
	Func = func(ps ...Any) Any
)

//==============================================================================
/*
NewMap creates a new Map.
*/
func NewMap() *Map {
	return &Map{lkp: map[Any]Any{}, seq: []Any{}}
}

//String of Map returns a string representation of its value.
func (m *Map) String() string {
	str, first := "{", true
	for _, k := range m.seq {
		v := m.lkp[k]
		if first {
			first = false
		} else {
			str = str + ", "
		}
		str = str + fmt.Sprintf("%#v: %#v", k, v)
	}
	return str + "}"
}

func (m *Map) Add(key, val Any) *Map {
	m.lkp[key] = val
	m.seq = append(m.seq, key)
	return m
}

func (m *Map) Remove(key Any) {
	//
}

func (m *Map) Len() int {
	return 0
}

func (m *Map) Val(key Any) Any {
	return nil
}

//==============================================================================

/*
Assign makes the first arg point to the second arg.
*/
func Assign(pp *interface{}, k interface{}) *interface{} {
	k = widen(k)
	*pp = k
	return pp
}

//TODO: MultipleAssign ???

/*
PlusAssign adds the value/s to a vector or the key-value pair/s to a map,
otherwise calls Plus and Assign on the args.
*/
func PlusAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	case Slice:
		*pp = Plus(*pp, k)
		return pp
	case Map:
		*pp = Plus(*pp, k)
		return pp

	default:
		//TODO: any logic for arbitrary object?

		panic(fmt.Sprintf("PlusAssign: invalid parameter/s: %T", k))
		return pp
	}
}

/*
MinusAssign deletes an entry from a Map, otherwise calls Minus and Assign on the args.
*/
func MinusAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {

	// TODO: add Slice functionality

	default:
		*pp = Minus(*pp, k)
		return pp

		//panic(fmt.Sprintf("MinusAssign: invalid parameter/s: %T", k))
	}
}

/*
MultAssign calls Mult and Assign on the args.
*/
func MultAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	default:
		*pp = Mult(*pp, k)
		return pp

		//panic(fmt.Sprintf("MultAssign: invalid parameter/s: %T", k))
	}
}

/*
DivideAssign calls Divide and Assign on the args.
*/
func DivideAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	default:
		*pp = Divide(*pp, k)
		return pp

		//panic(fmt.Sprintf("DivideAssign: invalid parameter/s: %T", k))
	}
}

/*
ModAssign calls Mod and Assign on the args.
*/
func ModAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	default:
		*pp = Mod(*pp, k)
		return pp

		//panic(fmt.Sprintf("ModAssign: invalid parameter/s: %T", k))
	}
}

/*
SeqAssign calls Seq and Assign on the args.
*/
func SeqAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	default:
		*pp = Seq(*pp, k)
		return pp

		//panic(fmt.Sprintf("SeqAssign: invalid parameter/s: %T", k))
	}
}

/*
AltAssign calls Alt and Assign on the args.
*/
func AltAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	default:
		*pp = Alt(*pp, k)
		return pp

		//panic(fmt.Sprintf("AltAssign: invalid parameter/s: %T", k))
	}
}

/*
XorAssign calls Xor and Assign on the args.
*/
func XorAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	default:
		*pp = Xor(*pp, k)
		return pp

		//panic(fmt.Sprintf("XorAssign: invalid parameter/s: %T", k))
	}
}

/*
SeqXorAssign calls SeqXor and Assign on the args.
*/
func SeqXorAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	default:
		*pp = SeqXor(*pp, k)
		return pp

		panic(fmt.Sprintf("SeqXorAssign: invalid parameter/s: %T", k))
	}
}

/*
LeftShiftAssign calls LeftShift and Assign on the args.
*/
func LeftShiftAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	case Slice, Map:
		*pp = LeftShift(*pp, k)
		return pp

	default:
		//TODO: any logic for arbitrary object?
		panic(fmt.Sprintf("LeftShiftAssign: invalid parameter/s: %T", k))
	}
}

/*
RightShiftAssign calls RightShift and Assign on the args.
*/
func RightShiftAssign(pp *interface{}, k interface{}) *interface{} {
	//TODO: widen pp
	k = widen(k)
	switch (*pp).(type) {
	case Map:
		m := (*pp).(Map)
		if _, ok := m.lkp[k]; ok {
			delete(m.lkp, k)

			idx := 0
			for ; idx < len(m.seq); idx++ { //####
				if m.seq[idx] == k {
					break
				}
			}
			m.seq = append(m.seq[:idx], m.seq[idx+1:len(m.seq)]...)
			//fmt.Println(">>>", m.seq)
		}
		return pp

	default:
		*pp = RightShift(*pp, k)
		return pp

		//panic(fmt.Sprintf("RightShiftAssign: invalid parameter/s: %T", k))
	}
}

//==============================================================================

/*
Incr increments the arg in-place.
*/
func Incr(pn *interface{}) interface{} {
	temp := *pn
	switch (*pn).(type) {
	case Int, BigInt:
		*pn = Plus(*pn, 1)
	default:
		panic("Incr: non-integral arg")
	}
	return temp
}

/*
Decr decrements the arg in-place.
*/
func Decr(pn *interface{}) interface{} {
	temp := *pn
	switch (*pn).(type) {
	case Int, BigInt:
		*pn = Minus(*pn, 1)
	default:
		panic("Decr: non-integral arg")
	}
	return temp
}

//==============================================================================

/*
GetIndex gets the index of the 1st arg as specified by the 2nd,
and maybe also 3rd, arg/s.
*/
func GetIndex(pa *interface{}, ns ...interface{}) *interface{} {
	var ab, ae interface{}
	if len(ns) > 0 {
		ab = widen(ns[0])
		switch ab.(type) {
		case Int:
			ab = ab.(Int)
		case BigInt:
			ab = ab.(BigInt)
		case Infinity:
			ab = Inf
		default:
			ab = ab
		}
	} else {
		panic(fmt.Sprintf("GetIndex: 1st index must be supplied"))
	}
	if len(ns) > 1 {
		ae = widen(ns[1])
		switch ae.(type) {
		case Int:
			ae = ae.(Int)
		case BigInt:
			ae = ae.(BigInt)
		case Infinity:
			ae = Inf
		default:
			panic("GetIndex: invalid type of 2nd index arg")
		}
	}

	wpa := widen(*pa)
	pa = &wpa

	switch (*pa).(type) {
	case u8.Text:
		b := ToInt(ab)
		var e Int
		if ae != nil {
			e = ToInt(ae)
		} //else 2nd index not supplied

		s := (*pa).(u8.Text)
		lenS := Int(len(s))
		var is interface{}
		if ae == nil {
			is = s[b]
		} else if ae == Inf {
			is = s[b:lenS]
		} else { // 0 <= e < maxInt {
			is = s[b:e]
		}
		return &is

	case Slice:
		b := ToInt(ab)
		var e Int
		if ae != nil {
			e = ToInt(ae)
		} //else 2nd index not supplied

		s := (*pa).(Slice)
		lenS := Int(len(s))
		if ae == nil {
			return &s[b]
		} else if ae == Inf {
			is := interface{}(s[b:lenS])
			return &is
		} else { // 0 <= e < maxInt {
			is := interface{}(s[b:e])
			return &is
		}

	case Map:
		//TODO: ensure only one key given for Maps
		n := widen(ns[0])
		an := (*pa).(Map).lkp[n]
		return &an

	default:
		panic(fmt.Sprintf("GetIndex: invalid aggregate type: %T", *pa))
	}
}

/*
SetIndex sets the index specified by the second arg of the first arg to the third arg.
*/
func SetIndex(pa *interface{}, n, v interface{}) *interface{} {
	n = widen(n)
	v = widen(v)

	switch (*pa).(type) {
	case u8.Text:
		panic("SetIndex: Texts are immutable")

	case Slice:
		switch n.(type) {
		case Int:
			n = n.(Int)
		case BigInt:
			n = n.(BigInt)
		case Infinity:
			n = Inf
		default:
			n = n
		}
		nn := ToInt(n)
		[]interface{}((*pa).(Slice))[nn] = v
		an := []interface{}((*pa).(Slice))[nn]
		return &an

	case Map:
		map[interface{}]interface{}((*pa).(Map).lkp)[n] = v
		an := map[interface{}]interface{}((*pa).(Map).lkp)[n]
		return &an

	default:
		panic("SetIndex: invalid type")
	}
}

//==============================================================================

/*
NewSlice creates a new Slice.
*/
func NewSlice(n interface{}) interface{} {
	return Slice(make([]interface{}, n.(int)))
}

/*
InitSlice creates a new Slice and initializes it to the supplied values.
*/
func InitSlice(vs ...interface{}) interface{} {
	s := make([]interface{}, 0)
	for _, v := range vs {
		s = append(s, widen(v))
	}
	return Slice(s)
}

//==============================================================================

/*
MakeMap makes a new Map.
*/
func MakeMap() interface{} {
	return Map{lkp: map[Any]Any{}, seq: []Any{}}
}

/*
InitMap makes a new Map and initializes it to the supplied pair-values.
*/
func InitMap(es ...MapEntry) interface{} {
	m := Map{lkp: map[interface{}]interface{}{}, seq: []interface{}{}}
	for _, e := range es {
		m.lkp[widen(e.Key())] = widen(e.Val())
		m.seq = append(m.seq, e.Key()) //####
	}
	return m
}

//==============================================================================

/*
Len returns the length of the supplied Slice or Map.
*/
func Len(s interface{}) int {
	switch s.(type) {
	case Slice:
		return len([]interface{}(s.(Slice)))
	case Map:
		return len(s.(Map).lkp)
	default:
		panic(fmt.Sprintf("Len: unknown type %T", s))
	}
}

/*
Copy copies the values of a Slice or the entries of a Map
to a new one.
*/
func Copy(s interface{}) interface{} {
	s = widen(s)
	switch s.(type) {
	case Slice:
		so := make([]interface{}, len([]interface{}(s.(Slice))))
		copy(so, s.(Slice))
		return Slice(so)
	case Map:
		mo := Map{lkp: map[interface{}]interface{}{}, seq: []interface{}{}}
		//for k, v := range s.(Map).lkp {
		for _, k := range s.(Map).seq { //####
			v := s.(Map).lkp[k]
			mo.lkp[k] = v
			mo.seq = append(mo.seq, k)
		}
		return Map(mo)
	default:
		panic("Copy: unknown type")
	}
}

//==============================================================================

/*
Unwrap unwraps tuply-returned args into a slice.

TODO: test it
*/
func Unwrap(a ...interface{}) interface{} {
	if len(a) == 0 { //never happens
		return nil
	} else if len(a) == 1 {
		return a[0]
	} else {
		return a
	}
}

/*
Assert asserts the arg is true.

TODO: test it
*/
func Assert(b interface{}) {
	if !b.(bool) {
		fmt.Printf("assert failed.\n....found:%v (%[1]T)\n", b)
	}
}

/*
StructToMap converts a struct to a map of string to data.

TODO: test it
*/
func StructToMap(f interface{}) map[string]interface{} {
	fv := reflect.ValueOf(f)
	ft := reflect.TypeOf(f)
	m := map[string]interface{}{}
	for i := 0; i < fv.NumField(); i++ {
		m[ft.Field(i).Name] = fv.Field(i).Interface()
	}
	return m
}

/*
ArrayToSlice converts an array to a slice.

TODO: test it
*/
func ArrayToSlice(f interface{}) []interface{} {
	fv := reflect.ValueOf(f)
	//ft:= reflect.TypeOf(f) //need to check it's really an array
	vec := []interface{}{}
	for i := 0; i < fv.Len(); i++ {
		vec = append(vec, fv.Index(i).Interface())
	}
	return vec
}

//==============================================================================

/*
Prf printf's an untyped parameter, accepting any number of additional args
as values.
*/
func Prf(fs interface{}, is ...interface{}) {
	fmt.Print(sprf(fs, is...))
}

func Sprf(fs interface{}, is ...interface{}) u8.Text {
	return u8.Desur(sprf(fs, is...))
}

func sprf(fs interface{}, is ...interface{}) string {
	switch fs.(type) {
	case u8.Text:
		return fmt.Sprintf(u8.Surr(fs.(u8.Text)), is...)
	case string:
		return fmt.Sprintf(fs.(string), is...)
	case nil:
		return fmt.Sprint(is...)
	default:
		as := []interface{}{fs}
		for _, i := range is {
			as = append(as, i)
		}
		return fmt.Sprint(as...)
	}
}

func Pr(is ...interface{}) {
	fmt.Print(Spr(is...))
}

func Spr(is ...interface{}) u8.Text {
	return u8.Desur(fmt.Sprint(is...))
}

//==============================================================================
