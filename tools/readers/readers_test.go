package readers

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestUint32(t *testing.T) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint32(32))
	res := ReadUint32(&buf)
	if res != 32 {
		t.Error(res)
	}
}

func TestUint64(t *testing.T) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint64(64))
	res := ReadUint64(&buf)
	if res != 64 {
		t.Error(res)
	}
}

func TestFloat64(t *testing.T) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, float64(64.01))
	res := ReadFloat64(&buf)
	if res != 64.01 {
		t.Error(res)
	}
}
