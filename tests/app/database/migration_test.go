package database

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	appdb "github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/models"
	// The same SQLite engine the application links, so the legacy table below is
	// created by the very driver that has to migrate it.
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeLegacyUsersTable creates the schema a database had before usernames were
// constrained: no unique index, so the same name could be inserted twice.
func writeLegacyUsersTable(t *testing.T, dbPath string, usernames ...string) {
	t.Helper()

	raw, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	defer raw.Close()

	_, err = raw.Exec(`CREATE TABLE users (
		id text PRIMARY KEY,
		username text,
		password text,
		role text,
		created_at datetime,
		updated_at datetime
	)`)
	require.NoError(t, err)

	now := time.Now()
	for index, username := range usernames {
		_, err = raw.Exec(
			"INSERT INTO users (id, username, password, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			fmt.Sprintf("%s-row-%d", username, index), username, "hashed", "user", now, now,
		)
		require.NoError(t, err)
	}
}

// AutoMigrate adds the unique index on username, so a database that predates it
// can no longer be opened unchanged when it holds collisions. Refusing to start
// is correct, but the operator has to be told which names collide: the
// driver-level failure names neither the accounts nor a way forward.
func TestInitNamesTheDuplicateUsernamesItRefusesToMigrate(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "tairitsu.db")
	writeLegacyUsersTable(t, dbPath, "admin", "admin", "bob", "carol", "carol")

	db, err := appdb.NewDatabase(appdb.Config{Type: appdb.SQLite, Path: dbPath})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	err = db.Init()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin", "the colliding names must be named")
	assert.Contains(t, err.Error(), "carol")
	assert.NotContains(t, err.Error(), "bob", "an account with a unique name is not a collision")
}

// The check must not be what stops an ordinary database from migrating.
func TestInitMigratesADatabaseWithUniqueUsernames(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "tairitsu.db")
	writeLegacyUsersTable(t, dbPath, "admin", "bob")

	db, err := appdb.NewDatabase(appdb.Config{Type: appdb.SQLite, Path: dbPath})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	require.NoError(t, db.Init())

	// The migrated schema must now refuse a second account with the same name.
	now := time.Now()
	require.NoError(t, db.CreateUser(&models.User{
		ID: "second", Username: "carol", Password: "hashed", Role: "user", CreatedAt: now, UpdatedAt: now,
	}))
	assert.Error(t, db.CreateUser(&models.User{
		ID: "third", Username: "carol", Password: "hashed", Role: "user", CreatedAt: now, UpdatedAt: now,
	}))
}

// A fresh installation has no users table, so the check has nothing to inspect.
func TestInitCreatesTheSchemaOnAnEmptyDatabase(t *testing.T) {
	db, err := appdb.NewDatabase(appdb.Config{
		Type: appdb.SQLite,
		Path: filepath.Join(t.TempDir(), "tairitsu.db"),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	require.NoError(t, db.Init())
}

// writeLegacyNetworkViewers creates the single-role sharing table that the RBAC
// migration replaces. The column names match the old GORM model exactly.
func writeLegacyNetworkViewers(t *testing.T, dbPath string, rows ...[3]string) {
	t.Helper()

	raw, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	defer raw.Close()

	_, err = raw.Exec(`CREATE TABLE network_viewers (
		network_id text,
		user_id text,
		granted_by text,
		created_at datetime,
		updated_at datetime,
		PRIMARY KEY (network_id, user_id)
	)`)
	require.NoError(t, err)

	now := time.Now()
	for _, row := range rows {
		_, err = raw.Exec(
			"INSERT INTO network_viewers (network_id, user_id, granted_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			row[0], row[1], row[2], now, now,
		)
		require.NoError(t, err)
	}
}

func hasLegacyViewerTable(t *testing.T, dbPath string) bool {
	t.Helper()

	raw, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	defer raw.Close()

	var count int
	err = raw.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'network_viewers'").Scan(&count)
	require.NoError(t, err)
	return count > 0
}

func TestInitMigratesLegacyNetworkViewers(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "tairitsu.db")
	writeLegacyNetworkViewers(t, dbPath,
		[3]string{"8056c2e21c000001", "user-1", "owner-1"},
		[3]string{"8056c2e21c000002", "user-1", "owner-1"},
	)

	db, err := appdb.NewDatabase(appdb.Config{Type: appdb.SQLite, Path: dbPath})
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	require.NoError(t, db.Init())

	grant, err := db.GetNetworkGrant("8056c2e21c000001", models.NetworkGrantSubjectUser, "user-1")
	require.NoError(t, err)
	require.NotNil(t, grant)
	assert.Equal(t, "viewer", grant.Role)
	assert.Equal(t, "owner-1", grant.GrantedBy)

	grants, err := db.GetNetworkGrants("8056c2e21c000001")
	require.NoError(t, err)
	assert.Len(t, grants, 1)

	assert.False(t, hasLegacyViewerTable(t, dbPath), "the legacy table must be removed after a successful copy")

	// A second start must be a no-op rather than a second copy.
	require.NoError(t, db.Init())
	grants, err = db.GetNetworkGrants("8056c2e21c000001")
	require.NoError(t, err)
	assert.Len(t, grants, 1)
}

func TestInitRemovesAnEmptyLegacyNetworkViewerTable(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "tairitsu.db")
	writeLegacyNetworkViewers(t, dbPath)

	db, err := appdb.NewDatabase(appdb.Config{Type: appdb.SQLite, Path: dbPath})
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	require.NoError(t, db.Init())
	assert.False(t, hasLegacyViewerTable(t, dbPath))
}

func TestInitLeavesAFreshDatabaseWithoutALegacyTable(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "tairitsu.db")
	db, err := appdb.NewDatabase(appdb.Config{Type: appdb.SQLite, Path: dbPath})
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	require.NoError(t, db.Init())
	assert.False(t, hasLegacyViewerTable(t, dbPath))
}
