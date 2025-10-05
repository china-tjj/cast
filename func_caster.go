// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getFuncCaster(s *Scope, fromType, toType reflect.Type) castFunc {
	switch fromType.Kind() {
	case reflect.Func:
		numIn := fromType.NumIn()
		if numIn != toType.NumIn() {
			return nil
		}
		numOut := fromType.NumOut()
		if numOut != toType.NumOut() {
			return nil
		}
		inCasters := make([]reflectCastFunc, numIn)
		outCasters := make([]reflectCastFunc, numOut)
		for i := 0; i < numIn; i++ {
			inCasters[i] = getReflectCaster(s, toType.In(i), fromType.In(i))
			if inCasters[i] == nil {
				return nil
			}
		}
		for i := 0; i < numOut; i++ {
			outCasters[i] = getReflectCaster(s, fromType.Out(i), toType.Out(i))
			if outCasters[i] == nil {
				return nil
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
			*(*unsafe.Pointer)(toAddr) = *(*unsafe.Pointer)(getValueAddr(to))
			return nil
		}
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Pointer:
		return getAddressingPointerCaster(s, fromType, toType)
	default:
		return nil
	}
}

type reflectCastFunc func(from reflect.Value) (reflect.Value, error)

func getReflectCaster(s *Scope, fromType reflect.Type, toType reflect.Type) reflectCastFunc {
	caster := getCaster(s, fromType, toType)
	if caster == nil {
		return nil
	}
	return func(from reflect.Value) (reflect.Value, error) {
		fromAddr := getValueAddr(from)
		to := reflect.New(toType).Elem()
		if fromAddr == nil {
			return to, nil
		}
		err := caster(fromAddr, getValueAddr(to))
		return to, err
	}
}
