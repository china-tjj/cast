// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"fmt"
	"reflect"
	"unsafe"
)

func getFuncCaster(s *Scope, fromType, toType reflect.Type) (castFunc, uint8) {
	switch fromType.Kind() {
	case reflect.Func:
		numFromIn := fromType.NumIn()
		if numFromIn != toType.NumIn() {
			return nil, 0
		}
		numFromOut := fromType.NumOut()
		numToOut := toType.NumOut()
		lastOutIsErr := false
		outAddErr := false
		// 目标方法返回值可以多一个error
		if numFromOut+1 == numToOut && toType.Out(numToOut-1) == errType {
			lastOutIsErr = true
			outAddErr = true
		} else if numFromOut == numToOut {
			lastOutIsErr = toType.Out(numToOut-1) == errType
		} else {
			return nil, 0
		}
		inCasters := make([]reflectCastFunc, numFromIn)
		outCasters := make([]reflectCastFunc, numFromOut)
		for i := 0; i < numFromIn; i++ {
			inCasters[i] = getReflectCaster(s, toType.In(i), fromType.In(i))
			if inCasters[i] == nil {
				return nil, 0
			}
		}
		for i := 0; i < numFromOut; i++ {
			outCasters[i] = getReflectCaster(s, fromType.Out(i), toType.Out(i))
			if outCasters[i] == nil {
				return nil, 0
			}
		}
		isVariadic := fromType.IsVariadic()
		if !lastOutIsErr {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				if *(*func())(fromAddr) == nil {
					*(*func())(toAddr) = nil
					return nil
				}
				from := reflect.NewAt(fromType, fromAddr).Elem()
				to := reflect.MakeFunc(toType, func(args []reflect.Value) []reflect.Value {
					for i, arg := range args {
						args[i], _ = inCasters[i](arg)
					}
					var outs []reflect.Value
					if isVariadic {
						outs = from.CallSlice(args)
					} else {
						outs = from.Call(args)
					}
					for i, out := range outs {
						outs[i], _ = outCasters[i](out)
					}
					return outs
				})
				_, fnPtr := unpackEface(to.Interface())
				*(*unsafe.Pointer)(toAddr) = fnPtr
				return nil
			}, 0
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			if *(*func())(fromAddr) == nil {
				*(*func())(toAddr) = nil
				return nil
			}
			from := reflect.NewAt(fromType, fromAddr).Elem()
			to := reflect.MakeFunc(toType, func(args []reflect.Value) []reflect.Value {
				var err error
				for i, arg := range args {
					args[i], err = inCasters[i](arg)
					if err != nil {
						err = fmt.Errorf("cast in arg %d from %s to %s failed: %w", i, toType.In(i), fromType.In(i), err)
						return returnErr(toType, err)
					}
				}
				var outs []reflect.Value
				if isVariadic {
					outs = from.CallSlice(args)
				} else {
					outs = from.Call(args)
				}
				for i, out := range outs {
					outs[i], err = outCasters[i](out)
					if err != nil {
						err = fmt.Errorf("cast out arg %d from %s to %s failed: %w", i, fromType.Out(i), toType.Out(i), err)
						return returnErr(toType, err)
					}
				}
				if outAddErr {
					outs = append(outs, nilErrValue)
				}
				return outs
			})
			_, fnPtr := unpackEface(to.Interface())
			*(*unsafe.Pointer)(toAddr) = fnPtr
			return nil
		}, 0
	case reflect.Interface:
		return getUnpackInterfaceCaster(s, fromType, toType)
	case reflect.Pointer:
		return getAddressingPointerCaster(s, fromType, toType)
	default:
		return nil, 0
	}
}

type reflectCastFunc func(from reflect.Value) (reflect.Value, error)

func getReflectCaster(s *Scope, fromType reflect.Type, toType reflect.Type) reflectCastFunc {
	caster, _ := getCaster(s, fromType, toType)
	if caster == nil {
		return nil
	}
	return func(from reflect.Value) (reflect.Value, error) {
		// 这里不需要检查from.IsValid()
		toPtr := reflect.New(toType)
		err := caster(getValueAddr(from), toPtr.UnsafePointer())
		return toPtr.Elem(), err
	}
}

func returnErr(funcType reflect.Type, err error) []reflect.Value {
	n := funcType.NumOut()
	outs := make([]reflect.Value, n)
	for i := 0; i < n-1; i++ {
		outs[i] = reflect.Zero(funcType.Out(i))
	}
	outs[n-1] = reflect.ValueOf(err)
	return outs
}
