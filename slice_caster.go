// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getSliceCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
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
			}, true
		}
		elemCaster, hasRef := getCaster(s, fromElemType, toElemType)
		if elemCaster == nil {
			return nil, false
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
			return nil, false
		}
		fromElemSize := fromElemType.Size()
		toElemSize := toElemType.Size()
		return func(fromAddr, toAddr unsafe.Pointer) error {
			from := *(*slice)(fromAddr)
			toPtr := (*slice)(toAddr)
			*toPtr = makeSlice(toElemType, from.len, from.cap)
			for i := 0; i < from.len; i++ {
				if err := elemCaster(offset(from.data, i, fromElemSize), offset(toPtr.data, i, toElemSize)); err != nil {
					*toPtr = slice{}
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
