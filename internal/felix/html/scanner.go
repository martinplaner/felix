// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package html

import (
	"context"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"

	"github.com/martinplaner/felix/internal/felix"
)

// LinkScanner parses r as an HTML document and extracts all links.
var LinkScanner = felix.ScanFunc(func(ctx context.Context, r io.Reader, e felix.Emitter) error {
	doc, err := goquery.NewDocumentFromReader(r)

	if err != nil {
		return errors.Wrap(err, "could not read HTML document")
	}

	doc.Find("a").Each(func(index int, item *goquery.Selection) {
		if href, ok := item.Attr("href"); ok {
			title := item.Text()
			if strings.TrimSpace(title) == "" {
				title = href
			}
			e.EmitLink(felix.Link{
				Title: title,
				URL:   href,
			})
		}
	})

	// TODO: Also parse by regex to capture non <a> links? (but maybe in separate package/scanner?)

	return nil
})
