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

// ItemTitleFilter only accepts items where the title matches
// at least one of the given titles. (After conversion to lower case and
// stripping of all non-alphanumeric characters)
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

type LinkFilter interface {
	Filter(link Link, next func(Link))
}

type LinkFilterFunc func(Link, func(Link))

func (f LinkFilterFunc) Filter(link Link, next func(Link)) {
	f(link, next)
}

type LinkFilterAdapter func(LinkFilter) LinkFilter

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
			if expr.MatchString(link.URL) {
				next(link)
				break
			}
		}
	}), nil
}
