// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package felix

import (
	"time"
)

// Item is a feed item that should be scraped for links.
type Item struct {
	Title   string
	URL     string
	PubDate time.Time
}

// Link is a link that was found in a feed or scraped from a page (Item)
type Link struct {
	Title string
	URL   string
}
