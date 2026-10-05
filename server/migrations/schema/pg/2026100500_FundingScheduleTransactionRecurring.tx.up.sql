ALTER TABLE "funding_schedules" ADD COLUMN "transaction_recurring_id" VARCHAR(32);

-- Only null out the recurring ID when a recurring transaction is deleted,
-- account and bank account are part of this key too and need to stay
ALTER TABLE "funding_schedules" ADD CONSTRAINT "fk_funding_schedules_transaction_recurring"
FOREIGN KEY ("transaction_recurring_id", "account_id", "bank_account_id")
REFERENCES "transaction_recurring" ("transaction_recurring_id", "account_id", "bank_account_id")
ON DELETE SET NULL ("transaction_recurring_id");

-- A recurring transaction can only be tracked by a single funding schedule.
-- Nulls are distinct so funding schedules without a recurring transaction are
-- fine.
ALTER TABLE "funding_schedules" ADD CONSTRAINT "uq_funding_schedules_transaction_recurring"
UNIQUE ("account_id", "bank_account_id", "transaction_recurring_id");
