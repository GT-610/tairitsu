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
	logger.Info("用户服务数据库连接已更新")
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
		logger.Error("服务层：注册失败，数据库未初始化")
		return nil, ErrUserDBUnavailable
	}

	username, err := normalizeUsername(req.Username)
	if err != nil {
		return nil, err
	}

	logger.Info("服务层：开始用户注册", zap.String("username", username))

	existingUser, err := db.GetUserByUsername(username)
	if err != nil {
		logger.Error("服务层：注册失败，检查用户名时出错", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if existingUser != nil {
		logger.Error("服务层：注册失败，用户名已存在", zap.String("username", username))
		return nil, ErrUsernameExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("服务层：注册失败，密码加密错误", zap.String("username", req.Username), zap.Error(err))
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
		logger.Error("服务层：注册失败，保存用户时出错", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("保存用户失败: %w", err)
	}

	logger.Info("服务层：用户注册成功", zap.String("user_id", user.ID), zap.String("username", user.Username), zap.String("role", userRole))

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
		logger.Error("服务层：生成临时密码失败", zap.String("admin_user_id", currentAdminID), zap.Error(err))
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

	logger.Info("服务层：管理员创建用户成功",
		zap.String("admin_user_id", currentAdminID),
		zap.String("user_id", user.ID),
		zap.String("username", user.Username))

	return user, temporaryPassword, nil
}

func (s *UserService) Login(req *models.LoginRequest) (*models.User, error) {
	db := s.getDB()
	if db == nil {
		logger.Error("服务层：登录失败，数据库未初始化")
		return nil, ErrUserDBUnavailable
	}

	username := strings.TrimSpace(req.Username)
	logger.Info("服务层：开始用户登录", zap.String("username", username))

	user, err := db.GetUserByUsername(username)
	if err != nil {
		logger.Error("服务层：登录失败，查询用户时出错", zap.String("username", username), zap.Error(err))
		return nil, ErrInvalidCredentials
	}
	if user == nil {
		logger.Error("服务层：登录失败，用户不存在", zap.String("username", username))
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		logger.Error("服务层：登录失败，密码错误", zap.String("user_id", user.ID))
		return nil, ErrInvalidCredentials
	}

	logger.Info("服务层：用户登录成功", zap.String("user_id", user.ID), zap.String("username", user.Username))

	return user, nil
}

func (s *UserService) GetUserByID(id string) (*models.User, error) {
	db := s.getDB()
	if db == nil {
		logger.Error("服务层：获取用户失败，数据库未初始化")
		return nil, ErrUserDBUnavailable
	}

	logger.Info("服务层：开始根据ID获取用户", zap.String("user_id", id))

	user, err := db.GetUserByID(id)
	if err != nil {
		logger.Error("服务层：根据ID获取用户失败", zap.String("user_id", id), zap.Error(err))
		return nil, fmt.Errorf("读取用户失败: %w", err)
	}

	if user == nil {
		logger.Error("服务层：用户不存在", zap.String("user_id", id))
		return nil, ErrUserNotFound
	}

	logger.Info("服务层：成功根据ID获取用户", zap.String("user_id", id), zap.String("username", user.Username))

	return user, nil
}

func (s *UserService) GetAllUsers() []*models.User {
	db := s.getDB()
	if db == nil {
		logger.Warn("服务层：数据库未初始化，返回空用户列表")
		return []*models.User{}
	}

	logger.Info("服务层：获取所有用户")

	users, err := db.GetAllUsers()
	if err != nil {
		logger.Error("服务层：获取所有用户失败", zap.Error(err))
		return []*models.User{}
	}

	return users
}

func (s *UserService) HasAdminUser() (bool, error) {
	db := s.getDB()
	if db == nil {
		logger.Warn("服务层：数据库未初始化，无法检查管理员用户")
		return false, ErrUserDBUnavailable
	}

	logger.Info("服务层：检查是否已存在管理员用户")

	hasAdmin, err := db.HasAdminUser()
	if err != nil {
		logger.Error("服务层：检查管理员用户失败", zap.Error(err))
		return false, err
	}

	logger.Info("服务层：检查管理员用户完成", zap.Bool("hasAdmin", hasAdmin))
	return hasAdmin, nil
}

func (s *UserService) UpdateUser(user *models.User) error {
	db := s.getDB()
	if db == nil {
		logger.Error("服务层：更新用户失败，数据库未初始化")
		return ErrUserDBUnavailable
	}

	logger.Info("服务层：开始更新用户信息", zap.String("user_id", user.ID))

	if err := db.UpdateUser(user); err != nil {
		logger.Error("服务层：更新用户失败，保存用户时出错", zap.String("user_id", user.ID), zap.Error(err))
		return fmt.Errorf("更新用户失败: %w", err)
	}

	logger.Info("服务层：用户信息更新成功", zap.String("user_id", user.ID))
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
		logger.Error("服务层：修改密码失败，数据库未初始化")
		return 0, ErrUserDBUnavailable
	}

	logger.Info("服务层：开始修改密码", zap.String("user_id", userID))

	user, err := db.GetUserByID(userID)
	if err != nil {
		logger.Error("服务层：修改密码失败，获取用户时出错", zap.String("user_id", userID), zap.Error(err))
		return 0, fmt.Errorf("读取用户失败: %w", err)
	}

	if user == nil {
		logger.Error("服务层：修改密码失败，用户不存在", zap.String("user_id", userID))
		return 0, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		logger.Error("服务层：修改密码失败，原密码错误", zap.String("user_id", userID))
		return 0, ErrOldPasswordIncorrect
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("服务层：修改密码失败，密码加密错误", zap.String("user_id", userID), zap.Error(err))
		return 0, err
	}

	revokedSessions := 0
	now := time.Now()
	if err := db.WithTransaction(func(tx database.DBInterface) error {
		user.Password = string(hashedPassword)
		user.UpdatedAt = now

		if err := tx.UpdateUser(user); err != nil {
			return fmt.Errorf("更新密码失败: %w", err)
		}

		if !revokeOtherSessions {
			return nil
		}

		sessions, err := tx.GetSessionsByUserID(userID)
		if err != nil {
			return fmt.Errorf("读取会话列表失败: %w", err)
		}

		for _, session := range sessions {
			if session.ID == currentSessionID || session.RevokedAt != nil {
				continue
			}
			session.RevokedAt = &now
			session.UpdatedAt = now
			if err := tx.UpdateSession(session); err != nil {
				return fmt.Errorf("吊销其他会话失败: %w", err)
			}
			revokedSessions++
		}

		return nil
	}); err != nil {
		logger.Error("服务层：修改密码失败，事务执行出错", zap.String("user_id", userID), zap.Error(err))
		return 0, err
	}

	logger.Info("服务层：密码修改成功", zap.String("user_id", userID))
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
		logger.Error("服务层：转让管理员失败，数据库未初始化")
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
		logger.Error("服务层：转让管理员失败，降级旧管理员时出错", zap.String("user_id", currentAdminID), zap.Error(err))
		return nil, fmt.Errorf("更新原管理员失败: %w", err)
	}

	if err := db.UpdateUser(targetUser); err != nil {
		logger.Error("服务层：转让管理员失败，提升新管理员时出错", zap.String("user_id", targetUserID), zap.Error(err))
		currentAdmin.Role = "admin"
		currentAdmin.UpdatedAt = time.Now()
		if rollbackErr := db.UpdateUser(currentAdmin); rollbackErr != nil {
			logger.Error("服务层：转让管理员失败，回滚原管理员角色时出错", zap.String("user_id", currentAdminID), zap.Error(rollbackErr))
		}
		return nil, fmt.Errorf("更新目标管理员失败: %w", err)
	}

	logger.Info("服务层：管理员身份转让成功",
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
		return "", fmt.Errorf("生成随机密码失败: %w", err)
	}

	for index, value := range random {
		password[index] = temporaryPasswordAlphabet[int(value)%int(max)]
	}

	return string(password), nil
}

func (s *UserService) ResetPasswordByAdmin(currentAdminID, targetUserID string) (*models.User, string, int, error) {
	db := s.getDB()
	if db == nil {
		logger.Error("服务层：重置密码失败，数据库未初始化")
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
		logger.Error("服务层：生成临时密码失败", zap.String("target_user_id", targetUserID), zap.Error(err))
		return nil, "", 0, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(temporaryPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("服务层：重置密码失败，密码加密错误", zap.String("target_user_id", targetUserID), zap.Error(err))
		return nil, "", 0, err
	}

	revokedSessions := 0
	now := time.Now()
	if err := db.WithTransaction(func(tx database.DBInterface) error {
		targetUser.Password = string(hashedPassword)
		targetUser.UpdatedAt = now
		if err := tx.UpdateUser(targetUser); err != nil {
			return fmt.Errorf("更新密码失败: %w", err)
		}

		sessions, err := tx.GetSessionsByUserID(targetUserID)
		if err != nil {
			return fmt.Errorf("读取会话列表失败: %w", err)
		}

		for _, session := range sessions {
			if session.RevokedAt != nil {
				continue
			}
			session.RevokedAt = &now
			session.UpdatedAt = now
			if err := tx.UpdateSession(session); err != nil {
				return fmt.Errorf("吊销用户会话失败: %w", err)
			}
			revokedSessions++
		}

		return nil
	}); err != nil {
		logger.Error("服务层：重置密码失败，事务执行出错", zap.String("target_user_id", targetUserID), zap.Error(err))
		return nil, "", 0, err
	}

	logger.Info("服务层：管理员重置用户密码成功",
		zap.String("admin_user_id", currentAdminID),
		zap.String("target_user_id", targetUserID),
		zap.String("target_username", targetUser.Username),
		zap.Int("revoked_sessions", revokedSessions))

	return targetUser, temporaryPassword, revokedSessions, nil
}

func (s *UserService) DeleteUserByAdmin(currentAdminID, targetUserID string) (*models.User, int, int, error) {
	db := s.getDB()
	if db == nil {
		logger.Error("服务层：删除用户失败，数据库未初始化")
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
			return fmt.Errorf("读取用户网络失败: %w", err)
		}

		for _, network := range networks {
			network.OwnerID = currentAdminID
			network.UpdatedAt = now
			if err := tx.UpdateNetwork(network); err != nil {
				return fmt.Errorf("转移网络所有权失败: %w", err)
			}
			transferredNetworks++
		}

		sessions, err := tx.GetSessionsByUserID(targetUserID)
		if err != nil {
			return fmt.Errorf("读取会话列表失败: %w", err)
		}

		for _, session := range sessions {
			if session.RevokedAt != nil {
				continue
			}
			session.RevokedAt = &now
			session.UpdatedAt = now
			if err := tx.UpdateSession(session); err != nil {
				return fmt.Errorf("吊销用户会话失败: %w", err)
			}
			revokedSessions++
		}

		if err := tx.DeleteUser(targetUserID); err != nil {
			return fmt.Errorf("删除用户失败: %w", err)
		}

		return nil
	}); err != nil {
		logger.Error("服务层：删除用户失败，事务执行出错", zap.String("target_user_id", targetUserID), zap.Error(err))
		return nil, 0, 0, err
	}

	logger.Info("服务层：管理员删除用户成功",
		zap.String("admin_user_id", currentAdminID),
		zap.String("target_user_id", targetUserID),
		zap.String("target_username", targetUser.Username),
		zap.Int("transferred_networks", transferredNetworks),
		zap.Int("revoked_sessions", revokedSessions))

	return targetUser, transferredNetworks, revokedSessions, nil
}
