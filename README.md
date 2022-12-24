[![Build](https://github.com/fornellas/rrb/workflows/build/badge.svg?branch=main)](https://github.com/fornellas/rrb/actions?query=workflow%3Abuild+branch%3Amain)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](https://github.com/fornellas/rrb/pulls)

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

### Pre-built binary

Check the [latest release](https://github.com/fornellas/rrb/releases) then:

```shell
curl -L https://github.com/fornellas/rrb/releases/download/${release}/rrb-linux-amd64 > rrb
chmod 755 rrb
./rrb -h
```

## Build

```shell
go install github.com/fornellas/rrb@latest
rrb -h
```