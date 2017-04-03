package action

import (
	"os"
	"testing"
	"time"
	"tracerun/db"
	"tracerun/lg"

	"github.com/boltdb/bolt"
	"github.com/drkaka/ulid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestActionMethods(t *testing.T) {
	lg.InitLogger(true, false, "")

	dbPath := "action_test.db"
	db.SetDBPath(dbPath)
	defer func() {
		assert.NoError(t, os.Remove(dbPath), "error while removing file")
	}()

	testEnqueue(t)
	testActionEncoding(t)
	testExpireAction(t)
	testCloseAction(t)
	testGetAllActions(t)
	waitTicker(t)
}

func testActionEncoding(t *testing.T) {
	rwDB, err := db.CreateRWDB()
	assert.NoError(t, err, "error while open read-write db")
	defer func() {
		if err := rwDB.Close(); err != nil {
			lg.L.Error("fail to close db", zap.Error(err))
		}
	}()

	err = rwDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(actionBucket))

		now := time.Now()
		// create a random target.
		randomTarget, err := ulid.NewFromTime(now)
		if err != nil {
			return err
		}
		target := randomTarget.String()

		// target should not be existed
		start, last, err := get(b, target)
		assert.NoError(t, err, "error while getting an action.")
		assert.Equal(t, uint32(0), start, "start should be 0")
		assert.Equal(t, uint32(0), last, "last should be 0")

		// create the target
		ts := uint32(now.Unix())
		expectLast := ts + 3
		err = put(b, randomTarget.String(), ts, expectLast)
		if err != nil {
			return err
		}

		// get back the target
		start, last, err = get(b, target)
		assert.NoError(t, err, "error while getting an action.")
		assert.Equal(t, ts, start, "start should be equal to ts")
		assert.Equal(t, expectLast, last, "last should be equal to expectLast")

		return nil
	})
	assert.NoError(t, err, "error while testing.")
}

func testEnqueue(t *testing.T) {
	// create a random target.
	randomTarget, err := ulid.NewFromTime(time.Now())
	assert.NoError(t, err, "error while creating random target.")
	for i := 0; i < 5; i++ {
		lg.L.Debug("add an action")
		AddToQ(randomTarget.String(), true)
	}

	time.Sleep(1 * time.Second)
}

func testExpireAction(t *testing.T) {
	// create a random target.
	randomTarget, err := ulid.NewFromTime(time.Now())
	assert.NoError(t, err, "error while creating random target.")

	AddToDB(randomTarget.String(), true)

	time.Sleep(time.Duration(Expired) * time.Second)
	time.Sleep(1100 * time.Millisecond)
	AddToDB(randomTarget.String(), true)
}

func testCloseAction(t *testing.T) {
	// create a random target.
	randomTarget, err := ulid.NewFromTime(time.Now())
	assert.NoError(t, err, "error while creating random target.")

	target := randomTarget.String()
	AddToDB(target, true)
	AddToDB(target, false)

	// Readonly mode has problem on windows, so create RWDB, TODO
	readDB, err := db.CreateRWDB()
	assert.NoError(t, err, "error while open read-only db")
	defer func() {
		if err := readDB.Close(); err != nil {
			lg.L.Error("fail to close db", zap.Error(err))
		}
	}()

	// after close, the action should be deleted.
	err = readDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(actionBucket))

		start, last, err := get(b, target)
		assert.NoError(t, err, "error while getting an action.")
		assert.Equal(t, uint32(0), start, "start should be 0")
		assert.Equal(t, uint32(0), last, "last should be 0")

		return nil
	})
	assert.NoError(t, err, "error while testing.")
}

func testGetAllActions(t *testing.T) {
	targets, starts, lasts, err := GetAll()
	assert.NoError(t, err, "error while getting all actions.")
	assert.Len(t, targets, 3, "should have 3 targets")
	assert.Len(t, starts, 3, "should have 3 targets")
	assert.Len(t, lasts, 3, "should have 3 targets")
}

func waitTicker(t *testing.T) {
	time.Sleep(time.Duration(50) * time.Second)
	targets, starts, lasts, err := GetAll()
	assert.NoError(t, err, "error while getting all actions.")
	assert.Len(t, targets, 0, "should have 0 targets")
	assert.Len(t, starts, 0, "should have 0 targets")
	assert.Len(t, lasts, 0, "should have 0 targets")
}
