package store

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ablankz/bloader/internal/config"
	"github.com/ablankz/bloader/internal/utils"
	"github.com/boltdb/bolt"
)

// BoltStore is a store that uses BoltDB
type BoltStore struct {
	db *bolt.DB
}

// SetupStore sets up the BoltStore
func (b *BoltStore) SetupStore(env string, conf config.ValidStoreConfig) error {
	for _, f := range conf.File {
		if f.Env == env {
			_, err := utils.CreateFileWithDir(f.Path)
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}
			db, err := bolt.Open(f.Path, 0600, &bolt.Options{
				// Timeout: 3 * time.Second,
			})
			if err != nil {
				return fmt.Errorf("failed to open bolt db: %v", err)
			}
			b.db = db
			if err := b.CreateBuckets(conf); err != nil {
				return fmt.Errorf("failed to create buckets: %v", err)
			}
			return nil
		}
	}

	return fmt.Errorf("no store config found for env %s", env)
}

// CreateBuckets creates buckets in the store
func (b *BoltStore) CreateBuckets(conf config.ValidStoreConfig) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range conf.Buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return fmt.Errorf("failed to create bucket: %v", err)
			}
		}
		return nil
	})
}

// PutObject puts an object in the store
func (b *BoltStore) PutObject(bucket, key string, data []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		return b.Put([]byte(key), data)
	})
}

// GetObject gets an object from the store
func (b *BoltStore) GetObject(bucket, key string) ([]byte, error) {
	var data []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		data = b.Get([]byte(key))
		return nil
	})
	return data, err
}

// PutObjectReader puts an object in the store from a reader
func (b *BoltStore) PutObjectReader(bucket, key string, reader io.Reader) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		data, err := io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("failed to read data: %v", err)
		}
		return b.Put([]byte(key), data)
	})
}

// GetObjectReader gets an object from the store as a reader
func (b *BoltStore) GetObjectReader(bucket, key string) (io.Reader, error) {
	var data []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		data = b.Get([]byte(key))
		return nil
	})
	return bytes.NewReader(data), err
}

// DeleteObject deletes an object from the store
func (b *BoltStore) DeleteObject(bucket, key string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		return b.Delete([]byte(key))
	})
}

// ListObjects lists all objects in a bucket
func (b *BoltStore) ListObjects(bucket string) ([]string, error) {
	var keys []string
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		return b.ForEach(func(k, v []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	})
	return keys, err
}

// ListBuckets lists all buckets in the store
func (b *BoltStore) ListBuckets() ([]string, error) {
	var buckets []string
	err := b.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			buckets = append(buckets, string(name))
			return nil
		})
	})
	return buckets, err
}

// Backup writes a backup of the store to a writer
func (b *BoltStore) Backup(writer io.Writer) (int, error) {
	var size int
	err := b.db.View(func(tx *bolt.Tx) error {
		if _, err := tx.WriteTo(writer); err != nil {
			return fmt.Errorf("failed to write backup: %v", err)
		}
		size = int(tx.Size())
		return nil
	})

	return size, err
}

// Clear clears the store
func (b *BoltStore) Clear() error {
	return b.db.Update(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			return tx.DeleteBucket(name)
		})
	})
}

// Close closes the store
func (b *BoltStore) Close() error {
	return b.db.Close()
}

var _ Store = &BoltStore{}
