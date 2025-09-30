// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getUnpackInterfaceCaster(fromType, toType reflect.Type) castFunc {
	fromTypeNumMethod := fromType.NumMethod()
	return func(fromAddr, toAddr unsafe.Pointer) error {
		fromElemType, fromElemAddr := unpackInterface(fromTypeNumMethod, fromAddr)
		elemCaster := getCaster(fromElemType, toType)
		if elemCaster == nil {
			return invalidCastErr(fromElemType, toType)
		}
		return elemCaster(fromElemAddr, toAddr)
	}
}

func getAddressingPointerCaster(fromType, toType reflect.Type) castFunc {
	elemCaster := getCaster(fromType.Elem(), toType)
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

func getFromStringAsSliceCaster(toType reflect.Type) castFunc {
	if toType.Elem().Kind() == reflect.Int32 {
		runesCaster := getCaster(runesType, toType)
		if runesCaster == nil {
			return nil
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			runes := []rune(*(*string)(fromAddr))
			return runesCaster(unsafe.Pointer(&runes), toAddr)
		}
	} else {
		bytesCaster := getCaster(bytesType, toType)
		if bytesCaster == nil {
			return nil
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			bytes := toBytes(*(*string)(fromAddr))
			return bytesCaster(unsafe.Pointer(&bytes), toAddr)
		}
	}
}

func getToStringAsBytesCaster(fromType reflect.Type) castFunc {
	if fromType.Elem().Kind() == reflect.Int32 {
		runesCaster := getCaster(fromType, runesType)
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
		bytesCaster := getCaster(fromType, bytesType)
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
