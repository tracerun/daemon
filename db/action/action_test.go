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
)

func TestActionMethods(t *testing.T) {
	lg.InitLogger(true, false, "")

	dbPath := "action_test.db"
	// set db path
	db.SetDBPath(dbPath)
	defer func() {
		assert.NoError(t, os.Remove(dbPath), "error while removing file")
	}()

	testEnqueue(t)
	testActionEncoding(t)
	testExpireAction(t)
	testCloseAction(t)
}

func testActionEncoding(t *testing.T) {
	rwDB, err := db.CreateRWDB()
	assert.NoError(t, err, "error while open read-write db")
	defer rwDB.Close()

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
		start, last := get(b, target)
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
		start, last = get(b, target)
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
		Add(randomTarget.String(), true)
	}

	time.Sleep(1 * time.Second)
}

func testExpireAction(t *testing.T) {
	// create a random target.
	randomTarget, err := ulid.NewFromTime(time.Now())
	assert.NoError(t, err, "error while creating random target.")

	Add(randomTarget.String(), true)

	time.Sleep(expired * time.Second)
	time.Sleep(1500 * time.Millisecond)
	Add(randomTarget.String(), true)
	time.Sleep(500 * time.Millisecond)
}

func testCloseAction(t *testing.T) {
	// create a random target.
	randomTarget, err := ulid.NewFromTime(time.Now())
	assert.NoError(t, err, "error while creating random target.")

	Add(randomTarget.String(), true)
	time.Sleep(1 * time.Second)
	Add(randomTarget.String(), false)
	time.Sleep(500 * time.Millisecond)
}
