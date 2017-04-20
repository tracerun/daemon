package socket

import (
	"net"
	"testing"
)

func TestSocket(t *testing.T) {
	port := ":8880"
	t.Logf("starting at %s", port)
	ln, err := net.Listen("tcp", port)
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConn(&conn)
	}
}

func handleConn(c *net.Conn) {

}
