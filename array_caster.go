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
		if isRefAble(s, fromElemType, toElemType) {
			var arrayTypePtr unsafe.Pointer
			if fromType.Len() <= toElemType.Len() {
				arrayTypePtr = typePtr(fromType)
			} else {
				arrayTypePtr = typePtr(toType)
			}
			return func(s *Scope, fromAddr, toAddr unsafe.Pointer) error {
				typedmemmove(arrayTypePtr, toAddr, fromAddr)
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
		return func(s *Scope, fromAddr, toAddr unsafe.Pointer) error {
			for i := 0; i < length; i++ {
				if err := elemCaster(s, offset(fromAddr, i, fromElemSize), offset(toAddr, i, toElemSize)); err != nil {
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
		if isRefAble(s, fromElemType, toElemType) {
			toElemTypePtr := typePtr(toElemType)
			toLen := toType.Len()
			return func(s *Scope, fromAddr, toAddr unsafe.Pointer) error {
				from := *(*slice)(fromAddr)
				typedslicecopy(toElemTypePtr, toAddr, toLen, from.data, from.len)
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
		return func(s *Scope, fromAddr, toAddr unsafe.Pointer) error {
			from := *(*slice)(fromAddr)
			length := min(from.len, toLen)
			for i := 0; i < length; i++ {
				if err := elemCaster(s, offset(from.data, i, fromElemSize), offset(toAddr, i, toElemSize)); err != nil {
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
