// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"reflect"
)

// To 将任意类型转为T, 相当于 Cast[any, T], 转换规则详见README
func To[T any](from any) (T, error) {
	return ToWithScope[T](defaultScope, from)
}

// Cast 将类型F转为T, 转换规则详见README
func Cast[F any, T any](from F) (to T, err error) {
	return CastWithScope[F, T](defaultScope, from)
}

// GetCaster 获取实例化的转换方法, 对比直接调用 Cast 少了查缓存的步骤, 性能会略微好一点
func GetCaster[F any, T any]() func(from F) (to T, err error) {
	return GetCasterWithScope[F, T](defaultScope)
}

// MustGetCaster 类似于 GetCaster, 区别为: 当这两个类型之间的转换不合法时, GetCaster 会返回一个必定返回 err 的转换器, MustGetCaster 会 panic
func MustGetCaster[F any, T any]() func(from F) (to T, err error) {
	return MustGetCasterWithScope[F, T](defaultScope)
}

// ReflectCast 使用反射将 from 反射值转换为 toType 对应的反射值，转换规则详见 README
func ReflectCast(from reflect.Value, toType reflect.Type) (to reflect.Value, err error) {
	return ReflectCastWithScope(defaultScope, from, toType)
}
