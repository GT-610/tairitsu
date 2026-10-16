package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The database path arrives in a setup request and a later reset deletes
// whatever it names, so anything outside the mounted data directory has to be
// refused rather than merely discouraged.
func TestResolveDataPathConfinesPathsToTheDataDirectory(t *testing.T) {
	dataRoot := t.TempDir()
	t.Setenv(config.DataDirEnv, dataRoot)

	t.Run("empty selects the default file", func(t *testing.T) {
		resolved, err := config.ResolveDataPath("")
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(dataRoot, "tairitsu.db"), resolved)
	})

	t.Run("absolute path inside the root", func(t *testing.T) {
		candidate := filepath.Join(dataRoot, "absolute.db")
		resolved, err := config.ResolveDataPath(candidate)
		require.NoError(t, err)
		assert.Equal(t, candidate, resolved)
	})

	t.Run("nested path that does not exist yet", func(t *testing.T) {
		candidate := filepath.Join(dataRoot, "nested", "deeper", "tairitsu.db")
		resolved, err := config.ResolveDataPath(candidate)
		require.NoError(t, err)
		assert.Equal(t, candidate, resolved)
	})

	t.Run("traversal out of the root is rejected", func(t *testing.T) {
		_, err := config.ResolveDataPath(filepath.Join(dataRoot, "..", "escape.db"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "outside the Tairitsu data directory")
	})

	t.Run("an absolute path elsewhere is rejected", func(t *testing.T) {
		_, err := config.ResolveDataPath("/etc/passwd")
		require.Error(t, err)
	})

	t.Run("the data directory itself is rejected", func(t *testing.T) {
		_, err := config.ResolveDataPath(dataRoot)
		require.Error(t, err)
	})

	t.Run("a sibling with the same prefix is rejected", func(t *testing.T) {
		_, err := config.ResolveDataPath(dataRoot + "-backup/tairitsu.db")
		require.Error(t, err)
	})
}

// A relative path keeps its historical meaning: it is resolved against the
// working directory, which is what the documented `data/tairitsu.db` default
// relies on when the data directory is not relocated.
func TestResolveDataPathResolvesRelativePathsAgainstTheWorkingDirectory(t *testing.T) {
	dataRoot := t.TempDir()
	t.Setenv(config.DataDirEnv, dataRoot)
	t.Chdir(dataRoot)

	resolved, err := config.ResolveDataPath("nested/tairitsu.db")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dataRoot, "nested", "tairitsu.db"), resolved)

	_, err = config.ResolveDataPath("../escape.db")
	require.Error(t, err)
}

// A symlink inside the data directory must not be usable to reach outside it,
// including when the target file does not exist yet.
func TestResolveDataPathRejectsSymlinkEscapes(t *testing.T) {
	dataRoot := t.TempDir()
	t.Setenv(config.DataDirEnv, dataRoot)

	outside := t.TempDir()
	link := filepath.Join(dataRoot, "link")
	require.NoError(t, os.Symlink(outside, link))

	_, err := config.ResolveDataPath(filepath.Join(link, "tairitsu.db"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside the Tairitsu data directory")
}

func TestDataRootFollowsTheEnvironment(t *testing.T) {
	dataRoot := t.TempDir()
	t.Setenv(config.DataDirEnv, dataRoot)

	assert.Equal(t, dataRoot, config.DataRoot())
	assert.Equal(t, filepath.Join(dataRoot, "config.json"), config.ConfigFilePath())
}

func TestDataRootDefaultsToTheWorkingDirectory(t *testing.T) {
	workingDirectory := t.TempDir()
	t.Setenv(config.DataDirEnv, "")
	t.Chdir(workingDirectory)

	assert.Equal(t, filepath.Join(workingDirectory, "data"), config.DataRoot())
}
