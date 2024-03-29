LAST_TAG := $(shell git describe --abbrev=0 --tags)
all: release-deps install release

clean:
	@rm -rf dist && mkdir dist

compile: clean
	GOOS=darwin go build -o dist/ecsy-$(LAST_TAG)-darwin-amd64 *.go && \
	GOOS=linux CGO_ENABLED=0 go build -o dist/ecsy-$(LAST_TAG)-linux *.go && \
	GOOS=darwin GOARCH=arm64 go build -o dist/ecsy-$(LAST_TAG)-darwin-arm64 *.go

release-deps:
	go install github.com/c4milo/github-release

install:
	go install -ldflags "-X main.Version=v1.0.8"

release: compile
	@./release.sh
