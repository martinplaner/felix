# felix

[![Current Release](https://img.shields.io/github/release/martinplaner/felix.svg)](https://github.com/martinplaner/felix/releases/latest)
[![Build Status](https://travis-ci.org/martinplaner/felix.svg?branch=master)](https://travis-ci.org/martinplaner/felix)
[![Go Report Card](https://goreportcard.com/badge/github.com/martinplaner/felix)](https://goreportcard.com/report/github.com/martinplaner/felix)
[![codecov](https://codecov.io/gh/martinplaner/felix/branch/master/graph/badge.svg)](https://codecov.io/gh/martinplaner/felix)
[![GoDoc](https://godoc.org/github.com/martinplaner/felix?status.svg)](https://godoc.org/github.com/martinplaner/felix)
[![License](https://img.shields.io/badge/LICENSE-BSD-ff69b4.svg)](https://github.com/martinplaner/felix/blob/master/LICENSE)

The **fe**ed **li**nk e**x**tractor.

Felix is a tool for fetching, extracting, filtering and aggregating links from multiple feed sources (RSS feeds, HTML pages, etc.).

![overview](doc/overview.png)

## Changelog

See [CHANGELOG](CHANGELOG.md).

For some vague notes about planned features and changes, see also [TODO](TODO.md).

## Installation

### Via go get:

`go get -u github.com/martinplaner/felix`, then run `felix`.

### Via docker:

Automated docker builds are available as [martinplaner/felix](https://hub.docker.com/r/martinplaner/felix/). Pull the image with `docker pull martinplaner/felix`.

The image expects the config file mapped to `/config.yml` and a data volume mounted to `/data`.

## Documentation

Developer documentation can be found on [GoDoc](https://godoc.org/github.com/martinplaner/felix).

User documentation is currently lacking. For the moment have a look at the example config and try to figure it out from there.

## License

Copyright 2017 Martin Planer. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
