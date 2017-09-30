// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package felix

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

type mockScanSource struct {
	sourceCalls int // number of times Source.Get was called
	scanCalls   int // number of times Scanner.Scan was called
}

func (ss *mockScanSource) Get(ctx context.Context, url string) (io.Reader, error) {
	ss.sourceCalls++

	switch ss.sourceCalls {
	case 1:
		return &bytes.Buffer{}, nil
	case 2:
		return nil, &tempNetError{"tempSourceError"}
	case 3:
		return nil, errors.New("otherSourceError")
	case 4, 5:
		return &bytes.Buffer{}, nil
	}

	return nil, errors.New("unexpectedError")
}

func (ss *mockScanSource) Scan(ctx context.Context, r io.Reader, e Emitter) error {
	ss.scanCalls++

	switch ss.scanCalls {
	case 1:
		e.EmitFollow("followurl1")
		e.EmitFollow("followurl2")
		e.EmitFollow("followurl3")
		e.EmitFollow("followurl4")
		return nil
	case 2:
		return errors.New("scanError")
	case 3:
		e.EmitItem(Item{Title: "emittedItem"})
		return nil
	}

	return errors.New("unexpectedError")
}

type tempNetError struct {
	msg string
}

func (e *tempNetError) Timeout() bool {
	panic("implement me")
}

func (e *tempNetError) Temporary() bool {
	return true
}

func (e *tempNetError) Error() string {
	return e.msg
}

func TestFetcher(t *testing.T) {
	items := make(chan Item)
	links := make(chan Link)

	scanSource := &mockScanSource{}

	nextFetch := func(n int) func() time.Duration {
		i := 0
		return func() time.Duration {
			defer func() { i++ }()
			if i < n {
				return 0
			} else {
				return 1000 * time.Hour
			}
		}
	}(1)

	logBuf := &bytes.Buffer{}
	log := NewLogger()
	log.SetOutput(logBuf)

	fetcher := &Fetcher{
		url:       "baseURL",
		source:    scanSource,
		scanner:   scanSource,
		items:     items,
		links:     links,
		log:       log,
		nextFetch: nextFetch,
	}

	quit := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		fetcher.Start(quit)
		wg.Done()
	}()

	<-items

	close(quit)
	wg.Wait()

	for _, s := range []string{"tempSourceError", "otherSourceError", "scanError"} {
		if !strings.Contains(logBuf.String(), s) {
			t.Errorf("could not find %q in log output", s)
		}
	}
}