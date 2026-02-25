// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
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
		// 空接口（eface）的第一个字是 *_type，所以只依赖具体类型；非空接口（iface）的第一个字是 *itab，而 itab 是 (interface type, concrete type) 的组合，里面还包含接口类型指针和该接口的方法表
		return fromType == toType || (fromType.NumMethod() == 0 && toType.NumMethod() == 0)
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
			fromName, skip1 := getFieldName(&fromField)
			if skip1 {
				return false
			}
			toName, skip2 := getFieldName(&toField)
			if skip2 {
				return false
			}
			if !(fromName == toName || foldNameStr(fromName) == foldNameStr(toName)) {
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

func getFieldName(field *reflect.StructField) (string, bool) {
	if castTag := field.Tag.Get("cast"); castTag == "-" {
		return "", true
	} else if castTag != "" {
		return strings.Split(castTag, ",")[0], false
	} else if jsonTag := field.Tag.Get("json"); jsonTag == "-" {
		return "", false
	} else if jsonTag != "" {
		return strings.Split(jsonTag, ",")[0], false
	}
	return field.Name, false
}

func foldNameStr(in string) string {
	return toString(foldName(toBytes(in)))
}

func foldName(in []byte) []byte {
	var arr [32]byte
	return appendFoldedName(arr[:0], in)
}

func appendFoldedName(out, in []byte) []byte {
	for i := 0; i < len(in); {
		if c := in[i]; c < utf8.RuneSelf {
			if c != '_' && c != '-' {
				if 'a' <= c && c <= 'z' {
					c -= 'a' - 'A'
				}
				out = append(out, c)
			}
			i++
			continue
		}
		r, n := utf8.DecodeRune(in[i:])
		out = utf8.AppendRune(out, foldRune(r))
		i += n
	}
	return out
}

func foldRune(r rune) rune {
	for {
		r2 := unicode.SimpleFold(r)
		if r2 <= r {
			return r2
		}
		r = r2
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
	switch typ.Kind() {
	// chan、map、func 其实就是一个指针
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer:
		return true
	default:
		return false
	}
}

func isNilableType(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
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

//go:linkname typedMemMove runtime.typedmemmove
func typedMemMove(typ, dst, src unsafe.Pointer)

//go:linkname typedSliceCopy runtime.typedslicecopy
func typedSliceCopy(typ, dstPtr unsafe.Pointer, dstLen int, srcPtr unsafe.Pointer, srcLen int) int

//go:linkname mallocGC runtime.mallocgc
func mallocGC(size uintptr, typ unsafe.Pointer, needzero bool) unsafe.Pointer

func newObject(typ reflect.Type) unsafe.Pointer {
	return mallocGC(typ.Size(), typePtr(typ), true)
}

func copyObject(typ reflect.Type, ptr unsafe.Pointer) unsafe.Pointer {
	tp := typePtr(typ)
	copiedPtr := mallocGC(typ.Size(), tp, true)
	typedMemMove(tp, copiedPtr, ptr)
	return copiedPtr
}

//go:linkname newArray runtime.newarray
func newArray(typ unsafe.Pointer, n int) unsafe.Pointer

type slice struct {
	data unsafe.Pointer
	len  int
	cap  int
}

func makeSlice(elemType reflect.Type, len, cap int) slice {
	return slice{
		data: newArray(typePtr(elemType), cap),
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

func getFinalElem(typ reflect.Type) (int, reflect.Type) {
	var depth int
	for typ.Kind() == reflect.Pointer {
		depth++
		typ = typ.Elem()
	}
	return depth, typ
}

var zeroMu sync.RWMutex
var zeroMap = make(map[unsafe.Pointer]unsafe.Pointer)
var zeros [zerosSize]byte

func getZeroPtr(typ reflect.Type) unsafe.Pointer {
	if typ.Size() < zerosSize {
		return unsafe.Pointer(&zeros[0])
	}

	typPtr := typePtr(typ)
	zeroMu.RLock()
	if ptr, ok := zeroMap[typPtr]; ok {
		zeroMu.RUnlock()
		return ptr
	}
	zeroMu.RUnlock()

	zeroMu.Lock()
	defer zeroMu.Unlock()
	var zeroPtr unsafe.Pointer
	if ptr, ok := zeroMap[typPtr]; ok {
		zeroPtr = ptr
	} else {
		zeroPtr = newObject(typ)
		zeroMap[typPtr] = zeroPtr
	}
	return zeroPtr
}

type structField struct {
	rawName    string // 原始字段名，仅打error用
	name       string // 一定非空
	foldedName string // 可能为空，为空说明是名称重复
	offset     uintptr
	typ        reflect.Type
	isRequired bool
	// 嵌套结构体指针相关字段
	parent        *structField
	parentElemTyp reflect.Type
}

// addr 为根结构体的地址，要求不为 nil。递归放到 getAddrSlow，使得该函数可以内联
func (f *structField) getAddr(addr unsafe.Pointer, newIfNil bool) unsafe.Pointer {
	if f.parent != nil {
		return f.getAddrSlow(addr, newIfNil)
	}
	return unsafe.Add(addr, f.offset)
}

func (f *structField) getAddrSlow(addr unsafe.Pointer, newIfNil bool) unsafe.Pointer {
	if f.parent != nil {
		parentAddr := f.parent.getAddrSlow(addr, newIfNil)
		addr = *(*unsafe.Pointer)(parentAddr)
		if addr == nil {
			if !newIfNil {
				return nil
			}
			addr = newObject(f.parentElemTyp)
			*(*unsafe.Pointer)(parentAddr) = addr
		}
	}
	return unsafe.Add(addr, f.offset)
}

type structFields struct {
	flattened    []*structField
	byActualName map[string]*structField
	byFoldedName map[string]*structField
}

var fieldCache sync.Map

type fieldCacheKey struct {
	castUnexported bool
	typ            reflect.Type
}

// 获取结构体的所有字段，包括提升的
func getAllFields(s *Scope, typ reflect.Type) structFields {
	key := fieldCacheKey{
		castUnexported: s.castUnexported,
		typ:            typ,
	}
	if f, ok := fieldCache.Load(key); ok {
		return f.(structFields)
	}
	f, _ := fieldCache.LoadOrStore(key, getAllFieldsInner(s, typ, 0, nil, make(map[reflect.Type]struct{})))
	return f.(structFields)
}

func getAllFieldsInner(s *Scope, typ reflect.Type, offset uintptr, parent *structField, visited map[reflect.Type]struct{}) structFields {
	if _, v := visited[typ]; v {
		return structFields{}
	}
	visited[typ] = struct{}{}

	n := typ.NumField()
	fields := make([]*structField, 0, n)
	var anonymousFields []*structField
	nameMap := make(map[string]*structField, n)
	foldedNameMap := make(map[string]*structField, n)
	foldedNameCandidateMap := make(map[string][]*structField, n)
	anonymousNameMap := make(map[string][]*structField)
	anonymousFoldedNameMap := make(map[string][]*structField)

	for i := 0; i < n; i++ {
		reflectField := typ.Field(i)
		if !s.castUnexported && !reflectField.IsExported() {
			continue
		}
		field := &structField{
			rawName: reflectField.Name,
			offset:  offset + reflectField.Offset,
			typ:     reflectField.Type,
		}
		if parent != nil {
			field.parent = parent
			field.parentElemTyp = parent.typ.Elem()
		}
		if castTag := typ.Field(i).Tag.Get("cast"); castTag == "-" {
			continue
		} else if castTag != "" {
			values := strings.Split(castTag, ",")
			field.name = values[0]
			for _, value := range values[1:] {
				switch value {
				case "required":
					field.isRequired = true
				default:
					break
				}
			}
		} else if jsonTag := typ.Field(i).Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			field.name = strings.Split(jsonTag, ",")[0]
		}
		if field.name == "" && reflectField.Anonymous {
			if field.typ.Kind() == reflect.Struct {
				subFields := getAllFieldsInner(s, field.typ, field.offset, nil, visited)
				for _, anonymousField := range subFields.flattened {
					anonymousFields = append(anonymousFields, anonymousField)
					name, foldedName := anonymousField.name, anonymousField.foldedName
					anonymousNameMap[name] = append(anonymousNameMap[name], anonymousField)
					if anonymousField.foldedName == "" {
						continue
					}
					anonymousFoldedNameMap[foldedName] = append(anonymousFoldedNameMap[foldedName], anonymousField)
				}
				continue
			} else if field.typ.Kind() == reflect.Ptr && field.typ.Elem().Kind() == reflect.Struct {
				subFields := getAllFieldsInner(s, field.typ.Elem(), 0, field, visited)
				for _, anonymousField := range subFields.flattened {
					anonymousFields = append(anonymousFields, anonymousField)
					name, foldedName := anonymousField.name, anonymousField.foldedName
					anonymousNameMap[name] = append(anonymousNameMap[name], anonymousField)
					if anonymousField.foldedName == "" {
						continue
					}
					anonymousFoldedNameMap[foldedName] = append(anonymousFoldedNameMap[foldedName], anonymousField)
				}
				continue
			}
		}
		if field.name == "" {
			field.name = reflectField.Name
		}
		field.foldedName = foldNameStr(field.name)
		fields = append(fields, field)
		nameMap[field.name] = field
		foldedNameCandidateMap[field.foldedName] = append(foldedNameCandidateMap[field.foldedName], field)
	}

	for foldedName, candidates := range foldedNameCandidateMap {
		if len(candidates) == 1 {
			foldedNameMap[foldedName] = candidates[0]
		}
	}

	for _, anonymousField := range anonymousFields {
		name := anonymousField.name
		if _, ok := nameMap[name]; ok || len(anonymousNameMap[name]) != 1 {
			continue
		}
		nameMap[name] = anonymousField
		fields = append(fields, anonymousField)
		foldedName := anonymousField.foldedName
		if _, ok := foldedNameMap[foldedName]; ok && len(anonymousFoldedNameMap[foldedName]) != 1 {
			anonymousField.foldedName = ""
			continue
		}
		foldedNameMap[foldedName] = anonymousField
	}
	return structFields{fields, nameMap, foldedNameMap}
}
