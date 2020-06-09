# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [0ver](https://0ver.org).

## [Unreleased]

### Changed

- Added the link to the MR in both the title and the footer
- Bumped all dependencies
- Bumped goreleaser to 0.137.0

## [0.1.3] - 2019-04-07

### Changed

- Fixed a bug in the fetching of slack users function causing wrong assignments of SlackUserIds
- Reverted the go-gitlab library to the xanzy one
- Bumped to go 1.14
- Use revive for lint tests
- Bumped goreleaser to 0.131.1
- Removed some unnecessary calls to the API in the merge function context
- Enhanced docker build times
- Switched the slack library from nlopes to slack-go (moved)
- Added links to MR in titles of slack messages

## [0.1.2] - 2019-10-23

### Changed

- Re-used fork for xanzy/go-gitlab to fix an issue on the allowed approvers update

## [0.1.1] - 2019-10-23

### Added

- Missed windows builds

### Changed

- Avoid duplicates on not found emails
- Fixed a bug where users get added twice if they commit with different emails and crash the MR creation
- Bumped goreleaser to `v0.131.1`
- Bumped go to `1.13`
- Moved back to upstream repo of go-gitlab (xanzy)
- Fixed docker builds

## [0.1.0] - 2019-09-03

### Added

- Working state of the app
- merge function
- slack notification support
- automated fetching of gitlab and slack users
- Makefile
- LICENSE
- README

[Unreleased]: https://github.com/mvisonneau/gitlab-merger/compare/0.1.3...HEAD
[0.1.3]: https://github.com/mvisonneau/gitlab-merger/tree/0.1.3
[0.1.2]: https://github.com/mvisonneau/gitlab-merger/tree/0.1.2
[0.1.1]: https://github.com/mvisonneau/gitlab-merger/tree/0.1.1
[0.1.0]: https://github.com/mvisonneau/gitlab-merger/tree/0.1.0
