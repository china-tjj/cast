// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getFinalElem(typ reflect.Type) (int, reflect.Type) {
	var depth int
	for typ.Kind() == reflect.Pointer {
		depth++
		typ = typ.Elem()
	}
	return depth, typ
}

func getPointerCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	switch fromType.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Map, reflect.String, reflect.Struct:
		return getNormalPtrCaster(s, fromType, toType.Elem())
	case reflect.Array:
		if isRefAble(s, fromType.Elem(), toType.Elem()) {
			return arrayToElemPtrCaster, true
		} else {
			return getNormalPtrCaster(s, fromType, toType.Elem())
		}
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Pointer:
		fromDepth, fromElemType := getFinalElem(fromType)
		toDepth, toElemType := getFinalElem(toType)
		if isRefAble(s, fromElemType, toElemType) {
			if fromDepth < toDepth {
				return func(fromAddr, toAddr unsafe.Pointer) error {
					for d := toDepth - 1; d > fromDepth; d-- {
						toPtr := (*unsafe.Pointer)(toAddr)
						if *toPtr == nil {
							*toPtr = unsafe.Pointer(new(unsafe.Pointer))
						}
						toAddr = *toPtr
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
							return nil
						}
					}
					*(*unsafe.Pointer)(toAddr) = fromAddr
					return nil
				}, false
			}
		}
		var caster castFunc
		switch fromElemType.Kind() {
		case reflect.Slice:
			if toElemType.Kind() == reflect.Array && isRefAble(s, fromElemType.Elem(), toElemType.Elem()) {
				toDepth--
				caster, _ = getSliceToArrayPtrCaster(toElemType)
			} else if isRefAble(s, fromElemType.Elem(), toElemType) {
				toDepth--
				caster = sliceToElemPtrCaster
			} else {
				caster, _ = getCaster(s, fromElemType, toElemType)
			}
		case reflect.Array:
			if isRefAble(s, fromElemType.Elem(), toElemType) {
				toDepth--
				caster = arrayToElemPtrCaster
			}
		default:
			caster, _ = getCaster(s, fromElemType, toElemType)
		}
		if caster == nil {
			return nil, false
		}
		depth := min(fromDepth, toDepth)
		return func(fromAddr, toAddr unsafe.Pointer) error {
			for d := fromDepth; d > toDepth; d-- {
				fromAddr = *(*unsafe.Pointer)(fromAddr)
				if fromAddr == nil {
					return nil
				}
			}
			for d := toDepth; d > fromDepth; d-- {
				toPtr := (*unsafe.Pointer)(toAddr)
				if *toPtr == nil {
					*toPtr = unsafe.Pointer(new(unsafe.Pointer))
				}
				toAddr = *toPtr
			}
			for d := depth; d > 0; d-- {
				fromAddr = *(*unsafe.Pointer)(fromAddr)
				if fromAddr == nil {
					return nil
				}
				toPtr := (*unsafe.Pointer)(toAddr)
				if *toPtr == nil {
					if d > 1 {
						*toPtr = unsafe.Pointer(new(unsafe.Pointer))
					} else {
						*toPtr = newObject(toElemType)
					}
				}
				toAddr = *toPtr
			}
			return caster(fromAddr, toAddr)
		}, false
	case reflect.Slice:
		toElemType := toType.Elem()
		if toElemType.Kind() == reflect.Array && isRefAble(s, fromType.Elem(), toElemType.Elem()) {
			return getSliceToArrayPtrCaster(toElemType)
		} else if isRefAble(s, fromType.Elem(), toElemType) {
			return sliceToElemPtrCaster, false
		} else {
			return getNormalPtrCaster(s, fromType, toElemType)
		}
	case reflect.UnsafePointer:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*unsafe.Pointer)(toAddr) = *(*unsafe.Pointer)(fromAddr)
			return nil
		}, false
	default:
		return nil, false
	}
}

func getNormalPtrCaster(s *Scope, fromType, toElemType reflect.Type) (castFunc, bool) {
	if isRefAble(s, fromType, toElemType) {
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*unsafe.Pointer)(toAddr) = fromAddr
			return nil
		}, true
	}
	caster, hasRef := getCaster(s, fromType, toElemType)
	if caster == nil {
		return nil, false
	}
	return func(fromAddr, toAddr unsafe.Pointer) error {
		toPtr := (*unsafe.Pointer)(toAddr)
		if *toPtr == nil {
			*toPtr = newObject(toElemType)
		}
		return caster(fromAddr, *toPtr)
	}, hasRef
}

func getSliceToArrayPtrCaster(arrayType reflect.Type) (castFunc, bool) {
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
	}, false
}

func sliceToElemPtrCaster(fromAddr, toAddr unsafe.Pointer) error {
	*(*unsafe.Pointer)(toAddr) = (*slice)(fromAddr).data
	return nil
}

func arrayToElemPtrCaster(fromAddr, toAddr unsafe.Pointer) error {
	*(*unsafe.Pointer)(toAddr) = fromAddr
	return nil
}
