package initializer

import (
	"fmt"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"go.uber.org/zap"
)

// DatabaseInitializer Database initializer
type DatabaseInitializer struct {
	db database.DBInterface
}

// NewDatabaseInitializer Create a new database initializer
func NewDatabaseInitializer() *DatabaseInitializer {
	return &DatabaseInitializer{}
}

// Initialize Initialize database
func (di *DatabaseInitializer) Initialize() (database.DBInterface, error) {
	// Load database configuration
	dbConfig := database.LoadConfig()

	// If database type is not configured, skip initialization
	if dbConfig.Type == "" {
		logger.Info("Database type not configured, skipping database initialization")
		return nil, nil
	}

	logger.Info("Starting database initialization", zap.String("type", string(dbConfig.Type)))

	// Create database instance
	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database instance: %w", err)
	}

	// Initialize database schema
	if err := db.Init(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	di.db = db
	logger.Info("Database initialization completed", zap.String("type", string(dbConfig.Type)))
	return db, nil
}

// GetDatabase Get initialized database instance
func (di *DatabaseInitializer) GetDatabase() database.DBInterface {
	return di.db
}

// Close Close database connection
func (di *DatabaseInitializer) Close() error {
	if di.db != nil {
		return di.db.Close()
	}
	return nil
}

// ResetDatabase Reset database (for testing or reinitialization)
func (di *DatabaseInitializer) ResetDatabase() error {
	if di.db != nil {
		di.Close()
	}

	dbConfig := database.LoadConfig()
	if dbConfig.Type == "" {
		return fmt.Errorf("database type not configured")
	}

	return database.ResetDatabase(dbConfig)
}
