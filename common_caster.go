package cast

import (
	"reflect"
	"unsafe"
)

type iface interface {
	M()
}

func getUnpackInterfaceCaster(fromType, toType reflect.Type) castFunc {
	return func(fromAddr, toAddr unsafe.Pointer) bool {
		var eface any
		if fromType.NumMethod() == 0 {
			eface = *(*any)(fromAddr)
		} else {
			eface = (any)(*(*iface)(fromAddr))
		}
		fromElem := reflect.ValueOf(eface)
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

func getUnpackSliceCaster(fromType, toType reflect.Type) castFunc {
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

func getUnpackArrayCaster(fromType, toType reflect.Type) castFunc {
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

func getUnpackStructCaster(fromType, toType reflect.Type) castFunc {
	if fromType.NumField() != 1 {
		return nil
	}
	caster := getCaster(fromType.Field(0).Type, toType)
	if caster == nil {
		return nil
	}
	return caster
}
