package db

import (
	"Bitcask_02/biface"
	"sync"
)

type DB struct {
	cfg     DBConfig      // 配置项
	mu      *sync.RWMutex // 互斥锁
	fileIDs []int         // 文件id只能用于加载文件索引时使用，不能在其他地方使用
	//activeFile     *data.DataFile            // 活跃文件 用于写入
	//oldFile        map[uint32]*data.DataFile // 旧数据文件，只用于读出
	index          biface.IIndexer // 内存索引
	seqNo          uint64          // 序列号
	isMerging      bool            //是否正在merge
	seqNoFileExist bool            // 存储事务序列号的文件是否存在
	isInitiated    bool            // 是否是第一次初始化数据目录
}

// Stat 存储引擎统计信息
type Stat struct {
	KeyNum          uint  // key 的总数量
	DataFileNum     uint  // 数据文件的数量
	ReclaimableSize int64 // 可以进行 merge 回收的数据量，字节为单位
	DiskSize        int64 // 数据目录所占磁盘空间大小
}

func Open(cfg DBConfig) (*DB, error) {
	return nil, nil
}
