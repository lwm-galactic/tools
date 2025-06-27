package btree

import (
	"fmt"
	"github.com/google/btree"
	"sync"
)

// KeyValue 是一个键值对结构
type KeyValue[K comparable] struct {
	Key   K
	Value interface{}
}

// LessFunc 是自定义比较函数类型
type LessFunc[K comparable] func(a, b K) bool

// BTreeLock 是封装后的 B+ 树结构,加锁线程安全
type BTreeLock[K comparable] struct {
	tree  *btree.BTree
	mutex *sync.RWMutex // 读写锁保证map插入线程安全
	less  LessFunc[K]
}

// New 创建一个新的 BTreeLock 实例 (泛型函数)
// 可以传入任意做key 需要传入一个比较传入泛型的比较方法
func New[K comparable](degree int, less LessFunc[K]) *BTreeLock[K] {
	if degree < 2 {
		panic("degree must >= 2")
	}
	return &BTreeLock[K]{
		tree:  btree.New(degree),
		mutex: &sync.RWMutex{},
		less:  less,
	}
}

// Insert 插入一个键值对
func (bt *BTreeLock[K]) Insert(key K, value interface{}) {
	item := &keyValueItem[K]{Key: key, Value: value, less: bt.less}
	bt.mutex.Lock()
	defer bt.mutex.Unlock()
	bt.tree.ReplaceOrInsert(item)
}

// Get 查找一个键对应的值
func (bt *BTreeLock[K]) Get(key K) (interface{}, bool) {
	item := &keyValueItem[K]{Key: key, less: bt.less}
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()
	found := bt.tree.Get(item)
	if found == nil {
		return nil, false
	}
	return found.(*keyValueItem[K]).Value, true
}

// Delete 删除一个键
func (bt *BTreeLock[K]) Delete(key K) {
	item := &keyValueItem[K]{Key: key, less: bt.less}
	bt.mutex.Lock()
	defer bt.mutex.Unlock()
	bt.tree.Delete(item)
}

// Size 返回当前树中的元素数量
func (bt *BTreeLock[K]) Size() int {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()
	return bt.tree.Len()
}

// IsEmpty 判断是否为空
func (bt *BTreeLock[K]) IsEmpty() bool {
	return bt.Size() == 0
}

// Ascend 按升序遍历所有元素
func (bt *BTreeLock[K]) Ascend(fn func(key K, value interface{})) {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()
	bt.tree.Ascend(func(i btree.Item) bool {
		item := i.(*keyValueItem[K])
		fn(item.Key, item.Value)
		return true
	})
}

// AscendRange 遍历指定范围 [start, end]
func (bt *BTreeLock[K]) AscendRange(start, end K, fn func(K, interface{})) {
	startItem := &keyValueItem[K]{Key: start, less: bt.less}
	endItem := &keyValueItem[K]{Key: end, less: bt.less}

	bt.mutex.RLock()
	defer bt.mutex.RUnlock()

	bt.tree.AscendRange(startItem, endItem, func(i btree.Item) bool {
		item := i.(*keyValueItem[K])
		fn(item.Key, item.Value)
		return true
	})
}

// Descend 按降序遍历所有元素
func (bt *BTreeLock[K]) Descend(fn func(key K, value interface{})) {
	bt.mutex.RLock()
	defer bt.mutex.RUnlock()
	bt.tree.Descend(func(i btree.Item) bool {
		item := i.(*keyValueItem[K])
		fn(item.Key, item.Value)
		return true
	})
}

// BatchInsert 批量插入键值对
func (bt *BTreeLock[K]) BatchInsert(pairs map[K]interface{}) {
	bt.mutex.Lock()
	defer bt.mutex.Unlock()
	for k, v := range pairs {
		item := &keyValueItem[K]{Key: k, Value: v, less: bt.less}
		bt.tree.ReplaceOrInsert(item)
	}
}

// keyValueItem 是内部使用的 Item 类型
type keyValueItem[K comparable] struct {
	Key   K
	Value interface{}
	less  LessFunc[K]
}

// Less 实现 btree.Item 接口
func (i *keyValueItem[K]) Less(than btree.Item) bool {
	other := than.(*keyValueItem[K])
	return i.less(i.Key, other.Key)
}

// String 返回字符串表示
func (i *keyValueItem[K]) String() string {
	return fmt.Sprintf("{Key: %v}", i.Key)
}
