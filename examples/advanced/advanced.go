package main

import (
	"fmt"
	"time"

	"github.com/china-tjj/cast"
)

func CastTimeToString(s *cast.Scope, t time.Time) (string, error) {
	return t.Format("2006-01-02 15:04:05"), nil
}

func main() {
	scope := cast.NewScope(cast.WithCaster(CastTimeToString))
	t := time.Now()

	str, err := cast.CastWithScope[time.Time, string](scope, t)
	fmt.Println(str, err) // 2025-10-04 19:55:54 <nil>
}
