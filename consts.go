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

var nilPtrErr = errors.New("can't address nil pointer")
var nilStringerErr = errors.New("stringer is nil")

func invalidCastErr(fromType, toType reflect.Type) error {
	return errors.New("invalid cast: can't cast type " + fromType.String() + " to " + toType.String())
}

var (
	boolType     = typeFor[bool]()
	intType      = typeFor[int]()
	int8Type     = typeFor[int8]()
	int16Type    = typeFor[int16]()
	int32Type    = typeFor[int32]()
	int64Type    = typeFor[int64]()
	uintType     = typeFor[uint]()
	uint8Type    = typeFor[uint8]()
	uint16Type   = typeFor[uint16]()
	uint32Type   = typeFor[uint32]()
	uint64Type   = typeFor[uint64]()
	uintptrType  = typeFor[uintptr]()
	float32Type  = typeFor[float32]()
	float64Type  = typeFor[float64]()
	stringType   = typeFor[string]()
	bytesType    = typeFor[[]byte]()
	runesType    = typeFor[[]rune]()
	jsonMapType  = typeFor[map[string]any]()
	jsonListType = typeFor[[]any]()
	stringerType = typeFor[fmt.Stringer]()
	anyType      = typeFor[any]()
)
