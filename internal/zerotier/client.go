package zerotier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/GT-610/tairitsu/internal/app/logger"
)

// Client is a ZeroTier controller API client.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

const responsePreviewLimit = 160

// Network represents a ZeroTier network.
type Network struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Config      NetworkConfig `json:"config"`
	Created     int64         `json:"creationTime"`
	Modified    int64         `json:"lastModifiedTime"`
	Status      string        `json:"status"`
}

// NetworkResponse is the raw flat network structure returned by the ZeroTier API (used for custom unmarshalling).
type NetworkResponse struct {
	ID                         string             `json:"id"`
	Name                       string             `json:"name"`
	Description                string             `json:"description"`
	Private                    bool               `json:"private"`
	AllowPassivePortForwarding bool               `json:"allowPassivePortForwarding"`
	EnableBroadcast            bool               `json:"enableBroadcast"`
	Mtu                        int                `json:"mtu"`
	MulticastLimit             int                `json:"multicastLimit"`
	IpAssignmentPools          []IpAssignmentPool `json:"ipAssignmentPools"`
	Routes                     []Route            `json:"routes"`
	Tags                       []Tag              `json:"tags"`
	Rules                      []Rule             `json:"rules"`
	DNS                        DNSConfig          `json:"dns"`
	V4AssignMode               AssignmentMode     `json:"v4AssignMode"`
	V6AssignMode               V6AssignmentMode   `json:"v6AssignMode"`
	CreationTime               int64              `json:"creationTime"`
	LastModifiedTime           int64              `json:"lastModifiedTime"`
	Status                     string             `json:"status"`
}

func (n *Network) UnmarshalJSON(data []byte) error {
	var resp NetworkResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal network JSON: %w", err)
	}

	n.ID = resp.ID
	n.Name = resp.Name
	n.Description = resp.Description
	n.Created = resp.CreationTime
	n.Modified = resp.LastModifiedTime
	n.Status = resp.Status

	n.Config = NetworkConfig{
		Private:                    resp.Private,
		AllowPassivePortForwarding: resp.AllowPassivePortForwarding,
		EnableBroadcast:            resp.EnableBroadcast,
		Mtu:                        resp.Mtu,
		MulticastLimit:             resp.MulticastLimit,
		IpAssignmentPools:          resp.IpAssignmentPools,
		Routes:                     resp.Routes,
		Tags:                       resp.Tags,
		Rules:                      resp.Rules,
		DNS:                        resp.DNS,
		V4AssignMode:               resp.V4AssignMode,
		V6AssignMode:               resp.V6AssignMode,
	}

	return nil
}

// NetworkUpdateRequest is a partial network update request (no required fields).
type NetworkUpdateRequest struct {
	Name                 string             `json:"name,omitempty"`
	Description          string             `json:"description,omitempty"`
	Private              bool               `json:"private"`
	AllowPassiveBridging *bool              `json:"allowPassiveBridging,omitempty"`
	EnableBroadcast      bool               `json:"enableBroadcast"`
	Mtu                  *int               `json:"mtu,omitempty"`
	MulticastLimit       *int               `json:"multicastLimit,omitempty"`
	IpAssignmentPools    []IpAssignmentPool `json:"ipAssignmentPools,omitempty"`
	Routes               []Route            `json:"routes,omitempty"`
	DNS                  *DNSConfig         `json:"dns,omitempty"`
	V4AssignMode         *AssignmentMode    `json:"v4AssignMode,omitempty"`
	V6AssignMode         *V6AssignmentMode  `json:"v6AssignMode,omitempty"`
}

// NetworkConfig holds the network configuration fields.
type NetworkConfig struct {
	Private                    bool               `json:"private"`
	AllowPassivePortForwarding bool               `json:"allowPassivePortForwarding"`
	EnableBroadcast            bool               `json:"enableBroadcast"`
	Mtu                        int                `json:"mtu"`
	MulticastLimit             int                `json:"multicastLimit"`
	IpAssignmentPools          []IpAssignmentPool `json:"ipAssignmentPools"`
	Routes                     []Route            `json:"routes"`
	Tags                       []Tag              `json:"tags"`
	Rules                      []Rule             `json:"rules"`
	DNS                        DNSConfig          `json:"dns"`
	V4AssignMode               AssignmentMode     `json:"v4AssignMode"`
	V6AssignMode               V6AssignmentMode   `json:"v6AssignMode"`
}

type DNSConfig struct {
	Domain  string   `json:"domain"`
	Servers []string `json:"servers"`
}

func (d *DNSConfig) UnmarshalJSON(data []byte) error {
	trimmed := string(bytes.TrimSpace(data))
	if trimmed == "" || trimmed == "null" {
		*d = DNSConfig{}
		return nil
	}

	if trimmed == "[]" {
		*d = DNSConfig{}
		return nil
	}

	type dnsAlias DNSConfig
	var alias dnsAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return fmt.Errorf("failed to unmarshal DNS config: %w", err)
	}

	*d = DNSConfig(alias)
	return nil
}

// V6AssignmentMode represents IPv6 assignment mode.
type V6AssignmentMode struct {
	ZT      bool `json:"zt"`
	Plane6  bool `json:"6plane"`
	Rfc4193 bool `json:"rfc4193"`
}

// IpAssignmentPool defines an IP address range for assignment.
type IpAssignmentPool struct {
	IpRangeStart string `json:"ipRangeStart"`
	IpRangeEnd   string `json:"ipRangeEnd"`
}

// Route represents a network route.
type Route struct {
	Target string `json:"target"`
	Via    string `json:"via,omitempty"`
}

// Tag represents a network tag.
type Tag struct {
	ID    int `json:"id"`
	Value int `json:"value"`
}

// Rule represents a network rule.
type Rule struct {
	Not       bool   `json:"not"`
	Or        bool   `json:"or,omitempty"`
	Type      string `json:"type"`
	Metric    int    `json:"metric"`
	PortMin   int    `json:"portMin,omitempty"`
	PortMax   int    `json:"portMax,omitempty"`
	EthType   int    `json:"etherType,omitempty"`
	IPVersion int    `json:"ipVersion,omitempty"`
	Action    string `json:"action"`
}

// AssignmentMode represents IPv4 assignment mode.
type AssignmentMode struct {
	ZT bool `json:"zt"`
}

// Member represents a network member.
type Member struct {
	ID              string       `json:"id"`
	Address         string       `json:"address"`
	Config          MemberConfig `json:"config"`
	Authorized      bool         `json:"authorized,omitempty"`
	ActiveBridge    bool         `json:"activeBridge,omitempty"`
	IPAssignments   []string     `json:"ipAssignments,omitempty"`
	Tags            []Tag        `json:"tags,omitempty"`
	Capabilities    []int        `json:"capabilities,omitempty"`
	NoAutoAssignIPs bool         `json:"noAutoAssignIps,omitempty"`
	Identity        string       `json:"identity"`
	Name            string       `json:"name"`
	Description     string       `json:"description"`
	ClientVersion   string       `json:"clientVersion,omitempty"`
	Online          bool         `json:"online,omitempty"`
	LastSeen        int64        `json:"lastOnline,omitempty"`
	CreationTime    int64        `json:"creationTime,omitempty"`
	VMajor          int          `json:"vMajor,omitempty"`
	VMinor          int          `json:"vMinor,omitempty"`
	VRev            int          `json:"vRev,omitempty"`
	PeerVersion     string       `json:"peerVersion,omitempty"`
	PeerLatency     int          `json:"peerLatency,omitempty"`
	PeerRole        string       `json:"peerRole,omitempty"`
	PreferredPath   string       `json:"preferredPath,omitempty"`
}

type memberAlias struct {
	ID              string       `json:"id"`
	Address         string       `json:"address"`
	Config          MemberConfig `json:"config"`
	Authorized      *bool        `json:"authorized"`
	ActiveBridge    *bool        `json:"activeBridge"`
	IPAssignments   []string     `json:"ipAssignments"`
	Tags            []Tag        `json:"tags"`
	Capabilities    []int        `json:"capabilities"`
	NoAutoAssignIPs *bool        `json:"noAutoAssignIps"`
	Identity        string       `json:"identity"`
	Name            string       `json:"name"`
	Description     string       `json:"description"`
	ClientVersion   string       `json:"clientVersion"`
	Online          bool         `json:"online"`
	LastSeen        int64        `json:"lastOnline"`
	CreationTime    int64        `json:"creationTime"`
	VMajor          int          `json:"vMajor"`
	VMinor          int          `json:"vMinor"`
	VRev            int          `json:"vRev"`
}

func (m *Member) UnmarshalJSON(data []byte) error {
	var raw memberAlias
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to unmarshal member JSON: %w", err)
	}

	m.ID = raw.ID
	m.Address = raw.Address
	m.Identity = raw.Identity
	m.Name = raw.Name
	m.Description = raw.Description
	m.Online = raw.Online
	m.LastSeen = raw.LastSeen
	m.CreationTime = raw.CreationTime
	m.VMajor = raw.VMajor
	m.VMinor = raw.VMinor
	m.VRev = raw.VRev

	m.Config = raw.Config

	if raw.Authorized != nil {
		m.Authorized = *raw.Authorized
		m.Config.Authorized = *raw.Authorized
	} else {
		m.Authorized = raw.Config.Authorized
	}

	if raw.ActiveBridge != nil {
		m.ActiveBridge = *raw.ActiveBridge
		m.Config.ActiveBridge = *raw.ActiveBridge
	} else {
		m.ActiveBridge = raw.Config.ActiveBridge
	}

	if raw.IPAssignments != nil {
		m.IPAssignments = raw.IPAssignments
		m.Config.IPAssignments = raw.IPAssignments
	} else {
		m.IPAssignments = raw.Config.IPAssignments
	}

	if raw.Tags != nil {
		m.Tags = raw.Tags
		m.Config.Tags = raw.Tags
	} else {
		m.Tags = raw.Config.Tags
	}

	if raw.Capabilities != nil {
		m.Capabilities = raw.Capabilities
		m.Config.Capabilities = raw.Capabilities
	} else {
		m.Capabilities = raw.Config.Capabilities
	}

	if raw.NoAutoAssignIPs != nil {
		m.NoAutoAssignIPs = *raw.NoAutoAssignIPs
		m.Config.NoAutoAssignIPs = *raw.NoAutoAssignIPs
	} else {
		m.NoAutoAssignIPs = raw.Config.NoAutoAssignIPs
	}

	if raw.ClientVersion != "" {
		m.ClientVersion = raw.ClientVersion
	} else if raw.VMajor > 0 || raw.VMinor > 0 || raw.VRev > 0 {
		m.ClientVersion = fmt.Sprintf("%d.%d.%d", raw.VMajor, raw.VMinor, raw.VRev)
	}

	return nil
}

type MemberUpdateRequest struct {
	Name            string   `json:"name,omitempty"`
	Authorized      *bool    `json:"authorized,omitempty"`
	ActiveBridge    *bool    `json:"activeBridge,omitempty"`
	IPAssignments   []string `json:"ipAssignments,omitempty"`
	NoAutoAssignIPs *bool    `json:"noAutoAssignIps,omitempty"`
}

// MemberConfig holds member configuration fields.
type MemberConfig struct {
	Authorized      bool     `json:"authorized"`
	ActiveBridge    bool     `json:"activeBridge"`
	IPAssignments   []string `json:"ipAssignments"`
	Tags            []Tag    `json:"tags"`
	Capabilities    []int    `json:"capabilities"`
	NoAutoAssignIPs bool     `json:"noAutoAssignIps"`
}

type Peer struct {
	Address       string     `json:"address"`
	Latency       int        `json:"latency"`
	Paths         []PeerPath `json:"paths"`
	PreferredPath PeerPath   `json:"preferredPath"`
	Role          string     `json:"role"`
	Version       string     `json:"version"`
	VersionMajor  int        `json:"versionMajor"`
	VersionMinor  int        `json:"versionMinor"`
	VersionRev    int        `json:"versionRev"`
}

type PeerPath struct {
	Active    bool   `json:"active"`
	Address   string `json:"address"`
	Preferred bool   `json:"preferred"`
}

// Status represents the ZeroTier controller status.
type Status struct {
	Version     string `json:"version"`
	Address     string `json:"address"`
	Online      bool   `json:"online"`
	TCPFallback bool   `json:"tcpFallbackAvailable"`
	APIReady    bool   `json:"apiReady"`
}

type statusAlias struct {
	Version              string `json:"version"`
	Address              string `json:"address"`
	Online               bool   `json:"online"`
	TCPFallbackAvailable bool   `json:"tcpFallbackAvailable"`
	TCPFallbackActive    bool   `json:"tcpFallbackActive"`
	APIReady             bool   `json:"apiReady"`
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var raw statusAlias
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to unmarshal status JSON: %w", err)
	}

	s.Version = raw.Version
	s.Address = raw.Address
	s.Online = raw.Online
	s.TCPFallback = raw.TCPFallbackAvailable || raw.TCPFallbackActive
	s.APIReady = raw.APIReady

	return nil
}

func NewClientWithConfig(cfg *config.Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration not loaded")
	}

	if err := config.LoadTokenFromPathInto(cfg, cfg.ZeroTier.TokenPath); err != nil && strings.TrimSpace(cfg.ZeroTier.Token) == "" {
		return nil, fmt.Errorf("failed to load ZeroTier token: %w", err)
	}

	token, err := config.GetZTTokenFrom(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get ZeroTier token: %w", err)
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        20,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	baseURL := cfg.ZeroTier.URL
	if baseURL == "" {
		baseURL = "http://localhost:9993"
		logger.Warn("ZeroTier URL not configured, falling back to default; this will not work if ZeroTier runs in a separate container")
	}

	return &Client{
		BaseURL:    baseURL,
		Token:      token,
		HTTPClient: httpClient,
	}, nil
}

// doRequest executes an HTTP request against the ZeroTier controller.
func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)

	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ZT1-Auth", c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("request failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetStatus retrieves the ZeroTier controller status.
func (c *Client) GetStatus() (*Status, error) {
	respBody, err := c.doRequest("GET", "/status", nil)
	if err != nil {
		return nil, err
	}

	var status Status
	if err := json.Unmarshal(respBody, &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status response: %w; preview: %s", err, responsePreview(respBody))
	}

	return &status, nil
}

// GetNetworkIDs retrieves only the network ID list (lightweight).
func (c *Client) GetNetworkIDs() ([]string, error) {
	respBody, err := c.doRequest("GET", "/controller/network", nil)
	if err != nil {
		return nil, err
	}

	networkIDs, err := parseNetworkIDs(respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse network ID list: %w; preview: %s", err, responsePreview(respBody))
	}

	return networkIDs, nil
}

func parseNetworkIDs(data []byte) ([]string, error) {
	trimmed := string(bytes.TrimSpace(data))
	if trimmed == "" || trimmed == "null" {
		return []string{}, nil
	}

	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	switch value := raw.(type) {
	case []any:
		return parseNetworkIDsFromArray(value), nil
	case map[string]any:
		return parseNetworkIDsFromObject(value), nil
	default:
		return nil, fmt.Errorf("unsupported network list format")
	}
}

func parseNetworkIDsFromArray(items []any) []string {
	networkIDs := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))

	for _, item := range items {
		switch value := item.(type) {
		case string:
			appendUniqueNetworkID(&networkIDs, seen, value)
		case map[string]any:
			appendUniqueNetworkID(&networkIDs, seen, extractNetworkIDFromObject(value))
		}
	}

	sort.Strings(networkIDs)
	return networkIDs
}

func parseNetworkIDsFromObject(items map[string]any) []string {
	networkIDs := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))

	if nested, ok := items["networks"]; ok {
		switch value := nested.(type) {
		case []any:
			return parseNetworkIDsFromArray(value)
		case map[string]any:
			return parseNetworkIDsFromObject(value)
		}
	}

	for key, value := range items {
		if looksLikeNetworkID(key) {
			appendUniqueNetworkID(&networkIDs, seen, key)
			continue
		}

		if object, ok := value.(map[string]any); ok {
			appendUniqueNetworkID(&networkIDs, seen, extractNetworkIDFromObject(object))
		}
	}

	sort.Strings(networkIDs)
	return networkIDs
}

func extractNetworkIDFromObject(item map[string]any) string {
	id, _ := item["id"].(string)
	if looksLikeNetworkID(id) {
		return id
	}
	return ""
}

func appendUniqueNetworkID(networkIDs *[]string, seen map[string]struct{}, id string) {
	if !looksLikeNetworkID(id) {
		return
	}
	if _, ok := seen[id]; ok {
		return
	}
	seen[id] = struct{}{}
	*networkIDs = append(*networkIDs, id)
}

func looksLikeNetworkID(id string) bool {
	if len(id) != 16 {
		return false
	}

	for _, r := range id {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}

	return true
}

// GetNetwork retrieves a single network by ID.
func (c *Client) GetNetwork(networkID string) (*Network, error) {
	endpoint := fmt.Sprintf("/controller/network/%s", networkID)
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var network Network
	if err := json.Unmarshal(respBody, &network); err != nil {
		return nil, fmt.Errorf("failed to unmarshal network detail: %w; preview: %s", err, responsePreview(respBody))
	}

	return &network, nil
}

// CreateNetwork creates a new network.
func (c *Client) CreateNetwork(network *Network) (*Network, error) {
	respBody, err := c.doRequest("POST", "/controller/network", network)
	if err != nil {
		return nil, err
	}

	var createdNetwork Network
	if err := json.Unmarshal(respBody, &createdNetwork); err != nil {
		return nil, fmt.Errorf("failed to unmarshal create network response: %w; preview: %s", err, responsePreview(respBody))
	}

	return &createdNetwork, nil
}

// PartialUpdateNetwork partially updates a network configuration.
func (c *Client) PartialUpdateNetwork(networkID string, updateReq *NetworkUpdateRequest) (*Network, error) {
	endpoint := fmt.Sprintf("/controller/network/%s", networkID)
	respBody, err := c.doRequest("POST", endpoint, updateReq)
	if err != nil {
		return nil, err
	}

	var updatedNetwork Network
	if err := json.Unmarshal(respBody, &updatedNetwork); err != nil {
		return nil, fmt.Errorf("failed to unmarshal update network response: %w; preview: %s", err, responsePreview(respBody))
	}

	return &updatedNetwork, nil
}

// DeleteNetwork deletes a network by ID.
func (c *Client) DeleteNetwork(networkID string) error {
	endpoint := fmt.Sprintf("/controller/network/%s", networkID)
	_, err := c.doRequest("DELETE", endpoint, nil)
	return err
}

// GetMembers retrieves all members of a network.
func (c *Client) GetMembers(networkID string) ([]Member, error) {
	endpoint := fmt.Sprintf("/controller/network/%s/member", networkID)
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	members, err := parseMemberList(respBody)
	if err == nil {
		return members, nil
	}

	memberIDs, indexErr := parseMemberIndexList(respBody)
	if indexErr != nil {
		return nil, err
	}

	members = make([]Member, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		member, detailErr := c.GetMember(networkID, memberID)
		if detailErr != nil {
			return nil, fmt.Errorf("failed to get member %s detail: %w", memberID, detailErr)
		}
		if member != nil {
			members = append(members, *member)
		}
	}

	sort.Slice(members, func(i, j int) bool {
		return members[i].ID < members[j].ID
	})

	return members, nil
}

// GetMember retrieves a single member by network ID and member ID.
func (c *Client) GetMember(networkID, memberID string) (*Member, error) {
	endpoint := fmt.Sprintf("/controller/network/%s/member/%s", networkID, memberID)
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var member Member
	if err := json.Unmarshal(respBody, &member); err != nil {
		return nil, fmt.Errorf("failed to unmarshal member detail: %w; preview: %s", err, responsePreview(respBody))
	}

	return &member, nil
}

// GetPeers retrieves all peer nodes.
func (c *Client) GetPeers() ([]Peer, error) {
	respBody, err := c.doRequest("GET", "/peer", nil)
	if err != nil {
		return nil, err
	}

	peers, err := parsePeerList(respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal peer list: %w; preview: %s", err, responsePreview(respBody))
	}

	sort.Slice(peers, func(i, j int) bool {
		return peers[i].Address < peers[j].Address
	})

	return peers, nil
}

// UpdateMember updates a member's configuration.
func (c *Client) UpdateMember(networkID, memberID string, member *MemberUpdateRequest) (*Member, error) {
	endpoint := fmt.Sprintf("/controller/network/%s/member/%s", networkID, memberID)
	respBody, err := c.doRequest("POST", endpoint, member)
	if err != nil {
		return nil, err
	}

	var updatedMember Member
	if err := json.Unmarshal(respBody, &updatedMember); err != nil {
		return nil, fmt.Errorf("failed to unmarshal update member response: %w; preview: %s", err, responsePreview(respBody))
	}

	return &updatedMember, nil
}

// DeleteMember removes a member from a network.
func (c *Client) DeleteMember(networkID, memberID string) error {
	endpoint := fmt.Sprintf("/controller/network/%s/member/%s", networkID, memberID)
	_, err := c.doRequest("DELETE", endpoint, nil)
	return err
}

func parseMemberList(respBody []byte) ([]Member, error) {
	var members []Member
	if err := json.Unmarshal(respBody, &members); err == nil {
		return members, nil
	}

	var memberMap map[string]Member
	if err := json.Unmarshal(respBody, &memberMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal member list: %w; preview: %s", err, responsePreview(respBody))
	}

	members = make([]Member, 0, len(memberMap))
	for memberID, member := range memberMap {
		if member.ID == "" {
			member.ID = memberID
		}
		members = append(members, member)
	}

	sort.Slice(members, func(i, j int) bool {
		return members[i].ID < members[j].ID
	})

	return members, nil
}

func parseMemberIndexList(respBody []byte) ([]string, error) {
	var memberIndex map[string]int
	if err := json.Unmarshal(respBody, &memberIndex); err != nil {
		return nil, fmt.Errorf("failed to parse member index: %w; preview: %s", err, responsePreview(respBody))
	}

	memberIDs := make([]string, 0, len(memberIndex))
	for memberID := range memberIndex {
		memberIDs = append(memberIDs, memberID)
	}
	sort.Strings(memberIDs)

	return memberIDs, nil
}

func parsePeerList(respBody []byte) ([]Peer, error) {
	var peers []Peer
	if err := json.Unmarshal(respBody, &peers); err == nil {
		return peers, nil
	}

	var peerMap map[string]Peer
	if err := json.Unmarshal(respBody, &peerMap); err != nil {
		return nil, err
	}

	peers = make([]Peer, 0, len(peerMap))
	for address, peer := range peerMap {
		if peer.Address == "" {
			peer.Address = address
		}
		peers = append(peers, peer)
	}

	return peers, nil
}

func responsePreview(respBody []byte) string {
	preview := bytes.TrimSpace(respBody)
	if len(preview) == 0 {
		return "<empty>"
	}
	if len(preview) > responsePreviewLimit {
		return string(preview[:responsePreviewLimit]) + "..."
	}
	return string(preview)
}
