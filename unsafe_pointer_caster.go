// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getUnsafePointerCaster(fromType, toType reflect.Type) castFunc {
	switch fromType.Kind() {
	case reflect.Uintptr:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*unsafe.Pointer)(toAddr) = unsafe.Pointer(*(*uintptr)(fromAddr))
			return true
		}
	case reflect.Interface:
		return getUnwrapInterfaceCaster(fromType, toType)
	case reflect.Pointer, reflect.UnsafePointer:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*unsafe.Pointer)(toAddr) = *(*unsafe.Pointer)(fromAddr)
			return true
		}
	default:
		return nil
	}
}
