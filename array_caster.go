// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getArrayCaster(s *Scope, fromType, toType reflect.Type) castFunc {
	switch fromType.Kind() {
	case reflect.Array:
		fromElemType := fromType.Elem()
		toElemType := toType.Elem()
		if isMemSame(fromElemType, toElemType) {
			size := min(fromType.Size(), toType.Size())
			return func(fromAddr, toAddr unsafe.Pointer) error {
				memCopy(toAddr, fromAddr, size)
				return nil
			}
		}
		elemCaster := getCaster(s, fromElemType, toElemType)
		if elemCaster == nil {
			return nil
		}
		length := min(fromType.Len(), toType.Len())
		fromElemSize := fromElemType.Size()
		toElemSize := toElemType.Size()
		return func(fromAddr, toAddr unsafe.Pointer) error {
			for i := 0; i < length; i++ {
				if err := elemCaster(offset(fromAddr, i, fromElemSize), offset(toAddr, i, toElemSize)); err != nil {
					return err
				}
			}
			return nil
		}
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Pointer:
		return getAddressingPointerCaster(s, fromType, toType)
	case reflect.Slice:
		fromElemType := fromType.Elem()
		toElemType := toType.Elem()
		if isMemSame(fromElemType, toElemType) {
			fromElemSize := fromElemType.Size()
			toSize := toType.Size()
			return func(fromAddr, toAddr unsafe.Pointer) error {
				from := *(*slice)(fromAddr)
				memCopy(toAddr, from.data, min(uintptr(from.len)*fromElemSize, toSize))
				return nil
			}
		}
		elemCaster := getCaster(s, fromElemType, toElemType)
		if elemCaster == nil {
			return nil
		}
		fromElemSize := fromElemType.Size()
		toElemSize := toElemType.Size()
		toLen := toElemType.Len()
		return func(fromAddr, toAddr unsafe.Pointer) error {
			from := *(*slice)(fromAddr)
			length := min(from.len, toLen)
			for i := 0; i < length; i++ {
				if err := elemCaster(offset(from.data, i, fromElemSize), offset(toAddr, i, toElemSize)); err != nil {
					return err
				}
			}
			return nil
		}
	case reflect.String:
		return getFromStringAsSliceCaster(s, toType)
	default:
		return getStringAsBridgeCaster(s, fromType, toType)
	}
}
