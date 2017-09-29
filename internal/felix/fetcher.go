// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package felix

import (
	"context"
	"net"
	"time"
)

// Fetcher is the default fetcher.
type Fetcher struct {
	url       string
	source    Source
	scanner   Scanner
	nextFetch func() (time.Duration, bool)
	items     chan<- Item
	links     chan<- Link
	log       Logger
}

func NewFetcher(url string, source Source, scanner Scanner, nextFetch func() (time.Duration, bool), items chan<- Item, links chan<- Link) *Fetcher {
	return &Fetcher{
		url:       url,
		source:    source,
		scanner:   scanner,
		nextFetch: nextFetch,
		items:     items,
		links:     links,
		log:       &NopLogger{},
	}
}

func (f *Fetcher) SetLogger(log Logger) {
	// TODO: consolidate into sublogger with common fields (url, etc.)
	f.log = log
}

// Start starts the fetching.
func (f *Fetcher) Start(quit <-chan struct{}) {

	bg := context.Background()
	e := &emitter{
		items: f.items,
		links: f.links,
	}

L:
	for {
		nextFetch, shouldContinue := f.nextFetch()
		if !shouldContinue {
			f.log.Info("will not try to continue. quitting.", "url", f.url)
			return
		}

		select {
		// TODO: abstract time stdlib away into clock, etc. for testing
		case <-time.After(nextFetch):
			ctx, cancel := context.WithTimeout(bg, 20*time.Second)
			e.EmitFollow(f.url)

			for e.HasFollow() {
				followURL := e.NextFollow()

				r, err := f.source.Get(ctx, followURL)

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
					f.log.Error("could not scan content", "err", err, "url", f.url, "follow", followURL)
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
