// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
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
	registerFastInterfaceCaster[complex64]()
	registerFastInterfaceCaster[complex128]()
	registerFastInterfaceCaster[string]()
	registerFastInterfaceCaster[[]byte]()
	registerFastInterfaceCaster[[]rune]()
	registerFastInterfaceCaster[map[string]any]()
	registerFastInterfaceCaster[[]any]()
}

// 不用闭包，手动封装结构体，以支持内联
type interfaceCastHandler struct {
	s                 *Scope
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
	elemCaster, elemFlag := getCaster(handler.s, fromElemType, handler.toType)
	if elemCaster == nil {
		return invalidCastErr(handler.s, fromElemType, handler.toType)
	}
	trueFromElemAddr := fromElemAddr
	if isPtrType(fromElemType) {
		// 接口存指针时，eface 的指针会指针拷贝这个指针，故需要再取一次地址
		trueFromElemAddr = noEscape(unsafe.Pointer(&fromElemAddr))
	} else if isHasRef(elemFlag) {
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

func getUnpackInterfaceCaster(s *Scope, fromType, toType reflect.Type) (castFunc, uint8) {
	return newInterfaceCastHandler(s, fromType, toType).getCaster(), 0
}

func getAddressingPointerCaster(s *Scope, fromType, toType reflect.Type) (castFunc, uint8) {
	depth, finalElemType := getFinalElem(fromType)
	elemCaster, _ := getCaster(s, finalElemType, toType)
	if elemCaster == nil {
		return nil, 0
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
	}, 0
}
