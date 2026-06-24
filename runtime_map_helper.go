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

//go:linkname mapAssign runtime.mapassign
func mapAssign(t unsafe.Pointer, h map[any]any, key unsafe.Pointer) unsafe.Pointer

//go:linkname mapAssignFastStr runtime.mapassign_faststr
func mapAssignFastStr(t unsafe.Pointer, h map[any]any, s string) unsafe.Pointer

//go:linkname mapIterInit runtime.mapiterinit
func mapIterInit(t unsafe.Pointer, h map[any]any, it *hiter)

//go:linkname mapIterNext runtime.mapiternext
func mapIterNext(it *hiter)

type mapHelper struct {
	fastStr     bool
	mapTypePtr  unsafe.Pointer
	elemTypePtr unsafe.Pointer
}

func newMapHelper(mapType reflect.Type) *mapHelper {
	return &mapHelper{
		fastStr:     mapType.Key().Kind() == reflect.String && mapType.Elem().Size() <= 128,
		mapTypePtr:  typePtr(mapType),
		elemTypePtr: typePtr(mapType.Elem()),
	}
}

func (h *mapHelper) Make(n int) map[any]any {
	return makeMap(h.mapTypePtr, n, nil)
}

func (h *mapHelper) Load(m map[any]any, key unsafe.Pointer) (v unsafe.Pointer, ok bool) {
	if h.fastStr {
		return mapAccess2FastStr(h.mapTypePtr, m, *(*string)(key))
	}
	return mapAccess2(h.mapTypePtr, m, key)
}

func (h *mapHelper) Store(m map[any]any, key, value unsafe.Pointer) {
	if h.fastStr {
		typedMemMove(h.elemTypePtr, mapAssignFastStr(h.mapTypePtr, m, *(*string)(key)), value)
	} else {
		typedMemMove(h.elemTypePtr, mapAssign(h.mapTypePtr, m, key), value)
	}
}

func (h *mapHelper) Range(m map[any]any, f func(key, value unsafe.Pointer) bool) {
	var iter hiter
	ptr := noEscapePtr(&iter)
	for mapIterInit(h.mapTypePtr, m, ptr); iter.key != nil; mapIterNext(ptr) {
		k, v := iter.key, iter.elem
		if !f(k, v) {
			return
		}
	}
}
