// Package testutil holds the test doubles and fixtures that more than one test
// package needs, so those packages can share one definition instead of each
// carrying a copy that drifts.
package testutil

import (
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/stretchr/testify/require"
)

// NewSQLiteDB opens an isolated SQLite database that is removed when the test
// ends. Every caller gets its own file in its own temporary directory.
func NewSQLiteDB(t *testing.T) database.DBInterface {
	t.Helper()

	db, err := database.NewDatabase(database.Config{
		Type: database.SQLite,
		Path: t.TempDir() + "/tairitsu.db",
	})
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}

// FakeDB is an in-memory database.DBInterface for tests that need to decide what
// the database reports without standing up a real backend. Callers seed the
// accounts they need through NewFakeDB.
//
// Accounts and sessions are backed by storage because callers need to read back
// what they wrote; every other operation answers with a zero value. It is
// deliberately not a general mock, so it is safe to hand to a service that
// touches only part of the interface.
type FakeDB struct {
	users    []*models.User
	sessions []*models.Session

	// GetUserByIDError makes a user lookup fail with this error, for callers
	// that have to distinguish an unreachable database from a missing account.
	// Sessions are stored, so a test can resolve a credential and then fail only
	// the lookup that follows.
	GetUserByIDError error
}

// NewFakeDB seeds an in-memory database with the supplied accounts.
func NewFakeDB(users ...*models.User) *FakeDB {
	return &FakeDB{users: users}
}

func (f *FakeDB) Init() error { return nil }

func (f *FakeDB) WithTransaction(fn func(database.DBInterface) error) error {
	return fn(f)
}

func (f *FakeDB) CreateUser(user *models.User) error {
	f.users = append(f.users, user)
	return nil
}

func (f *FakeDB) GetUserByID(id string) (*models.User, error) {
	if f.GetUserByIDError != nil {
		return nil, f.GetUserByIDError
	}

	for _, user := range f.users {
		if user.ID == id {
			return user, nil
		}
	}

	return nil, nil
}

func (f *FakeDB) GetUserByUsername(username string) (*models.User, error) {
	for _, user := range f.users {
		if user.Username == username {
			return user, nil
		}
	}

	return nil, nil
}

func (f *FakeDB) GetAllUsers() ([]*models.User, error) { return f.users, nil }

func (f *FakeDB) GetUsersByIDs(ids []string) ([]*models.User, error) {
	result := make([]*models.User, 0, len(ids))
	for _, id := range ids {
		for _, user := range f.users {
			if user.ID == id {
				result = append(result, user)
				break
			}
		}
	}

	return result, nil
}

func (f *FakeDB) UpdateUser(user *models.User) error { return nil }

func (f *FakeDB) DeleteUser(id string) error {
	kept := f.users[:0]
	for _, user := range f.users {
		if user.ID != id {
			kept = append(kept, user)
		}
	}
	f.users = kept

	return nil
}

func (f *FakeDB) HasAdminUser() (bool, error) {
	for _, user := range f.users {
		if user.Role == "admin" {
			return true, nil
		}
	}

	return false, nil
}

// PromoteFirstAdministrator mirrors the real backend: it grants the role only
// while no account holds it and reports whether the grant happened.
func (f *FakeDB) PromoteFirstAdministrator(userID string) (bool, error) {
	for _, user := range f.users {
		if user.Role == "admin" {
			return false, nil
		}
	}

	for _, user := range f.users {
		if user.ID == userID {
			user.Role = "admin"
			return true, nil
		}
	}

	return false, nil
}

func (f *FakeDB) CreateSession(session *models.Session) error {
	f.sessions = append(f.sessions, session)
	return nil
}

func (f *FakeDB) GetSessionByID(id string) (*models.Session, error) {
	for _, session := range f.sessions {
		if session.ID == id {
			return session, nil
		}
	}

	return nil, nil
}

func (f *FakeDB) GetSessionByTokenHash(tokenHash string) (*models.Session, error) {
	for _, session := range f.sessions {
		if session.TokenHash != nil && *session.TokenHash == tokenHash {
			return session, nil
		}
	}

	return nil, nil
}

func (f *FakeDB) GetSessionsByUserID(userID string) ([]*models.Session, error) {
	result := []*models.Session{}
	for _, session := range f.sessions {
		if session.UserID == userID {
			result = append(result, session)
		}
	}

	return result, nil
}

func (f *FakeDB) UpdateSession(session *models.Session) error { return nil }

func (f *FakeDB) DeleteExpiredSessions(before time.Time) error { return nil }

func (f *FakeDB) DeleteSessionsWithoutCredential() (int64, error) { return 0, nil }

func (f *FakeDB) CreateNetwork(network *models.Network) error { return nil }

func (f *FakeDB) GetNetworkByID(id string) (*models.Network, error) { return nil, nil }

func (f *FakeDB) GetNetworksByOwnerID(ownerID string) ([]*models.Network, error) {
	return []*models.Network{}, nil
}

func (f *FakeDB) GetAllNetworks() ([]*models.Network, error) {
	return []*models.Network{}, nil
}

func (f *FakeDB) UpdateNetwork(network *models.Network) error { return nil }

func (f *FakeDB) DeleteNetwork(id string) error { return nil }

func (f *FakeDB) UpsertNetworkViewer(viewer *models.NetworkViewer) error { return nil }

func (f *FakeDB) GetNetworkViewer(networkID, userID string) (*models.NetworkViewer, error) {
	return nil, nil
}

func (f *FakeDB) GetNetworkViewers(networkID string) ([]*models.NetworkViewer, error) {
	return []*models.NetworkViewer{}, nil
}

func (f *FakeDB) GetSharedNetworksByUserID(userID string) ([]*models.Network, error) {
	return []*models.Network{}, nil
}

func (f *FakeDB) DeleteNetworkViewer(networkID, userID string) error { return nil }

func (f *FakeDB) DeleteAllNetworkViewers(networkID string) error { return nil }

func (f *FakeDB) Ping() error { return nil }

func (f *FakeDB) Close() error { return nil }
