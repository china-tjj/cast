// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getSliceCaster(fromType, toType reflect.Type) castFunc {
	switch fromType.Kind() {
	case reflect.Array:
		fromElemType := fromType.Elem()
		toElemType := toType.Elem()
		length := fromType.Len()
		if isMemSame(fromElemType, toElemType) {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				to := (*slice)(toAddr)
				to.data = fromAddr
				to.len = length
				to.cap = length
				return nil
			}
		}
		elemCaster := getCaster(fromElemType, toElemType)
		if elemCaster == nil {
			return nil
		}
		fromElemSize := fromElemType.Size()
		toElemSize := toElemType.Size()
		return func(fromAddr, toAddr unsafe.Pointer) error {
			toPtr := (*slice)(toAddr)
			*toPtr = makeSlice(toElemSize, length, length)
			for i := 0; i < length; i++ {
				if err := elemCaster(offset(fromAddr, i, fromElemSize), offset(toPtr.data, i, toElemSize)); err != nil {
					return err
				}
			}
			return nil
		}
	case reflect.Interface:
		return getUnpackInterfaceCaster(fromType, toType)
	case reflect.Pointer:
		return getAddressingPointerCaster(fromType, toType)
	case reflect.Slice:
		fromElemType := fromType.Elem()
		toElemType := toType.Elem()
		if isMemSame(fromElemType, toElemType) {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*slice)(toAddr) = *(*slice)(fromAddr)
				return nil
			}
		}
		elemCaster := getCaster(fromElemType, toElemType)
		if elemCaster == nil {
			return nil
		}
		fromElemSize := fromElemType.Size()
		toElemSize := toElemType.Size()
		return func(fromAddr, toAddr unsafe.Pointer) error {
			from := *(*slice)(fromAddr)
			toPtr := (*slice)(toAddr)
			*toPtr = makeSlice(toElemSize, from.len, from.cap)
			for i := 0; i < from.len; i++ {
				if err := elemCaster(offset(from.data, i, fromElemSize), offset(toPtr.data, i, toElemSize)); err != nil {
					return err
				}
			}
			return nil
		}
	case reflect.String:
		return getFromStringAsSliceCaster(toType)
	default:
		return nil
	}
}
