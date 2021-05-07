CREATE TABLE products
(
    product_id        BIGSERIAL NOT NULL,
    product_code      TEXT      NOT NULL,
    name              TEXT      NOT NULL,
    description       TEXT,
    features          BIGINT    NOT NULL,
    stripe_product_id TEXT      NOT NULL,
    CONSTRAINT "pk_products" PRIMARY KEY ("product_id"),
    CONSTRAINT "uq_products_product_code" UNIQUE ("product_code")
);

CREATE TABLE prices
(
    price_id          BIGSERIAL NOT NULL,
    price_code        TEXT      NOT NULL,
    product_id        BIGINT    NOT NULL,
    interval          TEXT      NOT NULL,
    interval_count    INT       NOT NULL,
    trial_period_days INT,
    unit_amount       INT       NOT NULL,
    stripe_pricing_id TEXT      NOT NULL,
    CONSTRAINT "pk_prices" PRIMARY KEY ("price_id"),
    CONSTRAINT "uq_prices_price_code" UNIQUE ("price_code"),
    CONSTRAINT "fk_prices_products_product_id" FOREIGN KEY ("product_id") REFERENCES "products" ("product_id")
);

CREATE TABLE subscriptions
(
    subscription_id        BIGSERIAL NOT NULL,
    account_id             BIGINT    NOT NULL,
    owned_by_user_id       BIGINT    NOT NULL,
    price_id               BIGINT    NOT NULL,
    stripe_subscription_id TEXT      NOT NULL,
    stripe_customer_id     TEXT      NOT NULL,
    status                 TEXT      NOT NULL,
    CONSTRAINT "pk_subscriptions" PRIMARY KEY ("subscription_id", "account_id"),
    CONSTRAINT "fk_subscriptions_owner" FOREIGN KEY ("owned_by_user_id") REFERENCES "users" ("user_id"),
    CONSTRAINT "fk_subscriptions_price_id" FOREIGN KEY ("price_id") REFERENCES "prices" ("price_id")
);