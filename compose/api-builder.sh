#!/usr/bin/env bash

if ! command -v dlv &> /dev/null
then
    echo "[builder] delve could not be found, it will be installed"
    go install github.com/go-delve/delve/cmd/dlv@latest
fi

NOW=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
REVISION=$(git rev-parse HEAD)

LDFLAGS=""
LDFLAGS += " -X main.buildType=development"
LDFLAGS += " -X main.buildHost=${BUILD_HOST:-`hostname`}"
LDFLAGS += " -X main.buildTime=${NOW}"
LDFLAGS += " -X main.buildRevision=${REVISION}"

echo "[builder] building monetr now..."
go build -ldflags "${LDFLAGS}" -tags=mini,noui -o /usr/bin/monetr github.com/monetr/monetr/pkg/cmd
