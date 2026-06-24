package cast

const (
	flagHasRef        = uint8(1 << iota) // 转换后，可不可以访问到from的地址，若可以且该内存只读，需拷贝
	flagRequireInHeap                    // 是否要求from在堆上，是flagHasRef的必要条件
)

func isHasRef(flag uint8) bool {
	return flag&flagHasRef != 0
}

func isRequireInHeap(flag uint8) bool {
	return flag&flagRequireInHeap != 0
}
