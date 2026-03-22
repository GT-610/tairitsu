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

	if db == nil {
		db = database.GetGlobalDB()
	}
	return db
}

func NewUserServiceWithDB(db database.DBInterface) *UserService {
	return &UserService{
		db: db,
	}
}

func NewUserServiceWithoutDB() *UserService {
	return &UserService{
		db: nil,
	}
}

func NewUserService() *UserService {
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

func (s *UserService) Register(req *models.RegisterRequest, role ...string) (*models.User, error) {
	db := s.getDB()
	if db == nil {
		logger.Error("服务层：注册失败，数据库未初始化")
		return nil, errors.New("系统尚未配置数据库，请先完成初始设置")
	}

	logger.Info("服务层：开始用户注册", zap.String("username", req.Username))

	existingUser, err := db.GetUserByUsername(req.Username)
	if err != nil {
		logger.Error("服务层：注册失败，检查用户名时出错", zap.String("username", req.Username), zap.Error(err))
		return nil, err
	}
	if existingUser != nil {
		logger.Error("服务层：注册失败，用户名已存在", zap.String("username", req.Username))
		return nil, errors.New("用户名已存在")
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
		Username:  req.Username,
		Password:  string(hashedPassword),
		Role:      userRole,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.CreateUser(user); err != nil {
		logger.Error("服务层：注册失败，保存用户时出错", zap.String("username", req.Username), zap.Error(err))
		return nil, err
	}

	logger.Info("服务层：用户注册成功", zap.String("user_id", user.ID), zap.String("username", user.Username), zap.String("role", userRole))

	return user, nil
}

func (s *UserService) Login(req *models.LoginRequest) (*models.User, error) {
	db := s.getDB()
	if db == nil {
		logger.Error("服务层：登录失败，数据库未初始化")
		return nil, errors.New("系统尚未配置数据库，请先完成初始设置")
	}

	logger.Info("服务层：开始用户登录", zap.String("username", req.Username))

	user, err := db.GetUserByUsername(req.Username)
	if err != nil {
		logger.Error("服务层：登录失败，查询用户时出错", zap.String("username", req.Username), zap.Error(err))
		return nil, errors.New("用户名或密码错误")
	}
	if user == nil {
		logger.Error("服务层：登录失败，用户不存在", zap.String("username", req.Username))
		return nil, errors.New("用户名或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		logger.Error("服务层：登录失败，密码错误", zap.String("user_id", user.ID))
		return nil, errors.New("用户名或密码错误")
	}

	logger.Info("服务层：用户登录成功", zap.String("user_id", user.ID), zap.String("username", user.Username))

	return user, nil
}

func (s *UserService) GetUserByID(id string) (*models.User, error) {
	db := s.getDB()
	if db == nil {
		logger.Error("服务层：获取用户失败，数据库未初始化")
		return nil, errors.New("系统尚未配置数据库，请先完成初始设置")
	}

	logger.Info("服务层：开始根据ID获取用户", zap.String("user_id", id))

	user, err := db.GetUserByID(id)
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
		return false, errors.New("数据库未初始化")
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
		return errors.New("系统尚未配置数据库，请先完成初始设置")
	}

	logger.Info("服务层：开始更新用户信息", zap.String("user_id", user.ID))

	if err := db.UpdateUser(user); err != nil {
		logger.Error("服务层：更新用户失败，保存用户时出错", zap.String("user_id", user.ID), zap.Error(err))
		return err
	}

	logger.Info("服务层：用户信息更新成功", zap.String("user_id", user.ID))
	return nil
}

func (s *UserService) ChangePassword(userID, oldPassword, newPassword string) error {
	db := s.getDB()
	if db == nil {
		logger.Error("服务层：修改密码失败，数据库未初始化")
		return errors.New("系统尚未配置数据库，请先完成初始设置")
	}

	logger.Info("服务层：开始修改密码", zap.String("user_id", userID))

	user, err := db.GetUserByID(userID)
	if err != nil {
		logger.Error("服务层：修改密码失败，获取用户时出错", zap.String("user_id", userID), zap.Error(err))
		return err
	}

	if user == nil {
		logger.Error("服务层：修改密码失败，用户不存在", zap.String("user_id", userID))
		return errors.New("用户不存在")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		logger.Error("服务层：修改密码失败，原密码错误", zap.String("user_id", userID))
		return errors.New("原密码错误")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("服务层：修改密码失败，密码加密错误", zap.String("user_id", userID), zap.Error(err))
		return err
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := db.UpdateUser(user); err != nil {
		logger.Error("服务层：修改密码失败，更新用户时出错", zap.String("user_id", userID), zap.Error(err))
		return err
	}

	logger.Info("服务层：密码修改成功", zap.String("user_id", userID))
	return nil
}
