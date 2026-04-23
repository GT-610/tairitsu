package initializer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/database"
	appinit "github.com/GT-610/tairitsu/internal/app/initializer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseInitializerWithConfigUsesBoundConfig(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: config.SQLite,
			Path: filepath.Join(t.TempDir(), "compat.db"),
		},
	}

	initializer := appinit.NewDatabaseInitializerWithConfig(cfg)
	db, err := initializer.Initialize()
	require.NoError(t, err)
	require.NotNil(t, db)
	t.Cleanup(func() {
		_ = initializer.Close()
	})
}

func TestZeroTierInitializerWithConfigSaveConfigUsesBoundConfig(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			JWTSecret: "test-secret",
		},
	}

	initializer := appinit.NewZeroTierInitializerWithConfig(cfg)
	require.NoError(t, initializer.SaveConfig("http://127.0.0.1:9993", "encrypted:token"))

	assert.Equal(t, "http://127.0.0.1:9993", cfg.ZeroTier.URL)
	assert.Equal(t, "encrypted:token", cfg.ZeroTier.Token)
}

func TestInitGlobalDBFromAppConfigUsesExplicitConfig(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "global.db")
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: config.SQLite,
			Path: dbPath,
		},
	}

	require.NoError(t, database.InitGlobalDBFromAppConfig(cfg))
	require.NotNil(t, database.GetGlobalDB())
	require.NoError(t, database.CloseGlobalDB())

	_, err := os.Stat(dbPath)
	require.NoError(t, err)
}
