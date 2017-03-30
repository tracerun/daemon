package db

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/user"
	"runtime"
	"time"
	"tracerun/lg"

	"github.com/boltdb/bolt"
	"github.com/oklog/ulid"
	"go.uber.org/zap"
)

const (
	metaBucket = "__metadata__"
	version    = uint8(1)

	versionKey  = "version"
	tagKey      = "tag"
	hostKey     = "host"
	archKey     = "arch"
	usernameKey = "username"
	osKey       = "os"
)

var (
	// ErrMetaNotFound means metadata is not set
	ErrMetaNotFound = errors.New("metadata not existed")
)

// Metadata for certain db.
type Metadata struct {
	Version  uint8
	Tag      string
	Host     string
	Arch     string
	Username string
	OS       string
	Create   uint64
}

// initMeta to initialize the meta tag.
// Meta contains:
// "version" current db version
// "tag" a unique tag for db, ULID (https://github.com/alizain/ulid)
// "host" the hostname
// "arch" the arch
// "username" the user
// "os" the operation system
func initMeta(rw *bolt.DB) error {
	// check whether __metadata__ existed
	var existed bool
	if err := rw.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(metaBucket))
		existed = b != nil
		return nil
	}); err != nil {
		return err
	}

	// existed, return
	if existed {
		return nil
	}

	if rw.IsReadOnly() {
		return ErrReadOnlyDB
	}

	var err error
	var id ulid.ULID
	var host string
	var username string

	now := time.Now()
	entropy := rand.New(rand.NewSource(now.UnixNano()))
	if id, err = ulid.New(ulid.Timestamp(now), entropy); err != nil {
		return err
	}
	lg.L.Debug("id generated", zap.Any("id", id))

	if host, err = os.Hostname(); err != nil {
		return err
	}

	if usr, err := user.Current(); err != nil {
		return err
	} else {
		username = usr.Username
	}

	// Set __metadata__
	return rw.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(metaBucket))
		if err != nil {
			return err
		}
		if err := b.Put([]byte(versionKey), []byte{byte(version)}); err != nil {
			return err
		}
		if err := b.Put([]byte(tagKey), id[:]); err != nil {
			return err
		}
		if err := b.Put([]byte(hostKey), []byte(host)); err != nil {
			return err
		}
		if err := b.Put([]byte(archKey), []byte(runtime.GOARCH)); err != nil {
			return err
		}
		if err := b.Put([]byte(usernameKey), []byte(username)); err != nil {
			return err
		}
		if err := b.Put([]byte(osKey), []byte(runtime.GOOS)); err != nil {
			return err
		}
		lg.L.Debug("metadata set", zap.Uint8("version", version), zap.ByteString("tag", id[:]), zap.String("host", host), zap.String("arch", runtime.GOARCH), zap.String("username", username), zap.String("os", runtime.GOOS))
		return nil
	})
}

// GetMetaData to get metadata for the db.
func GetMetaData(rdb *bolt.DB) (Metadata, error) {
	var metaData Metadata
	err := rdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(metaBucket))
		if b == nil {
			return ErrMetaNotFound
		}

		versionBytes := b.Get([]byte(versionKey))
		if l := len(versionBytes); l != 1 {
			return fmt.Errorf("wrong version in metadata, byte length should be 1, but is %d", l)
		}
		metaData.Version = uint8(versionBytes[0])

		var uid ulid.ULID
		tagBytes := b.Get([]byte(tagKey))
		if l := len(tagBytes); l != 16 {
			return fmt.Errorf("wrong tag in metadata, byte length should be 1, but is %d (%s)", l, string(tagBytes))
		}
		copy(uid[:], tagBytes[:16])

		metaData.Tag = uid.String()
		metaData.Create = uid.Time() / 1000

		metaData.Host = string(b.Get([]byte(hostKey)))
		metaData.Arch = string(b.Get([]byte(archKey)))
		metaData.Username = string(b.Get([]byte(usernameKey)))
		metaData.OS = string(b.Get([]byte(osKey)))
		return nil
	})
	return metaData, err
}
