// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getArrayCaster(fromType, toType reflect.Type) castFunc {
	switch fromType.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		if toType.Len() <= 0 {
			return nil
		}
		caster := getCaster(fromType, toType.Elem())
		if caster == nil {
			return nil
		}
		return caster
	case reflect.Array:
		fromElemType := fromType.Elem()
		toElemType := toType.Elem()
		if isMemSame(fromElemType, toElemType) {
			size := minInt(int(fromType.Size()), int(toType.Size()))
			return func(fromAddr, toAddr unsafe.Pointer) bool {
				memCopy(toAddr, fromAddr, size)
				return true
			}
		}
		elemCaster := getCaster(fromElemType, toElemType)
		if elemCaster == nil {
			return nil
		}
		length := minInt(fromType.Len(), toType.Len())
		fromElemSize := fromElemType.Size()
		toElemSize := toElemType.Size()
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			for i := 0; i < length; i++ {
				elemCaster(offset(fromAddr, i, fromElemSize), offset(toAddr, i, toElemSize))
			}
			return true
		}
	case reflect.Interface:
		return getUnpackInterfaceCaster(fromType, toType)
	case reflect.Pointer:
		return getAddressingPointerCaster(fromType, toType)
	case reflect.Slice:
		fromElemType := fromType.Elem()
		toElemType := toType.Elem()
		if isMemSame(fromElemType, toElemType) {
			fromElemSize := int(fromElemType.Size())
			toSize := int(toType.Size())
			return func(fromAddr, toAddr unsafe.Pointer) bool {
				from := (*slice)(fromAddr)
				memCopy(toAddr, from.data, minInt(from.len*fromElemSize, toSize))
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
			for i := 0; i < from.len; i++ {
				elemCaster(offset(from.data, i, fromElemSize), offset(toAddr, i, toElemSize))
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
		return getUnpackStructCaster(fromType, toType)
	default:
		return nil
	}
}
