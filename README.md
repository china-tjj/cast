# cast

一个高性能、类型安全、支持复杂结构转换的 Go 泛型类型转换库。

> 目前版本为 v0.1.4，处于开发阶段，存在潜在 bug，不建议用于生产环境。欢迎通过 Issues 反馈问题，
> 稳定版（v1.0.0）将在充分验证后发布。

## 简介

`cast` 是一个基于泛型的通用类型转换工具，旨在简化 Go 语言中繁琐的类型转换，在保证类型安全与易用性的同时，性能接近手写转换。

### 核心特性

* **零拷贝支持**：对于内存布局一致的类型，通过指针重解释实现零拷贝。
* **低反射开销**：仅在首次构建转换器时使用反射；后续调用均通过缓存的转换函数执行，仅在极少数边缘场景下保留必要的反射逻辑。
* **递归支持复杂类型**：支持嵌套结构体、多重指针、map、slice、array、func、interface 等复合类型的转换。
* **类型安全的泛型接口**：提供 `Cast[F, T]` 泛型函数，编译期即可校验输入输出类型，避免运行时类型断言。

## 核心函数

```go
func Cast[F any, T any](from F) (to T, err error)

func GetCaster[F any, T any]() func (from F) (to T, err error)

func MustGetCaster[F any, T any]() func(from F) (to T, err error)
```

* `Cast[F, T]`：尝试将类型 `F` 的值 `from` 转换为类型 `T`。返回转换结果与错误信息。
* `GetCaster[F, T]`：返回转换器，避免每次调用时查询缓存，适用于高频转换场景，性能略优于直接调用 `Cast`。
* `MustGetCaster[F, T]`：类似于 GetCaster，区别为：当这两个类型之间的转换不合法时，GetCaster 会返回一个必定返回 err 的转换器，MustGetCaster 会 panic

## 适用场景

1. **配置解析**：将 `map[string]interface{}` 高效转换为目标结构体，适用于配置加载、动态插件等场景。
2. **跨模块结构体桥接**：不同服务/模块间存在结构体重复定义时，实现零拷贝互通。

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

## 性能表现

以下为结构体转换的基准测试，对比手写转换与 `cast` 的性能：

```go
package main

import (
	"github.com/china-tjj/cast"
	"strconv"
	"testing"
)

type FromStruct struct {
	V1 float64
	V2 []complex128
	V3 map[int]int
}

type ToStruct struct {
	V1 *string
	V2 []*string
	V3 map[string]*string
}

func ManualCast(from *FromStruct) (*ToStruct, error) {
	v1 := strconv.FormatFloat(from.V1, 'f', -1, 64)
	v2 := make([]*string, len(from.V2))
	for i, v := range from.V2 {
		newV := strconv.FormatComplex(v, 'f', -1, 128)
		v2[i] = &newV
	}
	v3 := map[string]*string{}
	for k, v := range from.V3 {
		newK := strconv.Itoa(k)
		newV := strconv.Itoa(v)
		v3[newK] = &newV
	}
	return &ToStruct{
		V1: &v1,
		V2: v2,
		V3: v3,
	}, nil
}

func BenchmarkStructCast(b *testing.B) {
	from := &FromStruct{
		V1: 1,
		V2: []complex128{2 + 3i, 4 + 5i, 6 + 7i, 8 + 9i},
		V3: map[int]int{10: 11, 12: 13, 14: 15, 16: 17, 18: 19},
	}
	b.Run("ManualCast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ManualCast(from)
		}
	})
	b.Run("GetCaster", func(b *testing.B) {
		caster := cast.GetCaster[*FromStruct, *ToStruct]()
		for i := 0; i < b.N; i++ {
			_, _ = caster(from)
		}
	})
	b.Run("Cast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = cast.Cast[*FromStruct, *ToStruct](from)
		}
	})
}
```

**测试结果（Apple M3 Pro, Go 1.21）**：

```text
BenchmarkStructCast/ManualCast-12         	 1393200	       850.0 ns/op
BenchmarkStructCast/GetCaster-12          	 1208700	       983.4 ns/op
BenchmarkStructCast/Cast-12               	 1000000	       1009 ns/op
```

在上面的例子里：

* `GetCaster` 的性能约为手写转换的 1.15 倍，`Cast` 约为 1.18 倍。
* `GetCaster` 的性能损耗主要是因为闭包无法内联、一些堆内存分配与读取，`Cast` 还有查询缓存的开销，因此整体会逊色于手写转换与代码生成。
* 在大多数场景下，性能损耗完全可以接受，且换来的是极高的开发效率与类型安全性。

## 高级用法-作用域

作用域是 `cast`
库中用于隔离转换规则的重要机制。在复杂项目中，不同模块可能需要不同的转换规则（例如，某些模块需要自定义时间格式转换，而其他模块则需要标准转换方式），通过作用域可以为不同上下文提供独立的转换规则，避免全局设置导致的意外行为。

#### 创建自定义作用域

可以通过 `cast.NewScope` 函数创建新的作用域，并进行自定义配置：

```go
scope := cast.NewScope(
    cast.WithCaster(CastTimeToString), // 添加自定义时间转换器
    cast.WithDisableZeroCopy(),        // 禁用零拷贝
    // 可添加更多配置选项
)
```

#### 使用自定义作用域

在转换时，可以指定使用特定的作用域：

```go
str, err := cast.CastWithScope[time.Time, string](scope, t)
```

或者使用 `GetCasterWithScope` 获取作用域特定的转换器：

```go
caster := cast.GetCasterWithScope[time.Time, string](scope)
str, err := caster(t)
```

#### 默认作用域

如果需要更改默认作用域（**不推荐**，除非你完全理解其影响），可以使用 `cast.SetDefaultScope`：

```go
cast.SetDefaultScope(scope)
```

### 1. 自定义转换器

支持自定义自定义转换器，例如以下示例自定义，自定义 `time.Time` 转 `string` 的格式：

```go
package main

import (
	"fmt"
	"github.com/china-tjj/cast"
	"time"
)

func CastTimeToString(s *cast.Scope, t time.Time) (string, error) {
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
```

### 2. 禁用零拷贝

零拷贝转换后，可能多个不同类型的变量共用一块内存，遇到并发问题时会加大排查难度，因此支持禁用零拷贝
（禁用内存布局相同时零拷贝强转，但是类型相同时仍只会浅拷贝），示例如下：

```go
scope := cast.NewScope(cast.WithDisableZeroCopy())
```

### 3. 深拷贝

所有转换均进行深拷贝，示例如下：

```go
scope := cast.NewScope(cast.WithDeepCopy())
```

类似于 `Cast` 和 `GetCaster`，本库简单封装了两个函数:

```go
func DeepCopy[T any](v T) (T, error)

func GetDeepCopier[T any]() func (T) (T, error)
```

### 4. 访问结构体未导出字段

本库在处理结构体相关转换时，默认会跳过未导出字段，但是支持访问结构体未导出字段，示例如下：

```go
scope := cast.NewScope(cast.WithUnexportedFields())
```

### 5. 严格 nil 检查

当源值为 nil 时（指无类型或指针类型的 nil），无论目标类型是什么，本库默认会将其转为目标类型对应的零值，支持开启严格 nil 检查，仅允许 nil 转为可以为 nil 的类型，示例如下：

```go
scope := cast.NewScope(cast.WithStrictNilCheck())
```

## 转换规则详解

转换规则可递归应用于复合类型（如 struct、slice、map 等）。以下规则按优先级和逻辑组织：

### 1. 错误判定机制

* 递归转换时，只要有一处出现 error，则整体返回 error，该行为与标准库一致

### 2. 零拷贝强转（内存布局一致）

当类型 `F` 与 `T` 的底层内存布局完全一致时，通过 `unsafe` 指针重解释实现零拷贝转换。判定条件如下：

* 基础类型（`reflect.Kind`）必须相同
* 复合类型递归判定：
    * **指针**：指向类型的内存布局一致；`unsafe.Pointer` 与任意指针类型视为兼容
    * **数组**：长度相同，且元素类型内存布局一致
    * **切片**：元素类型内存布局一致
    * **map**：键类型与值类型分别满足内存布局一致
    * **channel**：方向兼容（双向可转单向），且元素类型内存布局一致
    * **struct**：字段数量一致，每个字段满足：
        * 字段名相同或 `json` tag 相同
        * 对应字段类型内存布局一致
    * **interface**：两个接口的方法集互为子集（即互相实现）
    * **func**：参数和返回值的数量、顺序一致，且对应类型内存布局一致

### 3. 基础类型互转

* `bool`、各类 `int`、`uint`、`float` 系列、`string` 之间支持相互转换
* 支持 `complex` 系列与 `string` 互转
* 基于原生转换与标准库 `strconv` 实现，遵循标准语义
* 若类型实现了 `fmt.Stringer` 接口，在转为 `string` 时优先调用该接口

### 4. 序列类型互转

* 数组 `[N]F` 与切片 `[]T` 可互转，前提是 `F` 与 `T` 可转换；若内存布局一致，则支持零拷贝
* 当 `F` 和 `T` 内存布局相同时，`[N]F` 与 `[]F` 可以转为 `*T`，但安全起见，转换不可逆
* `string` 与 `[]byte` / `[N]byte` 支持零拷贝互转
* `string` 与 `[]rune` / `[N]rune` 使用原生互转（非零拷贝）
* `string` 与其他 `[]T` / `[N]T` 互转时，`string` 被视为 `[]byte` 处理

### 5. 指针与非接口类型互转

* 若 `F` 和 `T` 可互转，其各自的多重指针之间也可以互转（如 `*int` 和 `**string` 可互转）
* `pointer` 与非 `interface` 互转时，尝试用 `pointer` 指向的对象去互转

### 6. Map 与 Struct 转换

* `map[K1]V1` → `map[K2]V2`：要求 `K1`→`K2` 与 `V1`→`V2` 均可转换
* `struct` → `map`：
    * 键名优先使用 `json` tag，其次使用字段名
    * 字段值按转换规则映射为 map 的值，若存在无法转换的字段，则不允许整体的转换
* `map` → `struct`：
    * 键名优先使用 `json` tag，其次使用字段名
    * 成功匹配的字段值将按规则转换，若匹配成功但转换失败则会导致整体转换失败
* `struct` → `struct`：
    * 字段映射优先依据 `json` tag 匹配，其次使用字段名
    * 成功匹配的字段值将按规则转换，若匹配成功但转换失败则会导致整体转换失败

### 7. Interface 转换

* **目标为 interface**：要求源值的动态类型实现该接口
* **源为 interface**：使用其动态值（即"拆箱"后）进行后续转换

### 8. Func 转换

* `func(A1, A2) (R1, R2)` → `func(B1, B2) (S1, S2)`：
    * 参数数量与返回值数量必须一致
    * 每个 `Ai` 可转换为 `Bi`，每个 `Ri` 可转换为 `Si`
    * 转换后的函数在调用时自动完成参数与返回值的类型适配

### 9. 特殊处理

* `string` 转 `time.Time`：调用 `time.Parse`，依次尝试标准库里的所有格式进行转换
* `string` 转 `time.Duration`：若字符串中含有时间单位，则调用 `time.ParseDuration`，否则视为转为 `int64`
