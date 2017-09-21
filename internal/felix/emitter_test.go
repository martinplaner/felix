// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package felix

import "testing"

func TestEmitter(t *testing.T) {
	var testTitle = "title"
	var testURL = "http://example.com"
	var testFollow = []string{"http://example.com", "http://example.org"}

	// Initialize with capacity to not deadlock tests
	links := make(chan Link, 10)
	items := make(chan Item, 10)

	e := emitter{
		links: links,
		items: items,
	}

	// EmitItem

	e.EmitItem(Item{
		Title: testTitle,
	})

	if i, ok := <-items; !ok || i.Title != testTitle {
		t.Errorf("invalid item title received: expected %q, got %q", testTitle, i.Title)
	}

	// EmitLink

	e.EmitLink(Link{
		URL: testURL,
	})

	if l, ok := <-links; !ok || l.URL != testURL {
		t.Errorf("invalid link URL received: expected %q, got %q", testURL, l.URL)
	}

	// EmitFollow

	if hasFollow := e.HasFollow(); hasFollow {
		t.Errorf("invalid HasFollow: expected %v, got %v", false, hasFollow)
	}

	for _, f := range testFollow {
		e.EmitFollow(f)
	}

	i := 0
	for e.HasFollow() {
		f := e.NextFollow()

		if f != testFollow[i] {
			t.Errorf("invalid follow: expected %q, got %q", testFollow[i], f)
		}
		i++
	}
}
