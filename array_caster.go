// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getArrayCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
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
			return func(fromAddr, toAddr unsafe.Pointer) error {
				typedmemmove(arrayTypePtr, toAddr, fromAddr)
				return nil
			}, false
		}
		elemCaster, hasRef := getCaster(s, fromElemType, toElemType)
		if elemCaster == nil {
			return nil, false
		}
		length := min(fromType.Len(), toType.Len())
		fromElemSize := fromElemType.Size()
		toElemSize := toElemType.Size()
		zeroPtr := getZeroPtr(toType)
		return func(fromAddr, toAddr unsafe.Pointer) error {
			for i := 0; i < length; i++ {
				if err := elemCaster(offset(fromAddr, i, fromElemSize), offset(toAddr, i, toElemSize)); err != nil {
					typedmemmove(typePtr(toType), toAddr, zeroPtr)
					return err
				}
			}
			return nil
		}, hasRef
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
			return func(fromAddr, toAddr unsafe.Pointer) error {
				from := *(*slice)(fromAddr)
				typedslicecopy(toElemTypePtr, toAddr, toLen, from.data, from.len)
				return nil
			}, false
		}
		elemCaster, _ := getCaster(s, fromElemType, toElemType)
		if elemCaster == nil {
			return nil, false
		}
		fromElemSize := fromElemType.Size()
		toElemSize := toElemType.Size()
		toLen := toElemType.Len()
		zeroPtr := getZeroPtr(toType)
		return func(fromAddr, toAddr unsafe.Pointer) error {
			from := *(*slice)(fromAddr)
			length := min(from.len, toLen)
			for i := 0; i < length; i++ {
				if err := elemCaster(offset(from.data, i, fromElemSize), offset(toAddr, i, toElemSize)); err != nil {
					typedmemmove(typePtr(toType), toAddr, zeroPtr)
					return err
				}
			}
			return nil
		}, false
	case reflect.String:
		return getFromStringAsSliceCaster(s, toType)
	default:
		return getStringAsBridgeCaster(s, fromType, toType)
	}
}
