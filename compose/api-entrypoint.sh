#!/usr/bin/env bash

if ! command -v air &> /dev/null
then
    echo "[executor] cosmtrek/air could not be found, it will be installed"
    go install github.com/cosmtrek/air@v1.29.0
fi

# Air gives us a hot reloader for golang. I'm doing this instead of using the container as this will support other
# architectures a bit more gracefully at the moment.
air -c /build/air.toml
