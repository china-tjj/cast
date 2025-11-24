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
		fields, _, _ := getAllFields(s, toType)
		if len(fields) == 0 {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				return nil
			}, false
		}
		type medaData struct {
			// struct field 或 json tag 对应的 map key
			key1        unsafe.Pointer
			key2        unsafe.Pointer
			field       *structField
			fieldHasRef bool
			fieldCaster castFunc
		}
		data := make([]medaData, len(fields))
		for i, field := range fields {
			data[i].field = field
			data[i].fieldCaster, data[i].fieldHasRef = getCaster(s, fromElemType, field.typ)
			var strKey1, strKey2 string
			if field.jsonTag != "" && field.name != "" {
				if field.jsonTag != field.name {
					strKey1, strKey2 = field.jsonTag, field.name
				} else {
					strKey1 = field.jsonTag
				}
			} else if field.jsonTag != "" {
				strKey1 = field.jsonTag
			} else if field.name != "" {
				strKey1 = field.name
			}
			if strKey1 != "" {
				key := newObject(fromKeyType)
				if keyCaster(unsafe.Pointer(&strKey1), key) == nil {
					data[i].key1 = key
				}
			}
			if strKey2 != "" {
				key := newObject(fromKeyType)
				if keyCaster(unsafe.Pointer(&strKey2), key) == nil {
					data[i].key2 = key
				}
			}
			if data[i].key1 == nil && data[i].key2 == nil {
				return nil, false
			}
		}
		fromMapHelper := getMapHelper(fromType)
		zeroPtr := getZeroPtr(toType)
		return func(fromAddr, toAddr unsafe.Pointer) error {
			from := *(*map[any]any)(fromAddr)
			for i := range data {
				if data[i].key1 != nil {
					if v, ok := fromMapHelper.Load(from, data[i].key1, data[i].fieldHasRef); ok {
						if data[i].fieldCaster == nil {
							typedmemmove(typePtr(toType), toAddr, zeroPtr)
							return invalidCastErr(s, fromElemType, data[i].field.typ)
						}
						if err := data[i].fieldCaster(v, data[i].field.getAddr(toAddr, true)); err != nil {
							typedmemmove(typePtr(toType), toAddr, zeroPtr)
							return err
						}
						continue
					}
				}
				if data[i].key2 != nil {
					if v, ok := fromMapHelper.Load(from, data[i].key2, data[i].fieldHasRef); ok {
						if data[i].fieldCaster == nil {
							typedmemmove(typePtr(toType), toAddr, zeroPtr)
							return invalidCastErr(s, fromElemType, data[i].field.typ)
						}
						if err := data[i].fieldCaster(v, data[i].field.getAddr(toAddr, true)); err != nil {
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
		toFields, _, _ := getAllFields(s, toType)
		if len(toFields) == 0 {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				return nil
			}, false
		}
		_, fromFieldMap, fromJsonTagMap := getAllFields(s, fromType)
		type medaData struct {
			fromField     *structField
			toField       *structField
			caster        castFunc
			fromIsNilable bool
		}
		data := make([]medaData, 0, len(toFields))
		getMappedField := func(field *structField) *structField {
			if field.jsonTag != "" {
				if mappedField, ok := fromJsonTagMap[field.jsonTag]; ok {
					return mappedField
				}
			}
			if field.name != "" {
				if mappedField, ok := fromFieldMap[field.name]; ok {
					return mappedField
				}
			}
			return nil
		}
		hasRef := false
		for _, toField := range toFields {
			fromField := getMappedField(toField)
			if fromField == nil {
				continue
			}
			caster, fHasRef := getCaster(s, fromField.typ, toField.typ)
			if caster == nil {
				return nil, false
			}
			data = append(data, medaData{
				fromField:     fromField,
				toField:       toField,
				caster:        caster,
				fromIsNilable: isNilableType(fromField.typ),
			})
			hasRef = hasRef || fHasRef
		}
		if len(data) == 0 {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				return nil
			}, false
		}
		zeroPtr := getZeroPtr(toType)
		return func(fromAddr, toAddr unsafe.Pointer) error {
			for i := range data {
				fromFieldAddr := data[i].fromField.getAddr(fromAddr, false)
				if fromFieldAddr == nil {
					if s.strictNilCheck && !data[i].fromIsNilable {
						typedmemmove(typePtr(toType), toAddr, zeroPtr)
						return NilPtrErr
					}
					continue
				}
				if err := data[i].caster(fromFieldAddr, data[i].toField.getAddr(toAddr, true)); err != nil {
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
