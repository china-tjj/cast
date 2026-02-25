// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

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

//go:linkname makeMap runtime.makemap
func makeMap(t unsafe.Pointer, hint int, h map[any]any) map[any]any

//go:linkname mapAccess2 runtime.mapaccess2
func mapAccess2(t unsafe.Pointer, h map[any]any, key unsafe.Pointer) (unsafe.Pointer, bool)

//go:linkname mapAccess2FastStr runtime.mapaccess2_faststr
func mapAccess2FastStr(t unsafe.Pointer, h map[any]any, key string) (unsafe.Pointer, bool)

//go:linkname mapAccess2Fast32 runtime.mapaccess2_fast32
func mapAccess2Fast32(t unsafe.Pointer, h map[any]any, key uint32) (unsafe.Pointer, bool)

//go:linkname mapAccess2Fast64 runtime.mapaccess2_fast64
func mapAccess2Fast64(t unsafe.Pointer, h map[any]any, key uint64) (unsafe.Pointer, bool)

//go:linkname mapAssign runtime.mapassign
func mapAssign(t unsafe.Pointer, h map[any]any, key unsafe.Pointer) unsafe.Pointer

//go:linkname mapAssignFastStr runtime.mapassign_faststr
func mapAssignFastStr(t unsafe.Pointer, h map[any]any, s string) unsafe.Pointer

//go:linkname mapAssignFast32 runtime.mapassign_fast32
func mapAssignFast32(t unsafe.Pointer, h map[any]any, key uint32) unsafe.Pointer

//go:linkname mapAssignFast64 runtime.mapassign_fast64
func mapAssignFast64(t unsafe.Pointer, h map[any]any, key uint64) unsafe.Pointer

//go:linkname mapAssignFast32Ptr runtime.mapassign_fast32ptr
func mapAssignFast32Ptr(t unsafe.Pointer, h map[any]any, key unsafe.Pointer) unsafe.Pointer

//go:linkname mapAssignFast64Ptr runtime.mapassign_fast64ptr
func mapAssignFast64Ptr(t unsafe.Pointer, h map[any]any, key unsafe.Pointer) unsafe.Pointer

//go:linkname mapIterInit runtime.mapiterinit
func mapIterInit(t unsafe.Pointer, h map[any]any, it *hiter)

//go:linkname mapIterNext runtime.mapiternext
func mapIterNext(it *hiter)

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

func getKeyFlag(keyType reflect.Type) int8 {
	switch keyType.Kind() {
	case reflect.String:
		return keyFlagFastStr
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return keyFlagFast
	case reflect.UnsafePointer, reflect.Ptr:
		return keyFlagFastPtr
	case reflect.Array:
		length := keyType.Len()
		if uintptr(length)*keyType.Elem().Size() != keyType.Size() {
			return keyFlagNormal
		}
		switch length {
		case 0:
			return keyFlagEmpty
		case 1:
			return getKeyFlag(keyType.Elem())
		default:
			elemkeyFlag := getKeyFlag(keyType.Elem())
			if elemkeyFlag == keyFlagFastStr {
				return keyFlagNormal
			}
			return elemkeyFlag
		}
	case reflect.Struct:
		n := keyType.NumField()
		result := int8(keyFlagEmpty)
		var size uintptr
		for i := 0; i < n; i++ {
			field := keyType.Field(i)
			size += field.Type.Size()
			fieldKeyFlag := getKeyFlag(field.Type)
			if fieldKeyFlag == keyFlagEmpty {
				continue
			}
			if result == keyFlagEmpty {
				result = fieldKeyFlag
				continue
			}
			if result != fieldKeyFlag {
				return keyFlagNormal
			}
			// 出现多个str，不能用fastStr
			if result == keyFlagFastStr {
				return keyFlagNormal
			}
		}
		if size != keyType.Size() {
			return keyFlagNormal
		}
		return result
	default:
		return keyFlagNormal
	}
}

func newMapHelper(mapType reflect.Type) *runtimeMapHelper {
	return newRuntimeMapHelper(mapType)
}

type runtimeMapHelper struct {
	keyFlag     uint8
	mapTypePtr  unsafe.Pointer
	elemTypePtr unsafe.Pointer
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
	}
}

func (h *runtimeMapHelper) Make(n int) map[any]any {
	return makeMap(h.mapTypePtr, n, nil)
}

func (h *runtimeMapHelper) Load(m map[any]any, key unsafe.Pointer) (v unsafe.Pointer, ok bool) {
	switch keyFlag := h.keyFlag; keyFlag {
	case keyFlagFastStr:
		v, ok = mapAccess2FastStr(h.mapTypePtr, m, *(*string)(key))
	case keyFlagFast32, keyFlagFast32Ptr:
		v, ok = mapAccess2Fast32(h.mapTypePtr, m, *(*uint32)(key))
	case keyFlagFast64, keyFlagFast64Ptr:
		v, ok = mapAccess2Fast64(h.mapTypePtr, m, *(*uint64)(key))
	default:
		v, ok = mapAccess2(h.mapTypePtr, m, key)
	}
	if !ok {
		return nil, false
	}
	return v, true
}

func (h *runtimeMapHelper) Store(m map[any]any, key, value unsafe.Pointer) {
	switch keyFlag := h.keyFlag; keyFlag {
	case keyFlagFastStr:
		typedMemMove(h.elemTypePtr, mapAssignFastStr(h.mapTypePtr, m, *(*string)(key)), value)
	case keyFlagFast32:
		typedMemMove(h.elemTypePtr, mapAssignFast32(h.mapTypePtr, m, *(*uint32)(key)), value)
	case keyFlagFast32Ptr:
		typedMemMove(h.elemTypePtr, mapAssignFast32Ptr(h.mapTypePtr, m, *(*unsafe.Pointer)(key)), value)
	case keyFlagFast64:
		typedMemMove(h.elemTypePtr, mapAssignFast64(h.mapTypePtr, m, *(*uint64)(key)), value)
	case keyFlagFast64Ptr:
		typedMemMove(h.elemTypePtr, mapAssignFast64Ptr(h.mapTypePtr, m, *(*unsafe.Pointer)(key)), value)
	default:
		typedMemMove(h.elemTypePtr, mapAssign(h.mapTypePtr, m, key), value)
	}
}

func (h *runtimeMapHelper) Range(m map[any]any, f func(key, value unsafe.Pointer) bool) {
	var iter hiter
	ptr := noEscapePtr(&iter)
	for mapIterInit(h.mapTypePtr, m, ptr); iter.key != nil; mapIterNext(ptr) {
		k, v := iter.key, iter.elem
		if !f(k, v) {
			return
		}
	}
}
