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
	url     string
	source  Source
	scanner Scanner
	attempt Attempter
	items   chan<- Item
	links   chan<- Link
	log     Logger
}

// Attempter is used by Fetcher to determine if and when the next fetch attempt should be made.
type Attempter interface {
	// Next returns if and when the next attempt is scheduled
	Next(key string) (bool, time.Duration, error)
	// Inc increments the number of attempts by 1
	Inc(key string) error
}

// NextFetchFunc returns if the fetcher should continue to fetch and if so, how long to wait before the next attempt.
type NextFetchFunc func(url string) (bool, time.Duration)

func NewFetcher(url string, source Source, scanner Scanner, attempt Attempter, items chan<- Item, links chan<- Link) *Fetcher {
	return &Fetcher{
		url:     url,
		source:  source,
		scanner: scanner,
		attempt: attempt,
		items:   items,
		links:   links,
		log:     &NopLogger{},
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
		shouldContinue, nextFetch, err := f.attempt.Next(f.url)
		if err != nil {
			f.log.Error("could not get next attempt", "err", err)
			return
		}
		if !shouldContinue {
			f.log.Info("will not try to continue. quitting.", "url", f.url)
			return
		}
		f.log.Info("waiting", "nextFetch", nextFetch, "url", f.url)

		select {
		// TODO: abstract time stdlib away into clock, etc. for testing
		case <-time.After(nextFetch):
			if err := f.attempt.Inc(f.url); err != nil {
				f.log.Error("could not get increment attempt count", "err", err)
				return
			}
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

// attempt is the default Datastore-backed Attempter
type attempt struct {
	ds   Datastore
	next func(last time.Time, attempts int) (bool, time.Duration)
}

func (a attempt) Next(key string) (bool, time.Duration, error) {
	last, attempts, err := a.ds.LastAttempt(key)
	if err != nil {
		return false, 0, err
	}

	wait, untilNext := a.next(last, attempts)
	return wait, untilNext, nil
}

func (a attempt) Inc(key string) error {
	return a.ds.IncAttempt(key)
}

func PeriodicAttempter(ds Datastore, fetchInterval time.Duration) Attempter {
	next := func(last time.Time, attempts int) (bool, time.Duration) {
		nextTry := last.Add(fetchInterval)
		untilNext := nextTry.Sub(time.Now())
		return true, untilNext
	}

	return &attempt{
		ds:   ds,
		next: next,
	}
}

func FibAttempter(ds Datastore, baseInterval time.Duration, maxAttempts int) Attempter {
	var fib func(n int) int
	fib = func(n int) int {
		if n < 2 {
			return n
		}
		return fib(n-2) + fib(n-1)
	}

	next := func(last time.Time, attempts int) (bool, time.Duration) {
		if attempts >= maxAttempts {
			return false, 0
		}

		interval := time.Duration(fib(attempts)) * baseInterval
		nextTry := last.Add(interval)
		untilNext := nextTry.Sub(time.Now())
		return true, untilNext
	}

	return &attempt{
		ds:   ds,
		next: next,
	}
}
