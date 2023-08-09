#!/usr/bin/env bash

if ! command -v pnpm &> /dev/null
then
    echo "[ui] pnpm could not be found, it will be installed"
    npm install -g pnpm
else
    echo "[ui] pnpm is already installed, skipping..."
fi

pnpm start
