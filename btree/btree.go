package btree

import "github.com/google/btree"

// BTree 是一个非线程安全的 B+ 树实现，用于高性能场景
type BTree[K comparable] struct {
	tree *btree.BTree
	less LessFunc[K]
}

// NewNoLock 创建一个新的非线程安全的 B+ 树实例
func NewNoLock[K comparable](degree int, less LessFunc[K]) *BTree[K] {
	if degree < 2 {
		panic("degree must >= 2")
	}
	return &BTree[K]{
		tree: btree.New(degree),
		less: less,
	}
}

// Insert 插入键值对（无锁）
func (bt *BTree[K]) Insert(key K, value interface{}) {
	item := &keyValueItem[K]{Key: key, Value: value, less: bt.less}
	bt.tree.ReplaceOrInsert(item)
}

// Get 查找键对应的值（无锁）
func (bt *BTree[K]) Get(key K) (interface{}, bool) {
	item := &keyValueItem[K]{Key: key, less: bt.less}
	found := bt.tree.Get(item)
	if found == nil {
		return nil, false
	}
	return found.(*keyValueItem[K]).Value, true
}

// Delete 删除指定键（无锁）
func (bt *BTree[K]) Delete(key K) {
	item := &keyValueItem[K]{Key: key, less: bt.less}
	bt.tree.Delete(item)
}

// Size 返回当前树中的元素数量（无锁）
func (bt *BTree[K]) Size() int {
	return bt.tree.Len()
}

// IsEmpty 判断是否为空（无锁）
func (bt *BTree[K]) IsEmpty() bool {
	return bt.Size() == 0
}

// Ascend 按升序遍历所有元素（无锁）
func (bt *BTree[K]) Ascend(fn func(K, interface{})) {
	bt.tree.Ascend(func(i btree.Item) bool {
		item := i.(*keyValueItem[K])
		fn(item.Key, item.Value)
		return true
	})
}

// AscendRange 遍历指定范围 [start, end]（无锁）
func (bt *BTree[K]) AscendRange(start, end K, fn func(K, interface{})) {
	startItem := &keyValueItem[K]{Key: start, less: bt.less}
	endItem := &keyValueItem[K]{Key: end, less: bt.less}

	bt.tree.AscendRange(startItem, endItem, func(i btree.Item) bool {
		item := i.(*keyValueItem[K])
		fn(item.Key, item.Value)
		return true
	})
}

// Descend 按降序遍历所有元素（无锁）
func (bt *BTree[K]) Descend(fn func(K, interface{})) {
	bt.tree.Descend(func(i btree.Item) bool {
		item := i.(*keyValueItem[K])
		fn(item.Key, item.Value)
		return true
	})
}

// BatchInsert 批量插入键值对（无锁）
func (bt *BTree[K]) BatchInsert(pairs map[K]interface{}) {
	for k, v := range pairs {
		item := &keyValueItem[K]{Key: k, Value: v, less: bt.less}
		bt.tree.ReplaceOrInsert(item)
	}
}
