package config

import (
	"bufio"
	"os"
	"strings"
	"sync"

	"github.com/GT-610/tempest-zt/internal/app/logger"
	"go.uber.org/zap"
)

// dotEnvFileName is an optional file read from the working directory. It is a
// convenience for local runs and containers: a real environment variable always
// takes precedence over a value read from the file.
const dotEnvFileName = ".env"

var (
	dotEnvOnce     sync.Once
	dotEnvSettings map[string]string
)

// loadDotEnv reads dotEnvFileName once per process. A missing or unreadable
// file is not an error: the file is optional and every value in it has an
// environment-variable equivalent.
func loadDotEnv() map[string]string {
	dotEnvOnce.Do(func() {
		dotEnvSettings = map[string]string{}

		file, err := os.Open(dotEnvFileName)
		if err != nil {
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, value, found := strings.Cut(line, "=")
			if !found {
				continue
			}
			key = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(key), "export "))
			if key == "" {
				continue
			}
			dotEnvSettings[key] = strings.Trim(strings.TrimSpace(value), `"'`)
		}
	})

	return dotEnvSettings
}

// Env returns the value of an environment variable, falling back to the
// optional .env file in the working directory.
//
// An empty environment variable is treated as unset so that `FOO=` in a
// container definition does not shadow a value configured in .env.
func Env(name string) string {
	if value, ok := os.LookupEnv(name); ok {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}

	return loadDotEnv()[name]
}

// EnvPrefix is the prefix every variable that configures the server shares.
const EnvPrefix = "TEMPEST_ZT_"

// legacyEnvPrefix is the prefix these settings used before the rename. A
// deployment file that still spells its variables the old way keeps working, so
// an upgrade cannot silently drop a setting that has no visible failure: losing
// the trusted-proxy list, for example, only surfaces as rate limits keyed on the
// wrong address.
const legacyEnvPrefix = "TAIRITSU_"

var (
	legacyEnvReported  = map[string]struct{}{}
	legacyEnvReportMux sync.Mutex
)

// legacyEnvName maps a current variable name to the one it replaced. A name that
// is not ours returns "", because the ZeroTier client variables and the
// pre-existing JWT_SECRET have their own handling.
func legacyEnvName(name string) string {
	if !strings.HasPrefix(name, EnvPrefix) {
		return ""
	}
	return legacyEnvPrefix + strings.TrimPrefix(name, EnvPrefix)
}

// EnvWithLegacy reads name, falling back to its pre-rename spelling. The current
// name always wins, and a fallback is reported once per variable rather than on
// every read.
func EnvWithLegacy(name string) string {
	if value := Env(name); value != "" {
		return value
	}

	legacy := legacyEnvName(name)
	if legacy == "" {
		return ""
	}

	value := Env(legacy)
	if value == "" {
		return ""
	}

	reportLegacyEnv(legacy, name)
	return value
}

// EnvListWithLegacy is EnvList over EnvWithLegacy.
func EnvListWithLegacy(name string) []string {
	return splitEnvList(EnvWithLegacy(name))
}

func reportLegacyEnv(legacy, replacement string) {
	legacyEnvReportMux.Lock()
	defer legacyEnvReportMux.Unlock()

	if _, reported := legacyEnvReported[legacy]; reported {
		return
	}
	legacyEnvReported[legacy] = struct{}{}

	logger.Warn("reading a deprecated environment variable; rename it",
		zap.String("deprecated", legacy), zap.String("replacement", replacement))
}

// EnvList returns the entries of a comma, semicolon or whitespace separated
// environment variable. Empty entries are dropped.
func EnvList(name string) []string {
	return splitEnvList(Env(name))
}

func splitEnvList(raw string) []string {
	if raw == "" {
		return nil
	}

	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == ' ' || r == '\t' || r == '\n'
	})

	entries := make([]string, 0, len(fields))
	for _, field := range fields {
		if trimmed := strings.TrimSpace(field); trimmed != "" {
			entries = append(entries, trimmed)
		}
	}

	if len(entries) == 0 {
		return nil
	}
	return entries
}
