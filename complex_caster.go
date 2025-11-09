package cast

import (
	"reflect"
	"strconv"
	"unsafe"
)

type iComplex interface {
	~complex64 | ~complex128
}

func getComplexCaster[T iComplex](s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	switch fromType.Kind() {
	case reflect.Complex64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*T)(toAddr) = T(*(*complex64)(fromAddr))
			return nil
		}, false
	case reflect.Complex128:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*T)(toAddr) = T(*(*complex128)(fromAddr))
			return nil
		}, false
	case reflect.String:
		toBitSize := int(8 * toType.Size())
		return func(fromAddr, toAddr unsafe.Pointer) error {
			c128, err := strconv.ParseComplex(*(*string)(fromAddr), toBitSize)
			if err != nil {
				return err
			}
			*(*T)(toAddr) = T(c128)
			return nil
		}, false
	default:
		return nil, false
	}
}
