CREATE TABLE IF NOT EXISTS "logins" (
    "login_id"      bigserial NOT NULL,
    "email"         text      NOT NULL UNIQUE,
    "password_hash" text      NOT NULL,
    PRIMARY KEY ("login_id"),
    UNIQUE ("email")
);
CREATE TABLE IF NOT EXISTS "users" (
    "user_id"            bigserial NOT NULL,
    "login_id"           bigint    NOT NULL,
    "account_id"         bigint    NOT NULL,
    "stripe_customer_id" text,
    "first_name"         text      NOT NULL,
    "last_name"          text,
    PRIMARY KEY ("user_id"),
    FOREIGN KEY ("login_id") REFERENCES "logins" ("login_id") ON DELETE CASCADE,
    FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS "accounts" (
    "account_id"      bigserial NOT NULL,
    "billing_user_id" bigint    NOT NULL,
    "timezone"        text      NOT NULL DEFAULT 'UTC',
    PRIMARY KEY ("account_id")
);
CREATE TABLE IF NOT EXISTS "links" (
    "link_id"            bigserial NOT NULL,
    "account_id"         bigint    NOT NULL,
    "plaid_item_id"      text      NOT NULL,
    "plaid_access_token" text      NOT NULL,
    "webhook_url"        text,
    PRIMARY KEY ("link_id", "account_id"),
    FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS "bank_accounts" (
    "bank_account_id"   bigserial NOT NULL,
    "account_id"        bigserial NOT NULL,
    "link_id"           bigint    NOT NULL,
    "plaid_account_id"  text      NOT NULL,
    "available_balance" bigint    NOT NULL,
    "current_balance"   bigint    NOT NULL,
    "mask"              text      NOT NULL,
    "name"              text,
    "original_name"     text      NOT NULL,
    "official_name"     text      NOT NULL,
    "account_type"      text      NOT NULL,
    "account_sub_type"  text      NOT NULL,
    PRIMARY KEY ("bank_account_id", "account_id"),
    FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
    FOREIGN KEY ("link_id", "account_id") REFERENCES "links" ("link_id", "account_id") ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS "funding_schedules" (
    "funding_schedule_id" bigserial NOT NULL,
    "account_id"          bigint    NOT NULL,
    "bank_account_id"     bigint    NOT NULL,
    "name"                text      NOT NULL,
    "description"         text,
    "rule"                text      NOT NULL,
    "last_occurrence"     date,
    "next_occurrence"     date,
    PRIMARY KEY ("funding_schedule_id", "account_id", "bank_account_id"),
    UNIQUE ("bank_account_id", "name"),
    FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id") ON DELETE CASCADE,
    FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS "expenses" (
    "expense_id"               bigserial NOT NULL,
    "account_id"               bigint    NOT NULL,
    "bank_account_id"          bigint    NOT NULL,
    "funding_schedule_id"      bigint,
    "name"                     text      NOT NULL,
    "description"              text,
    "target_amount"            bigint    NOT NULL,
    "current_amount"           bigint    NOT NULL,
    "recurrence_rule"          text      NOT NULL,
    "last_recurrence"          date,
    "next_recurrence"          date      NOT NULL,
    "next_contribution_amount" bigint    NOT NULL,
    "is_behind"                boolean   NOT NULL,
    PRIMARY KEY ("expense_id", "account_id", "bank_account_id"),
    UNIQUE ("bank_account_id", "name"),
    FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id") ON DELETE CASCADE,
    FOREIGN KEY ("funding_schedule_id", "account_id",
                 "bank_account_id") REFERENCES "funding_schedules" ("funding_schedule_id",
                                                                    "account_id",
                                                                    "bank_account_id") ON DELETE SET NULL,
    FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS "transactions" (
    "transaction_id"         bigserial   NOT NULL,
    "account_id"             bigint      NOT NULL,
    "bank_account_id"        bigint      NOT NULL,
    "plaid_transaction_id"   text        NOT NULL,
    "amount"                 bigint      NOT NULL,
    "expense_id"             bigint,
    "categories"             text[],
    "original_categories"    text[],
    "date"                   date        NOT NULL,
    "authorized_date"        date,
    "name"                   text,
    "original_name"          text        NOT NULL,
    "merchant_name"          text,
    "original_merchant_name" text,
    "is_pending"             boolean     NOT NULL,
    "created_at"             timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY ("transaction_id", "account_id", "bank_account_id"),
    FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
    FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id") ON DELETE CASCADE,
    FOREIGN KEY ("expense_id", "account_id", "bank_account_id") REFERENCES "expenses" ("expense_id", "account_id", "bank_account_id") ON DELETE SET NULL
);
