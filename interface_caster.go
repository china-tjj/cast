// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getInterfaceCaster(fromType, toType reflect.Type) castFunc {
	if !fromType.Implements(toType) {
		return nil
	}
	return func(fromAddr, toAddr unsafe.Pointer) bool {
		reflect.NewAt(toType, toAddr).Elem().Set(reflect.NewAt(fromType, fromAddr).Elem())
		return true
	}
}
