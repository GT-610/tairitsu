package services

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tairitsu/tairitsu/internal/app/database"
	"github.com/tairitsu/tairitsu/internal/app/logger"
	"github.com/tairitsu/tairitsu/internal/app/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// UserService 用户服务
type UserService struct {
	db    database.DBInterface // 数据库接口
	mutex sync.RWMutex         // 读写锁，保护并发访问
}

// NewUserServiceWithDB 创建带数据库的用户服务实例
func NewUserServiceWithDB(db database.DBInterface) *UserService {
	return &UserService{
		db: db,
	}
}

// NewUserService 创建用户服务实例（使用默认的SQLite数据库）
func NewUserService() *UserService {
	// 创建默认的SQLite数据库实现
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

// Register 用户注册
func (s *UserService) Register(req *models.RegisterRequest) (*models.User, error) {
	logger.Info("服务层：开始用户注册", zap.String("username", req.Username), zap.String("email", req.Email))
	
	// 检查用户名是否已存在
	existingUser, err := s.db.GetUserByUsername(req.Username)
	if err != nil {
		logger.Error("服务层：注册失败，检查用户名时出错", zap.String("username", req.Username), zap.Error(err))
		return nil, err
	}
	if existingUser != nil {
		logger.Error("服务层：注册失败，用户名已存在", zap.String("username", req.Username))
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	existingUser, err = s.db.GetUserByEmail(req.Email)
	if err != nil {
		logger.Error("服务层：注册失败，检查邮箱时出错", zap.String("email", req.Email), zap.Error(err))
		return nil, err
	}
	if existingUser != nil {
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
	if err := s.db.CreateUser(user); err != nil {
		logger.Error("服务层：注册失败，保存用户时出错", zap.String("username", req.Username), zap.Error(err))
		return nil, err
	}
	
	logger.Info("服务层：用户注册成功", zap.String("user_id", user.ID), zap.String("username", user.Username))

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(req *models.LoginRequest) (*models.User, error) {
	logger.Info("服务层：用户尝试登录", zap.String("username", req.Username))
	
	// 查找用户
	user, err := s.db.GetUserByUsername(req.Username)
	if err != nil {
		logger.Error("服务层：登录失败，查询用户时出错", zap.String("username", req.Username), zap.Error(err))
		return nil, errors.New("用户名或密码错误")
	}
	if user == nil {
		logger.Error("服务层：登录失败，用户不存在", zap.String("username", req.Username))
		return nil, errors.New("用户名或密码错误")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
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
}// GetAllUsers 获取所有用户
func (s *UserService) GetAllUsers() []*models.User {
	users, err := s.db.GetAllUsers()
	if err != nil {
		logger.Error("服务层：获取所有用户失败", zap.Error(err))
		return []*models.User{}
	}
	
	// 转换为指针切片
	result := make([]*models.User, len(users))
	for i, user := range users {
		result[i] = user
	}
	
	return result
}