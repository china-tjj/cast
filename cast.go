// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"sync"
	"unsafe"
)

func Cast[F any, T any](from F) (to T, ok bool) {
	caster := getCaster(typeFor[F](), typeFor[T]())
	if caster == nil {
		return to, false
	}
	ok = caster(unsafe.Pointer(&from), unsafe.Pointer(&to))
	return to, ok
}

func GetCaster[F any, T any]() func(from F) (to T, ok bool) {
	caster := getCaster(typeFor[F](), typeFor[T]())
	if caster == nil {
		return nil
	}
	return func(from F) (to T, ok bool) {
		ok = caster(unsafe.Pointer(&from), unsafe.Pointer(&to))
		return to, ok
	}
}

type castFunc func(fromAddr, toAddr unsafe.Pointer) bool

// 读多写少的场景，sync.RWMutex的效率比sync.Map更高
var casterCache = make(map[casterCacheKey]castFunc)
var casterCacheMutex sync.RWMutex

type casterCacheKey struct {
	from reflect.Type
	to   reflect.Type
}

func getCaster(fromType, toType reflect.Type) castFunc {
	cacheKey := casterCacheKey{from: fromType, to: toType}
	casterCacheMutex.RLock()
	caster, ok := casterCache[cacheKey]
	if ok {
		casterCacheMutex.RUnlock()
		return caster
	}
	casterCacheMutex.RUnlock()
	caster = newCaster(fromType, toType)
	casterCacheMutex.Lock()
	casterCache[cacheKey] = caster
	casterCacheMutex.Unlock()
	return caster
}

func newCaster(fromType, toType reflect.Type) castFunc {
	switch toType.Kind() {
	case reflect.Bool:
		return getBoolCaster(fromType, toType)
	case reflect.Int:
		return getNumberCaster[int](fromType, toType)
	case reflect.Int8:
		return getNumberCaster[int8](fromType, toType)
	case reflect.Int16:
		return getNumberCaster[int16](fromType, toType)
	case reflect.Int32:
		return getNumberCaster[int32](fromType, toType)
	case reflect.Int64:
		return getNumberCaster[int64](fromType, toType)
	case reflect.Uint:
		return getNumberCaster[uint](fromType, toType)
	case reflect.Uint8:
		return getNumberCaster[uint8](fromType, toType)
	case reflect.Uint16:
		return getNumberCaster[uint16](fromType, toType)
	case reflect.Uint32:
		return getNumberCaster[uint32](fromType, toType)
	case reflect.Uint64:
		return getNumberCaster[uint64](fromType, toType)
	case reflect.Uintptr:
		return getNumberCaster[uintptr](fromType, toType)
	case reflect.Float32:
		return getNumberCaster[float32](fromType, toType)
	case reflect.Float64:
		return getNumberCaster[float64](fromType, toType)
	case reflect.Array:
		return getArrayCaster(fromType, toType)
	case reflect.Chan:
		return getChanCaster(fromType, toType)
	case reflect.Interface:
		return getInterfaceCaster(fromType, toType)
	case reflect.Map:
		return getMapCaster(fromType, toType)
	case reflect.Pointer:
		return getPointerCaster(fromType, toType)
	case reflect.Slice:
		return getSliceCaster(fromType, toType)
	case reflect.String:
		return getStringCaster(fromType, toType)
	case reflect.Struct:
		return getStructCaster(fromType, toType)
	case reflect.UnsafePointer:
		return getUnsafePointerCaster(fromType, toType)
	default:
		return nil
	}
}
