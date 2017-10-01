// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/martinplaner/felix/internal/felix"
	"github.com/martinplaner/felix/internal/felix/bolt"
	"github.com/martinplaner/felix/internal/felix/html"
	"github.com/martinplaner/felix/internal/felix/rss"
	"golang.org/x/net/context"
)

var (
	// Version is the version string in compliance with Semantic Versioning 2.x
	Version = "undefined"
	// BuildDate is the date and time of build (UTC)
	BuildDate = "undefined"
	// GitSummary is the output of `output of git describe --tags --dirty --always`
	GitSummary = "undefined"
)

var (
	log           = felix.NewLogger()
	source        = felix.NewHTTPSource(http.DefaultClient)
	newItems      = make(chan felix.Item)
	newLinks      = make(chan felix.Link)
	filteredItems = make(chan felix.Item)
	filteredLinks = make(chan felix.Link)

	// TODO: remove this hack
	titles []string
)

func main() {
	printVersion()

	configfile := flag.String("config", "config.yml", "location of the config file")
	datadir := flag.String("datadir", ".", "dir for auxilliary data")
	flag.Parse()

	// Initialize config and shared components

	config, err := felix.ConfigFromFile(*configfile)
	if err != nil {
		log.Fatal("could not read config file", "configfile", *configfile)
	}
	log.Info("read config from file.", "configfile", *configfile)

	datastorefile := filepath.Join(*datadir, "felix.db")
	db, err := bolt.NewDatastore(datastorefile)
	if err != nil {
		log.Fatal("could not create datastore", "file", datastorefile)
	} else {
		log.Info("initialized datastore", "datastorefile", datastorefile)
	}
	defer db.Close()

	// Configure fetchers and filters

	feedFetchers := initFeedFetchers(config, db)
	itemFilters := initItemFilters(config)
	linkFilters := initLinkFilters(config)

	quit := make(chan struct{})
	var wgFeeds sync.WaitGroup
	var wgItems sync.WaitGroup

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Start up components

	for _, f := range feedFetchers {
		log.Info("starting new feed fetcher")
		wgFeeds.Add(1)
		go func(f *felix.Fetcher) {
			f.Start(quit)
			wgFeeds.Done()
		}(f)
	}

	go felix.FilterItems(newItems, filteredItems, itemFilters...)
	go runPageFetchers(config, db, &wgItems)
	go felix.FilterLinks(newLinks, filteredLinks, linkFilters...)

	wgFeeds.Add(1)
	go func() {
		periodicCleanup(db, config.CleanupInterval, config.CleanupMaxAge, quit)
		wgFeeds.Done()
	}()

	http.Handle("/", felix.FeedHandler(db))
	http.Handle("/filters", felix.StringHandler(felix.FilterString(itemFilters, linkFilters)))

	// TODO: make host and port configurable
	server := &http.Server{Addr: fmt.Sprintf("%s:%d", "", 6554)}

	go func() {
		log.Info("starting http server", "addr", server.Addr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Error("http server error", "err", err)
		}
	}()

	// Shutdown handler
	go func() {
		<-sig
		log.Info("got shutdown signal, finishing up")
		close(quit)
		wgFeeds.Wait()
		close(newItems)
		wgItems.Wait()
		close(newLinks)
	}()

	// Insert filtered links into datastore
	for link := range filteredLinks {
		log.Info("found new link", "url", link.URL)
		if _, err := db.StoreLink(link); err != nil {
			log.Error("could not store link", "err", err, "link", link)
		}
	}

	if err := server.Shutdown(context.Background()); err != nil {
		log.Error("could not shutdown http server", "err", err)
	} else {
		log.Info("stopped http server")
	}

	log.Info("shutdown complete")
}

// run periodic datastore cleanup routine to remove entries older than maxAge
func periodicCleanup(db felix.Datastore, interval time.Duration, maxAge time.Duration, quit <-chan struct{}) {
L:
	for {
		select {
		case <-time.After(interval):
			if err := db.Cleanup(maxAge); err != nil {
				log.Error("could not cleanup datastore", "err", err)
			}
		case <-quit:
			break L
		}
	}
}

func initFeedFetchers(config felix.Config, data felix.Datastore) []*felix.Fetcher {
	var feedFetchers []*felix.Fetcher
	for _, fc := range config.Feeds {
		switch fc.Type {

		case "rss":
			fetchInterval := fc.FetchInterval
			if fetchInterval == 0 {
				fetchInterval = config.FetchInterval
			}

			nextFetch := felix.PeriodicAttempter(data, fetchInterval)
			f := felix.NewFetcher(fc.URL, source, rss.ItemScanner, nextFetch, newItems, newLinks)
			f.SetLogger(log)
			feedFetchers = append(feedFetchers, f)

		default:
			log.Fatal("unknown feed type", "type", fc.Type)
		}
	}
	return feedFetchers
}

func initItemFilters(config felix.Config) []felix.ItemFilter {
	var itemFilters []felix.ItemFilter
	for _, f := range config.ItemFilters {
		switch f.Type {

		case "title":
			var fc felix.ItemTitleFilterConfig
			if err := f.Unmarshal(&fc); err != nil {
				log.Fatal("could not decode filter config", "err", err, "type", f.Type)
			}

			itemFilters = append(itemFilters, felix.ItemTitleFilter(fc.Titles...))
			titles = fc.Titles

		default:
			log.Fatal("unsupported item filter type", "type", f.Type)
		}
	}
	return itemFilters
}

func initLinkFilters(config felix.Config) []felix.LinkFilter {
	var linkFilters []felix.LinkFilter
	for _, f := range config.LinkFilters {
		switch f.Type {

		case "domain":
			var fc felix.LinkDomainFilterConfig
			if err := f.Unmarshal(&fc); err != nil {
				log.Fatal("could not decode filter config", "err", err, "type", f.Type)
			}

			linkFilters = append(linkFilters, felix.LinkDomainFilter(fc.Domains...))

		case "regex":
			var fc felix.LinkURLRegexFilterConfig
			if err := f.Unmarshal(&fc); err != nil {
				log.Fatal("could not decode filter config", "err", err, "type", f.Type)
			}

			lurf, err := felix.LinkURLRegexFilter(fc.Exprs...)
			if err != nil {
				log.Fatal("could not create filter", "err", err, "type", f.Type)
			}

			linkFilters = append(linkFilters, lurf)

		case "filenameastitle":
			var fc felix.LinkFilenameAsTitleFilterConfig
			if err := f.Unmarshal(&fc); err != nil {
				log.Fatal("could not decode filter config", "err", err, "type", f.Type)
			}

			linkFilters = append(linkFilters, felix.LinkFilenameAsTitleFilter(fc.TrimExt))

		default:
			log.Fatal("unsupported link filter type", "type", f.Type)
		}
	}
	return linkFilters
}

// runPageFetchers restarts old page fetchers found in the datastore
// as well as new ones on demand when new items are found
func runPageFetchers(config felix.Config, db felix.Datastore, wg *sync.WaitGroup) {
	// TODO: refactor this mess.. erm.. component -.-
	quit := make(chan struct{})
	// Restore old item fetchers / scrapers
	oldItems, err := db.GetItems(config.CleanupMaxAge)
	if err != nil {
		log.Error("could not get items", "err", err, "cleanupMaxAge", config.CleanupMaxAge)
	} else {
		for _, item := range oldItems {
			// TODO: Make maxTries configurable
			log.Info("restarting item fetcher", "url", item.URL)
			nextFetch := felix.FibAttempter(db, config.FetchInterval, 7)
			f := felix.NewFetcher(item.URL, source, html.LinkScanner, nextFetch, newItems, newLinks)
			f.SetLogger(log)

			wg.Add(1)
			go func() {
				f.Start(quit)
				wg.Done()
			}()
		}
	}

	for item := range filteredItems {
		didExist, err := db.StoreItem(item)

		if err != nil {
			log.Error("could not store item", "err", err, "item", item)
			continue
		}

		if didExist {
			// Item did already exist. Skipping...
			continue
		}

		log.Info("starting new item fetcher", "url", item.URL)
		nextFetch := felix.FibAttempter(db, config.FetchInterval, 7)
		f := felix.NewFetcher(item.URL, source, html.LinkScanner, nextFetch, newItems, newLinks)
		f.SetLogger(log)

		wg.Add(1)
		go func() {
			f.Start(quit)
			wg.Done()
		}()
	}

	close(quit)
}

func printVersion() {
	fmt.Printf("%s version:%s git:%s build:%s\n", os.Args[0], Version, GitSummary, BuildDate)
}
