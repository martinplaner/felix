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

func TestDatastore_AddTry(t *testing.T) {
	ds, close := newDatastore(t)
	defer close()
	key := "key"

	t.Run("first try should return 0 tries", func(t *testing.T) {
		_, tries, err := ds.AddTry(key)

		assertNilError(t, err)

		if tries != 0 {
			t.Errorf("unexpected number of tries. expected %v, got %v", 0, tries)
		}
	})

	t.Run("second try should return 1 previous try", func(t *testing.T) {
		_, tries, err := ds.AddTry(key)

		assertNilError(t, err)

		if tries != 1 {
			t.Errorf("unexpected number of tries. expected %v, got %v", 1, tries)
		}
	})

	t.Run("another key should again return 0 tries", func(t *testing.T) {
		_, tries, err := ds.AddTry(key + "2")

		assertNilError(t, err)

		if tries != 0 {
			t.Errorf("unexpected number of tries. expected %v, got %v", 0, tries)
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
	_, _, err = ds.AddTry(tryKey)
	assertNilError(t, err)

	t.Run("should not remove entries in maxAge window", func(t *testing.T) {
		err := ds.Cleanup(10 * time.Hour)
		assertNilError(t, err)

		items, err := ds.GetItems(1 * time.Hour)
		assertNilError(t, err)
		links, err := ds.GetLinks(1 * time.Hour)
		assertNilError(t, err)
		_, tries, err := ds.AddTry(tryKey)
		assertNilError(t, err)

		if len(items) != 1 || len(links) != 1 || tries != 2 {
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
		_, tries, err := ds.AddTry(tryKey)
		assertNilError(t, err)

		if len(items) != 0 || len(links) != 0 || tries != 1 {
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
