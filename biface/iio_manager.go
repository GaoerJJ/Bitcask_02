package biface

type IIOManager interface {
	Read([]byte, int64) (int64, error)

	Write([]byte) (int64, error)

	Sync() error

	Close() error

	Size() (int64, error)
}
