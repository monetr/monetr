CREATE TABLE "spending_funding" (
  "spending_funding_id"      BIGSERIAL NOT NULL,
  "account_id"               BIGINT    NOT NULL,
  "bank_account_id"          BIGINT    NOT NULL,
  "spending_id"              BIGINT    NOT NULL,
  "funding_schedule_id"      BIGINT    NOT NULL,
  "next_contribution_amount" BIGINT    NOT NULL,
  CONSTRAINT "pk_spending_funding"          PRIMARY KEY ("spending_funding_id", "account_id", "bank_account_id"),
  CONSTRAINT "uq_spending_funding"          UNIQUE      ("spending_id", "funding_schedule_id"),
  CONSTRAINT "fk_spending_funding_account"  FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_spending_funding_bank"     FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id"),
  CONSTRAINT "fk_spending_funding_spending" FOREIGN KEY ("spending_id", "account_id", "bank_account_id") REFERENCES "spending" ("spending_id", "account_id", "bank_account_id"),
  CONSTRAINT "fk_spending_funding_funding"  FOREIGN KEY ("funding_schedule_id", "account_id", "bank_account_id") REFERENCES "funding_schedules" ("funding_schedule_id", "account_id", "bank_account_id")
);

INSERT INTO "spending_funding" ("account_id", "bank_account_id", "spending_id", "funding_schedule_id", "next_contribution_amount")
SELECT
  "spending"."account_id",
  "spending"."bank_account_id",
  "spending"."spending_id",
  "spending"."funding_schedule_id",
  "spending"."next_contribution_amount"
FROM "spending";

ALTER TABLE "spending" DROP CONSTRAINT "fk_spending_funding_schedules_funding_schedule_id_account_id_ba";
ALTER TABLE "spending" DROP COLUMN "funding_schedule_id" CASCADE;
ALTER TABLE "spending" DROP COLUMN "next_contribution_amount";
