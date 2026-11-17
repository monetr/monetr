-- Nullable for now: existing rows keep their secrets in the secrets table (or
-- plaintext TOTP on logins) until they are backfilled. Each table gets the same
-- columns and constraints from models.Encrypted:
--  - The FK restricts so a KEK cant be removed while a DEK is still wrapped by it
--  - The check makes sure encrypted data always records which KEK wrapped its DEK
--  - The index is for the RESTRICT check on KEK deletes and for finding rows
--    still wrapped by an old KEK

ALTER TABLE "plaid_links" ADD COLUMN "key_encryption_key_id" VARCHAR(32);
ALTER TABLE "plaid_links" ADD COLUMN "encrypted_data" TEXT;
ALTER TABLE "plaid_links" ADD CONSTRAINT "fk_plaid_links_key_encryption_key" FOREIGN KEY ("key_encryption_key_id") REFERENCES "key_encryption_keys" ("key_encryption_key_id") ON DELETE RESTRICT;
ALTER TABLE "plaid_links" ADD CONSTRAINT "ck_plaid_links_encrypted_data_key_encryption_key" CHECK ("encrypted_data" IS NULL OR "key_encryption_key_id" IS NOT NULL);
CREATE INDEX "ix_plaid_links_key_encryption_key" ON "plaid_links" ("key_encryption_key_id");

ALTER TABLE "lunch_flow_links" ADD COLUMN "key_encryption_key_id" VARCHAR(32);
ALTER TABLE "lunch_flow_links" ADD COLUMN "encrypted_data" TEXT;
ALTER TABLE "lunch_flow_links" ADD CONSTRAINT "fk_lunch_flow_links_key_encryption_key" FOREIGN KEY ("key_encryption_key_id") REFERENCES "key_encryption_keys" ("key_encryption_key_id") ON DELETE RESTRICT;
ALTER TABLE "lunch_flow_links" ADD CONSTRAINT "ck_lunch_flow_links_encrypted_data_key_encryption_key" CHECK ("encrypted_data" IS NULL OR "key_encryption_key_id" IS NOT NULL);
CREATE INDEX "ix_lunch_flow_links_key_encryption_key" ON "lunch_flow_links" ("key_encryption_key_id");

ALTER TABLE "simplefin_links" ADD COLUMN "key_encryption_key_id" VARCHAR(32);
ALTER TABLE "simplefin_links" ADD COLUMN "encrypted_data" TEXT;
ALTER TABLE "simplefin_links" ADD CONSTRAINT "fk_simplefin_links_key_encryption_key" FOREIGN KEY ("key_encryption_key_id") REFERENCES "key_encryption_keys" ("key_encryption_key_id") ON DELETE RESTRICT;
ALTER TABLE "simplefin_links" ADD CONSTRAINT "ck_simplefin_links_encrypted_data_key_encryption_key" CHECK ("encrypted_data" IS NULL OR "key_encryption_key_id" IS NOT NULL);
CREATE INDEX "ix_simplefin_links_key_encryption_key" ON "simplefin_links" ("key_encryption_key_id");

ALTER TABLE "logins" ADD COLUMN "key_encryption_key_id" VARCHAR(32);
ALTER TABLE "logins" ADD COLUMN "encrypted_data" TEXT;
ALTER TABLE "logins" ADD CONSTRAINT "fk_logins_key_encryption_key" FOREIGN KEY ("key_encryption_key_id") REFERENCES "key_encryption_keys" ("key_encryption_key_id") ON DELETE RESTRICT;
ALTER TABLE "logins" ADD CONSTRAINT "ck_logins_encrypted_data_key_encryption_key" CHECK ("encrypted_data" IS NULL OR "key_encryption_key_id" IS NOT NULL);
CREATE INDEX "ix_logins_key_encryption_key" ON "logins" ("key_encryption_key_id");
