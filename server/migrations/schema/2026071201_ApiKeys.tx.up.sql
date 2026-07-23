CREATE TABLE "api_keys" (
  "api_key_id" VARCHAR(32) NOT NULL,
  "account_id" VARCHAR(32) NOT NULL,
  "name"       TEXT        NOT NULL,
  "public_key" BYTEA       NOT NULL,
  "created_by" VARCHAR(32) NOT NULL,
  "created_at" TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "updated_at" TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "deleted_at" TIMESTAMP WITHOUT TIME ZONE,
  CONSTRAINT "pk_api_keys"            PRIMARY KEY ("api_key_id"),
  CONSTRAINT "fk_api_keys_account"    FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_api_keys_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("user_id") ON DELETE CASCADE
);

-- Do not allow duplicate API key names within the same account, excluding
-- deleted api keys.
CREATE UNIQUE INDEX "ix_uq_api_keys_name" ON "api_keys" ("account_id", "name") WHERE "deleted_at" IS NULL;
