package zerotier

import (
	"encoding/json"
	"testing"
)

func TestParseNetworkIDsSupportsCommonResponseShapes(t *testing.T) {
	testCases := []struct {
		name string
		data string
		want []string
	}{
		{
			name: "string array",
			data: `["f76fd3000b86b177","8056c2e21c000001"]`,
			want: []string{"8056c2e21c000001", "f76fd3000b86b177"},
		},
		{
			name: "object keyed by network id",
			data: `{
				"f76fd3000b86b177":{"id":"f76fd3000b86b177","name":"main-net"},
				"8056c2e21c000001":{"id":"8056c2e21c000001","name":"lab-net"}
			}`,
			want: []string{"8056c2e21c000001", "f76fd3000b86b177"},
		},
		{
			name: "object array",
			data: `[
				{"id":"f76fd3000b86b177","name":"main-net"},
				{"id":"8056c2e21c000001","name":"lab-net"}
			]`,
			want: []string{"8056c2e21c000001", "f76fd3000b86b177"},
		},
		{
			name: "nested networks field",
			data: `{
				"networks":{
					"f76fd3000b86b177":{"name":"main-net"},
					"8056c2e21c000001":{"name":"lab-net"}
				},
				"total":2
			}`,
			want: []string{"8056c2e21c000001", "f76fd3000b86b177"},
		},
		{
			name: "empty null response",
			data: `null`,
			want: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseNetworkIDs([]byte(tc.data))
			if err != nil {
				t.Fatalf("parseNetworkIDs() error = %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("len(got) = %d, want %d; got=%v", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("got[%d] = %q, want %q; got=%v", i, got[i], tc.want[i], got)
				}
			}
		})
	}
}

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
