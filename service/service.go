package service

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/golang/protobuf/proto"
	"github.com/tracerun/tdb"
	"github.com/tracerun/tracerun/lg"
)

const (
	headerBytes = 3
)

var (
	db       *tdb.TDB
	stopChan = make(chan bool, 1)
)

// RouteFunc to route handlers
type RouteFunc func([]byte, io.Writer)

// Start service
func Start(port uint16, dbFolder string) {
	go receiveActions()
	go checkActions()

	var err error
	db, err = tdb.Open(dbFolder)
	if err != nil {
		panic(err)
	}

	s := NewTCPServer(port, getRouter())
	go s.Start()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	select {
	case <-stopChan:
	case <-sigs:
	}

	stop(s)
}

func stop(s *TCPServer) {
	s.ln.Close()
	lg.L.Info("TCP service stopped.")
}

// WriteErrorMessage to write a message to writer
func WriteErrorMessage(msg string, w io.Writer) {
	var errMsg ErrorMessage
	errMsg.Message = msg

	buf, _ := proto.Marshal(&errMsg)
	headerBuf := GenerateHeaderBuf(uint16(len(buf)), uint8(255))

	w.Write(append(headerBuf, buf...))
}

// GenerateHeaderBuf to generate a header buf
func GenerateHeaderBuf(length uint16, route uint8) []byte {
	buf := make([]byte, 3)
	binary.LittleEndian.PutUint16(buf, length)
	buf[2] = byte(route)
	return buf
}

// ReadHeader to read header containing data count and route info
func ReadHeader(r io.Reader) (uint16, uint8, error) {
	byteCount := uint16(0)
	route := uint8(0)

	buf := make([]byte, headerBytes)
	n, err := io.ReadFull(r, buf)
	if err != nil {
		return byteCount, route, err
	}
	if n != headerBytes {
		return byteCount, route, fmt.Errorf("read header wrong")
	}

	byteCount = binary.LittleEndian.Uint16(buf)
	route = uint8(buf[2])
	return byteCount, route, nil
}

// ReadData to read certain amount of bytes
func ReadData(r io.Reader, count uint16) ([]byte, error) {
	buf := make([]byte, count)

	n, err := io.ReadFull(r, buf)
	if err != nil {
		return buf, err
	}
	if n != int(count) {
		return buf, fmt.Errorf("read data length wrong")
	}
	return buf, err
}
