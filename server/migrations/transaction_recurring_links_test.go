package migrations

import (
	"io/fs"
	"testing"

	"github.com/monetr/monetr/server/internal/testutils/testlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionRecurringLinksMigration(t *testing.T) {
	db := newCleanDatabase(t)
	log := testlog.GetLog(t)

	m := &PGMigrationsManager{log: log}
	pinned := pinnedExecutor(t, db)
	require.NoError(t, m.ensureSchemaTable(t.Context(), pinned), "must be able to create the schema table")

	const version = 2026100600
	before := make([]migrationFile, 0, len(pgMigrationFiles))
	var target migrationFile
	for _, f := range pgMigrationFiles {
		switch {
		case f.Version < version:
			before = append(before, f)
		case f.Version == version:
			target = f
		}
	}
	require.Equal(t, int64(version), target.Version, "must find the migration being tested")

	_, _, err := m.applyMigrations(t.Context(), pinned, pgMigrations, before)
	require.NoError(t, err, "must be able to migrate up to just before the links are moved")

	{ // Seed a debit and a credit, linked the way 2026100500 stored them
		require.NoError(t, pinned.Exec(t.Context(), `
INSERT INTO "accounts" ("account_id", "locale", "created_at")
VALUES ('acct_01', 'en_US', now());

INSERT INTO "logins" ("login_id", "email", "crypt", "is_enabled", "is_email_verified")
VALUES ('lgn_01', 'migration@monetr.local', '\x00', true, true);

INSERT INTO "users" ("user_id", "login_id", "account_id", "role")
VALUES ('user_01', 'lgn_01', 'acct_01', 'owner');

INSERT INTO "links" ("link_id", "account_id", "created_at", "created_by", "updated_at")
VALUES ('link_01', 'acct_01', now(), 'user_01', now());

INSERT INTO "bank_accounts" ("bank_account_id", "account_id", "link_id", "name", "available_balance", "current_balance", "created_at", "updated_at")
VALUES ('bac_01', 'acct_01', 'link_01', 'Checking', 0, 0, now(), now());

INSERT INTO "transaction_clusters" ("transaction_cluster_id", "account_id", "bank_account_id", "name", "original_name", "members", "created_at", "updated_at")
VALUES
  ('tcl_01', 'acct_01', 'bac_01', 'Github', 'Github', '{}', now(), now()),
  ('tcl_02', 'acct_01', 'bac_01', 'Spotify', 'Spotify', '{}', now(), now());

INSERT INTO "transaction_recurring" ("transaction_recurring_id", "account_id", "bank_account_id", "transaction_cluster_id", "direction", "window_type", "ruleset", "first", "last", "next", "confidence", "amounts", "last_amount")
VALUES
  ('txrc_debit', 'acct_01', 'bac_01', 'tcl_01', 'debit', 'monthly', 'RRULE:FREQ=MONTHLY', now(), now(), now(), 0.9, '{}', 800),
  ('txrc_credit', 'acct_01', 'bac_01', 'tcl_01', 'credit', 'monthly', 'RRULE:FREQ=MONTHLY', now(), now(), now(), 0.9, '{}', -500),
  ('txrc_unlinked', 'acct_01', 'bac_01', 'tcl_02', 'debit', 'weekly', 'RRULE:FREQ=WEEKLY', now(), now(), now(), 0.9, '{}', 100);

INSERT INTO "funding_schedules" ("funding_schedule_id", "account_id", "bank_account_id", "name", "ruleset", "next_recurrence", "next_recurrence_original", "transaction_recurring_id")
VALUES ('fund_01', 'acct_01', 'bac_01', 'Payday', 'RRULE:FREQ=MONTHLY', now(), now(), 'txrc_credit');

INSERT INTO "spending" ("spending_id", "account_id", "bank_account_id", "funding_schedule_id", "name", "target_amount", "current_amount", "used_amount", "next_recurrence", "next_contribution_amount", "is_behind", "is_paused", "created_at", "spending_type", "transaction_recurring_id")
VALUES ('spnd_01', 'acct_01', 'bac_01', 'fund_01', 'Github', 800, 0, 0, now(), 0, false, false, now(), 'expense', 'txrc_debit');
	`), "must be able to seed the old links")
	}

	_, newVersion, err := m.applyMigrations(t.Context(), pinned, pgMigrations, []migrationFile{target})
	require.NoError(t, err, "must be able to move the links")
	assert.EqualValues(t, version, newVersion, "should be at the links migration")

	type link struct {
		TransactionRecurringId string  `bun:"transaction_recurring_id"`
		SpendingId             *string `bun:"spending_id"`
		FundingScheduleId      *string `bun:"funding_schedule_id"`
	}
	var links []link
	require.NoError(t, pinned.Query(t.Context(), &links, `
SELECT "transaction_recurring_id", "spending_id", "funding_schedule_id"
FROM "transaction_recurring"
ORDER BY "transaction_recurring_id"`), "must be able to read the links")
	require.Len(t, links, 3, "must have all of the recurring transactions")
	byId := map[string]link{}
	for _, item := range links {
		byId[item.TransactionRecurringId] = item
	}

	if assert.NotNil(t, byId["txrc_debit"].SpendingId, "debit must carry the expense over") {
		assert.Equal(t, "spnd_01", *byId["txrc_debit"].SpendingId, "debit should point at the expense")
	}
	assert.Nil(t, byId["txrc_debit"].FundingScheduleId, "debit should not have a funding schedule")
	if assert.NotNil(t, byId["txrc_credit"].FundingScheduleId, "credit must carry the funding schedule over") {
		assert.Equal(t, "fund_01", *byId["txrc_credit"].FundingScheduleId, "credit should point at the funding schedule")
	}
	assert.Nil(t, byId["txrc_credit"].SpendingId, "credit should not have a spending")
	assert.Nil(t, byId["txrc_unlinked"].SpendingId, "unlinked should stay unlinked")
	assert.Nil(t, byId["txrc_unlinked"].FundingScheduleId, "unlinked should stay unlinked")

	var oldColumns int
	require.NoError(t, pinned.Get(t.Context(), &oldColumns, `
SELECT count(*)
FROM "information_schema"."columns"
WHERE "table_name" IN ('spending', 'funding_schedules') AND
      "column_name" = 'transaction_recurring_id'`), "must be able to count the old columns")
	assert.Zero(t, oldColumns, "the old link columns must be dropped")

	// The direction rules are enforced by the database now too
	assert.Error(t, pinned.Exec(
		t.Context(),
		`UPDATE "transaction_recurring" SET "spending_id" = 'spnd_01', "funding_schedule_id" = NULL WHERE "transaction_recurring_id" = 'txrc_credit'`,
	), "a credit must not be linked to a spending")
	assert.Error(t, pinned.Exec(
		t.Context(),
		`UPDATE "transaction_recurring" SET "funding_schedule_id" = 'fund_01' WHERE "transaction_recurring_id" = 'txrc_unlinked'`,
	), "a debit must not be linked to a funding schedule")

	// And down puts everything back where it was
	down, err := fs.ReadFile(pgMigrations, "2026100600_TransactionRecurringLinks.tx.down.sql")
	require.NoError(t, err, "must be able to read the down migration")
	require.NoError(t, pinned.Exec(t.Context(), string(down)), "must be able to roll the links back")

	var spendingRecurringId, fundingRecurringId *string
	require.NoError(t, pinned.Get(
		t.Context(),
		&spendingRecurringId,
		`SELECT "transaction_recurring_id" FROM "spending" WHERE "spending_id" = 'spnd_01'`,
	), "must be able to read the spending link")
	require.NoError(t, pinned.Get(
		t.Context(),
		&fundingRecurringId,
		`SELECT "transaction_recurring_id" FROM "funding_schedules" WHERE "funding_schedule_id" = 'fund_01'`,
	), "must be able to read the funding schedule link")
	if assert.NotNil(t, spendingRecurringId, "spending should be linked again") {
		assert.Equal(t, "txrc_debit", *spendingRecurringId, "spending should point at the debit")
	}
	if assert.NotNil(t, fundingRecurringId, "funding schedule should be linked again") {
		assert.Equal(t, "txrc_credit", *fundingRecurringId, "funding schedule should point at the credit")
	}
}
