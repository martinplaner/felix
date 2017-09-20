# felix

[![Current Release](https://img.shields.io/github/release/martinplaner/felix.svg)](https://github.com/martinplaner/felix/releases/latest)
[![Build Status](https://travis-ci.org/martinplaner/felix.svg?branch=master)](https://travis-ci.org/martinplaner/felix)
[![Go Report Card](https://goreportcard.com/badge/github.com/martinplaner/felix)](https://goreportcard.com/report/github.com/martinplaner/felix)
[![GoDoc](https://godoc.org/github.com/martinplaner/felix?status.svg)](https://godoc.org/github.com/martinplaner/felix)
[![License](https://img.shields.io/badge/LICENSE-BSD-ff69b4.svg)](https://github.com/martinplaner/felix/blob/master/LICENSE)

The **fe**ed **li**nk e**x**tractor.

## Installation

### Via go get:

`go get -u github.com/martinplaner/felix`, then run `felix`.

### Via docker:

Automated docker builds are available as [martinplaner/felix](https://hub.docker.com/r/martinplaner/felix/). Pull the image with `docker pull martinplaner/felix`.

The image expects the config file mapped to `/config.yml` and a data volume mounted to `/data`.
