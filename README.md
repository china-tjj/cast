# cast
泛型cast库
```go
func Cast[F any, T any](from F) (to T, ok bool)
```
## cast规则
1. 若`F`和`T`的内存布局完全相同，则直接强转，具体实现见[utils.go](utils.go)
   ```go
   func isMemSame(ta, tb reflect.Type) bool
   ```
2. `bool` `int` `uint` `uintptr` `float`可以互转
3. `array` `slice`与`bool` `int` `uint` `uintptr` `float`互转：当`len == 1`时，尝试用`array[0]`或`slice[0]`进行互转
4. `struct`与`bool` `int` `uint` `uintptr` `float`互转：若`struct`只有一个字段，则尝试用这个字段去互转
5. `string`与`bool` `int` `uint` `uintptr` `float`互转: 调用`strconv.ParseXXX`和`strconv.FormatXXX`。其中，若`strconv.ParseBool`出现`err != nil`，则改为`len(str) > 0`
6. `array` `slice`可以互转，若元素满足`isMemSame`，则不会发生拷贝
7. `string`与`array` `slice`互转时，视为`byte[]`
8. `interface`转非`interface`时，用实际的值进行转换
9. 某类型转`interface`时，要求实现该接口才能转
10. `pointer`与非`interface`互转时，尝试用`pointer`指向的内容去互转
11. `struct`与`map`互转时，`field`对应的`string-key`优先`json-tag`，其次为`field-name`，分别将`string-key`与`map-key`互转，`field-value`与`map-value`互转