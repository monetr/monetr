-- Nullable for now: existing links keep their secret in the secrets table
-- until they are backfilled
ALTER TABLE "plaid_links" ADD COLUMN "key_encryption_key_id" VARCHAR(32);
ALTER TABLE "plaid_links" ADD COLUMN "encrypted_data" TEXT;
-- Restrict so a KEK cant be removed while a DEK is still wrapped by it
ALTER TABLE "plaid_links" ADD CONSTRAINT "fk_plaid_links_key_encryption_key" FOREIGN KEY ("key_encryption_key_id") REFERENCES "key_encryption_keys" ("key_encryption_key_id") ON DELETE RESTRICT;
-- Encrypted data cant be decrypted without knowing which KEK wrapped its DEK
ALTER TABLE "plaid_links" ADD CONSTRAINT "ck_plaid_links_encrypted_data_key_encryption_key" CHECK ("encrypted_data" IS NULL OR "key_encryption_key_id" IS NOT NULL);
-- For the RESTRICT check on KEK deletes and for finding links still wrapped by
-- an old KEK
CREATE INDEX "ix_plaid_links_key_encryption_key" ON "plaid_links" ("key_encryption_key_id");
