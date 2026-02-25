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

const cacheSize = reflect.UnsafePointer

var cacheableTypePtrs [cacheSize]unsafe.Pointer

func registerCacheableType[T any]() {
	typ := typeFor[T]()
	idx := typ.Kind() - 1
	if idx < 0 || idx >= cacheSize {
		return
	}
	cacheableTypePtrs[idx] = typePtr(typ)
}

func getCacheIdx(fromType, toType reflect.Type) int {
	fromIdx := fromType.Kind() - 1
	if fromIdx < 0 || fromIdx >= cacheSize || cacheableTypePtrs[fromIdx] != typePtr(fromType) {
		return -1
	}
	toIdx := toType.Kind() - 1
	if toIdx < 0 || toIdx >= cacheSize || cacheableTypePtrs[toIdx] != typePtr(toType) {
		return -1
	}
	return int(fromIdx*cacheSize + toIdx)
}

func init() {
	registerCacheableType[bool]()
	registerCacheableType[int]()
	registerCacheableType[int8]()
	registerCacheableType[int16]()
	registerCacheableType[int32]()
	registerCacheableType[int64]()
	registerCacheableType[uint]()
	registerCacheableType[uint8]()
	registerCacheableType[uint16]()
	registerCacheableType[uint32]()
	registerCacheableType[uint64]()
	registerCacheableType[uintptr]()
	registerCacheableType[float32]()
	registerCacheableType[float64]()
	registerCacheableType[complex64]()
	registerCacheableType[complex128]()
	registerCacheableType[string]()
	registerCacheableType[unsafe.Pointer]()
	registerCacheableType[any]()
}

type cache [cacheSize * cacheSize]atomic.Pointer[casterValue]

func (c *cache) load(idx int) (casterValue, bool) {
	ptr := c[idx].Load()
	if ptr == nil {
		return casterValue{}, false
	}
	return *ptr, true
}

func (c *cache) store(idx int, value casterValue) {
	c[idx].Store(&value)
}
