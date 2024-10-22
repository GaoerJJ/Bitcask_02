package fio

import "os"

// FileIO 标准系统文件 IO
type FileIO struct {
	fd *os.File // 系统文件描述符
}

// NewFileIOManager 初始化标准文件 IO
func NewFileIOManager(fileName string) (*FileIO, error) {
	fd, err := os.OpenFile(
		fileName,
		// 不存在创建 读写模式打开文件 追加模式打开文件
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		DataFilePerm,
	)
	if err != nil {
		return nil, err
	}
	return &FileIO{fd: fd}, nil
}

func (fio *FileIO) Read(b []byte, offset int64) (int64, error) {
	size, err := fio.fd.ReadAt(b, offset)
	return int64(size), err
}

func (fio *FileIO) Write(b []byte) (int64, error) {
	size, err := fio.fd.Write(b)
	return int64(size), err
}

func (fio *FileIO) Sync() error {
	return fio.fd.Sync()
}

func (fio *FileIO) Close() error {
	return fio.fd.Close()
}
func (fio *FileIO) Size() (int64, error) {
	stat, err := fio.fd.Stat()
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}
