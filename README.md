# 🦊 gitlab-merger - Automated merge request creation for GitLab projects

[![GoDoc](https://godoc.org/github.com/mvisonneau/gitlab-merger?status.svg)](https://godoc.org/github.com/mvisonneau/gitlab-merger)
[![Go Report Card](https://goreportcard.com/badge/github.com/mvisonneau/gitlab-merger)](https://goreportcard.com/report/github.com/mvisonneau/gitlab-merger)
[![Docker Pulls](https://img.shields.io/docker/pulls/mvisonneau/gitlab-merger.svg)](https://hub.docker.com/r/mvisonneau/gitlab-merger/)
[![Build Status](https://cloud.drone.io/api/badges/mvisonneau/gitlab-merger/status.svg)](https://cloud.drone.io/mvisonneau/gitlab-merger)
[![Coverage Status](https://coveralls.io/repos/github/mvisonneau/gitlab-merger/badge.svg?branch=master)](https://coveralls.io/github/mvisonneau/gitlab-merger?branch=master)

`gitlab-merger` allows you to generate merge requests that can be used for release review and notification purposes.

## Usage

```bash
~$ gitlab-merger
NAME:
   gitlab-merger - Automate your MR creation

USAGE:
   main [global options] command [command options] [arguments...]

COMMANDS:
   merge    refs together
   refresh  users list
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --log-level level     log level (debug,info,warn,fatal,panic) (default: "info") [$GLM_LOG_LEVEL]
   --log-format format   log format (json,text) (default: "text") [$GLM_LOG_FORMAT]
   --gitlab-url url      url [$GLM_GITLAB_URL]
   --gitlab-token token  token [$GLM_GITLAB_TOKEN]
   --help, -h            show help
   --version, -v         print the version
```

## Install

Have a look onto the [latest release page](https://github.com/mvisonneau/gitlab-merger/releases/latest) and pick your flavor.

### Go

```bash
~$ go get -u github.com/mvisonneau/gitlab-merger
```

### Homebrew

```bash
~$ brew install mvisonneau/tap/gitlab-merger
```

### Docker

```bash
~$ docker run -it --rm mvisonneau/gitlab-merger
```

### Scoop

```bash
~$ scoop bucket add https://github.com/mvisonneau/scoops
~$ scoop install gitlab-merger
```

### Binaries, DEB and RPM packages

For the following ones, you need to know which version you want to install, to fetch the latest available :

```bash
~$ export GITLAB_MERGER_VERSION=$(curl -s "https://api.github.com/repos/mvisonneau/gitlab-merger/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
```

```bash
# Binary (eg: linux/amd64)
~$ wget https://github.com/mvisonneau/strongbox/releases/download/${GITLAB_MERGER_VERSION}/gitlab-merger_${GITLAB_MERGER_VERSION}_linux_amd64.tar.gz
~$ tar zxvf gitlab-merger_${GITLAB_MERGER_VERSION}_linux_amd64.tar.gz -C /usr/local/bin

# DEB package (eg: linux/386)
~$ wget https://github.com/mvisonneau/strongbox/releases/download/${GITLAB_MERGER_VERSION}/gitlab-merger_${GITLAB_MERGER_VERSION}_linux_386.deb
~$ dpkg -i gitlab-merger_${GITLAB_MERGER_VERSION}_linux_386.deb

# RPM package (eg: linux/arm64)
~$ wget https://github.com/mvisonneau/strongbox/releases/download/${GITLAB_MERGER_VERSION}/gitlab-merger_${GITLAB_MERGER_VERSION}_linux_arm64.rpm
~$ rpm -ivh gitlab-merger_${GITLAB_MERGER_VERSION}_linux_arm64.rpm
```

## Develop / Test

If you use docker, you can easily get started using :

```bash
~$ make dev-env
# You should then be able to use go commands to work onto the project, eg:
~docker$ make fmt
~docker$ gitlab-merger
```

This command will spin up a container with everything required in terms of **golang** dependencies to get started.

## Build / Release

If you want to build and/or release your own version of `gitlab-merger`, you need the following prerequisites :

- [git](https://git-scm.com/)
- [golang](https://golang.org/)
- [make](https://www.gnu.org/software/make/)
- [goreleaser](https://goreleaser.com/)

```bash
~$ git clone git@github.com:mvisonneau/gitlab-merger.git && cd gitlab-merger

# Build the binaries locally
~$ make build

# Build the binaries and release them (you will need a GITHUB_TOKEN and to reconfigure .goreleaser.yml)
~$ make release
```

## Contribute

Contributions are more than welcome! Feel free to submit a [PR](https://github.com/mvisonneau/gitlab-merger/pulls).
