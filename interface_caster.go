// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getInterfaceCaster(s *Scope, fromType, toType reflect.Type) (castFunc, bool) {
	fromKind := fromType.Kind()
	if toType.NumMethod() == 0 {
		if fromKind == reflect.Interface {
			if s.deepCopy {
				return getUnpackInterfaceCaster(s, fromType, toType)
			}
			if fromType.NumMethod() == 0 {
				return func(fromAddr, toAddr unsafe.Pointer) error {
					*(*any)(toAddr) = *(*any)(fromAddr)
					return nil
				}, false
			} else {
				return func(fromAddr, toAddr unsafe.Pointer) error {
					*(*any)(toAddr) = *(*iface)(fromAddr)
					return nil
				}, false
			}
		}
		if s.deepCopy {
			copier, _ := getCaster(s, fromType, fromType)
			if copier == nil {
				return nil, false
			}
			if isPtrType(fromType) {
				return func(fromAddr, toAddr unsafe.Pointer) error {
					copied := newObject(fromType)
					if err := copier(fromAddr, copied); err != nil {
						return err
					}
					*(*any)(toAddr) = packEface(fromType, *(*unsafe.Pointer)(copied))
					return nil
				}, false
			}
			return func(fromAddr, toAddr unsafe.Pointer) error {
				copied := newObject(fromType)
				if err := copier(fromAddr, copied); err != nil {
					return err
				}
				*(*any)(toAddr) = packEface(fromType, copied)
				return nil
			}, false
		}
		if isPtrType(fromType) {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = packEface(fromType, *(*unsafe.Pointer)(fromAddr))
				return nil
			}, false
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			*(*any)(toAddr) = packEface(fromType, fromAddr)
			return nil
		}, true
	}

	if fromKind == reflect.Interface {
		if !fromType.Implements(toType) || s.deepCopy {
			return getUnpackInterfaceCaster(s, fromType, toType)
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.NewAt(fromType, fromAddr).Elem())
			return nil
		}, false
	}

	if !fromType.Implements(toType) {
		return nil, false
	}

	if s.deepCopy {
		copier, _ := getCaster(s, fromType, fromType)
		if copier == nil {
			return nil, false
		}
		return func(fromAddr, toAddr unsafe.Pointer) error {
			copied := newObject(fromType)
			if err := copier(fromAddr, copied); err != nil {
				return err
			}
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.NewAt(fromType, copied).Elem())
			return nil
		}, false
	}
	return func(fromAddr, toAddr unsafe.Pointer) error {
		reflect.NewAt(toType, toAddr).Elem().Set(reflect.NewAt(fromType, fromAddr).Elem())
		return nil
	}, false
}
