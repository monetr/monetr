ALTER TABLE "spending" ADD COLUMN "auto_create_transaction" BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE "funding_schedules" ADD COLUMN "auto_create_transaction" BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE "transactions" ADD COLUMN "created_by_spending_id" VARCHAR(32);
ALTER TABLE "transactions" ADD COLUMN "created_by_funding_schedule_id" VARCHAR(32);

ALTER TABLE "transactions" ADD CONSTRAINT "fk_transactions_created_by_spending"
  FOREIGN KEY ("created_by_spending_id", "account_id", "bank_account_id") REFERENCES "spending" ("spending_id", "account_id", "bank_account_id")
  ON DELETE SET NULL;

ALTER TABLE "transactions" ADD CONSTRAINT "fk_transactions_created_by_funding_schedule"
  FOREIGN KEY ("created_by_funding_schedule_id", "account_id", "bank_account_id") REFERENCES "funding_schedules" ("funding_schedule_id", "account_id", "bank_account_id")
  ON DELETE SET NULL;
