package db

import (
	"github.com/hawx/phemera/models"
	"github.com/jmhodges/levigo"
	"log"
	"sort"
	"strconv"
	"time"
)

type Db interface {
	Get() models.Entries
	Save(time.Time, string)
	Close()
}

type LevelDb struct {
	me      *levigo.DB
	wo      *levigo.WriteOptions
	ro      *levigo.ReadOptions
	horizon time.Duration
}

func Open(path, horizon string) Db {
	opts := levigo.NewOptions()
	opts.SetCache(levigo.NewLRUCache(3 << 30))
	opts.SetCreateIfMissing(true)
	db, err := levigo.Open(path, opts)

	if err != nil {
		log.Fatal("db:", err)
	}

	wo := levigo.NewWriteOptions()

	ro := levigo.NewReadOptions()
	ro.SetFillCache(false)

	ho, _ := time.ParseDuration(horizon)

	return LevelDb{db, wo, ro, ho}
}

func timestamp(t time.Time) []byte {
	return []byte(strconv.FormatInt(t.Unix(), 10))
}

func (db LevelDb) Get() models.Entries {
	it := db.me.NewIterator(db.ro)
	defer it.Close()

	it.Seek(timestamp(time.Now().Add(db.horizon)))

	list := models.Entries{}
	for it = it; it.Valid(); it.Next() {
		list = append(list, models.Entry{string(it.Key()), string(it.Value())})
	}

	sort.Sort(sort.Reverse(list))
	return list
}

func (db LevelDb) Save(t time.Time, body string) {
	db.me.Put(db.wo, []byte(timestamp(t)), []byte(body))
}

func (db LevelDb) Close() {
	db.me.Close()
	db.wo.Close()
	db.ro.Close()
}
