#!/usr/bin/env bash

if [ ! -f "/tmp/locales-marker" ]
then
  echo "[executor] locales have not been installed yet, installing them now"
  apt-get update && apt-get install -y locales-all tzdata
fi

if ! command -v air &> /dev/null
then
    echo "[executor] cosmtrek/air could not be found, it will be installed"
    go install github.com/cosmtrek/air@v1.29.0
else
    echo "[executor] cosmtrek/air is already installed, skipping..."
fi

[ ! -f "/build/build/ed25519.key" ] && openssl genpkey -algorithm ED25519 -out /build/build/ed25519.key
[ -f "/build/build/ed25519.key" ] && chown -R $UID:$GID /build/build/ed25519.key

if [ "$DISABLE_GO_RELOAD" == "true" ]
then
  echo "[executor] hot reload is disabled, monetr will be run normally"
  bash -c $PWD/compose/api-builder.sh # Build the executable
  bash -c $PWD/compose/api-wrapper.sh # Execute the executable
else
  # Air gives us a hot reloader for golang.
  air -c /build/air.toml
fi

