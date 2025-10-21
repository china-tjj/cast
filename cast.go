// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"sync"
	"unsafe"
)

type Scope struct {
	casterMap map[casterKey]casterValue
	mu        sync.RWMutex // 读多写少的场景，sync.RWMutex的效率比sync.Map更高
	frozen    bool

	disableZeroCopy bool // 禁用零拷贝
	deepCopy        bool // 深拷贝
	castUnexported  bool // 转换未导出字段
}

func (s *Scope) DisableZeroCopy() bool {
	return s.disableZeroCopy
}

type ScopeOption func(s *Scope)

// NewScope 创建新的作用域
func NewScope(options ...ScopeOption) *Scope {
	scope := &Scope{
		casterMap: make(map[casterKey]casterValue),
	}
	for _, option := range defaultOptions {
		option(scope)
	}
	for _, option := range options {
		option(scope)
	}
	scope.frozen = true
	return scope
}

var defaultScope = NewScope()

// SetDefaultScope ！！慎用！！设置默认作用域，可以改变默认行为
func SetDefaultScope(s *Scope) {
	defaultScope = s
}

// Cast 将类型F转为T，转换规则详见README
func Cast[F any, T any](from F) (to T, err error) {
	return CastWithScope[F, T](defaultScope, from)
}

// GetCaster 获取实例化的转换方法，调用该方法返回的函数，对比直接调用 Cast 少了查缓存的步骤，性能会略微好一点
func GetCaster[F any, T any]() func(from F) (to T, err error) {
	return GetCasterWithScope[F, T](defaultScope)
}

// ReflectCast 以反射的方式，需输入待转换的值与要转换的类型
func ReflectCast(from reflect.Value, toType reflect.Type) (to reflect.Value, err error) {
	return ReflectCastWithScope(defaultScope, from, toType)
}

// CastWithScope 将类型F转为T，转换规则详见README
func CastWithScope[F any, T any](s *Scope, from F) (to T, err error) {
	fromType, toType := typeFor[F](), typeFor[T]()
	caster, _ := getCaster(s, fromType, toType)
	if caster == nil {
		return to, invalidCastErr(s, fromType, toType)
	}
	// 当from包含指针时，这部分指针指向的内容会逃逸
	escape(from)
	err = caster(noEscape(unsafe.Pointer(&from)), noEscape(unsafe.Pointer(&to)))
	return to, err
}

// GetCasterWithScope 获取实例化的转换方法，调用该方法返回的函数，对比直接调用 Cast 少了查缓存的步骤，性能会略微好一点
func GetCasterWithScope[F any, T any](s *Scope) func(from F) (to T, err error) {
	fromType, toType := typeFor[F](), typeFor[T]()
	caster, _ := getCaster(s, fromType, toType)
	if caster == nil {
		e := invalidCastErr(s, fromType, toType)
		return func(from F) (to T, err error) {
			return to, e
		}
	}
	return func(from F) (to T, err error) {
		// 当from包含指针时，这部分指针指向的内容会逃逸
		escape(from)
		err = caster(noEscape(unsafe.Pointer(&from)), noEscape(unsafe.Pointer(&to)))
		return to, err
	}
}

// ReflectCastWithScope 以反射的方式，需输入待转换的值与要转换的类型
func ReflectCastWithScope(s *Scope, from reflect.Value, toType reflect.Type) (to reflect.Value, err error) {
	if toType == nil {
		return reflect.Value{}, nilToTypeErr
	}
	if !from.IsValid() {
		return reflect.Zero(toType), nil
	}
	fromType := from.Type()
	caster, _ := getCaster(s, fromType, toType)
	if caster == nil {
		return reflect.Value{}, invalidCastErr(s, fromType, toType)
	}
	toPtr := reflect.New(toType)
	err = caster(getValueAddr(from), toPtr.UnsafePointer())
	return toPtr.Elem(), err
}

// WithCaster 注册自定义转换器，只能注册到新的作用域里，避免全局污染。允许传入nil，表示禁止这两个类型之间的转换
func WithCaster[F any, T any](caster func(s *Scope, from F) (to T, err error)) ScopeOption {
	fromType, toType := typeFor[F](), typeFor[T]()
	key := casterKey{fromTypePtr: typePtr(fromType), toTypePtr: typePtr(toType)}
	return func(s *Scope) {
		if s.frozen {
			return
		}
		var wrappedCaster castFunc
		if caster != nil {
			wrappedCaster = func(fromAddr, toAddr unsafe.Pointer) error {
				var err error
				*(*T)(toAddr), err = caster(s, *(*F)(fromAddr))
				return err
			}
		}
		s.casterMap[key] = casterValue{wrappedCaster, false}
	}
}

// WithDisableZeroCopy 禁用内存布局相同时零拷贝强转，但是类型相同时仍只会浅拷贝
func WithDisableZeroCopy() ScopeOption {
	return func(s *Scope) {
		if s.frozen {
			return
		}
		s.disableZeroCopy = true
	}
}

// WithDeepCopy 所有转换强制深拷贝
func WithDeepCopy() ScopeOption {
	return func(s *Scope) {
		if s.frozen {
			return
		}
		s.deepCopy = true
	}
}

// WithUnexportedFields 转换结构体的未导出字段
func WithUnexportedFields() ScopeOption {
	return func(s *Scope) {
		if s.frozen {
			return
		}
		s.castUnexported = true
	}
}
