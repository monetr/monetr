#!/usr/bin/env sh

echo "[docs] Installing swagger plugin!"
pip install mkdocs-render-swagger-plugin mike

if [[ ! -z "${CI}" ]]; then
  git config --global user.name ${GIT_USER}
  git config --global user.email ${GIT_EMAIL}
fi

mike deploy --config=$PWD/mkdocs.yaml $1
