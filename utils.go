// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

func isMemSame(s *Scope, fromType, toType reflect.Type) bool {
	if fromType == toType {
		return true
	}
	fromKind := fromType.Kind()
	toKind := toType.Kind()
	if fromKind != toKind {
		return fromKind == reflect.Pointer && toKind == reflect.UnsafePointer ||
			fromKind == reflect.UnsafePointer && toKind == reflect.Pointer
	}
	switch fromKind {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.String, reflect.UnsafePointer:
		return true
	case reflect.Array:
		return fromType.Len() == toType.Len() && isMemSame(s, fromType.Elem(), toType.Elem())
	case reflect.Chan:
		fromDir := fromType.ChanDir()
		return (fromDir == reflect.BothDir || fromDir == toType.ChanDir()) && isMemSame(s, fromType.Elem(), toType.Elem())
	case reflect.Func:
		numIn, numOut := fromType.NumIn(), fromType.NumOut()
		if numIn != toType.NumIn() || numOut != toType.NumOut() {
			return false
		}
		for i := 0; i < numIn; i++ {
			if !isMemSame(s, fromType.In(i), toType.In(i)) {
				return false
			}
		}
		for i := 0; i < numOut; i++ {
			if !isMemSame(s, fromType.Out(i), toType.Out(i)) {
				return false
			}
		}
		return true
	case reflect.Interface:
		return fromType.Implements(toType) && toType.Implements(fromType)
	case reflect.Map:
		return isMemSame(s, fromType.Key(), toType.Key()) && isMemSame(s, fromType.Elem(), toType.Elem())
	case reflect.Pointer, reflect.Slice:
		return isMemSame(s, fromType.Elem(), toType.Elem())
	case reflect.Struct:
		n := fromType.NumField()
		if n != toType.NumField() {
			return false
		}
		for i := 0; i < n; i++ {
			fromField := fromType.Field(i)
			toField := toType.Field(i)
			if !s.castUnexported && (!fromField.IsExported() || !toField.IsExported()) {
				return false
			}
			fromJsonTag, ok1 := fromField.Tag.Lookup("json")
			toJsonTag, ok2 := toField.Tag.Lookup("json")
			if fromField.Name != toField.Name || (ok1 && ok2 && fromJsonTag != toJsonTag) {
				return false
			}
			if !isMemSame(s, fromField.Type, toField.Type) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func isRefAble(s *Scope, fromType, toType reflect.Type) bool {
	if s.deepCopy {
		return false
	}
	if fromType == toType {
		return true
	}
	if s.disableZeroCopy {
		return false
	}
	return isMemSame(s, fromType, toType)
}

func typeFor[T any]() reflect.Type {
	var v T
	if t := reflect.TypeOf(v); t != nil {
		return t // optimize for T being a non-interface kind
	}
	return reflect.TypeOf((*T)(nil)).Elem() // only for an interface kind
}

func typePtr(t reflect.Type) unsafe.Pointer {
	return noEscape((*eface)(unsafe.Pointer(&t)).ptr)
}

//go:linkname typePtrToType reflect.toType
func typePtrToType(t unsafe.Pointer) reflect.Type

func min[T iNumber](a, b T) T {
	if a < b {
		return a
	}
	return b
}

//go:nosplit
func noEscape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

//go:nosplit
func noEscapePtr[T any](p *T) *T {
	x := uintptr(unsafe.Pointer(p))
	return (*T)(unsafe.Pointer(x ^ 0))
}

var alwaysFalse bool
var escapeSink any

func escape[T any](x T) T {
	if alwaysFalse {
		escapeSink = x
	}
	return x
}

type iface interface {
	M()
}

type eface struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

func packEface(typ reflect.Type, ptr unsafe.Pointer) any {
	return *(*any)(unsafe.Pointer(&eface{
		typ: typePtr(typ),
		ptr: ptr,
	}))
}

func loadEface(numMethod int, ptr unsafe.Pointer) any {
	if numMethod == 0 {
		return *(*any)(ptr)
	} else {
		return *(*iface)(ptr)
	}
}

func unpackEface(v any) (unpackedTypePtr unsafe.Pointer, unpackedDataPtr unsafe.Pointer) {
	e := *(*eface)(unsafe.Pointer(&v))
	return e.typ, e.ptr
}

func isRefType(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Array:
		return isRefType(typ.Elem())
	case reflect.Chan, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return true
	case reflect.Struct:
		n := typ.NumField()
		for i := 0; i < n; i++ {
			if isRefType(typ.Field(i).Type) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func isPtrType(typ reflect.Type) bool {
	if typ == nil {
		return false
	}
	switch typ.Kind() {
	// chan、map、func 其实就是一个指针
	case reflect.Chan, reflect.Map, reflect.Func, reflect.Pointer, reflect.UnsafePointer:
		return true
	default:
		return false
	}
}

func getValueAddr(v reflect.Value) unsafe.Pointer {
	if v.CanAddr() {
		return v.Addr().UnsafePointer()
	}
	copiedPtr := reflect.New(v.Type())
	copiedPtr.Elem().Set(v)
	return copiedPtr.UnsafePointer()
}

func offset(data unsafe.Pointer, idx int, elemSize uintptr) unsafe.Pointer {
	return unsafe.Add(data, uintptr(idx)*elemSize)
}

//go:linkname typedmemmove runtime.typedmemmove
func typedmemmove(typ, dst, src unsafe.Pointer)

//go:linkname typedslicecopy runtime.typedslicecopy
func typedslicecopy(typ, dstPtr unsafe.Pointer, dstLen int, srcPtr unsafe.Pointer, srcLen int) int

//go:linkname mallocgc runtime.mallocgc
func mallocgc(size uintptr, typ unsafe.Pointer, needzero bool) unsafe.Pointer

func newObject(typ reflect.Type) unsafe.Pointer {
	return mallocgc(typ.Size(), typePtr(typ), true)
}

func copyObject(typ reflect.Type, ptr unsafe.Pointer) unsafe.Pointer {
	tp := typePtr(typ)
	copiedPtr := mallocgc(typ.Size(), tp, false)
	typedmemmove(tp, copiedPtr, ptr)
	return copiedPtr
}

//go:linkname newarray runtime.newarray
func newarray(typ unsafe.Pointer, n int) unsafe.Pointer

type slice struct {
	data unsafe.Pointer
	len  int
	cap  int
}

func makeSlice(elemType reflect.Type, len, cap int) slice {
	return slice{
		data: newarray(typePtr(elemType), cap),
		len:  len,
		cap:  cap,
	}
}

type str struct {
	data unsafe.Pointer
	len  int
}

func toBytes(s string) (b []byte) {
	from := (*str)(unsafe.Pointer(&s))
	to := (*slice)(unsafe.Pointer(&b))
	to.data = from.data
	to.len = from.len
	to.cap = from.len
	return
}

func toString(b []byte) (s string) {
	from := (*slice)(unsafe.Pointer(&b))
	to := (*str)(unsafe.Pointer(&s))
	to.data = from.data
	to.len = from.len
	return
}

func isSpecialHash(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Array:
		return typ.Len() > 0 && isSpecialHash(typ.Elem())
	case reflect.String, reflect.Interface:
		return true
	case reflect.Struct:
		n := typ.NumField()
		for i := 0; i < n; i++ {
			if isSpecialHash(typ.Field(i).Type) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func getDiffJsonTag(f *reflect.StructField) (string, bool) {
	if jsonTag, ok := f.Tag.Lookup("json"); ok && jsonTag != f.Name {
		return jsonTag, true
	}
	return "", false
}
