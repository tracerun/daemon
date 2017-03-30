package db

import (
	"errors"
	"time"

	"github.com/boltdb/bolt"
)

var (
	// ErrReadOnlyDB error to do write operation
	ErrReadOnlyDB = errors.New("ReadOnly DB can't write")

	rwDB   *bolt.DB
	readDB *bolt.DB
)

// CreateRWDB to create a read-write connection.
func CreateRWDB(p string) error {
	var err error

	rwDB, err = bolt.Open(p, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	return initMeta(rwDB)
}

// CreateReadDB to create a read-only connection.
func CreateReadDB(p string) error {
	var err error
	readDB, err = bolt.Open(p, 0666, &bolt.Options{ReadOnly: true, Timeout: 1 * time.Second})
	return err
}

// CloseDB to close db
func CloseDB() error {
	if rwDB != nil {
		if err := rwDB.Close(); err != nil {
			return err
		}
	}
	if readDB != nil {
		if err := readDB.Close(); err != nil {
			return err
		}
	}
	return nil
}

// EnqueueAction to enqueue the action.
func EnqueueAction() {

}
