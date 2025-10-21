package cast

import (
	"testing"
)

func TestDeepCopy1(t *testing.T) {
	m := map[string]interface{}{
		"1": "1",
		"2": map[string]interface{}{
			"3": "3",
		},
	}
	m2, err := DeepCopy(m)
	if err != nil {
		t.Fatal(err)
	}
	m["1"] = "11"
	if m2["1"] != "1" {
		t.Fatal()
	}
	m["2"].(map[string]interface{})["3"] = "33"
	if m2["2"].(map[string]interface{})["3"] != "3" {
		t.Fatal()
	}
}

func TestDeepCopy2(t *testing.T) {
	type S struct {
		V1 *int
		V2 *string
		V3 *float64
	}
	s := &S{
		V1: ptr(1),
		V2: ptr("2"),
		V3: ptr(3.),
	}
	s2, err := DeepCopy(s)
	if err != nil {
		t.Fatal(err)
	}
	*s.V1 = 11
	*s.V2 = "22"
	*s.V3 = 33
	if *s2.V1 != 1 || *s2.V2 != "2" || *s2.V3 != 3 {
		t.Fatal()
	}
}

func TestDeepCopy3(t *testing.T) {
	_, err := DeepCopy[chan int](nil)
	if err == nil || err.Error() != "invalid deep copy: can't deep copy type 「chan int」" {
		t.Fatal(err)
	}
}
