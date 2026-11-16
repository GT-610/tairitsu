package database

import (
	"os"
	"path/filepath"
	"testing"

	appdb "github.com/GT-610/tairitsu/internal/app/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSQLiteFileIsOwnerOnly is the regression test for a world-readable database.
//
// The driver creates the file with the process umask, so a default deployment got
// 0644 next to a 0600 config.json — and that file holds every password hash and
// session hash. Measured before the fix: -rw-r--r--.
func TestSQLiteFileIsOwnerOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tairitsu.db")

	db, err := appdb.NewDatabase(appdb.Config{Type: appdb.SQLite, Path: path})
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(),
		"the database holds password and session hashes and must not be readable by other accounts")
}

// TestSQLiteFilePermissionsAreCorrectedOnRewrite covers a deployment that predates
// the fix: the file already exists with the wrong mode and has to be corrected
// rather than left as it was found.
func TestSQLiteFilePermissionsAreCorrectedOnRewrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tairitsu.db")
	require.NoError(t, os.WriteFile(path, nil, 0o644))

	before, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o644), before.Mode().Perm(), "precondition: the file starts world-readable")

	db, err := appdb.NewDatabase(appdb.Config{Type: appdb.SQLite, Path: path})
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	after, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), after.Mode().Perm())
}

// TestSQLiteWriteAheadLogInheritsOwnerOnly keeps the fix honest: the -wal file
// carries the same rows, and SQLite derives its mode from the database file. If
// that ever stops holding, this notices.
func TestSQLiteWriteAheadLogInheritsOwnerOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tairitsu.db")

	db, err := appdb.NewDatabase(appdb.Config{Type: appdb.SQLite, Path: path})
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	walInfo, err := os.Stat(path + "-wal")
	if err != nil {
		t.Skipf("no write-ahead log was created: %v", err)
	}

	assert.Equal(t, os.FileMode(0o600), walInfo.Mode().Perm(),
		"the write-ahead log holds the same rows and must match the database file")
}
