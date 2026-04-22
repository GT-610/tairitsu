package database

import (
	"os"
	"path/filepath"
	"testing"

	appdb "github.com/GT-610/tairitsu/internal/app/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDatabase_SQLiteCreatesParentDirectory(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "data", "nested", "tairitsu.db")

	db, err := appdb.NewDatabase(appdb.Config{
		Type: appdb.SQLite,
		Path: dbPath,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	require.NoError(t, db.Init())

	_, err = os.Stat(filepath.Dir(dbPath))
	assert.NoError(t, err)
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)
}
