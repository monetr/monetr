-- Add a new constraint that has the ON DELETE SET NULL clause.
ALTER TABLE "transactions" ADD CONSTRAINT "fk_transactions_spending"
FOREIGN KEY ("spending_id", "account_id", "bank_account_id") REFERENCES "spending" ("spending_id", "account_id", "bank_account_id")
ON DELETE SET NULL;

-- Then remove the old constraint.
ALTER TABLE "transactions" DROP CONSTRAINT "fk_transactions_spending_spending_id_account_id_bank_account_id";
