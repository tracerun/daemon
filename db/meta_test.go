package db

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/drkaka/ulid"
	"github.com/stretchr/testify/assert"
	"github.com/tracerun/tracerun/lg"
)

func TestMetaData(t *testing.T) {
	lg.InitLogger(true, false, "")

	id, err := ulid.NewFromTime(time.Now())
	assert.NoError(t, err, "error while creating an ULID")
	if err != nil {
		t.Fatal(err)
	}

	dbPath := fmt.Sprintf("%s.db", id.String())

	// create db
	err = CreateRWDB(dbPath)
	assert.NoError(t, err, "error while open read-write db")
	defer func() {
		err = CloseDB()
		assert.NoError(t, err, "error while closing rwdb")

		err = os.Remove(dbPath)
		assert.NoError(t, err, "error while removing file")
	}()

	// create readonly db should fail
	err = CreateReadDB(dbPath)
	assert.Equal(t, bolt.ErrTimeout, err, "error should pop while open a readonly db when rwdb is connected")

	// init again should be ok.
	err = initMeta(rwDB)
	assert.NoError(t, err, "error while init metadata")

	// get metadata
	metaData, err := GetMetaData(rwDB)
	assert.NoError(t, err, "error while getting metadata")
	if err != nil {
		t.Fatal(err)
	}

	// check version info
	assert.Equal(t, version, metaData.Version, "Version is not correct.")
	// check tag info
	assert.Equal(t, 26, len(metaData.Tag), "Tag length is not 26. For an ULID, string length should be 26.")
	// check host info
	host, err := os.Hostname()
	assert.NoError(t, err, "error while getting host information.")
	assert.EqualValues(t, host, metaData.Host, "Host is wrong.")
	// check username info
	usr, err := user.Current()
	assert.NoError(t, err, "error while getting username information.")
	assert.EqualValues(t, usr.Username, metaData.Username, "Host is wrong.")
	// check OS info
	assert.EqualValues(t, runtime.GOOS, metaData.OS, "OS is wrong.")
	// check arch info
	assert.EqualValues(t, runtime.GOARCH, metaData.Arch, "Arch is not correct.")
	// check creation info
	assert.InDelta(t, metaData.Create, id.Time()/1000, 1, "Creation is not correct.")
}
