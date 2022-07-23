#!/usr/bin/env bash

# Try to hit the ngrok API inside docker compose, if this succeeds then that means webhooks are enabled for plaid for
# local development.

if [[ ! -z ${GITPOD_WORKSPACE_ID} ]]; then
  echo "[wrapper] gitpod detected, will use gitpod URL for webhooks instead";

  export MONETR_PLAID_WEBHOOKS_DOMAIN="https://${MONETR_API_DOMAIN_NAME}";
  export MONETR_PLAID_WEBHOOKS_ENABLED="true";
else
  WEBHOOKS_DOMAIN=$(curl http://ngrok:4040/api/tunnels -s -m 0.1 | perl -pe '/\"public_url\":\"https:\/\/(\S*?)\",/g; print $1;' | cut -d "{" -f1);

  if [[ ! -z "${WEBHOOKS_DOMAIN}" ]]; then
    echo "[wrapper] ngrok detected, webhooks should target: ${WEBHOOKS_DOMAIN}";

    # If the domain name has been derived then enable webhooks for plaid.
    echo "[wrapper] Plaid webhooks have been enabled...";
    export MONETR_PLAID_WEBHOOKS_DOMAIN=${WEBHOOKS_DOMAIN};
    export MONETR_PLAID_WEBHOOKS_ENABLED="true";
  else
    echo "[wrapper] ngrok not detected, webhooks will not be available..."
  fi
fi


# If the stripe API key, webhook secret and price ID are provided then enable billing for local development.
# Stripe does require webhooks, as we rely on them in order to know when a subscription becomes active.
if [[ ! -z "${MONETR_STRIPE_API_KEY}" ]] && \
  [[ ! -z "${MONETR_STRIPE_WEBHOOK_SECRET}" ]] && \
  [[ ! -z "${MONETR_STRIPE_DEFAULT_PRICE_ID}" ]] && \
  [[ ! -z "${WEBHOOKS_DOMAIN}" ]]; then
  echo "[wrapper] Stripe credentials are available, stripe and billing will be enabled...";
  export MONETR_STRIPE_ENABLED="true";
  export MONETR_STRIPE_BILLING_ENABLED="true";
  export MONETR_STRIPE_WEBHOOKS_ENABLED="true";
  export MONETR_STRIPE_WEBHOOKS_DOMAIN=${WEBHOOKS_DOMAIN};
fi

if [[ ! -z "${MONETR_CAPTCHA_PUBLIC_KEY}" ]] && [[ ! -z "${MONETR_CAPTCHA_PRIVATE_KEY}" ]]; then
  echo "[wrapper] ReCAPTCHA credentials detected, requiring verification...";
  export MONETR_CAPTCHA_ENABLED="true";
fi

if [[ ! -z "${MONETR_SENTRY_DSN}" ]]; then
  echo "[wrapper] Sentry DSN detected, enabling...";
  export MONETR_SENTRY_ENABLED="true";
fi

# Sometimes the old process does not get killed properly. This should do it.
pkill monetr;
pkill dlv;

# Execute the command with the new environment variables.
/go/bin/dlv exec --continue --api-version 2 --accept-multiclient --listen=:2345 --headless=true --api-version=2 /usr/bin/monetr -- serve --migrate=true;
