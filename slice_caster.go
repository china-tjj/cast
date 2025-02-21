// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func makeSlice(ptr unsafe.Pointer, tpy reflect.Type, len, cap int) *slice {
	res := (*slice)(ptr)
	*res = *(*slice)(getValueAddr(reflect.MakeSlice(tpy, len, cap)))
	return res
}

func getSliceCaster(fromType, toType reflect.Type) castFunc {
	switch fromType.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		caster := getCaster(fromType, toType.Elem())
		if caster == nil {
			return nil
		}
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			to := makeSlice(toAddr, toType, 1, 1)
			return caster(fromAddr, to.data)
		}
	case reflect.Array:
		fromElemType := fromType.Elem()
		toElemType := toType.Elem()
		length := fromType.Len()
		if isMemSame(fromElemType, toElemType) {
			return func(fromAddr, toAddr unsafe.Pointer) bool {
				to := (*slice)(toAddr)
				to.data = fromAddr
				to.len = length
				to.cap = length
				return true
			}
		}
		elemCaster := getCaster(fromElemType, toElemType)
		if elemCaster == nil {
			return nil
		}
		fromElemSize := fromElemType.Size()
		toElemSize := toElemType.Size()
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			to := makeSlice(toAddr, toType, length, length)
			for i := 0; i < length; i++ {
				elemCaster(offset(fromAddr, i, fromElemSize), offset(to.data, i, toElemSize))
			}
			return true
		}
	case reflect.Interface:
		return getUnwrapInterfaceCaster(fromType, toType)
	case reflect.Pointer:
		return getAddressingPointerCaster(fromType, toType)
	case reflect.Slice:
		fromElemType := fromType.Elem()
		toElemType := toType.Elem()
		if isMemSame(fromElemType, toElemType) {
			return func(fromAddr, toAddr unsafe.Pointer) bool {
				*(*slice)(toAddr) = *(*slice)(fromAddr)
				return true
			}
		}
		elemCaster := getCaster(fromElemType, toElemType)
		if elemCaster == nil {
			return nil
		}
		fromElemSize := fromElemType.Size()
		toElemSize := toElemType.Size()
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			from := (*slice)(fromAddr)
			to := makeSlice(toAddr, toType, from.len, from.cap)
			for i := 0; i < from.len; i++ {
				elemCaster(offset(from.data, i, fromElemSize), offset(to.data, i, toElemSize))
			}
			return true
		}
	case reflect.String:
		bytesCaster := getCaster(typeFor[[]byte](), toType)
		if bytesCaster == nil {
			return nil
		}
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			bytes := toBytes(*(*string)(fromAddr))
			return bytesCaster(unsafe.Pointer(&bytes), toAddr)
		}
	case reflect.Struct:
		return getUnwrapStructCaster(fromType, toType)
	default:
		return nil
	}
}
