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
	t.Setenv("TEMPEST_ZT_TEST_VALUE", "from-environment")

	assert.Equal(t, "from-environment", Env("TEMPEST_ZT_TEST_VALUE"))
	assert.Empty(t, Env("TEMPEST_ZT_TEST_ABSENT"))
}

func TestEnvReadsTheDotEnvFile(t *testing.T) {
	writeDotEnv(t, "# a comment\n\nTEMPEST_ZT_TEST_VALUE=from-file\n")

	assert.Equal(t, "from-file", Env("TEMPEST_ZT_TEST_VALUE"))
}

func TestEnvPrefersTheProcessEnvironmentOverTheFile(t *testing.T) {
	writeDotEnv(t, "TEMPEST_ZT_TEST_VALUE=from-file\n")
	t.Setenv("TEMPEST_ZT_TEST_VALUE", "from-environment")

	assert.Equal(t, "from-environment", Env("TEMPEST_ZT_TEST_VALUE"))
}

// `FOO=` in a container definition must not shadow a value the operator put in
// .env, which is the only way to configure some deployments.
func TestEnvTreatsAnEmptyVariableAsUnset(t *testing.T) {
	writeDotEnv(t, "TEMPEST_ZT_TEST_VALUE=from-file\n")
	t.Setenv("TEMPEST_ZT_TEST_VALUE", "")

	assert.Equal(t, "from-file", Env("TEMPEST_ZT_TEST_VALUE"))
}

func TestDotEnvParsingHandlesQuotesCommentsAndExport(t *testing.T) {
	writeDotEnv(t, "  # comment with = sign\n"+
		"export TEMPEST_ZT_EXPORTED=plain value\n"+
		"TEMPEST_ZT_QUOTED=\"quoted value\"\n"+
		"TEMPEST_ZT_SINGLE='single value'\n"+
		"TEMPEST_ZT_SPACED = spaced \n"+
		"NOT_ASSIGNMENT\n"+
		"= no key\n"+
		"TEMPEST_ZT_EMPTY=\n")

	assert.Equal(t, "plain value", Env("TEMPEST_ZT_EXPORTED"))
	assert.Equal(t, "quoted value", Env("TEMPEST_ZT_QUOTED"))
	assert.Equal(t, "single value", Env("TEMPEST_ZT_SINGLE"))
	assert.Equal(t, "spaced", Env("TEMPEST_ZT_SPACED"))
	assert.Empty(t, Env("NOT_ASSIGNMENT"))
	assert.Empty(t, Env("TEMPEST_ZT_EMPTY"))
}

func TestEnvListSplitsTrimsAndDropsEmptyEntries(t *testing.T) {
	writeDotEnv(t, "")
	t.Setenv("TEMPEST_ZT_TEST_LIST", " 172.17.0.1 , 10.0.0.0/8;172.18.0.1\t172.19.0.1 ,, ")

	assert.Equal(t, []string{"172.17.0.1", "10.0.0.0/8", "172.18.0.1", "172.19.0.1"}, EnvList("TEMPEST_ZT_TEST_LIST"))
	assert.Nil(t, EnvList("TEMPEST_ZT_TEST_ABSENT"))
}

func TestEnvListReturnsNilForEmptyValues(t *testing.T) {
	writeDotEnv(t, "")
	t.Setenv("TEMPEST_ZT_TEST_LIST", " , ; ")

	assert.Nil(t, EnvList("TEMPEST_ZT_TEST_LIST"))
}

func TestEnvWithLegacyPrefersTheCurrentName(t *testing.T) {
	writeDotEnv(t, "")
	t.Setenv("TEMPEST_ZT_TEST_SETTING", "current")
	t.Setenv(legacyEnvPrefix+"TEST_SETTING", "legacy")

	assert.Equal(t, "current", EnvWithLegacy("TEMPEST_ZT_TEST_SETTING"))
}

func TestEnvWithLegacyFallsBackToTheOldName(t *testing.T) {
	writeDotEnv(t, "")
	t.Setenv(legacyEnvPrefix+"TEST_SETTING", "legacy")

	assert.Equal(t, "legacy", EnvWithLegacy("TEMPEST_ZT_TEST_SETTING"))
}

func TestEnvWithLegacyLeavesForeignNamesAlone(t *testing.T) {
	writeDotEnv(t, "")

	// ZT_* and the retired JWT_SECRET are not ours to redirect, so an absent
	// value stays absent instead of resolving to some unrelated old name.
	assert.Empty(t, EnvWithLegacy("ZT_CONTROLLER_URL"))
	assert.Empty(t, EnvWithLegacy("JWT_SECRET"))
}

func TestEnvListWithLegacyFallsBackToTheOldName(t *testing.T) {
	writeDotEnv(t, "")
	t.Setenv(legacyEnvPrefix+"TEST_LIST", "172.17.0.1, 10.0.0.0/8")

	assert.Equal(t, []string{"172.17.0.1", "10.0.0.0/8"}, EnvListWithLegacy("TEMPEST_ZT_TEST_LIST"))
	assert.Nil(t, EnvListWithLegacy("TEMPEST_ZT_TEST_ABSENT"))
}

func TestLegacyFallbackIsReportedOncePerVariable(t *testing.T) {
	writeDotEnv(t, "")
	t.Setenv(legacyEnvPrefix+"TEST_REPEATED", "legacy")
	t.Cleanup(func() {
		legacyEnvReportMux.Lock()
		delete(legacyEnvReported, legacyEnvPrefix+"TEST_REPEATED")
		legacyEnvReportMux.Unlock()
	})

	EnvWithLegacy("TEMPEST_ZT_TEST_REPEATED")

	legacyEnvReportMux.Lock()
	_, reported := legacyEnvReported[legacyEnvPrefix+"TEST_REPEATED"]
	legacyEnvReportMux.Unlock()

	assert.True(t, reported)
}
