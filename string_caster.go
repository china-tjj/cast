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

func getStringCaster(s *Scope, fromType, toType reflect.Type) (castFunc, uint8) {
	if fromType.Implements(stringerType) {
		fromTypeIsPtr := isPtrType(fromType)
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if fromTypeIsPtr {
				fromAddr = *(*unsafe.Pointer)(fromAddr)
				if fromAddr == nil {
					return NilStringerErr
				}
			}
			from, ok := packEface(fromType, fromAddr).(fmt.Stringer)
			if !ok || from == nil {
				return NilStringerErr
			}
			*(*string)(toAddr) = from.String()
			return nil
		}, 0
	}
	switch fromType.Kind() {
	case reflect.Bool:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatBool(*(*bool)(fromAddr))
			return nil
		}, 0
	case reflect.Int:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatInt(int64(*(*int)(fromAddr)), 10)
			return nil
		}, 0
	case reflect.Int8:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatInt(int64(*(*int8)(fromAddr)), 10)
			return nil
		}, 0
	case reflect.Int16:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatInt(int64(*(*int16)(fromAddr)), 10)
			return nil
		}, 0
	case reflect.Int32:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatInt(int64(*(*int32)(fromAddr)), 10)
			return nil
		}, 0
	case reflect.Int64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatInt(*(*int64)(fromAddr), 10)
			return nil
		}, 0
	case reflect.Uint:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatUint(uint64(*(*uint)(fromAddr)), 10)
			return nil
		}, 0
	case reflect.Uint8:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatUint(uint64(*(*uint8)(fromAddr)), 10)
			return nil
		}, 0
	case reflect.Uint16:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatUint(uint64(*(*uint16)(fromAddr)), 10)
			return nil
		}, 0
	case reflect.Uint32:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatUint(uint64(*(*uint32)(fromAddr)), 10)
			return nil
		}, 0
	case reflect.Uint64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatUint(*(*uint64)(fromAddr), 10)
			return nil
		}, 0
	case reflect.Uintptr:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatUint(uint64(*(*uintptr)(fromAddr)), 10)
			return nil
		}, 0
	case reflect.Float32:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatFloat(float64(*(*float32)(fromAddr)), 'g', -1, 32)
			return nil
		}, 0
	case reflect.Float64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatFloat(*(*float64)(fromAddr), 'g', -1, 64)
			return nil
		}, 0
	case reflect.Complex64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatComplex(complex128(*(*complex64)(fromAddr)), 'g', -1, 64)
			return nil
		}, 0
	case reflect.Complex128:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = strconv.FormatComplex(*(*complex128)(fromAddr), 'g', -1, 128)
			return nil
		}, 0
	case reflect.Array:
		switch fromType.Elem().Kind() {
		case reflect.Int32:
			length := fromType.Len()
			return func(fromAddr, toAddr unsafe.Pointer) error {
				fromPtr := &slice{
					data: fromAddr,
					len:  length,
					cap:  length,
				}
				*(*string)(toAddr) = string(*(*[]rune)(unsafe.Pointer(fromPtr)))
				return nil
			}, 0
		case reflect.Uint8:
			length := fromType.Len()
			if s.deepCopy || s.disableZeroCopy {
				return func(fromAddr, toAddr unsafe.Pointer) error {
					toPtr := (*str)(toAddr)
					toPtr.data = copyObject(fromType, fromAddr)
					toPtr.len = length
					return nil
				}, 0
			}
			return func(fromAddr, toAddr unsafe.Pointer) error {
				toPtr := (*str)(toAddr)
				toPtr.data = fromAddr
				toPtr.len = length
				return nil
			}, flagHasRef
		default:
			return nil, 0
		}
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Pointer:
		return getAddressingPointerCaster(s, fromType, toType)
	case reflect.Slice:
		switch fromType.Elem().Kind() {
		case reflect.Int32:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*string)(toAddr) = string(*(*[]rune)(fromAddr))
				return nil
			}, 0
		case reflect.Uint8:
			if s.deepCopy || s.disableZeroCopy {
				return func(fromAddr, toAddr unsafe.Pointer) error {
					// 在任何go版本，这样[]byte转string都会拷贝
					*(*string)(toAddr) = string(*(*[]byte)(fromAddr))
					return nil
				}, 0
			}
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*string)(toAddr) = toString(*(*[]byte)(fromAddr))
				return nil
			}, 0
		default:
			return nil, 0
		}
	case reflect.String:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*string)(toAddr) = *(*string)(fromAddr)
			return nil
		}, 0
	default:
		return nil, 0
	}
}
