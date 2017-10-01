// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package felix

import (
	"fmt"
	"net/http"
	"text/template"
	"time"
)

type Feed struct {
	PubDate time.Time
	Items   []Item
}

func StringHandler(s string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, s)
	})
}

func FeedHandler(ds Datastore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// TODO: make configurable
		links, err := ds.GetLinks(6 * time.Hour)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var items []Item
		for _, link := range links {
			items = append(items, Item{
				Title:   link.Title,
				URL:     link.URL,
				PubDate: time.Now(),
			})
		}

		feed := &Feed{
			PubDate: time.Now(),
			Items:   items,
		}

		w.Header().Set("Content-Type", "application/rss+xml")
		if err := feedTemplate.Execute(w, feed); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

var feedTemplate = template.Must(template.New("feed").Funcs(funcMap).Parse(templateString))

var funcMap = template.FuncMap{
	"formatTime": func(t time.Time) string { return t.Format("Mon, 02 Jan 2006 15:04:05 -0700") },
}

var templateString = `<?xml version="1.0" encoding="UTF-8"?><rss version="2.0">
<channel>
	<title>felix</title>
	<description>felix feed</description>
	<link>http://example.com</link>
	<pubDate>{{.PubDate | formatTime}}</pubDate>
	{{range .Items}}
	<item>
		<title>{{.Title}}</title>
		<guid>{{.URL}}</guid>
		<link>{{.URL}}</link>
		<description>{{.Title}}</description>
		<pubDate>{{.PubDate | formatTime}}</pubDate>
	</item>
	{{end}}
</channel>
</rss>
`
