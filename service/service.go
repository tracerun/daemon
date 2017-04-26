package service

import (
	"encoding/binary"
	"errors"
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
	// ErrDataLength the data length wrong
	ErrDataLength = errors.New("read data length wrong")

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
func WriteErrorMessage(err error, w io.Writer) {
	var errMsg ErrorMessage
	errMsg.Message = err.Error()

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

// ReadOne to read header containing data count and route info
func ReadOne(r io.Reader) ([]byte, uint8, error) {
	byteCount := uint16(0)
	route := uint8(0)

	headerBuf := make([]byte, headerBytes)
	n, err := io.ReadFull(r, headerBuf)
	if err != nil {
		return nil, route, err
	}
	if n != headerBytes {
		return nil, route, fmt.Errorf("read header wrong")
	}

	byteCount = binary.LittleEndian.Uint16(headerBuf)
	route = uint8(headerBuf[2])

	buf := make([]byte, byteCount)
	n, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, route, err
	}
	if n != int(byteCount) {
		return nil, route, ErrDataLength
	}

	return buf, route, nil
}
