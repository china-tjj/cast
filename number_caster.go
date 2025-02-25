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

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64
}

func getNumberCaster[T Number](fromType, toType reflect.Type) castFunc {
	switch fromType.Kind() {
	case reflect.Bool:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			if *(*bool)(fromAddr) {
				*(*T)(toAddr) = 1
			} else {
				*(*T)(toAddr) = 0
			}
			return true
		}
	case reflect.Int:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*T)(toAddr) = T(*(*int)(fromAddr))
			return true
		}
	case reflect.Int8:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*T)(toAddr) = T(*(*int8)(fromAddr))
			return true
		}
	case reflect.Int16:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*T)(toAddr) = T(*(*int16)(fromAddr))
			return true
		}
	case reflect.Int32:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*T)(toAddr) = T(*(*int32)(fromAddr))
			return true
		}
	case reflect.Int64:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*T)(toAddr) = T(*(*int64)(fromAddr))
			return true
		}
	case reflect.Uint:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*T)(toAddr) = T(*(*uint)(fromAddr))
			return true
		}
	case reflect.Uint8:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*T)(toAddr) = T(*(*uint8)(fromAddr))
			return true
		}
	case reflect.Uint16:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*T)(toAddr) = T(*(*uint16)(fromAddr))
			return true
		}
	case reflect.Uint32:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*T)(toAddr) = T(*(*uint32)(fromAddr))
			return true
		}
	case reflect.Uint64:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*T)(toAddr) = T(*(*uint64)(fromAddr))
			return true
		}
	case reflect.Uintptr:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*T)(toAddr) = T(*(*uintptr)(fromAddr))
			return true
		}
	case reflect.Float32:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*T)(toAddr) = T(*(*float32)(fromAddr))
			return true
		}
	case reflect.Float64:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*T)(toAddr) = T(*(*float64)(fromAddr))
			return true
		}
	case reflect.Array:
		return getUnpackArrayCaster(fromType, toType)
	case reflect.Interface:
		return getUnpackInterfaceCaster(fromType, toType)
	case reflect.Pointer:
		if toType.Kind() != reflect.Uintptr {
			return getAddressingPointerCaster(fromType, toType)
		}
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*uintptr)(toAddr) = uintptr(*(*unsafe.Pointer)(fromAddr))
			return true
		}
	case reflect.Slice:
		return getUnpackSliceCaster(fromType, toType)
	case reflect.String:
		switch toType.Kind() {
		case reflect.Float32, reflect.Float64:
			return func(fromAddr, toAddr unsafe.Pointer) bool {
				f64, err := strconv.ParseFloat(*(*string)(fromAddr), 10)
				if err != nil {
					return false
				}
				*(*T)(toAddr) = T(f64)
				return true
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return func(fromAddr, toAddr unsafe.Pointer) bool {
				i64, err := strconv.ParseInt(*(*string)(fromAddr), 10, 64)
				if err != nil {
					return false
				}
				*(*T)(toAddr) = T(i64)
				return true
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return func(fromAddr, toAddr unsafe.Pointer) bool {
				ui64, err := strconv.ParseUint(*(*string)(fromAddr), 10, 64)
				if err != nil {
					return false
				}
				*(*T)(toAddr) = T(ui64)
				return true
			}
		default:
			return nil
		}
	case reflect.Struct:
		return getUnpackStructCaster(fromType, toType)
	case reflect.UnsafePointer:
		if toType.Kind() != reflect.Uintptr {
			return nil
		}
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*uintptr)(toAddr) = uintptr(*(*unsafe.Pointer)(fromAddr))
			return true
		}
	default:
		return nil
	}
}
