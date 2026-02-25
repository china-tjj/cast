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

type iNumber interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64
}

func getNumberCaster[T iNumber](s *Scope, fromType, toType reflect.Type) (castFunc, uint8) {
	switch fromType.Kind() {
	case reflect.Bool:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if *(*bool)(fromAddr) {
				*(*T)(toAddr) = 1
			} else {
				*(*T)(toAddr) = 0
			}
			return nil
		}, 0
	case reflect.Int:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*T)(toAddr) = T(*(*int)(fromAddr))
			return nil
		}, 0
	case reflect.Int8:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*T)(toAddr) = T(*(*int8)(fromAddr))
			return nil
		}, 0
	case reflect.Int16:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*T)(toAddr) = T(*(*int16)(fromAddr))
			return nil
		}, 0
	case reflect.Int32:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*T)(toAddr) = T(*(*int32)(fromAddr))
			return nil
		}, 0
	case reflect.Int64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*T)(toAddr) = T(*(*int64)(fromAddr))
			return nil
		}, 0
	case reflect.Uint:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*T)(toAddr) = T(*(*uint)(fromAddr))
			return nil
		}, 0
	case reflect.Uint8:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*T)(toAddr) = T(*(*uint8)(fromAddr))
			return nil
		}, 0
	case reflect.Uint16:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*T)(toAddr) = T(*(*uint16)(fromAddr))
			return nil
		}, 0
	case reflect.Uint32:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*T)(toAddr) = T(*(*uint32)(fromAddr))
			return nil
		}, 0
	case reflect.Uint64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*T)(toAddr) = T(*(*uint64)(fromAddr))
			return nil
		}, 0
	case reflect.Uintptr:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*T)(toAddr) = T(*(*uintptr)(fromAddr))
			return nil
		}, 0
	case reflect.Float32:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*T)(toAddr) = T(*(*float32)(fromAddr))
			return nil
		}, 0
	case reflect.Float64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*T)(toAddr) = T(*(*float64)(fromAddr))
			return nil
		}, 0
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Pointer:
		if toType.Kind() != reflect.Uintptr {
			return getAddressingPointerCaster(s, fromType, toType)
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*uintptr)(toAddr) = uintptr(*(*unsafe.Pointer)(fromAddr))
			return nil
		}, 0
	case reflect.String:
		toBitSize := int(8 * toType.Size())
		switch toType.Kind() {
		case reflect.Float32, reflect.Float64:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				f64, err := strconv.ParseFloat(*(*string)(fromAddr), toBitSize)
				if err != nil {
					return err
				}
				*(*T)(toAddr) = T(f64)
				return nil
			}, 0
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				i64, err := strconv.ParseInt(*(*string)(fromAddr), 10, toBitSize)
				if err != nil {
					return err
				}
				*(*T)(toAddr) = T(i64)
				return nil
			}, 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				ui64, err := strconv.ParseUint(*(*string)(fromAddr), 10, toBitSize)
				if err != nil {
					return err
				}
				*(*T)(toAddr) = T(ui64)
				return nil
			}, 0
		default:
			return nil, 0
		}
	case reflect.UnsafePointer:
		if toType.Kind() != reflect.Uintptr {
			return nil, 0
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*uintptr)(toAddr) = uintptr(*(*unsafe.Pointer)(fromAddr))
			return nil
		}, 0
	default:
		return nil, 0
	}
}
