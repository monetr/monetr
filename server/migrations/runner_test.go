package migrations

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMigrationFilename(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		cases := []struct {
			name string
			file string
			want migrationFile
		}{
			{
				name: "plain up",
				file: "2021041100_Extensions.up.sql",
				want: migrationFile{
					Version:       2021041100,
					Name:          "Extensions",
					Transactional: false,
					Direction:     "up",
					Filename:      "2021041100_Extensions.up.sql",
				},
			},
			{
				name: "tx up",
				file: "2024041100_NewID.tx.up.sql",
				want: migrationFile{
					Version:       2024041100,
					Name:          "NewID",
					Transactional: true,
					Direction:     "up",
					Filename:      "2024041100_NewID.tx.up.sql",
				},
			},
			{
				name: "plain down",
				file: "2021041700_PlaidTransactions.down.sql",
				want: migrationFile{
					Version:       2021041700,
					Name:          "PlaidTransactions",
					Transactional: false,
					Direction:     "down",
					Filename:      "2021041700_PlaidTransactions.down.sql",
				},
			},
			{
				name: "tx down",
				file: "2021042700_LoginNames.tx.down.sql",
				want: migrationFile{
					Version:       2021042700,
					Name:          "LoginNames",
					Transactional: true,
					Direction:     "down",
					Filename:      "2021042700_LoginNames.tx.down.sql",
				},
			},
			{
				name: "name with underscore",
				file: "2021041102_Balances_View.up.sql",
				want: migrationFile{
					Version:       2021041102,
					Name:          "Balances_View",
					Transactional: false,
					Direction:     "up",
					Filename:      "2021041102_Balances_View.up.sql",
				},
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := parseMigrationFilename(tc.file)
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		bad := []string{
			"notamigration.sql",
			"foo_bar.up.sql",      // no leading digits
			"2021041700.up.sql",   // missing descriptive name
			"2021041700_X.up.txt", // wrong extension
			"2021041700_X.sql",    // missing direction
			"2021041700_X.up",     // missing .sql
			"README.md",
		}
		for _, name := range bad {
			t.Run(name, func(t *testing.T) {
				_, err := parseMigrationFilename(name)
				assert.Error(t, err)
			})
		}
	})
}

func TestDiscoverMigrations_SortsAscending(t *testing.T) {
	fsys := fstest.MapFS{
		"schema/2024010100_Third.tx.up.sql":  &fstest.MapFile{Data: []byte("SELECT 1;")},
		"schema/2021041100_First.up.sql":     &fstest.MapFile{Data: []byte("SELECT 1;")},
		"schema/2023060100_Second.tx.up.sql": &fstest.MapFile{Data: []byte("SELECT 1;")},
		"schema/2021041100_First.down.sql":   &fstest.MapFile{Data: []byte("SELECT 1;")},
	}
	ups, err := discoverMigrations(fsys)
	require.NoError(t, err)
	require.Len(t, ups, 3, "down files must not appear in the up list")
	assert.Equal(t, int64(2021041100), ups[0].Version)
	assert.Equal(t, int64(2023060100), ups[1].Version)
	assert.Equal(t, int64(2024010100), ups[2].Version)
}

func TestDiscoverMigrations_DuplicateUpVersionFails(t *testing.T) {
	fsys := fstest.MapFS{
		"schema/2024010100_First.tx.up.sql":  &fstest.MapFile{Data: []byte("SELECT 1;")},
		"schema/2024010100_Second.tx.up.sql": &fstest.MapFile{Data: []byte("SELECT 1;")},
	}
	_, err := discoverMigrations(fsys)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestDiscoverMigrations_BadFilenameFails(t *testing.T) {
	fsys := fstest.MapFS{
		"schema/2024010100_OK.tx.up.sql": &fstest.MapFile{Data: []byte("SELECT 1;")},
		"schema/junk.txt":                &fstest.MapFile{Data: []byte("nope")},
	}
	_, err := discoverMigrations(fsys)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unrecognized")
}

func TestDiscoverMigrations_AllowsUpWithoutDown(t *testing.T) {
	fsys := fstest.MapFS{
		"schema/2024010100_NoReverse.tx.up.sql": &fstest.MapFile{Data: []byte("SELECT 1;")},
	}
	ups, err := discoverMigrations(fsys)
	require.NoError(t, err)
	require.Len(t, ups, 1)
}

func TestDiscoverMigrations_RealEmbed(t *testing.T) {
	ups, err := discoverMigrations(embeddedMigrations)
	require.NoError(t, err)

	// 99 .tx.up.sql + 6 .up.sql = 105 up files in the production embed.
	// If a new migration is added, bump this assertion.
	require.Len(t, ups, 105)
	assert.Equal(t, int64(2021041100), ups[0].Version)
	assert.Equal(t, int64(2026050600), ups[len(ups)-1].Version)

	// The list must be strictly ascending.
	for i := 1; i < len(ups); i++ {
		require.Greater(t, ups[i].Version, ups[i-1].Version)
	}
}
