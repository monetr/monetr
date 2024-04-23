ALTER TABLE "logins"
ADD COLUMN "login_id_new" VARCHAR(64);

ALTER TABLE "accounts"
ADD COLUMN "account_id_new" VARCHAR(64);

ALTER TABLE "users"
ADD COLUMN "user_id_new" VARCHAR(64),
ADD COLUMN "login_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64);

ALTER TABLE "betas"
ADD COLUMN "beta_id_new" VARCHAR(64),
ADD COLUMN "used_by_user_id_new" VARCHAR(64);

ALTER TABLE "links"
ADD COLUMN "link_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "plaid_link_id_new" VARCHAR(64),
ADD COLUMN "teller_link_id_new" VARCHAR(64),
ADD COLUMN "created_by_user_id_new" VARCHAR(64);

ALTER TABLE "secrets"
ADD COLUMN "secret_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64);

ALTER TABLE "bank_accounts"
ADD COLUMN "bank_account_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "link_id_new" VARCHAR(64),
ADD COLUMN "plaid_bank_account_id_new" VARCHAR(64),
ADD COLUMN "teller_bank_account_id_new" VARCHAR(64);

ALTER TABLE "transactions"
ADD COLUMN "transaction_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "bank_account_id_new" VARCHAR(64),
ADD COLUMN "plaid_transaction_id_new" VARCHAR(64),
ADD COLUMN "pending_plaid_transaction_id_new" VARCHAR(64),
ADD COLUMN "teller_transaction_id_new" VARCHAR(64),
ADD COLUMN "spending_id_new" VARCHAR(64);

ALTER TABLE "transaction_clusters"
ADD COLUMN "transaction_cluster_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "bank_account_id_new" VARCHAR(64);

ALTER TABLE "spending"
ADD COLUMN "spending_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "bank_account_id_new" VARCHAR(64),
ADD COLUMN "funding_schedule_id_new" VARCHAR(64);

ALTER TABLE "funding_schedules"
ADD COLUMN "funding_schedule_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "bank_account_id_new" VARCHAR(64);

ALTER TABLE "files"
ADD COLUMN "file_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "bank_account_id_new" VARCHAR(64);

ALTER TABLE "jobs"
ADD COLUMN "job_id_new" VARCHAR(64);

ALTER TABLE "plaid_links"
ADD COLUMN "plaid_link_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "secret_id_new" VARCHAR(64),
ADD COLUMN "created_by_user_id_new" VARCHAR(64);

ALTER TABLE "plaid_syncs"
ADD COLUMN "plaid_sync_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "plaid_link_id_new" VARCHAR(64);

ALTER TABLE "plaid_bank_accounts"
ADD COLUMN "plaid_bank_account_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "plaid_link_id_new" VARCHAR(64),
ADD COLUMN "created_by_user_id_new" VARCHAR(64);

ALTER TABLE "plaid_transactions"
ADD COLUMN "plaid_transaction_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "plaid_bank_account_id_new" VARCHAR(64);

ALTER TABLE "teller_links"
ADD COLUMN "teller_link_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "secret_id_new" VARCHAR(64),
ADD COLUMN "created_by_user_id_new" VARCHAR(64);

ALTER TABLE "teller_bank_accounts"
ADD COLUMN "teller_bank_account_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "teller_link_id_new" VARCHAR(64);

ALTER TABLE "teller_syncs"
ADD COLUMN "teller_sync_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "teller_bank_account_id_new" VARCHAR(64);

ALTER TABLE "teller_transactions"
ADD COLUMN "teller_transaction_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64),
ADD COLUMN "teller_bank_account_id_new" VARCHAR(64);

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

CREATE FUNCTION generate_ulid(kind TEXT)
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
  unix_time = (EXTRACT(EPOCH FROM CLOCK_TIMESTAMP()) * 1000)::BIGINT;
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
	SELECT "accounts"."account_id", generate_ulid('acct') AS "id"
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
	SELECT "links"."link_id", generate_ulid('link') AS "id"
	FROM "links"
)
UPDATE "links"
SET "link_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."link_id" = "links"."link_id";

-- Secrets
WITH new_ids AS (
	SELECT "secrets"."secret_id", generate_ulid('scrt') AS "id"
	FROM "secrets"
)
UPDATE "secrets"
SET "secret_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."secret_id" = "secrets"."secret_id";

-- Bank Accounts
WITH new_ids AS (
	SELECT "bank_accounts"."bank_account_id", generate_ulid('bac') AS "id"
	FROM "bank_accounts"
)
UPDATE "bank_accounts"
SET "bank_account_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."bank_account_id" = "bank_accounts"."bank_account_id";

-- Transactions
WITH new_ids AS (
	SELECT "transactions"."transaction_id", generate_ulid('txn') AS "id"
	FROM "transactions"
)
UPDATE "transactions"
SET "transaction_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."transaction_id" = "transactions"."transaction_id";

-- Transaction Clusters
WITH new_ids AS (
	SELECT "transaction_clusters"."transaction_cluster_id", generate_ulid('tcl') AS "id"
	FROM "transaction_clusters"
)
UPDATE "transaction_clusters"
SET "transaction_cluster_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."transaction_cluster_id" = "transaction_clusters"."transaction_cluster_id";

-- Spending
WITH new_ids AS (
	SELECT "spending"."spending_id", generate_ulid('spnd') AS "id"
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
	SELECT "files"."file_id", generate_ulid('file') AS "id"
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
	SELECT "plaid_links"."plaid_link_id", generate_ulid('plx') AS "id"
	FROM "plaid_links"
)
UPDATE "plaid_links"
SET "plaid_link_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."plaid_link_id" = "plaid_links"."plaid_link_id";

-- Plaid Syncs
WITH new_ids AS (
	SELECT "plaid_syncs"."plaid_sync_id", generate_ulid('psyn') AS "id"
	FROM "plaid_syncs"
)
UPDATE "plaid_syncs"
SET "plaid_sync_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."plaid_sync_id" = "plaid_syncs"."plaid_sync_id";

-- Plaid Bank Accounts
WITH new_ids AS (
	SELECT "plaid_bank_accounts"."plaid_bank_account_id", generate_ulid('pbac') AS "id"
	FROM "plaid_bank_accounts"
)
UPDATE "plaid_bank_accounts"
SET "plaid_bank_account_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."plaid_bank_account_id" = "plaid_bank_accounts"."plaid_bank_account_id";

-- Plaid Transactions
WITH new_ids AS (
	SELECT "plaid_transactions"."plaid_transaction_id", generate_ulid('ptxn') AS "id"
	FROM "plaid_transactions"
)
UPDATE "plaid_transactions"
SET "plaid_transaction_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."plaid_transaction_id" = "plaid_transactions"."plaid_transaction_id";

-- Teller Links
WITH new_ids AS (
	SELECT "teller_links"."teller_link_id", generate_ulid('tlx') AS "id"
	FROM "teller_links"
)
UPDATE "teller_links"
SET "teller_link_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."teller_link_id" = "teller_links"."teller_link_id";

-- Teller bank accounts
WITH new_ids AS (
	SELECT "teller_bank_accounts"."teller_bank_account_id", generate_ulid('tbac') AS "id"
	FROM "teller_bank_accounts"
)
UPDATE "teller_bank_accounts"
SET "teller_bank_account_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."teller_bank_account_id" = "teller_bank_accounts"."teller_bank_account_id";

-- Teller syncs
WITH new_ids AS (
	SELECT "teller_syncs"."teller_sync_id", generate_ulid('tsyn') AS "id"
	FROM "teller_syncs"
)
UPDATE "teller_syncs"
SET "teller_sync_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."teller_sync_id" = "teller_syncs"."teller_sync_id";

-- Teller transactions
WITH new_ids AS (
	SELECT "teller_transactions"."teller_transaction_id", generate_ulid('ttxn') AS "id"
	FROM "teller_transactions"
)
UPDATE "teller_transactions"
SET "teller_transaction_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."teller_transaction_id" = "teller_transactions"."teller_transaction_id";
