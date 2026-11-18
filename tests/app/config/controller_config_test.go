package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GT-610/tempest-zt/internal/app/config"
	"github.com/GT-610/tempest-zt/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeControllerURLAcceptsHTTPAddresses(t *testing.T) {
	cases := map[string]string{
		"http://127.0.0.1:9993":        "http://127.0.0.1:9993",
		"https://controller.internal":  "https://controller.internal",
		"http://zerotier:9993":         "http://zerotier:9993",
		"  http://127.0.0.1:9993  ":    "http://127.0.0.1:9993",
		"http://127.0.0.1:9993/":       "http://127.0.0.1:9993",
		"https://example.test/zt/":     "https://example.test/zt",
		"http://[::1]:9993":            "http://[::1]:9993",
		"https://203.0.113.10:8443/zt": "https://203.0.113.10:8443/zt",
	}

	for input, expected := range cases {
		actual, err := config.NormalizeControllerURL(input)
		require.NoError(t, err, "input %q", input)
		assert.Equal(t, expected, actual, "input %q", input)
	}
}

// The client sends the controller token in a header to whatever this names, so a
// scheme the transport does not understand widens who can receive it.
func TestNormalizeControllerURLRejectsUnsafeAddresses(t *testing.T) {
	cases := map[string]string{
		"":                           "empty",
		"   ":                        "blank",
		"gopher://127.0.0.1:9993":    "gopher scheme",
		"file:///etc/passwd":         "file scheme",
		"ftp://127.0.0.1":            "ftp scheme",
		"javascript:alert(1)":        "javascript scheme",
		"127.0.0.1:9993":             "no scheme",
		"//127.0.0.1:9993":           "scheme-relative",
		"http://":                    "no host",
		"http://user:pass@127.0.0.1": "embedded credentials",

		// The client appends its request path by concatenation, so anything after the
		// base captures the appended endpoint instead of following it.
		"http://127.0.0.1:9993?mode=test": "query string",
		"http://127.0.0.1:9993/?a=b":      "query string after a slash",
		"http://127.0.0.1:9993?":          "bare question mark",
		"http://127.0.0.1:9993#fragment":  "fragment",
		"http://127.0.0.1:9993/zt#top":    "fragment on a path",
	}

	for input, description := range cases {
		_, err := config.NormalizeControllerURL(input)
		assert.Error(t, err, "%s (%q) must be rejected", description, input)
	}
}

// TestLoadTokenFromPathIntoRejectsFilesThatAreNotTokens is the control behind the
// setup surface's token path: the file lives wherever the controller was
// installed, so the path cannot be confined, and what has to hold instead is that
// the file contains a token rather than whatever the caller wanted to read.
func TestLoadTokenFromPathIntoRejectsFilesThatAreNotTokens(t *testing.T) {
	testutil.UseTempDataDir(t)

	notTokens := map[string]string{
		"multi-line passwd-style": "root:x:0:0:root:/root:/bin/sh\ndaemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin\n",
		"json configuration":      `{"security":{"encryption_key":"c2VjcmV0"}}`,
		"pem private key":         "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAA\n-----END OPENSSH PRIVATE KEY-----",
		"too short":               "abc123",
		"with spaces":             "abcdefgh ijklmnop qrstuvwx",
	}

	for name, content := range notTokens {
		dir := t.TempDir()
		path := filepath.Join(dir, "authtoken.secret")
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

		cfg := &config.Config{Security: config.SecurityConfig{EncryptionKey: "test-secret-key"}}
		err := config.LoadTokenFromPathInto(cfg, path)
		assert.Error(t, err, "%s must be rejected as a token", name)
		assert.Empty(t, cfg.ZeroTier.Token, "a rejected file must not be stored")
	}
}

func TestLoadTokenFromPathIntoAcceptsAControllerToken(t *testing.T) {
	testutil.UseTempDataDir(t)

	path := filepath.Join(t.TempDir(), "authtoken.secret")
	// The shape the controller actually writes, including its trailing newline.
	require.NoError(t, os.WriteFile(path, []byte("gvyqdnryjyhm9tms336s428i\n"), 0o600))

	cfg := &config.Config{Security: config.SecurityConfig{EncryptionKey: "test-secret-key"}}
	require.NoError(t, config.LoadTokenFromPathInto(cfg, path))
	assert.NotEmpty(t, cfg.ZeroTier.Token)

	plaintext, err := config.GetZTTokenFrom(cfg)
	require.NoError(t, err)
	assert.Equal(t, "gvyqdnryjyhm9tms336s428i", plaintext)
}

// TestSetZTConfigOnLeavesSettingsUntouchedWhenRejected guards the ordering: the
// token is read before anything is assigned, so a rejected request cannot leave
// the in-memory configuration half updated.
func TestSetZTConfigOnLeavesSettingsUntouchedWhenRejected(t *testing.T) {
	testutil.UseTempDataDir(t)

	badPath := filepath.Join(t.TempDir(), "not-a-token")
	require.NoError(t, os.WriteFile(badPath, []byte("root:x:0:0:root:/root:/bin/sh\n"), 0o600))

	cfg := &config.Config{
		ZeroTier: config.ZeroTierConfig{URL: "http://127.0.0.1:9993", TokenPath: "/var/lib/zerotier-one/authtoken.secret"},
		Security: config.SecurityConfig{EncryptionKey: "test-secret-key"},
	}

	err := config.SetZTConfigOn(cfg, "http://198.51.100.7:9993", badPath)
	require.Error(t, err)

	assert.Equal(t, "http://127.0.0.1:9993", cfg.ZeroTier.URL, "the stored URL must survive a rejected request")
	assert.Equal(t, "/var/lib/zerotier-one/authtoken.secret", cfg.ZeroTier.TokenPath)
}
