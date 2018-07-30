// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package html

import (
	"context"
	"io"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"

	"github.com/martinplaner/felix/internal/felix"
)

var urlPattern = regexp.MustCompile(`(?mi:https?://[^<"' (\r\n)\t]+)`)

// LinkScanner parses r as an HTML document and extracts all links.
// Links are uniquely identified by the links URL. Multiple instances of the same URL (e.g. href),
// will only be reported once (i.e. the first found instance).
var LinkScanner = felix.ScanFunc(func(ctx context.Context, r io.Reader, e felix.Emitter) error {
	doc, err := goquery.NewDocumentFromReader(r)

	if err != nil {
		return errors.Wrap(err, "could not read HTML document")
	}

	// TODO: only retrieve absolute links with schema?
	foundURLs := make(map[string]bool)

	// Look for proper HTML links in <a> elements
	doc.Find("a").Each(func(index int, item *goquery.Selection) {
		if href, ok := item.Attr("href"); ok && !foundURLs[href] {
			title := item.Text()
			if strings.TrimSpace(title) == "" {
				title = href
			}
			foundURLs[href] = true
			e.EmitLink(felix.Link{
				Title: title,
				URL:   href,
			})
		}
	})

	// Look for other links in complete source via regex
	if s, err := doc.Html(); err == nil {
		urls := urlPattern.FindAllString(s, -1)
		for _, u := range urls {
			if !foundURLs[u] {
				foundURLs[u] = true
				e.EmitLink(felix.Link{
					Title: u,
					URL:   u,
				})
			}
		}
	}

	return nil
})
