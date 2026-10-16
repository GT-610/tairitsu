package testutil

import (
	"testing"

	"github.com/GT-610/tairitsu/internal/app/config"
)

// UseTempDataDir points the configuration file at a temporary directory for the
// duration of one test.
//
// The configuration is written next to the process working directory by default,
// so without this a test that saves setup state drops a data/config.json into
// the source tree, where the next run reads it back and the failures depend on
// the order the tests happened to run in.
func UseTempDataDir(t *testing.T) {
	t.Helper()

	t.Setenv(config.DataDirEnv, t.TempDir())
}
