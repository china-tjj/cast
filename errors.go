// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
)

type strErr string

func (e strErr) Error() string {
	return string(e)
}

const NilToTypeErr = strErr("to type is <nil>")
const NilPtrErr = strErr("can't address nil pointer")
const NilStringerErr = strErr("stringer is nil")

func invalidCastErr(s *Scope, fromType, toType reflect.Type) error {
	if s.deepCopy && fromType == toType {
		return strErr("invalid deep copy: can't deep copy type <" + getTypeString(fromType) + ">")
	}
	return strErr("invalid cast: can't cast type <" + getTypeString(fromType) + "> to <" + getTypeString(toType) + ">")
}

func getTypeString(typ reflect.Type) string {
	if typ == nil {
		return "nil"
	}
	return typ.String()
}

func requiredFieldNotMatchErr(toType reflect.Type, fieldName string) error {
	return strErr("required field <" + toType.String() + "." + fieldName + "> not match")
}
