package cast

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"
)

func ref[F, T any](f F) T {
	return *(*T)(unsafe.Pointer(&f))
}

func testReflectValue[T any](t *testing.T, expectEqual bool) {
	var ins T
	p := ref[reflect.Value, value](reflect.ValueOf(ins)).ptr
	v := ref[T, unsafe.Pointer](ins)
	if (p == v) != expectEqual {
		t.Fatal(fmt.Sprintf("%T not expect: %p %p", ins, p, v))
	}
}

func TestReflectValue(t *testing.T) {
	testReflectValue[chan int](t, true)
	testReflectValue[func()](t, true)
	testReflectValue[func(string, string) string](t, true)
	testReflectValue[map[int]int](t, true)
	testReflectValue[*int](t, true)
	testReflectValue[unsafe.Pointer](t, true)
	testReflectValue[int](t, false)
}
