// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getMapCaster(s *Scope, fromType, toType reflect.Type) castFunc {
	switch fromType.Kind() {
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Map:
		toKeyType := toType.Key()
		keyCaster := getCaster(s, fromType.Key(), toKeyType)
		if keyCaster == nil {
			return nil
		}
		toElemType := toType.Elem()
		elemCaster := getCaster(s, fromType.Elem(), toElemType)
		if elemCaster == nil {
			return nil
		}
		mapNewer := getMapHelperNewer(fromType)
		mapMaker := getMapHelperMaker(toType)
		toKeySize := toKeyType.Size()
		toElemSize := toElemType.Size()
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if *(*unsafe.Pointer)(fromAddr) == nil {
				return nil
			}
			from := mapNewer(fromAddr)
			to := mapMaker(toAddr)
			var err error
			from.Range(func(key unsafe.Pointer, value unsafe.Pointer) bool {
				toK := malloc(toKeySize)
				if err = keyCaster(key, toK); err != nil {
					return false
				}
				toV := malloc(toElemSize)
				if err = elemCaster(value, toV); err != nil {
					return false
				}
				to.Store(toK, toV)
				return true
			})
			return err
		}
	case reflect.Pointer:
		return getAddressingPointerCaster(s, fromType, toType)
	case reflect.Struct:
		toKeyType := toType.Key()
		keyCaster := getCaster(s, stringType, toKeyType)
		if keyCaster == nil {
			return nil
		}
		toElemType := toType.Elem()
		n := fromType.NumField()
		mapMaker := getMapHelperMaker(toType)
		if n == 0 {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				mapMaker(toAddr)
				return nil
			}
		}
		jsonTags := make([]string, n)
		fieldNames := make([]string, n)
		fieldCasters := make([]castFunc, n)
		fieldOffsets := make([]uintptr, n)
		for i := 0; i < n; i++ {
			field := fromType.Field(i)
			fieldCasters[i] = getCaster(s, field.Type, toElemType)
			if fieldCasters[i] == nil {
				continue
			}
			jsonTags[i] = field.Tag.Get("json")
			fieldNames[i] = field.Name
			fieldOffsets[i] = field.Offset
		}
		toKeySize := toKeyType.Size()
		toElemSize := toElemType.Size()
		// map key会对外暴露，故每次new一个新的
		getKey := func(i int) (unsafe.Pointer, error) {
			k := malloc(toKeySize)
			if jsonTags[i] != "" && keyCaster(unsafe.Pointer(&jsonTags[i]), k) == nil {
				return k, nil
			}
			if err := keyCaster(unsafe.Pointer(&fieldNames[i]), k); err != nil {
				return nil, err
			}
			return k, nil
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			to := mapMaker(toAddr)
			for i := 0; i < n; i++ {
				if fieldCasters[i] == nil {
					continue
				}
				k, err := getKey(i)
				if err != nil {
					return err
				}
				v := malloc(toElemSize)
				if err = fieldCasters[i](unsafe.Add(fromAddr, fieldOffsets[i]), v); err != nil {
					return err
				}
				to.Store(k, v)
			}
			return nil
		}
	default:
		return nil
	}
}
