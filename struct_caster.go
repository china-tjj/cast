// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getStructCaster(s *Scope, fromType, toType reflect.Type) (castFunc, uint8) {
	switch fromType.Kind() {
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Map:
		fromKeyType := fromType.Key()
		keyCaster, _ := getCaster(s, fromKeyType, stringType)
		if keyCaster == nil {
			return nil, 0
		}
		fromElemType := fromType.Elem()
		// 转换步骤：在线把map[K]V先转为map[string]V，再匹配字段
		type metaField struct {
			structField
			caster castFunc
			flag   uint8
		}
		fields := getAllFields(s, toType)
		if len(fields.flattened) == 0 {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				return nil
			}, 0
		}
		metaFields := make([]metaField, 0, len(fields.flattened))
		for _, field := range fields.flattened {
			caster, flag := getCaster(s, fromElemType, field.typ)
			if caster == nil && field.isRequired {
				return nil, 0
			}
			metaFields = append(metaFields, metaField{
				structField: *field,
				caster:      caster,
				flag:        flag,
			})
		}
		keyIsStr := fromKeyType.Kind() == reflect.String
		fromMapHelper := newMapHelper(fromType)
		zeroPtr := getZeroPtr(toType)
		return func(fromAddr, toAddr unsafe.Pointer) error {
			from := *(*map[any]any)(fromAddr)
			var keyMap map[string]unsafe.Pointer
			if !keyIsStr {
				// 直接拿map里的指针，非并发安全
				keyMap = make(map[string]unsafe.Pointer, len(from))
				fromMapHelper.Range(from, func(key, value unsafe.Pointer) bool {
					var strKey string
					if err := keyCaster(key, noEscape(unsafe.Pointer(&strKey))); err != nil {
						// key转换失败直接跳过，required字段的校验在后面做
						return true
					}
					keyMap[strKey] = value
					return true
				})
			}
			var missField []*metaField
			for i := range metaFields {
				field := &metaFields[i]
				var v unsafe.Pointer
				var ok bool
				if keyIsStr {
					v, ok = fromMapHelper.Load(from, unsafe.Pointer(&field.name))
				} else {
					v, ok = keyMap[field.name]
				}
				if !ok {
					if field.foldedName != "" && field.foldedName != field.name {
						missField = append(missField, field)
						continue
					}
					if field.isRequired {
						typedMemMove(typePtr(toType), toAddr, zeroPtr)
						return requiredFieldNotMatchErr(toType, field.rawName)
					}
					continue
				}
				if field.caster == nil {
					return invalidCastErr(s, fromElemType, field.typ)
				}
				if isHasRef(field.flag) {
					v = copyObject(fromElemType, v)
				}
				if err := field.caster(v, field.getAddr(toAddr, true)); err != nil {
					typedMemMove(typePtr(toType), toAddr, zeroPtr)
					return err
				}
			}
			if len(missField) == 0 {
				return nil
			}

			// 这里key不能排除foldNameStr(k)==k的，因为前面已经排除了field.foldedName==field.name的，比如存在以下情况：
			// field.name="a", field.foldedName="A", k="A", foldNameStr(k)="A"
			foldedKeyMap := make(map[string]unsafe.Pointer, len(from))
			if !keyIsStr {
				for k, v := range keyMap {
					foldedKeyMap[foldNameStr(k)] = v
				}
			} else {
				fromMapHelper.Range(from, func(key, value unsafe.Pointer) bool {
					foldedKeyMap[foldNameStr(*(*string)(key))] = value
					return true
				})
			}
			for _, field := range missField {
				v, ok := foldedKeyMap[field.foldedName]
				if !ok {
					if field.isRequired {
						typedMemMove(typePtr(toType), toAddr, zeroPtr)
						return requiredFieldNotMatchErr(toType, field.rawName)
					}
					continue
				}
				if field.caster == nil {
					return invalidCastErr(s, fromElemType, field.typ)
				}
				if isHasRef(field.flag) {
					v = copyObject(fromElemType, v)
				}
				if err := field.caster(v, field.getAddr(toAddr, true)); err != nil {
					typedMemMove(typePtr(toType), toAddr, zeroPtr)
					return err
				}
			}
			return nil
		}, 0
	case reflect.Pointer:
		return getAddressingPointerCaster(s, fromType, toType)
	case reflect.Struct:
		type metaField struct {
			structField
			fromField     structField
			caster        castFunc
			fromIsNilable bool
		}
		toFields := getAllFields(s, toType)
		if len(toFields.flattened) == 0 {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				return nil
			}, 0
		}
		fromFields := getAllFields(s, fromType)
		var flag uint8
		metaFields := make([]metaField, 0, len(toFields.flattened))
		for _, toField := range toFields.flattened {
			fromField, ok := fromFields.byActualName[toField.name]
			if !ok && toField.foldedName != "" && toField.foldedName != toField.name {
				fromField, ok = fromFields.byFoldedName[toField.foldedName]
			}
			if !ok {
				if toField.isRequired {
					return nil, 0
				}
				continue
			}
			caster, fFlag := getCaster(s, fromField.typ, toField.typ)
			if caster == nil {
				return nil, 0
			}
			metaFields = append(metaFields, metaField{
				structField:   *toField,
				fromField:     *fromField,
				caster:        caster,
				fromIsNilable: isNilableType(fromField.typ),
			})
			if isHasRef(fFlag) {
				flag |= flagHasRef
			}
		}
		if len(metaFields) == 0 {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				return nil
			}, 0
		}
		zeroPtr := getZeroPtr(toType)
		return func(fromAddr, toAddr unsafe.Pointer) error {
			for i := range metaFields {
				field := &metaFields[i]
				fromFieldAddr := field.fromField.getAddr(fromAddr, false)
				if fromFieldAddr == nil {
					if s.strictNilCheck && !field.fromIsNilable {
						typedMemMove(typePtr(toType), toAddr, zeroPtr)
						return NilPtrErr
					}
					continue
				}
				if err := field.caster(fromFieldAddr, field.getAddr(toAddr, true)); err != nil {
					typedMemMove(typePtr(toType), toAddr, zeroPtr)
					return err
				}
			}
			return nil
		}, flag
	default:
		return nil, 0
	}
}
