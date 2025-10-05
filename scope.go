// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

type Scope struct {
	casterMap       map[casterKey]castFunc
	mu              sync.RWMutex // 读多写少的场景，sync.RWMutex的效率比sync.Map更高
	disableZeroCopy bool         // 禁用零拷贝
	frozen          bool
}

// NewScope 创建新的作用域
func NewScope(options ...ScopeOption) *Scope {
	scope := &Scope{
		casterMap: make(map[casterKey]castFunc),
	}
	for _, option := range defaultOptions {
		option(scope)
	}
	for _, option := range options {
		option(scope)
	}
	scope.frozen = true
	return scope
}

type casterKey struct {
	fromTypePtr unsafe.Pointer
	toTypePtr   unsafe.Pointer
}

// fromAddr, toAddr都不能为nil
type castFunc func(fromAddr, toAddr unsafe.Pointer) error

type ScopeOption func(s *Scope)

func getCaster(s *Scope, fromType, toType reflect.Type) castFunc {
	cacheKey := casterKey{fromTypePtr: typePtr(fromType), toTypePtr: typePtr(toType)}
	s.mu.RLock()
	caster, ok := s.casterMap[cacheKey]
	s.mu.RUnlock()
	if ok {
		return caster
	}
	caster = newCaster(s, fromType, toType)
	s.mu.Lock()
	s.casterMap[cacheKey] = caster
	s.mu.Unlock()
	return caster
}

func newCaster(s *Scope, fromType, toType reflect.Type) castFunc {
	if !s.disableZeroCopy && isMemSame(fromType, toType) {
		size := fromType.Size()
		return func(fromAddr unsafe.Pointer, toAddr unsafe.Pointer) error {
			memCopy(toAddr, fromAddr, size)
			return nil
		}
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
		return nil
	}
}

func castStringToTime(s string) (time.Time, error) {
	for _, format := range timeFormats {
		t, err := time.Parse(format, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("failed to parse %s as time", s)
}

func castStringToDuration(s string) (time.Duration, error) {
	if strings.ContainsAny(s, "nuµmsh") {
		return time.ParseDuration(s)
	}
	v, err := strconv.ParseInt(s, 10, 64)
	return time.Duration(v), err
}
