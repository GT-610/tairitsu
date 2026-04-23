package services

import (
	"testing"

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
