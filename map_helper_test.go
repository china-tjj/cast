package cast

import (
	"testing"
	"unsafe"
)

func BenchmarkMapHelper(b *testing.B) {
	b.Run("use-helper", doBenchmark)

	defaultScope.casterMap = make(map[casterKey]castFunc)
	makers, newers := helperMaker, helperNewer
	var zero [helperMaxSize + 1][helperMaxSize + 1]func(mapAddr unsafe.Pointer) mapHelper
	helperMaker, helperNewer = zero, zero
	defer func() {
		helperMaker, helperNewer = makers, newers
	}()

	b.Run("ban-helper", doBenchmark)
}

func doBenchmark(b *testing.B) {
	b.Run("map[int]int->map[string]string", func(b *testing.B) {
		m := map[int]int{
			0: 0,
			1: 1,
			2: 2,
			3: 3,
			4: 4,
			5: 5,
		}
		caster := GetCaster[map[int]int, map[string]string]()
		for i := 0; i < b.N; i++ {
			_, _ = caster(m)
		}
	})
}
