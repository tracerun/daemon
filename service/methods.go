package service

import (
	"io"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/tracerun/tracerun/lg"
	"go.uber.org/zap"
)

const (
	bufferCount   = 200
	tickerSeconds = 60
)

var (
	// actionChan is used to handle actions
	actionChan = make(chan *act, bufferCount)
)

type act struct {
	target string
	active bool
	ts     uint32
}

func receiveActions() {
	for {
		a := <-actionChan
		lg.L.Debug("action from Q", zap.Any("target", a.target), zap.Bool("active", a.active), zap.Uint32("ts", a.ts))

	Remaining:
		for i := 0; i < bufferCount-1; i++ {
			select {
			case a := <-actionChan:
				lg.L.Debug("action from Q", zap.Any("target", a.target), zap.Bool("active", a.active), zap.Uint32("ts", a.ts))

			default:
				break Remaining
			}
		}

		// TODO write file
	}
}

func checkActions() {
	// for tk := range time.Tick(tickerSeconds * time.Second) {
	// 	if err := oneCheck(tk); err != nil {
	// 		lg.L.Error("error while checking actions", zap.Error(err))
	// 	}
	// }
}

// ping uint8(0) used to extend readtimeout
func ping(b []byte, w io.Writer) {}

// action to receive action income.
func action(b []byte, w io.Writer) {
	var ac Action
	if err := proto.Unmarshal(b, &ac); err != nil {
		lg.L.Error("error parse action", zap.Error(err))
	}

	var a act
	a.target = ac.Target
	a.active = ac.Active
	a.ts = uint32(time.Now().Unix())

	// enqueue
	go func() { actionChan <- &a }()
}

func getRouter() map[uint8]RouteFunc {
	m := make(map[uint8]RouteFunc)

	m[uint8(0)] = ping
	m[uint8(1)] = action

	return m
}
