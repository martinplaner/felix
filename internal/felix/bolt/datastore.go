// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bolt

import (
	"time"

	"bytes"
	"encoding/gob"

	"github.com/boltdb/bolt"
	"github.com/martinplaner/felix/internal/felix"
	"github.com/pkg/errors"
)

// TODO: too much boilerplate... rewrite this using storm? (https://github.com/asdine/storm)

var (
	attemptBucket = []byte("attempts")
	itemBucket    = []byte("items")
	linkBucket    = []byte("links")
)

type datastore struct {
	db *bolt.DB
}

type attemptEntity struct {
	Last  time.Time
	Count int
}

type itemEntity struct {
	Item  felix.Item
	Added time.Time
}

type linkEntity struct {
	Link  felix.Link
	Added time.Time
}

func (ds datastore) Close() error {
	return ds.db.Close()
}

func (ds datastore) LastAttempt(key string) (time.Time, int, error) {
	var attempt attemptEntity

	err := ds.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(attemptBucket)
		buf := b.Get([]byte(key))

		if buf == nil {
			attempt.Last = time.Time{}
			attempt.Count = 0
			return nil
		}

		if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&attempt); err != nil {
			return errors.Wrap(err, "could not decode entity")
		}

		return nil
	})

	if err != nil {
		return time.Time{}, 0, err
	}

	return attempt.Last, attempt.Count, nil
}

func (ds datastore) IncAttempt(key string) error {
	return ds.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(attemptBucket)
		buf := b.Get([]byte(key))

		var attempt = attemptEntity{Last: time.Time{}, Count: 0}
		if buf != nil {
			if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&attempt); err != nil {
				return errors.Wrap(err, "could not decode entity")
			}
		}

		attempt.Count++
		attempt.Last = time.Now()

		if err := put(b, []byte(key), attempt); err != nil {
			return errors.Wrap(err, "could not store entity")
		}

		return nil
	})
}

func (ds datastore) StoreItem(item felix.Item) (exists bool, e error) {
	exists = false
	err := ds.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(itemBucket)
		key := []byte(item.URL)
		buf := b.Get(key)

		if buf != nil {
			// Do not store if it already exists
			exists = true
			return nil
		}

		entity := itemEntity{
			Item:  item,
			Added: time.Now(),
		}

		if err := put(b, key, entity); err != nil {
			return errors.Wrap(err, "could not store entity")
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (ds datastore) StoreLink(link felix.Link) (exists bool, e error) {
	exists = false
	err := ds.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(linkBucket)
		key := []byte(link.URL)
		buf := b.Get(key)

		if buf != nil {
			// Do not store if it already exists
			exists = true
			return nil
		}

		entity := linkEntity{
			Link:  link,
			Added: time.Now(),
		}

		if err := put(b, key, entity); err != nil {
			return errors.Wrap(err, "could not store entity")
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (ds datastore) GetItems(maxAge time.Duration) ([]felix.Item, error) {
	var items []felix.Item
	var cutoff = time.Now().Add(-maxAge)

	err := ds.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(itemBucket)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var entity itemEntity
			if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&entity); err != nil {
				return errors.Wrap(err, "could not decode entity")
			}

			if entity.Added.After(cutoff) {
				items = append(items, entity.Item)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return items, nil
}

func (ds datastore) GetLinks(maxAge time.Duration) ([]felix.Link, error) {
	var links []felix.Link
	var cutoff = time.Now().Add(-maxAge)

	err := ds.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(linkBucket)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var entity linkEntity
			if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&entity); err != nil {
				return errors.Wrap(err, "could not decode entity")
			}

			if entity.Added.After(cutoff) {
				links = append(links, entity.Link)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return links, nil
}

func (ds datastore) Cleanup(maxAge time.Duration) error {
	var cutoff = time.Now().Add(-maxAge)

	return ds.db.Update(func(tx *bolt.Tx) error {
		itemCursor := tx.Bucket(itemBucket).Cursor()
		for k, v := itemCursor.First(); k != nil; k, v = itemCursor.Next() {
			var entity itemEntity
			if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&entity); err != nil {
				return errors.Wrap(err, "could not decode entity")
			}

			if entity.Added.Before(cutoff) {
				itemCursor.Delete()
			}
		}

		linkCursor := tx.Bucket(linkBucket).Cursor()
		for k, v := linkCursor.First(); k != nil; k, v = linkCursor.Next() {
			var entity linkEntity
			if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&entity); err != nil {
				return errors.Wrap(err, "could not decode entity")
			}

			if entity.Added.Before(cutoff) {
				linkCursor.Delete()
			}
		}

		tryCursor := tx.Bucket(attemptBucket).Cursor()
		for k, v := tryCursor.First(); k != nil; k, v = tryCursor.Next() {
			var entity attemptEntity
			if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&entity); err != nil {
				return errors.Wrap(err, "could not decode entity")
			}

			if entity.Last.Before(cutoff) {
				tryCursor.Delete()
			}
		}

		return nil
	})
}

func put(b *bolt.Bucket, key []byte, v interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		return errors.Wrap(err, "could not encode entity")
	}

	if err := b.Put(key, buf.Bytes()); err != nil {
		return errors.Wrap(err, "could not store entity")
	}

	return nil
}

func NewDatastore(filename string) (felix.Datastore, error) {
	db, err := bolt.Open(filename, 0600, &bolt.Options{Timeout: 10 * time.Second})

	if err != nil {
		return nil, errors.Wrap(err, "could not open bolt db")
	}

	err = db.Update(func(tx *bolt.Tx) error {
		for _, name := range [][]byte{attemptBucket, itemBucket, linkBucket} {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return errors.Wrapf(err, "could not create bucket %s", name)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &datastore{db}, nil
}
