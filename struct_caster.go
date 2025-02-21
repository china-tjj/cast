// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getStructCaster(fromType, toType reflect.Type) castFunc {
	switch fromType.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Array, reflect.Slice, reflect.String:
		if toType.NumField() != 1 {
			return nil
		}
		caster := getCaster(fromType, toType.Field(0).Type)
		if caster == nil {
			return nil
		}
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			caster(fromAddr, toAddr)
			return true
		}
	case reflect.Interface:
		return getUnwrapInterfaceCaster(fromType, toType)
	case reflect.Map:
		fromKeyType := fromType.Key()
		fromKeyCaster := getCaster(typeFor[string](), fromKeyType)
		if fromKeyCaster == nil {
			return nil
		}
		fromElemType := fromType.Elem()
		n := toType.NumField()
		// struct field 对应的 map key
		fieldKeys := make([]reflect.Value, n)
		fieldOffsets := make([]uintptr, n)
		fieldCasters := make([]castFunc, n)
		for i := 0; i < n; i++ {
			field := toType.Field(i)
			caster := newCaster(fromElemType, field.Type)
			if caster == nil {
				continue
			}
			fieldKey := reflect.New(fromKeyType).Elem()
			fieldKeyAddr := getValueAddr(fieldKey)
			jsonTag, ok := field.Tag.Lookup("json")
			if ok {
				ok = fromKeyCaster(unsafe.Pointer(&jsonTag), fieldKeyAddr)
			} else {
				ok = fromKeyCaster(unsafe.Pointer(&field.Name), fieldKeyAddr)
			}
			if !ok {
				continue
			}
			fieldKeys[i] = fieldKey
			fieldOffsets[i] = field.Offset
			fieldCasters[i] = caster
		}
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			from := reflect.NewAt(fromType, fromAddr).Elem()
			for i := 0; i < n; i++ {
				if fieldCasters[i] == nil {
					continue
				}
				v := from.MapIndex(fieldKeys[i])
				if !v.IsValid() {
					continue
				}
				fieldCasters[i](getValueAddr(v), unsafe.Pointer(uintptr(toAddr)+fieldOffsets[i]))
			}
			return true
		}
	case reflect.Pointer:
		return getAddressingPointerCaster(fromType, toType)
	case reflect.Struct:
		if isMemSame(fromType, toType) {
			size := int(fromType.Size())
			return func(fromAddr, toAddr unsafe.Pointer) bool {
				memCopy(toAddr, fromAddr, size)
				return true
			}
		}
		nFrom := fromType.NumField()
		fromFieldMap := make(map[string]reflect.StructField, nFrom)
		fromJsonMap := make(map[string]reflect.StructField, nFrom)
		for i := 0; i < nFrom; i++ {
			field := fromType.Field(i)
			fromFieldMap[field.Name] = field
			if jsonTag, ok := field.Tag.Lookup("json"); ok {
				fromJsonMap[jsonTag] = field
			}
		}
		nTo := toType.NumField()
		fromOffsets := make([]uintptr, nTo)
		toOffsets := make([]uintptr, nTo)
		casters := make([]castFunc, nTo)
		casterCnt := 0
		for i := 0; i < nTo; i++ {
			toField := toType.Field(i)
			var fromField reflect.StructField
			var ok bool
			if jsonTag, hasTag := toField.Tag.Lookup("json"); hasTag {
				fromField, ok = fromJsonMap[jsonTag]
			}
			if !ok {
				fromField, ok = fromFieldMap[toField.Name]
			}
			if !ok {
				continue
			}
			fromOffsets[i] = fromField.Offset
			toOffsets[i] = toField.Offset
			casters[i] = getCaster(fromField.Type, toField.Type)
			if casters[i] == nil {
				continue
			}
			casterCnt++
		}
		if casterCnt == 0 && nTo > 0 {
			return nil
		}
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			for i := 0; i < nTo; i++ {
				if casters[i] == nil {
					continue
				}
				casters[i](unsafe.Pointer(uintptr(fromAddr)+fromOffsets[i]), unsafe.Pointer(uintptr(toAddr)+toOffsets[i]))
			}
			return true
		}
	default:
		return nil
	}
}
