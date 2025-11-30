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
		panic("Failed to create default database: " + err.Error())
	}

	return &UserService{
		db: db,
	}
}

// Register handles user registration with optional role specification
func (s *UserService) Register(req *models.RegisterRequest, role ...string) (*models.User, error) {
	// Check if database is initialized
	if s.db == nil {
		logger.Error("Service layer: Registration failed, database not initialized")
		return nil, errors.New("System database not configured yet, please complete initial setup first")
	}

	logger.Info("Service layer: Starting user registration", zap.String("username", req.Username), zap.String("email", req.Email))

	// Check if username already exists
	existingUser, err := s.db.GetUserByUsername(req.Username)
	if err != nil {
		logger.Error("Service layer: Registration failed, error checking username", zap.String("username", req.Username), zap.Error(err))
		return nil, err
	}
	if existingUser != nil {
		logger.Error("Service layer: Registration failed, username already exists", zap.String("username", req.Username))
		return nil, errors.New("Username already exists")
	}

	// Check if email already exists
	existingUser, err = s.db.GetUserByEmail(req.Email)
	if err != nil {
		logger.Error("Service layer: Registration failed, error checking email", zap.String("email", req.Email), zap.Error(err))
		return nil, err
	}
	if existingUser != nil {
		logger.Error("Service layer: Registration failed, email already in use", zap.String("email", req.Email))
		return nil, errors.New("Email already in use")
	}

	// Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Service layer: Registration failed, password encryption error", zap.String("username", req.Username), zap.Error(err))
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
		Email:     req.Email,
		Role:      userRole, // Use specified role or default 'user'
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save user to database
	if err := s.db.CreateUser(user); err != nil {
		logger.Error("Service layer: Registration failed, error saving user", zap.String("username", req.Username), zap.Error(err))
		return nil, err
	}

	logger.Info("Service layer: User registered successfully", zap.String("user_id", user.ID), zap.String("username", user.Username), zap.String("role", userRole))

	return user, nil
}

// Login authenticates a user and returns the user object
func (s *UserService) Login(req *models.LoginRequest) (*models.User, error) {
	// Check if database is initialized
	if s.db == nil {
		logger.Error("Service layer: Login failed, database not initialized")
		return nil, errors.New("System database not configured yet, please complete initial setup first")
	}

	logger.Info("Service layer: Starting user login", zap.String("username", req.Username))

	// Find user by username
	user, err := s.db.GetUserByUsername(req.Username)
	if err != nil {
		logger.Error("Service layer: Login failed, error querying user", zap.String("username", req.Username), zap.Error(err))
		return nil, errors.New("Invalid username or password")
	}
	if user == nil {
		logger.Error("Service layer: Login failed, user does not exist", zap.String("username", req.Username))
		return nil, errors.New("Invalid username or password")
	}

	// Verify password hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		logger.Error("Service layer: Login failed, incorrect password", zap.String("user_id", user.ID))
		return nil, errors.New("Invalid username or password")
	}

	logger.Info("Service layer: User logged in successfully", zap.String("user_id", user.ID), zap.String("username", user.Username))

	return user, nil
}

// GetUserByID retrieves a user by their unique ID
func (s *UserService) GetUserByID(id string) (*models.User, error) {
	// Check if database is initialized
	if s.db == nil {
		logger.Error("Service layer: Failed to get user, database not initialized")
		return nil, errors.New("System database not configured yet, please complete initial setup first")
	}

	logger.Info("Service layer: Starting to get user by ID", zap.String("user_id", id))

	user, err := s.db.GetUserByID(id)
	if err != nil {
		logger.Error("Service layer: Failed to get user by ID", zap.String("user_id", id), zap.Error(err))
		return nil, err
	}

	if user == nil {
		logger.Error("Service layer: User does not exist", zap.String("user_id", id))
		return nil, errors.New("User does not exist")
	}

	logger.Info("Service layer: Successfully got user by ID", zap.String("user_id", id), zap.String("username", user.Username))

	return user, nil
} // GetAllUsers retrieves all users from the database
func (s *UserService) GetAllUsers() []*models.User {
	// Check if database is initialized
	if s.db == nil {
		logger.Warn("Service layer: Database not initialized, returning empty user list")
		return []*models.User{}
	}

	logger.Info("Service layer: Getting all users")

	users, err := s.db.GetAllUsers()
	if err != nil {
		logger.Error("Service layer: Failed to get all users", zap.Error(err))
		return []*models.User{}
	}

	return users
}

// HasAdminUser checks if any admin user exists in the system
func (s *UserService) HasAdminUser() (bool, error) {
	// Check if database is initialized
	if s.db == nil {
		logger.Warn("Service layer: Database not initialized, cannot check admin user")
		return false, errors.New("Database not initialized")
	}

	logger.Info("Service layer: Checking if admin user already exists")

	// Check database for admin user existence
	hasAdmin, err := s.db.HasAdminUser()
	if err != nil {
		logger.Error("Service layer: Failed to check admin user", zap.Error(err))
		return false, err
	}

	logger.Info("Service layer: Admin user check completed", zap.Bool("hasAdmin", hasAdmin))
	return hasAdmin, nil
}
