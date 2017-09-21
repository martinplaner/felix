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

var twoLinkTags = `
<html>
	<body>
		<a href="http://example.com">Testlink1</a>
		<div>
			<a href="http://example.org">Testlink2</a>
		</div>

		<p>Hey look what I found on http://example.net.<p>
	</body>
</html>
`

func TestLinkScanner(t *testing.T) {
	testCases := []struct {
		desc          string
		content       string
		expectedLinks int
		shouldError   bool
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
			desc:          "two link tags",
			content:       twoLinkTags,
			expectedLinks: 2,
			shouldError:   false,
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
		})
	}
}
