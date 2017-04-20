package segment

import (
	"os"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/drkaka/ulid"
	"github.com/stretchr/testify/assert"
	"github.com/tracerun/tracerun/db"
	"github.com/tracerun/tracerun/lg"
	"go.uber.org/zap"
)

func TestSegmentMethods(t *testing.T) {
	lg.InitLogger(true, false, "")

	dbPath := "segment_test.db"
	// set db path
	db.SetDBPath(dbPath)
	defer func() {
		assert.NoError(t, os.Remove(dbPath), "error while removing file")
	}()

	testGeneration(t)
	testPutGet(t)
}

func testGeneration(t *testing.T) {
	// create a random target.
	rwDB, err := db.CreateRWDB()
	assert.NoError(t, err, "error while open read-write db")
	defer func() {
		if err := rwDB.Close(); err != nil {
			lg.L.Error("fail to close db", zap.Error(err))
		}
	}()

	err = rwDB.Update(func(tx *bolt.Tx) error {
		now := time.Now()
		// create a random target.
		randomTarget, err := ulid.NewFromTime(now)
		if err != nil {
			return err
		}
		target := randomTarget.String()

		start := uint32(now.Unix())
		long := uint32(20)

		return Generate(tx, target, start, long)
	})
	assert.NoError(t, err, "error while testing.")
}

func testPutGet(t *testing.T) {
	rwDB, err := db.CreateRWDB()
	assert.NoError(t, err, "error while open read-write db")
	defer func() {
		if err := rwDB.Close(); err != nil {
			lg.L.Error("fail to close db", zap.Error(err))
		}
	}()

	err = rwDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(segBucket))

		now := time.Now()
		// create a random target.
		randomTarget, err := ulid.NewFromTime(now)
		if err != nil {
			return err
		}
		target := randomTarget.String()

		var targetB *bolt.Bucket
		if targetB, err = b.CreateBucket([]byte(target)); err != nil {
			return err
		}

		start := uint32(now.Unix())
		seg := uint32(20)

		if err = put(targetB, start, seg); err != nil {
			return err
		}

		// target should not be existed
		long := get(targetB, start)

		assert.Equal(t, seg, long, "result should be equal")
		return nil
	})
	assert.NoError(t, err, "error while testing.")
}
