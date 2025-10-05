package cast

import (
	"errors"
	"fmt"
	"testing"
	"unsafe"
)

func testHash[T comparable](t *testing.T, key T) {
	m := make(map[T]struct{})
	m[key] = struct{}{}
	var ok bool
	switch unsafe.Sizeof(key) {
	case 0:
		_, ok = ref[map[T]struct{}, map[[0]byte]struct{}](m)[ref[T, [0]byte](key)]
	case 1:
		_, ok = ref[map[T]struct{}, map[[1]byte]struct{}](m)[ref[T, [1]byte](key)]
	case 2:
		_, ok = ref[map[T]struct{}, map[[2]byte]struct{}](m)[ref[T, [2]byte](key)]
	case 3:
		_, ok = ref[map[T]struct{}, map[[3]byte]struct{}](m)[ref[T, [3]byte](key)]
	case 4:
		_, ok = ref[map[T]struct{}, map[[4]byte]struct{}](m)[ref[T, [4]byte](key)]
	case 5:
		_, ok = ref[map[T]struct{}, map[[5]byte]struct{}](m)[ref[T, [5]byte](key)]
	case 6:
		_, ok = ref[map[T]struct{}, map[[6]byte]struct{}](m)[ref[T, [6]byte](key)]
	case 7:
		_, ok = ref[map[T]struct{}, map[[7]byte]struct{}](m)[ref[T, [7]byte](key)]
	case 8:
		_, ok = ref[map[T]struct{}, map[[8]byte]struct{}](m)[ref[T, [8]byte](key)]
	case 9:
		_, ok = ref[map[T]struct{}, map[[9]byte]struct{}](m)[ref[T, [9]byte](key)]
	case 10:
		_, ok = ref[map[T]struct{}, map[[10]byte]struct{}](m)[ref[T, [10]byte](key)]
	case 11:
		_, ok = ref[map[T]struct{}, map[[11]byte]struct{}](m)[ref[T, [11]byte](key)]
	case 12:
		_, ok = ref[map[T]struct{}, map[[12]byte]struct{}](m)[ref[T, [12]byte](key)]
	case 13:
		_, ok = ref[map[T]struct{}, map[[13]byte]struct{}](m)[ref[T, [13]byte](key)]
	case 14:
		_, ok = ref[map[T]struct{}, map[[14]byte]struct{}](m)[ref[T, [14]byte](key)]
	case 15:
		_, ok = ref[map[T]struct{}, map[[15]byte]struct{}](m)[ref[T, [15]byte](key)]
	case 16:
		_, ok = ref[map[T]struct{}, map[[16]byte]struct{}](m)[ref[T, [16]byte](key)]
	case 17:
		_, ok = ref[map[T]struct{}, map[[17]byte]struct{}](m)[ref[T, [17]byte](key)]
	case 18:
		_, ok = ref[map[T]struct{}, map[[18]byte]struct{}](m)[ref[T, [18]byte](key)]
	case 19:
		_, ok = ref[map[T]struct{}, map[[19]byte]struct{}](m)[ref[T, [19]byte](key)]
	case 20:
		_, ok = ref[map[T]struct{}, map[[20]byte]struct{}](m)[ref[T, [20]byte](key)]
	case 21:
		_, ok = ref[map[T]struct{}, map[[21]byte]struct{}](m)[ref[T, [21]byte](key)]
	case 22:
		_, ok = ref[map[T]struct{}, map[[22]byte]struct{}](m)[ref[T, [22]byte](key)]
	case 23:
		_, ok = ref[map[T]struct{}, map[[23]byte]struct{}](m)[ref[T, [23]byte](key)]
	case 24:
		_, ok = ref[map[T]struct{}, map[[24]byte]struct{}](m)[ref[T, [24]byte](key)]
	default:
		t.Fatal("too big")
	}
	if ok == isSpecialHash(typeFor[T]()) {
		t.Fatal(fmt.Sprintf("%T not expect", key))
	}
}

func TestHash(t *testing.T) {
	testHash(t, 1)
	testHash(t, "1")
	testHash(t, [1]string{"1"})
	testHash(t, errors.New("1"))
	testHash(t, 1)
	testHash(t, ptr(1))
}
