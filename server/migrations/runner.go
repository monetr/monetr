package migrations

import (
	"io/fs"
	"regexp"
	"sort"
	"strconv"

	"github.com/pkg/errors"
)

// migrationFilenameRegex matches the four shapes of migration filenames we
// support:
//
//	YYYYMMDDNN_Name.up.sql
//	YYYYMMDDNN_Name.tx.up.sql
//	YYYYMMDDNN_Name.down.sql
//	YYYYMMDDNN_Name.tx.down.sql
//
// The version prefix is a fixed ten digits: an eight-digit date plus a
// two-digit per-day sequence. Pinning the width keeps a fat-fingered prefix
// from quietly parsing as some wildly different version, which has tripped us
// up before. Names may contain underscores (2021041102_Balances_View is real)
// but not dots, since we use dots to mark the tx/direction suffix.
var migrationFilenameRegex = regexp.MustCompile(`^(\d{10})_([^.]+)\.(tx\.)?(up|down)\.sql$`)

type migrationFile struct {
	Version       int64
	Name          string
	Transactional bool
	Direction     string
	Filename      string
}

func parseMigrationFilename(name string) (migrationFile, error) {
	// A successful match has exactly five elements: the whole string plus the
	// four capture groups. Checking the length rather than just != nil keeps the
	// match[1..4] indexing below safe even if someone later adds or drops a group
	// in the pattern.
	match := migrationFilenameRegex.FindStringSubmatch(name)
	if len(match) != 5 {
		return migrationFile{}, errors.Errorf("unrecognized migration filename %q", name)
	}

	version, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return migrationFile{}, errors.Wrapf(err, "invalid version in filename %q", name)
	}

	return migrationFile{
		Version:       version,
		Name:          match[2],
		Transactional: match[3] == "tx.",
		Direction:     match[4],
		Filename:      name,
	}, nil
}

// discoverMigrations walks the root of fsys, parses every filename, and returns
// the up-files sorted ascending by version. Down files are validated for shape but
// stripped out, since the runtime never invokes them.
func discoverMigrations(fsys fs.FS) ([]migrationFile, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read embedded migrations directory")
	}

	ups := make([]migrationFile, 0, len(entries))
	seen := make(map[int64]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			return nil, errors.Errorf("unexpected directory in migrations: %q", entry.Name())
		}
		mf, err := parseMigrationFilename(entry.Name())
		if err != nil {
			return nil, err
		}
		if mf.Direction != "up" {
			continue
		}
		if existing, ok := seen[mf.Version]; ok {
			return nil, errors.Errorf(
				"duplicate migration version %d in both %q and %q",
				mf.Version, existing, mf.Filename,
			)
		}
		seen[mf.Version] = mf.Filename
		ups = append(ups, mf)
	}

	sort.Slice(ups, func(i, j int) bool {
		return ups[i].Version < ups[j].Version
	})

	return ups, nil
}

// mustSubMigrations roots fsys at dir, which is how each database engine gets
// its own set of migration files out of the shared embed. Like
// mustDiscoverMigrations, a failure here is a programming mistake compiled into
// the build.
func mustSubMigrations(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(errors.Wrapf(err, "failed to open embedded migrations directory %q", dir))
	}
	return sub
}

// mustDiscoverMigrations is discoverMigrations for the embedded set, where a
// parse failure is a programming mistake compiled into the build and there is
// nothing sensible to do but refuse to start.
func mustDiscoverMigrations(fsys fs.FS) []migrationFile {
	files, err := discoverMigrations(fsys)
	if err != nil {
		panic(errors.Wrap(err, "failed to parse embedded migrations"))
	}
	return files
}
