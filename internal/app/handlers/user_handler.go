package handlers

import (
	"net/http"

	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new instance of UserHandler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetAllUsers retrieves all users
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	logger.Info("开始获取所有用户")

	// Get all users from service
	users := h.userService.GetAllUsers()

	// Convert users to response format
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	logger.Info("成功获取所有用户", zap.Int("count", len(userResponses)))
	c.JSON(http.StatusOK, userResponses)
}

// UpdateUserRole updates a user's role
type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin user"`
}

func (h *UserHandler) UpdateUserRole(c *gin.Context) {
	userId := c.Param("userId")
	logger.Info("开始更新用户角色", zap.String("user_id", userId))

	// Bind and validate request body
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("更新用户角色失败，请求参数绑定失败", zap.String("user_id", userId), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user by ID
	user, err := h.userService.GetUserByID(userId)
	if err != nil {
		logger.Error("更新用户角色失败，获取用户时出错", zap.String("user_id", userId), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Update user role
	user.Role = req.Role

	// Save updated user to database
	if err := h.userService.UpdateUser(user); err != nil {
		logger.Error("更新用户角色失败，保存用户时出错", zap.String("user_id", userId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户角色失败"})
		return
	}

	logger.Info("成功更新用户角色", zap.String("user_id", userId), zap.String("new_role", req.Role))
	c.JSON(http.StatusOK, user.ToResponse())
}