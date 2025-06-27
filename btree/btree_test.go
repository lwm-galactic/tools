package btree

import (
	"fmt"
	"testing"
)

func TestBTree(t *testing.T) {
	// 创建一个 BTree，键为 int，值为 string
	bt := New[int](4, func(a, b int) bool {
		return a < b
	})
	// 插入数据
	bt.Insert(10, "ten")
	bt.Insert(5, "five")
	bt.Insert(15, "fifteen")

	// 查询数据
	if val, ok := bt.Get(5); ok {
		fmt.Println("Get 5:", val)
	}

	// 遍历所有数据
	fmt.Println("All items:")
	bt.Ascend(func(k int, v interface{}) {
		fmt.Printf("%d -> %s\n", k, v)
	})

	// 范围查询 [5, 10]
	fmt.Println("Range [5, 10]:")
	bt.AscendRange(5, 10, func(k int, v interface{}) {
		fmt.Printf("%d -> %s\n", k, v)
	})

	// 删除一个键
	bt.Delete(5)

	// 批量插入
	bt.BatchInsert(map[int]interface{}{
		1: "one",
		2: "two",
		3: "three",
	})

	// 再次遍历
	fmt.Println("After batch insert:")
	bt.Ascend(func(k int, v interface{}) {
		fmt.Printf("%d -> %s\n", k, v)
	})
}
