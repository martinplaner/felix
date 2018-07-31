// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package felix

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestFilterItems(t *testing.T) {
	in := make(chan Item, 10)
	out := make(chan Item, 10)

	items := []Item{{Title: "test"}, {Title: "foobar"}}

	go FilterItems(in, out, []ItemFilter{}...)

	for _, item := range items {
		in <- item
	}
	close(in)

	var got []Item
	for item := range out {
		got = append(got, item)
	}

	if len(got) != len(items) {
		t.Errorf("unexpected number of items returned by filter, expected %d, got %d", len(items), len(got))
	}

	if !reflect.DeepEqual(got, items) {
		t.Errorf("unexpected items returned by filter, expected %#v, got %#v", items, got)
	}
}

func Test_buildItemFilterChain(t *testing.T) {
	filterAppendA := ItemFilterFunc(func(item Item, next func(Item)) {
		item.Title += "A"
		next(item)
	})
	filterAppendB := ItemFilterFunc(func(item Item, next func(Item)) {
		item.Title += "B"
		next(item)
	})
	filterAppendC := ItemFilterFunc(func(item Item, next func(Item)) {
		item.Title += "C"
		next(item)
	})

	chain := buildItemFilterChain([]ItemFilter{filterAppendA, filterAppendB, filterAppendC}...)

	in := Item{}
	var got Item
	expectedTitle := "ABC"

	chain.Filter(in, func(item Item) {
		got = item
	})

	if got.Title != expectedTitle {
		t.Errorf("wrong title after filter chain, expected %q, got %q", expectedTitle, got.Title)
	}
}

func runItemFilter(filter ItemFilter, input []Item) []Item {
	output := []Item{}

	var add = func(item Item) {
		output = append(output, item)
	}

	for _, item := range input {
		filter.Filter(item, add)
	}

	return output
}

func TestItemTitleFilter(t *testing.T) {
	testCases := []struct {
		desc     string
		filter   ItemFilter
		input    []Item
		expected []Item
	}{
		{
			desc:     "empty filter criteria",
			filter:   ItemTitleFilter(),
			input:    []Item{{Title: "a title"}, {Title: "another title"}},
			expected: []Item{},
		},
		{
			// TODO: should this even be allowed?
			desc:     "empty string matches everything",
			filter:   ItemTitleFilter(""),
			input:    []Item{{Title: "a title"}, {Title: "another title"}},
			expected: []Item{{Title: "a title"}, {Title: "another title"}},
		},
		{
			desc:     "matching filter",
			filter:   ItemTitleFilter("title", "another"),
			input:    []Item{{Title: "a title"}, {Title: "another title"}},
			expected: []Item{{Title: "a title"}, {Title: "another title"}},
		},
		{
			desc:     "special characters",
			filter:   ItemTitleFilter("A Title & With: Special Characters", "@deutscher titel"),
			input:    []Item{{Title: "A.title.with.special.characters"}, {Title: "Ein deutscher Titel"}, {Title: "Un intitulé"}},
			expected: []Item{{Title: "A.title.with.special.characters"}, {Title: "Ein deutscher Titel"}},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := runItemFilter(tC.filter, tC.input)

			if len(got) != len(tC.expected) {
				t.Errorf("unexpected number of items returned by filter, expected %d, got %d", len(tC.expected), len(got))
			}

			if !reflect.DeepEqual(got, tC.expected) {
				t.Errorf("unexpected items returned by filter, expected %#v, got %#v", tC.expected, got)
			}
		})
	}
}

func TestFilterLinks(t *testing.T) {
	in := make(chan Link, 10)
	out := make(chan Link, 10)

	links := []Link{{Title: "test"}, {Title: "foobar"}}

	go FilterLinks(in, out, []LinkFilter{}...)

	for _, link := range links {
		in <- link
	}
	close(in)

	var got []Link
	for link := range out {
		got = append(got, link)
	}

	if len(got) != len(links) {
		t.Errorf("unexpected number of links returned by filter, expected %d, got %d", len(links), len(got))
	}

	if !reflect.DeepEqual(got, links) {
		t.Errorf("unexpected links returned by filter, expected %#v, got %#v", links, got)
	}
}

func Test_buildLinkFilterChain(t *testing.T) {
	filterAppendA := LinkFilterFunc(func(link Link, next func(Link)) {
		link.Title += "A"
		next(link)
	})
	filterAppendB := LinkFilterFunc(func(link Link, next func(Link)) {
		link.Title += "B"
		next(link)
	})
	filterAppendC := LinkFilterFunc(func(link Link, next func(Link)) {
		link.Title += "C"
		next(link)
	})

	chain := buildLinkFilterChain([]LinkFilter{filterAppendA, filterAppendB, filterAppendC}...)

	in := Link{}
	var got Link
	expectedTitle := "ABC"

	chain.Filter(in, func(link Link) {
		got = link
	})

	if got.Title != expectedTitle {
		t.Errorf("wrong title after filter chain, expected %q, got %q", expectedTitle, got.Title)
	}
}

func runLinkFilter(filter LinkFilter, input []Link) []Link {
	output := []Link{}

	var add = func(link Link) {
		output = append(output, link)
	}

	for _, link := range input {
		filter.Filter(link, add)
	}

	return output
}

func Test_sanitizeTitle(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"   ", ""},
		{"  &  a : ", "a"},
		{"Title", "title"},
		{"another Interesting Title", "another interesting title"},
		{"The   Title (2017)", "the title 2017"},
		{"The TitleRRR", "the titlerrr"},
		{"A Title & With: Special Characters", "a title with special characters"},
		{"@title", "title"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := sanitizeTitle(tt.in); got != tt.want {
				t.Errorf("sanitizeTitle(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestLinkDomainFilter(t *testing.T) {
	testCases := []struct {
		desc     string
		filter   LinkFilter
		input    []Link
		expected []Link
	}{
		{
			desc:     "empty filter criteria",
			filter:   LinkDomainFilter(),
			input:    []Link{{URL: "http://example.com/test"}},
			expected: []Link{},
		},
		{
			desc:     "matching filter",
			filter:   LinkDomainFilter("example.com"),
			input:    []Link{{URL: "http://example.com/test1"}, {URL: "feed://example.com/test2"}, {URL: "http://example.org/testOrg"}},
			expected: []Link{{URL: "http://example.com/test1"}, {URL: "feed://example.com/test2"}},
		},
		{
			desc:     "untrimmed filter criteria",
			filter:   LinkDomainFilter("  example.com     "),
			input:    []Link{{URL: "http://example.com/test1"}, {URL: "http://example.com/test2"}, {URL: "http://example.org/testOrg"}},
			expected: []Link{{URL: "http://example.com/test1"}, {URL: "http://example.com/test2"}},
		},
		{
			desc:     "invalid URLs",
			filter:   LinkDomainFilter("example.com"),
			input:    []Link{{URL: "http://example.com/test1"}, {URL: "example.com/test2"}, {URL: "http:////example.com?/test3"}},
			expected: []Link{{URL: "http://example.com/test1"}},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := runLinkFilter(tC.filter, tC.input)

			if len(got) != len(tC.expected) {
				t.Errorf("unexpected number of links returned by filter, expected %d, got %d", len(tC.expected), len(got))
			}

			if !reflect.DeepEqual(got, tC.expected) {
				t.Errorf("unexpected links returned by filter, expected %#v, got %#v", tC.expected, got)
			}
		})
	}
}

func TestLinkURLRegexFilter(t *testing.T) {

	// must create new filter
	var newFilter = func(exprs ...string) LinkFilter {
		t.Helper()
		f, err := LinkURLRegexFilter(exprs...)
		if err != nil {
			panic(err)
		}
		return f
	}

	testCases := []struct {
		desc     string
		filter   LinkFilter
		input    []Link
		expected []Link
	}{
		{
			desc:     "empty filter criteria",
			filter:   newFilter(),
			input:    []Link{{URL: "http://example.com/test.mp4"}},
			expected: []Link{},
		},
		{
			desc:     "matching filter",
			filter:   newFilter(`.*\.mp4$`, `.*\.mkv$`),
			input:    []Link{{URL: "http://example.com/test.mp4"}, {URL: "http://example.com/test.mkv"}},
			expected: []Link{{URL: "http://example.com/test.mp4"}, {URL: "http://example.com/test.mkv"}},
		},
		{
			desc:     "non-matching filter",
			filter:   newFilter(`.*\.mp4`),
			input:    []Link{{URL: "http://example.com/test.mkv"}},
			expected: []Link{},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := runLinkFilter(tC.filter, tC.input)

			if len(got) != len(tC.expected) {
				t.Errorf("unexpected number of links returned by filter, expected %d, got %d", len(tC.expected), len(got))
			}

			if !reflect.DeepEqual(got, tC.expected) {
				t.Errorf("unexpected links returned by filter, expected %#v, got %#v", tC.expected, got)
			}
		})
	}
}

func TestLinkFilenameAsTitleFilter(t *testing.T) {
	testCases := []struct {
		desc     string
		filter   LinkFilter
		input    []Link
		expected []Link
	}{
		{
			desc:     "valid filename",
			filter:   LinkFilenameAsTitleFilter(false),
			input:    []Link{{Title: "title", URL: "http://example.com/image.jpg"}, {Title: "title", URL: "http://example.com/dl/testfile"}},
			expected: []Link{{Title: "image.jpg", URL: "http://example.com/image.jpg"}, {Title: "testfile", URL: "http://example.com/dl/testfile"}},
		},
		{
			desc:     "strip file extension",
			filter:   LinkFilenameAsTitleFilter(true),
			input:    []Link{{Title: "title", URL: "http://example.com/image.jpg"}, {Title: "title", URL: "http://example.com/dl/testfile"}},
			expected: []Link{{Title: "image", URL: "http://example.com/image.jpg"}, {Title: "testfile", URL: "http://example.com/dl/testfile"}},
		},
		{
			desc:     "empty title and url",
			filter:   LinkFilenameAsTitleFilter(false),
			input:    []Link{{Title: "", URL: ""}},
			expected: []Link{{Title: "", URL: ""}},
		},
		{
			desc:     "empty path in url",
			filter:   LinkFilenameAsTitleFilter(false),
			input:    []Link{{Title: "title", URL: "http://example.com"}, {Title: "title", URL: "http://example.com/"}},
			expected: []Link{{Title: "title", URL: "http://example.com"}, {Title: "title", URL: "http://example.com/"}},
		},
		{
			desc:     "non-empty path but without filename",
			filter:   LinkFilenameAsTitleFilter(false),
			input:    []Link{{Title: "title", URL: "http://example.com/category/announcements/"}, {Title: "title", URL: "http://example.com/news/   "}},
			expected: []Link{{Title: "title", URL: "http://example.com/category/announcements/"}, {Title: "title", URL: "http://example.com/news/   "}},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := runLinkFilter(tC.filter, tC.input)

			if len(got) != len(tC.expected) {
				t.Errorf("unexpected number of links returned by filter, expected %d, got %d", len(tC.expected), len(got))
			}

			if !reflect.DeepEqual(got, tC.expected) {
				t.Errorf("unexpected links returned by filter, expected %#v, got %#v", tC.expected, got)
			}
		})
	}
}

type mockSource func(ctx context.Context, url string) (io.Reader, error)

func (f mockSource) Get(ctx context.Context, url string) (io.Reader, error) {
	return f(ctx, url)
}

func TestLinkUploadedExpandFilenameFilter(t *testing.T) {

	buildFilter := func(filename string, success bool) LinkFilter {
		source := mockSource(func(ctx context.Context, url string) (io.Reader, error) {
			if !success {
				return nil, errors.New("mock source error")
			}

			return strings.NewReader(fmt.Sprintf("%s\n198,90 KB", filename)), nil
		})

		return LinkUploadedExpandFilenameFilter(source)
	}

	testCases := []struct {
		desc     string
		filter   LinkFilter
		input    []Link
		expected []Link
	}{
		{
			desc:     "uploaded.net domain, already expanded form",
			filter:   buildFilter("", false),
			input:    []Link{{Title: "title", URL: "http://uploaded.net/file/xxxxxxxx/file.ext"}},
			expected: []Link{{Title: "title", URL: "http://uploaded.net/file/xxxxxxxx/file.ext"}},
		},
		{
			desc:     "uploaded.net domain, not a file URL",
			filter:   buildFilter("", false),
			input:    []Link{{Title: "title", URL: "http://uploaded.net/some/other/url"}},
			expected: []Link{{Title: "title", URL: "http://uploaded.net/some/other/url"}},
		},
		{
			desc:     "uploaded.net domain, existing file",
			filter:   buildFilter("file.ext", true),
			input:    []Link{{Title: "title", URL: "http://uploaded.net/file/xxxxxxxx"}},
			expected: []Link{{Title: "title", URL: "http://uploaded.net/file/xxxxxxxx/file.ext"}},
		},
		{
			desc:     "ul.to domain, existing file",
			filter:   buildFilter("file.ext", true),
			input:    []Link{{Title: "title", URL: "http://ul.to/file/xxxxxxxx"}},
			expected: []Link{{Title: "title", URL: "http://ul.to/file/xxxxxxxx/file.ext"}},
		},
		{
			desc:     "ul.to domain, existing file, ultra short form",
			filter:   buildFilter("file.ext", true),
			input:    []Link{{Title: "title", URL: "http://ul.to/xxxxxxxx"}},
			expected: []Link{{Title: "title", URL: "http://ul.to/file/xxxxxxxx/file.ext"}},
		},
		{
			desc:     "existing file, trailing slash",
			filter:   buildFilter("file.ext", true),
			input:    []Link{{Title: "title", URL: "http://uploaded.net/file/xxxxxxxx/"}},
			expected: []Link{{Title: "title", URL: "http://uploaded.net/file/xxxxxxxx/file.ext"}},
		},
		{
			desc:     "valid domain, fetch error",
			filter:   buildFilter("", false),
			input:    []Link{{Title: "title", URL: "http://uploaded.net/file/xxxxxxxx"}},
			expected: []Link{},
		},
		{
			desc:     "valid domain, empty filename, PARSING ERROR??",
			filter:   buildFilter("", true),
			input:    []Link{{Title: "title", URL: "http://uploaded.net/file/xxxxxxxx"}},
			expected: []Link{},
		},
		{
			desc:     "invalid domain, should not touch link",
			filter:   buildFilter("", true),
			input:    []Link{{Title: "title", URL: "http://example.com/files/foobar.ext"}},
			expected: []Link{{Title: "title", URL: "http://example.com/files/foobar.ext"}},
		},
		{
			desc:     "invalid domain, fetch error, should not touch link",
			filter:   buildFilter("", false),
			input:    []Link{{Title: "title", URL: "http://sub.example.org/files/foobar.ext"}},
			expected: []Link{{Title: "title", URL: "http://sub.example.org/files/foobar.ext"}},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := runLinkFilter(tC.filter, tC.input)

			if len(got) != len(tC.expected) {
				t.Errorf("unexpected number of links returned by filter, expected %d, got %d", len(tC.expected), len(got))
			}

			if !reflect.DeepEqual(got, tC.expected) {
				t.Errorf("unexpected links returned by filter, expected %#v, got %#v", tC.expected, got)
			}
		})
	}
}

func TestLinkDuplicatesFilter(t *testing.T) {
	testCases := []struct {
		desc     string
		filter   LinkFilter
		input    []Link
		expected []Link
	}{
		{
			desc:     "unique urls",
			filter:   LinkDuplicatesFilter(100),
			input:    []Link{{"", "A"}, {"", "B"}, {"", "C"}},
			expected: []Link{{"", "A"}, {"", "B"}, {"", "C"}},
		},
		{
			desc:     "some duplicate urls with different titles",
			filter:   LinkDuplicatesFilter(100),
			input:    []Link{{"", "A"}, {"", "B"}, {"a", "A"}, {"b", "A"}, {"c", "A"}},
			expected: []Link{{"", "A"}, {"", "B"}},
		},
		{
			desc:     "sliding search window overflow",
			filter:   LinkDuplicatesFilter(1),
			input:    []Link{{"", "A"}, {"", "B"}, {"a", "A"}, {"b", "A"}, {"c", "A"}},
			expected: []Link{{"", "A"}, {"", "B"}, {"a", "A"}},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := runLinkFilter(tC.filter, tC.input)

			if len(got) != len(tC.expected) {
				t.Errorf("unexpected number of links returned by filter, expected %d, got %d", len(tC.expected), len(got))
			}

			if !reflect.DeepEqual(got, tC.expected) {
				t.Errorf("unexpected links returned by filter, expected %#v, got %#v", tC.expected, got)
			}
		})
	}
}
