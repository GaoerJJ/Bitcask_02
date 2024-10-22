package index

import (
	"Bitcask_02/biface"
	"Bitcask_02/data"
	"bytes"
	art "github.com/plar/go-adaptive-radix-tree"
	"sort"
	"sync"
)

// AdaptiveRadixTree 自适应基数树索引
type AdaptiveRadixTree struct {
	tree art.Tree
	lock *sync.RWMutex
}

// NewART 创建一个 ART 树
func NewART() *AdaptiveRadixTree {
	return &AdaptiveRadixTree{
		tree: art.New(),
		lock: new(sync.RWMutex),
	}
}

// Put 插入
func (art *AdaptiveRadixTree) Put(key []byte, pos *data.LogRecordPos) bool {
	art.lock.Lock()
	art.tree.Insert(key, pos)
	art.lock.Unlock()
	return true // todo 为什么 insert 不会出错?
}

func (art *AdaptiveRadixTree) Get(key []byte) *data.LogRecordPos {
	art.lock.RLock()
	defer art.lock.RUnlock()
	value, found := art.tree.Search(key)
	if !found {
		return nil
	}
	return value.(*data.LogRecordPos)
}

func (art *AdaptiveRadixTree) Delete(key []byte) bool {
	art.lock.Lock()
	_, deleted := art.tree.Delete(key)
	art.lock.Unlock()
	return deleted
}

// Close art树没有显示关闭的方法,只为了配合 IIndexer
func (art *AdaptiveRadixTree) Close() error {
	return nil
}

func (art *AdaptiveRadixTree) Iterator(reverse bool) biface.IIterator {
	art.lock.RLock()
	defer art.lock.RUnlock()
	return newARTIterator(art.tree, reverse)
}

func (art *AdaptiveRadixTree) Size() int {
	art.lock.RLock()
	size := art.tree.Size()
	art.lock.RUnlock()
	return size
}

// // Art 索引迭代器
type artIterator struct {
	currIndex int     // 当前遍历的下标位置
	reverse   bool    // 是否是反向遍历
	values    []*Item // key+位置索引信息
}

// newARTIterator creates a new artIterator object.
// It takes an art.Tree object and a boolean value indicating whether to iterate in reverse order.
// It returns a pointer to the artIterator object.
func newARTIterator(tree art.Tree, reverse bool) *artIterator {
	// Initialize the index variable
	var idx int
	if reverse {
		idx = tree.Size() - 1
	}

	// Create a slice to store the values of each node
	values := make([]*Item, tree.Size())

	// Define the saveValues function to save the values of each node
	saveValues := func(node art.Node) bool {
		// Create a new Item object with the key and position of the node
		item := &Item{
			key: node.Key(),
			pos: node.Value().(*data.LogRecordPos),
		}

		// Save the item in the values slice at the current index
		values[idx] = item

		// Update the index based on the reverse flag
		if reverse {
			idx--
		} else {
			idx++
		}

		// Return true to indicate that the node should be saved
		return true
	}

	// Iterate over each node in the tree and save its value
	tree.ForEach(saveValues)

	// Create and return the artIterator object
	return &artIterator{
		currIndex: 0,
		reverse:   reverse,
		values:    values,
	}
}

// Rewind resets the current index to 0.
func (ai *artIterator) Rewind() {
	ai.currIndex = 0
}

// Seek sets the current index to the position of the first value whose key is greater than or equal to the given key.
// If reverse is true, it sets the current index to the position of the first value whose key is less than or equal to the given key.
func (ai *artIterator) Seek(key []byte) {
	if ai.reverse {
		ai.currIndex = sort.Search(len(ai.values), func(i int) bool {
			return bytes.Compare(ai.values[i].key, key) <= 0
		})
	} else {
		ai.currIndex = sort.Search(len(ai.values), func(i int) bool {
			return bytes.Compare(ai.values[i].key, key) >= 0
		})
	}
}

// Next increments the current index by 1.
func (ai *artIterator) Next() {
	ai.currIndex += 1
}

// Valid returns true if the current index is within the bounds of the values slice.
func (ai *artIterator) Valid() bool {
	return ai.currIndex < len(ai.values)
}

// Key returns the key of the current value.
func (ai *artIterator) Key() []byte {
	return ai.values[ai.currIndex].key
}

// Value returns the position of the current value.
func (ai *artIterator) Value() *data.LogRecordPos {
	return ai.values[ai.currIndex].pos
}

// Close sets the values slice to nil.
func (ai *artIterator) Close() {
	ai.values = nil
}
