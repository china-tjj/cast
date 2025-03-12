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

func getFieldName(field *reflect.StructField) string {
	if tag, ok := field.Tag.Lookup("json"); ok {
		return tag
	} else {
		return field.Name
	}
}

func isMemSame(ta, tb reflect.Type) bool {
	if ta == tb {
		return true
	}
	ka := ta.Kind()
	kb := tb.Kind()
	if ka != kb {
		return ka == reflect.Pointer && kb == reflect.UnsafePointer ||
			ka == reflect.UnsafePointer && kb == reflect.Pointer
	}
	switch ka {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.String, reflect.UnsafePointer:
		return true
	case reflect.Array:
		return ta.Len() == tb.Len() && isMemSame(ta.Elem(), tb.Elem())
	case reflect.Interface:
		return ta.Implements(tb) && tb.Implements(ta)
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
			if getFieldName(&fa) != getFieldName(&fb) {
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

type iface interface {
	M()
}

type eface struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

func unpackInterface(typ reflect.Type, ptr unsafe.Pointer) (reflect.Type, unsafe.Pointer) {
	var elem any
	if typ.NumMethod() == 0 {
		elem = *(*any)(ptr)
	} else {
		elem = any(*(*iface)(ptr))
	}
	unpackedType := reflect.TypeOf(elem)
	unpackedPtr := (*eface)(unsafe.Pointer(&elem)).ptr
	// 指针类型对应的eface/iface/value里的data/ptr直接会直接copy这个指针
	// 值类型转对应的eface/iface/value里的data/ptr会指向这个值的拷贝
	switch unpackedType.Kind() {
	// chan 和 map 其实就是一个指针
	case reflect.Chan, reflect.Map, reflect.Pointer:
		return unpackedType, unsafe.Pointer(&unpackedPtr)
	default:
		return unpackedType, unpackedPtr
	}
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
