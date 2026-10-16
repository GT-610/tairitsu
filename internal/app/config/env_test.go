package config

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// resetDotEnvCache clears the once-per-process cache so a test can install its
// own .env file. Production reads the file once and keeps it.
func resetDotEnvCache(t *testing.T) {
	t.Helper()

	dotEnvOnce = sync.Once{}
	dotEnvSettings = nil
	t.Cleanup(func() {
		dotEnvOnce = sync.Once{}
		dotEnvSettings = nil
	})
}

func writeDotEnv(t *testing.T, content string) {
	t.Helper()

	t.Chdir(t.TempDir())
	if err := os.WriteFile(dotEnvFileName, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write %s: %v", dotEnvFileName, err)
	}
	resetDotEnvCache(t)
}

func TestEnvReadsTheProcessEnvironment(t *testing.T) {
	writeDotEnv(t, "")
	t.Setenv("TAIRITSU_TEST_VALUE", "from-environment")

	assert.Equal(t, "from-environment", Env("TAIRITSU_TEST_VALUE"))
	assert.Empty(t, Env("TAIRITSU_TEST_ABSENT"))
}

func TestEnvReadsTheDotEnvFile(t *testing.T) {
	writeDotEnv(t, "# a comment\n\nTAIRITSU_TEST_VALUE=from-file\n")

	assert.Equal(t, "from-file", Env("TAIRITSU_TEST_VALUE"))
}

func TestEnvPrefersTheProcessEnvironmentOverTheFile(t *testing.T) {
	writeDotEnv(t, "TAIRITSU_TEST_VALUE=from-file\n")
	t.Setenv("TAIRITSU_TEST_VALUE", "from-environment")

	assert.Equal(t, "from-environment", Env("TAIRITSU_TEST_VALUE"))
}

// `FOO=` in a container definition must not shadow a value the operator put in
// .env, which is the only way to configure some deployments.
func TestEnvTreatsAnEmptyVariableAsUnset(t *testing.T) {
	writeDotEnv(t, "TAIRITSU_TEST_VALUE=from-file\n")
	t.Setenv("TAIRITSU_TEST_VALUE", "")

	assert.Equal(t, "from-file", Env("TAIRITSU_TEST_VALUE"))
}

func TestDotEnvParsingHandlesQuotesCommentsAndExport(t *testing.T) {
	writeDotEnv(t, "  # comment with = sign\n"+
		"export TAIRITSU_EXPORTED=plain value\n"+
		"TAIRITSU_QUOTED=\"quoted value\"\n"+
		"TAIRITSU_SINGLE='single value'\n"+
		"TAIRITSU_SPACED = spaced \n"+
		"NOT_ASSIGNMENT\n"+
		"= no key\n"+
		"TAIRITSU_EMPTY=\n")

	assert.Equal(t, "plain value", Env("TAIRITSU_EXPORTED"))
	assert.Equal(t, "quoted value", Env("TAIRITSU_QUOTED"))
	assert.Equal(t, "single value", Env("TAIRITSU_SINGLE"))
	assert.Equal(t, "spaced", Env("TAIRITSU_SPACED"))
	assert.Empty(t, Env("NOT_ASSIGNMENT"))
	assert.Empty(t, Env("TAIRITSU_EMPTY"))
}

func TestEnvListSplitsTrimsAndDropsEmptyEntries(t *testing.T) {
	writeDotEnv(t, "")
	t.Setenv("TAIRITSU_TEST_LIST", " 172.17.0.1 , 10.0.0.0/8;172.18.0.1\t172.19.0.1 ,, ")

	assert.Equal(t, []string{"172.17.0.1", "10.0.0.0/8", "172.18.0.1", "172.19.0.1"}, EnvList("TAIRITSU_TEST_LIST"))
	assert.Nil(t, EnvList("TAIRITSU_TEST_ABSENT"))
}

func TestEnvListReturnsNilForEmptyValues(t *testing.T) {
	writeDotEnv(t, "")
	t.Setenv("TAIRITSU_TEST_LIST", " , ; ")

	assert.Nil(t, EnvList("TAIRITSU_TEST_LIST"))
}
