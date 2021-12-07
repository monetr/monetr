CREATE UNIQUE INDEX ix_uq_plaid_links_item_id ON plaid_links (item_id);
ALTER TABLE plaid_links ADD CONSTRAINT uq_plaid_links_item_id UNIQUE USING INDEX ix_uq_plaid_links_item_id;