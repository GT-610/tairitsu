package handlers

import (
	"net/http"

	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	userService *services.UserService
	jwtService  *services.JWTService
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler(userService *services.UserService, jwtService *services.JWTService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtService:  jwtService,
	}
}

// Register handles user registration requests
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind registration request parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Info("Starting user registration", zap.String("username", req.Username), zap.String("email", req.Email))

	// Determine user role - first registered user becomes admin
	role := "user"
	hasAdmin, err := h.userService.HasAdminUser()
	if err != nil {
		// If admin check fails, log warning but continue with default role
		logger.Warn("Failed to check admin user, using default role", zap.Error(err))
	} else if !hasAdmin {
		// If no admin exists, set current user as admin
		role = "admin"
		logger.Info("No admin user in system, setting current user as admin", zap.String("username", req.Username))
	}

	user, err := h.userService.Register(&req, role)
	if err != nil {
		logger.Error("Failed to register user", zap.String("username", req.Username), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Info("User registered successfully", zap.String("user_id", user.ID), zap.String("username", user.Username), zap.String("role", user.Role))

	// Initialize ZeroTier client after successful registration
	// This ensures the client is instantiated before any ZeroTier operations
	logger.Info("Initializing ZeroTier client")
	ztClient, err := zerotier.NewClient()
	if err != nil {
		// Log error but don't block registration
		logger.Warn("Failed to initialize ZeroTier client, will try again later", zap.Error(err))
	} else {
		// Verify connection if client initialized successfully
		_, err = ztClient.GetStatus()
		if err != nil {
			logger.Warn("Failed to verify ZeroTier connection, will try again later", zap.Error(err))
		} else {
			logger.Info("ZeroTier client initialized and connection verified successfully")
			// Connection verification only - client will be created on-demand when needed
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":    user.ToResponse(),
		"message": "Registration successful",
	})
}

// Login handles user authentication requests
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind login request parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Info("User attempting to login", zap.String("username", req.Username))

	user, err := h.userService.Login(&req)
	if err != nil {
		logger.Error("Failed to login user", zap.String("username", req.Username), zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	logger.Info("User logged in successfully", zap.String("user_id", user.ID), zap.String("username", user.Username))

	// Generate JWT token for authenticated user
	token, err := h.jwtService.GenerateToken(user)
	if err != nil {
		logger.Error("Failed to generate JWT token", zap.String("user_id", user.ID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	logger.Info("JWT token generated successfully", zap.String("user_id", user.ID))

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user.ToResponse(),
	})
}

// GetProfile retrieves the authenticated user's profile information
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("Failed to get user information: not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	logger.Info("Getting user information", zap.String("user_id", userID.(string)))

	user, err := h.userService.GetUserByID(userID.(string))
	if err != nil {
		logger.Error("Failed to get user information", zap.String("user_id", userID.(string)), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	logger.Info("Successfully got user information", zap.String("user_id", user.ID), zap.String("username", user.Username))

	c.JSON(http.StatusOK, user.ToResponse())
}
