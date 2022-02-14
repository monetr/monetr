ALTER TABLE "transactions"
    ADD COLUMN "date_fixed" TIMESTAMPTZ;
UPDATE "transactions"
SET "date_fixed" = "date"::timestamptz;
ALTER TABLE "transactions"
    ALTER COLUMN "date_fixed" SET NOT NULL;
ALTER TABLE "transactions"
    DROP COLUMN "date";
ALTER TABLE "transactions"
    RENAME COLUMN "date_fixed" TO "date";


ALTER TABLE "transactions"
    ADD COLUMN "authorized_fixed" TIMESTAMPTZ;
UPDATE "transactions"
SET "authorized_fixed" = "authorized_date"::timestamptz;
ALTER TABLE "transactions"
    DROP COLUMN "authorized_date";
ALTER TABLE "transactions"
    RENAME COLUMN "authorized_fixed" TO "authorized_date";
