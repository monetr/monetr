CREATE TABLE "api_keys" (
  "api_key_id" VARCHAR(32) NOT NULL,
  "account_id" VARCHAR(32) NOT NULL,
  "public_key" BYTEA       NOT NULL,
  "created_by" VARCHAR(32) NOT NULL,
  "created_at" TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "updated_at" TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "deleted_at" TIMESTAMP WITHOUT TIME ZONE,
  CONSTRAINT "pk_api_keys"            PRIMARY KEY ("api_key_id"),
  CONSTRAINT "fk_api_keys_account"    FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_api_keys_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("user_id") ON DELETE CASCADE
);
