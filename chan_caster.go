// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
)

func getChanCaster(s *Scope, fromType, toType reflect.Type) castFunc {
	switch fromType.Kind() {
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Pointer:
		return getAddressingPointerCaster(s, fromType, toType)
	default:
		return nil
	}
}
