package zerotier

import (
	"encoding/json"
	"testing"
)

func TestDNSConfigUnmarshalJSONSupportsArrayNullAndObject(t *testing.T) {
	testCases := []struct {
		name string
		data string
		want DNSConfig
	}{
		{name: "empty array", data: `[]`, want: DNSConfig{}},
		{name: "null", data: `null`, want: DNSConfig{}},
		{name: "object", data: `{"domain":"home.arpa","servers":["1.1.1.1"]}`, want: DNSConfig{
			Domain:  "home.arpa",
			Servers: []string{"1.1.1.1"},
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got DNSConfig
			if err := json.Unmarshal([]byte(tc.data), &got); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			if got.Domain != tc.want.Domain {
				t.Fatalf("domain = %q, want %q", got.Domain, tc.want.Domain)
			}
			if len(got.Servers) != len(tc.want.Servers) {
				t.Fatalf("servers len = %d, want %d", len(got.Servers), len(tc.want.Servers))
			}
			for i := range got.Servers {
				if got.Servers[i] != tc.want.Servers[i] {
					t.Fatalf("server[%d] = %q, want %q", i, got.Servers[i], tc.want.Servers[i])
				}
			}
		})
	}
}

func TestMemberUnmarshalJSONDoesNotFallbackToZeroVersion(t *testing.T) {
	var member Member
	if err := json.Unmarshal([]byte(`{"id":"member-1","config":{"authorized":false}}`), &member); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if member.ClientVersion != "" {
		t.Fatalf("clientVersion = %q, want empty string", member.ClientVersion)
	}
}

func TestMemberUnmarshalJSONUsesVersionFallbackWhenAvailable(t *testing.T) {
	var member Member
	if err := json.Unmarshal([]byte(`{"id":"member-1","config":{"authorized":false},"vMajor":1,"vMinor":14,"vRev":2}`), &member); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if member.ClientVersion != "1.14.2" {
		t.Fatalf("clientVersion = %q, want %q", member.ClientVersion, "1.14.2")
	}
}
