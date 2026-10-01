ALTER TABLE "transaction_uploads" DROP CONSTRAINT "fk_transaction_uploads_bank_account";
ALTER TABLE "transaction_uploads" ADD CONSTRAINT "fk_transaction_uploads_bank_account"
FOREIGN KEY ("bank_account_id", "account_id")
REFERENCES "bank_accounts" ("bank_account_id", "account_id");

ALTER TABLE "transaction_clusters" DROP CONSTRAINT "fk_transaction_clusters_bank_account";
ALTER TABLE "transaction_clusters" ADD CONSTRAINT "fk_transaction_clusters_bank_account"
FOREIGN KEY ("bank_account_id", "account_id")
REFERENCES "bank_accounts" ("bank_account_id", "account_id");

ALTER TABLE "funding_schedules" DROP CONSTRAINT "fk_funding_schedules_bank_account";
ALTER TABLE "funding_schedules" ADD CONSTRAINT "fk_funding_schedules_bank_account"
FOREIGN KEY ("bank_account_id", "account_id")
REFERENCES "bank_accounts" ("bank_account_id", "account_id");

ALTER TABLE "spending" DROP CONSTRAINT "fk_spending_bank_account";
ALTER TABLE "spending" ADD CONSTRAINT "fk_spending_bank_account"
FOREIGN KEY ("bank_account_id", "account_id")
REFERENCES "bank_accounts" ("bank_account_id", "account_id");

ALTER TABLE "transactions" DROP CONSTRAINT "fk_transactions_bank_account";
ALTER TABLE "transactions" ADD CONSTRAINT "fk_transactions_bank_account"
FOREIGN KEY ("bank_account_id", "account_id")
REFERENCES "bank_accounts" ("bank_account_id", "account_id");

ALTER TABLE "bank_accounts" DROP CONSTRAINT "fk_bank_accounts_link";
ALTER TABLE "bank_accounts" ADD CONSTRAINT "fk_bank_accounts_link"
FOREIGN KEY ("link_id", "account_id")
REFERENCES "links" ("link_id", "account_id");
