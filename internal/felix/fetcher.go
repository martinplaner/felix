// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package felix

import (
	"context"
	"net"
	"time"
)

// PeriodicFetcher is the default fetcher that fetches using a fixed time interval.
type PeriodicFetcher struct {
	url           string
	src           Source
	scanner       Scanner
	items         chan<- Item
	links         chan<- Link
	log           Logger
	fetchInterval time.Duration
}

// Start starts a new periodic ASDFFOOBAR
func (f *PeriodicFetcher) Start(quit <-chan struct{}) {

	bg := context.Background()
	e := &emitter{
		items: f.items,
		links: f.links,
	}

L:
	for {
		select {
		case <-time.After(f.sleepDuration()):
			// TODO: make timeout configurable
			ctx, cancel := context.WithTimeout(bg, 30*time.Second)
			e.EmitFollow(f.url)

			for e.HasFollow() {
				followURL := e.NextFollow()

				r, err := f.src.Get(ctx, followURL)

				if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
					// TODO: handle temporary errors (backoff, retry)
					f.log.Error("temporary net error", "err", err, "url", f.url, "follow", followURL)
					cancel()
					continue
				}

				if err != nil {
					f.log.Error("could not get resource", "err", err, "url", f.url, "follow", followURL)
					cancel()
					continue
				}

				err = f.scanner.Scan(ctx, r, e)

				if err != nil {
					f.log.Error("could not scan content", "url", f.url, "follow", followURL)
					cancel()
					continue
				}
			}

			cancel()

		case <-quit:
			break L
		}
	}
}

func (f *PeriodicFetcher) sleepDuration() time.Duration {
	// TODO: calculate proper sleep duration since last try
	return f.fetchInterval
}
