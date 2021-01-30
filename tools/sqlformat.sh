#!/bin/bash

#
# author: Michał Skrzyński (skrzynski.michal@gmail.com)
# version: 1.0.2 (28.11.2019)
#

[ ! -x /usr/bin/curl ] && { echo 'Error: curl is not installed.' >&2 ;exit 1; }
# [ ! -x /usr/bin/jq ] && { echo 'Error: jq is not installed.' >&2 ;exit 1; }

[[ -f "${1}" || -p /dev/stdin ]] && { sql=$(cat "${1:--}"); } || { sql="${1}"; }

curl -s -X POST -d "reindent=1&sql=${sql}" https://sqlformat.org/api/v1/format|jq -r .result