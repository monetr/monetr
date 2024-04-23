ALTER TABLE "logins"
ADD COLUMN "login_id_new" VARCHAR(64);

ALTER TABLE "accounts"
ADD COLUMN "account_id_new" VARCHAR(64);

ALTER TABLE "users"
ADD COLUMN "user_id_new" VARCHAR(64),
ADD COLUMN "login_id_new" VARCHAR(64),
ADD COLUMN "account_id_new" VARCHAR(64);

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
WITH id_generator AS (
	SELECT generate_ulid('login') AS "id"
),
new_ids AS (
	SELECT "logins"."login_id", "id_generator"."id"
	FROM "logins"
	CROSS JOIN "id_generator"
)
UPDATE "logins"
SET "login_id_new" = "new_ids"."id"
FROM "new_ids"
WHERE "new_ids"."login_id" = "logins"."login_id";
