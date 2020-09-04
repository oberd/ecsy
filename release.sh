#!/bin/bash

set -e

LAST_TAG=$(git describe --abbrev=0 --tags)
DOCKER_IMAGE_NAME="oberd/ecsy:${LAST_TAG:1}"
export GITHUB_TOKEN="${GITHUB_TOKEN:-$GITHUB_OAUTH_TOKEN}"
docker build -t "$DOCKER_IMAGE_NAME" .
docker push "$DOCKER_IMAGE_NAME"

github-release "oberd/ecsy" "$LAST_TAG" "$(git rev-parse --abbrev-ref HEAD)" "" "dist/ecsy-$LAST_TAG-*"
