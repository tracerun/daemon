package segment

import (
	"tracerun/lg"

	"github.com/boltdb/bolt"
	"go.uber.org/zap"
)

const (
	segBucket = "__segments__"
)

// Generate to generate a segment for a given target.
func Generate(tx *bolt.Tx, target string, start uint32, seg uint16) error {
	lg.L.Debug("new segment", zap.String("target", target), zap.Uint32("start", start), zap.Uint16("seg", seg))
	// TODO
	return nil
}
