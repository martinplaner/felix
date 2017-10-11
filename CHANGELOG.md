# Change Log
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

Currently no changes or additions.

### Added

- Made `FeedOutputMaxAge`, the maximum age for links included in the output feed, configurable.

### Fixed

- Only report new links the first time they are found.

## [0.2.0] - 2017-10-01

### Added

- Added new filter (LinkFilenameAsTitleFilter) that tries to extract the filename from the URL and sets the link title accordingly.

## [0.1.0] - 2017-10-01

This is the first working release.

### Added

- Periodic fetching and parsing of RSS feeds.
- Fetching of links from HTML pages with fibonacci based backoff.
- Item filters based on item titles.
- Link filters based on link URL, according to domain and/or regular expression.
- Dynamic feed and filter configuration with YAML config file.
- Graceful shutdown and persistence of fetch status across restarts.
- Travis CI pipeline and automated Docker builds.


[Unreleased]: https://github.com/martinplaner/felix/tree/develop
[0.1.0]: https://github.com/martinplaner/felix/releases/tag/v0.1.0
[0.2.0]: https://github.com/martinplaner/felix/releases/tag/v0.2.0
