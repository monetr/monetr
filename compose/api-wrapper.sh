#!/usr/bin/env bash

# Try to hit the ngrok API inside docker compose, if this succeeds then that means webhooks are enabled for plaid for
# local development.
export MONETR_PLAID_WEBHOOKS_DOMAIN=$(curl http://ngrok:4040/api/tunnels -s -m 0.1 | perl -pe '/\"public_url\":\"https:\/\/(\S*?)\",/g; print $1;' | cut -d "{" -f1)
if [[ ! -z "${MONETR_PLAID_WEBHOOKS_DOMAIN}" ]]; then
  # If the domain name has been derived then enable webhooks for plaid.
  export MONETR_PLAID_WEBHOOKS_ENABLED="true";
fi

# Execute the command with the new environment variables.
/go/bin/dlv exec --continue --accept-multiclient --listen=:2345 --headless=true --api-version=2 /usr/bin/monetr -- serve --migrate=true;
