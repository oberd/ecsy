all: release-deps deps install release

clean:
	@rm -rf dist && mkdir dist

compile: clean
	GOOS=darwin go build -o dist/ecsy *.go

release-deps:
	go get github.com/c4milo/github-release

deps:
	go list -f '{{join .Deps "\n"}}' | xargs go list -e -f '{{if not .Standard}}{{.ImportPath}}{{end}}' | xargs go get -u

install:
	go install -ldflags "-X main.Version=v1.0.8"

release: compile
	@./release.sh
