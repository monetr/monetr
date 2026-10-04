ALTER TABLE "spending" ADD COLUMN "transaction_recurring_id" VARCHAR(32);

-- Only null out the recurring ID when a recurring transaction is deleted,
-- account and bank account are part of this key too and need to stay
ALTER TABLE "spending" ADD CONSTRAINT "fk_spending_transaction_recurring"
FOREIGN KEY ("transaction_recurring_id", "account_id", "bank_account_id")
REFERENCES "transaction_recurring" ("transaction_recurring_id", "account_id", "bank_account_id")
ON DELETE SET NULL ("transaction_recurring_id");

-- A recurring transaction can only be tracked by a single spending object.
-- Nulls are distinct so spending without a recurring transaction is fine.
ALTER TABLE "spending" ADD CONSTRAINT "uq_spending_transaction_recurring"
UNIQUE ("account_id", "bank_account_id", "transaction_recurring_id");
