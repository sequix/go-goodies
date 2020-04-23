package unit

const (
	Bit  = 1
	Byte = 8 * Bit
)

type IECInfoSize uint64

const (
	KiB IECInfoSize = 1024 * Byte
	MiB             = 1024 * KiB
	GiB             = 1024 * MiB
	TiB             = 1024 * GiB
	PiB             = 1024 * TiB
	EiB             = 1024 * PiB
	ZiB             = 1024 * EiB // overflow uint64
	YiB             = 1024 * ZiB // overflow uint64
)

type SIInfoSize uint64

const (
	KB SIInfoSize = 1000 * Byte
	MB            = 1000 * KB
	GB            = 1000 * MB
	TB            = 1000 * GB
	PB            = 1000 * TB
	EB            = 1000 * PB
	ZB            = 1000 * EB // overflow uint64
	YB            = 1000 * ZB // overflow uint64
)

// todo stringer and parser