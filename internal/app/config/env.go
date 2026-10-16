package config

import (
	"bufio"
	"os"
	"strings"
	"sync"
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

// EnvList returns the entries of a comma, semicolon or whitespace separated
// environment variable. Empty entries are dropped.
func EnvList(name string) []string {
	raw := Env(name)
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
