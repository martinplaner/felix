package felix

import "time"

type Datastore interface {
	AddTry(key string) (last time.Time, tries int, err error)
	StoreItem(item Item) (bool, error)
	StoreLink(link Link) (bool, error)
	GetItems(maxAge time.Duration) ([]Item, error)
	GetLinks(maxAge time.Duration) ([]Link, error)
	Cleanup(maxAge time.Duration) error
	Close() error
}
