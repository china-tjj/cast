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

type fastEfaceUnpacker func(from any, toAddr unsafe.Pointer) (error, bool)

var unpackerNewerMap = make(map[unsafe.Pointer]func(s *Scope) fastEfaceUnpacker)

func registerFastEFaceUnpacker[T any]() {
	toType := typeFor[T]()
	isSimple := !isRefType(toType)
	unpackerNewerMap[typePtr(toType)] = func(s *Scope) fastEfaceUnpacker {
		if !s.deepCopy || isSimple {
			return func(from any, toAddr unsafe.Pointer) (error, bool) {
				var ok bool
				*(*T)(toAddr), ok = from.(T)
				return nil, ok
			}
		}

		copier, _ := getCaster(s, toType, toType)
		if copier == nil {
			err := invalidCastErr(s, toType, toType)
			return func(from any, toAddr unsafe.Pointer) (error, bool) {
				_, ok := from.(T)
				if !ok {
					return nil, false
				}
				return err, true
			}
		}
		return func(from any, toAddr unsafe.Pointer) (error, bool) {
			to, ok := from.(T)
			if !ok {
				return nil, false
			}
			return copier(noEscape(unsafe.Pointer(&to)), toAddr), true
		}
	}
}

func getFastEFaceUnpacker(s *Scope, toType reflect.Type) fastEfaceUnpacker {
	if newer := unpackerNewerMap[typePtr(toType)]; newer != nil {
		return newer(s)
	}
	return nil
}

func init() {
	registerFastEFaceUnpacker[bool]()
	registerFastEFaceUnpacker[int]()
	registerFastEFaceUnpacker[int8]()
	registerFastEFaceUnpacker[int16]()
	registerFastEFaceUnpacker[int32]()
	registerFastEFaceUnpacker[int64]()
	registerFastEFaceUnpacker[uint]()
	registerFastEFaceUnpacker[uint8]()
	registerFastEFaceUnpacker[uint16]()
	registerFastEFaceUnpacker[uint32]()
	registerFastEFaceUnpacker[uint64]()
	registerFastEFaceUnpacker[uintptr]()
	registerFastEFaceUnpacker[float32]()
	registerFastEFaceUnpacker[float64]()
	registerFastEFaceUnpacker[string]()
	registerFastEFaceUnpacker[[]byte]()
	registerFastEFaceUnpacker[[]rune]()
	registerFastEFaceUnpacker[map[string]any]()
	registerFastEFaceUnpacker[[]any]()
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

func getUnpackInterfaceCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	fromTypeNumMethod := fromType.NumMethod()
	fastUnpacker := getFastEFaceUnpacker(s, toType)
	// 每次拆箱后都需动态查 map 获取 caster，略微有点慢，加一层 cache
	var c cache
	return func(fromAddr, toAddr unsafe.Pointer) error {
		from := loadEface(fromTypeNumMethod, fromAddr)
		if fastUnpacker != nil {
			if err, ok := fastUnpacker(from, toAddr); ok {
				return err
			}
		}
		fromElemTypePtr, fromElemAddr := unpackEface(from)
		fromElemType := typePtrToType(fromElemTypePtr)
		key := casterKey{fromTypePtr: fromElemTypePtr, toTypePtr: typePtr(toType)}
		var elemCaster castFunc
		var elemHasRef bool
		if value, ok := c.load(key); ok {
			elemCaster, elemHasRef = value.caster, value.hasRef
		} else {
			elemCaster, elemHasRef = getCaster(s, fromElemType, toType)
			value = casterValue{caster: elemCaster, hasRef: elemHasRef}
			c.store(key, value)
		}
		if elemCaster == nil {
			return invalidCastErr(s, fromElemType, toType)
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
	}, false
}

func getAddressingPointerCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	elemCaster, _ := getCaster(s, fromType.Elem(), toType)
	if elemCaster == nil {
		return nil, false
	}
	return func(fromAddr, toAddr unsafe.Pointer) error {
		fromAddr = *(*unsafe.Pointer)(fromAddr)
		if fromAddr == nil {
			return nilPtrErr
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
