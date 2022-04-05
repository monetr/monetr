#!/usr/bin/env sh

echo "[docs] Installing swagger plugin!"
pip install mkdocs-render-swagger-plugin

mkdocs serve --dev-addr=0.0.0.0:8000
