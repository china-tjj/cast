// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getChanCaster(fromType, toType reflect.Type) castFunc {
	switch fromType.Kind() {
	case reflect.Chan:
		if !isMemSame(fromType, toType) {
			return nil
		}
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*unsafe.Pointer)(toAddr) = *(*unsafe.Pointer)(fromAddr)
			return true
		}
	case reflect.Array:
		return getUnpackArrayCaster(fromType, toType)
	case reflect.Interface:
		return getUnpackInterfaceCaster(fromType, toType)
	case reflect.Pointer:
		return getAddressingPointerCaster(fromType, toType)
	case reflect.Slice:
		return getUnpackSliceCaster(fromType, toType)
	case reflect.Struct:
		return getUnpackStructCaster(fromType, toType)
	default:
		return nil
	}
}
