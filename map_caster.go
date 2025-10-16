// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
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
		return func(fromAddr, toAddr unsafe.Pointer) error {
			from := *(*map[any]any)(fromAddr)
			if from == nil {
				return nil
			}
			to := toMapHelper.Make(toAddr, len(from))
			var err error
			fromMapHelper.Range(from, func(key unsafe.Pointer, value unsafe.Pointer) bool {
				toK := newObject(toKeyType)
				if err = keyCaster(key, toK); err != nil {
					return false
				}
				toV := newObject(toElemType)
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
		jsonTags := make([]string, n)
		fieldNames := make([]string, n)
		fieldCasters := make([]castFunc, n)
		fieldOffsets := make([]uintptr, n)
		for i := 0; i < n; i++ {
			field := fromType.Field(i)
			fieldCasters[i], _ = getCaster(s, field.Type, toElemType)
			if fieldCasters[i] == nil {
				continue
			}
			jsonTags[i], _ = getDiffJsonTag(&field)
			fieldNames[i] = field.Name
			fieldOffsets[i] = field.Offset
		}
		// map key会对外暴露，故每次new一个新的
		getKey := func(i int) (unsafe.Pointer, error) {
			k := newObject(toKeyType)
			if jsonTags[i] != "" && keyCaster(unsafe.Pointer(&jsonTags[i]), k) == nil {
				return k, nil
			}
			if err := keyCaster(unsafe.Pointer(&fieldNames[i]), k); err != nil {
				return nil, err
			}
			return k, nil
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			to := toMapHelper.Make(toAddr, n)
			for i := 0; i < n; i++ {
				if fieldCasters[i] == nil {
					continue
				}
				k, err := getKey(i)
				if err != nil {
					return err
				}
				v := newObject(toElemType)
				if err = fieldCasters[i](unsafe.Add(fromAddr, fieldOffsets[i]), v); err != nil {
					return err
				}
				toMapHelper.Store(to, k, v)
			}
			return nil
		}, false
	default:
		return nil, false
	}
}
