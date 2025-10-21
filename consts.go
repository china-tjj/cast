// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"errors"
	"fmt"
	"reflect"
)

var nilToTypeErr = errors.New("to type is <nil>")
var nilPtrErr = errors.New("can't address nil pointer")
var nilStringerErr = errors.New("stringer is nil")

func invalidCastErr(s *Scope, fromType, toType reflect.Type) error {
	if s.deepCopy && fromType == toType {
		return errors.New("invalid deep copy: can't deep copy type 「" + fromType.String() + "」")
	}
	return errors.New("invalid cast: can't cast type 「" + fromType.String() + "」 to 「" + toType.String() + "」")
}

var (
	stringType   = typeFor[string]()
	bytesType    = typeFor[[]byte]()
	runesType    = typeFor[[]rune]()
	stringerType = typeFor[fmt.Stringer]()
)
