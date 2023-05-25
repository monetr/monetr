#!/usr/bin/env sh

echo "STABLE_GIT_REVISION $(git rev-parse HEAD)"
echo "STABLE_GIT_RELEASE $(git describe --tags --dirty)"
echo BUILD_TIME "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
echo "STABLE_WORKSPACE $(pwd)"
