package socket

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/tracerun/tracerun/lg"
	"go.uber.org/zap"
)

const (
	headerBytes = 3
)

// RouteFunc to route handlers
type RouteFunc func([]byte, io.Writer)

// Server to define a socket server
type Server struct {
	port   uint16
	router map[uint8]RouteFunc
	ln     net.Listener
}

// NewServer to create a server instance
func NewServer(port uint16, router map[uint8]RouteFunc) *Server {
	return &Server{
		port:   port,
		router: router,
	}
}

// Start the server
func (s *Server) Start() error {
	if s.ln != nil {
		return fmt.Errorf("already started")
	}

	port := fmt.Sprintf(":%d", s.port)
	ln, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	s.ln = ln
	lg.L.Info("started to listen socket connections", zap.Uint16("port", s.port))

	for {
		conn, err := ln.Accept()
		if err != nil {
			lg.L.Error("error accept connection", zap.Error(err))
		}
		lg.L.Debug("new connection come")

		go s.handleConn(conn)
	}
}

// Stop the server
func (s *Server) Stop() error {
	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}

func (s *Server) handleConn(c net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			lg.L.Warn("recovered", zap.Any("error", r), zap.Stack("info"))
		}
	}()

	for {
		count, route, err := readHeader(c)
		if err != nil {
			lg.L.Error("error to read header", zap.Error(err))
		}
		lg.L.Debug("one", zap.Uint8("route", route), zap.Uint16("count", count))

		bytes, err := readData(c, count)
		if err != nil {
			lg.L.Error("error to read data", zap.Error(err))
		}
		lg.L.Debug("data", zap.Binary("data", bytes))

		fn, ok := s.router[route]
		if !ok {
			lg.L.Warn("not found")
		} else {
			fn(bytes, c)
		}
	}
}

// readHeader to read header containing data count and route info
func readHeader(c net.Conn) (uint16, uint8, error) {
	byteCount := uint16(0)
	route := uint8(0)

	buf := make([]byte, headerBytes)
	n, err := io.ReadFull(c, buf)
	n, err := c.Read(buf)
	if n != headerBytes {
		return byteCount, route, fmt.Errorf("read header wrong")
	}
	if err != nil {
		return byteCount, route, err
	}
	byteCount = binary.LittleEndian.Uint16(buf)
	route = uint8(buf[2])
	return byteCount, route, nil
}

// readData to read certain amount of bytes
func readData(c net.Conn, count uint16) ([]byte, error) {
	buf := make([]byte, count)
	n, err := io.ReadFull(c, buf)
	if n != int(count) {
		return buf, fmt.Errorf("read data length wrong")
	}
	return buf, err
}
