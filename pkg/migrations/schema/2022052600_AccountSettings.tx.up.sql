CREATE TABLE settings
(
    account_id        BIGINT NOT NULL PRIMARY KEY,
    max_safe_to_spend JSONB  NOT NULL,
    CONSTRAINT fk_settings_accounts FOREIGN KEY (account_id) REFERENCES accounts (account_id) ON DELETE CASCADE
);
