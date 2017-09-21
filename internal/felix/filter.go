// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package felix

import "unicode"
import "strings"
import "net/url"

type ItemFilter interface {
	Filter(Item)
}

// ItemFilterFunc is an adapter to allow the use of ordinary functions as filters.
// If f is a function with the appropriate signature, ItemFilterFunc(f) is a ItemFilter that calls f.
type ItemFilterFunc func(Item)

func (f ItemFilterFunc) Filter(i Item) {
	f(i)
}

type ItemFilterAdapter func(ItemFilter) ItemFilter

// FilterItems should just filter until in-Channel is closed? Or is quit channel needed?
func FilterItems(in <-chan Item, out chan<- Item, filters ...ItemFilterAdapter) {

	// final filter chain output
	var chain ItemFilter = ItemFilterFunc(func(item Item) {
		out <- item
	})

	// construct filter chain (preserve filter order)
	for i := len(filters) - 1; i >= 0; i-- {
		adapter := filters[i]
		chain = adapter(chain)
	}

	for item := range in {
		chain.Filter(item)
	}

	close(out)
}

// ItemTitleFilter only allows items through where the title matches
// at least one of the given titles. (After conversion to lower case and
// stripping of all non-alphanumeric characters)
func ItemTitleFilter(titles []string) ItemFilterAdapter {

	validTitles := make([][]string, 0, len(titles))
	for _, t := range titles {
		validTitles = append(validTitles, strings.Split(sanitizeTitle(t), " "))
	}

	return func(next ItemFilter) ItemFilter {
		return ItemFilterFunc(func(item Item) {
			itemTitle := sanitizeTitle(item.Title)
			for _, title := range validTitles {
				found := true
				for _, tf := range title {
					if !strings.Contains(tf, itemTitle) {
						found = false
					}
				}
				if found {
					next.Filter(item)
					return
				}
			}
		})
	}
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
	Filter(Link)
}

type LinkFilterFunc func(Link)

func (f LinkFilterFunc) Filter(l Link) {
	f(l)
}

type LinkFilterAdapter func(LinkFilter) LinkFilter

// FilterLinks should just filter until in-Channel is closed? Or is quit channel needed?
func FilterLinks(in <-chan Link, out chan<- Link, filters ...LinkFilterAdapter) {

	// final filter chain output
	var chain LinkFilter = LinkFilterFunc(func(link Link) {
		out <- link
	})

	// construct filter chain
	for i := len(filters) - 1; i >= 0; i-- {
		adapter := filters[i]
		chain = adapter(chain)
	}

	for item := range in {
		chain.Filter(item)
	}

	close(out)
}

func LinkDomainFilter(domains ...string) LinkFilterAdapter {

	validDomains := make([]string, 0, len(domains))
	for _, domain := range domains {
		validDomains = append(validDomains, strings.ToLower(strings.TrimSpace(domain)))
	}

	return func(next LinkFilter) LinkFilter {
		return LinkFilterFunc(func(link Link) {
			u, err := url.Parse(link.URL)

			if err != nil {
				return
			}

			hostname := strings.ToLower(u.Hostname())

			for _, domain := range validDomains {
				if domain == hostname {
					next.Filter(link)
					return
				}
			}
		})
	}
}
