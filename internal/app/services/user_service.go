package services

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tairitsu/tairitsu/internal/app/logger"
	"github.com/tairitsu/tairitsu/internal/app/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// UserService 用户服务
type UserService struct {
	users           map[string]*models.User // 内存存储，实际项目中应使用数据库
	usersByUsername map[string]*models.User // 按用户名索引的用户映射
	usersByEmail    map[string]*models.User // 按邮箱索引的用户映射
	mutex           sync.RWMutex            // 读写锁，保护并发访问
}

// NewUserService 创建用户服务实例
func NewUserService() *UserService {
	return &UserService{
		users:           make(map[string]*models.User),
		usersByUsername: make(map[string]*models.User),
		usersByEmail:    make(map[string]*models.User),
	}
}

// Register 用户注册
func (s *UserService) Register(req *models.RegisterRequest) (*models.User, error) {
	logger.Info("服务层：开始用户注册", zap.String("username", req.Username), zap.String("email", req.Email))
	
	// 检查用户名是否已存在
	if _, exists := s.usersByUsername[req.Username]; exists {
		logger.Error("服务层：注册失败，用户名已存在", zap.String("username", req.Username))
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if _, exists := s.usersByEmail[req.Email]; exists {
		logger.Error("服务层：注册失败，邮箱已被使用", zap.String("email", req.Email))
		return nil, errors.New("邮箱已被使用")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("服务层：注册失败，密码加密错误", zap.String("username", req.Username), zap.Error(err))
		return nil, err
	}

	// 创建用户
	user := &models.User{
		ID:        uuid.New().String(),
		Username:  req.Username,
		Password:  string(hashedPassword),
		Email:     req.Email,
		Role:      "user", // 默认角色为普通用户
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 保存用户
	s.mutex.Lock()
	s.users[user.ID] = user
	s.usersByUsername[user.Username] = user
	s.usersByEmail[user.Email] = user
	s.mutex.Unlock()
	
	logger.Info("服务层：用户注册成功", zap.String("user_id", user.ID), zap.String("username", user.Username))

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(req *models.LoginRequest) (*models.User, error) {
	logger.Info("服务层：用户尝试登录", zap.String("username", req.Username))
	
	// 查找用户
	user, exists := s.usersByUsername[req.Username]
	if !exists {
		logger.Error("服务层：登录失败，用户不存在", zap.String("username", req.Username))
		return nil, errors.New("用户名或密码错误")
	}

	// 验证密码
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		logger.Error("服务层：登录失败，密码错误", zap.String("username", req.Username))
		return nil, errors.New("用户名或密码错误")
	}
	
	logger.Info("服务层：用户登录成功", zap.String("user_id", user.ID), zap.String("username", user.Username))

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id string) (*models.User, error) {
	logger.Info("服务层：开始根据ID获取用户", zap.String("user_id", id))
	
	s.mutex.RLock()
	user, exists := s.users[id]
	s.mutex.RUnlock()
	
	if !exists {
		logger.Error("服务层：用户不存在", zap.String("user_id", id))
		return nil, errors.New("用户不存在")
	}
	
	logger.Info("服务层：成功根据ID获取用户", zap.String("user_id", id), zap.String("username", user.Username))

	return user, nil
}// GetAllUsers 获取所有用户
func (s *UserService) GetAllUsers() []*models.User {
	users := make([]*models.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users
}