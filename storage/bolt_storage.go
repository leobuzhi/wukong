package storage

import (
	"time"

	"github.com/boltdb/bolt"
)

var wukongDocuments = []byte("wukongDocuments")

type boltStorage struct {
	db *bolt.DB
}

func openBoltStorage(path string) (Storage, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 3600 * time.Second})
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(wukongDocuments)
		return err
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	return &boltStorage{db}, nil
}

func (s *boltStorage) WALName() string {
	return s.db.Path()
}

func (s *boltStorage) Set(k []byte, v []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(wukongDocuments).Put(k, v)
	})
}

func (s *boltStorage) Get(k []byte) (b []byte, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		b = tx.Bucket(wukongDocuments).Get(k)
		return nil
	})
	return
}

func (s *boltStorage) Delete(k []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(wukongDocuments).Delete(k)
	})
}

func (s *boltStorage) ForEach(fn func(k, v []byte) error) error {
	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(wukongDocuments)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if err := fn(k, v); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *boltStorage) Close() error {
	return s.db.Close()
}
