package main

import (
	"fmt"
	"github.com/china-tjj/cast"
)

type Config struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func main() {
	input := map[string]interface{}{
		"host": "localhost",
		"port": 8080,
	}

	cfg, err := cast.Cast[any, Config](input)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Config: %+v\n", cfg) // Config: {Host:localhost Port:8080}
}
