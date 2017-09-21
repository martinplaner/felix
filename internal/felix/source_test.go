// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package felix

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSource(t *testing.T) {

	testCases := []struct {
		desc       string
		content    []byte
		testServer bool
		wantError  bool
	}{
		{
			desc:       "normal, nonempty content",
			content:    []byte("veryimportanttestdata"),
			testServer: true,
			wantError:  false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var url = "http://invalid.url"

			if tC.testServer {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write(tC.content)
				}))
				url = ts.URL
				defer ts.Close()
			}

			source := NewHTTPSource(nil)

			r, err := source.Get(context.Background(), url)

			if (err != nil) != tC.wantError {
				t.Errorf("unexpected error return value! wantError = %q, got: %q", tC.wantError, err)
			}

			got, err := ioutil.ReadAll(r)

			if err != nil {
				t.Error("could not read response:", err)
			}

			if !bytes.Equal(got, tC.content) {
				t.Errorf("received content does not match expected! want: %v, got: %v", tC.content, got)
			}
		})
	}

}
