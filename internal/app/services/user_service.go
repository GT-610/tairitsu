package services

import (
	"errors"
	"sync"
	"time"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// UserService handles user-related operations and business logic
type UserService struct {
	db    database.DBInterface // Database interface for user data operations
	mutex sync.RWMutex         // Read-write mutex to protect concurrent access
}

// NewUserServiceWithDB creates a new UserService instance with database connection
func NewUserServiceWithDB(db database.DBInterface) *UserService {
	return &UserService{
		db: db,
	}
}

// NewUserServiceWithoutDB creates a new UserService instance without database connection
// This is typically used for testing or in-memory operations
func NewUserServiceWithoutDB() *UserService {
	return &UserService{
		db: nil,
	}
}

// NewUserService creates a new UserService instance using the default SQLite database
func NewUserService() *UserService {
	// Create default SQLite database implementation
	db, err := database.NewDatabase(database.Config{
		Type: database.SQLite,
	})
	if err != nil {
		panic("无法创建默认数据库: " + err.Error())
	}

	return &UserService{
		db: db,
	}
}

// Register handles user registration with optional role specification
func (s *UserService) Register(req *models.RegisterRequest, role ...string) (*models.User, error) {
	// Check if database is initialized
	if s.db == nil {
		logger.Error("服务层：注册失败，数据库未初始化")
		return nil, errors.New("系统尚未配置数据库，请先完成初始设置")
	}

	logger.Info("服务层：开始用户注册", zap.String("username", req.Username))

	// Check if username already exists
	existingUser, err := s.db.GetUserByUsername(req.Username)
	if err != nil {
		logger.Error("服务层：注册失败，检查用户名时出错", zap.String("username", req.Username), zap.Error(err))
		return nil, err
	}
	if existingUser != nil {
		logger.Error("服务层：注册失败，用户名已存在", zap.String("username", req.Username))
		return nil, errors.New("用户名已存在")
	}

	// Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("服务层：注册失败，密码加密错误", zap.String("username", req.Username), zap.Error(err))
		return nil, err
	}

	// Determine user role - default to 'user' if not specified
	userRole := "user"
	if len(role) > 0 && role[0] != "" {
		userRole = role[0]
	}

	// Create user object with generated UUID and timestamps
	user := &models.User{
		ID:        uuid.New().String(),
		Username:  req.Username,
		Password:  string(hashedPassword),
		Role:      userRole, // Use specified role or default 'user'
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save user to database
	if err := s.db.CreateUser(user); err != nil {
		logger.Error("服务层：注册失败，保存用户时出错", zap.String("username", req.Username), zap.Error(err))
		return nil, err
	}

	logger.Info("服务层：用户注册成功", zap.String("user_id", user.ID), zap.String("username", user.Username), zap.String("role", userRole))

	return user, nil
}

// Login authenticates a user and returns the user object
func (s *UserService) Login(req *models.LoginRequest) (*models.User, error) {
	// Check if database is initialized
	if s.db == nil {
		logger.Error("服务层：登录失败，数据库未初始化")
		return nil, errors.New("系统尚未配置数据库，请先完成初始设置")
	}

	logger.Info("服务层：开始用户登录", zap.String("username", req.Username))

	// Find user by username
	user, err := s.db.GetUserByUsername(req.Username)
	if err != nil {
		logger.Error("服务层：登录失败，查询用户时出错", zap.String("username", req.Username), zap.Error(err))
		return nil, errors.New("用户名或密码错误")
	}
	if user == nil {
		logger.Error("服务层：登录失败，用户不存在", zap.String("username", req.Username))
		return nil, errors.New("用户名或密码错误")
	}

	// Verify password hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		logger.Error("服务层：登录失败，密码错误", zap.String("user_id", user.ID))
		return nil, errors.New("用户名或密码错误")
	}

	logger.Info("服务层：用户登录成功", zap.String("user_id", user.ID), zap.String("username", user.Username))

	return user, nil
}

// GetUserByID retrieves a user by their unique ID
func (s *UserService) GetUserByID(id string) (*models.User, error) {
	// Check if database is initialized
	if s.db == nil {
		logger.Error("服务层：获取用户失败，数据库未初始化")
		return nil, errors.New("系统尚未配置数据库，请先完成初始设置")
	}

	logger.Info("服务层：开始根据ID获取用户", zap.String("user_id", id))

	user, err := s.db.GetUserByID(id)
	if err != nil {
		logger.Error("服务层：根据ID获取用户失败", zap.String("user_id", id), zap.Error(err))
		return nil, err
	}

	if user == nil {
		logger.Error("服务层：用户不存在", zap.String("user_id", id))
		return nil, errors.New("用户不存在")
	}

	logger.Info("服务层：成功根据ID获取用户", zap.String("user_id", id), zap.String("username", user.Username))

	return user, nil
} // GetAllUsers retrieves all users from the database
func (s *UserService) GetAllUsers() []*models.User {
	// Check if database is initialized
	if s.db == nil {
		logger.Warn("服务层：数据库未初始化，返回空用户列表")
		return []*models.User{}
	}

	logger.Info("服务层：获取所有用户")

	users, err := s.db.GetAllUsers()
	if err != nil {
		logger.Error("服务层：获取所有用户失败", zap.Error(err))
		return []*models.User{}
	}

	return users
}

// HasAdminUser checks if any admin user exists in the system
func (s *UserService) HasAdminUser() (bool, error) {
	// Check if database is initialized
	if s.db == nil {
		logger.Warn("服务层：数据库未初始化，无法检查管理员用户")
		return false, errors.New("数据库未初始化")
	}

	logger.Info("服务层：检查是否已存在管理员用户")

	// Check database for admin user existence
	hasAdmin, err := s.db.HasAdminUser()
	if err != nil {
		logger.Error("服务层：检查管理员用户失败", zap.Error(err))
		return false, err
	}

	logger.Info("服务层：检查管理员用户完成", zap.Bool("hasAdmin", hasAdmin))
	return hasAdmin, nil
}

// UpdateUser updates a user's information
func (s *UserService) UpdateUser(user *models.User) error {
	// Check if database is initialized
	if s.db == nil {
		logger.Error("服务层：更新用户失败，数据库未初始化")
		return errors.New("系统尚未配置数据库，请先完成初始设置")
	}

	logger.Info("服务层：开始更新用户信息", zap.String("user_id", user.ID))

	// Update user in database
	if err := s.db.UpdateUser(user); err != nil {
		logger.Error("服务层：更新用户失败，保存用户时出错", zap.String("user_id", user.ID), zap.Error(err))
		return err
	}

	logger.Info("服务层：用户信息更新成功", zap.String("user_id", user.ID))
	return nil
}

// ChangePassword handles user password change requests
func (s *UserService) ChangePassword(userID, oldPassword, newPassword string) error {
	// Check if database is initialized
	if s.db == nil {
		logger.Error("服务层：修改密码失败，数据库未初始化")
		return errors.New("系统尚未配置数据库，请先完成初始设置")
	}

	logger.Info("服务层：开始修改密码", zap.String("user_id", userID))

	// Get user by ID
	user, err := s.db.GetUserByID(userID)
	if err != nil {
		logger.Error("服务层：修改密码失败，获取用户时出错", zap.String("user_id", userID), zap.Error(err))
		return err
	}

	if user == nil {
		logger.Error("服务层：修改密码失败，用户不存在", zap.String("user_id", userID))
		return errors.New("用户不存在")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		logger.Error("服务层：修改密码失败，原密码错误", zap.String("user_id", userID))
		return errors.New("原密码错误")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("服务层：修改密码失败，密码加密错误", zap.String("user_id", userID), zap.Error(err))
		return err
	}

	// Update user password
	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := s.db.UpdateUser(user); err != nil {
		logger.Error("服务层：修改密码失败，更新用户时出错", zap.String("user_id", userID), zap.Error(err))
		return err
	}

	logger.Info("服务层：密码修改成功", zap.String("user_id", userID))
	return nil
}
