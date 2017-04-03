package action

import (
	"encoding/binary"
	"errors"
	"time"
	"tracerun/db"
	"tracerun/db/segment"
	"tracerun/lg"

	"github.com/boltdb/bolt"
	"go.uber.org/zap"
)

const (
	actionBucket = "__actions__"

	bufferCount   = 200
	tickerSeconds = 60
)

type action struct {
	target string
	active bool
	ts     uint32
}

var (
	// ErrActionValue error for action value field
	ErrActionValue = errors.New("action value bytes wrong")

	// Expired an action after a given seconds
	Expired uint32 = 15

	// actionChan is used to handle actions
	actionChan = make(chan *action, bufferCount)
)

func init() {
	go checkActions()
	go receiveActions()
}

// GetAll to get all the actions.
func GetAll() ([]string, []uint32, []uint32, error) {
	// Readonly mode has problem on windows, so create RWDB, TODO
	readDB, err := db.CreateRWDB()
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() {
		if err := readDB.Close(); err != nil {
			lg.L.Error("fail to close db", zap.Error(err))
		}
	}()

	targets, starts, lasts, err := getAll(readDB)
	return targets, starts, lasts, err
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
		return
	}
	defer func() {
		if err := rwDB.Close(); err != nil {
			lg.L.Error("fail to close db", zap.Error(err))
		}
	}()

	ts := uint32(time.Now().Unix())
	lg.L.Debug("action directly to DB", zap.Any("target", target), zap.Bool("active", active), zap.Uint32("ts", ts))

	if err = handle(rwDB, target, active, ts); err != nil {
		lg.L.Error("error while handling action", zap.Error(err))
	}
}

func checkActions() {
	for tk := range time.Tick(tickerSeconds * time.Second) {
		if err := oneCheck(tk); err != nil {
			lg.L.Error("error while checking actions", zap.Error(err))
		}
	}
}

func oneCheck(tk time.Time) error {
	rwDB, err := db.CreateRWDB()
	if err != nil {
		return err
	}
	defer func() {
		if err := rwDB.Close(); err != nil {
			lg.L.Error("fail to close db", zap.Error(err))
		}
	}()

	now := uint32(tk.Unix())
	lg.L.Debug("one action check", zap.String("at", time.Unix(int64(now), 0).Format("2006-01-02 15:04:05")))

	var targets []string
	var starts, lasts []uint32
	err = rwDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(actionBucket))
		if b == nil {
			return nil
		}

		return b.ForEach(func(k []byte, v []byte) error {
			start, last, err := decodeValue(v)
			if err != nil {
				return err
			}
			if now-last > Expired {
				targets = append(targets, string(k))
				starts = append(starts, start)
				lasts = append(lasts, last)
			}
			return nil
		})
	})

	if err != nil {
		return err
	}

	l := len(targets)
	if l == 0 {
		return nil
	}

	return rwDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(actionBucket))
		if b == nil {
			return nil
		}

		for i := 0; i < l; i++ {
			if err := segment.Generate(tx, targets[i], starts[i], lasts[i]-starts[i]); err != nil {
				return err
			}
			// delete the old action
			if err := b.Delete([]byte(targets[i])); err != nil {
				return err
			}
			lg.L.Debug("expired action", zap.String("key", string(targets[i])))
		}
		return nil
	})
}

func receiveActions() {
	for {
		a := <-actionChan
		lg.L.Debug("action from Q", zap.Any("target", a.target), zap.Bool("active", a.active), zap.Uint32("ts", a.ts))

		// create a readwrite db.
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

		if err := rwDB.Close(); err != nil {
			lg.L.Error("fail to close db", zap.Error(err))
		} else {
			lg.L.Debug("RW DB connection closed.")
		}
	}
}

// handle an action.
func handle(rwDB *bolt.DB, target string, active bool, ts uint32) error {
	return rwDB.Update(func(tx *bolt.Tx) error {
		var err error

		b := tx.Bucket([]byte(actionBucket))
		// If actions bucket not existed, create one.
		if b == nil {
			b, err = tx.CreateBucket([]byte(actionBucket))
			if err != nil {
				return err
			}
			lg.L.Info("action bucket created")
		}

		start, last, err := get(b, target)
		if err != nil {
			return err
		}

		if !active {
			// a close action comes
			if start == 0 {
				// action not existed, add a 1 second segment
				lg.L.Debug("single close action", zap.String("target", target), zap.Uint32("ts", ts))
				return segment.Generate(tx, target, ts, uint32(1))
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
			err = segment.Generate(tx, target, start, long)
			if err != nil {
				return err
			}
			// delete the old action
			return b.Delete([]byte(target))
		}
		// an active action comes
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
			if err := segment.Generate(tx, target, start, long); err != nil {
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
	return b.Put([]byte(target), encodeValue(start, last))
}

func encodeValue(start, last uint32) []byte {
	// bytes value: first 4 bytes is uint32 for last active, last 4 bytes for uint32 start.
	bs := make([]byte, 8)
	// encode to bs with start timestamp
	binary.LittleEndian.PutUint32(bs[4:], start)
	// encode to bs with last active timestamp
	binary.LittleEndian.PutUint32(bs[0:], last)
	return bs
}

func decodeValue(bs []byte) (uint32, uint32, error) {
	if len(bs) != 8 {
		return 0, 0, ErrActionValue
	}
	return binary.LittleEndian.Uint32(bs[4:]), binary.LittleEndian.Uint32(bs), nil
}

// get an action.
// start and last active timestamp will be returned
func get(b *bolt.Bucket, target string) (uint32, uint32, error) {
	bs := b.Get([]byte(target))
	if bs == nil {
		return 0, 0, nil
	}
	return decodeValue(bs)
}

// getAll to get all the actions.
func getAll(readDB *bolt.DB) ([]string, []uint32, []uint32, error) {
	var targets []string
	var starts, lasts []uint32

	err := readDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(actionBucket))
		if b == nil {
			return nil
		}

		return b.ForEach(func(k []byte, v []byte) error {
			start, last, err := decodeValue(v)
			if err != nil {
				return err
			}
			targets = append(targets, string(k))
			starts = append(starts, start)
			lasts = append(lasts, last)
			return nil
		})
	})
	return targets, starts, lasts, err
}
