#!/usr/bin/env bash
set -e

echo "DOING DOCKER LOGIN"
if ! echo "$GITHUB_TOKEN" | docker login -u docker --password-stdin ghcr.io ; then
  exit 1
fi

echo "CONTINUE WITH GORELEASER"
exec goreleaser "$@"
