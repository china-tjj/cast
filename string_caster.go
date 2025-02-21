// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"strconv"
	"unsafe"
)

func getStringCaster(fromType, toType reflect.Type) castFunc {
	switch fromType.Kind() {
	case reflect.Bool:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*string)(toAddr) = strconv.FormatBool(*(*bool)(fromAddr))
			return true
		}
	case reflect.Int:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*string)(toAddr) = strconv.FormatInt(int64(*(*int)(fromAddr)), 10)
			return true
		}
	case reflect.Int8:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*string)(toAddr) = strconv.FormatInt(int64(*(*int8)(fromAddr)), 10)
			return true
		}
	case reflect.Int16:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*string)(toAddr) = strconv.FormatInt(int64(*(*int16)(fromAddr)), 10)
			return true
		}
	case reflect.Int32:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*string)(toAddr) = strconv.FormatInt(int64(*(*int32)(fromAddr)), 10)
			return true
		}
	case reflect.Int64:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*string)(toAddr) = strconv.FormatInt(*(*int64)(fromAddr), 10)
			return true
		}
	case reflect.Uint:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*string)(toAddr) = strconv.FormatUint(uint64(*(*uint)(fromAddr)), 10)
			return true
		}
	case reflect.Uint8:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*string)(toAddr) = strconv.FormatUint(uint64(*(*uint8)(fromAddr)), 10)
			return true
		}
	case reflect.Uint16:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*string)(toAddr) = strconv.FormatUint(uint64(*(*uint16)(fromAddr)), 10)
			return true
		}
	case reflect.Uint32:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*string)(toAddr) = strconv.FormatUint(uint64(*(*uint32)(fromAddr)), 10)
			return true
		}
	case reflect.Uint64:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*string)(toAddr) = strconv.FormatUint(*(*uint64)(fromAddr), 10)
			return true
		}
	case reflect.Uintptr:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*string)(toAddr) = strconv.FormatUint(uint64(*(*uintptr)(fromAddr)), 10)
			return true
		}
	case reflect.Float32:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*string)(toAddr) = strconv.FormatFloat(float64(*(*float32)(fromAddr)), 'g', -1, 32)
			return true
		}
	case reflect.Float64:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*string)(toAddr) = strconv.FormatFloat(*(*float64)(fromAddr), 'g', -1, 64)
			return true
		}
	case reflect.Array, reflect.Slice:
		bytesCaster := getCaster(fromType, typeFor[[]byte]())
		if bytesCaster == nil {
			return nil
		}
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			var bytes []byte
			ok := bytesCaster(fromAddr, unsafe.Pointer(&bytes))
			if !ok {
				return false
			}
			*(*string)(toAddr) = toString(bytes)
			return true
		}
	case reflect.Interface:
		return getUnwrapInterfaceCaster(fromType, toType)
	case reflect.Pointer:
		return getAddressingPointerCaster(fromType, toType)
	case reflect.String:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*string)(toAddr) = *(*string)(fromAddr)
			return true
		}
	case reflect.Struct:
		return getUnwrapStructCaster(fromType, toType)
	default:
		return nil
	}
}
