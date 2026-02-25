// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getArrayCaster(s *Scope, fromType, toType reflect.Type) (castFunc, uint8) {
	switch fromType.Kind() {
	case reflect.Array:
		fromElemType := fromType.Elem()
		toElemType := toType.Elem()
		if isRefAble(s, fromElemType, toElemType) {
			var arrayTypePtr unsafe.Pointer
			if fromType.Len() <= toType.Len() {
				arrayTypePtr = typePtr(fromType)
			} else {
				arrayTypePtr = typePtr(toType)
			}
			return func(fromAddr, toAddr unsafe.Pointer) error {
				typedMemMove(arrayTypePtr, toAddr, fromAddr)
				return nil
			}, 0
		}
		elemCaster, flag := getCaster(s, fromElemType, toElemType)
		if elemCaster == nil {
			return nil, 0
		}
		length := min(fromType.Len(), toType.Len())
		fromElemSize := fromElemType.Size()
		toElemSize := toElemType.Size()
		zeroPtr := getZeroPtr(toType)
		return func(fromAddr, toAddr unsafe.Pointer) error {
			for i := 0; i < length; i++ {
				if err := elemCaster(offset(fromAddr, i, fromElemSize), offset(toAddr, i, toElemSize)); err != nil {
					typedMemMove(typePtr(toType), toAddr, zeroPtr)
					return err
				}
			}
			return nil
		}, flag
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
				fromPtr := (*slice)(fromAddr)
				typedSliceCopy(toElemTypePtr, toAddr, toLen, fromPtr.data, fromPtr.len)
				return nil
			}, 0
		}
		elemCaster, _ := getCaster(s, fromElemType, toElemType)
		if elemCaster == nil {
			return nil, 0
		}
		fromElemSize := fromElemType.Size()
		toElemSize := toElemType.Size()
		toLen := toType.Len()
		zeroPtr := getZeroPtr(toType)
		return func(fromAddr, toAddr unsafe.Pointer) error {
			from := *(*slice)(fromAddr)
			length := min(from.len, toLen)
			for i := 0; i < length; i++ {
				if err := elemCaster(offset(from.data, i, fromElemSize), offset(toAddr, i, toElemSize)); err != nil {
					typedMemMove(typePtr(toType), toAddr, zeroPtr)
					return err
				}
			}
			return nil
		}, 0
	case reflect.String:
		switch toType.Elem().Kind() {
		case reflect.Int32:
			toLen := toType.Len()
			toElemTypePtr := typePtr(toType.Elem())
			return func(fromAddr, toAddr unsafe.Pointer) error {
				fromRunes := []rune(*(*string)(fromAddr))
				fromRunesPtr := (*slice)(unsafe.Pointer(&fromRunes))
				typedSliceCopy(toElemTypePtr, toAddr, toLen, fromRunesPtr.data, fromRunesPtr.len)
				return nil
			}, 0
		case reflect.Uint8:
			toLen := toType.Len()
			toElemTypePtr := typePtr(toType.Elem())
			return func(fromAddr, toAddr unsafe.Pointer) error {
				fromPtr := (*str)(fromAddr)
				typedSliceCopy(toElemTypePtr, toAddr, toLen, fromPtr.data, fromPtr.len)
				return nil
			}, 0
		default:
			return nil, 0
		}
	default:
		return nil, 0
	}
}
