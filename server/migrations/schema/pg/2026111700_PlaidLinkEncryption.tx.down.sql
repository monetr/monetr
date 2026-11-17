DROP INDEX IF EXISTS "ix_plaid_links_key_encryption_key";
ALTER TABLE "plaid_links" DROP CONSTRAINT IF EXISTS "ck_plaid_links_encrypted_data_key_encryption_key";
ALTER TABLE "plaid_links" DROP CONSTRAINT IF EXISTS "fk_plaid_links_key_encryption_key";
ALTER TABLE "plaid_links" DROP COLUMN IF EXISTS "encrypted_data";
ALTER TABLE "plaid_links" DROP COLUMN IF EXISTS "key_encryption_key_id";
