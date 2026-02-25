// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type casterKey struct {
	fromTypePtr unsafe.Pointer
	toTypePtr   unsafe.Pointer
}

// fromAddr, toAddr都不能为 nil
// 要求 toAddr 指向的内存为零值 -> 转换失败时，若有部分写入，写回零值
type castFunc func(fromAddr, toAddr unsafe.Pointer) error

type casterValue struct {
	caster castFunc
	flag   uint8
}

func getCaster(s *Scope, fromType, toType reflect.Type) (castFunc, uint8) {
	cacheIdx := getCacheIdx(fromType, toType)
	if cacheIdx != -1 {
		v, ok := s.casterCache.load(cacheIdx)
		if ok {
			return v.caster, v.flag
		}
	}
	key := casterKey{fromTypePtr: typePtr(fromType), toTypePtr: typePtr(toType)}
	s.mu.RLock()
	v, ok := s.casterMap[key]
	s.mu.RUnlock()
	if ok {
		return v.caster, v.flag
	}
	caster, flag := newCaster(s, fromType, toType)
	v = casterValue{caster, flag}
	if cacheIdx != -1 {
		s.casterCache.store(cacheIdx, v)
	}
	s.mu.Lock()
	s.casterMap[key] = v
	s.mu.Unlock()
	return caster, flag
}

func newCaster(s *Scope, fromType, toType reflect.Type) (castFunc, uint8) {
	// 内存布局相同，直接强转
	if isRefAble(s, fromType, toType) {
		fromTypePtr := typePtr(fromType)
		return func(fromAddr, toAddr unsafe.Pointer) error {
			typedMemMove(fromTypePtr, toAddr, fromAddr)
			return nil
		}, 0
	}
	switch toType.Kind() {
	case reflect.Bool:
		return getBoolCaster(s, fromType, toType)
	case reflect.Int:
		return getNumberCaster[int](s, fromType, toType)
	case reflect.Int8:
		return getNumberCaster[int8](s, fromType, toType)
	case reflect.Int16:
		return getNumberCaster[int16](s, fromType, toType)
	case reflect.Int32:
		return getNumberCaster[int32](s, fromType, toType)
	case reflect.Int64:
		return getNumberCaster[int64](s, fromType, toType)
	case reflect.Uint:
		return getNumberCaster[uint](s, fromType, toType)
	case reflect.Uint8:
		return getNumberCaster[uint8](s, fromType, toType)
	case reflect.Uint16:
		return getNumberCaster[uint16](s, fromType, toType)
	case reflect.Uint32:
		return getNumberCaster[uint32](s, fromType, toType)
	case reflect.Uint64:
		return getNumberCaster[uint64](s, fromType, toType)
	case reflect.Uintptr:
		return getNumberCaster[uintptr](s, fromType, toType)
	case reflect.Float32:
		return getNumberCaster[float32](s, fromType, toType)
	case reflect.Float64:
		return getNumberCaster[float64](s, fromType, toType)
	case reflect.Complex64:
		return getComplexCaster[complex64](s, fromType, toType)
	case reflect.Complex128:
		return getComplexCaster[complex128](s, fromType, toType)
	case reflect.Array:
		return getArrayCaster(s, fromType, toType)
	case reflect.Chan:
		return getChanCaster(s, fromType, toType)
	case reflect.Func:
		return getFuncCaster(s, fromType, toType)
	case reflect.Interface:
		return getInterfaceCaster(s, fromType, toType)
	case reflect.Map:
		return getMapCaster(s, fromType, toType)
	case reflect.Pointer:
		return getPointerCaster(s, fromType, toType)
	case reflect.Slice:
		return getSliceCaster(s, fromType, toType)
	case reflect.String:
		return getStringCaster(s, fromType, toType)
	case reflect.Struct:
		return getStructCaster(s, fromType, toType)
	case reflect.UnsafePointer:
		return getUnsafePointerCaster(s, fromType, toType)
	default:
		return nil, 0
	}
}

var timeFormats = []string{
	time.Layout,
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339,
	time.RFC3339Nano,
	time.Kitchen,
	time.Stamp,
	time.StampMilli,
	time.StampMicro,
	time.StampNano,
	time.DateTime,
	time.DateOnly,
	time.TimeOnly,
}

func castStringToTime(s *Scope, str string) (time.Time, error) {
	for _, format := range timeFormats {
		t, err := time.Parse(format, str)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("failed to parse " + str + " as time")
}

func castStringToDuration(s *Scope, str string) (time.Duration, error) {
	if strings.ContainsAny(str, "nuµmsh") {
		return time.ParseDuration(str)
	}
	v, err := strconv.ParseInt(str, 10, 64)
	return time.Duration(v), err
}

var defaultOptions = []ScopeOption{
	WithCaster(castStringToTime),
	WithCaster(castStringToDuration),
}
