package services

import (
	"crypto/rand"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const temporaryPasswordAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"

type UserService struct {
	db    database.DBInterface
	mutex sync.RWMutex
}

func (s *UserService) SetDB(db database.DBInterface) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.db = db
	logger.Info("user service database connection updated")
}

func (s *UserService) GetDB() database.DBInterface {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.db
}

func (s *UserService) getDB() database.DBInterface {
	s.mutex.RLock()
	db := s.db
	s.mutex.RUnlock()
	return db
}

func NewUserService(db database.DBInterface) *UserService {
	return &UserService{
		db: db,
	}
}

func normalizeUsername(username string) (string, error) {
	normalized := strings.TrimSpace(username)
	if normalized == "" {
		return "", ErrInvalidUsername
	}
	return normalized, nil
}

func (s *UserService) Register(req *models.RegisterRequest, role ...string) (*models.User, error) {
	db := s.getDB()
	if db == nil {
		logger.Error("service: registration failed; database is not initialized")
		return nil, ErrUserDBUnavailable
	}

	username, err := normalizeUsername(req.Username)
	if err != nil {
		return nil, err
	}

	logger.Info("service: starting user registration", zap.String("username", username))

	existingUser, err := db.GetUserByUsername(username)
	if err != nil {
		logger.Error("service: registration failed while checking username", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if existingUser != nil {
		logger.Error("service: registration failed; username already exists", zap.String("username", username))
		return nil, ErrUsernameExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("service: registration failed while hashing password", zap.String("username", req.Username), zap.Error(err))
		return nil, err
	}

	userRole := "user"
	if len(role) > 0 && role[0] != "" {
		userRole = role[0]
	}

	user := &models.User{
		ID:        uuid.New().String(),
		Username:  username,
		Password:  string(hashedPassword),
		Role:      userRole,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.CreateUser(user); err != nil {
		logger.Error("service: registration failed while saving user", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	logger.Info("service: user registered successfully", zap.String("user_id", user.ID), zap.String("username", user.Username), zap.String("role", userRole))

	return user, nil
}

func (s *UserService) CreateUserByAdmin(currentAdminID, username string) (*models.User, string, error) {
	currentAdmin, err := s.GetUserByID(currentAdminID)
	if err != nil {
		return nil, "", err
	}
	if currentAdmin.Role != "admin" {
		return nil, "", ErrAdminAccessDenied
	}

	temporaryPassword, err := generateTemporaryPassword(16)
	if err != nil {
		logger.Error("service: failed to generate temporary password", zap.String("admin_user_id", currentAdminID), zap.Error(err))
		return nil, "", err
	}

	normalizedUsername, err := normalizeUsername(username)
	if err != nil {
		return nil, "", err
	}

	user, err := s.Register(&models.RegisterRequest{
		Username: normalizedUsername,
		Password: temporaryPassword,
	}, "user")
	if err != nil {
		return nil, "", err
	}

	logger.Info("service: administrator created user successfully",
		zap.String("admin_user_id", currentAdminID),
		zap.String("user_id", user.ID),
		zap.String("username", user.Username))

	return user, temporaryPassword, nil
}

func (s *UserService) Login(req *models.LoginRequest) (*models.User, error) {
	db := s.getDB()
	if db == nil {
		logger.Error("service: login failed; database is not initialized")
		return nil, ErrUserDBUnavailable
	}

	username := strings.TrimSpace(req.Username)
	logger.Info("service: starting user login", zap.String("username", username))

	user, err := db.GetUserByUsername(username)
	if err != nil {
		logger.Error("service: login failed while querying user", zap.String("username", username), zap.Error(err))
		return nil, ErrInvalidCredentials
	}
	if user == nil {
		logger.Error("service: login failed; user does not exist", zap.String("username", username))
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		logger.Error("service: login failed; password mismatch", zap.String("user_id", user.ID))
		return nil, ErrInvalidCredentials
	}

	logger.Info("service: user logged in successfully", zap.String("user_id", user.ID), zap.String("username", user.Username))

	return user, nil
}

func (s *UserService) GetUserByID(id string) (*models.User, error) {
	db := s.getDB()
	if db == nil {
		logger.Error("service: get user failed; database is not initialized")
		return nil, ErrUserDBUnavailable
	}

	logger.Info("service: getting user by ID", zap.String("user_id", id))

	user, err := db.GetUserByID(id)
	if err != nil {
		logger.Error("service: failed to get user by ID", zap.String("user_id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to read user: %w", err)
	}

	if user == nil {
		logger.Error("service: user not found", zap.String("user_id", id))
		return nil, ErrUserNotFound
	}

	logger.Info("service: user retrieved by ID successfully", zap.String("user_id", id), zap.String("username", user.Username))

	return user, nil
}

func (s *UserService) GetAllUsers() []*models.User {
	db := s.getDB()
	if db == nil {
		logger.Warn("service: database is not initialized; returning empty user list")
		return []*models.User{}
	}

	logger.Info("service: getting all users")

	users, err := db.GetAllUsers()
	if err != nil {
		logger.Error("service: failed to get all users", zap.Error(err))
		return []*models.User{}
	}

	return users
}

func (s *UserService) HasAdminUser() (bool, error) {
	db := s.getDB()
	if db == nil {
		logger.Warn("service: database is not initialized; cannot check administrator users")
		return false, ErrUserDBUnavailable
	}

	logger.Info("service: checking for existing administrator user")

	hasAdmin, err := db.HasAdminUser()
	if err != nil {
		logger.Error("service: failed to check administrator users", zap.Error(err))
		return false, err
	}

	logger.Info("service: administrator user check completed", zap.Bool("hasAdmin", hasAdmin))
	return hasAdmin, nil
}

func (s *UserService) UpdateUser(user *models.User) error {
	db := s.getDB()
	if db == nil {
		logger.Error("service: update user failed; database is not initialized")
		return ErrUserDBUnavailable
	}

	logger.Info("service: updating user information", zap.String("user_id", user.ID))

	if err := db.UpdateUser(user); err != nil {
		logger.Error("service: update user failed while saving user", zap.String("user_id", user.ID), zap.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}

	logger.Info("service: user information updated successfully", zap.String("user_id", user.ID))
	return nil
}

func (s *UserService) ChangePassword(userID, oldPassword, newPassword string) error {
	_, err := s.changePassword(userID, oldPassword, newPassword, "", false)
	return err
}

func (s *UserService) ChangePasswordAndRevokeOtherSessions(userID, oldPassword, newPassword, currentSessionID string) (int, error) {
	return s.changePassword(userID, oldPassword, newPassword, currentSessionID, true)
}

func (s *UserService) changePassword(userID, oldPassword, newPassword, currentSessionID string, revokeOtherSessions bool) (int, error) {
	db := s.getDB()
	if db == nil {
		logger.Error("service: password change failed; database is not initialized")
		return 0, ErrUserDBUnavailable
	}

	logger.Info("service: changing password", zap.String("user_id", userID))

	user, err := db.GetUserByID(userID)
	if err != nil {
		logger.Error("service: password change failed while getting user", zap.String("user_id", userID), zap.Error(err))
		return 0, fmt.Errorf("failed to read user: %w", err)
	}

	if user == nil {
		logger.Error("service: password change failed; user not found", zap.String("user_id", userID))
		return 0, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		logger.Error("service: password change failed; current password is incorrect", zap.String("user_id", userID))
		return 0, ErrOldPasswordIncorrect
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("service: password change failed while hashing password", zap.String("user_id", userID), zap.Error(err))
		return 0, err
	}

	revokedSessions := 0
	now := time.Now()
	if err := db.WithTransaction(func(tx database.DBInterface) error {
		user.Password = string(hashedPassword)
		user.UpdatedAt = now

		if err := tx.UpdateUser(user); err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		if !revokeOtherSessions {
			return nil
		}

		sessions, err := tx.GetSessionsByUserID(userID)
		if err != nil {
			return fmt.Errorf("failed to read session list: %w", err)
		}

		for _, session := range sessions {
			if session.ID == currentSessionID || session.RevokedAt != nil {
				continue
			}
			session.RevokedAt = &now
			session.UpdatedAt = now
			if err := tx.UpdateSession(session); err != nil {
				return fmt.Errorf("failed to revoke other sessions: %w", err)
			}
			revokedSessions++
		}

		return nil
	}); err != nil {
		logger.Error("service: password change failed during transaction", zap.String("user_id", userID), zap.Error(err))
		return 0, err
	}

	logger.Info("service: password changed successfully", zap.String("user_id", userID))
	return revokedSessions, nil
}

func (s *UserService) UpdateUserRole(userID, role string) (*models.User, error) {
	if role != "admin" && role != "user" {
		return nil, ErrInvalidUserRole
	}

	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	user.Role = role
	if err := s.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) TransferAdmin(currentAdminID, targetUserID string) (*models.User, error) {
	db := s.getDB()
	if db == nil {
		logger.Error("service: administrator transfer failed; database is not initialized")
		return nil, ErrUserDBUnavailable
	}

	if currentAdminID == targetUserID {
		return nil, ErrAdminTransferSelf
	}

	currentAdmin, err := s.GetUserByID(currentAdminID)
	if err != nil {
		return nil, err
	}
	if currentAdmin.Role != "admin" {
		return nil, ErrAdminAccessDenied
	}

	targetUser, err := s.GetUserByID(targetUserID)
	if err != nil {
		return nil, err
	}
	if targetUser.Role == "admin" {
		return nil, ErrTransferTargetAdmin
	}

	now := time.Now()
	currentAdmin.Role = "user"
	currentAdmin.UpdatedAt = now
	targetUser.Role = "admin"
	targetUser.UpdatedAt = now

	if err := db.UpdateUser(currentAdmin); err != nil {
		logger.Error("service: administrator transfer failed while demoting previous administrator", zap.String("user_id", currentAdminID), zap.Error(err))
		return nil, fmt.Errorf("failed to update previous administrator: %w", err)
	}

	if err := db.UpdateUser(targetUser); err != nil {
		logger.Error("service: administrator transfer failed while promoting target user", zap.String("user_id", targetUserID), zap.Error(err))
		currentAdmin.Role = "admin"
		currentAdmin.UpdatedAt = time.Now()
		if rollbackErr := db.UpdateUser(currentAdmin); rollbackErr != nil {
			logger.Error("service: administrator transfer failed while rolling back previous administrator role", zap.String("user_id", currentAdminID), zap.Error(rollbackErr))
		}
		return nil, fmt.Errorf("failed to update target administrator: %w", err)
	}

	logger.Info("service: administrator role transferred successfully",
		zap.String("from_user_id", currentAdminID),
		zap.String("to_user_id", targetUserID),
		zap.String("to_username", targetUser.Username))

	return targetUser, nil
}

func generateTemporaryPassword(length int) (string, error) {
	password := make([]byte, length)
	max := byte(len(temporaryPasswordAlphabet))
	random := make([]byte, length)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("failed to generate random password: %w", err)
	}

	for index, value := range random {
		password[index] = temporaryPasswordAlphabet[int(value)%int(max)]
	}

	return string(password), nil
}

func (s *UserService) ResetPasswordByAdmin(currentAdminID, targetUserID string) (*models.User, string, int, error) {
	db := s.getDB()
	if db == nil {
		logger.Error("service: password reset failed; database is not initialized")
		return nil, "", 0, ErrUserDBUnavailable
	}

	if currentAdminID == targetUserID {
		return nil, "", 0, ErrAdminResetSelf
	}

	currentAdmin, err := s.GetUserByID(currentAdminID)
	if err != nil {
		return nil, "", 0, err
	}
	if currentAdmin.Role != "admin" {
		return nil, "", 0, ErrAdminAccessDenied
	}

	targetUser, err := s.GetUserByID(targetUserID)
	if err != nil {
		return nil, "", 0, err
	}

	temporaryPassword, err := generateTemporaryPassword(16)
	if err != nil {
		logger.Error("service: failed to generate temporary password", zap.String("target_user_id", targetUserID), zap.Error(err))
		return nil, "", 0, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(temporaryPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("service: password reset failed while hashing password", zap.String("target_user_id", targetUserID), zap.Error(err))
		return nil, "", 0, err
	}

	revokedSessions := 0
	now := time.Now()
	if err := db.WithTransaction(func(tx database.DBInterface) error {
		targetUser.Password = string(hashedPassword)
		targetUser.UpdatedAt = now
		if err := tx.UpdateUser(targetUser); err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		sessions, err := tx.GetSessionsByUserID(targetUserID)
		if err != nil {
			return fmt.Errorf("failed to read session list: %w", err)
		}

		for _, session := range sessions {
			if session.RevokedAt != nil {
				continue
			}
			session.RevokedAt = &now
			session.UpdatedAt = now
			if err := tx.UpdateSession(session); err != nil {
				return fmt.Errorf("failed to revoke user sessions: %w", err)
			}
			revokedSessions++
		}

		return nil
	}); err != nil {
		logger.Error("service: password reset failed during transaction", zap.String("target_user_id", targetUserID), zap.Error(err))
		return nil, "", 0, err
	}

	logger.Info("service: administrator reset user password successfully",
		zap.String("admin_user_id", currentAdminID),
		zap.String("target_user_id", targetUserID),
		zap.String("target_username", targetUser.Username),
		zap.Int("revoked_sessions", revokedSessions))

	return targetUser, temporaryPassword, revokedSessions, nil
}

func (s *UserService) DeleteUserByAdmin(currentAdminID, targetUserID string) (*models.User, int, int, error) {
	db := s.getDB()
	if db == nil {
		logger.Error("service: delete user failed; database is not initialized")
		return nil, 0, 0, ErrUserDBUnavailable
	}

	if currentAdminID == targetUserID {
		return nil, 0, 0, ErrAdminDeleteSelf
	}

	currentAdmin, err := s.GetUserByID(currentAdminID)
	if err != nil {
		return nil, 0, 0, err
	}
	if currentAdmin.Role != "admin" {
		return nil, 0, 0, ErrAdminAccessDenied
	}

	targetUser, err := s.GetUserByID(targetUserID)
	if err != nil {
		return nil, 0, 0, err
	}
	if targetUser.Role == "admin" {
		return nil, 0, 0, ErrAdminDeleteBlocked
	}

	now := time.Now()
	transferredNetworks := 0
	revokedSessions := 0
	if err := db.WithTransaction(func(tx database.DBInterface) error {
		networks, err := tx.GetNetworksByOwnerID(targetUserID)
		if err != nil {
			return fmt.Errorf("failed to read user networks: %w", err)
		}

		for _, network := range networks {
			network.OwnerID = currentAdminID
			network.UpdatedAt = now
			if err := tx.UpdateNetwork(network); err != nil {
				return fmt.Errorf("failed to transfer network ownership: %w", err)
			}
			transferredNetworks++
		}

		sessions, err := tx.GetSessionsByUserID(targetUserID)
		if err != nil {
			return fmt.Errorf("failed to read session list: %w", err)
		}

		for _, session := range sessions {
			if session.RevokedAt != nil {
				continue
			}
			session.RevokedAt = &now
			session.UpdatedAt = now
			if err := tx.UpdateSession(session); err != nil {
				return fmt.Errorf("failed to revoke user sessions: %w", err)
			}
			revokedSessions++
		}

		sharedNetworks, err := tx.GetSharedNetworksByUserID(targetUserID)
		if err != nil {
			return fmt.Errorf("failed to read shared network grants for user: %w", err)
		}

		for _, network := range sharedNetworks {
			if network == nil {
				continue
			}
			if err := tx.DeleteNetworkViewer(network.ID, targetUserID); err != nil {
				return fmt.Errorf("failed to delete shared network grants for user: %w", err)
			}
		}

		if err := tx.DeleteUser(targetUserID); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		return nil
	}); err != nil {
		logger.Error("service: delete user failed during transaction", zap.String("target_user_id", targetUserID), zap.Error(err))
		return nil, 0, 0, err
	}

	logger.Info("service: administrator deleted user successfully",
		zap.String("admin_user_id", currentAdminID),
		zap.String("target_user_id", targetUserID),
		zap.String("target_username", targetUser.Username),
		zap.Int("transferred_networks", transferredNetworks),
		zap.Int("revoked_sessions", revokedSessions))

	return targetUser, transferredNetworks, revokedSessions, nil
}
