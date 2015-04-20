package db

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"hawx.me/code/phemera/models"
)

type Db interface {
	Get() models.Entries
	Save(time.Time, string)
	Close()
}

type BoltDb struct {
	me      *bolt.DB
	horizon time.Duration
}

const bucketName = "phemera"

func Open(path, horizon string) Db {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	ho, _ := time.ParseDuration(horizon)

	return BoltDb{db, ho}
}

func (db BoltDb) Get() models.Entries {
	list := models.Entries{}

	db.me.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		c := b.Cursor()

		bound := timestamp(time.Now().Add(db.horizon))
		for k, v := c.Last(); k != nil && bytes.Compare(k, bound) >= 0; k, v = c.Prev() {
			list = append(list, models.Entry{string(k), string(v)})
		}

		return nil
	})

	return list
}

func (db BoltDb) Save(key time.Time, value string) {
	db.me.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		err := b.Put([]byte(timestamp(key)), []byte(value))
		return err
	})
}

func (db BoltDb) Close() {
	db.me.Close()
}

func timestamp(t time.Time) []byte {
	return []byte(strconv.FormatInt(t.Unix(), 10))
}
