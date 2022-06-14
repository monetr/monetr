#!/usr/bin/env sh

echo "[docs] Installing swagger plugin!"
pip install mkdocs-render-swagger-plugin
pip install git+https://github.com/jimporter/mike.git

mkdocs serve --dev-addr=0.0.0.0:8000
