ALTER TABLE "accounts" ADD COLUMN "stripe_customer_id" TEXT NULL;
ALTER TABLE "accounts" ADD CONSTRAINT "uq_accounts_stripe_customer_id" UNIQUE ("stripe_customer_id");

ALTER TABLE "accounts" ADD COLUMN "stripe_subscription_id" TEXT NULL;
ALTER TABLE "accounts" ADD CONSTRAINT "uq_accounts_stripe_subscription_id" UNIQUE ("stripe_subscription_id");

ALTER TABLE "accounts" ADD COLUMN "subscription_active_until" TIMESTAMPTZ NULL;

DROP TABLE "subscriptions";