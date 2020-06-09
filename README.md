# 🦊 gitlab-merger - Automated merge request creation for GitLab projects

[![GoDoc](https://godoc.org/github.com/mvisonneau/gitlab-merger?status.svg)](https://godoc.org/github.com/mvisonneau/gitlab-merger)
[![Go Report Card](https://goreportcard.com/badge/github.com/mvisonneau/gitlab-merger)](https://goreportcard.com/report/github.com/mvisonneau/gitlab-merger)
[![Docker Pulls](https://img.shields.io/docker/pulls/mvisonneau/gitlab-merger.svg)](https://hub.docker.com/r/mvisonneau/gitlab-merger/)
[![Build Status](https://cloud.drone.io/api/badges/mvisonneau/gitlab-merger/status.svg)](https://cloud.drone.io/mvisonneau/gitlab-merger)
[![Coverage Status](https://coveralls.io/repos/github/mvisonneau/gitlab-merger/badge.svg?branch=master)](https://coveralls.io/github/mvisonneau/gitlab-merger?branch=master)

`gitlab-merger` allows you to generate merge requests that can be used for release review and notification purposes.

## Disclaimer

The current implementation of this project is very opinionated but I am very open to make it more generic. It also only runs on GitLab Enterprise with a license >= starter/bronze.

## TL;DR

```bash
# Trigger it from wherever you want, I personally have a manual CI job at the end of a dev branch which allow me to trigger
# dev -> master merge requests

gitlab-merger \
  --log-level debug \
  --gitlab-url https://gitlab.com \
  --gitlab-token xxxx \
  merge \
  --source-ref foo \
  --target-ref bar \
  --project 'foo/bar' \
  --slack-channel '#random' \
  --slack-token 'xoxp-xxx-xxxx-xxxx-xxxxxxx'
INFO[2020-06-09T16:47:55+02:00] checking existing merge requests..
INFO[2020-06-09T16:47:55+02:00] no open MR found, comparing refs..
INFO[2020-06-09T16:47:56+02:00] found differences between refs                commit-count=1
DEBU[2020-06-09T16:47:56+02:00] Found current user ID: 1
DEBU[2020-06-09T16:47:56+02:00] found existing snippet                        snippet-id=1
DEBU[2020-06-09T16:47:56+02:00] loading content from snippet                  snippet-id=1
INFO[2020-06-09T16:47:56+02:00] done matching committers in GitLab            committer-count=1
INFO[2020-06-09T16:47:56+02:00] found GitLab user                             email=maxime@sphere.me gitlab-user-id=1
INFO[2020-06-09T16:47:57+02:00] merge request created                         merge-request-id=1 merge-request-url="https://gitlab.com/foo/bar/-/merge_requests/1"
INFO[2020-06-09T16:47:57+02:00] notifying slack channel about the new MR      slack-channel="#test"
DEBU[2020-06-09T16:47:57+02:00] exiting..                                     execution-duration=2.45031955s
```

Every contributor gets notified in the Slack channel:

![slack-message](docs/images/slack-message.png)

And here is the MR you would get

![gitlab-mr](docs/images/gitlab-mr.png)

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
~$ wget https://github.com/mvisonneau/gitlab-merger/releases/download/${GITLAB_MERGER_VERSION}/gitlab-merger_${GITLAB_MERGER_VERSION}_linux_amd64.tar.gz
~$ tar zxvf gitlab-merger_${GITLAB_MERGER_VERSION}_linux_amd64.tar.gz -C /usr/local/bin

# DEB package (eg: linux/386)
~$ wget https://github.com/mvisonneau/gitlab-merger/releases/download/${GITLAB_MERGER_VERSION}/gitlab-merger_${GITLAB_MERGER_VERSION}_linux_386.deb
~$ dpkg -i gitlab-merger_${GITLAB_MERGER_VERSION}_linux_386.deb

# RPM package (eg: linux/arm64)
~$ wget https://github.com/mvisonneau/gitlab-merger/releases/download/${GITLAB_MERGER_VERSION}/gitlab-merger_${GITLAB_MERGER_VERSION}_linux_arm64.rpm
~$ rpm -ivh gitlab-merger_${GITLAB_MERGER_VERSION}_linux_arm64.rpm
```

## Usage

```bash
~$ gitlab-merger
NAME:
   gitlab-merger - Automate your MR creation

USAGE:
   gitlab-merger [global options] command [command options] [arguments...]

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

### Refresh the GitLab users email mappings

In order to figure out which commit is being made by which GitLab user and get this data in a efficient fashion, we store this value into a snippet. It will be automatically created if it does not exists but you can manually refresh it using the following command

```bash
gitlab-merger refresh gitlab-users
```

### Refresh the GitLab users / Slack users mappings

In order to map out which GitLab users with Slack users, you can use the following command:

```bash
gitlab-merger refresh slack-users
```

It will store the mapping under the same snippet as the email/gitlab mappings

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
