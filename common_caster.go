// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getUnpackInterfaceCaster(s *Scope, fromType, toType reflect.Type) castFunc {
	fromTypeNumMethod := fromType.NumMethod()
	return func(fromAddr, toAddr unsafe.Pointer) error {
		fromElemType, fromElemAddr := unpackInterface(fromTypeNumMethod, fromAddr)
		elemCaster := getCaster(s, fromElemType, toType)
		if elemCaster == nil {
			return invalidCastErr(fromElemType, toType)
		}
		return elemCaster(fromElemAddr, toAddr)
	}
}

func getAddressingPointerCaster(s *Scope, fromType, toType reflect.Type) castFunc {
	elemCaster := getCaster(s, fromType.Elem(), toType)
	if elemCaster == nil {
		return nil
	}
	return func(fromAddr, toAddr unsafe.Pointer) error {
		fromAddr = *(*unsafe.Pointer)(fromAddr)
		if fromAddr == nil {
			return nilPtrErr
		}
		return elemCaster(fromAddr, toAddr)
	}
}

func getFromStringAsSliceCaster(s *Scope, toType reflect.Type) castFunc {
	if toType.Elem().Kind() == reflect.Int32 {
		runesCaster := getCaster(s, runesType, toType)
		if runesCaster == nil {
			return nil
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			runes := []rune(*(*string)(fromAddr))
			return runesCaster(unsafe.Pointer(&runes), toAddr)
		}
	} else {
		bytesCaster := getCaster(s, bytesType, toType)
		if bytesCaster == nil {
			return nil
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			bytes := toBytes(*(*string)(fromAddr))
			return bytesCaster(unsafe.Pointer(&bytes), toAddr)
		}
	}
}

func getToStringAsSliceCaster(s *Scope, fromType reflect.Type) castFunc {
	if fromType.Elem().Kind() == reflect.Int32 {
		runesCaster := getCaster(s, fromType, runesType)
		if runesCaster == nil {
			return nil
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			var runes []rune
			if err := runesCaster(fromAddr, unsafe.Pointer(&runes)); err != nil {
				return err
			}
			*(*string)(toAddr) = string(runes)
			return nil
		}
	} else {
		bytesCaster := getCaster(s, fromType, bytesType)
		if bytesCaster == nil {
			return nil
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			var bytes []byte
			if err := bytesCaster(fromAddr, unsafe.Pointer(&bytes)); err != nil {
				return err
			}
			*(*string)(toAddr) = toString(bytes)
			return nil
		}
	}
}

// 在无法转为 []byte / [N]byte / []rune / [N]rune 时，尝试先转为 string，再转为目标类型
func getStringAsBridgeCaster(s *Scope, fromType, toType reflect.Type) castFunc {
	toElemKind := toType.Elem().Kind()
	if toElemKind != reflect.Uint8 && toElemKind != reflect.Int32 {
		return nil
	}
	toStringCaster := getCaster(s, fromType, stringType)
	if toStringCaster == nil {
		return nil
	}
	toTargetCaster := getFromStringAsSliceCaster(s, toType)
	if toTargetCaster == nil {
		return nil
	}
	return func(fromAddr, toAddr unsafe.Pointer) error {
		var st string
		err := toStringCaster(fromAddr, unsafe.Pointer(&st))
		if err != nil {
			return err
		}
		err = toTargetCaster(unsafe.Pointer(&st), toAddr)
		if err != nil {
			return err
		}
		return nil
	}
}
