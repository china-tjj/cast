// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build !go1.26

package cast

import (
	"reflect"
	"unsafe"
)

type hiter struct {
	key         unsafe.Pointer
	elem        unsafe.Pointer
	t           unsafe.Pointer
	h           unsafe.Pointer
	buckets     unsafe.Pointer
	bptr        unsafe.Pointer
	overflow    *[]unsafe.Pointer
	oldoverflow *[]unsafe.Pointer
	startBucket uintptr
	offset      uint8
	wrapped     bool
	B           uint8
	i           uint8
	bucket      uintptr
	checkBucket uintptr
}

//go:linkname makemap runtime.makemap
func makemap(t unsafe.Pointer, hint int, h map[any]any) map[any]any

//go:linkname mapaccess2 runtime.mapaccess2
func mapaccess2(t unsafe.Pointer, h map[any]any, key unsafe.Pointer) (unsafe.Pointer, bool)

//go:linkname mapaccess2FastStr runtime.mapaccess2_faststr
func mapaccess2FastStr(t unsafe.Pointer, h map[any]any, ky string) (unsafe.Pointer, bool)

//go:linkname mapaccess2Fast32 runtime.mapaccess2_fast32
func mapaccess2Fast32(t unsafe.Pointer, h map[any]any, key uint32) (unsafe.Pointer, bool)

//go:linkname mapaccess2Fast64 runtime.mapaccess2_fast64
func mapaccess2Fast64(t unsafe.Pointer, h map[any]any, key uint64) (unsafe.Pointer, bool)

//go:linkname mapassign runtime.mapassign
func mapassign(t unsafe.Pointer, h map[any]any, key unsafe.Pointer) unsafe.Pointer

//go:linkname mapassignFastStr runtime.mapassign_faststr
func mapassignFastStr(t unsafe.Pointer, h map[any]any, s string) unsafe.Pointer

//go:linkname mapassignFast32 runtime.mapassign_fast32
func mapassignFast32(t unsafe.Pointer, h map[any]any, key uint32) unsafe.Pointer

//go:linkname mapassignFast64 runtime.mapassign_fast64
func mapassignFast64(t unsafe.Pointer, h map[any]any, key uint64) unsafe.Pointer

//go:linkname mapassignFast32Ptr runtime.mapassign_fast32ptr
func mapassignFast32Ptr(t unsafe.Pointer, h map[any]any, key unsafe.Pointer) unsafe.Pointer

//go:linkname mapassignFast64Ptr runtime.mapassign_fast64ptr
func mapassignFast64Ptr(t unsafe.Pointer, h map[any]any, key unsafe.Pointer) unsafe.Pointer

//go:linkname mapiterinit runtime.mapiterinit
func mapiterinit(t unsafe.Pointer, h map[any]any, it *hiter)

//go:linkname mapiternext runtime.mapiternext
func mapiternext(it *hiter)

const (
	keyFlagFastStr = iota
	keyFlagFast32
	keyFlagFast32Ptr
	keyFlagFast64
	keyFlagFast64Ptr

	keyFlagEmpty
	keyFlagFast    // fast32 / fast64
	keyFlagFastPtr // fast32ptr / fast64ptr
	keyFlagNormal
)

func getKeyFlag(keyFlag reflect.Type) int8 {
	switch keyFlag.Kind() {
	case reflect.String:
		return keyFlagFastStr
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return keyFlagFast
	case reflect.UnsafePointer, reflect.Ptr:
		return keyFlagFastPtr
	case reflect.Array:
		switch keyFlag.Len() {
		case 0:
			return keyFlagEmpty
		case 1:
			return getKeyFlag(keyFlag.Elem())
		default:
			elemkeyFlag := getKeyFlag(keyFlag.Elem())
			if elemkeyFlag == keyFlagFastStr {
				return keyFlagNormal
			}
			return elemkeyFlag
		}
	case reflect.Struct:
		n := keyFlag.NumField()
		result := int8(keyFlagEmpty)
		for i := 0; i < n; i++ {
			fieldkeyFlag := getKeyFlag(keyFlag.Field(i).Type)
			if fieldkeyFlag == keyFlagEmpty {
				continue
			}
			if result == keyFlagEmpty {
				result = fieldkeyFlag
				continue
			}
			if result != fieldkeyFlag {
				return keyFlagNormal
			}
			// 出现多个str，不能用fastStr
			if result == keyFlagFastStr {
				return keyFlagNormal
			}
		}
		return result
	default:
		return keyFlagNormal
	}
}

func getMapHelper(mapType reflect.Type) *runtimeMapHelper {
	return newRuntimeMapHelper(mapType)
}

type runtimeMapHelper struct {
	keyFlag     uint8
	mapTypePtr  unsafe.Pointer
	elemTypePtr unsafe.Pointer
	keyType     reflect.Type
	elemType    reflect.Type
}

func newRuntimeMapHelper(mapType reflect.Type) *runtimeMapHelper {
	keyType := mapType.Key()
	var keyFlag uint8
	switch getKeyFlag(keyType) {
	case keyFlagFastStr:
		keyFlag = keyFlagFastStr
	case keyFlagFast:
		switch keyType.Size() {
		case 4:
			keyFlag = keyFlagFast32
		case 8:
			keyFlag = keyFlagFast64
		default:
			keyFlag = keyFlagNormal
		}
	case keyFlagFastPtr:
		switch keyType.Size() {
		case 4:
			keyFlag = keyFlagFast32Ptr
		case 8:
			keyFlag = keyFlagFast64Ptr
		default:
			keyFlag = keyFlagNormal
		}
	default:
		keyFlag = keyFlagNormal
	}
	return &runtimeMapHelper{
		keyFlag:     keyFlag,
		mapTypePtr:  typePtr(mapType),
		elemTypePtr: typePtr(mapType.Elem()),
		keyType:     mapType.Key(),
		elemType:    mapType.Elem(),
	}
}

func (h *runtimeMapHelper) Make(mapAddr unsafe.Pointer, n int) map[any]any {
	m := makemap(h.mapTypePtr, n, nil)
	*(*map[any]any)(mapAddr) = m
	return m
}

func (h *runtimeMapHelper) Load(m map[any]any, key unsafe.Pointer, valueHasRef bool) (v unsafe.Pointer, ok bool) {
	switch keyFlag := h.keyFlag; keyFlag {
	case keyFlagFastStr:
		v, ok = mapaccess2FastStr(h.mapTypePtr, m, *(*string)(key))
	case keyFlagFast32, keyFlagFast32Ptr:
		v, ok = mapaccess2Fast32(h.mapTypePtr, m, *(*uint32)(key))
	case keyFlagFast64, keyFlagFast64Ptr:
		v, ok = mapaccess2Fast64(h.mapTypePtr, m, *(*uint64)(key))
	default:
		v, ok = mapaccess2(h.mapTypePtr, m, key)
	}
	if !ok {
		return nil, false
	}
	if valueHasRef {
		v = copyObject(h.elemType, v)
	}
	return v, true
}

func (h *runtimeMapHelper) Store(m map[any]any, key, value unsafe.Pointer) {
	switch keyFlag := h.keyFlag; keyFlag {
	case keyFlagFastStr:
		typedmemmove(h.elemTypePtr, mapassignFastStr(h.mapTypePtr, m, *(*string)(key)), value)
	case keyFlagFast32:
		typedmemmove(h.elemTypePtr, mapassignFast32(h.mapTypePtr, m, *(*uint32)(key)), value)
	case keyFlagFast32Ptr:
		typedmemmove(h.elemTypePtr, mapassignFast32Ptr(h.mapTypePtr, m, *(*unsafe.Pointer)(key)), value)
	case keyFlagFast64:
		typedmemmove(h.elemTypePtr, mapassignFast64(h.mapTypePtr, m, *(*uint64)(key)), value)
	case keyFlagFast64Ptr:
		typedmemmove(h.elemTypePtr, mapassignFast64Ptr(h.mapTypePtr, m, *(*unsafe.Pointer)(key)), value)
	default:
		typedmemmove(h.elemTypePtr, mapassign(h.mapTypePtr, m, key), value)
	}
}

func (h *runtimeMapHelper) Range(m map[any]any, f func(key, value unsafe.Pointer) bool, keyHasRef, valueHasRef bool) {
	var iter hiter
	for mapiterinit(h.mapTypePtr, m, noEscapePtr(&iter)); iter.key != nil; mapiternext(noEscapePtr(&iter)) {
		k, v := iter.key, iter.elem
		if keyHasRef {
			k = copyObject(h.keyType, k)
		}
		if valueHasRef {
			v = copyObject(h.elemType, v)
		}
		if !f(k, v) {
			return
		}
	}
}
