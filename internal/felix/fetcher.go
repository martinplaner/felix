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
	nextFetch NextFetchFunc
	items     chan<- Item
	links     chan<- Link
	log       Logger
}

// NextFetchFunc returns if the fetcher should continue to fetch and if so, how long to wait before the next attempt.
type NextFetchFunc func(url string) (bool, time.Duration)

func NewFetcher(url string, source Source, scanner Scanner, nextFetch NextFetchFunc, items chan<- Item, links chan<- Link) *Fetcher {
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

	f.log.Info("started fetcher", "url", f.url)

	bg := context.Background()
	e := &emitter{
		items: f.items,
		links: f.links,
	}

L:
	for {
		shouldContinue, nextFetch := f.nextFetch(f.url)
		if !shouldContinue {
			f.log.Info("will not try to continue. quitting.", "url", f.url)
			return
		}
		f.log.Info("waiting", "nextFetch", nextFetch, "url", f.url)

		select {
		// TODO: abstract time stdlib away into clock, etc. for testing
		case <-time.After(nextFetch):
			ctx, cancel := context.WithTimeout(bg, 15*time.Second)
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

// FibWait returns a NextFetchFunc that
func FibWait(ds Datastore, baseInterval time.Duration, maxTries int) NextFetchFunc {
	var fib func(n int) int
	fib = func(n int) int {
		if n < 2 {
			return n
		}
		return fib(n-2) + fib(n-1)
	}

	return func(url string) (bool, time.Duration) {
		lastTry, previousTries, err := ds.AddTry(url)

		if err != nil {
			// TODO: re-evaluate this decision. this silently hides an error :-/
			return true, baseInterval
		}

		if previousTries >= maxTries {
			return false, time.Duration(0)
		}

		interval := time.Duration(fib(previousTries)) * baseInterval
		nextTry := lastTry.Add(interval)
		untilNext := nextTry.Sub(time.Now())

		return true, untilNext
	}
}

func PeriodicWait(ds Datastore, fetchInterval time.Duration) NextFetchFunc {
	return func(url string) (bool, time.Duration) {
		lastTry, _, err := ds.AddTry(url)

		if err != nil {
			// TODO: re-evaluate this decision. this silently hides an error :-/
			return true, fetchInterval
		}

		nextTry := lastTry.Add(fetchInterval)
		untilNext := nextTry.Sub(time.Now())

		return true, untilNext
	}
}
