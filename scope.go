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

func (s *Scope) DisableZeroCopy() bool {
	return s.disableZeroCopy
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
type castFunc func(s *Scope, fromAddr, toAddr unsafe.Pointer) error

type ScopeOption func(s *Scope)

func getCaster(s *Scope, fromType, toType reflect.Type) castFunc {
	key := casterKey{fromTypePtr: typePtr(fromType), toTypePtr: typePtr(toType)}
	s.mu.RLock()
	caster, ok := s.casterMap[key]
	s.mu.RUnlock()
	if ok {
		return caster
	}
	caster = newCaster(s, fromType, toType)
	s.mu.Lock()
	s.casterMap[key] = caster
	s.mu.Unlock()
	return caster
}

func newCaster(s *Scope, fromType, toType reflect.Type) castFunc {
	if isRefAble(s, fromType, toType) {
		fromTypePtr := typePtr(fromType)
		return func(s *Scope, fromAddr, toAddr unsafe.Pointer) error {
			typedmemmove(fromTypePtr, toAddr, fromAddr)
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

func getAnyToTCaster[T any]() func(s *Scope, from any) (to T, err error) {
	toType := typeFor[T]()
	return func(s *Scope, from any) (to T, err error) {
		if t, ok := from.(T); ok {
			return t, nil
		}
		unpackCaster := getUnpackInterfaceCaster(s, anyType, toType)
		err = unpackCaster(s, noEscape(unsafe.Pointer(&from)), noEscape(unsafe.Pointer(&to)))
		return to, err
	}
}

var defaultOptions = []ScopeOption{
	WithCaster(castStringToTime),
	WithCaster(castStringToDuration),
	WithCaster(getAnyToTCaster[bool]()),
	WithCaster(getAnyToTCaster[int]()),
	WithCaster(getAnyToTCaster[int8]()),
	WithCaster(getAnyToTCaster[int16]()),
	WithCaster(getAnyToTCaster[int32]()),
	WithCaster(getAnyToTCaster[int64]()),
	WithCaster(getAnyToTCaster[uint]()),
	WithCaster(getAnyToTCaster[uint8]()),
	WithCaster(getAnyToTCaster[uint16]()),
	WithCaster(getAnyToTCaster[uint32]()),
	WithCaster(getAnyToTCaster[uint64]()),
	WithCaster(getAnyToTCaster[uintptr]()),
	WithCaster(getAnyToTCaster[float32]()),
	WithCaster(getAnyToTCaster[float64]()),
	WithCaster(getAnyToTCaster[string]()),
	WithCaster(getAnyToTCaster[[]byte]()),
	WithCaster(getAnyToTCaster[[]rune]()),
	WithCaster(getAnyToTCaster[map[string]any]()),
	WithCaster(getAnyToTCaster[[]any]()),
}
