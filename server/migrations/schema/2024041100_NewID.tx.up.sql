ALTER TABLE "logins"
ADD COLUMN "login_id_new" VARCHAR(64);

ALTER TABLE "accounts"
ADD COLUMN "account_id_new" VARCHAR(64);

ALTER TABLE "users"
ADD COLUMN "user_id_new" VARCHAR(64),
ADD COLUMN "login_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64);

ALTER TABLE "links"
ADD COLUMN "link_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "plaid_link_id_new" VARCHAR(64),
ADD COLUMN "teller_link_id_new" VARCHAR(64),
ADD COLUMN "created_by_user_id_new" VARCHAR(64);

ALTER TABLE "plaid_links"
ADD COLUMN "plaid_link_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "secret_id_new" VARCHAR(64),
ADD COLUMN "created_by_user_id_new" VARCHAR(64);

ALTER TABLE "plaid_syncs"
ADD COLUMN "plaid_sync_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "plaid_link_id_new" VARCHAR(64);

ALTER TABLE "plaid_bank_accounts"
ADD COLUMN "plaid_bank_account_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "plaid_link_id_new" VARCHAR(64),
ADD COLUMN "created_by_user_id_new" VARCHAR(64);

ALTER TABLE "plaid_transactions"
ADD COLUMN "plaid_transaction_id_new" VARCHAR(64);

ALTER TABLE "teller_links"
ADD COLUMN "teller_link_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "secret_id_new" VARCHAR(64),
ADD COLUMN "created_by_user_id_new" VARCHAR(64);

ALTER TABLE "teller_bank_accounts"
ADD COLUMN "teller_bank_account_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "teller_link_id_new" VARCHAR(64);

ALTER TABLE "teller_syncs"
ADD COLUMN "teller_sync_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "teller_bank_account_id_new" VARCHAR(64);

ALTER TABLE "teller_transactions"
ADD COLUMN "teller_transaction_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "teller_bank_account_id_new" VARCHAR(64);
