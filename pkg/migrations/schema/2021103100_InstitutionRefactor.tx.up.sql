ALTER TABLE "links"
    DROP COLUMN "institution_id";

DROP TABLE "institutions";

ALTER TABLE "links"
    ADD COLUMN "plaid_institution_id" TEXT NULL;

UPDATE "links"
SET "plaid_institution_id"="plaid_links"."institution_id"
FROM (
         SELECT "plaid_links"."institution_id", "plaid_links"."plaid_link_id" FROM "plaid_links"
     ) AS "plaid_links"
WHERE "links"."plaid_link_id" = "plaid_links"."plaid_link_id";
