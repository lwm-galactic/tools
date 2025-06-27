package redblacktree

// ✅ 红黑树定义特性（回顾）
// 红黑树满足以下 5 条规则：
// 每个节点要么是红色，要么是黑色。
// 根节点是黑色。
// 所有叶子节点（nil 节点）是黑色。
// 如果一个节点是红色，则它的子节点必须是黑色。（不能有两个连续的红色节点）
// 从任一节点到其每个叶子的所有路径都包含相同数目的黑色节点。

// Color 颜色常量
type Color bool

const (
	Red   Color = true
	Black Color = false
)

// Node 节点结构体
type Node[T any] struct {
	Key    T
	Value  interface{}
	Left   *Node[T]
	Right  *Node[T]
	Parent *Node[T]
	Color  Color
}

// RedBlackTree 红黑树结构体（带比较器）
type RedBlackTree[T any] struct {
	Root       *Node[T]
	nilNode    *Node[T]
	Comparator func(a, b T) int
}

// NewRedBlackTree 创建新的红黑树 自定义key 结构 和 比较函数
func NewRedBlackTree[T any](comparator func(a, b T) int) *RedBlackTree[T] {
	nilNode := &Node[T]{
		Color: Black,
	}
	return &RedBlackTree[T]{
		nilNode:    nilNode,
		Comparator: comparator,
	}
}

// 新建一个空节点
func (tree *RedBlackTree[T]) newNode(key T, value interface{}) *Node[T] {
	return &Node[T]{
		Key:    key,
		Value:  value,
		Left:   tree.nilNode,
		Right:  tree.nilNode,
		Parent: tree.nilNode,
		Color:  Red,
	}
}

func (tree *RedBlackTree[T]) Insert(key T, value interface{}) {

}

func (tree *RedBlackTree[T]) Delete(key T) {

}
