-- Add the old constraint back.
ALTER TABLE "transactions" ADD CONSTRAINT "fk_transactions_spending_spending_id_account_id_bank_account_id"
FOREIGN KEY ("spending_id", "account_id", "bank_account_id") REFERENCES "spending" ("spending_id", "account_id", "bank_account_id");

-- Remove the new constraint.
ALTER TABLE "transactions" DROP CONSTRAINT "fk_transactions_spending";
