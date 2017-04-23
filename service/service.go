package service

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/tracerun/tracerun/lg"
	"go.uber.org/zap"
)

const (
	headerBytes = 3
)

// RouteFunc to route handlers
type RouteFunc func([]byte, io.Writer)

// Start service
func Start(port uint16) {
	go receiveActions()
	go checkActions()

	s := NewTCPServer(port, getRouter())

	if err := s.Start(); err != nil {
		lg.L.Error("error start service", zap.Error(err))
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	<-sigs

	stop(s)
}

func stop(s *TCPServer) {
	s.ln.Close()
	lg.L.Info("TCP service stopped.")
}

// readHeader to read header containing data count and route info
func readHeader(r io.Reader) (uint16, uint8, error) {
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

// readData to read certain amount of bytes
func readData(r io.Reader, count uint16) ([]byte, error) {
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
