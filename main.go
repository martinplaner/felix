// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/martinplaner/felix/internal/felix"
)

var (
	// Version is the version string in compliance with Semantic Versioning 2.x
	Version = "undefined"
	// BuildDate is the date and time of build (UTC)
	BuildDate = "undefined"
	// GitSummary is the output of `output of git describe --tags --dirty --always`
	GitSummary = "undefined"
)

func main() {

	item := &felix.Item{
		Title:   "Title",
		URL:     "http://example.com",
		PubDate: time.Now().Add(-1 * time.Second),
	}

	link := &felix.Link{
		Title: "Title",
		URL:   "http://example.com",
	}

	printVersion()
	fmt.Println("Hello from felix!")
	fmt.Println("Item:", item)
	fmt.Println("Link:", link)
}

func printVersion() {
	fmt.Printf("%s version:%s git:%s build:%s\n", os.Args[0], Version, GitSummary, BuildDate)
}
