package segment

import (
	"encoding/binary"
	"tracerun/lg"

	"github.com/boltdb/bolt"
	"go.uber.org/zap"
)

const (
	segBucket = "__segments__"
)

// Generate to generate a segment for a given target.
func Generate(tx *bolt.Tx, target string, start uint32, seg uint16) error {
	// get segment bucket
	var err error
	b := tx.Bucket([]byte(segBucket))
	if b == nil {
		if b, err = tx.CreateBucket([]byte(segBucket)); err != nil {
			return err
		}
		lg.L.Info("segments bucket created")
	}

	// get target segment bucket
	targetB := b.Bucket([]byte(target))
	if targetB == nil {
		if targetB, err = b.CreateBucket([]byte(target)); err != nil {
			return err
		}
		lg.L.Info("target segment bucket created", zap.String("target", target))
	}

	long := get(targetB, start)
	if long == 0 {
		// segment for a start time not existed
		err = put(targetB, start, seg)
		lg.L.Debug("new segment", zap.String("target", target), zap.Uint32("start", start), zap.Uint16("seg", seg))
	} else if seg > long {
		// segment for a start time existed, if new is longer that old, put new.
		err = put(targetB, start, seg)
		lg.L.Debug("update segment", zap.String("target", target), zap.Uint32("start", start), zap.Uint16("seg", seg))
	} else {
		lg.L.Debug("segment not change", zap.String("target", target), zap.Uint32("start", start))
	}
	return err
}

func getUInt32Bytes(ts uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, ts)
	return b
}

func getUInt16Bytes(ts uint16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, ts)
	return b
}

// get seconds for the segment with a given start time
func get(targetB *bolt.Bucket, start uint32) uint16 {
	var long uint16

	bs := targetB.Get(getUInt32Bytes(start))
	if bs != nil {
		long = binary.LittleEndian.Uint16(bs)
	}
	return long
}

func put(targetB *bolt.Bucket, start uint32, long uint16) error {
	return targetB.Put(getUInt32Bytes(start), getUInt16Bytes(long))
}
