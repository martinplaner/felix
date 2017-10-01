package bolt

import (
	"testing"

	"io/ioutil"

	"os"

	"time"

	"github.com/martinplaner/felix/internal/felix"
)

// TODO: look for assertion library?

var _ felix.Datastore = new(datastore)

func TestDatastore_GetItems(t *testing.T) {
	ds, close := newDatastore(t)
	defer close()

	item := felix.Item{
		URL:     "http://example.com",
		Title:   "Item Title 1",
		PubDate: time.Now(),
	}

	t.Run("empty datastore should not return any items", func(t *testing.T) {
		items, err := ds.GetItems(1 * time.Hour)

		assertNilError(t, err)

		if len(items) != 0 {
			t.Errorf("unexpected number of returned items. expected %v, got %v", 0, len(items))
		}
	})

	t.Run("should return item after storing", func(t *testing.T) {
		if _, err := ds.StoreItem(item); err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		items, err := ds.GetItems(1 * time.Hour)

		assertNilError(t, err)

		if len(items) != 1 {
			t.Errorf("unexpected number of returned items. expected %v, got %v", 1, len(items))
		}
	})

	t.Run("should not return items with 0 maxAge", func(t *testing.T) {
		items, err := ds.GetItems(0 * time.Second)

		assertNilError(t, err)

		if len(items) != 0 {
			t.Errorf("unexpected number of returned items. expected %v, got %v", 0, len(items))
		}
	})
}

func TestDatastore_StoreItem(t *testing.T) {
	ds, close := newDatastore(t)
	defer close()

	item := felix.Item{
		URL:     "http://example.com",
		Title:   "Item Title 1",
		PubDate: time.Now(),
	}

	t.Run("item should not exist on first store", func(t *testing.T) {
		didExist, err := ds.StoreItem(item)

		assertNilError(t, err)

		if didExist != false {
			t.Errorf("unexpected exist status. expected %v, got %v", false, didExist)
		}
	})

	t.Run("item should already exist on second store", func(t *testing.T) {
		didExist, err := ds.StoreItem(item)

		assertNilError(t, err)

		if didExist != true {
			t.Errorf("unexpected exist status. expected %v, got %v", true, didExist)
		}
	})
}

func TestDatastore_GetLinks(t *testing.T) {
	ds, close := newDatastore(t)
	defer close()

	link := felix.Link{
		URL:   "http://example.com",
		Title: "Item Title 1",
	}

	t.Run("empty datastore should not return any links", func(t *testing.T) {
		items, err := ds.GetLinks(1 * time.Hour)

		assertNilError(t, err)

		if len(items) != 0 {
			t.Errorf("unexpected number of returned items. expected %v, got %v", 0, len(items))
		}
	})

	t.Run("should return link after storing", func(t *testing.T) {
		if _, err := ds.StoreLink(link); err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		items, err := ds.GetLinks(1 * time.Hour)

		assertNilError(t, err)

		if len(items) != 1 {
			t.Errorf("unexpected number of returned items. expected %v, got %v", 1, len(items))
		}
	})

	t.Run("should not return items with 0 maxAge", func(t *testing.T) {
		items, err := ds.GetLinks(0 * time.Second)

		assertNilError(t, err)

		if len(items) != 0 {
			t.Errorf("unexpected number of returned items. expected %v, got %v", 0, len(items))
		}
	})
}

func TestDatastore_StoreLink(t *testing.T) {
	ds, close := newDatastore(t)
	defer close()

	link := felix.Link{
		URL:   "http://example.com",
		Title: "Item Title 1",
	}

	t.Run("link should not exist on first store", func(t *testing.T) {
		didExist, err := ds.StoreLink(link)

		assertNilError(t, err)

		if didExist != false {
			t.Errorf("unexpected exist status. expected %v, got %v", false, didExist)
		}
	})

	t.Run("link should already exist on second store", func(t *testing.T) {
		didExist, err := ds.StoreLink(link)

		assertNilError(t, err)

		if didExist != true {
			t.Errorf("unexpected exist status. expected %v, got %v", true, didExist)
		}
	})
}

func TestDatastore_Attempts(t *testing.T) {
	ds, close := newDatastore(t)
	defer close()
	key := "key"

	t.Run("new key should return 0 attempts on every call without incremeting", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			_, attempts, err := ds.LastAttempt(key)

			assertNilError(t, err)

			if attempts != 0 {
				t.Errorf("unexpected number of attempts. expected %v, got %v", 0, attempts)
			}
		}
	})

	t.Run("same key should return incremented attempt on calls after incrementing", func(t *testing.T) {
		for i := 1; i < 5; i++ {
			assertNilError(t, ds.IncAttempt(key))
			_, attempts, err := ds.LastAttempt(key)
			assertNilError(t, err)

			if attempts != i {
				t.Errorf("unexpected number of attempts. expected %v, got %v", i, attempts)
			}
		}
	})
}

func TestDatastore_Cleanup(t *testing.T) {
	ds, close := newDatastore(t)
	defer close()
	tryKey := "key"

	// Insert some test data
	_, err := ds.StoreLink(felix.Link{URL: "http://example.com"})
	assertNilError(t, err)
	_, err = ds.StoreItem(felix.Item{URL: "http://example.com"})
	assertNilError(t, err)
	assertNilError(t, ds.IncAttempt(tryKey))

	t.Run("should not remove entries in maxAge window", func(t *testing.T) {
		err := ds.Cleanup(10 * time.Hour)
		assertNilError(t, err)

		items, err := ds.GetItems(1 * time.Hour)
		assertNilError(t, err)
		links, err := ds.GetLinks(1 * time.Hour)
		assertNilError(t, err)
		_, attempts, err := ds.LastAttempt(tryKey)
		assertNilError(t, err)
		assertNilError(t, ds.IncAttempt(tryKey))

		if len(items) != 1 || len(links) != 1 || attempts != 1 {
			t.Error("inconsistent state. cleanup should not have removed anything.")
		}
	})

	t.Run("should remove entries in zero maxAge", func(t *testing.T) {
		err := ds.Cleanup(0 * time.Second)
		assertNilError(t, err)

		items, err := ds.GetItems(1 * time.Hour)
		assertNilError(t, err)
		links, err := ds.GetLinks(1 * time.Hour)
		assertNilError(t, err)
		_, attempts, err := ds.LastAttempt(tryKey)
		assertNilError(t, err)
		assertNilError(t, ds.IncAttempt(tryKey))

		if len(items) != 0 || len(links) != 0 || attempts != 0 {
			t.Error("inconsistent state. cleanup should have removed everything.")
		}
	})
}

func assertNilError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func newDatastore(t *testing.T) (felix.Datastore, func()) {
	t.Helper()
	f, err := ioutil.TempFile("", "bolt_test")
	if err != nil {
		t.Fatal("could not create temp file")
	}
	if f.Close() != nil {
		t.Fatal("could not close temp file")
	}

	ds, err := NewDatastore(f.Name())
	if err != nil {
		t.Fatal("could not create datastore")
	}

	close := func() {
		if ds.Close() != nil {
			t.Error("could not close datastore")
		}
		if os.Remove(f.Name()) != nil {
			t.Error("could not remove temp file")
		}
	}

	return ds, close
}
