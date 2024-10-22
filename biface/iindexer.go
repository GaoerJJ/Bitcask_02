package biface

import "Bitcask_02/data"

// IIndexer 抽象内存索引接口,后续如果要接入其他的数据结构,实现这个接口即可
type IIndexer interface {
	// Put 向索引中存储 key 对应的数据位置信息
	Put(key []byte, pos *data.LogRecordPos) bool

	// Get 根据 key 取出对应的索引位置信息
	Get(key []byte) *data.LogRecordPos

	// Delete 根据 key 删除对应的索引位置信息
	Delete(key []byte) bool
	// Iterator 索引迭代器
	Iterator(reverse bool) IIterator
	//Size 索引中的数据量
	Size() int
	// Close 关闭内存索引
	Close() error
}
