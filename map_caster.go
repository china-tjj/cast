// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"sync/atomic"
	"unsafe"
)

func getMapCaster(s *Scope, fromType, toType reflect.Type) (castFunc, uint8) {
	switch fromType.Kind() {
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Map:
		fromKeyType := fromType.Key()
		toKeyType := toType.Key()
		keyCaster, keyFlag := getCaster(s, fromKeyType, toKeyType)
		if keyCaster == nil {
			return nil, 0
		}
		fromElemType := fromType.Elem()
		toElemType := toType.Elem()
		elemCaster, elemFlag := getCaster(s, fromElemType, toElemType)
		if elemCaster == nil {
			return nil, 0
		}
		fromMapHelper := newMapHelper(fromType)
		toMapHelper := newMapHelper(toType)
		// 小优化，减少堆内存分配
		globalKeyBuffer := newObject(toKeyType)
		globalValueBuffer := newObject(toElemType)
		keyZeroPtr := getZeroPtr(toKeyType)
		valueZeroPtr := getZeroPtr(toElemType)
		var mu atomic.Uint32
		return func(fromAddr, toAddr unsafe.Pointer) error {
			from := *(*map[any]any)(fromAddr)
			if from == nil {
				*(*map[any]any)(toAddr) = nil
				return nil
			}
			to := toMapHelper.Make(len(from))
			*(*map[any]any)(toAddr) = to
			if len(from) == 0 {
				return nil
			}
			var err error
			var toK, toV unsafe.Pointer
			if mu.CompareAndSwap(0, 1) {
				defer mu.Store(0)
				toK = globalKeyBuffer
				toV = globalValueBuffer
			} else {
				toK = newObject(toKeyType)
				toV = newObject(toElemType)
			}
			fromMapHelper.Range(from, func(key unsafe.Pointer, value unsafe.Pointer) bool {
				if isHasRef(keyFlag) {
					key = copyObject(fromKeyType, key)
				}
				typedMemMove(typePtr(toKeyType), toK, keyZeroPtr)
				if err = keyCaster(key, toK); err != nil {
					*(*map[any]any)(toAddr) = nil
					return false
				}
				if isHasRef(elemFlag) {
					value = copyObject(fromElemType, value)
				}
				typedMemMove(typePtr(toElemType), toV, valueZeroPtr)
				if err = elemCaster(value, toV); err != nil {
					*(*map[any]any)(toAddr) = nil
					return false
				}
				toMapHelper.Store(to, toK, toV)
				return true
			})
			return err
		}, 0
	case reflect.Pointer:
		return getAddressingPointerCaster(s, fromType, toType)
	case reflect.Struct:
		toKeyType := toType.Key()
		keyCaster, _ := getCaster(s, stringType, toKeyType)
		if keyCaster == nil {
			return nil, 0
		}
		toElemType := toType.Elem()
		type metaField struct {
			structField
			key    unsafe.Pointer
			caster castFunc
		}
		fields := getAllFields(s, fromType)
		toMapHelper := newMapHelper(toType)
		if len(fields.flattened) == 0 {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*map[any]any)(toAddr) = toMapHelper.Make(0)
				return nil
			}, 0
		}
		var flag uint8
		keyIsStr := toKeyType.Kind() == reflect.String
		keyIsRefType := isRefType(toKeyType)
		metaFields := make([]metaField, 0, len(fields.flattened))
		for _, field := range fields.flattened {
			caster, fFlag := getCaster(s, field.typ, toElemType)
			if caster == nil {
				return nil, 0
			}
			var fieldKey unsafe.Pointer
			if keyIsStr {
				fieldKey = unsafe.Pointer(&field.name)
			} else if !keyIsRefType {
				fieldKey = newObject(toKeyType)
				if err := keyCaster(unsafe.Pointer(&field.name), fieldKey); err != nil {
					return nil, 0
				}
			}
			metaFields = append(metaFields, metaField{
				structField: *field,
				key:         fieldKey,
				caster:      caster,
			})
			if isHasRef(fFlag) {
				flag |= flagHasRef
			}
		}
		// 小优化，减少堆内存分配
		globalValueBuffer := newObject(toElemType)
		var mu atomic.Uint32
		valueZeroPtr := getZeroPtr(toElemType)
		return func(fromAddr, toAddr unsafe.Pointer) error {
			to := toMapHelper.Make(len(fields.flattened))
			*(*map[any]any)(toAddr) = to
			var v unsafe.Pointer
			if mu.CompareAndSwap(0, 1) {
				defer mu.Store(0)
				v = globalValueBuffer
			} else {
				v = newObject(toElemType)
			}
			for i := range metaFields {
				field := &metaFields[i]
				var k unsafe.Pointer
				if keyIsRefType {
					// key是引用类型时，如指针，保证每次生成新的key
					k = newObject(toKeyType)
					name := field.name // 浅拷贝一下，避免被hack
					if err := keyCaster(unsafe.Pointer(&name), k); err != nil {
						*(*map[any]any)(toAddr) = nil
						return err
					}
				} else {
					k = field.key
				}
				fromFieldAddr := field.getAddr(fromAddr, false)
				if fromFieldAddr == nil {
					continue
				}
				typedMemMove(typePtr(toElemType), v, valueZeroPtr)
				if err := field.caster(fromFieldAddr, v); err != nil {
					*(*map[any]any)(toAddr) = nil
					return err
				}
				toMapHelper.Store(to, k, v)
			}
			return nil
		}, flag
	default:
		return nil, 0
	}
}
