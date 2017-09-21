// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package felix

// Emitter is used by a Scanner to emit Items, Links and Follow URLs that should be processed.
type Emitter interface {
	EmitItem(item Item)
	EmitLink(link Link)
	EmitFollow(follow string)
}

type emitter struct {
	items   chan<- Item
	links   chan<- Link
	follows []string
}

// EmitItem emits an Item to be processed.
func (e *emitter) EmitItem(item Item) {
	e.items <- item
}

// EmitLink emits an Link to be processed.
func (e *emitter) EmitLink(link Link) {
	e.links <- link
}

// EmitFollow emits a URL to be followed by the fetcher.
func (e *emitter) EmitFollow(follow string) {
	e.follows = append(e.follows, follow)
}

// HasFollow returns true if at least one follow URL is available.
func (e *emitter) HasFollow() bool {
	return len(e.follows) > 0
}

// NextFollow returns the next follow URL.
//
// This funtion panics if a follow URL is not available,
// therefore HasFollow must be called beforehand.
func (e *emitter) NextFollow() string {
	next := e.follows[0]
	e.follows = e.follows[1:]
	return next
}
