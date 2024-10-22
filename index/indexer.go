package index

import (
	"Bitcask_02/biface"
	"Bitcask_02/data"
	"bytes"
	"fmt"
	"github.com/google/btree"
)

// IndexerType 索引类型
type IndexerType = int8

const (
	Btree IndexerType = iota + 1
	ART
	BPTree
)

// NewIndexer 初试化索引
func NewIndexer(tp IndexerType, dir string, sync bool) (biface.IIndexer, error) {
	switch tp {
	case Btree:
		return NewBTree(), nil
	case ART:
		//todo
		return NewART(), nil
	//case BPTree:
	//	return NewBPlusTree(dir, sync), nil
	default:
		return nil, fmt.Errorf("unsupported index type: %v", tp)
	}
}

// Item 索引节点
type Item struct {
	key []byte
	pos *data.LogRecordPos
}

// Less 比较函数
func (ai *Item) Less(bi btree.Item) bool {
	return bytes.Compare(ai.key, bi.(*Item).key) == -1
}
