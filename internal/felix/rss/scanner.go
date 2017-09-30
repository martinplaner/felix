// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rss

import (
	"context"
	"io"

	"github.com/pkg/errors"

	"time"

	"github.com/martinplaner/felix/internal/felix"
	"github.com/mmcdole/gofeed"
)

// ItemScanner parses r as an RSS feed and extracts all feed items.
var ItemScanner = felix.ScanFunc(func(ctx context.Context, r io.Reader, e felix.Emitter) error {
	fp := gofeed.NewParser()
	feed, err := fp.Parse(r)

	if err != nil {
		return errors.Wrap(err, "could not parse RSS feed")
	}

	for _, item := range feed.Items {
		e.EmitItem(felix.Item{
			Title:   item.Title,
			URL:     item.Link,
			PubDate: time.Now(),
			//PubDate: *item.PublishedParsed, // -> crash on nil, fix maybe
		})
	}

	return nil
})
