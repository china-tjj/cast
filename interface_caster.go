// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func getFallbackInterfaceCaster(fromType, toType reflect.Type) castFunc {
	return func(fromAddr, toAddr unsafe.Pointer) error {
		reflect.NewAt(toType, toAddr).Elem().Set(reflect.NewAt(fromType, fromAddr).Elem())
		return nil
	}
}

func getInterfaceCaster(s *Scope, fromType, toType reflect.Type) castFunc {
	if toType.NumMethod() == 0 {
		switch fromType.Kind() {
		case reflect.Bool:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*bool)(fromAddr)
				return nil
			}
		case reflect.Int:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*int)(fromAddr)
				return nil
			}
		case reflect.Int8:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*int8)(fromAddr)
				return nil
			}
		case reflect.Int16:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*int16)(fromAddr)
				return nil
			}
		case reflect.Int32:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*int32)(fromAddr)
				return nil
			}
		case reflect.Int64:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*int64)(fromAddr)
				return nil
			}
		case reflect.Uint:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*uint)(fromAddr)
				return nil
			}
		case reflect.Uint8:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*uint8)(fromAddr)
				return nil
			}
		case reflect.Uint16:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*uint16)(fromAddr)
				return nil
			}
		case reflect.Uint32:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*uint32)(fromAddr)
				return nil
			}
		case reflect.Uint64:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*uint64)(fromAddr)
				return nil
			}
		case reflect.Uintptr:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*uintptr)(fromAddr)
				return nil
			}
		case reflect.Float32:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*float32)(fromAddr)
				return nil
			}
		case reflect.Float64:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*float64)(fromAddr)
				return nil
			}
		case reflect.Interface:
			if fromType.NumMethod() == 0 {
				return func(fromAddr, toAddr unsafe.Pointer) error {
					*(*any)(toAddr) = *(*any)(fromAddr)
					return nil
				}
			}
			return getFallbackInterfaceCaster(fromType, toType)
		case reflect.String:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*string)(fromAddr)
				return nil
			}
		case reflect.UnsafePointer:
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*any)(toAddr) = *(*unsafe.Pointer)(fromAddr)
				return nil
			}
		default:
			return getFallbackInterfaceCaster(fromType, toType)
		}
	}

	if !fromType.Implements(toType) {
		return nil
	}
	// 转为非空接口时，虽然可以手动构建一个 Tab 来避免反射，但是这太 trick 和 unsafe 了，安全起见，还是用反射吧
	switch fromType.Kind() {
	case reflect.Bool:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.ValueOf(*(*bool)(fromAddr)))
			return nil
		}
	case reflect.Int:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.ValueOf(*(*int)(fromAddr)))
			return nil
		}
	case reflect.Int8:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.ValueOf(*(*int8)(fromAddr)))
			return nil
		}
	case reflect.Int16:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.ValueOf(*(*int16)(fromAddr)))
			return nil
		}
	case reflect.Int32:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.ValueOf(*(*int32)(fromAddr)))
			return nil
		}
	case reflect.Int64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.ValueOf(*(*int64)(fromAddr)))
			return nil
		}
	case reflect.Uint:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.ValueOf(*(*uint)(fromAddr)))
			return nil
		}
	case reflect.Uint8:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.ValueOf(*(*uint8)(fromAddr)))
			return nil
		}
	case reflect.Uint16:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.ValueOf(*(*uint16)(fromAddr)))
			return nil
		}
	case reflect.Uint32:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.ValueOf(*(*uint32)(fromAddr)))
			return nil
		}
	case reflect.Uint64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.ValueOf(*(*uint64)(fromAddr)))
			return nil
		}
	case reflect.Uintptr:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.ValueOf(*(*uintptr)(fromAddr)))
			return nil
		}
	case reflect.Float32:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.ValueOf(*(*float32)(fromAddr)))
			return nil
		}
	case reflect.Float64:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.ValueOf(*(*float64)(fromAddr)))
			return nil
		}
	case reflect.Interface:
		if fromType.NumMethod() == 0 {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				reflect.NewAt(toType, toAddr).Elem().Set(reflect.ValueOf(*(*any)(fromAddr)))
				return nil
			}
		} else if toType.Implements(fromType) {
			return func(fromAddr, toAddr unsafe.Pointer) error {
				*(*iface)(toAddr) = *(*iface)(fromAddr)
				return nil
			}
		}
		return getFallbackInterfaceCaster(fromType, toType)
	case reflect.String:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.ValueOf(*(*string)(fromAddr)))
			return nil
		}
	case reflect.UnsafePointer:
		return func(fromAddr, toAddr unsafe.Pointer) error {
			reflect.NewAt(toType, toAddr).Elem().Set(reflect.ValueOf(*(*unsafe.Pointer)(fromAddr)))
			return nil
		}
	default:
		return getFallbackInterfaceCaster(fromType, toType)
	}
}
