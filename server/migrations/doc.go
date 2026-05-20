// Package migrations applies monetr's SQL schema migrations against
// Postgres. The migration files themselves live in ./schema/ and are pulled
// in at build time via //go:embed.
//
// To add a new migration:
//
//  1. Drop a file into server/migrations/schema/ named
//     YYYYMMDDNN_DescriptiveName.tx.up.sql, where YYYYMMDD is the date and
//     NN is a two-digit sequence within that day (e.g.
//     2026060100_AddFooColumn.tx.up.sql).
//  2. Use .tx.up.sql when the body can run inside a transaction. That's
//     what we want for basically everything. Reach for plain .up.sql only
//     when Postgres won't let you wrap the statement in a transaction:
//     CREATE INDEX CONCURRENTLY, some CREATE EXTENSION variants,
//     ALTER TYPE ... ADD VALUE, and so on.
//  3. Optionally add a matching .tx.down.sql or .down.sql with the inverse
//     statements. We validate down files for shape at startup but never
//     actually run them, they're just there as documentation.
//  4. Rebuild. //go:embed picks the new file up automatically, no code
//     changes required.
//
// The version (the integer prefix on the filename) must be unique across
// up files and must be greater than every previously applied migration.
// The runner applies any file whose version is greater than MAX(version) in
// schema_migrations, in ascending order.
package migrations
