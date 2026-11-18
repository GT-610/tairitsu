package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	appdb "github.com/GT-610/tempest-zt/internal/app/database"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSQLiteFileIsOwnerOnly is the regression test for a world-readable database.
//
// The driver creates the file with the process umask, so a default deployment got
// 0644 next to a 0600 config.json — and that file holds every password hash and
// session hash. Measured before the fix: -rw-r--r--.
func TestSQLiteFileIsOwnerOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tempest-zt.db")

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
	path := filepath.Join(t.TempDir(), "tempest-zt.db")
	require.NoError(t, os.WriteFile(path, nil, 0o644))
	// The umask applies at creation, so the mode is set explicitly rather than
	// assumed: under umask 0077 the file above is 0600 and there is nothing to fix.
	require.NoError(t, os.Chmod(path, 0o644))

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

// TestSQLiteWriteAheadLogInheritsOwnerOnly covers a database that SQLite created
// in this run: it derives the -wal mode from the database file, so no separate
// correction is needed.
func TestSQLiteWriteAheadLogInheritsOwnerOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tempest-zt.db")

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

// TestSQLiteRestrictsAPreExistingWriteAheadLog is the case a fresh database does
// not reach. SQLite copies the database file's permissions only when it *creates*
// a write-ahead log; one left behind by an unclean shutdown keeps whatever mode it
// already had, and the next start reuses it with the same rows inside.
//
// The database is snapshotted while a connection still holds the log open, which is
// how a leftover log looks on disk, rather than after a clean close — a clean close
// checkpoints and removes it, and the test would then assert nothing.
func TestSQLiteRestrictsAPreExistingWriteAheadLog(t *testing.T) {
	live := filepath.Join(t.TempDir(), "tempest-zt.db")

	handle, err := sql.Open("sqlite3", live+"?_journal_mode=WAL&_busy_timeout=5000")
	require.NoError(t, err)
	t.Cleanup(func() { _ = handle.Close() })

	_, err = handle.Exec(`CREATE TABLE probe (id INTEGER PRIMARY KEY, payload TEXT)`)
	require.NoError(t, err)
	_, err = handle.Exec(`INSERT INTO probe (payload) VALUES ('left over from a write')`)
	require.NoError(t, err)

	// Copy the trio while the log is still on disk. Nothing is writing, so the
	// snapshot is consistent, and it is what a killed process leaves behind.
	target := filepath.Join(t.TempDir(), "tempest-zt.db")
	for _, suffix := range []string{"", "-wal", "-shm"} {
		content, readErr := os.ReadFile(live + suffix)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				continue
			}
			require.NoError(t, readErr)
		}
		require.NoError(t, os.WriteFile(target+suffix, content, 0o644))
		// The umask applies at creation, so set the permissive mode explicitly.
		require.NoError(t, os.Chmod(target+suffix, 0o644))
	}

	_, statErr := os.Stat(target + "-wal")
	require.NoError(t, statErr, "precondition: the snapshot must include a write-ahead log")

	// Correcting the database file alone must not be enough.
	db, err := appdb.NewDatabase(appdb.Config{Type: appdb.SQLite, Path: target})
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	for _, suffix := range []string{"", "-wal", "-shm"} {
		info, statErr := os.Stat(target + suffix)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				continue
			}
			require.NoError(t, statErr)
		}
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(),
			"tempest-zt.db%s holds the same rows as the database and must be owner-only", suffix)
	}
}

// TestSQLiteRestrictsSidecarsOfASymlinkedDatabase covers the resolution SQLite does
// itself: it follows a symlinked database filename and derives the -wal and -shm names
// from the target, so correcting the link's own name reaches neither sidecar. The
// database file happens to be corrected either way, because os.Chmod follows the link —
// the sidecars are where the difference shows, so the test keeps one on disk.
func TestSQLiteRestrictsSidecarsOfASymlinkedDatabase(t *testing.T) {
	targetDir := t.TempDir()
	linkDir := t.TempDir()

	target := filepath.Join(targetDir, "tempest-zt.db")
	link := filepath.Join(linkDir, "tempest-zt.db")

	// A connection left open, so the write-ahead log is still on disk to be corrected.
	handle, err := sql.Open("sqlite3", target+"?_journal_mode=WAL&_busy_timeout=5000")
	require.NoError(t, err)
	t.Cleanup(func() { _ = handle.Close() })
	_, err = handle.Exec(`CREATE TABLE probe (id INTEGER PRIMARY KEY, payload TEXT)`)
	require.NoError(t, err)
	_, err = handle.Exec(`INSERT INTO probe (payload) VALUES ('row in the log')`)
	require.NoError(t, err)

	for _, name := range []string{target, target + "-wal", target + "-shm"} {
		if err := os.Chmod(name, 0o644); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			require.NoError(t, err)
		}
	}
	_, err = os.Stat(target + "-wal")
	require.NoError(t, err, "precondition: the snapshot must include a write-ahead log")

	require.NoError(t, os.Symlink(target, link))

	// Opened through the link, corrected on the target.
	db, err := appdb.NewDatabase(appdb.Config{Type: appdb.SQLite, Path: link})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	for _, suffix := range []string{"", "-wal", "-shm"} {
		info, statErr := os.Stat(target + suffix)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				continue
			}
			require.NoError(t, statErr)
		}
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(),
			"tempest-zt.db%s is what SQLite actually uses and must be owner-only", suffix)
	}
}
