// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
	"unsafe"
)

var defaultScope = NewScope()

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
	caster := getCaster(s, fromType, toType)
	if caster == nil {
		return to, invalidCastErr(fromType, toType)
	}
	err = caster(unsafe.Pointer(&from), unsafe.Pointer(&to))
	return to, err
}

// GetCasterWithScope 获取实例化的转换方法，调用该方法返回的函数，对比直接调用 Cast 少了查缓存的步骤，性能会略微好一点
func GetCasterWithScope[F any, T any](s *Scope) func(from F) (to T, err error) {
	fromType, toType := typeFor[F](), typeFor[T]()
	caster := getCaster(s, fromType, toType)
	if caster == nil {
		e := invalidCastErr(fromType, toType)
		return func(from F) (to T, err error) {
			return to, e
		}
	}
	return func(from F) (to T, err error) {
		err = caster(unsafe.Pointer(&from), unsafe.Pointer(&to))
		return to, err
	}
}

// ReflectCastWithScope 以反射的方式，需输入待转换的值与要转换的类型
func ReflectCastWithScope(s *Scope, from reflect.Value, toType reflect.Type) (to reflect.Value, err error) {
	fromType := from.Type()
	caster := getCaster(s, fromType, toType)
	if caster == nil {
		return to, invalidCastErr(fromType, toType)
	}
	fromAddr := getValueAddr(from)
	to = reflect.New(toType).Elem()
	if fromAddr == nil {
		return to, nil
	}
	err = caster(fromAddr, getValueAddr(to))
	return to, err
}

// WithCaster 注册自定义转换器，只能注册到新的作用域里，避免全局污染。允许传入nil，表示禁止这两个类型之间的转换
func WithCaster[F any, T any](caster func(from F) (to T, err error)) ScopeOption {
	var wrappedCaster castFunc
	if caster != nil {
		wrappedCaster = func(fromAddr, toAddr unsafe.Pointer) error {
			var err error
			*(*T)(toAddr), err = caster(*(*F)(fromAddr))
			return err
		}
	}
	fromType, toType := typeFor[F](), typeFor[T]()
	cacheKey := casterKey{fromTypePtr: typePtr(fromType), toTypePtr: typePtr(toType)}
	return func(s *Scope) {
		if s.frozen {
			return
		}
		s.casterMap[cacheKey] = wrappedCaster
	}
}

// WithDisableZeroCopy 禁用零拷贝，所有场景均会浅拷贝
func WithDisableZeroCopy() ScopeOption {
	return func(s *Scope) {
		if s.frozen {
			return
		}
		s.disableZeroCopy = true
	}
}

// SetDefaultScope ！！慎用！！设置默认作用域，可以改变默认行为
func SetDefaultScope(s *Scope) {
	defaultScope = s
}
