// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getMapCaster(fromType, toType reflect.Type) castFunc {
	switch fromType.Kind() {
	case reflect.Array:
		return getUnwrapArrayCaster(fromType, toType)
	case reflect.Interface:
		return getUnwrapInterfaceCaster(fromType, toType)
	case reflect.Map:
		if isMemSame(fromType, toType) {
			return func(fromAddr, toAddr unsafe.Pointer) bool {
				*(*unsafe.Pointer)(toAddr) = *(*unsafe.Pointer)(fromAddr)
				return true
			}
		}
		toKeyType := toType.Key()
		keyCaster := getCaster(fromType.Key(), toKeyType)
		if keyCaster == nil {
			return nil
		}
		toElemType := toType.Elem()
		elemCaster := getCaster(fromType.Elem(), toElemType)
		if elemCaster == nil {
			return nil
		}
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			if *(*unsafe.Pointer)(fromAddr) == nil {
				return true
			}
			to := reflect.MakeMap(toType)
			iter := reflect.NewAt(fromType, fromAddr).Elem().MapRange()
			for iter.Next() {
				fromK := iter.Key()
				toK := reflect.New(toKeyType).Elem()
				if !keyCaster(getValueAddr(fromK), getValueAddr(toK)) {
					continue
				}
				fromV := iter.Value()
				toV := reflect.New(toElemType).Elem()
				if !keyCaster(getValueAddr(fromV), getValueAddr(toV)) {
					continue
				}
				to.SetMapIndex(toK, toV)
			}
			*(*unsafe.Pointer)(toAddr) = *(*unsafe.Pointer)(getValueAddr(to))
			return true
		}
	case reflect.Pointer:
		return getAddressingPointerCaster(fromType, toType)
	case reflect.Slice:
		return getUnwrapSliceCaster(fromType, toType)
	case reflect.Struct:
		toKeyType := toType.Key()
		toKeyCaster := getCaster(typeFor[string](), toKeyType)
		if toKeyCaster == nil {
			return nil
		}
		toElemType := toType.Elem()
		n := fromType.NumField()
		// struct field 对应的 map key
		fieldKeys := make([]*reflect.Value, n)
		fieldCasters := make([]castFunc, n)
		for i := 0; i < n; i++ {
			field := fromType.Field(i)
			fieldCasters[i] = newCaster(field.Type, toElemType)
			if fieldCasters[i] == nil {
				continue
			}
			fieldKey := reflect.New(toKeyType).Elem()
			fieldKeyAddr := getValueAddr(fieldKey)
			jsonTag, ok := field.Tag.Lookup("json")
			if ok {
				ok = toKeyCaster(unsafe.Pointer(&jsonTag), fieldKeyAddr)
			} else {
				ok = toKeyCaster(unsafe.Pointer(&field.Name), fieldKeyAddr)
			}
			if !ok {
				continue
			}
			fieldKeys[i] = &fieldKey
		}
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			from := reflect.NewAt(fromType, fromAddr).Elem()
			to := reflect.MakeMap(toType)
			for i := 0; i < n; i++ {
				if fieldKeys[i] == nil || fieldCasters[i] == nil {
					continue
				}
				v := reflect.New(toElemType).Elem()
				if !fieldCasters[i](getValueAddr(from.Field(i)), getValueAddr(v)) {
					continue
				}
				to.SetMapIndex(*fieldKeys[i], v)
			}
			*(*unsafe.Pointer)(toAddr) = *(*unsafe.Pointer)(getValueAddr(to))
			return true
		}
	default:
		return nil
	}
}
