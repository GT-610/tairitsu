package services

import (
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/models"
	appservices "github.com/GT-610/tairitsu/internal/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserServiceRegisterReturnsSentinelForDuplicateUsername(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserServiceWithDB(db)

	_, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)

	_, err = service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.ErrorIs(t, err, appservices.ErrUsernameExists)
}

func TestUserServiceLoginReturnsSentinelForInvalidCredentials(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserServiceWithDB(db)

	_, err := service.Login(&models.LoginRequest{Username: "missing", Password: "secret123"})
	require.ErrorIs(t, err, appservices.ErrInvalidCredentials)
}

func TestUserServiceUpdateUserRoleValidatesRoleAndMissingUser(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserServiceWithDB(db)

	_, err := service.UpdateUserRole("missing", "super-admin")
	require.ErrorIs(t, err, appservices.ErrInvalidUserRole)

	_, err = service.UpdateUserRole("missing", "admin")
	require.ErrorIs(t, err, appservices.ErrUserNotFound)
}

func TestUserServiceChangePasswordReturnsSentinelForWrongPassword(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserServiceWithDB(db)

	user, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)

	err = service.ChangePassword(user.ID, "wrong-password", "next-secret")
	require.ErrorIs(t, err, appservices.ErrOldPasswordIncorrect)
}

func TestUserServiceUpdateUserRoleUpdatesStoredUser(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserServiceWithDB(db)

	user, err := service.Register(&models.RegisterRequest{Username: "user-1", Password: "secret123"}, "user")
	require.NoError(t, err)

	updated, err := service.UpdateUserRole(user.ID, "admin")
	require.NoError(t, err)
	assert.Equal(t, "admin", updated.Role)
}

func TestUserServiceTransferAdminTransfersSingleAdminRole(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserServiceWithDB(db)

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
	service := appservices.NewUserServiceWithDB(db)

	currentAdmin, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)

	_, err = service.TransferAdmin(currentAdmin.ID, currentAdmin.ID)
	require.ErrorIs(t, err, appservices.ErrAdminTransferSelf)
}

func TestUserServiceResetPasswordByAdminRejectsSelfReset(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserServiceWithDB(db)

	admin, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)

	_, _, _, err = service.ResetPasswordByAdmin(admin.ID, admin.ID)
	require.ErrorIs(t, err, appservices.ErrAdminResetSelf)
}

func TestUserServiceResetPasswordByAdminUpdatesPasswordAndRevokesSessions(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserServiceWithDB(db)

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

func TestUserServiceCreateUserByAdminCreatesUserWithTemporaryPassword(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserServiceWithDB(db)

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

func TestUserServiceDeleteUserByAdminTransfersNetworksAndRevokesSessions(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserServiceWithDB(db)

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

func TestUserServiceDeleteUserByAdminRejectsDeletingSelf(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserServiceWithDB(db)

	admin, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)

	_, _, _, err = service.DeleteUserByAdmin(admin.ID, admin.ID)
	require.ErrorIs(t, err, appservices.ErrAdminDeleteSelf)
}

func TestUserServiceDeleteUserByAdminRejectsDeletingAdmin(t *testing.T) {
	db := newTestSQLiteDB(t)
	service := appservices.NewUserServiceWithDB(db)

	admin, err := service.Register(&models.RegisterRequest{Username: "admin", Password: "secret123"}, "admin")
	require.NoError(t, err)
	otherAdmin, err := service.Register(&models.RegisterRequest{Username: "other-admin", Password: "secret123"}, "admin")
	require.NoError(t, err)

	_, _, _, err = service.DeleteUserByAdmin(admin.ID, otherAdmin.ID)
	require.ErrorIs(t, err, appservices.ErrAdminDeleteBlocked)
}
