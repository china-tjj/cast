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

func getMapCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	switch fromType.Kind() {
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Map:
		fromKeyType := fromType.Key()
		toKeyType := toType.Key()
		keyCaster, keyHasRef := getCaster(s, fromKeyType, toKeyType)
		if keyCaster == nil {
			return nil, false
		}
		fromElemType := fromType.Elem()
		toElemType := toType.Elem()
		elemCaster, elemHasRef := getCaster(s, fromElemType, toElemType)
		if elemCaster == nil {
			return nil, false
		}
		fromMapHelper := getMapHelper(fromType)
		toMapHelper := getMapHelper(toType)
		// 小优化，减少堆内存分配
		globalKeyBuffer := newObject(toKeyType)
		globalValueBuffer := newObject(toElemType)
		var mu atomic.Uint32
		return func(fromAddr, toAddr unsafe.Pointer) error {
			from := *(*map[any]any)(fromAddr)
			if from == nil {
				*(*map[any]any)(toAddr) = nil
				return nil
			}
			to := toMapHelper.Make(toAddr, len(from))
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
				if err = keyCaster(key, toK); err != nil {
					return false
				}
				if err = elemCaster(value, toV); err != nil {
					return false
				}
				toMapHelper.Store(to, toK, toV)
				return true
			}, keyHasRef, elemHasRef)
			return err
		}, false
	case reflect.Pointer:
		return getAddressingPointerCaster(s, fromType, toType)
	case reflect.Struct:
		toKeyType := toType.Key()
		keyCaster, _ := getCaster(s, stringType, toKeyType)
		if keyCaster == nil {
			return nil, false
		}
		toElemType := toType.Elem()
		n := fromType.NumField()
		toMapHelper := getMapHelper(toType)
		if n == 0 {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				toMapHelper.Make(toAddr, 0)
				return nil
			}, false
		}
		type medaData struct {
			strKey      string
			fieldKey    unsafe.Pointer
			fieldCaster castFunc
			fieldOffset uintptr
		}
		data := make([]medaData, n)
		hasRef := false
		keyIsStr := toKeyType.Kind() == reflect.String
		keyIsRefType := isRefType(toKeyType)
		for i := 0; i < n; i++ {
			field := fromType.Field(i)
			if !s.castUnexported && !field.IsExported() {
				continue
			}
			var fHasRef bool
			data[i].fieldCaster, fHasRef = getCaster(s, field.Type, toElemType)
			if data[i].fieldCaster == nil {
				return nil, false
			}
			data[i].strKey = field.Name
			if jsonTag, ok := field.Tag.Lookup("json"); ok {
				data[i].strKey = jsonTag
			}
			if !keyIsStr && !keyIsRefType {
				data[i].fieldKey = newObject(toKeyType)
				if err := keyCaster(unsafe.Pointer(&data[i].strKey), data[i].fieldKey); err != nil {
					return nil, false
				}
			}
			data[i].fieldOffset = field.Offset
			hasRef = hasRef || fHasRef
		}
		// 小优化，减少堆内存分配
		globalValueBuffer := newObject(toElemType)
		var mu atomic.Uint32
		return func(fromAddr, toAddr unsafe.Pointer) error {
			to := toMapHelper.Make(toAddr, n)
			var v unsafe.Pointer
			if mu.CompareAndSwap(0, 1) {
				defer mu.Store(0)
				v = globalValueBuffer
			} else {
				v = newObject(toElemType)
			}
			for i := 0; i < n; i++ {
				if data[i].fieldCaster == nil {
					continue
				}
				var k unsafe.Pointer
				if keyIsStr {
					k = unsafe.Pointer(&data[i].strKey)
				} else if keyIsRefType {
					k = newObject(toKeyType)
					if err := keyCaster(unsafe.Pointer(&data[i].strKey), k); err != nil {
						return err
					}
				} else {
					k = data[i].fieldKey
				}
				if err := data[i].fieldCaster(unsafe.Add(fromAddr, data[i].fieldOffset), v); err != nil {
					return err
				}
				toMapHelper.Store(to, k, v)
			}
			return nil
		}, hasRef
	default:
		return nil, false
	}
}
