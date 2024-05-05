ALTER TABLE "jobs"
ADD COLUMN "job_id_new" VARCHAR(64);

ALTER TABLE "logins"
ADD COLUMN "login_id_new" VARCHAR(64);

ALTER TABLE "accounts"
ADD COLUMN "account_id_new" VARCHAR(64);

ALTER TABLE "users"
ADD COLUMN "user_id_new" VARCHAR(64);

ALTER TABLE "betas"
ADD COLUMN "beta_id_new" VARCHAR(64);

ALTER TABLE "plaid_links"
ADD COLUMN "plaid_link_id_new" VARCHAR(64);

ALTER TABLE "links"
ADD COLUMN "link_id_new" VARCHAR(64);

ALTER TABLE "secrets"
ADD COLUMN "secret_id_new" VARCHAR(64);

ALTER TABLE "plaid_bank_accounts"
ADD COLUMN "plaid_bank_account_id_new" VARCHAR(64);

ALTER TABLE "bank_accounts"
ADD COLUMN "bank_account_id_new" VARCHAR(64);

ALTER TABLE "spending"
ADD COLUMN "spending_id_new" VARCHAR(64);

ALTER TABLE "funding_schedules"
ADD COLUMN "funding_schedule_id_new" VARCHAR(64);

ALTER TABLE "files"
ADD COLUMN "file_id_new" VARCHAR(64);

ALTER TABLE "plaid_transactions"
ADD COLUMN "plaid_transaction_id_new" VARCHAR(64);

ALTER TABLE "transaction_clusters"
ADD COLUMN "transaction_cluster_id_new" VARCHAR(64);

ALTER TABLE "transactions"
ADD COLUMN "transaction_id_new" VARCHAR(64);

ALTER TABLE "plaid_syncs"
ADD COLUMN "plaid_sync_id_new" VARCHAR(64);

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

ALTER TABLE "jobs" RENAME CONSTRAINT "pk_jobs" TO "pk_jobs_old";
ALTER TABLE "jobs" DROP CONSTRAINT "uq_jobs_signature";
ALTER TABLE "jobs" RENAME TO "jobs_old";

CREATE TABLE "jobs" (
  "job_id"       VARCHAR(32) NOT NULL,
  "queue"        VARCHAR(200) NOT NULL,
  "signature"    VARCHAR(100) NOT NULL,
  "status"       VARCHAR(50) NOT NULL,
  "input"        JSONB,
  "output"       JSONB,
  "created_at"   TIMESTAMP WITH TIME ZONE NOT NULL,
  "updated_at"   TIMESTAMP WITH TIME ZONE NOT NULL,
  "started_at"   TIMESTAMP WITH TIME ZONE,
  "completed_at" TIMESTAMP WITH TIME ZONE,
  CONSTRAINT "pk_jobs" PRIMARY KEY ("job_id"),
  CONSTRAINT "uq_jobs_signature" UNIQUE ("signature")
);

INSERT INTO "jobs" ("job_id", "queue", "signature", "status", "input", "output", "created_at", "updated_at", "started_at", "completed_at")
SELECT
  "j"."job_id_new",
  "j"."queue",
  "j"."signature",
  "j"."status",
  "j"."input",
  "j"."output",
  "j"."created_at",
  "j"."updated_at",
  "j"."started_at",
  "j"."completed_at"
FROM "jobs_old" AS "j";

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

ALTER TABLE "files" RENAME CONSTRAINT "pk_files" TO "pk_files_old";
ALTER TABLE "files" DROP CONSTRAINT "fk_files_account";
ALTER TABLE "files" DROP CONSTRAINT "fk_files_bank_account";
ALTER TABLE "files" DROP CONSTRAINT "fk_files_users_created_by_user_id";
ALTER TABLE "files" RENAME TO "files_old";

CREATE TABLE "files" (
  "file_id"       VARCHAR(32) NOT NULL,
  "account_id"    VARCHAR(32) NOT NULL,
  "content_type"  VARCHAR(200) NOT NULL,
  "name"          VARCHAR(200) NOT NULL,
  "size"          INTEGER NOT NULL,
  "blob_uri"      TEXT NOT NULL,
  "created_at"    TIMESTAMP WITH TIME ZONE NOT NULL,
  "created_by"    VARCHAR(32) NOT NULL,
  "deleted_at"    TIMESTAMP WITH TIME ZONE NOT NULL,
  "reconciled_at" TIMESTAMP WITH TIME ZONE,
  CONSTRAINT "pk_files" PRIMARY KEY ("file_id", "account_id"),
  CONSTRAINT "fk_files_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_files_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("user_id")
);

INSERT INTO "files" ("file_id", "account_id", "content_type", "name", "size", "blob_uri", "created_at", "created_by", "deleted_at", "reconciled_at")
SELECT
  "f"."file_id_new",
  "a"."account_id_new",
  "f"."content_type",
  "f"."name",
  "f"."size",
  "f"."object_uri",
  "f"."created_at",
  "u"."user_id_new",
  NULL,
  NULL
FROM "files_old" AS "f"
INNER JOIN "accounts_old" AS "a" ON "a"."account_id" = "f"."account_id"
INNER JOIN "users_old" AS "u" ON "u"."user_id" = "f"."created_by_user_id";

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
  "description"              VARCHAR(500),
  "ruleset"                  TEXT NOT NULL,
  "last_recurrence"          TIMESTAMP WITH TIME ZONE,
  "next_recurrence"          TIMESTAMP WITH TIME ZONE NOT NULL,
  "next_recurrence_original" TIMESTAMP WITH TIME ZONE NOT NULL,
  "exclude_weekends"         BOOLEAN NOT NULL DEFAULT false,
  "wait_for_deposit"         BOOLEAN NOT NULL DEFAULT false,
  "estimated_deposit"        BIGINT,
  CONSTRAINT "pk_funding_schedules" PRIMARY KEY ("funding_schedule_id", "account_id", "bank_account_id"),
  CONSTRAINT "uq_funding_schedules_name" UNIQUE ("bank_account_id", "name"),
  CONSTRAINT "fk_funding_schedules_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_funding_schedules_bank_account" FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id")
);

INSERT INTO "funding_schedules" ("funding_schedule_id", "account_id", "bank_account_id", "name", "description", "ruleset", "last_recurrence", "next_recurrence", "next_recurrence_original", "exclude_weekends", "wait_for_deposit", "estimated_deposit")
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

ALTER TABLE "spending" RENAME CONSTRAINT "pk_spending" TO "pk_spending_old";
ALTER TABLE "spending" DROP CONSTRAINT "uq_spending_bank_account_id_spending_type_name";
ALTER TABLE "spending" DROP CONSTRAINT "fk_spending_accounts_account_id";
ALTER TABLE "spending" DROP CONSTRAINT "fk_spending_bank_accounts_bank_account_id_account_id";
ALTER TABLE "spending" DROP CONSTRAINT "fk_spending_funding_schedules_funding_schedule_id_account_id_ba";
ALTER TABLE "spending" RENAME TO "spending_old";

CREATE TABLE "spending" (
  "spending_id"              VARCHAR(32) NOT NULL,
  "account_id"               VARCHAR(32) NOT NULL,
  "bank_account_id"          VARCHAR(32) NOT NULL,
  "funding_schedule_id"      VARCHAR(32) NOT NULL,
  "spending_type"            SMALLINT NOT NULL,
  "name"                     VARCHAR(200) NOT NULL,
  "description"              VARCHAR(500),
  "ruleset"                  TEXT,
  "target_amount"            BIGINT NOT NULL,
  "current_amount"           BIGINT NOT NULL,
  "used_amount"              BIGINT NOT NULL,
  "last_recurrence"          TIMESTAMP WITH TIME ZONE,
  "next_recurrence"          TIMESTAMP WITH TIME ZONE NOT NULL,
  "last_spent_from"          TIMESTAMP WITH TIME ZONE,
  "next_contribution_amount" BIGINT NOT NULL,
  "is_behind"                BOOLEAN NOT NULL,
  "is_paused"                BOOLEAN NOT NULL,
  "created_at"               TIMESTAMP WITH TIME ZONE NOT NULL,
  CONSTRAINT "pk_spending" PRIMARY KEY ("spending_id", "account_id", "bank_account_id"),
  CONSTRAINT "uq_spending_type_name" UNIQUE ("bank_account_id", "spending_type", "name"),
  CONSTRAINT "fk_spending_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_spending_bank_account" FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id"),
  CONSTRAINT "fk_spending_funding_schedule" FOREIGN KEY ("funding_schedule_id", "account_id", "bank_account_id") REFERENCES "funding_schedules" ("funding_schedule_id", "account_id", "bank_account_id")
);

INSERT INTO "spending" ("spending_id", "account_id", "bank_account_id", "funding_schedule_id", "spending_type", "name", "description", "ruleset", "target_amount", "current_amount", "used_amount", "last_recurrence", "next_recurrence", "last_spent_from", "next_contribution_amount", "is_behind", "is_paused", "created_at")
SELECT
  "s"."spending_id_new",
  "a"."account_id_new",
  "b"."bank_account_id_new",
  "f"."funding_schedule_id_new",
  "s"."spending_type",
  "s"."name",
  "s"."description",
  "s"."ruleset",
  "s"."target_amount",
  "s"."current_amount",
  "s"."used_amount",
  "s"."last_recurrence",
  "s"."next_recurrence",
  "s"."last_spent_from",
  "s"."next_contribution_amount",
  "s"."is_behind",
  "s"."is_paused",
  "s"."date_created"
FROM "spending_old" AS "s"
INNER JOIN "accounts_old" AS "a" ON "a"."account_id" = "s"."account_id"
INNER JOIN "bank_accounts_old" AS "b" ON "b"."bank_account_id" = "s"."bank_account_id"
INNER JOIN "funding_schedules_old" AS "f" ON "f"."funding_schedule_id" = "s"."funding_schedule_id";

ALTER TABLE "plaid_transactions" RENAME CONSTRAINT "pk_plaid_transactions" TO "pk_plaid_transactions_old";
ALTER TABLE "plaid_transactions" DROP CONSTRAINT "fk_plaid_transactions_account";
ALTER TABLE "plaid_transactions" DROP CONSTRAINT "fk_plaid_transactions_plaid_bank_account";
ALTER TABLE "plaid_transactions" RENAME TO "plaid_transactions_old";

CREATE TABLE "plaid_transactions" (
  "plaid_transaction_id"  VARCHAR(32) NOT NULL,
  "account_id"            VARCHAR(32) NOT NULL,
  "plaid_bank_account_id" VARCHAR(32) NOT NULL,
  "plaid_id"              TEXT NOT NULL,
  "pending_plaid_id"      TEXT,
  "categories"            TEXT[],
  "date"                  TIMESTAMP WITH TIME ZONE NOT NULL,
  "authorized_date"       TIMESTAMP WITH TIME ZONE,
  "name"                  VARCHAR(200) NOT NULL,
  "merchant_name"         VARCHAR(200),
  "amount"                BIGINT NOT NULL,
  "currency"              VARCHAR(50) NOT NULL,
  "is_pending"            BOOLEAN NOT NULL,
  "created_at"            TIMESTAMP WITH TIME ZONE NOT NULL,
  "deleted_at"            TIMESTAMP WITH TIME ZONE,
  CONSTRAINT "pk_plaid_transactions" PRIMARY KEY ("plaid_transaction_id", "account_id"),
  CONSTRAINT "fk_plaid_transactions_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_plaid_transactions_plaid_bank_account" FOREIGN KEY ("plaid_bank_account_id", "account_id") REFERENCES "plaid_bank_accounts" ("plaid_bank_account_id", "account_id")
);

INSERT INTO "plaid_transactions" ("plaid_transaction_id", "account_id", "plaid_bank_account_id", "plaid_id", "pending_plaid_id", "categories", "date", "authorized_date", "name", "merchant_name", "amount", "currency", "is_pending", "created_at", "deleted_at")
SELECT
  "t"."plaid_transaction_id_new",
  "a"."account_id_new",
  "b"."plaid_bank_account_id_new",
  "t"."plaid_id",
  "t"."pending_plaid_id",
  "t"."categories",
  "t"."date",
  "t"."authorized_date",
  "t"."name",
  "t"."merchant_name",
  "t"."amount",
  "t"."currency",
  "t"."is_pending",
  "t"."created_at",
  "t"."deleted_at"
FROM "plaid_transactions_old" AS "t"
INNER JOIN "accounts_old" AS "a" ON "a"."account_id" = "t"."account_id"
INNER JOIN "plaid_bank_accounts_old" AS "b" ON "b"."plaid_bank_account_id" = "t"."plaid_bank_account_id";

ALTER TABLE "transaction_clusters" RENAME CONSTRAINT "pk_transaction_clusters" TO "pk_transaction_clusters_old";
ALTER TABLE "transaction_clusters" DROP CONSTRAINT "fk_transaction_clusters_account";
ALTER TABLE "transaction_clusters" DROP CONSTRAINT "fk_transaction_clusters_bank_account";
DROP INDEX "ix_transaction_clusters_bank_account";
DROP INDEX "ix_transaction_clusters_members";
ALTER TABLE "transaction_clusters" RENAME TO "transaction_clusters_old";

-- Don't backfill this table, it will need to be regenerated
CREATE TABLE "transaction_clusters" (
  "transaction_cluster_id" VARCHAR(32) NOT NULL,
  "account_id"             VARCHAR(32) NOT NULL,
  "bank_account_id"        VARCHAR(32) NOT NULL,
  "name"                   VARCHAR(200) NOT NULL,
  "members"                VARCHAR(32)[] NOT NULL,
  "created_at"             TIMESTAMP WITH TIME ZONE NOT NULL,
  CONSTRAINT "pk_transaction_clusters" PRIMARY KEY ("transaction_cluster_id", "account_id"),
  CONSTRAINT "fk_transaction_clusters_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_transaction_clusters_bank_account" FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id")
);

-- For querying by the members contents.
CREATE INDEX "ix_transaction_clusters_members" ON "transaction_clusters" USING GIN ("members");
-- For narrowing down the results to a single bank account.
CREATE INDEX "ix_transaction_clusters_bank_account" ON "transaction_clusters" ("account_id", "bank_account_id");

ALTER TABLE "plaid_syncs" RENAME CONSTRAINT "plaid_syncs_pkey" TO "plaid_syncs_pkey_old";
ALTER TABLE "plaid_syncs" DROP CONSTRAINT "fk_plaid_syncs_account";
ALTER TABLE "plaid_syncs" DROP CONSTRAINT "fk_plaid_syncs_plaid_link";
DROP INDEX "ix_plaid_syncs_timestamp";
ALTER TABLE "plaid_syncs" RENAME TO "plaid_syncs_old";

CREATE TABLE "plaid_syncs" (
  "plaid_sync_id" VARCHAR(32) NOT NULL,
  "account_id" VARCHAR(32) NOT NULL,
  "plaid_link_id" VARCHAR(32) NOT NULL,
  "timestamp" TIMESTAMP WITH TIME ZONE NOT NULL,
  "trigger" VARCHAR(50) NOT NULL,
  "cursor" VARCHAR(500) NOT NULL,
  "added" INTEGER NOT NULL,
  "modified" INTEGER NOT NULL,
  "removed" INTEGER NOT NULL,
  CONSTRAINT "pk_plaid_syncs" PRIMARY KEY ("plaid_sync_id", "account_id"),
  CONSTRAINT "fk_plaid_syncs_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_plaid_syncs_plaid_link" FOREIGN KEY ("plaid_link_id", "account_id") REFERENCES "plaid_links" ("plaid_link_id", "account_id")
);

CREATE INDEX "ix_plaid_syncs_timestamp" ON "plaid_syncs" ("plaid_link_id", "timestamp" DESC);

INSERT INTO "plaid_syncs" ("plaid_sync_id", "account_id", "plaid_link_id", "timestamp", "trigger", "cursor", "added", "modified", "removed")
SELECT
  "s"."plaid_sync_id_new",
  "a"."account_id_new",
  "p"."plaid_link_id_new",
  "s"."timestamp",
  "s"."trigger",
  "s"."cursor",
  "s"."added",
  "s"."modified",
  "s"."removed"
FROM "plaid_syncs_old" AS "s"
INNER JOIN "accounts_old" AS "a" ON "a"."account_id" = "s"."account_id"
INNER JOIN "plaid_links_old" AS "p" ON "p"."plaid_link_id" = "s"."plaid_link_id";

ALTER TABLE "transactions" RENAME CONSTRAINT "pk_transactions" TO "pk_transactions_old";
ALTER TABLE "transactions" DROP CONSTRAINT "fk_transactions_accounts_account_id";
ALTER TABLE "transactions" DROP CONSTRAINT "fk_transactions_bank_accounts_bank_account_id_account_id";
ALTER TABLE "transactions" DROP CONSTRAINT "fk_transactions_plaid_transactions";
ALTER TABLE "transactions" DROP CONSTRAINT "fk_transactions_plaid_transactions_pending";
ALTER TABLE "transactions" DROP CONSTRAINT "fk_transactions_spending";
ALTER TABLE "transactions" DROP CONSTRAINT "fk_transactions_teller_transaction";
DROP INDEX "ix_transactions_opt_order";
DROP INDEX "ix_transactions_soft_delete";
ALTER TABLE "transactions" RENAME TO "transactions_old";

CREATE TABLE "transactions" (
  "transaction_id"               VARCHAR(32) NOT NULL,
  "account_id"                   VARCHAR(32) NOT NULL,
  "bank_account_id"              VARCHAR(32) NOT NULL,
  "spending_id"                  VARCHAR(32),
  "plaid_transaction_id"         VARCHAR(32),
  "pending_plaid_transaction_id" VARCHAR(32),
  "name"                         VARCHAR(200),
  "original_name"                VARCHAR(200) NOT NULL,
  "merchant_name"                VARCHAR(200),
  "original_merchant_name"       VARCHAR(200),
  "categories"                   TEXT[],
  "amount"                       BIGINT NOT NULL,
  "spending_amount"              BIGINT,
  "is_pending"                   BOOLEAN NOT NULL,
  "date"                         TIMESTAMP WITH TIME ZONE NOT NULL,
  "created_at"                   TIMESTAMP WITH TIME ZONE NOT NULL,
  "deleted_at"                   TIMESTAMP WITH TIME ZONE,
  CONSTRAINT "pk_transactions" PRIMARY KEY ("transaction_id", "account_id", "bank_account_id"),
  CONSTRAINT "fk_transactions_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_transactions_bank_account" FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id"),
  CONSTRAINT "fk_transactions_spending" FOREIGN KEY ("spending_id", "bank_account_id", "account_id") REFERENCES "spending" ("spending_id", "bank_account_id", "account_id"),
  CONSTRAINT "fk_transactions_plaid_transaction" FOREIGN KEY ("plaid_transaction_id", "account_id") REFERENCES "plaid_transactions" ("plaid_transaction_id", "account_id"),
  CONSTRAINT "fk_transactions_pending_plaid_transaction" FOREIGN KEY ("pending_plaid_transaction_id", "account_id") REFERENCES "plaid_transactions" ("plaid_transaction_id", "account_id")
);

CREATE INDEX "ix_transactions_opt_order" ON "transactions" ("account_id", "bank_account_id", "date" DESC, "transaction_id" DESC);
CREATE INDEX "ix_transactions_soft_delete" ON "transactions" ("account_id", "bank_account_id", "date" DESC, "transaction_id" DESC) WHERE "deleted_at" IS NULL;

INSERT INTO "transactions" ("transaction_id", "account_id", "bank_account_id", "spending_id", "plaid_transaction_id", "pending_plaid_transaction_id", "name", "original_name", "merchant_name", "original_merchant_name", "categories", "amount", "spending_amount", "is_pending", "date", "created_at", "deleted_at")
SELECT
  "t"."transaction_id_new",
  "a"."account_id_new",
  "b"."bank_account_id_new",
  "s"."spending_id_new",
  "p"."plaid_transaction_id_new", -- non pending
  "pp"."plaid_transaction_id_new", -- pending
  "t"."name",
  "t"."original_name",
  "t"."merchant_name",
  "t"."original_merchant_name",
  "t"."categories",
  "t"."amount",
  "t"."spending_amount",
  "t"."is_pending",
  "t"."date",
  "t"."created_at",
  "t"."deleted_at"
FROM "transactions_old" AS "t"
INNER JOIN "accounts_old" AS "a" ON "a"."account_id" = "t"."account_id"
INNER JOIN "bank_accounts_old" AS "b" ON "b"."bank_account_id" = "t"."bank_account_id"
LEFT JOIN "spending_old" AS "s" ON "s"."spending_id" = "t"."spending_id"
LEFT JOIN "plaid_transactions_old" AS "p" ON "p"."plaid_transaction_id" = "t"."plaid_transaction_id"
LEFT JOIN "plaid_transactions_old" AS "pp" ON "pp"."plaid_transaction_id" = "t"."pending_plaid_transaction_id"
WHERE "t"."teller_transaction_id" IS NULL;

DROP VIEW "balances";
DROP TABLE "teller_syncs" CASCADE;
DROP TABLE "plaid_syncs_old" CASCADE;
DROP TABLE "transaction_clusters_old" CASCADE;
DROP TABLE "transactions_old" CASCADE;
DROP TABLE "teller_transactions" CASCADE;
DROP TABLE "plaid_transactions_old" CASCADE;
DROP TABLE "spending_old" CASCADE; 
DROP TABLE "funding_schedules_old" CASCADE;
DROP TABLE "bank_accounts_old" CASCADE;
DROP TABLE "plaid_bank_accounts_old" CASCADE;
DROP TABLE "teller_bank_accounts" CASCADE;
DROP TABLE "links_old" CASCADE;
DROP TABLE "plaid_links_old" CASCADE;
DROP TABLE "teller_links" CASCADE;
DROP TABLE "secrets_old" CASCADE;
DROP TABLE "files_old" CASCADE;
DROP TABLE "jobs_old" CASCADE;
DROP TABLE "users_old" CASCADE;
DROP TABLE "accounts_old" CASCADE;
DROP TABLE "logins_old" CASCADE;
DROP TABLE "betas_old" CASCADE;
DROP TABLE IF EXISTS "funding_stats" CASCADE;

CREATE OR REPLACE VIEW balances AS
 SELECT bank_account.bank_account_id,
    bank_account.account_id,
    bank_account.current_balance AS current,
    bank_account.available_balance AS available,
    bank_account.available_balance::numeric - sum(COALESCE(expense.current_amount, 0::numeric)) - sum(COALESCE(goal.current_amount, 0::numeric)) AS free,
    sum(COALESCE(expense.current_amount, 0::numeric)) AS expenses,
    sum(COALESCE(goal.current_amount, 0::numeric)) AS goals
   FROM bank_accounts bank_account
     LEFT JOIN ( SELECT spending.bank_account_id,
            spending.account_id,
            sum(spending.current_amount) AS current_amount
           FROM spending
          WHERE spending.spending_type = 0
          GROUP BY spending.bank_account_id, spending.account_id) expense ON expense.bank_account_id = bank_account.bank_account_id AND expense.account_id = bank_account.account_id
     LEFT JOIN ( SELECT spending.bank_account_id,
            spending.account_id,
            sum(spending.current_amount) AS current_amount
           FROM spending
          WHERE spending.spending_type = 1
          GROUP BY spending.bank_account_id, spending.account_id) goal ON goal.bank_account_id = bank_account.bank_account_id AND goal.account_id = bank_account.account_id
  GROUP BY bank_account.bank_account_id, bank_account.account_id;
