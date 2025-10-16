// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getFallbackInterfaceCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	return func(fromAddr, toAddr unsafe.Pointer) error {
		reflect.NewAt(toType, toAddr).Elem().Set(reflect.NewAt(fromType, fromAddr).Elem())
		return nil
	}, false
}

func getInterfaceCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	fromKind := fromType.Kind()
	if toType.NumMethod() == 0 {
		switch fromKind {
		case reflect.Bool:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*bool)(fromAddr)
				return nil
			}, false
		case reflect.Int:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*int)(fromAddr)
				return nil
			}, false
		case reflect.Int8:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*int8)(fromAddr)
				return nil
			}, false
		case reflect.Int16:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*int16)(fromAddr)
				return nil
			}, false
		case reflect.Int32:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*int32)(fromAddr)
				return nil
			}, false
		case reflect.Int64:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*int64)(fromAddr)
				return nil
			}, false
		case reflect.Uint:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*uint)(fromAddr)
				return nil
			}, false
		case reflect.Uint8:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*uint8)(fromAddr)
				return nil
			}, false
		case reflect.Uint16:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*uint16)(fromAddr)
				return nil
			}, false
		case reflect.Uint32:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*uint32)(fromAddr)
				return nil
			}, false
		case reflect.Uint64:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*uint64)(fromAddr)
				return nil
			}, false
		case reflect.Uintptr:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*uintptr)(fromAddr)
				return nil
			}, false
		case reflect.Float32:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*float32)(fromAddr)
				return nil
			}, false
		case reflect.Float64:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*float64)(fromAddr)
				return nil
			}, false
		case reflect.Interface:
			if fromType.NumMethod() == 0 {
				return func(fromAddr, toAddr unsafe.Pointer) error {
					*(*any)(toAddr) = *(*any)(fromAddr)
					return nil
				}, false
			}
			return getFallbackInterfaceCaster(s, fromType, toType)
		case reflect.String:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*string)(fromAddr)
				return nil
			}, false
		case reflect.UnsafePointer:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*unsafe.Pointer)(fromAddr)
				return nil
			}, false
		default:
			return getFallbackInterfaceCaster(s, fromType, toType)
		}
	}

	if !fromType.Implements(toType) {
		if fromKind == reflect.Interface {
			return getUnpackInterfaceCaster(s, fromType, toType)
		}
		return nil, false
	}
	return getFallbackInterfaceCaster(s, fromType, toType)
}
