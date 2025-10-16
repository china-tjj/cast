// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getFuncCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	switch fromType.Kind() {
	case reflect.Func:
		numIn := fromType.NumIn()
		if numIn != toType.NumIn() {
			return nil, false
		}
		numOut := fromType.NumOut()
		if numOut != toType.NumOut() {
			return nil, false
		}
		inCasters := make([]reflectCastFunc, numIn)
		outCasters := make([]reflectCastFunc, numOut)
		for i := 0; i < numIn; i++ {
			inCasters[i] = getReflectCaster(s, toType.In(i), fromType.In(i))
			if inCasters[i] == nil {
				return nil, false
			}
		}
		for i := 0; i < numOut; i++ {
			outCasters[i] = getReflectCaster(s, fromType.Out(i), toType.Out(i))
			if outCasters[i] == nil {
				return nil, false
			}
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			from := reflect.NewAt(fromType, fromAddr).Elem()
			to := reflect.MakeFunc(toType, func(args []reflect.Value) []reflect.Value {
				for i, arg := range args {
					args[i], _ = inCasters[i](arg)
				}
				outs := from.Call(args)
				for i, out := range outs {
					outs[i], _ = outCasters[i](out)
				}
				return outs
			})
			_, fnPtr, _ := unpackEface(to.Interface())
			*(*unsafe.Pointer)(toAddr) = fnPtr
			return nil
		}, false
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Pointer:
		return getAddressingPointerCaster(s, fromType, toType)
	default:
		return nil, false
	}
}

type reflectCastFunc func(from reflect.Value) (reflect.Value, error)

func getReflectCaster(s *Scope, fromType reflect.Type, toType reflect.Type) reflectCastFunc {
	caster, _ := getCaster(s, fromType, toType)
	if caster == nil {
		return nil
	}
	return func(from reflect.Value) (reflect.Value, error) {
		if !from.IsValid() {
			return reflect.Zero(toType), nil
		}
		toPtr := reflect.New(toType)
		err := caster(getValueAddr(from), toPtr.UnsafePointer())
		return toPtr.Elem(), err
	}
}
