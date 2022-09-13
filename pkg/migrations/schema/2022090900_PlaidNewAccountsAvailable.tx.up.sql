ALTER TABLE "links" ADD COLUMN plaid_new_accounts_available BOOL DEFAULT FALSE;
UPDATE "links" SET plaid_new_accounts_available = false;
ALTER TABLE "links" ALTER COLUMN plaid_new_accounts_available SET NOT NULL;
