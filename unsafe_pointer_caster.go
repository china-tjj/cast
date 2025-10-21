// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getUnsafePointerCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	if fromType == nil {
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*[]any)(toAddr) = nil
			return nil
		}, false
	}
	switch fromType.Kind() {
	case reflect.Uintptr:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*unsafe.Pointer)(toAddr) = unsafe.Pointer(*(*uintptr)(fromAddr))
			return nil
		}, false
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Pointer, reflect.UnsafePointer:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*unsafe.Pointer)(toAddr) = *(*unsafe.Pointer)(fromAddr)
			return nil
		}, false
	default:
		return nil, false
	}
}
