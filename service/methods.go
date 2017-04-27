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
	ts     uint32
}

func receiveActions() {
	for {
		a := <-actionChan
		addOneAction(a)
	Remaining:
		for i := 0; i < bufferCount-1; i++ {
			select {
			case a := <-actionChan:
				addOneAction(a)
			default:
				break Remaining
			}
		}
	}
}

func addOneAction(a *act) {
	lg.L.Debug("action from Q", zap.Any("target", a.target), zap.Uint32("ts", a.ts))
	if err := db.AddAction(a.target, a.ts); err != nil {
		lg.L.Error("error add action", zap.Error(err))
	}
}

func checkActions() {
	for _ = range time.Tick(tickerSeconds * time.Second) {
		if err := db.CheckExpirations(); err != nil {
			lg.L.Error("error while checking actions", zap.Error(err))
		}
	}
}

// exit uint8(0) to stop the server
func exit(b []byte, w io.Writer) {
	stopChan <- true
}

// ping uint8(1) used to extend readtimeout
func ping(b []byte, w io.Writer) {}

// getMeta uint8(2) to get meta information
func getMeta(b []byte, w io.Writer) {
	thisRoute := uint8(2)

	var meta Meta

	v, err := db.Version()
	if err != nil {
		WriteErrorMessage(err, w)
		return
	}
	meta.Version = uint32(v)

	tag, err := db.Tag()
	if err != nil {
		WriteErrorMessage(err, w)
		return
	}
	meta.Tag = tag

	cre, err := db.CreateAt()
	if err != nil {
		WriteErrorMessage(err, w)
		return
	}
	meta.CreateAt = cre

	host, err := db.Host()
	if err != nil {
		WriteErrorMessage(err, w)
		return
	}
	meta.Host = host

	username, err := db.Username()
	if err != nil {
		WriteErrorMessage(err, w)
		return
	}
	meta.Username = username

	arch, err := db.Arch()
	if err != nil {
		WriteErrorMessage(err, w)
		return
	}
	meta.Arch = arch

	osInfo, err := db.OS()
	if err != nil {
		WriteErrorMessage(err, w)
		return
	}
	meta.Os = osInfo

	offset, err := db.ZoneOffset()
	if err != nil {
		WriteErrorMessage(err, w)
		return
	}
	meta.ZoneOffset = offset

	buf, err := proto.Marshal(&meta)
	if err != nil {
		WriteErrorMessage(err, w)
		return
	}

	headerBuf := GenerateHeaderBuf(uint16(len(buf)), thisRoute)
	if _, err := w.Write(append(headerBuf, buf...)); err != nil {
		lg.L.Error("error writing", zap.Error(err))
	}
}

// action uint8(10) to receive action income.
func action(b []byte, w io.Writer) {
	// enqueue
	go func() {
		actionChan <- &act{
			target: string(b),
			ts:     uint32(time.Now().Unix()),
		}
	}()
}

// getActions uint8(11) to get all actions
func getActions(b []byte, w io.Writer) {
	thisRoute := uint8(11)

	var all AllActions
	targets, starts, lasts, err := db.GetActions()
	if err != nil {
		lg.L.Error("error getting actions", zap.Error(err))
		WriteErrorMessage(err, w)
		return
	}

	for i := 0; i < len(targets); i++ {
		all.Actions = append(all.Actions, &AllActions_Act{
			Target: targets[i],
			Start:  starts[i],
			Last:   lasts[i],
		})
	}

	buf, err := proto.Marshal(&all)
	if err != nil {
		WriteErrorMessage(err, w)
		return
	}

	headerBuf := GenerateHeaderBuf(uint16(len(buf)), thisRoute)
	if _, err := w.Write(append(headerBuf, buf...)); err != nil {
		lg.L.Error("error writing", zap.Error(err))
	}
}

// getTargets uint8(20) to get all targets
func getTargets(b []byte, w io.Writer) {
	thisRoute := uint8(20)

	targets := db.GetTargets()

	var all Targets
	all.Target = targets

	buf, err := proto.Marshal(&all)
	if err != nil {
		WriteErrorMessage(err, w)
		return
	}

	headerBuf := GenerateHeaderBuf(uint16(len(buf)), thisRoute)
	if _, err := w.Write(append(headerBuf, buf...)); err != nil {
		lg.L.Error("error writing", zap.Error(err))
	}
}

// getSlots uint8(21) to get slots of a target in a range
func getSlots(b []byte, w io.Writer) {
	thisRoute := uint8(21)

	var rang SlotRange
	if err := proto.Unmarshal(b, &rang); err != nil {
		WriteErrorMessage(err, w)
		return
	}

	startsResult, slotsResult, err := db.GetSlots(rang.Target, rang.Start, rang.End)
	if err != nil {
		WriteErrorMessage(err, w)
		return
	}
	var slots []*Slot
	for i := 0; i < len(startsResult); i++ {
		oneStarts, oneSlots := startsResult[i], slotsResult[i]
		for j := 0; j < len(oneStarts); j++ {
			slots = append(slots, &Slot{
				Start: oneStarts[j],
				Slot:  oneSlots[j],
			})
		}
	}

	var all Slots
	all.Slots = slots

	buf, err := proto.Marshal(&all)
	if err != nil {
		WriteErrorMessage(err, w)
		return
	}

	headerBuf := GenerateHeaderBuf(uint16(len(buf)), thisRoute)
	if _, err := w.Write(append(headerBuf, buf...)); err != nil {
		lg.L.Error("error writing", zap.Error(err))
	}
}

func getRouter() map[uint8]RouteFunc {
	m := make(map[uint8]RouteFunc)

	m[uint8(0)] = exit
	m[uint8(1)] = ping
	m[uint8(2)] = getMeta
	m[uint8(10)] = action
	m[uint8(11)] = getActions
	m[uint8(20)] = getTargets
	m[uint8(21)] = getSlots

	return m
}
