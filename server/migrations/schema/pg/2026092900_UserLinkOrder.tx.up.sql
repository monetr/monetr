-- User defined ordering of links in the UI, stored as a list of link IDs.
ALTER TABLE "users" ADD COLUMN "link_order" VARCHAR(32)[];
