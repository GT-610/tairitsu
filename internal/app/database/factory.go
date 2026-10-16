package database

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"go.uber.org/zap"
)

// DatabaseType names a storage backend. SQLite is the only implementation.
type DatabaseType string

// SQLite is the only supported backend.
const SQLite DatabaseType = "sqlite"

// Config holds the database configuration
type Config struct {
	Type DatabaseType
	Path string // SQLite database path
}

// NewDatabase creates a database instance based on the given configuration
func NewDatabase(dbConfig Config) (DBInterface, error) {
	switch dbConfig.Type {
	case SQLite:
		// Use default path if none is specified
		if dbConfig.Path == "" {
			dbConfig.Path = filepath.Join(config.DataRoot(), "tairitsu.db")
		}

		if err := ensureSQLiteDir(dbConfig.Path); err != nil {
			return nil, err
		}

		// TranslateError turns constraint violations into gorm sentinels so a
		// lost race can be recognised without matching driver message text.
		db, err := gorm.Open(sqlite.Open(dbConfig.Path+"?_journal_mode=WAL&_busy_timeout=5000"), &gorm.Config{TranslateError: true})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get SQLite database instance: %w", err)
		}

		// Configure connection pool
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
		sqlDB.SetConnMaxIdleTime(30 * time.Minute)

		return &GormDB{db: db}, nil

	case "":
		return nil, fmt.Errorf("database type is required")

	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}
}

func LoadConfigFromApp(cfg *config.Config) Config {
	if cfg == nil {
		return Config{}
	}

	return Config{
		Type: DatabaseType(cfg.Database.Type),
		Path: cfg.Database.Path,
	}
}

// ResetDatabase removes the SQLite database file so the next start begins from
// an empty schema.
// WARNING: This operation will delete all data in the database. Use with caution.
//
// The path is resolved and confined to Tairitsu's data directory, and the
// target must actually look like a SQLite database, because the setup wizard
// supplies the path: without both checks a request could name any file on the
// host and have this function unlink it.
func ResetDatabase(dbConfig Config) error {
	logger.Info("starting database reset", zap.String("type", string(dbConfig.Type)))

	switch dbConfig.Type {
	case SQLite:
		safePath, err := config.ResolveDataPath(dbConfig.Path)
		if err != nil {
			logger.Error("refusing to reset the database at a rejected path", zap.String("path", dbConfig.Path), zap.Error(err))
			return fmt.Errorf("refusing to reset the database: %w", err)
		}

		logger.Info("resetting SQLite database", zap.String("path", safePath))

		removed, err := removeSQLiteDatabase(safePath)
		if err != nil {
			logger.Error("failed to reset SQLite database", zap.String("path", safePath), zap.Error(err))
			return err
		}
		if !removed {
			// Not an error: a fresh instance simply has no file yet.
			logger.Info("SQLite database file does not exist; a new file will be created", zap.String("path", safePath))
			return nil
		}

		logger.Info("SQLite database reset successfully", zap.String("path", safePath))
		return nil

	default:
		logger.Error("unsupported database type", zap.String("type", string(dbConfig.Type)))
		return fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}
}

// sqliteFileHeader is the magic string at the start of every SQLite database.
const sqliteFileHeader = "SQLite format 3\x00"

// removeSQLiteDatabase deletes a SQLite database together with its write-ahead
// log siblings, and reports whether anything was removed. It refuses to touch a
// file that is neither empty nor a SQLite database.
func removeSQLiteDatabase(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to inspect SQLite database file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("refusing to reset %q: not a regular file", path)
	}

	// An empty file holds nothing: it is what an interrupted setup leaves behind.
	if info.Size() > 0 {
		looksLikeSQLite, err := hasSQLiteHeader(path)
		if err != nil {
			return false, fmt.Errorf("failed to inspect SQLite database file: %w", err)
		}
		if !looksLikeSQLite {
			return false, fmt.Errorf("refusing to reset %q: not a SQLite database", path)
		}
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to reset SQLite database: %w", err)
	}

	// The -wal and -shm files belong to this database and are derived from the
	// validated path, so they are safe to remove alongside it.
	for _, suffix := range []string{"-wal", "-shm"} {
		if err := os.Remove(path + suffix); err != nil && !os.IsNotExist(err) {
			return true, fmt.Errorf("failed to remove %q: %w", path+suffix, err)
		}
	}

	return true, nil
}

// hasSQLiteHeader reports whether path starts with the SQLite magic string.
func hasSQLiteHeader(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	header := make([]byte, len(sqliteFileHeader))
	if _, err := io.ReadFull(file, header); err != nil {
		// A file shorter than the header cannot be a SQLite database.
		return false, nil
	}

	return string(header) == sqliteFileHeader, nil
}

func ensureSQLiteDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create SQLite database directory: %w", err)
	}

	return nil
}

func SaveConfigToApp(cfg *config.Config, dbConfig Config) error {
	if cfg == nil {
		return fmt.Errorf("configuration is not initialized")
	}

	// Update database configuration
	cfg.Database.Type = string(dbConfig.Type)
	cfg.Database.Path = dbConfig.Path

	return config.SaveConfig(cfg)
}
