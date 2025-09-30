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

// Cast 将类型F转为T，转换规则详见README
func Cast[F any, T any](from F) (to T, err error) {
	fromType := typeFor[F]()
	toType := typeFor[T]()
	caster := getCaster(fromType, toType)
	if caster == nil {
		return to, invalidCastErr(fromType, toType)
	}
	err = caster(unsafe.Pointer(&from), unsafe.Pointer(&to))
	return to, err
}

// GetCaster 获取实例化的转换方法，调用该方法返回的函数，对比直接调用 Cast 少了查缓存的步骤，性能会略微好一点
func GetCaster[F any, T any]() func(from F) (to T, err error) {
	fromType := typeFor[F]()
	toType := typeFor[T]()
	caster := getCaster(fromType, toType)
	if caster == nil {
		e := invalidCastErr(fromType, toType)
		return func(from F) (to T, err error) {
			return to, e
		}
	}
	return func(from F) (to T, err error) {
		err = caster(unsafe.Pointer(&from), unsafe.Pointer(&to))
		return to, err
	}
}

// ReflectCast 以反射的方式，需输入待转换的值与要转换的类型
func ReflectCast(from reflect.Value, toType reflect.Type) (to reflect.Value, err error) {
	fromType := from.Type()
	caster := getCaster(fromType, toType)
	if caster == nil {
		return to, invalidCastErr(fromType, toType)
	}
	fromAddr := getValueAddr(from)
	to = reflect.New(toType).Elem()
	if fromAddr == nil {
		return to, nil
	}
	err = caster(fromAddr, getValueAddr(to))
	return to, err
}

func getReflectCaster(fromType reflect.Type, toType reflect.Type) reflectCastFunc {
	caster := getCaster(fromType, toType)
	if caster == nil {
		return nil
	}
	return func(from reflect.Value) (reflect.Value, error) {
		fromAddr := getValueAddr(from)
		to := reflect.New(toType).Elem()
		if fromAddr == nil {
			return to, nil
		}
		err := caster(fromAddr, getValueAddr(to))
		return to, err
	}
}

// fromAddr, toAddr都不能为nil
type castFunc func(fromAddr, toAddr unsafe.Pointer) error
type reflectCastFunc func(from reflect.Value) (reflect.Value, error)

// 读多写少的场景，sync.RWMutex的效率比sync.Map更高
var casterCache = make(map[casterCacheKey]castFunc)
var casterCacheMutex sync.RWMutex

type casterCacheKey struct {
	fromTypePtr unsafe.Pointer
	toTypePtr   unsafe.Pointer
}

func getCaster(fromType, toType reflect.Type) castFunc {
	cacheKey := casterCacheKey{fromTypePtr: typePtr(fromType), toTypePtr: typePtr(toType)}
	casterCacheMutex.RLock()
	caster, ok := casterCache[cacheKey]
	casterCacheMutex.RUnlock()
	if ok {
		return caster
	}
	caster = newCaster(fromType, toType)
	casterCacheMutex.Lock()
	casterCache[cacheKey] = caster
	casterCacheMutex.Unlock()
	return caster
}

func newCaster(fromType, toType reflect.Type) castFunc {
	if isMemSame(fromType, toType) {
		size := fromType.Size()
		return func(fromAddr unsafe.Pointer, toAddr unsafe.Pointer) error {
			memCopy(toAddr, fromAddr, size)
			return nil
		}
	}
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
	case reflect.Func:
		return getFuncCaster(fromType, toType)
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
