// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getFinalElem(tpy reflect.Type) (int, reflect.Type) {
	var depth int
	for tpy.Kind() == reflect.Pointer {
		depth++
		tpy = tpy.Elem()
	}
	return depth, tpy
}

func getPointerCaster(fromType, toType reflect.Type) castFunc {
	switch fromType.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Array, reflect.Map, reflect.Slice, reflect.String, reflect.Struct:
		toElemType := toType.Elem()
		caster := getCaster(fromType, toElemType)
		if caster == nil {
			return nil
		}
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			toPtr := (*unsafe.Pointer)(toAddr)
			if *toPtr == nil {
				*toPtr = unsafe.Pointer(reflect.New(toElemType).Pointer())
			}
			caster(fromAddr, *toPtr)
			return true
		}
	case reflect.Uintptr:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*unsafe.Pointer)(toAddr) = unsafe.Pointer(*(*uintptr)(fromAddr))
			return true
		}
	case reflect.Interface:
		return getUnpackInterfaceCaster(fromType, toType)
	case reflect.Pointer:
		fromDepth, fromElemType := getFinalElem(fromType)
		toDepth, toElemType := getFinalElem(toType)
		if isMemSame(fromElemType, toElemType) {
			return func(fromAddr, toAddr unsafe.Pointer) bool {
				for d := fromDepth; d > toDepth; d-- {
					fromAddr = *(*unsafe.Pointer)(fromAddr)
					if fromAddr == nil {
						return true
					}
				}
				for d := toDepth; d > fromDepth; d-- {
					toPtr := (*unsafe.Pointer)(toAddr)
					if *toPtr == nil {
						*toPtr = unsafe.Pointer(new(unsafe.Pointer))
					}
					toAddr = *toPtr
				}
				*(*unsafe.Pointer)(toAddr) = *(*unsafe.Pointer)(fromAddr)
				return true
			}
		}
		caster := newCaster(fromElemType, toElemType)
		if caster == nil {
			return nil
		}
		depth := minInt(fromDepth, toDepth)
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			for d := fromDepth; d > toDepth; d-- {
				fromAddr = *(*unsafe.Pointer)(fromAddr)
				if fromAddr == nil {
					return true
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
					return true
				}
				toPtr := (*unsafe.Pointer)(toAddr)
				if *toPtr == nil {
					if d > 1 {
						*toPtr = unsafe.Pointer(new(unsafe.Pointer))
					} else {
						*toPtr = unsafe.Pointer(reflect.New(toElemType).Pointer())
					}
				}
				toAddr = *toPtr
			}
			return caster(fromAddr, toAddr)
		}
	case reflect.UnsafePointer:
		return func(fromAddr, toAddr unsafe.Pointer) bool {
			*(*unsafe.Pointer)(toAddr) = *(*unsafe.Pointer)(fromAddr)
			return true
		}
	default:
		return nil
	}
}
