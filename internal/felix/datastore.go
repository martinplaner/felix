// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package felix

import "time"

type Datastore interface {
	// TODO: rename 'try' to 'attempt'?
	AddTry(key string) (last time.Time, tries int, err error)
	StoreItem(item Item) (bool, error)
	StoreLink(link Link) (bool, error)
	GetItems(maxAge time.Duration) ([]Item, error)
	GetLinks(maxAge time.Duration) ([]Link, error)
	Cleanup(maxAge time.Duration) error
	Close() error
}
