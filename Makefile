GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
GIT_COMMIT=$(shell git rev-parse --short HEAD)
LDFLAGS=-ldflags "-s -w -X main.GitBranch=${GIT_BRANCH} -X main.GitCommit=${GIT_COMMIT} -X main.BuildDate=`date -u +%Y-%m-%d.%H:%M:%S`"

build:
	@[ -d .build ] || mkdir -p .build
	CGO_ENABLED=0 go build ${LDFLAGS} -o .build/nats-streaming-proxy cmd/nats-streaming-proxy/main.go
	@file  .build/nats-streaming-proxy
	@du -h .build/nats-streaming-proxy

deb: build
	@nfpm pkg --target .build/nats-streaming-proxy.deb
	@dpkg-deb -I .build/nats-streaming-proxy.deb

PHONY: build