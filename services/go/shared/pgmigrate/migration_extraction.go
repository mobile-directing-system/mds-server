package pgmigrate

import (
	"github.com/lefinal/meh"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

// MigrationsFromFS reads files in the given fs.FS and parses the name for
// migration meta-data. We expect filenames to be in the format of
// ###_human_readable_name_up.sql. The amount of version number digits is not
// fixed. When parsing, Migration.Name will be set to human_readable_name. The
// '_up.sql' suffix is mandatory.
func MigrationsFromFS(migrationsFS fs.FS) ([]Migration, error) {
	basePath := "."
	dirEntries, err := fs.ReadDir(migrationsFS, basePath)
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "read migrations fs directory", meh.Details{"base_path": basePath})
	}
	migrations := make([]Migration, 0, len(dirEntries))
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		migration, err := migrationFromFile(migrationsFS, filepath.Join(basePath, entry.Name()))
		if err != nil {
			return nil, meh.Wrap(err, "migration from file", meh.Details{"filename": entry.Name()})
		}
		migrations = append(migrations, migration)
	}
	return migrations, nil
}

const (
	fileNameSeparator      = "_"
	expectedFileNameSuffix = "up.sql"
)

// migrationFromFile parses and reads the migration file with the given name
// from the given fs.FS.
func migrationFromFile(migrationsFS fs.FS, fileName string) (Migration, error) {
	// Split filename into segments.
	fileNameSegments := strings.Split(fileName, fileNameSeparator)
	if len(fileNameSegments) < 3 {
		return Migration{}, meh.NewInternalErr("invalid migration filename format", meh.Details{
			"filename_segments": len(fileNameSegments),
			"separator":         fileNameSeparator,
		})
	}
	targetVersionStr := fileNameSegments[0]
	nameSuffix := fileNameSegments[len(fileNameSegments)-1]
	var migration Migration
	// Parse first segment as target version.
	targetVersion, err := strconv.Atoi(targetVersionStr)
	if err != nil {
		return Migration{}, meh.NewInternalErrFromErr(err, "parse target version",
			meh.Details{"segment": targetVersionStr})
	}
	migration.TargetVersion = targetVersion
	// Extract name.
	migration.Name = strings.TrimPrefix(targetVersionStr+fileNameSeparator, strings.TrimSuffix(fileName, fileNameSeparator+nameSuffix))
	// Check suffix.
	if nameSuffix != expectedFileNameSuffix {
		return Migration{}, meh.NewInternalErr("invalid filename suffix", meh.Details{
			"was":      nameSuffix,
			"expected": expectedFileNameSuffix,
		})
	}
	// Read migration content.
	f, err := migrationsFS.Open(fileName)
	if err != nil {
		return Migration{}, meh.NewInternalErrFromErr(err, "open migration file", meh.Details{"filename": fileName})
	}
	defer func() { _ = f.Close() }()
	sqlRaw, err := ioutil.ReadAll(f)
	if err != nil {
		return Migration{}, meh.NewInternalErrFromErr(err, "read migration file", nil)
	}
	migration.SQL = string(sqlRaw)
	return migration, nil
}
