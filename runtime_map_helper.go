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
	keyTypeEmpty = iota
	keyTypeFastStr
	keyTypeFast    // fast32 / fast64
	keyTypeFastPtr // fast32ptr / fast64ptr
	keyTypeNormal
)

func getKeyType(keyType reflect.Type) int8 {
	switch keyType.Kind() {
	case reflect.String:
		return keyTypeFastStr
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return keyTypeFast
	case reflect.UnsafePointer, reflect.Ptr:
		return keyTypeFastPtr
	case reflect.Array:
		switch keyType.Len() {
		case 0:
			return keyTypeEmpty
		case 1:
			return getKeyType(keyType.Elem())
		default:
			elemKeyType := getKeyType(keyType.Elem())
			if elemKeyType == keyTypeFastStr {
				return keyTypeNormal
			}
			return elemKeyType
		}
	case reflect.Struct:
		n := keyType.NumField()
		result := int8(keyTypeEmpty)
		for i := 0; i < n; i++ {
			fieldKeyType := getKeyType(keyType.Field(i).Type)
			if fieldKeyType == keyTypeEmpty {
				continue
			}
			if result == keyTypeEmpty {
				result = fieldKeyType
				continue
			}
			if result != fieldKeyType {
				return keyTypeNormal
			}
			// 出现多个str，不能用fastStr
			if result == keyTypeFastStr {
				return keyTypeNormal
			}
		}
		return result
	default:
		return keyTypeNormal
	}
}

func getMapHelper(mapType reflect.Type) iMapHelper {
	keyType := mapType.Key()
	switch getKeyType(keyType) {
	case keyTypeFastStr:
		return newRuntimeMapHelperFastStr(mapType)
	case keyTypeFast:
		switch keyType.Size() {
		case 4:
			return newRuntimeMapHelperFast32(mapType)
		case 8:
			return newRuntimeMapHelperFast64(mapType)
		default:
			return newRuntimeMapHelper(mapType)
		}
	case keyTypeFastPtr:
		switch keyType.Size() {
		case 4:
			return newRuntimeMapHelperFast32Ptr(mapType)
		case 8:
			return newRuntimeMapHelperFast64Ptr(mapType)
		default:
			return newRuntimeMapHelper(mapType)
		}
	default:
		return newRuntimeMapHelper(mapType)
	}
}

type runtimeMapHelper struct {
	mapTypePtr  unsafe.Pointer
	elemTypePtr unsafe.Pointer
	keyType     reflect.Type
	elemType    reflect.Type
}

func newRuntimeMapHelper(mapType reflect.Type) iMapHelper {
	return &runtimeMapHelper{
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

func (h *runtimeMapHelper) Load(m map[any]any, key unsafe.Pointer, valueHasRef bool) (unsafe.Pointer, bool) {
	v, ok := mapaccess2(h.mapTypePtr, m, key)
	if !ok {
		return nil, false
	}
	if valueHasRef {
		v = copyObject(h.elemType, v)
	}
	return v, true
}

func (h *runtimeMapHelper) Store(m map[any]any, key, value unsafe.Pointer) {
	typedmemmove(h.elemTypePtr, mapassign(h.mapTypePtr, m, key), value)
}

func (h *runtimeMapHelper) Range(m map[any]any, f func(key, value unsafe.Pointer) bool, keyHasRef, valueHasRef bool) {
	var iter hiter
	mapiterinit(h.mapTypePtr, m, noEscapePtr(&iter))
	for {
		if iter.key == nil {
			return
		}
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
		mapiternext(noEscapePtr(&iter))
	}
}

type runtimeMapHelperFastStr struct {
	runtimeMapHelper
}

func newRuntimeMapHelperFastStr(mapType reflect.Type) iMapHelper {
	return &runtimeMapHelperFastStr{
		runtimeMapHelper{
			mapTypePtr:  typePtr(mapType),
			elemTypePtr: typePtr(mapType.Elem()),
			keyType:     mapType.Key(),
			elemType:    mapType.Elem(),
		},
	}
}

func (h *runtimeMapHelperFastStr) Load(m map[any]any, key unsafe.Pointer, valueHasRef bool) (unsafe.Pointer, bool) {
	v, ok := mapaccess2FastStr(h.mapTypePtr, m, *(*string)(key))
	if !ok {
		return nil, false
	}
	if valueHasRef {
		v = copyObject(h.elemType, v)
	}
	return v, true
}

func (h *runtimeMapHelperFastStr) Store(m map[any]any, key, value unsafe.Pointer) {
	typedmemmove(h.elemTypePtr, mapassignFastStr(h.mapTypePtr, m, *(*string)(key)), value)
}

type runtimeMapHelperFast32 struct {
	runtimeMapHelper
}

func newRuntimeMapHelperFast32(mapType reflect.Type) iMapHelper {
	return &runtimeMapHelperFast32{
		runtimeMapHelper{
			mapTypePtr:  typePtr(mapType),
			elemTypePtr: typePtr(mapType.Elem()),
			keyType:     mapType.Key(),
			elemType:    mapType.Elem(),
		},
	}
}

func (h *runtimeMapHelperFast32) Load(m map[any]any, key unsafe.Pointer, valueHasRef bool) (unsafe.Pointer, bool) {
	v, ok := mapaccess2Fast32(h.mapTypePtr, m, *(*uint32)(key))
	if !ok {
		return nil, false
	}
	if valueHasRef {
		v = copyObject(h.elemType, v)
	}
	return v, true
}

func (h *runtimeMapHelperFast32) Store(m map[any]any, key, value unsafe.Pointer) {
	typedmemmove(h.elemTypePtr, mapassignFast32(h.mapTypePtr, m, *(*uint32)(key)), value)
}

type runtimeMapHelperFast32Ptr struct {
	runtimeMapHelperFast32
}

func newRuntimeMapHelperFast32Ptr(mapType reflect.Type) iMapHelper {
	return &runtimeMapHelperFast32Ptr{
		runtimeMapHelperFast32{
			runtimeMapHelper{
				mapTypePtr:  typePtr(mapType),
				elemTypePtr: typePtr(mapType.Elem()),
				keyType:     mapType.Key(),
				elemType:    mapType.Elem(),
			},
		},
	}
}

func (h *runtimeMapHelperFast32Ptr) Store(m map[any]any, key, value unsafe.Pointer) {
	typedmemmove(h.elemTypePtr, mapassignFast32Ptr(h.mapTypePtr, m, *(*unsafe.Pointer)(key)), value)
}

type runtimeMapHelperFast64 struct {
	runtimeMapHelper
}

func newRuntimeMapHelperFast64(mapType reflect.Type) iMapHelper {
	return &runtimeMapHelperFast64{
		runtimeMapHelper{
			mapTypePtr:  typePtr(mapType),
			elemTypePtr: typePtr(mapType.Elem()),
			keyType:     mapType.Key(),
			elemType:    mapType.Elem(),
		},
	}
}

func (h *runtimeMapHelperFast64) Load(m map[any]any, key unsafe.Pointer, valueHasRef bool) (unsafe.Pointer, bool) {
	v, ok := mapaccess2Fast64(h.mapTypePtr, m, *(*uint64)(key))
	if !ok {
		return nil, false
	}
	if valueHasRef {
		v = copyObject(h.elemType, v)
	}
	return v, true
}

func (h *runtimeMapHelperFast64) Store(m map[any]any, key, value unsafe.Pointer) {
	typedmemmove(h.elemTypePtr, mapassignFast64(h.mapTypePtr, m, *(*uint64)(key)), value)
}

type runtimeMapHelperFast64Ptr struct {
	runtimeMapHelperFast64
}

func newRuntimeMapHelperFast64Ptr(mapType reflect.Type) iMapHelper {
	return &runtimeMapHelperFast64Ptr{
		runtimeMapHelperFast64{
			runtimeMapHelper{
				mapTypePtr:  typePtr(mapType),
				elemTypePtr: typePtr(mapType.Elem()),
				keyType:     mapType.Key(),
				elemType:    mapType.Elem(),
			},
		},
	}
}

func (h *runtimeMapHelperFast64Ptr) Store(m map[any]any, key, value unsafe.Pointer) {
	typedmemmove(h.elemTypePtr, mapassignFast64Ptr(h.mapTypePtr, m, *(*unsafe.Pointer)(key)), value)
}
