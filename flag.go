package cast

const (
	flagHasRef = uint8(1 << iota) // 转换后，可不可以访问到from的地址，若可以且该内存只读，需拷贝
	flagCustom                    // 是否是用户自定义
)

func isHasRef(flag uint8) bool {
	return flag&flagHasRef != 0
}

func isCustom(flag uint8) bool {
	return flag&flagCustom != 0
}
