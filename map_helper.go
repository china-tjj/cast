// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build go1.26

package cast

import (
	"reflect"
	"unsafe"
)

const helperMaxSize = 24

// 特殊hash规则
var stringHelpers [helperMaxSize + 1]iMapHelper
var interfaceHelpers [helperMaxSize + 1]iMapHelper

// 通用hash规则
var helpers [helperMaxSize + 1][helperMaxSize + 1]iMapHelper

func getHelperList(keyType reflect.Type) *[helperMaxSize + 1]iMapHelper {
	switch keyType.Kind() {
	case reflect.Array:
		arrayElemType := keyType.Elem()
		if isSpecialHash(arrayElemType) {
			if keyType.Len() == 1 {
				return getHelperList(arrayElemType)
			}
			return nil
		}
	case reflect.String:
		return &stringHelpers
	case reflect.Interface:
		return &interfaceHelpers
	case reflect.Struct:
		n := keyType.NumField()
		var nonZeroFieldType reflect.Type
		nonZeroFieldCnt := 0
		hasSpecialHash := false
		for i := 0; i < n; i++ {
			fieldType := keyType.Field(i).Type
			hasSpecialHash = hasSpecialHash || isSpecialHash(fieldType)
			if fieldType.Size() > 0 {
				nonZeroFieldType = fieldType
				nonZeroFieldCnt++
			}
		}
		if hasSpecialHash {
			if nonZeroFieldCnt == 1 {
				return getHelperList(nonZeroFieldType)
			}
			return nil
		}
	default:
		break
	}
	keySize := keyType.Size()
	if keySize <= helperMaxSize {
		return &helpers[keySize]
	}
	return nil
}

func getMapHelper(mapType reflect.Type) iMapHelper {
	helperList := getHelperList(mapType.Key())
	if helperList != nil {
		valueSize := mapType.Elem().Size()
		if valueSize <= helperMaxSize {
			if helper := helperList[valueSize]; helper != nil {
				return helper
			}
		}
	}
	return newReflectMapHelper(mapType)
}

type nativeMapHelper[K comparable, V any] struct{}

func newNativeMapHelper[K comparable, V any]() iMapHelper {
	return nativeMapHelper[K, V]{}
}

func (h nativeMapHelper[K, V]) Make(mapAddr unsafe.Pointer, n int) map[any]any {
	m := make(map[K]V, n)
	*(*map[K]V)(mapAddr) = m
	return *(*map[any]any)(unsafe.Pointer(&m))
}

func (h nativeMapHelper[K, V]) Load(m map[any]any, key unsafe.Pointer, valueHasRef bool) (unsafe.Pointer, bool) {
	v, ok := (*(*map[K]V)(unsafe.Pointer(&m)))[*(*K)(key)]
	if !ok {
		return nil, false
	}
	return unsafe.Pointer(&v), ok
}

func (h nativeMapHelper[K, V]) Store(m map[any]any, key, value unsafe.Pointer) {
	(*(*map[K]V)(unsafe.Pointer(&m)))[*(*K)(key)] = *(*V)(value)
}

func (h nativeMapHelper[K, V]) Range(m map[any]any, f func(key, value unsafe.Pointer) bool, keyHasRef, valueHasRef bool) {
	for k, v := range *(*map[K]V)(unsafe.Pointer(&m)) {
		var kAddr, vAddr unsafe.Pointer
		if keyHasRef {
			kk := k
			kAddr = unsafe.Pointer(&kk)
		} else {
			kAddr = noEscape(unsafe.Pointer(&k))
		}
		if valueHasRef {
			vv := v
			vAddr = unsafe.Pointer(&vv)
		} else {
			vAddr = noEscape(unsafe.Pointer(&v))
		}
		if !f(kAddr, vAddr) {
			return
		}
	}
}

type reflectMapHelper struct {
	mapType   reflect.Type
	keyType   reflect.Type
	valueType reflect.Type
}

func newReflectMapHelper(mapType reflect.Type) iMapHelper {
	return &reflectMapHelper{
		mapType:   mapType,
		keyType:   mapType.Key(),
		valueType: mapType.Elem(),
	}
}

func (h *reflectMapHelper) Make(mapAddr unsafe.Pointer, n int) map[any]any {
	mv := reflect.MakeMapWithSize(h.mapType, n)
	m := *(*map[any]any)(getValueAddr(mv))
	*(*map[any]any)(mapAddr) = m
	return m
}

func (h *reflectMapHelper) Load(m map[any]any, key unsafe.Pointer, valueHasRef bool) (unsafe.Pointer, bool) {
	mv := reflect.NewAt(h.mapType, unsafe.Pointer(&m)).Elem()
	v := mv.MapIndex(reflect.NewAt(h.keyType, key).Elem())
	if !v.IsValid() {
		return nil, false
	}
	return getValueAddr(v), true
}

func (h *reflectMapHelper) Store(m map[any]any, key, value unsafe.Pointer) {
	mv := reflect.NewAt(h.mapType, unsafe.Pointer(&m)).Elem()
	mv.SetMapIndex(reflect.NewAt(h.keyType, key).Elem(), reflect.NewAt(h.valueType, value).Elem())
}

func (h *reflectMapHelper) Range(m map[any]any, f func(key, value unsafe.Pointer) bool, keyHasRef, valueHasRef bool) {
	mv := reflect.NewAt(h.mapType, unsafe.Pointer(&m)).Elem()
	iter := mv.MapRange()
	for iter.Next() {
		if !f(getValueAddr(iter.Key()), getValueAddr(iter.Value())) {
			return
		}
	}
}
