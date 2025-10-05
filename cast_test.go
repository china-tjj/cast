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
}

type I interface {
	f()
}

type S string

func (s S) f() {}

func TestIfaceCast(t *testing.T) {
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
					v1 *int
					v2 **string
				}{ptr(1), ptr2("2")},
			),
			struct {
				v1 ***string
				v2 ****int
			}{ptr3("1"), ptr4(2)},
			map[string]int{"v1": 1, "v2": 2},
			map[string]*string{"v1": ptr("1"), "v2": ptr("2")},
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
			map[int]bool{1: true, 2: true, 3: true},
			map[string]bool{"1": true, "2": true, "3": true},
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

type S1 struct {
	V1 int
	V2 []string
}

type S2 struct {
	V1 *string
	V2 []float64
}

func handCaster(s1 *S1) (*S2, error) {
	v1 := strconv.Itoa(s1.V1)
	v2 := make([]float64, 0, len(s1.V2))
	for _, v := range s1.V2 {
		newV, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, err
		}
		v2 = append(v2, newV)
	}
	return &S2{
		V1: &v1,
		V2: v2,
	}, nil
}

func BenchmarkStructCast(b *testing.B) {
	s1 := &S1{
		V1: 1,
		V2: []string{"1", "2"},
	}
	b.Run("HandCast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = handCaster(s1)
		}
	})
	b.Run("GetCaster", func(b *testing.B) {
		caster := GetCaster[*S1, *S2]()
		for i := 0; i < b.N; i++ {
			_, _ = caster(s1)
		}
	})
	b.Run("Cast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Cast[*S1, *S2](s1)
		}
	})
}
