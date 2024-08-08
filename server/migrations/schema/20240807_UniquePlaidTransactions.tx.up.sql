ALTER TABLE "plaid_transactions"
ADD CONSTRAINT "uq_plaid_transactions_plaid_id" UNIQUE ("account_id", "plaid_bank_account_id", "plaid_id");
