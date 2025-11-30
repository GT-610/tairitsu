package handlers

import (
	"crypto/rand"
	"encoding/base64"
	mathrand "math/rand"
	"net/http"
	"time"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SystemHandler handles system-related API endpoints and operations
type SystemHandler struct {
	networkService   *services.NetworkService
	userService      *services.UserService
	reloadRoutesFunc func() // Function to reload application routes
	// Database configuration is stored in config file
}

// NewSystemHandler creates a new system handler instance
func NewSystemHandler(networkService *services.NetworkService, userService *services.UserService, reloadRoutesFunc func()) *SystemHandler {
	return &SystemHandler{
		networkService:   networkService,
		userService:      userService,
		reloadRoutesFunc: reloadRoutesFunc,
	}
}

// GetSystemStatus retrieves the current system status
func (h *SystemHandler) GetSystemStatus(c *gin.Context) {
	// Get system initialization status from config module
	sysConfig := config.AppConfig
	initialized := false

	if sysConfig != nil {
		// Check initialized field from config.json first
		initialized = sysConfig.Initialized

		// If not initialized, return uninitialized status directly
		if !initialized {
			c.JSON(http.StatusOK, map[string]interface{}{
				"initialized": false,
			})
			return
		}
	}

	// If initialized, retrieve additional status information
	hasDatabase := false
	if sysConfig != nil && initialized {
		// Check if database is configured
		hasDatabase = sysConfig.Database.Type != ""
	}

	// Check for admin user only if database is configured
	hasAdmin := false
	if hasDatabase && h.userService != nil {
		users := h.userService.GetAllUsers()
		for _, user := range users {
			if user.Role == "admin" {
				hasAdmin = true
				break
			}
		}
	}

	// Check ZeroTier connection status
	ztStatus, err := h.networkService.GetStatus()
	if err != nil {
		logger.Debug("[System Status] Failed to get ZeroTier status", zap.Error(err))
		// Return system status even if ZeroTier connection fails
		ztStatus = &zerotier.Status{
			Version: "unknown",
			Address: "",
			Online:  false,
		}
	}

	// When initialized, return complete status information
	response := map[string]interface{}{
		"initialized": true,
		"hasDatabase": hasDatabase,
		"hasAdmin":    hasAdmin,
		"ztStatus":    ztStatus,
	}

	c.JSON(http.StatusOK, response)
}

// ConfigureDatabase configures the database connection settings
func (h *SystemHandler) ConfigureDatabase(c *gin.Context) {
	var dbConfig models.DatabaseConfig
	if err := c.ShouldBindJSON(&dbConfig); err != nil {
		logger.Error("Failed to bind database configuration parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Info("Starting database configuration", zap.String("type", dbConfig.Type))

	// Validate database configuration
	dbCfg := database.Config{
		Type: database.DatabaseType(dbConfig.Type),
		Path: dbConfig.Path,
		Host: dbConfig.Host,
		Port: dbConfig.Port,
		User: dbConfig.User,
		Pass: dbConfig.Pass,
		Name: dbConfig.Name,
	}

	// Try connecting to database
	db, err := database.NewDatabase(dbCfg)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to connect to database: " + err.Error()})
		return
	}

	// Initialize database schema
	if err := db.Init(); err != nil {
		logger.Error("Failed to initialize database", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to initialize database: " + err.Error()})
		return
	}

	// Close database connection
	if err := db.Close(); err != nil {
		logger.Warn("Warning when closing database connection", zap.Error(err))
	}

	// For SQLite, ensure path is properly saved to config
	// NewDatabase function might set default path if Path is empty
	if dbCfg.Type == database.SQLite {
		// From factory.go we know if Path is empty, default value "data/tairitsu.db" will be used
		if dbConfig.Path == "" {
			dbCfg.Path = "data/tairitsu.db"
		}
		logger.Info("SQLite database path has been set", zap.String("path", dbCfg.Path))
	}

	// Save database configuration to unified config management module
	if err := database.SaveConfig(dbCfg); err != nil {
		logger.Error("Failed to save database configuration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save database configuration: " + err.Error()})
		return
	}

	logger.Info("Database configuration successful", zap.String("type", dbConfig.Type))
	c.JSON(http.StatusOK, gin.H{
		"message": "Database configuration successful",
		"config":  dbConfig,
	})
}

// TestZeroTierConnection tests connectivity to the ZeroTier controller
func (h *SystemHandler) TestZeroTierConnection(c *gin.Context) {
	// Dynamically create ZeroTier client
	ztClient, err := zerotier.NewClient()
	if err != nil {
		logger.Error("[ZeroTier] Connection test failed: failed to create client", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ZeroTier client: " + err.Error()})
		return
	}

	// Get ZeroTier controller status
	ztStatus, err := ztClient.GetStatus()
	if err != nil {
		logger.Error("[ZeroTier] Connection test failed: failed to get status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to ZeroTier controller: " + err.Error()})
		return
	}

	logger.Info("[ZeroTier] Connection test successful")
	c.JSON(http.StatusOK, ztStatus)
}

// InitZeroTierClient initializes the ZeroTier client for the application
func (h *SystemHandler) InitZeroTierClient(c *gin.Context) {

	// Dynamically create ZeroTier client
	ztClient, err := zerotier.NewClient()
	if err != nil {
		logger.Error("Failed to create ZeroTier client", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ZeroTier client: " + err.Error()})
		return
	}

	// Set client in network service
	h.networkService.SetZTClient(ztClient)

	// Verify client is working properly
	status, err := h.networkService.GetStatus()
	if err != nil {
		logger.Error("ZeroTier client verification failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ZeroTier client verification failed after initialization: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ZeroTier client initialization successful", "status": status})
}

// SaveZeroTierConfig saves ZeroTier configuration and initializes connection
func (h *SystemHandler) SaveZeroTierConfig(c *gin.Context) {
	var req struct {
		ControllerURL string `json:"controllerUrl" binding:"required"`
		TokenPath     string `json:"tokenPath" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind ZeroTier configuration parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Info("Saving ZeroTier configuration", zap.String("controllerUrl", req.ControllerURL), zap.String("tokenPath", req.TokenPath))

	// Call config module to save ZeroTier configuration
	if err := config.SetZTConfig(req.ControllerURL, req.TokenPath); err != nil {
		logger.Error("Failed to save ZeroTier configuration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save ZeroTier configuration: " + err.Error()})
		return
	}

	// Try creating and validating ZeroTier client with new config
	ztClient, err := zerotier.NewClient()
	if err != nil {
		logger.Error("Failed to create ZeroTier client", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ZeroTier client: " + err.Error()})
		return
	}

	// Validate client works and get status information
	status, err := ztClient.GetStatus()
	if err != nil {
		logger.Error("ZeroTier client verification failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ZeroTier client verification failed: " + err.Error()})
		return
	}

	// Set client in network service
	h.networkService.SetZTClient(ztClient)

	logger.Info("ZeroTier configuration saved and verified successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "ZeroTier configuration saved successfully",
		"config":  req,
		"status":  status,
	})
}

// InitializeAdminCreation prepares the system for admin account creation
// This function is called when user enters the admin creation step to ensure correct database state
func (h *SystemHandler) InitializeAdminCreation(c *gin.Context) {
	logger.Info("Initializing admin account creation step")

	// Check if reset operation has already been performed
	resetDoneKey := "admin_creation_reset_done"
	resetDone := config.GetTempSetting(resetDoneKey)

	// Skip reset if already performed
	if resetDone == "true" {
		logger.Info("Database reset was performed earlier, skipping reset operation")
		c.JSON(http.StatusOK, gin.H{
			"message":   "Admin account creation step has been initialized",
			"resetDone": true,
		})
		return
	}

	// Get current database configuration
	dbConfig := database.LoadConfig()
	logger.Info("Getting database configuration", zap.String("type", string(dbConfig.Type)))

	// Only perform reset for SQLite databases
	if dbConfig.Type == database.SQLite {
		logger.Info("Preparing to reset SQLite database")

		// Execute database reset
		if err := database.ResetDatabase(dbConfig); err != nil {
			logger.Error("Failed to reset database", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize admin account creation step: " + err.Error()})
			return
		}

		logger.Info("SQLite database reset successful")
	} else {
		logger.Info("Current database type is not SQLite, skipping database reset", zap.String("type", string(dbConfig.Type)))
	}

	// Set flag indicating reset operation is complete
	config.SetTempSetting(resetDoneKey, "true")
	logger.Info("Setting reset operation completion flag")

	c.JSON(http.StatusOK, gin.H{
		"message":      "Admin account creation step initialized successfully",
		"resetDone":    true,
		"databaseType": string(dbConfig.Type),
	})
}

// SetInitialized updates the system initialization status
func (h *SystemHandler) SetInitialized(c *gin.Context) {
	var req struct {
		Initialized bool `json:"initialized" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind initialization status parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Info("Setting system initialization status", zap.Bool("initialized", req.Initialized))

	// Generate JWT and session secrets if not already set
	if req.Initialized {
		logger.Info("Starting to generate security keys")

		// Generate random JWT secret if not already set
		if config.AppConfig.Security.JWTSecret == "" {
			// Generate 32-byte random secret
			jwtSecret := generateRandomSecret(32)
			config.AppConfig.Security.JWTSecret = jwtSecret
			logger.Info("Generated new JWT secret")
		}

		// Generate random session secret if not already set
		if config.AppConfig.Security.SessionSecret == "" {
			// Generate 32-byte random secret
			sessionSecret := generateRandomSecret(32)
			config.AppConfig.Security.SessionSecret = sessionSecret
			logger.Info("Generated new session secret")
		}

		// Save updated configuration with new secrets
		if err := config.SaveConfig(config.AppConfig); err != nil {
			logger.Error("Failed to save secret configuration", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate security keys: " + err.Error()})
			return
		}
	}

	// Call config module to set initialization status
	if err := config.SetInitialized(req.Initialized); err != nil {
		logger.Error("Failed to set initialization status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set initialization status: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Initialization status updated successfully"})
}

// generateRandomSecret generates a random secret string of specified length
func generateRandomSecret(length int) string {
	// Use crypto/rand to generate secure random bytes
	// This is a simple implementation, in production you might want to use a more robust method
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a less secure method if crypto/rand fails
		logger.Warn("Failed to generate random secret using crypto/rand, will use math/rand as fallback", zap.Error(err))

		// Use math/rand as fallback
		r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
		for i := range bytes {
			bytes[i] = byte(r.Intn(256))
		}
	}

	// Encode to base64 for a safe string representation
	return base64.URLEncoding.EncodeToString(bytes)
}

// ReloadRoutes handles API request to reload application routes
func (h *SystemHandler) ReloadRoutes(c *gin.Context) {

	// Call internal reload function
	if h.reloadRoutesFunc != nil {
		h.reloadRoutesFunc()
		c.JSON(http.StatusOK, gin.H{"message": "Routes reloaded successfully"})
	} else {
		logger.Warn("Route reload function not defined")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Route reload functionality unavailable"})
	}
}