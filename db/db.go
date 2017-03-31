package db

import (
	"errors"
	"time"

	"github.com/boltdb/bolt"
)

var (
	// ErrReadOnlyDB error to do write operation
	ErrReadOnlyDB = errors.New("ReadOnly DB can't write")

	dbPath  string
	timeout = 3 * time.Second
)

// SetDBPath to set db file path.
func SetDBPath(p string) {
	dbPath = p
}

// CreateRWDB to create a read-write connection.
func CreateRWDB() (*bolt.DB, error) {
	rwDB, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: timeout})
	if err != nil {
		return nil, err
	}
	return rwDB, initMeta(rwDB)
}

// CreateReadDB to create a read-only connection.
func CreateReadDB() (*bolt.DB, error) {
	return bolt.Open(dbPath, 0666, &bolt.Options{ReadOnly: true, Timeout: timeout})
}
