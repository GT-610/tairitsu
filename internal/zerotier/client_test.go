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

// The controller answers a network in one flat object, and NetworkConfig is the
// nested shape the API publishes. Every configuration field has to survive that
// move: dropping one does not fail to compile, it silently reports the zero
// value, so a network with the setting on is described as having it off.
func TestNetworkUnmarshalJSONKeepsEveryConfigurationField(t *testing.T) {
	// Every value differs from the field's zero value, so a missed copy shows up
	// as a mismatch rather than agreeing by accident.
	payload := `{
		"id": "f76fd3000b86b177",
		"name": "lab",
		"description": "lab network",
		"private": true,
		"allowPassivePortForwarding": true,
		"enableBroadcast": true,
		"mtu": 2800,
		"multicastLimit": 32,
		"ipAssignmentPools": [{"ipRangeStart": "10.147.17.1", "ipRangeEnd": "10.147.17.254"}],
		"routes": [{"target": "10.0.0.0/8", "via": "10.147.17.1"}],
		"tags": [{"id": 1, "value": 2}],
		"rules": [{"not": true, "type": "ACTION_ACCEPT", "metric": 5}],
		"dns": {"domain": "lab.example", "servers": ["10.147.17.1"]},
		"v4AssignMode": {"zt": true},
		"v6AssignMode": {"zt": true, "6plane": true, "rfc4193": true},
		"creationTime": 1700000000000,
		"lastModifiedTime": 1700000001000,
		"status": "OK"
	}`

	var network Network
	if err := json.Unmarshal([]byte(payload), &network); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if network.ID != "f76fd3000b86b177" || network.Name != "lab" || network.Description != "lab network" {
		t.Fatalf("identity fields not carried over: %+v", network)
	}
	if network.Created != 1700000000000 || network.Modified != 1700000001000 || network.Status != "OK" {
		t.Fatalf("controller metadata not carried over: %+v", network)
	}

	config := network.Config
	if !config.Private {
		t.Error("Private was dropped")
	}
	if !config.AllowPassivePortForwarding {
		t.Error("AllowPassivePortForwarding was dropped")
	}
	if !config.EnableBroadcast {
		t.Error("EnableBroadcast was dropped")
	}
	if config.Mtu != 2800 {
		t.Errorf("Mtu = %d, want 2800", config.Mtu)
	}
	if config.MulticastLimit != 32 {
		t.Errorf("MulticastLimit = %d, want 32", config.MulticastLimit)
	}
	if len(config.IpAssignmentPools) != 1 || config.IpAssignmentPools[0].IpRangeStart != "10.147.17.1" {
		t.Errorf("IpAssignmentPools = %+v", config.IpAssignmentPools)
	}
	if len(config.Routes) != 1 || config.Routes[0].Target != "10.0.0.0/8" {
		t.Errorf("Routes = %+v", config.Routes)
	}
	if len(config.Tags) != 1 || config.Tags[0].ID != 1 || config.Tags[0].Value != 2 {
		t.Errorf("Tags = %+v", config.Tags)
	}
	if len(config.Rules) != 1 || !config.Rules[0].Not || config.Rules[0].Metric != 5 {
		t.Errorf("Rules = %+v", config.Rules)
	}
	if config.DNS.Domain != "lab.example" || len(config.DNS.Servers) != 1 {
		t.Errorf("DNS = %+v", config.DNS)
	}
	if !config.V4AssignMode.ZT {
		t.Error("V4AssignMode was dropped")
	}
	if !config.V6AssignMode.ZT || !config.V6AssignMode.Plane6 || !config.V6AssignMode.Rfc4193 {
		t.Errorf("V6AssignMode = %+v", config.V6AssignMode)
	}
}
