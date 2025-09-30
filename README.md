# cast

一个高性能、类型安全、支持复杂结构转换的 Go 泛型类型转换库。

***

## 简介

`cast` 是一个基于泛型的通用类型转换工具，旨在简化 Go 语言中繁琐的类型转换，在保证类型安全与易用性的同时，性能开销不超过手写转换的
2 倍。

### 核心特性

* **零拷贝支持**：对于内存布局一致的类型，通过指针重解释实现零拷贝。
* **低反射开销**：仅在首次构建转换器时使用反射；后续调用均通过缓存的转换函数执行，仅在极少数边缘场景下保留必要的反射逻辑。
* **递归支持复杂类型**：支持嵌套结构体、多重指针、map、slice、array、func、interface 等复合类型的转换。
* **类型安全的泛型接口**：提供 `Cast[F, T]` 泛型函数，编译期即可校验输入输出类型，避免运行时类型断言。

***

## 核心函数

```go
func Cast[F any, T any](from F) (to T, err error)

func GetCaster[F any, T any]() func (from F) (to T, err error)
```

* `Cast[F, T]`：尝试将类型 `F` 的值 `from` 转换为类型 `T`。返回转换结果与错误信息。
* `GetCaster[F, T]`：返回转换器，避免每次调用时查询缓存，适用于高频转换场景，性能略优于直接调用 `Cast`。

***

## 适用场景

1. **配置解析**：将 `map[string]interface{}` 高效转换为目标结构体，适用于配置加载、动态插件等场景。
2. **跨模块结构体桥接**：不同服务/模块间存在结构体重复定义时，实现零拷贝互通。

***

## 快速开始

### 安装

```text
go get github.com/china-tjj/cast
```

### 示例

```go
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
```

***

## 性能表现

以下为结构体转换的基准测试，对比手写转换与 `cast` 的性能：

```go
package main

import (
	"github.com/china-tjj/cast"
	"strconv"
	"testing"
)

type S1 struct {
	V1 int
	V2 []string
}

type S2 struct {
	V1 *string
	V2 []float64
}

func handCaster(s1 *S1) (*S2, error) {
	v1 := strconv.Itoa(s1.V1)
	v2 := make([]float64, 0, len(s1.V2))
	for _, v := range s1.V2 {
		newV, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, err
		}
		v2 = append(v2, newV)
	}
	return &S2{
		V1: &v1,
		V2: v2,
	}, nil
}

func BenchmarkStructCast(b *testing.B) {
	s1 := &S1{
		V1: 1,
		V2: []string{"1", "2"},
	}
	b.Run("HandCast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = handCaster(s1)
		}
	})
	b.Run("GetCaster", func(b *testing.B) {
		caster := cast.GetCaster[*S1, *S2]()
		for i := 0; i < b.N; i++ {
			_, _ = caster(s1)
		}
	})
	b.Run("Cast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = cast.Cast[*S1, *S2](s1)
		}
	})
}
```

**测试结果（Apple M3 Pro, Go 1.23）**：

```text
BenchmarkStructCast/HandCast-12         	19024240	        62.02 ns/op
BenchmarkStructCast/GetCaster-12        	12475611	        95.67 ns/op
BenchmarkStructCast/Cast-12             	11074729	        110.6 ns/op
```

* `GetCaster` 的性能约为手写转换的 1.5 倍，`Cast` 约为 1.8 倍。
* 在大多数场景下，性能损耗完全可以接受，且换来的是极高的开发效率与类型安全性。

***

## 转换规则详解

转换规则可递归应用于复合类型（如 struct、slice、map 等）。以下规则按优先级和逻辑组织：

### 1. 零拷贝强转（内存布局一致）

当类型 `F` 与 `T` 的底层内存布局完全一致时，通过 `unsafe` 指针重解释实现零拷贝转换。判定条件如下：

* 基础类型（`reflect.Kind`）必须相同。
* 复合类型递归判定：
    * **指针**：指向类型的内存布局一致；`unsafe.Pointer` 与任意指针类型视为兼容。
    * **数组**：长度相同，且元素类型内存布局一致。
    * **切片**：元素类型内存布局一致。
    * **map**：键类型与值类型分别满足内存布局一致。
    * **channel**：方向兼容（双向可转单向），且元素类型内存布局一致。
    * **struct**：字段数量一致，每个字段满足：
        * 字段名相同 **或** `json` tag 相同；
        * 对应字段类型内存布局一致。
    * **interface**：两个接口的方法集互为子集（即互相实现）。
    * **func**：参数和返回值的数量、顺序一致，且对应类型内存布局一致。

### 2. 基础类型互转

* `bool`、各类 `int`、`uint`、`float` 系列、`string` 之间支持相互转换。
* 基于原生转换与标准库 `strconv` 实现，遵循标准语义（如 `"123"` → `123`，`"true"` → `true`）。

### 3. 序列类型互转

* 数组 `[N]F` 与切片 `[]T` 可互转，前提是 `F` 与 `T` 可转换；若内存布局一致，则支持零拷贝。
* `string` 与 `[]byte` / `[N]byte` 支持零拷贝互转。
* `string` 与 `[]rune` / `[N]rune` 使用原生互转（非零拷贝）。
* `string` 与其他 `[]T` / `[N]T`互转时，`string` 被视为 `[]byte` 处理。

### 4. 指针与非接口类型互转

* 若 `F` 和 `T` 可互转，其各自的多重指针之间也可以互转（如 `*int` 和 `****string` 可互转）。

### 5. Map 转换

* `map[K1]V1` → `map[K2]V2`：要求 `K1`→`K2` 与 `V1`→`V2` 均可转换。
* `map` → `struct`：
    * 优先使用字段的 `json` tag 作为键名；
    * 若未匹配，则回退至字段名；
    * 成功匹配的字段值将按规则转换并赋值。
* `struct` → `map`：
    * 键名优先使用 `json` tag，其次使用字段名；
    * 字段值按转换规则映射为 map 的值。

### 6. Struct 互转

* 字段映射优先依据 `json` tag 对齐；
* 未通过 tag 匹配的字段，尝试使用字段名对齐；
* 所有成功对齐的字段必须满足类型可转换。

### 7. Interface 转换

* **目标为 interface**：要求源值的动态类型实现该接口。
* **源为 interface**：使用其动态值（即“拆箱”后）进行后续转换。

### 8. Func 转换

* `func(A1, A2) (R1, R2)` → `func(B1, B2) (S1, S2)`：
    * 参数数量与返回值数量必须一致；
    * 每个 `Ai` 可转换为 `Bi`，每个 `Ri` 可转换为 `Si`；
    * 转换后的函数在调用时自动完成参数与返回值的类型适配。

### 9. 错误判定机制

* 递归转换时，只要有一处出现 error，则整体返回 error，该行为与标准库一致。
