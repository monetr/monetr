#!/usr/bin/env bash

if ! command -v air &> /dev/null
then
    echo "[executor] cosmtrek/air could not be found, it will be installed"
    go install github.com/cosmtrek/air@v1.29.0
else
    echo "[executor] cosmtrek/air is already installed, skipping..."
fi


if [ "$DISABLE_GO_RELOAD" == "true" ]
then
  echo "[executor] hot reload is disabled, monetr will be run normally"
  bash -c $PWD/compose/api-builder.sh # Build the executable
  bash -c $PWD/compose/api-wrapper.sh # Execute the executable
else
  # Air gives us a hot reloader for golang. I'm doing this instead of using the container as this will support other
  # architectures a bit more gracefully at the moment.
  air -c /build/air.toml
fi

