# Change Log
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

Currently no changes or additions.

## [0.5.0] - 2018-07-31

### Added

- Implemented additional HTML link scanning via regex for text-based links.
- Added LinkDuplicatesFilter to filter duplicate links using a configurable sliding comparison window.

## [0.4.0] - 2018-07-19

### Added

- LinkUploadedExpandFilenameFilter ("expanduploadedlinks") for resolving short form uploaded links to long form including file name.

### Fixed

- Added TLS certificates to Docker image for HTTPS support.

### Changed

- Updated library dependencies.

## [0.3.2] - 2018-05-16

### Fixed

- Fixed an inconsistency with os signals blocking on channels.

### Changed

- Moved to [dep](https://github.com/golang/dep) for dependency management.
- Updated README with binary releases info

## [0.3.1] - 2017-10-13

### Fixed

- Fixed graceful shutdown on http port error.
- Linter and vet recommendations.

## [0.3.0] - 2017-10-12

### Added

- Made some previously hard-coded values configurable:
    - `FeedOutputMaxAge`, the maximum age for links included in the output feed (default: 6h).
    - `Port`, the TCP port the output feed should listen on (default: 6554).
- Added overview diagram to README.md.
- Automated binary releases (linux x86_64 only for now).

### Fixed

- Only report new links the first time they are found.
- Do not upload coverage for tag CI builds.

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
[0.3.0]: https://github.com/martinplaner/felix/releases/tag/v0.3.0
[0.3.1]: https://github.com/martinplaner/felix/releases/tag/v0.3.1
[0.3.2]: https://github.com/martinplaner/felix/releases/tag/v0.3.2
[0.4.0]: https://github.com/martinplaner/felix/releases/tag/v0.4.0
[0.5.0]: https://github.com/martinplaner/felix/releases/tag/v0.5.0
