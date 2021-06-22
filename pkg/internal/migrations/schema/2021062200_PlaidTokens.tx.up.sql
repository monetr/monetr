CREATE TABLE "plaid_tokens" (
    item_id TEXT NOT NULL,
    account_id BIGINT NOT NULL,
    access_token TEXT NOT NULL,
    CONSTRAINT pk_plaid_tokens PRIMARY KEY (account_id, item_id),
    CONSTRAINT fk_plaid_tokens_account FOREIGN KEY (account_id) REFERENCES accounts (account_id) ON DELETE CASCADE
);

INSERT INTO "plaid_tokens"(item_id, access_token, account_id)
SELECT
    plaid_link.item_id,
    plaid_link.access_token,
    link.account_id
FROM plaid_links AS plaid_link
INNER JOIN links AS link ON link.plaid_link_id = plaid_link.plaid_link_id;

ALTER TABLE "plaid_links" DROP COLUMN "access_token";