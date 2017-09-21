// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mock

import "github.com/martinplaner/felix/internal/felix"

// Emitter is a mock for felix.Emitter
type Emitter struct {
	Items   []felix.Item
	Links   []felix.Link
	Follows []string
}

// EmitItem emits an Item to be processed.
func (e *Emitter) EmitItem(item felix.Item) {
	e.Items = append(e.Items, item)
}

// EmitLink emits an Link to be processed.
func (e *Emitter) EmitLink(link felix.Link) {
	e.Links = append(e.Links, link)
}

// EmitFollow emits a URL to be followed by the fetcher.
func (e *Emitter) EmitFollow(follow string) {
	e.Follows = append(e.Follows, follow)
}
