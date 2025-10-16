// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getUnpackInterfaceCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	fromTypeNumMethod := fromType.NumMethod()
	return func(fromAddr, toAddr unsafe.Pointer) error {
		fromElemType, fromElemAddr, isPtr := unpackInterface(fromTypeNumMethod, fromAddr)
		elemCaster, elemHasRef := getCaster(s, fromElemType, toType)
		if elemCaster == nil {
			return invalidCastErr(fromElemType, toType)
		}
		trueFromElemAddr := fromElemAddr
		if isPtr {
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
