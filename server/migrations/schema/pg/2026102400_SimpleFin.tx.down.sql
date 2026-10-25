ALTER TABLE "transactions" DROP CONSTRAINT IF EXISTS "uq_transactions_simplefin_transaction";
ALTER TABLE "transactions" DROP CONSTRAINT IF EXISTS "fk_transactions_simplefin_transaction";
ALTER TABLE "transactions" DROP COLUMN IF EXISTS "simplefin_transaction_id";
DROP TABLE IF EXISTS "simplefin_transactions";

ALTER TABLE "bank_accounts" DROP CONSTRAINT IF EXISTS "uq_bank_accounts_simplefin_bank_account";
ALTER TABLE "bank_accounts" DROP CONSTRAINT IF EXISTS "fk_bank_accounts_simplefin_bank_account";
ALTER TABLE "bank_accounts" DROP COLUMN IF EXISTS "simplefin_bank_account_id";
DROP TABLE IF EXISTS "simplefin_bank_accounts";

ALTER TABLE "links" DROP CONSTRAINT IF EXISTS "uq_links_simplefin_link";
ALTER TABLE "links" DROP CONSTRAINT IF EXISTS "fk_links_simplefin_link";
ALTER TABLE "links" DROP COLUMN IF EXISTS "simplefin_link_id";
DROP TABLE IF EXISTS "simplefin_links";
