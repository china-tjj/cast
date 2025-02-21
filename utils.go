// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func typeFor[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func isMemSame(ta, tb reflect.Type) bool {
	if ta == tb {
		return true
	}
	kind := ta.Kind()
	if kind != tb.Kind() {
		return false
	}
	switch kind {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Interface, reflect.String:
		return true
	case reflect.Array:
		return ta.Len() == tb.Len() && isMemSame(ta.Elem(), tb.Elem())
	case reflect.Map:
		return isMemSame(ta.Key(), tb.Key()) && isMemSame(ta.Elem(), tb.Elem())
	case reflect.Chan, reflect.Pointer, reflect.Slice:
		return isMemSame(ta.Elem(), tb.Elem())
	case reflect.Struct:
		n := ta.NumField()
		if n != tb.NumField() {
			return false
		}
		for i := 0; i < n; i++ {
			fa := ta.Field(i)
			fb := tb.Field(i)
			tagA, okA := fa.Tag.Lookup("json")
			tagB, okB := fb.Tag.Lookup("json")
			if okA && okB && tagA != tagB {
				return false
			}
			if fa.Name != fb.Name {
				return false
			}
			if !isMemSame(fa.Type, fb.Type) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

type value struct {
	typ  unsafe.Pointer
	ptr  unsafe.Pointer
	flag uintptr
}

func getValueAddr(x reflect.Value) unsafe.Pointer {
	ptr := (*value)(unsafe.Pointer(&x)).ptr
	// 当可寻址时，ptr即该value的地址
	if x.CanAddr() {
		return ptr
	}
	// 指针类型对应的eface/iface/value里的data/ptr直接会直接copy这个指针
	// 值类型转对应的eface/iface/value里的data/ptr会指向这个值的拷贝
	switch x.Kind() {
	// chan 和 map 其实就是一个指针
	case reflect.Chan, reflect.Map, reflect.Pointer:
		return unsafe.Pointer(&ptr)
	default:
		return ptr
	}
}

func offset(data unsafe.Pointer, idx int, elemSize uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(data) + uintptr(idx)*elemSize)
}

func memCopy(dst, src unsafe.Pointer, size int) {
	copy(unsafe.Slice((*byte)(dst), size), unsafe.Slice((*byte)(src), size))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type slice struct {
	data unsafe.Pointer
	len  int
	cap  int
}

type str struct {
	data unsafe.Pointer
	len  int
}

func toBytes(s string) (b []byte) {
	from := (*str)(unsafe.Pointer(&s))
	to := (*slice)(unsafe.Pointer(&b))
	to.data = from.data
	to.len = from.len
	to.cap = from.len
	return
}

func toString(b []byte) (s string) {
	from := (*slice)(unsafe.Pointer(&b))
	to := (*str)(unsafe.Pointer(&s))
	to.data = from.data
	to.len = from.len
	return
}
