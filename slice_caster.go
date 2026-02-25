// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getSliceCaster(s *Scope, fromType, toType reflect.Type) (castFunc, uint8) {
	switch fromType.Kind() {
	case reflect.Array:
		fromElemType := fromType.Elem()
		toElemType := toType.Elem()
		length := fromType.Len()
		if isRefAble(s, fromElemType, toElemType) {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				to := (*slice)(toAddr)
				to.data = fromAddr
				to.len = length
				to.cap = length
				return nil
			}, flagHasRef
		}
		elemCaster, hasRef := getCaster(s, fromElemType, toElemType)
		if elemCaster == nil {
			return nil, 0
		}
		fromElemSize := fromElemType.Size()
		toElemSize := toElemType.Size()
		return func(fromAddr, toAddr unsafe.Pointer) error {
			toPtr := (*slice)(toAddr)
			*toPtr = makeSlice(toElemType, length, length)
			for i := 0; i < length; i++ {
				if err := elemCaster(offset(fromAddr, i, fromElemSize), offset(toPtr.data, i, toElemSize)); err != nil {
					*toPtr = slice{}
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
		fromElemType, toElemType := fromType.Elem(), toType.Elem()
		elemCaster, _ := getCaster(s, fromElemType, toElemType)
		if elemCaster == nil {
			return nil, 0
		}
		fromElemSize := fromElemType.Size()
		toElemSize := toElemType.Size()
		return func(fromAddr, toAddr unsafe.Pointer) error {
			from := *(*slice)(fromAddr)
			if from.data == nil {
				return nil
			}
			toPtr := (*slice)(toAddr)
			*toPtr = makeSlice(toElemType, from.len, from.cap)
			for i := 0; i < from.len; i++ {
				if err := elemCaster(offset(from.data, i, fromElemSize), offset(toPtr.data, i, toElemSize)); err != nil {
					*toPtr = slice{}
					return err
				}
			}
			return nil
		}, 0
	case reflect.String:
		switch toType.Elem().Kind() {
		case reflect.Int32:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*[]rune)(toAddr) = []rune(*(*string)(fromAddr))
				return nil
			}, 0
		case reflect.Uint8:
			if s.deepCopy || s.disableZeroCopy {
				return func(fromAddr, toAddr unsafe.Pointer) error {
					// 在任何go版本，这样string转[]byte都会拷贝
					*(*[]byte)(toAddr) = []byte(*(*string)(fromAddr))
					return nil
				}, 0
			}
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*[]byte)(toAddr) = toBytes(*(*string)(fromAddr))
				return nil
			}, 0
		default:
			return nil, 0
		}
	default:
		return nil, 0
	}
}
