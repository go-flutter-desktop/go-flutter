package plugin

import (
	"encoding/binary"
	"unsafe"
)

var endian binary.ByteOrder

func init() {
	// find out endiannes of the host
	const intSize int = int(unsafe.Sizeof(0))
	var i = 0x1
	bs := (*[intSize]byte)(unsafe.Pointer(&i))
	if bs[0] == 0 {
		endian = binary.BigEndian
	} else {
		endian = binary.LittleEndian
	}
}
