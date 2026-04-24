package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/models"
	appservices "github.com/GT-610/tairitsu/internal/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type txFailingDB struct {
	inner                 database.DBInterface
	failSessionUpdateAt   int
	failDeleteUser        bool
	sessionUpdateAttempts int
}

func (d *txFailingDB) Init() error { return d.inner.Init() }
func (d *txFailingDB) WithTransaction(fn func(database.DBInterface) error) error {
	return d.inner.WithTransaction(func(tx database.DBInterface) error {
		return fn(&txFailingDB{
			inner:               tx,
			failSessionUpdateAt: d.failSessionUpdateAt,
			failDeleteUser:      d.failDeleteUser,
		})
	})
}
func (d *txFailingDB) CreateUser(user *models.User) error { return d.inner.CreateUser(user) }
func (d *txFailingDB) GetUserByID(id string) (*models.User, error) {
	return d.inner.GetUserByID(id)
}
func (d *txFailingDB) GetUserByUsername(username string) (*models.User, error) {
	return d.inner.GetUserByUsername(username)
}
func (d *txFailingDB) GetAllUsers() ([]*models.User, error) { return d.inner.GetAllUsers() }
func (d *txFailingDB) UpdateUser(user *models.User) error   { return d.inner.UpdateUser(user) }
func (d *txFailingDB) DeleteUser(id string) error {
	if d.failDeleteUser {
		return fmt.Errorf("forced delete failure")
	}
	return d.inner.DeleteUser(id)
}
func (d *txFailingDB) CreateSession(session *models.Session) error {
	return d.inner.CreateSession(session)
}
func (d *txFailingDB) GetSessionByID(id string) (*models.Session, error) {
	return d.inner.GetSessionByID(id)
}
func (d *txFailingDB) GetSessionsByUserID(userID string) ([]*models.Session, error) {
	return d.inner.GetSessionsByUserID(userID)
}
func (d *txFailingDB) UpdateSession(session *models.Session) error {
	d.sessionUpdateAttempts++
	if d.failSessionUpdateAt > 0 && d.sessionUpdateAttempts == d.failSessionUpdateAt {
		return fmt.Errorf("forced session update failure")
	}
	return d.inner.UpdateSession(session)
}
func (d *txFailingDB) CreateNetwork(network *models.Network) error {
	return d.inner.CreateNetwork(network)
}
func (d *txFailingDB) GetNetworkByID(id string) (*models.Network, error) {
	return d.inner.GetNetworkByID(id)
}
func (d *txFailingDB) GetNetworksByOwnerID(ownerID string) ([]*models.Network, error) {
	return d.inner.GetNetworksByOwnerID(ownerID)
}
func (d *txFailingDB) GetAllNetworks() ([]*models.Network, error) { return d.inner.GetAllNetworks() }
func (d *txFailingDB) UpdateNetwork(network *models.Network) error {
	return d.inner.UpdateNetwork(network)
}
func (d *txFailingDB) DeleteNetwork(id string) error { return d.inner.DeleteNetwork(id) }
func (d *txFailingDB) UpsertNetworkViewer(viewer *models.NetworkViewer) error {
	return d.inner.UpsertNetworkViewer(viewer)
}
func (d *txFailingDB) GetNetworkViewer(networkID, userID string) (*models.NetworkViewer, error) {
	return d.inner.GetNetworkViewer(networkID, userID)
}
func (d *txFailingDB) GetNetworkViewers(networkID string) ([]*models.NetworkViewer, error) {
	return d.inner.GetNetworkViewers(networkID)
}
func (d *txFailingDB) GetSharedNetworksByUserID(userID string) ([]*models.Network, error) {
	return d.inner.GetSharedNetworksByUserID(userID)
}
func (d *txFailingDB) DeleteNetworkViewer(networkID, userID string) error {
	return d.inner.DeleteNetworkViewer(networkID, userID)
}
func (d *txFailingDB) HasAdminUser() (bool, error)   { return d.inner.HasAdminUser() }
func (d *txFailingDB) Close() error                  { return d.inner.Close() }

func TestUserServiceRegisterReturnsSentinelForDuplicateUsername(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(db)

	_, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)

	_, err = service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.ErrorIs(t, err, appservices.ErrUsernameExists)
}

func TestUserServiceRegisterRejectsEmptyOrWhitespaceUsername(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(db)

	_, err := service.Register(&models.RegisterRequest{Username: "", Password: "secret123"}, "user")
	require.ErrorIs(t, err, appservices.ErrInvalidUsername)

	_, err = service.Register(&models.RegisterRequest{Username: "   ", Password: "secret123"}, "user")
	require.ErrorIs(t, err, appservices.ErrInvalidUsername)
}

func TestUserServiceRegisterNormalizesUsernameBeforeCheckingDuplicates(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(db)

	user, err := service.Register(&models.RegisterRequest{Username: "  alice  ", Password: "secret123"}, "user")
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Username)

	_, err = service.Register(&models.RegisterRequest{Username: "alice", Password: "secret123"}, "user")
	require.ErrorIs(t, err, appservices.ErrUsernameExists)
}

func TestUserServiceLoginReturnsSentinelForInvalidCredentials(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(db)

	_, err := service.Login(&models.LoginRequest{Username: "missing", Password: "secret123"})
	require.ErrorIs(t, err, appservices.ErrInvalidCredentials)
}

func TestUserServiceUpdateUserRoleValidatesRoleAndMissingUser(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(db)

	_, err := service.UpdateUserRole("missing", "super-admin")
	require.ErrorIs(t, err, appservices.ErrInvalidUserRole)

	_, err = service.UpdateUserRole("missing", "admin")
	require.ErrorIs(t, err, appservices.ErrUserNotFound)
}

func TestUserServiceChangePasswordReturnsSentinelForWrongPassword(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(db)

	user, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)

	err = service.ChangePassword(user.ID, "wrong-password", "next-secret")
	require.ErrorIs(t, err, appservices.ErrOldPasswordIncorrect)
}

func TestUserServiceChangePasswordAndRevokeOtherSessionsRollsBackOnSessionFailure(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(&txFailingDB{
		inner:               db,
		failSessionUpdateAt: 1,
	})

	user, err := service.Register(&models.RegisterRequest{Username: "alice", Password: "secret123"}, "user")
	require.NoError(t, err)

	now := time.Now()
	currentSession := &models.Session{
		ID:         "session-current",
		UserID:     user.ID,
		UserAgent:  "browser-a",
		IPAddress:  "127.0.0.1",
		LastSeenAt: now,
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	otherSession := &models.Session{
		ID:         "session-other",
		UserID:     user.ID,
		UserAgent:  "browser-b",
		IPAddress:  "127.0.0.2",
		LastSeenAt: now,
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	require.NoError(t, db.CreateSession(currentSession))
	require.NoError(t, db.CreateSession(otherSession))

	_, err = service.ChangePasswordAndRevokeOtherSessions(user.ID, "secret123", "updated456", currentSession.ID)
	require.Error(t, err)

	_, err = service.Login(&models.LoginRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)

	_, err = service.Login(&models.LoginRequest{Username: "alice", Password: "updated456"})
	require.ErrorIs(t, err, appservices.ErrInvalidCredentials)

	reloadedOtherSession, err := db.GetSessionByID(otherSession.ID)
	require.NoError(t, err)
	assert.Nil(t, reloadedOtherSession.RevokedAt)
}

func TestUserServiceUpdateUserRoleUpdatesStoredUser(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(db)

	user, err := service.Register(&models.RegisterRequest{Username: "user-1", Password: "secret123"}, "user")
	require.NoError(t, err)

	updated, err := service.UpdateUserRole(user.ID, "admin")
	require.NoError(t, err)
	assert.Equal(t, "admin", updated.Role)
}

func TestUserServiceTransferAdminTransfersSingleAdminRole(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(db)

	currentAdmin, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)
	targetUser, err := service.Register(&models.RegisterRequest{Username: "alice", Password: "secret123"}, "user")
	require.NoError(t, err)

	updatedTarget, err := service.TransferAdmin(currentAdmin.ID, targetUser.ID)
	require.NoError(t, err)
	assert.Equal(t, "admin", updatedTarget.Role)

	reloadedCurrent, err := service.GetUserByID(currentAdmin.ID)
	require.NoError(t, err)
	assert.Equal(t, "user", reloadedCurrent.Role)

	reloadedTarget, err := service.GetUserByID(targetUser.ID)
	require.NoError(t, err)
	assert.Equal(t, "admin", reloadedTarget.Role)
}

func TestUserServiceTransferAdminRejectsTransferToSelf(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(db)

	currentAdmin, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)

	_, err = service.TransferAdmin(currentAdmin.ID, currentAdmin.ID)
	require.ErrorIs(t, err, appservices.ErrAdminTransferSelf)
}

func TestUserServiceResetPasswordByAdminRejectsSelfReset(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(db)

	admin, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)

	_, _, _, err = service.ResetPasswordByAdmin(admin.ID, admin.ID)
	require.ErrorIs(t, err, appservices.ErrAdminResetSelf)
}

func TestUserServiceResetPasswordByAdminUpdatesPasswordAndRevokesSessions(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(db)

	admin, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)
	target, err := service.Register(&models.RegisterRequest{Username: "alice", Password: "secret123"}, "user")
	require.NoError(t, err)

	now := time.Now()
	sessionA := &models.Session{
		ID:         "session-a",
		UserID:     target.ID,
		UserAgent:  "browser-a",
		IPAddress:  "127.0.0.1",
		LastSeenAt: now,
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	sessionB := &models.Session{
		ID:         "session-b",
		UserID:     target.ID,
		UserAgent:  "browser-b",
		IPAddress:  "127.0.0.2",
		LastSeenAt: now,
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	require.NoError(t, db.CreateSession(sessionA))
	require.NoError(t, db.CreateSession(sessionB))

	updatedUser, temporaryPassword, revokedSessions, err := service.ResetPasswordByAdmin(admin.ID, target.ID)
	require.NoError(t, err)
	assert.Equal(t, target.ID, updatedUser.ID)
	assert.Len(t, temporaryPassword, 16)
	assert.Equal(t, 2, revokedSessions)

	_, err = service.Login(&models.LoginRequest{Username: "alice", Password: "secret123"})
	require.ErrorIs(t, err, appservices.ErrInvalidCredentials)

	reloaded, err := service.Login(&models.LoginRequest{Username: "alice", Password: temporaryPassword})
	require.NoError(t, err)
	assert.Equal(t, target.ID, reloaded.ID)

	reloadedSessionA, err := db.GetSessionByID(sessionA.ID)
	require.NoError(t, err)
	assert.NotNil(t, reloadedSessionA.RevokedAt)

	reloadedSessionB, err := db.GetSessionByID(sessionB.ID)
	require.NoError(t, err)
	assert.NotNil(t, reloadedSessionB.RevokedAt)
}

func TestUserServiceResetPasswordByAdminRollsBackOnSessionRevokeFailure(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(&txFailingDB{
		inner:               db,
		failSessionUpdateAt: 1,
	})

	admin, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)
	target, err := service.Register(&models.RegisterRequest{Username: "alice", Password: "secret123"}, "user")
	require.NoError(t, err)

	now := time.Now()
	session := &models.Session{
		ID:         "session-a",
		UserID:     target.ID,
		UserAgent:  "browser-a",
		IPAddress:  "127.0.0.1",
		LastSeenAt: now,
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	require.NoError(t, db.CreateSession(session))

	_, _, _, err = service.ResetPasswordByAdmin(admin.ID, target.ID)
	require.Error(t, err)

	reloaded, err := service.Login(&models.LoginRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)
	assert.Equal(t, target.ID, reloaded.ID)

	reloadedSession, err := db.GetSessionByID(session.ID)
	require.NoError(t, err)
	assert.Nil(t, reloadedSession.RevokedAt)
}

func TestUserServiceCreateUserByAdminCreatesUserWithTemporaryPassword(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(db)

	admin, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)

	user, temporaryPassword, err := service.CreateUserByAdmin(admin.ID, "bob")
	require.NoError(t, err)
	assert.Equal(t, "bob", user.Username)
	assert.Equal(t, "user", user.Role)
	assert.Len(t, temporaryPassword, 16)

	loggedIn, err := service.Login(&models.LoginRequest{Username: "bob", Password: temporaryPassword})
	require.NoError(t, err)
	assert.Equal(t, user.ID, loggedIn.ID)
}

func TestUserServiceCreateUserByAdminRejectsWhitespaceUsername(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(db)

	admin, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)

	_, _, err = service.CreateUserByAdmin(admin.ID, "   ")
	require.ErrorIs(t, err, appservices.ErrInvalidUsername)
}

func TestUserServiceDeleteUserByAdminTransfersNetworksAndRevokesSessions(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(db)

	admin, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)
	target, err := service.Register(&models.RegisterRequest{Username: "alice", Password: "secret123"}, "user")
	require.NoError(t, err)

	now := time.Now()
	network := &models.Network{
		ID:          "net-1",
		Name:        "alice-network",
		Description: "test network",
		OwnerID:     target.ID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, db.CreateNetwork(network))

	session := &models.Session{
		ID:         "session-1",
		UserID:     target.ID,
		UserAgent:  "browser-a",
		IPAddress:  "127.0.0.1",
		LastSeenAt: now,
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	require.NoError(t, db.CreateSession(session))

	deletedUser, transferredNetworks, revokedSessions, err := service.DeleteUserByAdmin(admin.ID, target.ID)
	require.NoError(t, err)
	assert.Equal(t, target.ID, deletedUser.ID)
	assert.Equal(t, 1, transferredNetworks)
	assert.Equal(t, 1, revokedSessions)

	reloadedNetwork, err := db.GetNetworkByID(network.ID)
	require.NoError(t, err)
	assert.Equal(t, admin.ID, reloadedNetwork.OwnerID)

	reloadedSession, err := db.GetSessionByID(session.ID)
	require.NoError(t, err)
	assert.NotNil(t, reloadedSession.RevokedAt)

	reloadedUser, err := db.GetUserByID(target.ID)
	require.NoError(t, err)
	assert.Nil(t, reloadedUser)

	_, err = service.Login(&models.LoginRequest{Username: "alice", Password: "secret123"})
	require.ErrorIs(t, err, appservices.ErrInvalidCredentials)
}

func TestUserServiceDeleteUserByAdminRollsBackOnDeleteFailure(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(&txFailingDB{
		inner:          db,
		failDeleteUser: true,
	})

	admin, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)
	target, err := service.Register(&models.RegisterRequest{Username: "alice", Password: "secret123"}, "user")
	require.NoError(t, err)

	now := time.Now()
	network := &models.Network{
		ID:          "net-1",
		Name:        "alice-network",
		Description: "test network",
		OwnerID:     target.ID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, db.CreateNetwork(network))

	session := &models.Session{
		ID:         "session-1",
		UserID:     target.ID,
		UserAgent:  "browser-a",
		IPAddress:  "127.0.0.1",
		LastSeenAt: now,
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	require.NoError(t, db.CreateSession(session))

	_, _, _, err = service.DeleteUserByAdmin(admin.ID, target.ID)
	require.Error(t, err)

	reloadedUser, err := db.GetUserByID(target.ID)
	require.NoError(t, err)
	assert.NotNil(t, reloadedUser)

	reloadedNetwork, err := db.GetNetworkByID(network.ID)
	require.NoError(t, err)
	assert.Equal(t, target.ID, reloadedNetwork.OwnerID)

	reloadedSession, err := db.GetSessionByID(session.ID)
	require.NoError(t, err)
	assert.Nil(t, reloadedSession.RevokedAt)
}

func TestUserServiceDeleteUserByAdminRejectsDeletingSelf(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(db)

	admin, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)

	_, _, _, err = service.DeleteUserByAdmin(admin.ID, admin.ID)
	require.ErrorIs(t, err, appservices.ErrAdminDeleteSelf)
}

func TestUserServiceDeleteUserByAdminRejectsDeletingAdmin(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserService(db)

	admin, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)
	otherAdmin, err := service.Register(&models.RegisterRequest{Username: "other-admin", Password: "secret123"}, "admin")
	require.NoError(t, err)

	_, _, _, err = service.DeleteUserByAdmin(admin.ID, otherAdmin.ID)
	require.ErrorIs(t, err, appservices.ErrAdminDeleteBlocked)
}
