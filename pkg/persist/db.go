package persist

import (
	"github.com/dgraph-io/badger/v3"
	"log"
)

func OpenDB(path string) *badger.DB {
	opts := badger.DefaultOptions(path).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func SaveBlock(db *badger.DB, hash string, data []byte) error {
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(hash), data)
	})
}

func LoadBlock(db *badger.DB, hash string) ([]byte, error) {
	var val []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(hash))
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		return err
	})
	return val, err
}
