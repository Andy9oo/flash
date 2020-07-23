package readers

import (
	"encoding/binary"
	"io"
)

// ReadUint32 reads a uint32 from the reader
func ReadUint32(reader io.Reader) uint32 {
	var val uint32
	binary.Read(reader, binary.LittleEndian, &val)
	return val
}

// ReadUint64 reads a uint64 from the reader
func ReadUint64(reader io.Reader) uint64 {
	var val uint64
	binary.Read(reader, binary.LittleEndian, &val)
	return val
}

// ReadFloat64 reads a float64 from the reader
func ReadFloat64(reader io.Reader) float64 {
	var val float64
	binary.Read(reader, binary.LittleEndian, &val)
	return val
}
