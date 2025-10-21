// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getStructCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	if fromType == nil {
		return nil, false
	}
	switch fromType.Kind() {
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Map:
		fromKeyType := fromType.Key()
		keyCaster, _ := getCaster(s, stringType, fromKeyType)
		if keyCaster == nil {
			return nil, false
		}
		fromElemType := fromType.Elem()
		n := toType.NumField()
		if n == 0 {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				return nil
			}, false
		}
		// json tag 对应的 map key
		jsonTagKeys := make([]unsafe.Pointer, n)
		// struct field 对应的 map key
		fieldKeys := make([]unsafe.Pointer, n)
		fieldOffsets := make([]uintptr, n)
		fieldCasters := make([]castFunc, n)
		fieldHasRefs := make([]bool, n)
		fieldTypes := make([]reflect.Type, n)
		for i := 0; i < n; i++ {
			field := toType.Field(i)
			if !s.castUnexported && !field.IsExported() {
				continue
			}
			fieldOffsets[i] = field.Offset
			fieldTypes[i] = field.Type
			fieldCasters[i], fieldHasRefs[i] = getCaster(s, fromElemType, field.Type)
			if jsonTag, ok := getDiffJsonTag(&field); ok {
				jsonTagKey := newObject(fromKeyType)
				if keyCaster(unsafe.Pointer(&jsonTag), jsonTagKey) == nil {
					jsonTagKeys[i] = jsonTagKey
				}
			}
			fieldKey := newObject(fromKeyType)
			if keyCaster(unsafe.Pointer(&field.Name), fieldKey) == nil {
				fieldKeys[i] = fieldKey
			}
			if jsonTagKeys[i] == nil && fieldKeys[i] == nil {
				return nil, false
			}
		}
		fromMapHelper := getMapHelper(fromType)
		return func(fromAddr, toAddr unsafe.Pointer) error {
			from := *(*map[any]any)(fromAddr)
			for i := 0; i < n; i++ {
				var err1, err2 error
				if jsonTagKeys[i] != nil {
					if v, ok := fromMapHelper.Load(from, jsonTagKeys[i], fieldHasRefs[i]); ok {
						if fieldCasters[i] == nil {
							return invalidCastErr(s, fromElemType, fieldTypes[i])
						}
						if err1 = fieldCasters[i](v, unsafe.Add(toAddr, fieldOffsets[i])); err1 == nil {
							continue
						}
					}
				}
				if fieldKeys[i] != nil {
					if v, ok := fromMapHelper.Load(from, fieldKeys[i], fieldHasRefs[i]); ok {
						if fieldCasters[i] == nil {
							return invalidCastErr(s, fromElemType, fieldTypes[i])
						}
						if err2 = fieldCasters[i](v, unsafe.Add(toAddr, fieldOffsets[i])); err2 == nil {
							continue
						}
					}
				}
				if err1 != nil {
					return err1
				}
				if err2 != nil {
					return err2
				}
			}
			return nil
		}, false
	case reflect.Pointer:
		return getAddressingPointerCaster(s, fromType, toType)
	case reflect.Struct:
		nTo := toType.NumField()
		if nTo == 0 {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				return nil
			}, false
		}

		nFrom := fromType.NumField()
		fromJsonTagMap := make(map[string]reflect.StructField, nFrom)
		fromFieldMap := make(map[string]reflect.StructField, nFrom)
		for i := 0; i < nFrom; i++ {
			field := fromType.Field(i)
			if !s.castUnexported && !field.IsExported() {
				continue
			}
			if jsonTag, ok := field.Tag.Lookup("json"); ok {
				fromJsonTagMap[jsonTag] = field
			}
			fromFieldMap[field.Name] = field
		}
		fromOffsets := make([]uintptr, nTo)
		toOffsets := make([]uintptr, nTo)
		casters := make([]castFunc, nTo)
		casterCnt := 0
		getMappedField := func(field *reflect.StructField) *reflect.StructField {
			if jsonTag, ok := field.Tag.Lookup("json"); ok {
				if mappedField, ok := fromJsonTagMap[jsonTag]; ok {
					return &mappedField
				}
			}
			if mappedField, ok := fromFieldMap[field.Name]; ok {
				return &mappedField
			}
			return nil
		}
		hasRef := false
		for i := 0; i < nTo; i++ {
			toField := toType.Field(i)
			if !s.castUnexported && !toField.IsExported() {
				continue
			}
			fromField := getMappedField(&toField)
			if fromField == nil {
				continue
			}
			fromOffsets[i] = fromField.Offset
			toOffsets[i] = toField.Offset
			var fHasRef bool
			casters[i], fHasRef = getCaster(s, fromField.Type, toField.Type)
			if casters[i] == nil {
				return nil, false
			}
			casterCnt++
			hasRef = hasRef || fHasRef
		}
		if casterCnt == 0 && nFrom > 0 {
			return nil, false
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			for i := 0; i < nTo; i++ {
				if casters[i] == nil {
					continue
				}
				if err := casters[i](unsafe.Add(fromAddr, fromOffsets[i]), unsafe.Add(toAddr, toOffsets[i])); err != nil {
					return err
				}
			}
			return nil
		}, hasRef
	default:
		return nil, false
	}
}
