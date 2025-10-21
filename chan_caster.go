// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getChanCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	if fromType == nil {
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*chan any)(toAddr) = nil
			return nil
		}, false
	}
	switch fromType.Kind() {
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Pointer:
		return getAddressingPointerCaster(s, fromType, toType)
	default:
		return nil, false
	}
}
