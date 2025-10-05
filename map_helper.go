// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

type mapHelper interface {
	Load(key unsafe.Pointer) (value unsafe.Pointer, ok bool)
	Store(key, value unsafe.Pointer)
	Range(f func(key, value unsafe.Pointer) bool)
	Len() int
}

const helperMaxSize = 24

// 特殊hash规则
var stringHelperMaker [helperMaxSize + 1]func(mapAddr unsafe.Pointer) mapHelper
var stringHelperNewer [helperMaxSize + 1]func(mapAddr unsafe.Pointer) mapHelper
var interfaceHelperMaker [helperMaxSize + 1]func(mapAddr unsafe.Pointer) mapHelper
var interfaceHelperNewer [helperMaxSize + 1]func(mapAddr unsafe.Pointer) mapHelper

// 通用hash规则
var helperMaker [helperMaxSize + 1][helperMaxSize + 1]func(mapAddr unsafe.Pointer) mapHelper
var helperNewer [helperMaxSize + 1][helperMaxSize + 1]func(mapAddr unsafe.Pointer) mapHelper

func getMakerList(keyType reflect.Type) *[helperMaxSize + 1]func(mapAddr unsafe.Pointer) mapHelper {
	switch keyType.Kind() {
	case reflect.Array:
		arrayElemType := keyType.Elem()
		if isSpecialHash(arrayElemType) {
			if keyType.Len() == 1 {
				return getMakerList(arrayElemType)
			}
			return nil
		}
	case reflect.String:
		return &stringHelperMaker
	case reflect.Interface:
		return &interfaceHelperMaker
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
				return getMakerList(nonZeroFieldType)
			}
			return nil
		}
	default:
		break
	}
	keySize := keyType.Size()
	if keySize <= helperMaxSize {
		return &helperMaker[keySize]
	}
	return nil
}

func getMapHelperMaker(mapType reflect.Type) func(mapAddr unsafe.Pointer) mapHelper {
	makerList := getMakerList(mapType.Key())
	if makerList != nil {
		valueSize := mapType.Elem().Size()
		if valueSize <= helperMaxSize {
			if maker := makerList[valueSize]; maker != nil {
				return maker
			}
		}
	}
	return func(mapAddr unsafe.Pointer) mapHelper {
		return makeReflectMapHelper(mapType, mapAddr)
	}
}

func getNewerList(keyType reflect.Type) *[helperMaxSize + 1]func(mapAddr unsafe.Pointer) mapHelper {
	switch keyType.Kind() {
	case reflect.Array:
		arrayElemType := keyType.Elem()
		if isSpecialHash(arrayElemType) {
			if keyType.Len() == 1 {
				return getNewerList(arrayElemType)
			}
			return nil
		}
	case reflect.String:
		return &stringHelperNewer
	case reflect.Interface:
		return &interfaceHelperNewer
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
				return getNewerList(nonZeroFieldType)
			}
			return nil
		}
	default:
		break
	}
	keySize := keyType.Size()
	if keySize <= helperMaxSize {
		return &helperNewer[keySize]
	}
	return nil
}

func getMapHelperNewer(mapType reflect.Type) func(mapAddr unsafe.Pointer) mapHelper {
	newerList := getNewerList(mapType.Key())
	if newerList != nil {
		valueSize := mapType.Elem().Size()
		if valueSize <= helperMaxSize {
			if newer := newerList[valueSize]; newer != nil {
				return newer
			}
		}
	}
	return func(mapAddr unsafe.Pointer) mapHelper {
		return newReflectMapHelper(mapType, mapAddr)
	}
}

type nativeMapHelper[K comparable, V any] map[K]V

func makeNativeMapHelper[K comparable, V any](mapAddr unsafe.Pointer) mapHelper {
	m := make(nativeMapHelper[K, V])
	*(*nativeMapHelper[K, V])(mapAddr) = m
	return m
}

func newNativeMapHelper[K comparable, V any](mapAddr unsafe.Pointer) mapHelper {
	return *(*nativeMapHelper[K, V])(mapAddr)
}

func (m nativeMapHelper[K, V]) Load(key unsafe.Pointer) (unsafe.Pointer, bool) {
	v, ok := m[*(*K)(key)]
	if !ok {
		return nil, false
	}
	return unsafe.Pointer(&v), ok
}

func (m nativeMapHelper[K, V]) Store(key, value unsafe.Pointer) {
	m[*(*K)(key)] = *(*V)(value)
}

func (m nativeMapHelper[K, V]) Range(f func(key, value unsafe.Pointer) bool) {
	for k, v := range m {
		if !f(unsafe.Pointer(&k), unsafe.Pointer(&v)) {
			return
		}
	}
}

func (m nativeMapHelper[K, V]) Len() int {
	return len(m)
}

type reflectMapHelper struct {
	keyType   reflect.Type
	valueType reflect.Type
	m         reflect.Value
}

func makeReflectMapHelper(mapType reflect.Type, mapAddr unsafe.Pointer) mapHelper {
	m := reflect.MakeMap(mapType)
	*(*unsafe.Pointer)(mapAddr) = *(*unsafe.Pointer)(getValueAddr(m))
	return &reflectMapHelper{
		keyType:   mapType.Key(),
		valueType: mapType.Elem(),
		m:         m,
	}
}

func newReflectMapHelper(mapType reflect.Type, mapAddr unsafe.Pointer) mapHelper {
	return &reflectMapHelper{
		keyType:   mapType.Key(),
		valueType: mapType.Elem(),
		m:         reflect.NewAt(mapType, mapAddr).Elem(),
	}
}

func (m *reflectMapHelper) Load(key unsafe.Pointer) (unsafe.Pointer, bool) {
	v := m.m.MapIndex(reflect.NewAt(m.keyType, key).Elem())
	if !v.IsValid() {
		return nil, false
	}
	return getValueAddr(v), true
}

func (m *reflectMapHelper) Store(key, value unsafe.Pointer) {
	m.m.SetMapIndex(reflect.NewAt(m.keyType, key).Elem(), reflect.NewAt(m.valueType, value).Elem())
}

func (m *reflectMapHelper) Range(f func(key, value unsafe.Pointer) bool) {
	iter := m.m.MapRange()
	for iter.Next() {
		if !f(getValueAddr(iter.Key()), getValueAddr(iter.Value())) {
			return
		}
	}
}

func (m *reflectMapHelper) Len() int {
	return m.m.Len()
}
