package cast

import (
	"reflect"
	"unsafe"
)

func getUnwrapInterfaceCaster(fromType, toType reflect.Type) castFunc {
	return func(fromAddr, toAddr unsafe.Pointer) bool {
		fromElem := reflect.NewAt(fromType, fromAddr).Elem().Elem()
		elemCaster := getCaster(fromElem.Type(), toType)
		if elemCaster == nil {
			return false
		}
		return elemCaster(getValueAddr(fromElem), toAddr)
	}
}

func getAddressingPointerCaster(fromType, toType reflect.Type) castFunc {
	elemCaster := getCaster(fromType.Elem(), toType)
	if elemCaster == nil {
		return nil
	}
	return func(fromAddr, toAddr unsafe.Pointer) bool {
		fromAddr = *(*unsafe.Pointer)(fromAddr)
		if fromAddr == nil {
			return false
		}
		return elemCaster(fromAddr, toAddr)
	}
}

func getUnwrapSliceCaster(fromType, toType reflect.Type) castFunc {
	fromElemType := fromType.Elem()
	elemCaster := getCaster(fromElemType, toType)
	if elemCaster == nil {
		return nil
	}
	return func(fromAddr, toAddr unsafe.Pointer) bool {
		from := (*slice)(fromAddr)
		if from.len != 1 {
			return false
		}
		return elemCaster(from.data, toAddr)
	}
}

func getUnwrapArrayCaster(fromType, toType reflect.Type) castFunc {
	if fromType.Len() != 1 {
		return nil
	}
	fromElemType := fromType.Elem()
	elemCaster := getCaster(fromElemType, toType)
	if elemCaster == nil {
		return nil
	}
	return func(fromAddr, toAddr unsafe.Pointer) bool {
		return elemCaster(fromAddr, toAddr)
	}
}

func getUnwrapStructCaster(fromType, toType reflect.Type) castFunc {
	if fromType.NumField() != 1 {
		return nil
	}
	caster := getCaster(fromType.Field(0).Type, toType)
	if caster == nil {
		return nil
	}
	return caster
}
