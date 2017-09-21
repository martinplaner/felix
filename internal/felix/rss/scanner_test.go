// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rss

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/martinplaner/felix/internal/felix/mock"
)

var (
	empty      = ""
	invalidRSS = "<html></html>"
)

func TestRSSScanner(t *testing.T) {
	testCases := []struct {
		desc          string
		content       string
		expectedItems int
		wantError     bool
	}{
		{
			desc:          "invalid feed",
			content:       invalidRSS,
			expectedItems: 0,
			wantError:     true,
		},
		{
			desc:          "empty feed",
			content:       empty,
			expectedItems: 0,
			wantError:     true,
		},
		{
			desc:          "valid feed",
			content:       mustReadFile(t, "testdata/feed_rss_utf8.xml"),
			expectedItems: 3,
			wantError:     false,
		},
	}

	ctx := context.Background()

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			e := &mock.Emitter{}
			r := strings.NewReader(tC.content)
			err := ItemScanner.Scan(ctx, r, e)

			if (err != nil) != tC.wantError {
				t.Errorf("unexpected error return value! wantError = %v, got: %v", tC.wantError, err)
			}

			if len(e.Items) != tC.expectedItems {
				t.Errorf("invalid number of items found! expected: %v, got: %v", tC.expectedItems, len(e.Items))
			}
		})
	}
}

// read file content to string, fatal on error
func mustReadFile(t *testing.T, filename string) string {
	t.Helper()

	b, err := ioutil.ReadFile(filename)

	if err != nil {
		t.Fatal("could not read file:", err)
	}

	return string(b)
}
