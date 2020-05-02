package readers

import (
	"encoding/binary"
	"io"
)

// ReadUint32 reads a uint32 from the reader
func ReadUint32(reader io.Reader) uint32 {
	buf := make([]byte, 4)
	reader.Read(buf)

	return binary.LittleEndian.Uint32(buf)
}

// ReadUint64 reads a uint64 from the reader
func ReadUint64(reader io.Reader) uint64 {
	buf := make([]byte, 8)
	reader.Read(buf)

	return binary.LittleEndian.Uint64(buf)
}
