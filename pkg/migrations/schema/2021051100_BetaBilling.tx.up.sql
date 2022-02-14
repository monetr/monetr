ALTER TABLE "prices"
    DROP COLUMN "trial_period_days";
ALTER TABLE "prices"
    DROP COLUMN "price_code";
ALTER TABLE "products"
    DROP COLUMN "product_code";

DROP TABLE "subscriptions";

CREATE TABLE "subscriptions"
(
    subscription_id        BIGSERIAL   NOT NULL,
    account_id             BIGINT      NOT NULL,
    owned_by_user_id       BIGINT      NOT NULL,
    stripe_subscription_id TEXT        NOT NULL,
    stripe_customer_id     TEXT        NOT NULL,
    status                 TEXT        NOT NULL,
    trial_start            TIMESTAMPTZ NULL,
    trial_end              TIMESTAMPTZ NULL,
    CONSTRAINT pk_subscriptions PRIMARY KEY ("subscription_id", "account_id"),
    CONSTRAINT fk_subscriptions_account FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
    CONSTRAINT fk_subscriptions_owner FOREIGN KEY ("owned_by_user_id") REFERENCES "users" ("user_id"),
    CONSTRAINT uq_subscriptions_stripe_id UNIQUE ("stripe_subscription_id")
);

CREATE TABLE "subscription_items"
(
    subscription_item_id        BIGSERIAL NOT NULL,
    subscription_id             BIGINT    NOT NULL,
    account_id                  BIGINT    NOT NULL,
    stripe_subscription_item_id TEXT      NOT NULL,
    price_id                    BIGINT    NOT NULL,
    quantity                    SMALLINT  NOT NULL,
    CONSTRAINT pk_subscription_items PRIMARY KEY ("subscription_item_id", "subscription_id", "account_id"),
    CONSTRAINT fk_subscription_items_subscription FOREIGN KEY ("subscription_id", "account_id") REFERENCES "subscriptions" ("subscription_id", "account_id") ON DELETE CASCADE,
    CONSTRAINT fk_subscription_items_account FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
    CONSTRAINT fk_subscription_items_price FOREIGN KEY ("price_id") REFERENCES "prices" ("price_id") ON DELETE RESTRICT,
    CONSTRAINT uq_subscription_item_stripe_id UNIQUE ("stripe_subscription_item_id")
);
;