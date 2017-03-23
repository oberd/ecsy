#!/bin/bash

LAST_TAG=$(git describe --abbrev=0 --tags)
github-release "oberd/ecsy" "$LAST_TAG" "$(git rev-parse --abbrev-ref HEAD)" "" "dist/ecsy-$LAST_TAG-darwin-amd64"
