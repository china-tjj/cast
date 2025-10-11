package cast

import "unsafe"

type iMapHelper interface {
	Make(mapAddr unsafe.Pointer, n int) map[any]any
	Load(m map[any]any, key unsafe.Pointer) (value unsafe.Pointer, ok bool)
	Store(m map[any]any, key, value unsafe.Pointer)
	Range(m map[any]any, f func(key, value unsafe.Pointer) bool)
}
