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
		type medaData struct {
			jsonTagKey   unsafe.Pointer // json tag 对应的 map key
			fieldNameKey unsafe.Pointer // struct field 对应的 map key
			fieldHasRef  bool
			fieldCaster  castFunc
			fieldOffset  uintptr
			fieldType    reflect.Type
		}
		data := make([]medaData, n)
		for i := 0; i < n; i++ {
			field := toType.Field(i)
			if !s.castUnexported && !field.IsExported() {
				continue
			}
			data[i].fieldOffset = field.Offset
			data[i].fieldType = field.Type
			data[i].fieldCaster, data[i].fieldHasRef = getCaster(s, fromElemType, field.Type)
			if jsonTag, ok := field.Tag.Lookup("json"); ok && jsonTag != field.Name {
				jsonTagKey := newObject(fromKeyType)
				if keyCaster(unsafe.Pointer(&jsonTag), jsonTagKey) == nil {
					data[i].jsonTagKey = jsonTagKey
				}
			}
			fieldNameKey := newObject(fromKeyType)
			if keyCaster(unsafe.Pointer(&field.Name), fieldNameKey) == nil {
				data[i].fieldNameKey = fieldNameKey
			}
			if data[i].jsonTagKey == nil && data[i].fieldNameKey == nil {
				return nil, false
			}
		}
		fromMapHelper := getMapHelper(fromType)
		zeroPtr := getZeroPtr(toType)
		return func(fromAddr, toAddr unsafe.Pointer) error {
			from := *(*map[any]any)(fromAddr)
			for i := 0; i < n; i++ {
				if data[i].jsonTagKey != nil {
					if v, ok := fromMapHelper.Load(from, data[i].jsonTagKey, data[i].fieldHasRef); ok {
						if data[i].fieldCaster == nil {
							typedmemmove(typePtr(toType), toAddr, zeroPtr)
							return invalidCastErr(s, fromElemType, data[i].fieldType)
						}
						if err := data[i].fieldCaster(v, unsafe.Add(toAddr, data[i].fieldOffset)); err != nil {
							typedmemmove(typePtr(toType), toAddr, zeroPtr)
							return err
						}
						continue
					}
				}
				if data[i].fieldNameKey != nil {
					if v, ok := fromMapHelper.Load(from, data[i].fieldNameKey, data[i].fieldHasRef); ok {
						if data[i].fieldCaster == nil {
							typedmemmove(typePtr(toType), toAddr, zeroPtr)
							return invalidCastErr(s, fromElemType, data[i].fieldType)
						}
						if err := data[i].fieldCaster(v, unsafe.Add(toAddr, data[i].fieldOffset)); err != nil {
							typedmemmove(typePtr(toType), toAddr, zeroPtr)
							return err
						}
						continue
					}
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
		type medaData struct {
			fromOffset uintptr
			toOffset   uintptr
			caster     castFunc
		}
		data := make([]medaData, nTo)
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
		needCast := false
		for i := 0; i < nTo; i++ {
			toField := toType.Field(i)
			if !s.castUnexported && !toField.IsExported() {
				continue
			}
			fromField := getMappedField(&toField)
			if fromField == nil {
				continue
			}
			data[i].fromOffset = fromField.Offset
			data[i].toOffset = toField.Offset
			var fHasRef bool
			data[i].caster, fHasRef = getCaster(s, fromField.Type, toField.Type)
			if data[i].caster == nil {
				return nil, false
			}
			needCast = true
			hasRef = hasRef || fHasRef
		}
		if !needCast {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				return nil
			}, false
		}
		zeroPtr := getZeroPtr(toType)
		return func(fromAddr, toAddr unsafe.Pointer) error {
			for i := 0; i < nTo; i++ {
				if data[i].caster == nil {
					continue
				}
				if err := data[i].caster(unsafe.Add(fromAddr, data[i].fromOffset), unsafe.Add(toAddr, data[i].toOffset)); err != nil {
					typedmemmove(typePtr(toType), toAddr, zeroPtr)
					return err
				}
			}
			return nil
		}, hasRef
	default:
		return nil, false
	}
}
