package db

import "github.com/boltdb/bolt"

const (
	actionBucket = "__actions__"

	// action will be expired after 15 seconds
	expired = 15
)

// GetAction to get an action.
// start and last active timestamp will be returned
func GetAction(tx *bolt.Tx, target string) (uint32, uint32) {
	var start, last uint32
	return start, last
}
