#!/usr/bin/env bash

[ ! -f "$PWD/build/ed25519.key" ] && openssl genpkey -algorithm ED25519 -out $PWD/build/ed25519.key
[ -f "$PWD/build/ed25519.key" ] && chown -R $UID:$GID $PWD/build/ed25519.key

if [ "$DISABLE_GO_RELOAD" == "true" ]
then
  echo "[executor] hot reload is disabled, monetr will be run normally"
  bash -c $PWD/compose/api-builder.sh # Build the executable
  bash -c $PWD/compose/api-wrapper.sh # Execute the executable
else
  # Air gives us a hot reloader for golang.
  air -c $PWD/air.toml
fi

