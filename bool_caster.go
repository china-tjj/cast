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

func getBoolCaster(fromType, toType reflect.Type) castFunc {
	switch fromType.Kind() {
	case reflect.Bool:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*bool)(toAddr) = *(*bool)(fromAddr)
			return true
		}
	case reflect.Int:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			if *(*int)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return true
		}
	case reflect.Int8:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			if *(*int8)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return true
		}
	case reflect.Int16:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			if *(*int16)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return true
		}
	case reflect.Int32:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			if *(*int32)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return true
		}
	case reflect.Int64:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			if *(*int64)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return true
		}
	case reflect.Uint:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			if *(*uint)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return true
		}
	case reflect.Uint8:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			if *(*uint8)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return true
		}
	case reflect.Uint16:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			if *(*uint16)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return true
		}
	case reflect.Uint32:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			if *(*uint32)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return true
		}
	case reflect.Uint64:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			if *(*uint64)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return true
		}
	case reflect.Uintptr:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			if *(*uintptr)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return true
		}
	case reflect.Float32:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			if *(*float32)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return true
		}
	case reflect.Float64:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			if *(*float64)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return true
		}
	case reflect.Array:
		return getUnwrapArrayCaster(fromType, toType)
	case reflect.Interface:
		return getUnwrapInterfaceCaster(fromType, toType)
	case reflect.Pointer:
		return getAddressingPointerCaster(fromType, toType)
	case reflect.Slice:
		return getUnwrapSliceCaster(fromType, toType)
	case reflect.String:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			from := *(*string)(fromAddr)
			res, err := strconv.ParseBool(from)
			if err == nil {
				*(*bool)(toAddr) = res
			} else {
				*(*bool)(toAddr) = len(from) > 0
			}
			return true
		}
	case reflect.Struct:
		return getUnwrapStructCaster(fromType, toType)
	default:
		return nil
	}
}
