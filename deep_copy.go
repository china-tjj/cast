package cast

var deepCopyScope = NewScope(WithDeepCopy())

// DeepCopy 深拷贝
func DeepCopy[T any](v T) (T, error) {
	return CastWithScope[T, T](deepCopyScope, v)
}

// GetDeepCopier 获取类型 T 的深拷贝函数，对比直接调用 DeepCopy 少了查缓存的步骤，性能会略微好一点
func GetDeepCopier[T any]() func(T) (T, error) {
	return GetCasterWithScope[T, T](deepCopyScope)
}
