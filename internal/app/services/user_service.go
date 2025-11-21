package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tairitsu/tairitsu/internal/app/models"
	"golang.org/x/crypto/bcrypt"
)

// UserService 用户服务
type UserService struct {
	users map[string]*models.User // 内存存储，实际项目中应使用数据库
}

// NewUserService 创建用户服务实例
func NewUserService() *UserService {
	return &UserService{
		users: make(map[string]*models.User),
	}
}

// Register 用户注册
func (s *UserService) Register(req *models.RegisterRequest) (*models.User, error) {
	// 检查用户名是否已存在
	for _, user := range s.users {
		if user.Username == req.Username {
			return nil, errors.New("用户名已存在")
		}
		if user.Email == req.Email {
			return nil, errors.New("邮箱已被注册")
		}
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &models.User{
		ID:        uuid.New().String(),
		Username:  req.Username,
		Password:  string(hashedPassword),
		Email:     req.Email,
		Role:      "user", // 默认角色
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 存储用户
	s.users[user.ID] = user

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(req *models.LoginRequest) (*models.User, error) {
	// 查找用户
	var user *models.User
	for _, u := range s.users {
		if u.Username == req.Username {
			user = u
			break
		}
	}

	if user == nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 验证密码
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id string) (*models.User, error) {
	user, exists := s.users[id]
	if !exists {
		return nil, errors.New("用户不存在")
	}
	return user, nil
}

// GetAllUsers 获取所有用户
func (s *UserService) GetAllUsers() []*models.User {
	users := make([]*models.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users
}