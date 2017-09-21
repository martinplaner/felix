// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package felix

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/html/charset"

	"github.com/pkg/errors"
)

// Source retrieves the resource from the given URL and returns a reader for the content.
// If applicable, the reader shall only return UTF-8 encoded text.
// TODO: Find a better name?
type Source interface {
	Get(ctx context.Context, url string) (io.Reader, error)
}

// NewHTTPSource return a new Source for HTTP requests.
// A nil client will default to http.DefaultClient.
func NewHTTPSource(client *http.Client) Source {
	if client == nil {
		client = http.DefaultClient
	}

	return &httpSource{
		client: client,
	}
}

type httpSource struct {
	client *http.Client
}

var _ Source = new(httpSource)

func (s *httpSource) Get(ctx context.Context, url string) (io.Reader, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, errors.Wrap(err, "could not create request")
	}

	resp, err := s.client.Do(req.WithContext(ctx))

	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve resource")
	}

	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	r, err := charset.NewReader(resp.Body, contentType)

	if err != nil {
		return nil, errors.Wrap(err, "could not create UTF-8 charset reader")
	}

	b, err := ioutil.ReadAll(r)

	if err != nil {
		return nil, errors.Wrap(err, "could not read response body")
	}

	return bytes.NewReader(b), nil
}
