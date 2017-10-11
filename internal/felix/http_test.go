// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package felix

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"time"

	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
)

func TestStringHandler(t *testing.T) {
	testCases := []struct {
		desc   string
		s      string
		status int
	}{
		{
			desc:   "empty string",
			s:      "",
			status: http.StatusOK,
		},
		{
			desc:   "test string",
			s:      "teststring",
			status: http.StatusOK,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			StringHandler(tC.s).ServeHTTP(w, req)

			resp := w.Result()
			body, _ := ioutil.ReadAll(resp.Body)

			if resp.StatusCode != tC.status {
				t.Errorf("unexpected status code. expected %v, got %v", tC.status, resp.StatusCode)
			}

			if string(body) != tC.s {
				t.Errorf("unexpected response. expected %q, got %q", tC.s, string(body))
			}
		})
	}
}

func TestFeedHandler(t *testing.T) {
	testCases := []struct {
		desc   string
		links  []Link
		err    error
		status int
	}{
		{
			desc:   "feed with no links",
			links:  []Link{},
			status: http.StatusOK,
			err:    nil,
		},
		{
			desc: "feed with two links",
			links: []Link{
				{Title: "title1", URL: "http://example.com"},
				{Title: "title2", URL: "http://example.org"},
			},
			status: http.StatusOK,
			err:    nil,
		},
		{
			desc:   "datastore error",
			links:  []Link{},
			status: http.StatusInternalServerError,
			err:    errors.New("expected test failure"),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			ds := &mockDatastore{links: tC.links, err: tC.err}
			FeedHandler(ds, 0).ServeHTTP(w, req)

			resp := w.Result()

			if resp.StatusCode != tC.status {
				t.Errorf("unexpected status code. expected %v, got %v", tC.status, resp.StatusCode)
			}

			if tC.err != nil {
				// do not test feed parsing to failed requests
				return
			}

			feed, err := gofeed.NewParser().Parse(resp.Body)
			if err != nil {
				t.Error("could not parse output feed:", err)
			}

			for i, item := range feed.Items {
				if item.Title != tC.links[i].Title {
					t.Errorf("feed item %d title: %q != %q", item.Title, tC.links[i].Title)
				}
			}
		})
	}
}

type mockDatastore struct {
	Datastore

	links []Link
	err   error
}

func (m *mockDatastore) GetLinks(maxAge time.Duration) ([]Link, error) {
	return m.links, m.err
}
