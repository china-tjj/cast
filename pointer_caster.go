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

func getPointerCaster(s *Scope, fromType, toType reflect.Type) castFunc {
	switch fromType.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Map, reflect.String, reflect.Struct:
		return getNormalPtrCaster(s, fromType, toType.Elem())
	case reflect.Array:
		if isMemSame(fromType.Elem(), toType.Elem()) {
			return arrayToElemPtrCaster
		} else {
			return getNormalPtrCaster(s, fromType, toType.Elem())
		}
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Pointer:
		fromDepth, fromElemType := getFinalElem(fromType)
		toDepth, toElemType := getFinalElem(toType)
		if !s.disableZeroCopy && isMemSame(fromElemType, toElemType) {
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
				*(*unsafe.Pointer)(toAddr) = *(*unsafe.Pointer)(fromAddr)
				return nil
			}
		}
		var caster castFunc
		switch fromElemType.Kind() {
		case reflect.Slice:
			if !s.disableZeroCopy && toElemType.Kind() == reflect.Array && isMemSame(fromElemType.Elem(), toElemType.Elem()) {
				toDepth--
				caster = getSliceToArrayPtrCaster(toElemType)
			} else if isMemSame(fromElemType.Elem(), toElemType) {
				toDepth--
				caster = sliceToElemPtrCaster
			} else {
				caster = getCaster(s, fromElemType, toElemType)
			}
		case reflect.Array:
			if isMemSame(fromElemType.Elem(), toElemType) {
				toDepth--
				caster = arrayToElemPtrCaster
			}
		default:
			caster = getCaster(s, fromElemType, toElemType)
		}
		if caster == nil {
			return nil
		}
		depth := min(fromDepth, toDepth)
		toElemSize := toElemType.Size()
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
						*toPtr = malloc(toElemSize)
					}
				}
				toAddr = *toPtr
			}
			return caster(fromAddr, toAddr)
		}
	case reflect.Slice:
		toElemType := toType.Elem()
		if !s.disableZeroCopy && toElemType.Kind() == reflect.Array && isMemSame(fromType.Elem(), toElemType.Elem()) {
			return getSliceToArrayPtrCaster(toElemType)
		} else if isMemSame(fromType.Elem(), toElemType) {
			return sliceToElemPtrCaster
		} else {
			return getNormalPtrCaster(s, fromType, toElemType)
		}
	case reflect.UnsafePointer:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*unsafe.Pointer)(toAddr) = *(*unsafe.Pointer)(fromAddr)
			return nil
		}
	default:
		return nil
	}
}

func getNormalPtrCaster(s *Scope, fromType, toElemType reflect.Type) castFunc {
	caster := getCaster(s, fromType, toElemType)
	if caster == nil {
		return nil
	}
	toElemSize := toElemType.Size()
	return func(fromAddr, toAddr unsafe.Pointer) error {
		toPtr := (*unsafe.Pointer)(toAddr)
		if *toPtr == nil {
			*toPtr = malloc(toElemSize)
		}
		return caster(fromAddr, *toPtr)
	}
}

func getSliceToArrayPtrCaster(arrayType reflect.Type) castFunc {
	length := arrayType.Len()
	arraySize := arrayType.Size()
	elemSize := arrayType.Elem().Size()
	return func(fromAddr, toAddr unsafe.Pointer) error {
		from := *(*slice)(fromAddr)
		toPtr := (*unsafe.Pointer)(toAddr)
		if from.cap >= length {
			*toPtr = from.data
		} else {
			*toPtr = malloc(arraySize)
			memCopy(*toPtr, from.data, uintptr(min(length, from.len))*elemSize)
		}
		return nil
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
