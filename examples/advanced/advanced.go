package main

import (
	"fmt"
	"github.com/china-tjj/cast"
	"time"
)

func CastTimeToString(t time.Time) (string, error) {
	return t.Format(time.DateTime), nil
}

func main() {
	scope := cast.NewScope(cast.WithCaster(CastTimeToString))
	t := time.Now()

	str, err := cast.CastWithScope[time.Time, string](scope, t)
	fmt.Println(str, err) // 2025-10-04 19:55:54 <nil>

	bytes, err := cast.CastWithScope[time.Time, []byte](scope, t)
	fmt.Println(string(bytes), err) // 2025-10-04 19:55:54 <nil>
}
