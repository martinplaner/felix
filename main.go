// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	mux := http.NewServeMux()

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello from felix web!")
	}))

	srv := &http.Server{Addr: ":6554", Handler: mux}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	go func() {
	L:
		for {
			select {
			case <-time.After(2 * time.Second):
				fmt.Println("Hello (again) from felix!")
			case <-stopChan:
				break L
			}
		}
	}()

	<-stopChan // wait for SIGINT
	log.Println("Shutting down server...")

	// shut down gracefully, but wait no longer than 5 seconds before halting
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	cancel()

	log.Println("Server gracefully stopped")
}

func printVersion() {
	fmt.Printf("%s version:%s git:%s build:%s\n", os.Args[0], Version, GitSummary, BuildDate)
}
