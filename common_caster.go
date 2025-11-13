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

func init() {
	registerFastInterfaceCaster[bool]()
	registerFastInterfaceCaster[int]()
	registerFastInterfaceCaster[int8]()
	registerFastInterfaceCaster[int16]()
	registerFastInterfaceCaster[int32]()
	registerFastInterfaceCaster[int64]()
	registerFastInterfaceCaster[uint]()
	registerFastInterfaceCaster[uint8]()
	registerFastInterfaceCaster[uint16]()
	registerFastInterfaceCaster[uint32]()
	registerFastInterfaceCaster[uint64]()
	registerFastInterfaceCaster[uintptr]()
	registerFastInterfaceCaster[float32]()
	registerFastInterfaceCaster[float64]()
	registerFastInterfaceCaster[string]()
	registerFastInterfaceCaster[[]byte]()
	registerFastInterfaceCaster[[]rune]()
	registerFastInterfaceCaster[map[string]any]()
	registerFastInterfaceCaster[[]any]()
}

const cacheSize = 8

type cache struct {
	mu     sync.RWMutex
	keys   [cacheSize]casterKey
	values [cacheSize]casterValue
	n      int
}

func (c *cache) load(key casterKey) (casterValue, bool) {
	if !c.mu.TryRLock() {
		return casterValue{}, false
	}
	defer c.mu.RUnlock()
	n := min(cacheSize, c.n)
	for i := 0; i < n; i++ {
		if c.keys[i] == key {
			return c.values[i], true
		}
	}
	return casterValue{}, false
}

func (c *cache) store(key casterKey, value casterValue) {
	if !c.mu.TryLock() {
		return
	}
	defer c.mu.Unlock()
	n := min(cacheSize, c.n)
	for i := 0; i < n; i++ {
		if c.keys[i] == key {
			c.values[i] = value
			return
		}
	}
	i := c.n % cacheSize
	c.keys[i] = key
	c.values[i] = value
	c.n++
	if c.n > 2*cacheSize {
		c.n -= cacheSize
	}
}

// 不用闭包，手动封装结构体，以支持内联
type interfaceCastHandler struct {
	s                 *Scope
	c                 cache // 每次拆箱后都需动态查 map 获取 caster，略微有点慢，加一层 cache
	fromType          reflect.Type
	toType            reflect.Type
	fromTypeNumMethod int
	toTypePtr         unsafe.Pointer
	toTypeIsNilable   bool
}

func newInterfaceCastHandler(s *Scope, fromType, toType reflect.Type) *interfaceCastHandler {
	return &interfaceCastHandler{
		s:                 s,
		fromType:          fromType,
		toType:            toType,
		fromTypeNumMethod: fromType.NumMethod(),
		toTypePtr:         typePtr(toType),
		toTypeIsNilable:   isNilableType(toType),
	}
}

func (handler *interfaceCastHandler) normalCast(from any, toAddr unsafe.Pointer) error {
	fromElemTypePtr, fromElemAddr := unpackEface(from)
	fromElemType := typePtrToType(fromElemTypePtr)
	if fromElemTypePtr == nil {
		if handler.s.strictNilCheck && !handler.toTypeIsNilable {
			return invalidCastErr(handler.s, fromElemType, handler.toType)
		}
		return nil
	}
	key := casterKey{fromTypePtr: fromElemTypePtr, toTypePtr: handler.toTypePtr}
	var elemCaster castFunc
	var elemHasRef bool
	if value, ok := handler.c.load(key); ok {
		elemCaster, elemHasRef = value.caster, value.hasRef
	} else {
		elemCaster, elemHasRef = getCaster(handler.s, fromElemType, handler.toType)
		value = casterValue{caster: elemCaster, hasRef: elemHasRef}
		handler.c.store(key, value)
	}
	if elemCaster == nil {
		return invalidCastErr(handler.s, fromElemType, handler.toType)
	}
	trueFromElemAddr := fromElemAddr
	if isPtrType(fromElemType) {
		// 接口存指针时，eface 的指针会指针拷贝这个指针，故需要再取一次地址
		trueFromElemAddr = noEscape(unsafe.Pointer(&fromElemAddr))
	} else if elemHasRef {
		// 接口存非指针时，eface 的指针指向的内存是只读的，若有指针指向该块内存，需拷贝
		trueFromElemAddr = copyObject(fromElemType, trueFromElemAddr)
	}
	return elemCaster(trueFromElemAddr, toAddr)
}

func (handler *interfaceCastHandler) getCaster() castFunc {
	if newer := fastInterfaceCasterNewerMap[handler.toTypePtr]; newer != nil {
		return newer(handler)
	}
	return func(fromAddr, toAddr unsafe.Pointer) error {
		return handler.normalCast(loadEface(handler.fromTypeNumMethod, fromAddr), toAddr)
	}
}

var fastInterfaceCasterNewerMap = make(map[unsafe.Pointer]func(handler *interfaceCastHandler) castFunc)

func registerFastInterfaceCaster[T any]() {
	toType := typeFor[T]()
	isSimple := !isRefType(toType)
	fastInterfaceCasterNewerMap[typePtr(toType)] = func(handler *interfaceCastHandler) castFunc {
		if !handler.s.deepCopy || isSimple {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				from := loadEface(handler.fromTypeNumMethod, fromAddr)
				var ok bool
				if *(*T)(toAddr), ok = from.(T); ok {
					return nil
				}
				return handler.normalCast(from, toAddr)
			}
		}
		copier, _ := getCaster(handler.s, toType, toType)
		if copier == nil {
			err := invalidCastErr(handler.s, toType, toType)
			return func(fromAddr, toAddr unsafe.Pointer) error {
				from := loadEface(handler.fromTypeNumMethod, fromAddr)
				if _, ok := from.(T); ok {
					return err
				}
				return handler.normalCast(from, toAddr)
			}
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			from := loadEface(handler.fromTypeNumMethod, fromAddr)
			if to, ok := from.(T); ok {
				return copier(noEscape(unsafe.Pointer(&to)), toAddr)
			}
			return handler.normalCast(from, toAddr)
		}
	}
}

func getUnpackInterfaceCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	return newInterfaceCastHandler(s, fromType, toType).getCaster(), false
}

func getAddressingPointerCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	depth, finalElemType := getFinalElem(fromType)
	elemCaster, _ := getCaster(s, finalElemType, toType)
	if elemCaster == nil {
		return nil, false
	}
	toTypeIsNilable := isNilableType(toType)
	return func(fromAddr, toAddr unsafe.Pointer) error {
		for i := depth; i > 0; i-- {
			fromAddr = *(*unsafe.Pointer)(fromAddr)
			if fromAddr == nil {
				if s.strictNilCheck && !toTypeIsNilable {
					return NilPtrErr
				}
				return nil
			}
		}
		return elemCaster(fromAddr, toAddr)
	}, false
}

func getFromStringAsSliceCaster(s *Scope, toType reflect.Type) (castFunc, bool) {
	if toType.Elem().Kind() == reflect.Int32 {
		runesCaster, _ := getCaster(s, runesType, toType)
		if runesCaster == nil {
			return nil, false
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			//todo: runes默认逃逸到堆，考虑动态选择是否逃逸
			runes := []rune(*(*string)(fromAddr))
			return runesCaster(unsafe.Pointer(&runes), toAddr)
		}, false
	} else {
		bytesCaster, _ := getCaster(s, bytesType, toType)
		if bytesCaster == nil {
			return nil, false
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			bytes := toBytes(*(*string)(fromAddr))
			return bytesCaster(noEscape(unsafe.Pointer(&bytes)), toAddr)
		}, false
	}
}

func getToStringAsSliceCaster(s *Scope, fromType reflect.Type) (castFunc, bool) {
	if fromType.Elem().Kind() == reflect.Int32 {
		runesCaster, _ := getCaster(s, fromType, runesType)
		if runesCaster == nil {
			return nil, false
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			var runes []rune
			if err := runesCaster(fromAddr, noEscape(unsafe.Pointer(&runes))); err != nil {
				return err
			}
			*(*string)(toAddr) = string(runes)
			return nil
		}, false
	} else {
		bytesCaster, hasRef := getCaster(s, fromType, bytesType)
		if bytesCaster == nil {
			return nil, false
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			var bytes []byte
			if err := bytesCaster(fromAddr, noEscape(unsafe.Pointer(&bytes))); err != nil {
				return err
			}
			*(*string)(toAddr) = toString(bytes)
			return nil
		}, hasRef
	}
}

// 在无法转为 []byte / [N]byte / []rune / [N]rune 时，尝试先转为 string，再转为目标类型
func getStringAsBridgeCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	toElemKind := toType.Elem().Kind()
	if toElemKind != reflect.Uint8 && toElemKind != reflect.Int32 {
		return nil, false
	}
	toStringCaster, toStringHasRef := getCaster(s, fromType, stringType)
	if toStringCaster == nil {
		return nil, false
	}
	toTargetCaster, _ := getFromStringAsSliceCaster(s, toType)
	if toTargetCaster == nil {
		return nil, false
	}
	hasRef := toStringHasRef && toElemKind == reflect.Uint8
	return func(fromAddr, toAddr unsafe.Pointer) error {
		var st string
		err := toStringCaster(fromAddr, noEscape(unsafe.Pointer(&st)))
		if err != nil {
			return err
		}
		err = toTargetCaster(noEscape(unsafe.Pointer(&st)), toAddr)
		if err != nil {
			return err
		}
		return nil
	}, hasRef
}
