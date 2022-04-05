#!/usr/bin/env sh

echo "[docs] Installing swagger plugin!"
pip install mkdocs-render-swagger-plugin

mkdocs build
