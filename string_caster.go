// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"fmt"
	"reflect"
	"strconv"
	"unsafe"
)

func getStringCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	if fromType == nil {
		return nil, false
	}
	if fromType.Implements(stringerType) {
		fromTypeIsPtr := isPtrType(fromType)
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if fromTypeIsPtr {
				fromAddr = *(*unsafe.Pointer)(fromAddr)
				if fromAddr == nil {
					return nilPtrErr
				}
			}
			from, _ := packEface(fromType, fromAddr).(fmt.Stringer)
			if from == nil {
				return nilStringerErr
			}
			*(*string)(toAddr) = from.String()
			return nil
		}, false
	}
	switch fromType.Kind() {
	case reflect.Bool:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatBool(*(*bool)(fromAddr))
			return nil
		}, false
	case reflect.Int:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatInt(int64(*(*int)(fromAddr)), 10)
			return nil
		}, false
	case reflect.Int8:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatInt(int64(*(*int8)(fromAddr)), 10)
			return nil
		}, false
	case reflect.Int16:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatInt(int64(*(*int16)(fromAddr)), 10)
			return nil
		}, false
	case reflect.Int32:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatInt(int64(*(*int32)(fromAddr)), 10)
			return nil
		}, false
	case reflect.Int64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatInt(*(*int64)(fromAddr), 10)
			return nil
		}, false
	case reflect.Uint:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatUint(uint64(*(*uint)(fromAddr)), 10)
			return nil
		}, false
	case reflect.Uint8:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatUint(uint64(*(*uint8)(fromAddr)), 10)
			return nil
		}, false
	case reflect.Uint16:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatUint(uint64(*(*uint16)(fromAddr)), 10)
			return nil
		}, false
	case reflect.Uint32:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatUint(uint64(*(*uint32)(fromAddr)), 10)
			return nil
		}, false
	case reflect.Uint64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatUint(*(*uint64)(fromAddr), 10)
			return nil
		}, false
	case reflect.Uintptr:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatUint(uint64(*(*uintptr)(fromAddr)), 10)
			return nil
		}, false
	case reflect.Float32:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatFloat(float64(*(*float32)(fromAddr)), 'g', -1, 32)
			return nil
		}, false
	case reflect.Float64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatFloat(*(*float64)(fromAddr), 'g', -1, 64)
			return nil
		}, false
	case reflect.Complex64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatComplex(complex128(*(*complex64)(fromAddr)), 'g', -1, 64)
			return nil
		}, false
	case reflect.Complex128:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatComplex(*(*complex128)(fromAddr), 'g', -1, 128)
			return nil
		}, false
	case reflect.Array, reflect.Slice:
		return getToStringAsSliceCaster(s, fromType)
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Pointer:
		return getAddressingPointerCaster(s, fromType, toType)
	case reflect.String:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = *(*string)(fromAddr)
			return nil
		}, false
	default:
		return nil, false
	}
}
