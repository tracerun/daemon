package socket

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	headerBytes = 2
	routerBits  = 5
)

func readHeader(c net.Conn) (uint16, uint8, error) {
	byteCount := uint16(0)
	route := uint8(0)

	buf := make([]byte, headerBytes)
	n, err := c.Read(buf)
	if n != headerBytes {
		return byteCount, route, fmt.Errorf("read header wrong")
	}
	if err != nil {
		return byteCount, route, err
	}
	header := binary.LittleEndian.Uint16(buf)
	route = uint8(header & 0x1f)
	byteCount = uint16(header >> routerBits)
	return byteCount, route, nil
}
