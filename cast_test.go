package cast

import (
	"errors"
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

func TestCast(t *testing.T) {
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
		},
		{
			"一二三",
			[]rune("一二三"),
		},
		{
			"123",
			[]byte("123"),
			[]rune("123"),
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
			ptr(any(errors.New("123"))),
			ptr(errors.New("123")),
		},
	}
	for _, group := range equalGroups {
		for i := 0; i < len(group); i++ {
			for j := 0; j < len(group); j++ {
				func() {
					a, b := group[i], group[j]
					defer func() {
						if r := recover(); r != nil {
							const size = 64 << 10
							buf := make([]byte, size)
							buf = buf[:runtime.Stack(buf, false)]
							t.Fatal(fmt.Sprintf("cast %T(%+v) to %T(%+v) panic: \nerror: %v\nstack: %s", a, a, b, b, r, buf))
						}
					}()
					res, err := ReflectCast(reflect.ValueOf(a), reflect.TypeOf(b))
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
}
