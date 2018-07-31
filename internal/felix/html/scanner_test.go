// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package html

import (
	"context"
	"strings"
	"testing"

	"github.com/martinplaner/felix/internal/felix/mock"
)

var empty = ""

// HTML parser wont throw error even with this garbage!?
// var invalidHTML = `!@#^%$#)~+)_(*$%?><":LPO:öœø}}{:}{:>>?</.})`

var noLinks = `
<html>
<body>
</body>
</html>
`

var validLinks = `
<html>
	<body>
		<a href="http://example.com">Testlink1</a>
		<div>
			<a href="http://example.org">Testlink2</a>
		</div>

		<p>Hey look what I found on http://example.net<p>
		<p>And now again on http://example.org<p>

		<h4 class="links">Test:</h4><pre class="links" id="l123456">http://example.net/file/file.zip</pre><div class="ppsd2h" id="l123456" style="display:none">http://example.net/file/file.zip</div>
	</body>
</html>
`

func TestLinkScanner(t *testing.T) {
	testCases := []struct {
		desc             string
		content          string
		expectedLinks    int
		expectedLinkURLs []string
		shouldError      bool
	}{
		{
			desc:          "empty document",
			content:       empty,
			expectedLinks: 0,
			shouldError:   false,
		},
		// {
		// 	desc:          "invalid HTML",
		// 	content:       invalidHTML,
		// 	expectedLinks: 0,
		// 	shouldError:   true,
		// },
		{
			desc:          "no links",
			content:       noLinks,
			expectedLinks: 0,
			shouldError:   false,
		},
		{
			desc:             "two link tags, one unique text link",
			content:          validLinks,
			expectedLinks:    4,
			shouldError:      false,
			expectedLinkURLs: []string{"http://example.com", "http://example.org", "http://example.net", "http://example.net/file/file.zip"},
		},
	}

	ctx := context.Background()

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			e := &mock.Emitter{}
			r := strings.NewReader(tC.content)
			err := LinkScanner.Scan(ctx, r, e)

			if (err != nil) != tC.shouldError {
				t.Errorf("unexpected error return value! shouldError = %v, got: %v", tC.shouldError, err)
			}

			if len(e.Links) != tC.expectedLinks {
				t.Errorf("invalid number of links found! expected: %v, got: %v", tC.expectedLinks, len(e.Links))
			}

			if len(tC.expectedLinkURLs) > 0 {
				for _, l := range e.Links {
					if !in(l.URL, tC.expectedLinkURLs) {
						t.Errorf("unexpected link URL found!: %v", l.URL)
					}
				}
			}
		})
	}
}

func in(v string, ss []string) bool {
	for _, s := range ss {
		if v == s {
			return true
		}
	}
	return false
}
