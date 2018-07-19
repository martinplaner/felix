// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package felix

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"path/filepath"

	"context"

	"time"

	"io"

	"io/ioutil"

	"github.com/pkg/errors"
)

// Stringer is an optional interface for all filters to provide a 'native' textual representation.
type Stringer interface {
	String() string
}

// FilterString return the concatenated string output of all passed filters that implement the Stringer interface.
func FilterString(itemFilters []ItemFilter, linkFilters []LinkFilter) string {
	var b bytes.Buffer

	for _, f := range itemFilters {
		if s, ok := f.(Stringer); ok {
			b.WriteString(s.String())
		}
	}

	for _, f := range linkFilters {
		if s, ok := f.(Stringer); ok {
			b.WriteString(s.String())
		}
	}

	return b.String()
}

// ItemFilter wraps the Filter method for items.
//
// Filter evaluates the given item, optionally modifies it, and passes it
// to the next filter in the filter chain, if it matches the filter criteria.
type ItemFilter interface {
	Filter(item Item, next func(Item))
}

// ItemFilterFunc is an adapter to allow the use of ordinary functions as filters.
// If f is a function with the appropriate signature, ItemFilterFunc(f) is a ItemFilter that calls f.
type ItemFilterFunc func(Item, func(Item))

// Filter calls the underlying ItemFilterFunc
func (f ItemFilterFunc) Filter(item Item, next func(Item)) {
	f(item, next)
}

// internal helper type to provide an additional Stringer implementation
type itemFilter struct {
	ItemFilter
	s string
}

func (f itemFilter) String() string {
	return f.s
}

// FilterItems should just filter until in-Channel is closed? Or is quit channel needed?
func FilterItems(in <-chan Item, out chan<- Item, filters ...ItemFilter) {

	// final filter chain output
	var final = func(item Item) {
		out <- item
	}

	chain := buildItemFilterChain(filters...)

	for item := range in {
		chain.Filter(item, final)
	}

	close(out)
}

func buildItemFilterChain(filters ...ItemFilter) ItemFilter {
	if len(filters) > 1 {
		return ItemFilterFunc(func(item Item, next func(Item)) {
			filters[0].Filter(item, func(item Item) {
				buildItemFilterChain(filters[1:]...).Filter(item, next)
			})
		})
	} else if len(filters) == 1 {
		return filters[0]
	} else {
		return ItemFilterFunc(func(item Item, next func(Item)) {
			next(item)
		})
	}
}

// ItemTitleFilter filters items based on the given title strings.
// (After conversion to lower case and stripping of all non-alphanumeric characters)
func ItemTitleFilter(titles ...string) ItemFilter {

	validTitles := make([][]string, 0, len(titles))
	var b bytes.Buffer
	for _, t := range titles {
		validTitles = append(validTitles, strings.Split(sanitizeTitle(t), " "))
		fmt.Fprintf(&b, "ITEM_TITLE:%s\n", t)
	}

	filter := ItemFilterFunc(func(item Item, next func(Item)) {
		itemTitle := sanitizeTitle(item.Title)
		for _, title := range validTitles {
			found := true
			for _, tf := range title {
				if !strings.Contains(itemTitle, tf) {
					found = false
				}
			}
			if found {
				next(item)
				return
			}
		}
	})

	return itemFilter{filter, b.String()}
}

// sanitizeTitle strips all non-alphanumeric characters from a string
// and converts it to lower case for easier comparison.
func sanitizeTitle(title string) string {

	t := make([]rune, 0, len(title))
	emitted := false
	skipped := false

	for _, r := range title {
		if unicode.IsDigit(r) || unicode.IsLetter(r) {
			if skipped && emitted {
				t = append(t, ' ')
			}
			t = append(t, unicode.ToLower(r))
			emitted = true
			skipped = false
		} else {
			skipped = true
		}
	}

	return string(t)
}

// LinkFilter wraps the Filter method for links.
//
// Filter evaluates the given link, optionally modifies it, and passes it
// to the next filter in the filter chain, if it matches the filter criteria.
type LinkFilter interface {
	Filter(link Link, next func(Link))
}

// LinkFilterFunc is an adapter to allow the use of ordinary functions as filters.
// If f is a function with the appropriate signature, LinkFilterFunc(f) is a LinkFilter that calls f.
type LinkFilterFunc func(Link, func(Link))

// Filter calls the underlying LinkFilterFunc and implements LinkFilter.
func (f LinkFilterFunc) Filter(link Link, next func(Link)) {
	f(link, next)
}

// FilterLinks should just filter until in-Channel is closed? Or is quit channel needed?
func FilterLinks(in <-chan Link, out chan<- Link, filters ...LinkFilter) {

	// final filter chain output
	var final = func(link Link) {
		out <- link
	}

	chain := buildLinkFilterChain(filters...)

	for link := range in {
		chain.Filter(link, final)
	}

	close(out)
}

func buildLinkFilterChain(filters ...LinkFilter) LinkFilter {
	if len(filters) > 1 {
		return LinkFilterFunc(func(link Link, next func(Link)) {
			filters[0].Filter(link, func(link Link) {
				buildLinkFilterChain(filters[1:]...).Filter(link, next)
			})
		})
	} else if len(filters) == 1 {
		return filters[0]
	} else {
		return LinkFilterFunc(func(link Link, next func(Link)) {
			next(link)
		})
	}
}

// LinkDomainFilter filters links based on the given domains.
func LinkDomainFilter(domains ...string) LinkFilter {

	validDomains := make([]string, 0, len(domains))
	for _, domain := range domains {
		validDomains = append(validDomains, strings.ToLower(strings.TrimSpace(domain)))
	}

	return LinkFilterFunc(func(link Link, next func(Link)) {
		u, err := url.Parse(link.URL)

		if err != nil {
			return
		}

		hostname := strings.ToLower(u.Hostname())

		for _, domain := range validDomains {
			if domain == hostname {
				next(link)
				return
			}
		}
	})
}

// LinkURLRegexFilter filters links based their URLs matching the given regular expressions.
func LinkURLRegexFilter(exprs ...string) (LinkFilter, error) {

	var regexes []*regexp.Regexp
	for _, expr := range exprs {
		regex, err := regexp.Compile(expr)
		if err != nil {
			return nil, errors.Wrap(err, "could not compile regular expression")
		}
		regexes = append(regexes, regex)
	}

	return LinkFilterFunc(func(link Link, next func(Link)) {
		for _, expr := range regexes {
			if expr.MatchString(strings.TrimSpace(link.URL)) {
				next(link)
				break
			}
		}
	}), nil
}

// LinkFilenameAsTitleFilter extracts the filename from the URL and sets it as the new link title.
// When trimExt is set, the filter tries to remove the file extension, if one is present.
func LinkFilenameAsTitleFilter(trimExt bool) LinkFilter {
	return LinkFilterFunc(func(link Link, next func(Link)) {
		u, err := url.Parse(strings.TrimSpace(link.URL))

		if err != nil {
			next(link)
			return
		}

		if strings.HasSuffix(u.Path, "/") {
			next(link)
			return
		}

		filename := filepath.Base(u.Path)

		if filename == "." || strings.Contains(filename, "/") {
			next(link)
			return
		}

		if trimExt {
			filename = strings.TrimSuffix(filename, filepath.Ext(filename))
		}

		link.Title = filename
		next(link)
	})
}

// LinkUploadedExpandFilenameFilter expands the filename from an uploaded file URL and sets the appropriate new URL,
// e.g. uploaded.net/file/xxxxxxxx -> uploaded.net/file/xxxxxxxx/file.ext.
// This is sometime needed for easier filtering down the filter chain.
func LinkUploadedExpandFilenameFilter(source Source) LinkFilter {

	return LinkFilterFunc(func(link Link, next func(Link)) {
		u, err := url.Parse(strings.TrimSpace(link.URL))

		// TODO: accept or reject non-parsable URLs?
		if err != nil {
			return
		}

		// Only process "uploaded" domains
		if u.Hostname() != "ul.to" && u.Hostname() != "uploaded.net" {
			next(link)
			return
		}

		pathSegments := strings.Split(strings.Trim(u.Path, "/"), "/")

		// Only process short form file URLs
		if len(pathSegments) < 1 || len(pathSegments) > 2 {
			next(link)
			return
		}

		var id string
		if len(pathSegments) == 1 {
			id = pathSegments[0]
		} else if len(pathSegments) == 2 && pathSegments[0] == "file" {
			id = pathSegments[1]
		} else {
			next(link)
			return
		}

		statusURL := fmt.Sprintf("%s://%s/file/%s/status", u.Scheme, u.Hostname(), id)
		// TODO: pass context to filters?
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		reader, err := source.Get(ctx, statusURL)

		if err != nil {
			// fetch failed, skip link
			return
		}

		filename := parseULFilename(reader)
		if filename == "" {
			return
		}

		link.URL = fmt.Sprintf("%s://%s/file/%s/%s", u.Scheme, u.Hostname(), id, filename)
		next(link)
	})
}

func parseULFilename(r io.Reader) string {
	s, err := ioutil.ReadAll(r)

	if err != nil {
		return ""
	}

	split := strings.Split(string(s), "\n")
	if len(split) < 1 {
		return ""
	}

	return split[0]
}
