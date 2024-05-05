ALTER TABLE "logins"
ADD COLUMN "login_id_new" VARCHAR(64);

ALTER TABLE "accounts"
ADD COLUMN "account_id_new" VARCHAR(64);

ALTER TABLE "users"
ADD COLUMN "user_id_new" VARCHAR(64);

ALTER TABLE "betas"
ADD COLUMN "beta_id_new" VARCHAR(64);

ALTER TABLE "links"
ADD COLUMN "link_id_new" VARCHAR(64);

ALTER TABLE "secrets"
ADD COLUMN "secret_id_new" VARCHAR(64);

ALTER TABLE "bank_accounts"
ADD COLUMN "bank_account_id_new" VARCHAR(64);

ALTER TABLE "transactions"
ADD COLUMN "transaction_id_new" VARCHAR(64);

ALTER TABLE "transaction_clusters"
ADD COLUMN "transaction_cluster_id_new" VARCHAR(64);

ALTER TABLE "spending"
ADD COLUMN "spending_id_new" VARCHAR(64);

ALTER TABLE "funding_schedules"
ADD COLUMN "funding_schedule_id_new" VARCHAR(64);

ALTER TABLE "files"
ADD COLUMN "file_id_new" VARCHAR(64);

ALTER TABLE "jobs"
ADD COLUMN "job_id_new" VARCHAR(64);

ALTER TABLE "plaid_links"
ADD COLUMN "plaid_link_id_new" VARCHAR(64);

ALTER TABLE "plaid_syncs"
ADD COLUMN "plaid_sync_id_new" VARCHAR(64);

ALTER TABLE "plaid_bank_accounts"
ADD COLUMN "plaid_bank_account_id_new" VARCHAR(64);

ALTER TABLE "plaid_transactions"
ADD COLUMN "plaid_transaction_id_new" VARCHAR(64);

-- https://github.com/geckoboard/pgulid/blob/master/pgulid.sql
-- pgulid is based on OK Log's Go implementation of the ULID spec
--
-- https://github.com/oklog/ulid
-- https://github.com/ulid/spec
--
-- Copyright 2016 The Oklog Authors
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
-- http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION generate_ulid(kind TEXT, clock TIMESTAMP WITH TIME ZONE DEFAULT CLOCK_TIMESTAMP())
RETURNS TEXT
AS $$
DECLARE
  -- Crockford's Base32
  encoding   BYTEA = '0123456789ABCDEFGHJKMNPQRSTVWXYZ';
  timestamp  BYTEA = E'\\000\\000\\000\\000\\000\\000';
  output     TEXT = '';

  unix_time  BIGINT;
  ulid       BYTEA;
BEGIN
  -- 6 timestamp bytes
  unix_time = (EXTRACT(EPOCH FROM clock) * 1000)::BIGINT;
  timestamp = SET_BYTE(timestamp, 0, (unix_time >> 40)::BIT(8)::INTEGER);
  timestamp = SET_BYTE(timestamp, 1, (unix_time >> 32)::BIT(8)::INTEGER);
  timestamp = SET_BYTE(timestamp, 2, (unix_time >> 24)::BIT(8)::INTEGER);
  timestamp = SET_BYTE(timestamp, 3, (unix_time >> 16)::BIT(8)::INTEGER);
  timestamp = SET_BYTE(timestamp, 4, (unix_time >> 8)::BIT(8)::INTEGER);
  timestamp = SET_BYTE(timestamp, 5, unix_time::BIT(8)::INTEGER);

  -- 10 entropy bytes
  ulid = timestamp || gen_random_bytes(10);

  -- Encode the timestamp
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 0) & 224) >> 5));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 0) & 31)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 1) & 248) >> 3));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 1) & 7) << 2) | ((GET_BYTE(ulid, 2) & 192) >> 6)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 2) & 62) >> 1));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 2) & 1) << 4) | ((GET_BYTE(ulid, 3) & 240) >> 4)));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 3) & 15) << 1) | ((GET_BYTE(ulid, 4) & 128) >> 7)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 4) & 124) >> 2));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 4) & 3) << 3) | ((GET_BYTE(ulid, 5) & 224) >> 5)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 5) & 31)));

  -- Encode the entropy
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 6) & 248) >> 3));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 6) & 7) << 2) | ((GET_BYTE(ulid, 7) & 192) >> 6)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 7) & 62) >> 1));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 7) & 1) << 4) | ((GET_BYTE(ulid, 8) & 240) >> 4)));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 8) & 15) << 1) | ((GET_BYTE(ulid, 9) & 128) >> 7)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 9) & 124) >> 2));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 9) & 3) << 3) | ((GET_BYTE(ulid, 10) & 224) >> 5)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 10) & 31)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 11) & 248) >> 3));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 11) & 7) << 2) | ((GET_BYTE(ulid, 12) & 192) >> 6)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 12) & 62) >> 1));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 12) & 1) << 4) | ((GET_BYTE(ulid, 13) & 240) >> 4)));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 13) & 15) << 1) | ((GET_BYTE(ulid, 14) & 128) >> 7)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 14) & 124) >> 2));
  output = output || CHR(GET_BYTE(encoding, ((GET_BYTE(ulid, 14) & 3) << 3) | ((GET_BYTE(ulid, 15) & 224) >> 5)));
  output = output || CHR(GET_BYTE(encoding, (GET_BYTE(ulid, 15) & 31)));

  RETURN kind || '_' || LOWER(output);
END
$$
LANGUAGE plpgsql
VOLATILE;

-- Repeat for every table.

-- Logins
WITH new_ids AS (
	SELECT "logins"."login_id", generate_ulid('lgn') AS "id"
	FROM "logins"
)
UPDATE "logins"
SET "login_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."login_id" = "logins"."login_id";

-- Accounts
WITH new_ids AS (
	SELECT "accounts"."account_id", generate_ulid('acct', "accounts"."created_at") AS "id"
	FROM "accounts"
)
UPDATE "accounts"
SET "account_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."account_id" = "accounts"."account_id";

-- Users
WITH new_ids AS (
	SELECT "users"."user_id", generate_ulid('user') AS "id"
	FROM "users"
)
UPDATE "users"
SET "user_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."user_id" = "users"."user_id";

-- Betas
WITH new_ids AS (
	SELECT "betas"."beta_id", generate_ulid('beta') AS "id"
	FROM "betas"
)
UPDATE "betas"
SET "beta_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."beta_id" = "betas"."beta_id";

-- Links
WITH new_ids AS (
	SELECT "links"."link_id", generate_ulid('link', "links"."created_at") AS "id"
	FROM "links"
)
UPDATE "links"
SET "link_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."link_id" = "links"."link_id";

-- Secrets
WITH new_ids AS (
	SELECT "secrets"."secret_id", generate_ulid('scrt', "secrets"."created_at") AS "id"
	FROM "secrets"
)
UPDATE "secrets"
SET "secret_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."secret_id" = "secrets"."secret_id";

-- Bank Accounts
WITH new_ids AS (
	SELECT "bank_accounts"."bank_account_id", generate_ulid('bac', "bank_accounts"."created_at") AS "id"
	FROM "bank_accounts"
)
UPDATE "bank_accounts"
SET "bank_account_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."bank_account_id" = "bank_accounts"."bank_account_id";

-- Transactions
WITH new_ids AS (
	SELECT "transactions"."transaction_id", generate_ulid('txn', "transactions"."created_at") AS "id"
	FROM "transactions"
)
UPDATE "transactions"
SET "transaction_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."transaction_id" = "transactions"."transaction_id";

-- Transaction Clusters
WITH new_ids AS (
	SELECT "transaction_clusters"."transaction_cluster_id", generate_ulid('tcl', "transaction_clusters"."created_at") AS "id"
	FROM "transaction_clusters"
)
UPDATE "transaction_clusters"
SET "transaction_cluster_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."transaction_cluster_id" = "transaction_clusters"."transaction_cluster_id";

-- Spending
WITH new_ids AS (
	SELECT "spending"."spending_id", generate_ulid('spnd', "spending"."date_created") AS "id"
	FROM "spending"
)
UPDATE "spending"
SET "spending_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."spending_id" = "spending"."spending_id";

-- Funding schedule
WITH new_ids AS (
	SELECT "funding_schedules"."funding_schedule_id", generate_ulid('fund') AS "id"
	FROM "funding_schedules"
)
UPDATE "funding_schedules"
SET "funding_schedule_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."funding_schedule_id" = "funding_schedules"."funding_schedule_id";

-- Files
WITH new_ids AS (
	SELECT "files"."file_id", generate_ulid('file', "files"."created_at") AS "id"
	FROM "files"
)
UPDATE "files"
SET "file_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."file_id" = "files"."file_id";

-- Jobs
WITH new_ids AS (
	SELECT "jobs"."job_id", generate_ulid('job') AS "id"
	FROM "jobs"
)
UPDATE "jobs"
SET "job_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."job_id" = "jobs"."job_id";

-- Plaid Links
WITH new_ids AS (
	SELECT "plaid_links"."plaid_link_id", generate_ulid('plx', "plaid_links"."created_at") AS "id"
	FROM "plaid_links"
)
UPDATE "plaid_links"
SET "plaid_link_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."plaid_link_id" = "plaid_links"."plaid_link_id";

-- Plaid Syncs
WITH new_ids AS (
	SELECT "plaid_syncs"."plaid_sync_id", generate_ulid('psyn', "plaid_syncs"."timestamp") AS "id"
	FROM "plaid_syncs"
)
UPDATE "plaid_syncs"
SET "plaid_sync_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."plaid_sync_id" = "plaid_syncs"."plaid_sync_id";

-- Plaid Bank Accounts
WITH new_ids AS (
	SELECT "plaid_bank_accounts"."plaid_bank_account_id", generate_ulid('pbac', "plaid_bank_accounts"."created_at") AS "id"
	FROM "plaid_bank_accounts"
)
UPDATE "plaid_bank_accounts"
SET "plaid_bank_account_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."plaid_bank_account_id" = "plaid_bank_accounts"."plaid_bank_account_id";

-- Plaid Transactions
WITH new_ids AS (
	SELECT "plaid_transactions"."plaid_transaction_id", generate_ulid('ptxn', "plaid_transactions"."created_at") AS "id"
	FROM "plaid_transactions"
)
UPDATE "plaid_transactions"
SET "plaid_transaction_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."plaid_transaction_id" = "plaid_transactions"."plaid_transaction_id";

-- Swap tables

ALTER TABLE "logins" RENAME CONSTRAINT "pk_logins" TO "pk_logins_old";
ALTER TABLE "logins" DROP CONSTRAINT "uq_logins_email";
ALTER TABLE "logins" RENAME TO "logins_old";

CREATE TABLE "logins" (
  "login_id"          VARCHAR(32) NOT NULL,
  "email"             VARCHAR(250) NOT NULL,
  "first_name"        VARCHAR(250),
  "last_name"         VARCHAR(250),
  "totp"              TEXT,
  "totp_enabled_at"   TIMESTAMP WITH TIME ZONE,
  "crypt"             BYTEA NOT NULL,
  "is_enabled"        BOOLEAN NOT NULL,
  "is_email_verified" BOOLEAN NOT NULL,
  "email_verified_at" TIMESTAMP WITH TIME ZONE,
  "password_reset_at" TIMESTAMP WITH TIME ZONE,
  CONSTRAINT "pk_logins" PRIMARY KEY ("login_id"),
  CONSTRAINT "uq_logins_email" UNIQUE ("email")
);

INSERT INTO "logins" ("login_id", "email", "first_name", "last_name", "totp", "totp_enabled_at", "crypt", "is_enabled", "is_email_verified", "email_verified_at", "password_reset_at")
SELECT
  "l"."login_id_new",
  "l"."email",
  "l"."first_name",
  "l"."last_name",
  "l"."totp",
  "l"."totp_enabled_at",
  "l"."crypt",
  "l"."is_enabled",
  "l"."is_email_verified",
  "l"."email_verified_at",
  "l"."password_reset_at"
FROM "logins_old" AS "l";

ALTER TABLE "accounts" RENAME CONSTRAINT "pk_accounts" TO "pk_accounts_old";
ALTER TABLE "accounts" DROP CONSTRAINT "uq_accounts_stripe_customer_id";
ALTER TABLE "accounts" DROP CONSTRAINT "uq_accounts_stripe_subscription_id";
ALTER TABLE "accounts" RENAME TO "accounts_old";

CREATE TABLE "accounts" (
  "account_id"                      VARCHAR(32) NOT NULL,
  "timezone"                        VARCHAR(50) NOT NULL DEFAULT 'UTC',
  "locale"                          VARCHAR(50) NOT NULL,
  "stripe_customer_id"              VARCHAR(250),
  "stripe_subscription_id"          VARCHAR(250),
  "subscription_active_until"       TIMESTAMP WITH TIME ZONE,
  "stripe_webhook_latest_timestamp" TIMESTAMP WITH TIME ZONE,
  "subscription_status"             VARCHAR(50),
  "trial_ends_at"                   TIMESTAMP WITH TIME ZONE,
  "created_at"                      TIMESTAMP WITH TIME ZONE NOT NULL,
  CONSTRAINT "pk_accounts" PRIMARY KEY ("account_id"),
  CONSTRAINT "uq_accounts_stripe_customer_id" UNIQUE ("stripe_customer_id"),
  CONSTRAINT "uq_accounts_stripe_subscription_id" UNIQUE ("stripe_subscription_id")
);

INSERT INTO "accounts" ("account_id", "timezone", "locale", "stripe_customer_id", "stripe_subscription_id", "stripe_webhook_latest_timestamp", "subscription_status", "trial_ends_at", "created_at")
SELECT
  "a"."account_id_new",
  "a"."timezone",
  "a"."locale",
  "a"."stripe_customer_id",
  "a"."stripe_subscription_id",
  "a"."stripe_webhook_latest_timestamp",
  "a"."subscription_status",
  "a"."trial_ends_at",
  "a"."created_at"
FROM "accounts_old" AS "a";

ALTER TABLE "users" RENAME CONSTRAINT "pk_users" TO "pk_users_old";
ALTER TABLE "users" DROP CONSTRAINT "uq_users_login_id_account_id";
ALTER TABLE "users" DROP CONSTRAINT "fk_users_accounts_account_id";
ALTER TABLE "users" DROP CONSTRAINT "fk_users_logins_login_id";
ALTER TABLE "users" RENAME TO "users_old";

CREATE TABLE "users" (
  "user_id"            VARCHAR(32) NOT NULL,
  "login_id"           VARCHAR(32) NOT NULL,
  "account_id"         VARCHAR(32) NOT NULL,
  "stripe_customer_id" TEXT,
  CONSTRAINT "pk_users" PRIMARY KEY ("user_id"),
  CONSTRAINT "uq_users_login_account" UNIQUE ("login_id", "account_id"),
  CONSTRAINT "fk_users_login" FOREIGN KEY ("login_id") REFERENCES "logins" ("login_id"),
  CONSTRAINT "fk_users_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id")
);

INSERT INTO "users" ("user_id", "login_id", "account_id", "stripe_customer_id")
SELECT 
  "u"."user_id_new",
  "l"."login_id_new",
  "a"."account_id_new",
  "u"."stripe_customer_id"
FROM "users_old" AS "u"
INNER JOIN "logins_old" AS "l" ON "l"."login_id" = "u"."login_id"
INNER JOIN "accounts_old" AS "a" ON "a"."account_id" = "u"."account_id";

ALTER TABLE "betas" RENAME CONSTRAINT "pk_betas" TO "pk_betas_old";
ALTER TABLE "betas" DROP CONSTRAINT "uq_betas_code_hash";
ALTER TABLE "betas" DROP CONSTRAINT "fk_betas_used_by";
ALTER TABLE "betas" RENAME TO "betas_old";

CREATE TABLE "betas" (
  "beta_id"    VARCHAR(32) NOT NULL,
  "code_hash"  TEXT NOT NULL,
  "used_by"    VARCHAR(32),
  "expires_at" TIMESTAMP WITH TIME ZONE NOT NULL,
  CONSTRAINT "pk_betas" PRIMARY KEY ("beta_id"),
  CONSTRAINT "uq_betas_code_hash" UNIQUE ("code_hash"),
  CONSTRAINT "fk_betas_used_by" FOREIGN KEY ("used_by") REFERENCES "users" ("user_id")
);

INSERT INTO "betas" ("beta_id", "code_hash", "used_by", "expires_at")
SELECT
  "b"."beta_id_new",
  "b"."code_hash",
  "u"."user_id_new",
  "b"."expires_at"
FROM "betas_old" AS "b"
INNER JOIN "users_old" AS "u" ON "u"."user_id" = "b"."used_by_user_id";

ALTER TABLE "secrets" RENAME CONSTRAINT "pk_secrets" TO "pk_secrets_old";
ALTER TABLE "secrets" DROP CONSTRAINT "fk_secrets_account";
ALTER TABLE "secrets" RENAME TO "secrets_old";

CREATE TABLE "secrets" (
  "secret_id"  VARCHAR(32) NOT NULL,
  "account_id" VARCHAR(32) NOT NULL,
  "kind"       VARCHAR(100) NOT NULL,
  "key_id"     TEXT,
  "version"    TEXT,
  "secret"     TEXT NOT NULL,
  "updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  CONSTRAINT "pk_secrets" PRIMARY KEY ("secret_id", "account_id"),
  CONSTRAINT "fk_secrets_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id")
);

INSERT INTO "secrets" ("secret_id", "account_id", "kind", "key_id", "version", "secret", "updated_at", "created_at")
SELECT
  "s"."secret_id_new",
  "a"."account_id_new",
  "s"."kind",
  "s"."key_id",
  "s"."version",
  "s"."secret",
  "s"."updated_at",
  "s"."created_at"
FROM "secrets_old" AS "s"
INNER JOIN "accounts_old" AS "a" ON "a"."account_id" = "s"."account_id";

ALTER TABLE "plaid_links" RENAME CONSTRAINT "plaid_links_pkey" TO "plid_links_pkey_old";
ALTER TABLE "plaid_links" DROP CONSTRAINT "uq_plaid_links_item_id";
ALTER TABLE "plaid_links" DROP CONSTRAINT "fk_plaid_links_account";
ALTER TABLE "plaid_links" DROP CONSTRAINT "fk_plaid_links_secret";
ALTER TABLE "plaid_links" RENAME TO "plaid_links_old";

CREATE TABLE "plaid_links" (
  "plaid_link_id"          VARCHAR(32) NOT NULL,
  "account_id"             VARCHAR(32) NOT NULL,
  "secret_id"              VARCHAR(32) NOT NULL,
  "item_id"                TEXT NOT NULL,
  "products"               TEXT[] NOT NULL,
  "status"                 INT NOT NULL DEFAULT 0,
  "error_code"             TEXT,
  "expiration_date"        TIMESTAMP WITH TIME ZONE,
  "new_accounts_available" BOOLEAN NOT NULL,
  "webhook_url"            TEXT,
  "institution_id"         TEXT NOT NULL,
  "institution_name"       TEXT NOT NULL,
  "last_manual_sync"       TIMESTAMP WITH TIME ZONE,
  "last_successful_update" TIMESTAMP WITH TIME ZONE,
  "last_attempted_update"  TIMESTAMP WITH TIME ZONE,
  "updated_at"             TIMESTAMP WITH TIME ZONE NOT NULL,
  "created_at"             TIMESTAMP WITH TIME ZONE NOT NULL,
  "created_by"             VARCHAR(32) NOT NULL,
  CONSTRAINT "pk_plaid_links" PRIMARY KEY ("plaid_link_id", "account_id"),
  CONSTRAINT "uq_plaid_links_item_id" UNIQUE ("item_id"),
  CONSTRAINT "fk_plaid_links_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_plaid_links_secret" FOREIGN KEY ("secret_id", "account_id") REFERENCES "secrets" ("secret_id", "account_id"),
  CONSTRAINT "fk_plaid_links_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("user_id")
);

INSERT INTO "plaid_links" ("plaid_link_id", "account_id", "secret_id", "item_id", "products", "status", "error_code", "expiration_date", "new_accounts_available", "webhook_url", "institution_id", "institution_name", "last_manual_sync", "last_successful_update", "last_attempted_update", "updated_at", "created_at", "created_by")
SELECT
  "p"."plaid_link_id_new",
  "a"."account_id_new",
  "s"."secret_id_new",
  "p"."item_id",
  "p"."products",
  "p"."status",
  "p"."error_code",
  "p"."expiration_date",
  "p"."new_accounts_available",
  "p"."webhook_url",
  "p"."institution_id",
  "p"."institution_name",
  "p"."last_manual_sync",
  "p"."last_successful_update",
  "p"."last_attempted_update",
  "p"."updated_at",
  "p"."created_at",
  "u"."user_id_new"
FROM "plaid_links_old" AS "p"
INNER JOIN "accounts_old" AS "a" ON "a"."account_id" = "p"."account_id"
INNER JOIN "secrets_old" AS "s" ON "s"."secret_id" = "p"."secret_id" AND "s"."account_id" = "a"."account_id"
INNER JOIN "users_old" AS "u" ON "u"."user_id" = "p"."created_by_user_id"; 

ALTER TABLE "links" RENAME CONSTRAINT "pk_links" TO "pk_links_old";
ALTER TABLE "links" DROP CONSTRAINT "fk_links_accounts_account_id";
ALTER TABLE "links" DROP CONSTRAINT "fk_links_plaid_link";
ALTER TABLE "links" DROP CONSTRAINT "fk_links_teller_link";
ALTER TABLE "links" DROP CONSTRAINT "fk_links_users_created_by_user_id";
ALTER TABLE "links" RENAME TO "links_old";

CREATE TABLE "links" (
  "link_id"          VARCHAR(32) NOT NULL,
  "account_id"       VARCHAR(32) NOT NULL,
  "link_type"        SMALLINT NOT NULL,
  "plaid_link_id"    VARCHAR(32),
  "institution_name" VARCHAR(200),
  "description"      VARCHAR(500),
  "created_at"       TIMESTAMP WITH TIME ZONE NOT NULL,
  "created_by"       VARCHAR(32) NOT NULL,
  "updated_at"       TIMESTAMP WITH TIME ZONE NOT NULL,
  "deleted_at"       TIMESTAMP WITH TIME ZONE,
  CONSTRAINT "pk_links" PRIMARY KEY ("link_id", "account_id"),
  CONSTRAINT "fk_links_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_links_plaid_link" FOREIGN KEY ("plaid_link_id", "account_id") REFERENCES "plaid_links" ("plaid_link_id", "account_id"),
  CONSTRAINT "fk_links_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("user_id")
);

INSERT INTO "links" ("link_id", "account_id", "link_type", "plaid_link_id", "institution_name", "description", "created_at", "created_by", "updated_at", "deleted_at")
SELECT
  "l"."link_id_new",
  "a"."account_id_new",
  "l"."link_type",
  NULL,
  "l"."institution_name",
  "l"."description",
  "l"."created_at",
  "u"."user_id_new",
  "l"."updated_at",
  "l"."deleted_at"
FROM "links_old" AS "l"
INNER JOIN "accounts_old" AS "a" ON "a"."account_id" = "l"."account_id"
INNER JOIN "users_old" AS "u" ON "u"."user_id" = "l"."created_by_user_id"
WHERE "l"."plaid_link_id" IS NULL AND "l"."teller_link_id" IS NUll;

INSERT INTO "links" ("link_id", "account_id", "link_type", "plaid_link_id", "institution_name", "description", "created_at", "created_by", "updated_at", "deleted_at")
SELECT
  "l"."link_id_new",
  "a"."account_id_new",
  "l"."link_type",
  "p"."plaid_link_id_new",
  "l"."institution_name",
  "l"."description",
  "l"."created_at",
  "u"."user_id_new",
  "l"."updated_at",
  "l"."deleted_at"
FROM "links_old" AS "l"
INNER JOIN "accounts_old" AS "a" ON "a"."account_id" = "l"."account_id"
INNER JOIN "users_old" AS "u" ON "u"."user_id" = "l"."created_by_user_id"
INNER JOIN "plaid_links_old" AS "p" ON "p"."plaid_link_id" = "l"."plaid_link_id"
WHERE "l"."plaid_link_id" IS NOT NULL AND "l"."teller_link_id" IS NUll;

ALTER TABLE "plaid_bank_accounts" RENAME CONSTRAINT "pk_plaid_bank_accounts" TO "pk_plaid_bank_accounts_old";
ALTER TABLE "plaid_bank_accounts" DROP CONSTRAINT "fk_plaid_bank_accounts_account";
ALTER TABLE "plaid_bank_accounts" DROP CONSTRAINT "fk_plaid_bank_accounts_plaid_link";
ALTER TABLE "plaid_bank_accounts" DROP CONSTRAINT "fk_plaid_bank_accounts_users_created_by_user_id";
ALTER TABLE "plaid_bank_accounts" RENAME TO "plaid_bank_accounts_old";

CREATE TABLE "plaid_bank_accounts" (
  "plaid_bank_account_id" VARCHAR(32) NOT NULL,
  "account_id"            VARCHAR(32) NOT NULL,
  "plaid_link_id"         VARCHAR(32) NOT NULL,
  "plaid_id"              TEXT NOT NULL,
  "name"                  VARCHAR(200) NOT NULL,
  "official_name"         VARCHAR(200),
  "mask"                  VARCHAR(50),
  "available_balance"     BIGINT NOT NULL,
  "current_balance"       BIGINT NOT NULL,
  "limit_balance"         BIGINT,
  "created_at"            TIMESTAMP WITH TIME ZONE NOT NULL,
  "created_by"            VARCHAR(32) NOT NULL,
  CONSTRAINT "pk_plaid_bank_accounts" PRIMARY KEY ("plaid_bank_account_id", "account_id"),
  CONSTRAINT "fk_plaid_bank_accounts_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_plaid_bank_accounts_plaid_link" FOREIGN KEY ("plaid_link_id", "account_id") REFERENCES "plaid_links" ("plaid_link_id", "account_id"),
  CONSTRAINT "fk_plaid_bank_accounts_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("user_id")
);

INSERT INTO "plaid_bank_accounts" ("plaid_bank_account_id", "account_id", "plaid_link_id", "plaid_id", "name", "official_name", "mask", "available_balance", "current_balance", "limit_balance", "created_at", "created_by")
SELECT
  "b"."plaid_bank_account_id_new",
  "a"."account_id_new",
  "p"."plaid_link_id_new",
  "b"."plaid_id",
  "b"."name",
  "b"."official_name",
  "b"."mask",
  "b"."available_balance",
  "b"."current_balance",
  "b"."limit_balance",
  "b"."created_at",
  "u"."user_id_new"
FROM "plaid_bank_accounts_old" AS "b"
INNER JOIN "accounts_old" AS "a" ON "a"."account_id" = "b"."account_id"
INNER JOIN "plaid_links_old" AS "p" ON "p"."plaid_link_id" = "b"."plaid_link_id"
INNER JOIN "users_old" AS "u" ON "u"."user_id" = "b"."created_by_user_id";

ALTER TABLE "bank_accounts" RENAME CONSTRAINT "pk_bank_accounts" TO "pk_bank_accounts_old";
ALTER TABLE "bank_accounts" DROP CONSTRAINT "fk_bank_accounts_accounts_account_id";
ALTER TABLE "bank_accounts" DROP CONSTRAINT "fk_bank_accounts_links_link_id_account_id";
ALTER TABLE "bank_accounts" DROP CONSTRAINT "fk_bank_accounts_plaid_bank_accounts";
ALTER TABLE "bank_accounts" DROP CONSTRAINT "fk_bank_accounts_teller_bank_account";
ALTER TABLE "bank_accounts" RENAME TO "bank_accounts_old";

CREATE TABLE "bank_accounts" (
  "bank_account_id"       VARCHAR(32) NOT NULL,
  "account_id"            VARCHAR(32) NOT NULL,
  "link_id"               VARCHAR(32) NOT NULL,
  "plaid_bank_account_id" VARCHAR(32),
  "name"                  VARCHAR(200) NOT NULL,
  "mask"                  VARCHAR(50),
  "account_type"          VARCHAR(200),
  "account_sub_type"      VARCHAR(200),
  "status"                VARCHAR(100) NOT NULL DEFAULT 'active',
  "available_balance"     BIGINT NOT NULL,
  "current_balance"       BIGINT NOT NULL,
  "last_updated"          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() AT TIME ZONE 'UTC'), -- TODO Do we even need this anymore?
  "created_at"            TIMESTAMP WITH TIME ZONE NOT NULL,
  "updated_at"            TIMESTAMP WITH TIME ZONE NOT NULL,
  CONSTRAINT "pk_bank_accounts" PRIMARY KEY ("bank_account_id", "account_id"),
  CONSTRAINT "fk_bank_accounts_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_bank_accounts_link" FOREIGN KEY ("link_id", "account_id") REFERENCES "links" ("link_id", "account_id"),
  CONSTRAINT "fk_bank_accounts_plaid_bank_account" FOREIGN KEY ("plaid_bank_account_id", "account_id") REFERENCES "plaid_bank_accounts" ("plaid_bank_account_id", "account_id")
);

INSERT INTO "bank_accounts" ("bank_account_id", "account_id", "link_id", "plaid_bank_account_id", "name", "mask", "account_type", "account_sub_type", "status", "available_balance", "current_balance", "last_updated", "created_at", "updated_at")
SELECT
  "b"."bank_account_id_new",
  "a"."account_id_new",
  "l"."link_id_new",
  "p"."plaid_bank_account_id_new",
  "b"."name",
  "b"."mask",
  "b"."account_type",
  "b"."account_sub_type",
  "b"."status",
  "b"."available_balance",
  "b"."current_balance",
  "b"."last_updated",
  "b"."created_at",
  "b"."updated_at"
FROM "bank_accounts_old" AS "b"
INNER JOIN "accounts_old" AS "a" ON "a"."account_id" = "b"."account_id"
INNER JOIN "links_old" AS "l" ON "l"."link_id" = "b"."link_id"
LEFT JOIN "plaid_bank_accounts_old" AS "p" ON "p"."plaid_bank_account_id" = "b"."plaid_bank_account_id"
WHERE "b"."teller_bank_account_id" IS NULL;

ALTER TABLE "funding_schedules" RENAME CONSTRAINT "pk_funding_schedules" TO "pk_funding_schedules_old";
ALTER TABLE "funding_schedules" DROP CONSTRAINT "uq_funding_schedules_bank_account_id_name";
ALTER TABLE "funding_schedules" DROP CONSTRAINT "fk_funding_schedules_accounts_account_id";
ALTER TABLE "funding_schedules" DROP CONSTRAINT "fk_funding_schedules_bank_accounts_bank_account_id_account_id";
ALTER TABLE "funding_schedules" RENAME TO "funding_schedules_old";

CREATE TABLE "funding_schedules" (
  "funding_schedule_id"      VARCHAR(32) NOT NULL,
  "account_id"               VARCHAR(32) NOT NULL,
  "bank_account_id"          VARCHAR(32) NOT NULL,
  "name"                     VARCHAR(200) NOT NULL,
  "description"              VARCHAR(500) NOT NULL,
  "ruleset"                  TEXT NOT NULL,
  "last_occurrence"          TIMESTAMP WITH TIME ZONE,
  "next_occurrence"          TIMESTAMP WITH TIME ZONE NOT NULL,
  "next_occurrence_original" TIMESTAMP WITH TIME ZONE NOT NULL,
  "exclude_weekends"         BOOLEAN NOT NULL DEFAULT false,
  "wait_for_deposit"         BOOLEAN NOT NULL DEFAULT false,
  "estimated_deposit"        BIGINT,
  CONSTRAINT "pk_funding_schedules" PRIMARY KEY ("funding_schedule_id", "account_id", "bank_account_id"),
  CONSTRAINT "uq_funding_schedules_name" UNIQUE ("bank_account_id", "name"),
  CONSTRAINT "fk_funding_schedules_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_funding_schedules_bank_account" FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id")
);

INSERT INTO "funding_schedules" ("funding_schedule_id", "account_id", "bank_account_id", "name", "description", "ruleset", "last_occurrence", "next_occurrence", "next_occurrence_original", "exclude_weekends", "wait_for_deposit", "estimated_deposit")
SELECT
  "f"."funding_schedule_id_new",
  "a"."account_id_new",
  "b"."bank_account_id_new",
  "f"."name",
  "f"."description",
  "f"."ruleset",
  "f"."last_occurrence",
  "f"."next_occurrence",
  "f"."next_occurrence_original",
  "f"."exclude_weekends",
  "f"."wait_for_deposit",
  "f"."estimated_deposit"
FROM "funding_schedules_old" AS "f"
INNER JOIN "accounts_old" AS "a" ON "a"."account_id" = "f"."account_id"
INNER JOIN "bank_accounts_old" AS "b" ON "b"."bank_account_id" = "f"."bank_account_id";
