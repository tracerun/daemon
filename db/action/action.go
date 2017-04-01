package action

import (
	"encoding/binary"
	"time"
	"tracerun/db"
	"tracerun/db/segment"
	"tracerun/lg"

	"github.com/boltdb/bolt"
	"go.uber.org/zap"
)

const (
	actionBucket = "__actions__"

	bufferCount = 100
)

type action struct {
	target string
	active bool
	ts     uint32
}

var (
	// Expired an action after a given seconds
	Expired uint32 = 15

	// actionChan is used to handle actions
	actionChan = make(chan *action, bufferCount)
)

func init() {
	go receiveAction()
}

// AddToQ to enqueue an action
func AddToQ(target string, active bool) {
	var a action
	a.target = target
	a.active = active
	a.ts = uint32(time.Now().Unix())

	// enqueue
	go func() { actionChan <- &a }()
}

// AddToDB to directly add to database.
func AddToDB(target string, active bool) {
	rwDB, err := db.CreateRWDB()
	if err != nil {
		lg.L.Error("error create rw DB", zap.Error(err))
	}
	defer rwDB.Close()

	ts := uint32(time.Now().Unix())
	lg.L.Debug("action directly to DB", zap.Any("target", target), zap.Bool("active", active), zap.Uint32("ts", ts))

	if err = handle(rwDB, target, active, ts); err != nil {
		lg.L.Error("error while handling action", zap.Error(err))
	}
}

func receiveAction() {
	for {
		a := <-actionChan
		lg.L.Debug("action from Q", zap.Any("target", a.target), zap.Bool("active", a.active), zap.Uint32("ts", a.ts))

		rwDB, err := db.CreateRWDB()
		if err != nil {
			lg.L.Error("error create rw DB", zap.Error(err))
		}
		lg.L.Debug("RW DB connection created.")

		if err = handle(rwDB, a.target, a.active, a.ts); err != nil {
			lg.L.Error("error while handling action", zap.Error(err))
		}
	Remaining:
		for i := 0; i < bufferCount-1; i++ {
			select {
			case a := <-actionChan:
				lg.L.Debug("action from Q", zap.Any("target", a.target), zap.Bool("active", a.active), zap.Uint32("ts", a.ts))
				if err = handle(rwDB, a.target, a.active, a.ts); err != nil {
					lg.L.Error("error while handling action", zap.Error(err))
				}
			default:
				break Remaining
			}
		}
		rwDB.Close()
		lg.L.Debug("RW DB connection closed.")
	}
}

// handle an action.
func handle(rwDB *bolt.DB, target string, active bool, ts uint32) error {
	return rwDB.Update(func(tx *bolt.Tx) error {
		var err error

		b := tx.Bucket([]byte(actionBucket))
		if b == nil {
			b, err = tx.CreateBucket([]byte(actionBucket))
			if err != nil {
				return err
			}
			lg.L.Info("action bucket created")
		}

		start, last := get(b, target)
		if !active {
			// a close action comes
			if start == 0 {
				// action not existed, add a 1 second segment
				lg.L.Debug("single close action", zap.String("target", target), zap.Uint32("ts", ts))
				return segment.Generate(tx, target, ts, uint16(1))
			}

			// calculate how long the segment is
			var long uint32
			if ts < last {
				// timestamp is earlier than last active
				long = last - start
			} else if ts-last > Expired {
				// timestamp is expired, just add 1 second
				long = last - start + 1
			} else {
				// timestamp - start
				long = ts - start
			}
			return segment.Generate(tx, target, start, uint16(long))
		}
		// a active action comes
		if start == 0 {
			// action not existed, create the action
			return put(b, target, ts, ts+1)
		}
		if ts <= last {
			// timestamp is no later than last active, do nothing
			lg.L.Debug("earlier action", zap.String("target", target), zap.Uint32("ts", ts))
			return nil
		} else if ts-last > Expired {
			// timestamp is expired, create a segment and a new action
			long := last - start
			if err := segment.Generate(tx, target, start, uint16(long)); err != nil {
				return err
			}

			// set start and ts for new action
			start = ts
			ts = start + 1
		}
		return put(b, target, start, ts)
	})
}

// put an action in bucket, with timestamp of start and last active.
func put(b *bolt.Bucket, target string, start, last uint32) error {
	lg.L.Debug("put new action", zap.String("target", target), zap.Uint32("start", start), zap.Uint32("last", last))

	// bytes value: first 4 bytes is uint32 for last active, last 4 bytes for uint32 start.
	bs := make([]byte, 8)
	// encode to bs with start timestamp
	binary.LittleEndian.PutUint32(bs[4:], start)
	// encode to bs with last active timestamp
	binary.LittleEndian.PutUint32(bs[0:], last)
	return b.Put([]byte(target), bs)
}

// get an action.
// start and last active timestamp will be returned
func get(b *bolt.Bucket, target string) (uint32, uint32) {
	var start, last uint32
	bs := b.Get([]byte(target))
	if bs != nil {
		last = binary.LittleEndian.Uint32(bs)
		start = binary.LittleEndian.Uint32(bs[4:])
	}
	return start, last
}
