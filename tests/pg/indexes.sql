BEGIN;
SELECT plan(2);

-- Make sure that the index we specify does get renamed. I don't want to change the migration script as it's old. So I
-- want to make sure that it is properly applied. If this fails then that means the PostgreSQL behavior has changed in
-- whatever version that is being used. And that this index (when applied from scratch) is no longer being renamed.
SELECT hasnt_index('plaid_links', 'ix_uq_plaid_links_item_id');
SELECT has_index('plaid_links', 'uq_plaid_links_item_id');

SELECT *
FROM finish();
ROLLBACK;
