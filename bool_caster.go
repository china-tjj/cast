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

func getBoolCaster(s *Scope, fromType, toType reflect.Type) castFunc {
	switch fromType.Kind() {
	case reflect.Bool:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*bool)(toAddr) = *(*bool)(fromAddr)
			return nil
		}
	case reflect.Int:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if *(*int)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return nil
		}
	case reflect.Int8:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if *(*int8)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return nil
		}
	case reflect.Int16:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if *(*int16)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return nil
		}
	case reflect.Int32:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if *(*int32)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return nil
		}
	case reflect.Int64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if *(*int64)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return nil
		}
	case reflect.Uint:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if *(*uint)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return nil
		}
	case reflect.Uint8:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if *(*uint8)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return nil
		}
	case reflect.Uint16:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if *(*uint16)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return nil
		}
	case reflect.Uint32:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if *(*uint32)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return nil
		}
	case reflect.Uint64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if *(*uint64)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return nil
		}
	case reflect.Uintptr:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if *(*uintptr)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return nil
		}
	case reflect.Float32:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if *(*float32)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return nil
		}
	case reflect.Float64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if *(*float64)(fromAddr) != 0 {
				*(*bool)(toAddr) = true
			} else {
				*(*bool)(toAddr) = false
			}
			return nil
		}
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Pointer:
		return getAddressingPointerCaster(s, fromType, toType)
	case reflect.String:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			from := *(*string)(fromAddr)
			res, err := strconv.ParseBool(from)
			if err != nil {
				return err
			}
			*(*bool)(toAddr) = res
			return nil
		}
	default:
		return nil
	}
}
