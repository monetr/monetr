-- Restoring NOT NULL on secret_id fails if any link only has an encrypted
-- secret, rolling back past this point would lose those secrets.
DROP INDEX IF EXISTS "ix_logins_key_encryption_key";
ALTER TABLE "logins" DROP CONSTRAINT IF EXISTS "ck_logins_encrypted_data_key_encryption_key";
ALTER TABLE "logins" DROP CONSTRAINT IF EXISTS "fk_logins_key_encryption_key";
ALTER TABLE "logins" DROP COLUMN IF EXISTS "encrypted_data";
ALTER TABLE "logins" DROP COLUMN IF EXISTS "key_encryption_key_id";

DROP INDEX IF EXISTS "ix_simplefin_links_key_encryption_key";
ALTER TABLE "simplefin_links" DROP CONSTRAINT IF EXISTS "ck_simplefin_links_secret_or_encrypted";
ALTER TABLE "simplefin_links" ALTER COLUMN "secret_id" SET NOT NULL;
ALTER TABLE "simplefin_links" DROP CONSTRAINT IF EXISTS "ck_simplefin_links_encrypted_data_key_encryption_key";
ALTER TABLE "simplefin_links" DROP CONSTRAINT IF EXISTS "fk_simplefin_links_key_encryption_key";
ALTER TABLE "simplefin_links" DROP COLUMN IF EXISTS "encrypted_data";
ALTER TABLE "simplefin_links" DROP COLUMN IF EXISTS "key_encryption_key_id";

DROP INDEX IF EXISTS "ix_lunch_flow_links_key_encryption_key";
ALTER TABLE "lunch_flow_links" DROP CONSTRAINT IF EXISTS "ck_lunch_flow_links_secret_or_encrypted";
ALTER TABLE "lunch_flow_links" ALTER COLUMN "secret_id" SET NOT NULL;
ALTER TABLE "lunch_flow_links" DROP CONSTRAINT IF EXISTS "ck_lunch_flow_links_encrypted_data_key_encryption_key";
ALTER TABLE "lunch_flow_links" DROP CONSTRAINT IF EXISTS "fk_lunch_flow_links_key_encryption_key";
ALTER TABLE "lunch_flow_links" DROP COLUMN IF EXISTS "encrypted_data";
ALTER TABLE "lunch_flow_links" DROP COLUMN IF EXISTS "key_encryption_key_id";

DROP INDEX IF EXISTS "ix_plaid_links_key_encryption_key";
ALTER TABLE "plaid_links" DROP CONSTRAINT IF EXISTS "ck_plaid_links_secret_or_encrypted";
ALTER TABLE "plaid_links" ALTER COLUMN "secret_id" SET NOT NULL;
ALTER TABLE "plaid_links" DROP CONSTRAINT IF EXISTS "ck_plaid_links_encrypted_data_key_encryption_key";
ALTER TABLE "plaid_links" DROP CONSTRAINT IF EXISTS "fk_plaid_links_key_encryption_key";
ALTER TABLE "plaid_links" DROP COLUMN IF EXISTS "encrypted_data";
ALTER TABLE "plaid_links" DROP COLUMN IF EXISTS "key_encryption_key_id";
