package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDataRoot(t *testing.T) string {
	t.Helper()

	dataRoot := t.TempDir()
	t.Setenv(config.DataDirEnv, dataRoot)
	return dataRoot
}

func TestResetDatabaseRemovesSQLiteFileAndItsWriteAheadLog(t *testing.T) {
	dataRoot := newDataRoot(t)
	dbPath := filepath.Join(dataRoot, "tairitsu.db")

	db, err := database.NewDatabase(database.Config{Type: database.SQLite, Path: dbPath})
	require.NoError(t, err)
	require.NoError(t, db.Init())
	require.NoError(t, db.Close())

	require.NoError(t, os.WriteFile(dbPath+"-wal", []byte("wal"), 0600))
	require.NoError(t, os.WriteFile(dbPath+"-shm", []byte("shm"), 0600))

	require.NoError(t, database.ResetDatabase(database.Config{Type: database.SQLite, Path: dbPath}))

	_, err = os.Stat(dbPath)
	assert.True(t, os.IsNotExist(err), "the database file must be gone")
	for _, suffix := range []string{"-wal", "-shm"} {
		_, err := os.Stat(dbPath + suffix)
		assert.True(t, os.IsNotExist(err), "the %s sibling must be gone", suffix)
	}
}

func TestResetDatabaseIsNotAnErrorWhenNothingExistsYet(t *testing.T) {
	dataRoot := newDataRoot(t)

	require.NoError(t, database.ResetDatabase(database.Config{
		Type: database.SQLite,
		Path: filepath.Join(dataRoot, "missing.db"),
	}))
}

// The setup wizard supplies this path, so a reset must never unlink a file that
// merely happens to sit where a database was expected.
func TestResetDatabaseRefusesFilesOutsideTheDataDirectory(t *testing.T) {
	newDataRoot(t)

	victim := filepath.Join(t.TempDir(), "important.txt")
	require.NoError(t, os.WriteFile(victim, []byte("do not delete"), 0600))

	err := database.ResetDatabase(database.Config{Type: database.SQLite, Path: victim})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside the Tairitsu data directory")

	content, readErr := os.ReadFile(victim)
	require.NoError(t, readErr)
	assert.Equal(t, "do not delete", string(content))
}

func TestResetDatabaseRefusesNonSQLiteFilesInsideTheDataDirectory(t *testing.T) {
	dataRoot := newDataRoot(t)

	victim := filepath.Join(dataRoot, "notes.txt")
	require.NoError(t, os.WriteFile(victim, []byte("do not delete"), 0600))

	err := database.ResetDatabase(database.Config{Type: database.SQLite, Path: victim})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a SQLite database")

	content, readErr := os.ReadFile(victim)
	require.NoError(t, readErr)
	assert.Equal(t, "do not delete", string(content))
}

func TestResetDatabaseRefusesDirectories(t *testing.T) {
	dataRoot := newDataRoot(t)

	directory := filepath.Join(dataRoot, "a-directory")
	require.NoError(t, os.MkdirAll(directory, 0755))

	err := database.ResetDatabase(database.Config{Type: database.SQLite, Path: directory})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a regular file")

	_, statErr := os.Stat(directory)
	require.NoError(t, statErr, "the directory must still exist")
}

// An interrupted setup leaves an empty file where the database would be; that
// should not block the next attempt.
func TestResetDatabaseRemovesEmptyLeftovers(t *testing.T) {
	dataRoot := newDataRoot(t)

	leftover := filepath.Join(dataRoot, "tairitsu.db")
	require.NoError(t, os.WriteFile(leftover, nil, 0600))

	require.NoError(t, database.ResetDatabase(database.Config{Type: database.SQLite, Path: leftover}))

	_, err := os.Stat(leftover)
	assert.True(t, os.IsNotExist(err))
}
