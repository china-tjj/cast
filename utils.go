// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func isMemSame(fromType, toType reflect.Type) bool {
	if fromType == toType {
		return true
	}
	fromKind := fromType.Kind()
	toKind := toType.Kind()
	if fromKind != toKind {
		return fromKind == reflect.Pointer && toKind == reflect.UnsafePointer ||
			fromKind == reflect.UnsafePointer && toKind == reflect.Pointer
	}
	switch fromKind {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.String, reflect.UnsafePointer:
		return true
	case reflect.Array:
		return fromType.Len() == toType.Len() && isMemSame(fromType.Elem(), toType.Elem())
	case reflect.Chan:
		fromDir := fromType.ChanDir()
		return (fromDir == reflect.BothDir || fromDir == toType.ChanDir()) && isMemSame(fromType.Elem(), toType.Elem())
	case reflect.Func:
		numIn, numOut := fromType.NumIn(), fromType.NumOut()
		if numIn != toType.NumIn() || numOut != toType.NumOut() {
			return false
		}
		for i := 0; i < numIn; i++ {
			if !isMemSame(fromType.In(i), toType.In(i)) {
				return false
			}
		}
		for i := 0; i < numOut; i++ {
			if !isMemSame(fromType.Out(i), toType.Out(i)) {
				return false
			}
		}
		return true
	case reflect.Interface:
		return fromType.Implements(toType) && toType.Implements(fromType)
	case reflect.Map:
		return isMemSame(fromType.Key(), toType.Key()) && isMemSame(fromType.Elem(), toType.Elem())
	case reflect.Pointer, reflect.Slice:
		return isMemSame(fromType.Elem(), toType.Elem())
	case reflect.Struct:
		n := fromType.NumField()
		if n != toType.NumField() {
			return false
		}
		for i := 0; i < n; i++ {
			fromField := fromType.Field(i)
			toField := toType.Field(i)
			fromJsonTag, ok1 := fromField.Tag.Lookup("json")
			toJsonTag, ok2 := toField.Tag.Lookup("json")
			if fromField.Name != toField.Name || (ok1 && ok2 && fromJsonTag != toJsonTag) {
				return false
			}
			if !isMemSame(fromField.Type, toField.Type) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func typeFor[T any]() reflect.Type {
	var v T
	if t := reflect.TypeOf(v); t != nil {
		return t // optimize for T being a non-interface kind
	}
	return reflect.TypeOf((*T)(nil)).Elem() // only for an interface kind
}

func typePtr(t reflect.Type) unsafe.Pointer {
	return (*eface)(unsafe.Pointer(&t)).ptr
}

func min[T Number](a, b T) T {
	if a < b {
		return a
	}
	return b
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

func unpackInterface(numMethod int, ptr unsafe.Pointer) (reflect.Type, unsafe.Pointer) {
	var elem any
	if numMethod == 0 {
		elem = *(*any)(ptr)
	} else {
		elem = any(*(*iface)(ptr))
	}
	switch v := (elem).(type) {
	case bool:
		return boolType, unsafe.Pointer(&v)
	case int:
		return intType, unsafe.Pointer(&v)
	case int8:
		return int8Type, unsafe.Pointer(&v)
	case int16:
		return int16Type, unsafe.Pointer(&v)
	case int32:
		return int32Type, unsafe.Pointer(&v)
	case int64:
		return int64Type, unsafe.Pointer(&v)
	case uint:
		return uintType, unsafe.Pointer(&v)
	case uint8:
		return uint8Type, unsafe.Pointer(&v)
	case uint16:
		return uint16Type, unsafe.Pointer(&v)
	case uint32:
		return uint32Type, unsafe.Pointer(&v)
	case uint64:
		return uint64Type, unsafe.Pointer(&v)
	case uintptr:
		return uintptrType, unsafe.Pointer(&v)
	case float32:
		return float32Type, unsafe.Pointer(&v)
	case float64:
		return float64Type, unsafe.Pointer(&v)
	case string:
		return stringType, unsafe.Pointer(&v)
	case []byte:
		return bytesType, unsafe.Pointer(&v)
	case []rune:
		return runesType, unsafe.Pointer(&v)
	case map[string]any:
		return jsonMapType, unsafe.Pointer(&v)
	case []any:
		return jsonListType, unsafe.Pointer(&v)
	default:
		break
	}
	unpackedType := reflect.TypeOf(elem)
	unpackedPtr := (*eface)(unsafe.Pointer(&elem)).ptr
	// 指针类型对应的eface/iface/value里的data/ptr直接会直接copy这个指针
	// 值类型转对应的eface/iface/value里的data/ptr会指向这个值的拷贝
	switch unpackedType.Kind() {
	// chan、map、func 其实就是一个指针
	case reflect.Chan, reflect.Map, reflect.Func, reflect.Pointer, reflect.UnsafePointer:
		return unpackedType, unsafe.Pointer(&unpackedPtr)
	default:
		return unpackedType, unpackedPtr
	}
}

func getValueAddr(v reflect.Value) unsafe.Pointer {
	ptr := (*value)(unsafe.Pointer(&v)).ptr
	// 当可寻址时，ptr即该value的地址
	if v.CanAddr() {
		return ptr
	}
	// 指针类型对应的eface/iface/value里的data/ptr直接会直接copy这个指针
	// 值类型转对应的eface/iface/value里的data/ptr会指向这个值的拷贝
	switch v.Kind() {
	// chan、map、func 其实就是一个指针
	case reflect.Chan, reflect.Map, reflect.Func, reflect.Pointer, reflect.UnsafePointer:
		return unsafe.Pointer(&ptr)
	default:
		return ptr
	}
}

func offset(data unsafe.Pointer, idx int, elemSize uintptr) unsafe.Pointer {
	return unsafe.Add(data, uintptr(idx)*elemSize)
}

func memCopy(dst, src unsafe.Pointer, size uintptr) {
	copy(unsafe.Slice((*byte)(dst), size), unsafe.Slice((*byte)(src), size))
}

func malloc(size uintptr) unsafe.Pointer {
	bytes := make([]byte, size)
	return (*slice)(unsafe.Pointer(&bytes)).data
}

type slice struct {
	data unsafe.Pointer
	len  int
	cap  int
}

func makeSlice(elemSize uintptr, len, cap int) slice {
	return slice{
		data: malloc(elemSize * uintptr(cap)),
		len:  len,
		cap:  cap,
	}
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

func isSpecialHash(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Array:
		return typ.Len() > 0 && isSpecialHash(typ.Elem())
	case reflect.String, reflect.Interface:
		return true
	case reflect.Struct:
		n := typ.NumField()
		for i := 0; i < n; i++ {
			if isSpecialHash(typ.Field(i).Type) {
				return true
			}
		}
		return false
	default:
		return false
	}
}
