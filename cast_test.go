package cast

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"testing"
	"unsafe"
)

func ptr[T any](v T) *T {
	return &v
}

func ptr2[T any](v T) **T {
	return ptr(ptr(v))
}

func ptr3[T any](v T) ***T {
	return ptr(ptr2(v))
}

func ptr4[T any](v T) ****T {
	return ptr(ptr3(v))
}

func TestSliceToArrayPtr(t *testing.T) {
	s := []int{1, 2}
	a1, err1 := Cast[[]int, *[1]int](s)
	if err1 != nil {
		t.Fatal()
	}
	a2, err2 := Cast[[]int, *[2]int](s)
	if err2 != nil {
		t.Fatal()
	}
	a3, err3 := Cast[[]int, *[3]int](s)
	if err3 != nil {
		t.Fatal()
	}
	s[0] = 100
	if a1[0] != 100 {
		t.Fatal()
	}
	if a2[0] != 100 {
		t.Fatal()
	}
	if a3[0] != 1 {
		t.Fatal()
	}
}

func TestSlicePtrToArrayPtr(t *testing.T) {
	s := ptr3[[]int]([]int{1, 2})
	a1, err1 := Cast[***[]int, ***[1]int](s)
	if err1 != nil {
		t.Fatal()
	}
	a2, err2 := Cast[***[]int, ***[2]int](s)
	if err2 != nil {
		t.Fatal()
	}
	a3, err3 := Cast[***[]int, ***[3]int](s)
	if err3 != nil {
		t.Fatal()
	}
	(***s)[0] = 100
	if (***a1)[0] != 100 {
		t.Fatal()
	}
	if (***a2)[0] != 100 {
		t.Fatal()
	}
	if (***a3)[0] != 1 {
		t.Fatal()
	}
}

func TestArrayPtrToPtr(t *testing.T) {
	var arr [10]int
	ptr_, err := Cast[*[10]int, *int](&arr)
	if err != nil || unsafe.Pointer(ptr_) != unsafe.Pointer(&arr) {
		t.Fatal()
	}
}

func TestSliceToPtr(t *testing.T) {
	var arr = []int{1, 2}
	ptr_, err := Cast[[]int, *int](arr)
	if err != nil || unsafe.Pointer(ptr_) != unsafe.Pointer(&arr[0]) {
		t.Fatal()
	}
}

func TestSlicePtrToPtr(t *testing.T) {
	var arr = []int{1, 2}
	ptr_, err := Cast[*[]int, *int](&arr)
	if err != nil || unsafe.Pointer(ptr_) != unsafe.Pointer(&arr[0]) {
		t.Fatal()
	}
}

func TestFuncCast(t *testing.T) {
	add := func(a int, b int) int {
		return a + b
	}
	strAdd, err := Cast[func(a int, b int) int, func(a string, b string) string](add)
	if err != nil {
		t.Fatal()
	}
	c := strAdd("1", "2")
	if c != "3" {
		t.Fatal(c)
	}

	strAdd2, err := Cast[func(a int, b int) int, func(a string, b string) (string, error)](add)
	if err != nil {
		t.Fatal()
	}
	c, err = strAdd2("1", "2")
	if c != "3" || err != nil {
		t.Fatal(c, err)
	}
	c, err = strAdd2("1", "abc")
	if err == nil || err.Error() != `cast in arg 1 from string to int failed: strconv.ParseInt: parsing "abc": invalid syntax` {
		t.Fatal(err)
	}
}

func TestSliceFuncCast(t *testing.T) {
	add := func(a int, b int, others ...int) int {
		res := a + b
		for _, v := range others {
			res += v
		}
		return res
	}
	strAdd, err := Cast[func(a int, b int, others ...int) int, func(a string, b string, others ...string) string](add)
	if err != nil {
		t.Fatal()
	}
	c := strAdd("1", "2", "3", "4")
	if c != "10" {
		t.Fatal(c)
	}
}

type I interface {
	f()
}

type S string

func (s S) f() {}

func TestIfaceCast(t *testing.T) {
	s, err := Cast[I, string](S("123"))
	if err != nil || s != "123" {
		t.Fatal(err)
	}
}

func TestIfaceMapCast(t *testing.T) {
	iMap := map[I]struct{}{
		S("123"): {},
	}
	stringMap, err := Cast[map[I]struct{}, map[string]struct{}](iMap)
	if err != nil {
		t.Fatal()
	}
	_, exist := stringMap["123"]
	if !exist {
		t.Fatal()
	}
}

func TestNilAnyCast(t *testing.T) {
	if v, err := Cast[any, *int](nil); err != nil || v != nil {
		t.Fatal()
	}
	if v, err := Cast[any, chan any](nil); err != nil || v != nil {
		t.Fatal()
	}
	if v, err := Cast[any, map[any]any](nil); err != nil || v != nil {
		t.Fatal()
	}
	if v, err := Cast[any, error](nil); err != nil || v != nil {
		t.Fatal()
	}
	if v, err := Cast[any, []any](nil); err != nil || v != nil {
		t.Fatal()
	}
}

type S1 struct {
	*S2
	V1 int
}
type S2 struct {
	*S1
	V2 int
}

type S3 struct {
	*S1 `cast:"S1"`
	V3  int
}

func TestSelfAnonymousStruct(t *testing.T) {
	s1 := &S1{
		V1: 1,
		S2: &S2{
			V2: 2,
			S1: &S1{
				V1: 3,
			},
		},
	}
	m1, err := To[map[string]any](s1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m1, map[string]any{"V1": 1, "V2": 2}) {
		t.Fatal()
	}

	s3 := &S3{
		V3: 3,
		S1: &S1{
			V1: 1,
		},
	}
	m3, err := To[map[string]any](s3)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m3, map[string]any{"V3": 3, "S1": &S1{V1: 1}}) {
		t.Fatal()
	}
}

func TestInternalHack(t *testing.T) {
	type S struct {
		V int
	}
	m, err := To[map[*string]any](S{})
	if err != nil || len(m) == 0 {
		t.Fatal(err)
	}
	for k := range m {
		*k = "hack"
	}
	m, err = To[map[*string]any](S{})
	if err != nil {
		t.Fatal(err)
	}
	for k := range m {
		if *k != "V" {
			t.Fatal(*k)
		}
	}
}

type IntValue int

func (i IntValue) String() string {
	return fmt.Sprintf("value_%d", i)
}

type Struct struct {
	Value1 int `cast:"value_1"`
}

func TestMapToStruct(t *testing.T) {
	m := map[IntValue]int{1: 1}
	s, err := To[Struct](m)
	if err != nil || s.Value1 != 1 {
		t.Fatal(err)
	}
}

func TestCast(t *testing.T) {
	type S1 struct {
		A int
	}
	type S2 struct {
		S1
		B int
	}
	type S3 struct {
		*S1
		B int
	}
	type S4 struct {
		A int
		B int
	}
	equalGroups := [][]any{
		{
			int(123), int8(123), int16(123), int32(123), int64(123),
			uint(123), uint8(123), uint16(123), uint32(123), uint64(123), uintptr(123),
			"123", float32(123), float64(123),
		},
		{
			int(123), int8(123), int16(123), int32(123), int64(123),
			uint(123), uint8(123), uint16(123), uint32(123), uint64(123),
			"123", float32(123), float64(123),
			ptr("123"), ptr2(123), ptr3(123.),
		},
		{
			[1]int{123},
			[]int{123},
			[1]int32{123},
			[]int32{123},
		},
		{
			ptr4(
				struct {
					V1 *int
					V2 **string
				}{ptr(1), ptr2("2")},
			),
			struct {
				V1 ***string
				V2 ****int
			}{ptr3("1"), ptr4(2)},
			map[string]int{"V1": 1, "V2": 2},
			map[string]*string{"V1": ptr("1"), "V2": ptr("2")},
		},
		{
			struct {
				V1 int    `json:"v1"`
				V2 string `json:"v2"`
			}{1, "2"},
			struct {
				V1 string `json:"v1"`
				V2 int    `json:"v2"`
			}{"1", 2},
			map[string]int{"v1": 1, "v2": 2},
			map[string]string{"v1": "1", "v2": "2"},
		},
		{
			"一二三",
			[]byte("一二三"),
			*(*[9]byte)([]byte("一二三")),
		},
		{
			"一二三",
			[]rune("一二三"),
			*(*[3]rune)([]rune("一二三")),
		},
		{
			"123",
			[]byte("123"),
			[]rune("123"),
			*(*[3]byte)([]byte("123")),
			*(*[3]rune)([]rune("123")),
		},
		{
			map[string]bool{"1": true, "2": true, "3": true},
			map[int]bool{1: true, 2: true, 3: true},
			map[int32]bool{1: true, 2: true, 3: true},
			map[int64]bool{1: true, 2: true, 3: true},
			map[uint]bool{1: true, 2: true, 3: true},
			map[uint32]bool{1: true, 2: true, 3: true},
			map[uint64]bool{1: true, 2: true, 3: true},
			map[uintptr]bool{1: true, 2: true, 3: true},
		},
		{
			ptr(any(error(strErr("123")))),
			ptr(error(strErr("123"))),
		},
		{
			S2{S1{1}, 2},
			S3{&S1{1}, 2},
			S4{1, 2},
			map[string]any{
				"A": 1,
				"B": 2,
			},
		},
	}
	testScopes := []*Scope{defaultScope, deepCopyScope}
	for _, s := range testScopes {
		for _, group := range equalGroups {
			for _, a := range group {
				for _, b := range group {
					func() {
						defer func() {
							if r := recover(); r != nil {
								const size = 64 << 10
								buf := make([]byte, size)
								buf = buf[:runtime.Stack(buf, false)]
								t.Fatal(fmt.Sprintf("cast %T(%+v) to %T(%+v) panic: \nerror: %v\nstack: %s", a, a, b, b, r, buf))
							}
						}()
						res, err := ReflectCastWithScope(s, reflect.ValueOf(a), reflect.TypeOf(b))
						if err != nil {
							t.Fatal(fmt.Sprintf("cast %T(%+v) to %T(%+v) error: %v", a, a, b, b, err))
						}
						v := res.Interface()
						if !reflect.DeepEqual(v, b) {
							t.Fatal(fmt.Sprintf("cast %T(%+v) to %T(%+v) not equal, casted value is %T(%+v)",
								a, a, b, b, v, v))
						}
					}()
				}
			}
		}
	}
}

type FromStruct struct {
	V1 float64
	V2 []complex128
	V3 map[int]int
}

type ToStruct struct {
	V1 *string
	V2 []*string
	V3 map[string]*string
}

func ManualCast(from *FromStruct) (*ToStruct, error) {
	v1 := strconv.FormatFloat(from.V1, 'f', -1, 64)
	v2 := make([]*string, len(from.V2))
	for i, v := range from.V2 {
		newV := strconv.FormatComplex(v, 'f', -1, 128)
		v2[i] = &newV
	}
	v3 := map[string]*string{}
	for k, v := range from.V3 {
		newK := strconv.Itoa(k)
		newV := strconv.Itoa(v)
		v3[newK] = &newV
	}
	return &ToStruct{
		V1: &v1,
		V2: v2,
		V3: v3,
	}, nil
}

func BenchmarkStructCast(b *testing.B) {
	from := &FromStruct{
		V1: 1,
		V2: []complex128{2 + 3i, 4 + 5i, 6 + 7i, 8 + 9i},
		V3: map[int]int{10: 11, 12: 13, 14: 15, 16: 17, 18: 19},
	}
	b.Run("ManualCast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ManualCast(from)
		}
	})
	b.Run("GetCaster", func(b *testing.B) {
		caster := GetCaster[*FromStruct, *ToStruct]()
		for i := 0; i < b.N; i++ {
			_, _ = caster(from)
		}
	})
	b.Run("Cast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Cast[*FromStruct, *ToStruct](from)
		}
	})
	b.Run("To", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = To[*ToStruct](from)
		}
	})
}
