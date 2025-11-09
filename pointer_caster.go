// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getPointerCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	switch fromType.Kind() {
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.UnsafePointer:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*unsafe.Pointer)(toAddr) = *(*unsafe.Pointer)(fromAddr)
			return nil
		}, false
	default:
		fromDepth, fromElemType := getFinalElem(fromType) // fromDepth >= 0
		toDepth, toElemType := getFinalElem(toType)       // toDepth >= 1
		if isRefAble(s, fromElemType, toElemType) {
			if fromDepth < toDepth {
				return func(fromAddr, toAddr unsafe.Pointer) error {
					for d := toDepth - 1; d > fromDepth; d-- {
						ptr := unsafe.Pointer(new(unsafe.Pointer))
						*(*unsafe.Pointer)(toAddr) = ptr
						toAddr = ptr
					}
					*(*unsafe.Pointer)(toAddr) = fromAddr
					return nil
				}, true
			} else {
				// fromDepth >= toDepth
				return func(fromAddr, toAddr unsafe.Pointer) error {
					for d := fromDepth; d > toDepth-1; d-- {
						fromAddr = *(*unsafe.Pointer)(fromAddr)
						if fromAddr == nil {
							*(*unsafe.Pointer)(toAddr) = nil
							return nil
						}
					}
					*(*unsafe.Pointer)(toAddr) = fromAddr
					return nil
				}, false
			}
		}
		var caster castFunc
		var hasRef bool
		fromElemKind := fromElemType.Kind()
		if fromElemKind == reflect.Array && isRefAble(s, fromElemType.Elem(), toElemType) {
			// [N]T -> *T
			toDepth--
			caster, hasRef = arrayToElemPtrCaster, true
		} else if fromElemKind == reflect.Slice && isRefAble(s, fromElemType.Elem(), toElemType) {
			// []T -> *T
			toDepth--
			caster, hasRef = sliceToElemPtrCaster, false
		} else if fromElemKind == reflect.Slice && toElemType.Kind() == reflect.Array && isRefAble(s, fromElemType.Elem(), toElemType.Elem()) {
			// []T -> *[N]T
			toDepth--
			caster, hasRef = getSliceToArrayPtrCaster(toElemType), false
		} else if fromElemKind == reflect.String && isRefAble(s, byteType, toElemType) {
			// string -> *byte
			toDepth--
			caster, hasRef = strToBytePtrCaster, false
		} else if fromElemKind == reflect.String && toElemType.Kind() == reflect.Array && isRefAble(s, byteType, toElemType.Elem()) {
			// string -> *[N]byte
			toDepth--
			caster, hasRef = getStrToArrayPtrCaster(toElemType), false
		} else {
			caster, hasRef = getCaster(s, fromElemType, toElemType)
		}
		if caster == nil {
			return nil, false
		}
		depth := min(fromDepth, toDepth)
		return func(fromAddr, toAddr unsafe.Pointer) error {
			for d := fromDepth; d > toDepth; d-- {
				fromAddr = *(*unsafe.Pointer)(fromAddr)
				if fromAddr == nil {
					*(*unsafe.Pointer)(toAddr) = nil
					return nil
				}
			}
			for d := toDepth; d > fromDepth; d-- {
				if d > 1 {
					ptr := unsafe.Pointer(new(unsafe.Pointer))
					*(*unsafe.Pointer)(toAddr) = ptr
					toAddr = ptr
				} else {
					ptr := newObject(toElemType)
					*(*unsafe.Pointer)(toAddr) = ptr
					toAddr = ptr
				}
			}
			for d := depth; d > 0; d-- {
				fromAddr = *(*unsafe.Pointer)(fromAddr)
				if fromAddr == nil {
					*(*unsafe.Pointer)(toAddr) = nil
					return nil
				}
				if d > 1 {
					ptr := unsafe.Pointer(new(unsafe.Pointer))
					*(*unsafe.Pointer)(toAddr) = ptr
					toAddr = ptr
				} else {
					ptr := newObject(toElemType)
					*(*unsafe.Pointer)(toAddr) = ptr
					toAddr = ptr
				}
			}
			return caster(fromAddr, toAddr)
		}, hasRef && fromDepth <= toDepth && depth == 0
	}
}

func sliceToElemPtrCaster(fromAddr, toAddr unsafe.Pointer) error {
	*(*unsafe.Pointer)(toAddr) = (*slice)(fromAddr).data
	return nil
}

func arrayToElemPtrCaster(fromAddr, toAddr unsafe.Pointer) error {
	*(*unsafe.Pointer)(toAddr) = fromAddr
	return nil
}

func getSliceToArrayPtrCaster(arrayType reflect.Type) castFunc {
	length := arrayType.Len()
	elemTypePtr := typePtr(arrayType.Elem())
	return func(fromAddr, toAddr unsafe.Pointer) error {
		from := *(*slice)(fromAddr)
		toPtr := (*unsafe.Pointer)(toAddr)
		if from.cap >= length {
			*toPtr = from.data
		} else {
			to := newObject(arrayType)
			*toPtr = to
			typedslicecopy(elemTypePtr, to, length, from.data, from.len)
		}
		return nil
	}
}

func strToBytePtrCaster(fromAddr, toAddr unsafe.Pointer) error {
	from := (*str)(fromAddr)
	*(*unsafe.Pointer)(toAddr) = from.data
	return nil
}

func getStrToArrayPtrCaster(arrayType reflect.Type) castFunc {
	length := arrayType.Len()
	elemTypePtr := typePtr(arrayType.Elem())
	return func(fromAddr, toAddr unsafe.Pointer) error {
		from := *(*str)(fromAddr)
		toPtr := (*unsafe.Pointer)(toAddr)
		if from.len >= length {
			*toPtr = from.data
		} else {
			to := newObject(arrayType)
			*toPtr = to
			typedslicecopy(elemTypePtr, to, length, from.data, from.len)
		}
		return nil
	}
}
