package socket

import (
	"testing"

	"github.com/tracerun/tracerun/lg"
)

func TestSocket(t *testing.T) {
	lg.InitLogger(true, false, "")

	port := uint16(8880)
	s := NewServer(port, nil)

	s.Start()
}
