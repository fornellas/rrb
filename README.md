[![Latest Release](https://img.shields.io/github/v/release/fornellas/rrb)](https://github.com/fornellas/rrb/releases)
[![Push](https://github.com/fornellas/rrb/actions/workflows/push.yaml/badge.svg)](https://github.com/fornellas/rrb/actions/workflows/push.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/fornellas/rrb)](https://goreportcard.com/report/github.com/fornellas/rrb)
[![Go Reference](https://pkg.go.dev/badge/github.com/fornellas/rrb.svg)](https://pkg.go.dev/github.com/fornellas/rrb)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Buy me a beer: donate](https://img.shields.io/badge/Donate-Buy%20me%20a%20beer-yellow)](https://www.paypal.com/donate?hosted_button_id=AX26JVRT2GS2Q)

# rrb

**Re-Run Build** helps continuously build your projects based on source file changes, much like a CI server. **It works with any project, written in any language**. It helps iterate quick when writing code, by giving fast & reliable build results as you do changes to the source.

## How does it work?

Just call `rrb [your build command]`, that simple! In this example we use it to run our build with the `make lint test build` command:


```
$ rrb make lint test build
> make lint test build
goimports -w -local github.com/fornellas/ .
go generate ./...
go mod tidy
gocyclo -over 10 -avg .
Average: 4.37
golangci-lint run --timeout 10m
go vet ./...
gotest ./... -v -race -cover -count=1
?   	github.com/fornellas/rrb	[no test files]
?   	github.com/fornellas/rrb/cmd	[no test files]
?   	github.com/fornellas/rrb/log	[no test files]
?   	github.com/fornellas/rrb/process	[no test files]
?   	github.com/fornellas/rrb/runner	[no test files]
?   	github.com/fornellas/rrb/watcher	[no test files]
go build .
Success: exit status 0
```

From now onward, just edit the source code with your preferred text editor. **As soon as you save a source file `rrb` automatically runs the build**:

```
Changed: /home/fornellas/src/rrb/runner/runner.go (WRITE)
Changed: /home/fornellas/src/rrb/runner/runner.go (WRITE)
Changed: /home/fornellas/src/rrb/runner/runner.go (WRITE)
> make lint test build
goimports -w -local github.com/fornellas/ .
go generate ./...
go mod tidy
gocyclo -over 10 -avg .
Average: 4.37
golangci-lint run --timeout 10m
go vet ./...
gotest ./... -v -race -cover -count=1
?   	github.com/fornellas/rrb	[no test files]
?   	github.com/fornellas/rrb/cmd	[no test files]
?   	github.com/fornellas/rrb/log	[no test files]
?   	github.com/fornellas/rrb/process	[no test files]
?   	github.com/fornellas/rrb/runner	[no test files]
?   	github.com/fornellas/rrb/watcher	[no test files]
go build .
Success: exit status 0
```

But what happens if while the build is running, a source file changes? This build will yield stale / broken results, so there's no point in waiting for it to finish. In this case, `rrb` **terminates the stale in-flight build** before starting a new build:

```
Changed: /home/fornellas/src/rrb/runner/runner.go (WRITE)
Changed: /home/fornellas/src/rrb/runner/runner.go (WRITE)
Changed: /home/fornellas/src/rrb/runner/runner.go (WRITE)
Killing...
make: *** [Makefile:133: golangci-lint] Terminated
Failure: signal: terminated
> make lint test build
goimports -w -local github.com/fornellas/ .
go generate ./...
go mod tidy
gocyclo -over 10 -avg .
Average: 4.37
golangci-lint run --timeout 10m
go vet ./...
gotest ./... -v -race -cover -count=1
?   	github.com/fornellas/rrb	[no test files]
?   	github.com/fornellas/rrb/cmd	[no test files]
?   	github.com/fornellas/rrb/log	[no test files]
?   	github.com/fornellas/rrb/process	[no test files]
?   	github.com/fornellas/rrb/runner	[no test files]
?   	github.com/fornellas/rrb/watcher	[no test files]
go build .
Success: exit status 0

```

`rrb` does this *deterministically*, ensuring that there are **no stray process from a stale build left behind**[^1] before starting a new build. No matter how often you save your files, **you will *never* get a stale build signal with `rrb`**. This sets `rrb` apart from most similar tools.

[^1]: This is achieved by use of `PR_SET_CHILD_SUBREAPER` (see [prctl(2) ](https://man7.org/linux/man-pages/man2/prctl.2.html)).

## Install

Pick the [latest release](https://github.com/fornellas/rrb/releases) with:

```bash
GOARCH=$(case $(uname -m) in i[23456]86) echo 386;; x86_64) echo amd64;; armv6l|armv7l) echo arm;; aarch64) echo arm64;; *) echo Unknown machine $(uname -m) 1>&2 ; exit 1 ;; esac) && wget -O- https://github.com/fornellas/rrb/releases/latest/download/rrb.linux.$GOARCH.gz | gunzip > rrb && chmod 755 rrb
./rrb --help
```

## Development

[Docker](https://www.docker.com/) is used to create a reproducible development environment on any machine:

```bash
git clone git@github.com:fornellas/rrb.git
cd rrb/
./builld.sh
```

Typically you'll want to stick to `./builld.sh rrb`, as it enables you to edit files as preferred, and the build will automatically be triggered on any file changes.