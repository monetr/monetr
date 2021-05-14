
-- I cannot for the life of me make up my mind about how I want to address billing in MVP. So we are making it stupid
-- simple. God I hate billing.
DROP TABLE "subscription_items";
DROP TABLE "prices";
DROP TABLE "products";

ALTER TABLE "subscriptions" ADD COLUMN "stripe_price_id" TEXT;
ALTER TABLE "subscriptions" ALTER COLUMN "stripe_price_id" SET NOT NULL;
ALTER TABLE "subscriptions" ADD COLUMN "features" TEXT[];