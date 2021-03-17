CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TABLE IF NOT EXISTS "logins" (
    "login_id" BIGSERIAL NOT NULL,
    "email" TEXT NOT NULL,
    "password_hash" TEXT NOT NULL,
    "phone_number" TEXT,
    "is_enabled" BOOLEAN NOT NULL,
    "is_email_verified" BOOLEAN NOT NULL,
    "is_phone_verified" BOOLEAN NOT NULL,
    CONSTRAINT "pk_logins" PRIMARY KEY ("login_id"),
    CONSTRAINT "uq_logins_email" UNIQUE ("email"),
    CONSTRAINT "fk_logins_email_verifications" FOREIGN KEY ("login_id") REFERENCES "email_verifications" ("login_id"),
    CONSTRAINT "fk_logins_phone_verifications" FOREIGN KEY ("login_id") REFERENCES "phone_verifications" ("login_id"),
    CONSTRAINT "fk_logins_users" FOREIGN KEY ("login_id") REFERENCES "users" ("login_id")
);

CREATE TABLE IF NOT EXISTS "registrations" (
    "registration_id" UUID NOT NULL DEFAULT uuid_generate_v4(),
    "login_id" BIGINT NOT NULL,
    "is_complete" BOOLEAN NOT NULL,
    "date_created" TIMESTAMPTZ NOT NULL,
    "date_expires" TIMESTAMPTZ NOT NULL,
    CONSTRAINT "pk_registrations" PRIMARY KEY ("registration_id"),
    CONSTRAINT "fk_registrations_logins" FOREIGN KEY ("login_id") REFERENCES "logins" ("login_id")
);

CREATE TABLE IF NOT EXISTS "email_verifications" (
    "email_verification_id" BIGSERIAL NOT NULL,
    "login_id" BIGINT NOT NULL,
    "email_address" TEXT NOT NULL,
    "is_verified" BOOLEAN NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "expires_at" TIMESTAMPTZ NOT NULL,
    "verified_at" TIMESTAMPTZ,
    CONSTRAINT "pk_email_verifications" PRIMARY KEY ("email_verification_id"),
    CONSTRAINT "uq_email_verifications_login_id_email_address" UNIQUE ("login_id", "email_address"),
    CONSTRAINT "fk_email_verifications_logins" FOREIGN KEY ("login_id") REFERENCES "logins" ("login_id")
);

CREATE TABLE IF NOT EXISTS "phone_verifications" (
    "phone_verification_id" BIGSERIAL NOT NULL,
    "login_id" BIGINT NOT NULL,
    "code" TEXT NOT NULL,
    "phone_number" TEXT NOT NULL,
    "is_verified" BOOLEAN NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "expires_at" TIMESTAMPTZ NOT NULL,
    "verified_at" TIMESTAMPTZ,
    CONSTRAINT "pk_phone_verifications" PRIMARY KEY ("phone_verification_id"),
    CONSTRAINT "uq_phone_verifications_login_id_code" UNIQUE ("login_id", "code"),
    CONSTRAINT "uq_phone_verifications_login_id_phone_number" UNIQUE ("login_id", "phone_number"),
    CONSTRAINT "fk_phone_verifications_logins" FOREIGN KEY ("login_id") REFERENCES "logins" ("login_id")
);

CREATE TABLE IF NOT EXISTS "accounts" (
    "account_id" BIGSERIAL NOT NULL,
    "timezone" TEXT NOT NULL DEFAULT 'UTC',
    CONSTRAINT "pk_accounts" PRIMARY KEY ("account_id")
);

CREATE TABLE IF NOT EXISTS "users" (
    "user_id" BIGSERIAL NOT NULL,
    "login_id" BIGINT NOT NULL,
    "account_id" BIGINT NOT NULL,
    "first_name" TEXT NOT NULL,
    "last_name" TEXT,
    CONSTRAINT "pk_users" PRIMARY KEY ("user_id"),
    CONSTRAINT "uq_users_login_id_account_id" UNIQUE ("login_id", "account_id"),
    CONSTRAINT "fk_users_accounts" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
    CONSTRAINT "fk_users_logins" FOREIGN KEY ("login_id") REFERENCES "logins" ("login_id")
);

CREATE TABLE IF NOT EXISTS "jobs" (
    "job_id" TEXT NOT NULL,
    "account_id" BIGINT,
    "name" TEXT NOT NULL,
    "arguments" JSONB,
    "enqueued_at" TIMESTAMPTZ NOT NULL,
    "started_at" TIMESTAMPTZ,
    "finished_at" TIMESTAMPTZ,
    "retries" BIGINT,
    CONSTRAINT "pk_jobs" PRIMARY KEY ("job_id"),
    CONSTRAINT "fk_jobs_accounts" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id")
);

CREATE TABLE IF NOT EXISTS "plaid_links" (
    "plaid_link_id" BIGSERIAL NOT NULL,
    "item_id" TEXT NOT NULL,
    "access_token" TEXT NOT NULL,
    "products" TEXT [ ],
    "webhook_url" TEXT,
    "institution_id" TEXT,
    "institution_name" TEXT,
    CONSTRAINT "pk_plaid_links" PRIMARY KEY ("plaid_link_id")
);

CREATE TABLE IF NOT EXISTS "links" (
    "link_id" BIGSERIAL NOT NULL,
    "account_id" BIGINT NOT NULL,
    "link_type" SMALLINT NOT NULL,
    "plaid_link_id" BIGINT,
    "institution_name" TEXT,
    "custom_institution_name" TEXT,
    "created_at" TIMESTAMPTZ NOT NULL,
    "created_by_user_id" BIGINT NOT NULL,
    "updated_at" TIMESTAMPTZ NOT NULL,
    "updated_by_user_id" BIGINT,
    CONSTRAINT "pk_links" PRIMARY KEY ("link_id", "account_id"),
    CONSTRAINT "fk_links_accounts" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
    CONSTRAINT "fk_links_bank_accounts" FOREIGN KEY ("link_id", "account_id") REFERENCES "bank_accounts" ("link_id", "account_id"),
    CONSTRAINT "fk_links_plaid_links" FOREIGN KEY ("plaid_link_id") REFERENCES "plaid_links" ("plaid_link_id"),
    CONSTRAINT "fk_links_users" FOREIGN KEY ("created_by_user_id") REFERENCES "users" ("user_id"),
    CONSTRAINT "fk_links_users" FOREIGN KEY ("updated_by_user_id") REFERENCES "users" ("user_id")
);

CREATE TABLE IF NOT EXISTS "bank_accounts" (
    "bank_account_id" BIGSERIAL NOT NULL,
    "account_id" BIGINT NOT NULL,
    "link_id" BIGINT NOT NULL,
    "plaid_account_id" TEXT,
    "available_balance" BIGINT NOT NULL,
    "current_balance" BIGINT NOT NULL,
    "mask" TEXT,
    "name" TEXT NOT NULL,
    "plaid_name" TEXT,
    "plaid_official_name" TEXT,
    "account_type" TEXT,
    "account_sub_type" TEXT,
    CONSTRAINT "pk_bank_accounts" PRIMARY KEY ("bank_account_id", "account_id"),
    CONSTRAINT "fk_bank_accounts_accounts" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
    CONSTRAINT "fk_bank_accounts_links" FOREIGN KEY ("link_id", "account_id") REFERENCES "links" ("link_id", "account_id")
);

CREATE TABLE IF NOT EXISTS "funding_schedules" (
    "funding_schedule_id" BIGSERIAL NOT NULL,
    "account_id" BIGINT NOT NULL,
    "bank_account_id" BIGINT NOT NULL,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "rule" TEXT NOT NULL,
    "last_occurrence" DATE,
    "next_occurrence" DATE NOT NULL,
    CONSTRAINT "pk_funding_schedules" PRIMARY KEY (
        "funding_schedule_id",
        "account_id",
        "bank_account_id"
    ),
    CONSTRAINT "uq_funding_schedules_bank_account_id_name" UNIQUE ("bank_account_id", "name"),
    CONSTRAINT "fk_funding_schedules_accounts" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
    CONSTRAINT "fk_funding_schedules_bank_accounts" FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id")
);

CREATE TABLE IF NOT EXISTS "expenses" (
    "expense_id" BIGSERIAL NOT NULL,
    "account_id" BIGINT NOT NULL,
    "bank_account_id" BIGINT NOT NULL,
    "funding_schedule_id" BIGINT NOT NULL,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "target_amount" BIGINT NOT NULL,
    "current_amount" BIGINT NOT NULL,
    "recurrence_rule" TEXT NOT NULL,
    "last_recurrence" DATE,
    "next_recurrence" DATE NOT NULL,
    "next_contribution_amount" BIGINT NOT NULL,
    "is_behind" BOOLEAN NOT NULL,
    CONSTRAINT "pk_expenses" PRIMARY KEY ("expense_id", "account_id", "bank_account_id"),
    CONSTRAINT "uq_expenses_bank_account_id_name" UNIQUE ("bank_account_id", "name"),
    CONSTRAINT "fk_expenses_accounts" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
    CONSTRAINT "fk_expenses_bank_accounts" FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id"),
    CONSTRAINT "fk_expenses_funding_schedules" FOREIGN KEY (
        "funding_schedule_id",
        "account_id",
        "bank_account_id"
    ) REFERENCES "funding_schedules" (
        "funding_schedule_id",
        "account_id",
        "bank_account_id"
    )
);

CREATE TABLE IF NOT EXISTS "transactions" (
    "transaction_id" BIGSERIAL NOT NULL,
    "account_id" BIGINT NOT NULL,
    "bank_account_id" BIGINT NOT NULL,
    "plaid_transaction_id" TEXT,
    "amount" BIGINT NOT NULL,
    "expense_id" BIGINT,
    "expense_amount" BIGINT,
    "categories" TEXT [ ],
    "original_categories" TEXT [ ],
    "date" DATE NOT NULL,
    "authorized_date" DATE,
    "name" TEXT,
    "original_name" TEXT NOT NULL,
    "merchant_name" TEXT,
    "original_merchant_name" TEXT,
    "is_pending" BOOLEAN NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT "pk_transactions" PRIMARY KEY ("transaction_id", "account_id", "bank_account_id"),
    CONSTRAINT "uq_transactions_bank_account_id_plaid_transaction_id" UNIQUE ("bank_account_id", "plaid_transaction_id"),
    CONSTRAINT "fk_transactions_accounts" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
    CONSTRAINT "fk_transactions_bank_accounts" FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id"),
    CONSTRAINT "fk_transactions_expenses" FOREIGN KEY ("expense_id", "account_id", "bank_account_id") REFERENCES "expenses" ("expense_id", "account_id", "bank_account_id")
);

INSERT INTO
    "logins" (
        "login_id",
        "email",
        "password_hash",
        "phone_number",
        "is_enabled",
        "is_email_verified",
        "is_phone_verified"
    )
VALUES
    (
        -1,
        'support@harderthanitneedstobe.com',
        '',
        DEFAULT,
        FALSE,
        FALSE,
        FALSE
    ) RETURNING "phone_number";

INSERT INTO
    "accounts" ("account_id", "timezone")
VALUES
    (-1, 'UTC');

INSERT INTO
    "users" (
        "user_id",
        "login_id",
        "account_id",
        "first_name",
        "last_name"
    )
VALUES
    (-1, -1, -1, 'System', 'Bot');
