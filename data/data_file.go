package data

import (
	"Bitcask_02/biface"
	"Bitcask_02/fio"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"path/filepath"
)

var (
	ErrInvalidCRC = errors.New("invalid crc value, log record maybe corrupted")
)

// 文件类型
const (
	DataFileNameSuffix    = ".data"
	HintFileName          = "hint-index"
	MergeFinishedFileName = "merge-finished"
	SeqNoFileName         = "seq-no"
)

type DataFile struct {
	FileId      uint32
	WriteOffset int64
	IoManager   biface.IIOManager
}

// OpenDataFile 根据目录路径和 fileID 打开数据文件
// 此方法将根据目录路径和 fileID 在磁盘打开或创建对应的数据文件, 并生成 DataFile 返回
func OpenDataFile(dirPath string, fileId uint32, ioType fio.FileIOType) (*DataFile, error) {
	fileName := GetDataFileName(dirPath, fileId)
	return newDataFile(fileName, fileId, ioType)
}

// OpenHintFile 打开 Hint 索引文件
// 此方法将目录路径和 HintFileName 组合,在磁盘打开或创建对应的 Hint 文件
// Hint 文件 fileID 为 0 ,打开方式默认为标准文件 io
func OpenHintFile(dirPath string) (*DataFile, error) {
	fileName := filepath.Join(dirPath, HintFileName)
	return newDataFile(fileName, 0, fio.StandardFIO)
}

// OpenSeqNoFile 打开 Seq 事务文件
// 此方法将目录路径和 SeqNoFileName 组合,在磁盘打开或创建对应的 Seq 文件
// Seq 文件 fileID 为 0 , 打开方式默认为标准文件 io
func OpenSeqNoFile(dirPath string) (*DataFile, error) {
	fileName := filepath.Join(dirPath, SeqNoFileName)
	return newDataFile(fileName, 0, fio.StandardFIO)
}

// GetDataFileName 根据磁盘路径和文件 ID 组合成文件的磁盘名
func GetDataFileName(dirPath string, fileId uint32) string {
	return filepath.Join(dirPath, fmt.Sprintf("%09d", fileId)+DataFileNameSuffix)
}

// newDataFile 根据磁盘路径和 fileID 创建新的 DataFile 类型
// 磁盘路径的文件名和 fileID 是独立的关系
func newDataFile(fileName string, fileID uint32, ioType fio.FileIOType) (*DataFile, error) {
	// 初始化 IIOManager 接口
	ioManager, err := fio.NewIOManager(fileName, ioType)
	if err != nil {
		return nil, err
	}
	return &DataFile{
		FileId:      fileID,
		WriteOffset: 0,
		IoManager:   ioManager,
	}, nil
}

// ReadLogRecord 根据文件偏移量 offset 取出对应 LogRecord 数据
// 和磁盘数据记录的大小(包括 header key value 的总大小)
// 此方法将根据文件偏移 offset 从对应文件位置取出 header 头文件
// 再从 Header 头文件得到对应文件大小,从而取出 LogRecord 数据
func (df *DataFile) ReadLogRecord(offset int64) (*LogRecord, int64, error) {
	fileSize, err := df.IoManager.Size()
	if err != nil {
		return nil, 0, err
	}

	// 如果最大 header 长度已经超过文件剩余长度,则只需要读取剩余长度即可
	var headerBytes int64 = maxLogRecordHeaderSize
	if offset+maxLogRecordHeaderSize > fileSize {
		headerBytes = fileSize - offset
	}

	// 读取 Header 信息
	headerBuf, err := df.readNBytes(headerBytes, offset)
	if err != nil {
		return nil, 0, err
	}

	header, headerSize := decodeLogRecordHeader(headerBuf)
	// 下面两个条件表示读取到了文件末尾,直接返回 EOF 错误
	// todo 为什么能表示文件末尾 ,是因为数据存储时的某种规则吗
	if header == nil {
		return nil, 0, io.EOF
	}
	if header.crc == 0 && header.keySize == 0 && header.valueSize == 0 {
		return nil, 0, io.EOF
	}

	// 取出对应的 key 和 value 的长度
	keySize, valueSize := int64(header.keySize), int64(header.valueSize)

	logRecord := &LogRecord{Type: header.recordType}
	// 开始读取用户实际存储的 key/value 数据
	// todo 如果 keySize 和 valueSize 都是 0 表示什么意思 ?
	if keySize > 0 || valueSize > 0 {
		kvBuf, err := df.readNBytes(keySize+valueSize, offset+headerSize)
		if err != nil {
			return nil, 0, err
		}
		//	解出 key 和 value
		logRecord.Key = kvBuf[:keySize]
		logRecord.Value = kvBuf[keySize:]
	}

	crc := getCrc(headerBuf[crc32.Size:], logRecord.Key, logRecord.Value)
	if crc != header.crc {
		return nil, 0, ErrInvalidCRC
	}

	recordSize := headerSize + keySize + valueSize

	return logRecord, recordSize, nil
}

func (df *DataFile) Write(buf []byte) error {
	n, err := df.IoManager.Write(buf)
	if err != nil {
		return err
	}
	df.WriteOffset += n
	return nil
}

// WriteHintRecord 写入索引信息到 hint 文件中
func (df *DataFile) WriteHintRecord(key []byte, pos *LogRecordPos) error {
	record := &LogRecord{
		Key:   key,
		Value: EncodeLogRecordPos(pos),
	}
	encRecord, _ := EncodeLogRecord(record)
	return df.Write(encRecord)
}

func (df *DataFile) Sync() error {
	return df.IoManager.Sync()
}

func (df *DataFile) Close() error {
	return df.IoManager.Close()
}

// SetIOManager 根据 ioType 更改当前 datafile 的 ioManager
// 此方法会将当前 datafile 的 ioManager 关闭, 并开启一个新类型的 ioManager
func (df *DataFile) SetIOManager(dirPath string, ioType fio.FileIOType) error {
	if err := df.IoManager.Close(); err != nil {
		return err
	}
	ioManager, err := fio.NewIOManager(GetDataFileName(dirPath, df.FileId), ioType)
	if err != nil {
		return err
	}
	df.IoManager = ioManager
	return nil
}

// 从文件偏移处读取 n byte 的数据并返回
func (df *DataFile) readNBytes(n int64, offset int64) (b []byte, err error) {
	b = make([]byte, n)
	_, err = df.IoManager.Read(b, offset)
	return
}
