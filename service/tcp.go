package service

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/tracerun/tracerun/lg"
	"go.uber.org/zap"
)

const (
	readTimeout = 30
)

// TCPServer to define a TCP server
type TCPServer struct {
	port   uint16
	router map[uint8]RouteFunc
	ln     net.Listener
}

// NewTCPServer to create a TCP server instance
func NewTCPServer(port uint16, router map[uint8]RouteFunc) *TCPServer {
	return &TCPServer{
		port:   port,
		router: router,
	}
}

// Start the server
func (s *TCPServer) Start() error {
	if s.ln != nil {
		return fmt.Errorf("already started")
	}

	port := fmt.Sprintf(":%d", s.port)
	// net.ListenUDP("udp", laddr*net.UDPAddr)
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
func (s *TCPServer) Stop() error {
	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}

func (s *TCPServer) handleConn(c net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			lg.L.Warn("recovered", zap.Any("error", r), zap.Stack("info"))
		}
	}()

	for {
		// read header
		c.SetReadDeadline(time.Now().Add(readTimeout * time.Second))
		count, route, err := ReadHeader(c)

		if err != nil {
			recordConnError(err)
			break
		}
		lg.L.Debug("header", zap.Uint8("route", route), zap.Uint16("count", count))

		// read data
		var bytes []byte
		if count > 0 {
			c.SetReadDeadline(time.Now().Add(readTimeout * time.Second))
			bytes, err = ReadData(c, count)

			if err != nil {
				recordConnError(err)
				break
			}
			lg.L.Debug("data", zap.Binary("data", bytes))
		}

		// get routed function
		fn, ok := s.router[route]
		if !ok {
			lg.L.Warn("not found")
		} else {
			fn(bytes, c)
		}
	}

	if err := c.Close(); err != nil {
		lg.L.Error("error close", zap.Error(err))
	}
	lg.L.Debug("connection closed")
}

func recordConnError(err error) {
	if err == io.EOF {
		lg.L.Debug("EOF")
		return
	}
	if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
		lg.L.Debug("timeout")
		return
	}
	if _, ok := err.(*net.OpError); ok {
		lg.L.Debug("operror")
		return
	}

	lg.L.Error("error to read data", zap.Error(err))
}
