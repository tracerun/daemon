package db

import (
	"time"
	"tracerun/lg"

	"github.com/boltdb/bolt"
	"go.uber.org/zap"
)

const (
	metaTag = "__metadata__"
)

var (
	rwDB   *bolt.DB
	readDB *bolt.DB
)

// CreateRWDB to create a read-write connection.
func CreateRWDB(p string) {
	var err error

	rwDB, err = bolt.Open(p, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		lg.L.Fatal("can't open read-write db", zap.Error(err))
	}

}

// CreateReadDB to create a read-only connection.
func CreateReadDB(p string) {
	var err error

	readDB, err = bolt.Open(p, 0666, &bolt.Options{ReadOnly: true})
	if err != nil {
		lg.L.Fatal("can't open read-only db", zap.Error(err))
	}
}

// EnqueueAction to enqueue the action.
func EnqueueAction() {

}
