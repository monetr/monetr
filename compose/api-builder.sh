#!/usr/bin/env bash

NOW=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
REVISION=$(git rev-parse HEAD)

LDFLAGS=""
LDFLAGS="${LDFLAGS} -X main.buildType=development"
LDFLAGS="${LDFLAGS} -X main.buildHost=${BUILD_HOST:-`hostname`}"
LDFLAGS="${LDFLAGS} -X main.buildTime=${NOW}"
LDFLAGS="${LDFLAGS} -X main.buildRevision=${REVISION}"

TAGS="development,local,noui,icons"
[ -d "server/icons/sources/simple-icons" ] && TAGS="${TAGS},simple_icons"

echo "[builder] building monetr now with tags (${TAGS})..."
go build -buildvcs=false -ldflags "${LDFLAGS}" -tags=${TAGS} -o /home/monetr/bin/monetr github.com/monetr/monetr/server/cmd
