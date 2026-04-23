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
)

// Client ZeroTier API客户端
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

const responsePreviewLimit = 160

// Network 网络结构
type Network struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Config      NetworkConfig `json:"config"`
	Created     int64         `json:"creationTime"`
	Modified    int64         `json:"lastModifiedTime"`
	Status      string        `json:"status"`
}

// NetworkResponse ZeroTier API返回的扁平网络结构（用于自定义解析）
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
		return fmt.Errorf("解析网络JSON失败: %w", err)
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

// NetworkUpdateRequest 部分更新网络请求（不强制要求必填字段）
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

// NetworkConfig 网络配置
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
	trimmed := strings.TrimSpace(string(data))
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
		return fmt.Errorf("解析 DNS 配置失败: %w", err)
	}

	*d = DNSConfig(alias)
	return nil
}

// V6AssignmentMode IPv6分配模式
type V6AssignmentMode struct {
	ZT      bool `json:"zt"`
	Plane6  bool `json:"6plane"`
	Rfc4193 bool `json:"rfc4193"`
}

// IpAssignmentPool IP分配池
type IpAssignmentPool struct {
	IpRangeStart string `json:"ipRangeStart"`
	IpRangeEnd   string `json:"ipRangeEnd"`
}

// Route 路由
type Route struct {
	Target string `json:"target"`
	Via    string `json:"via,omitempty"`
}

// Tag 标签
type Tag struct {
	ID    int `json:"id"`
	Value int `json:"value"`
}

// Rule 规则
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

// AssignmentMode 分配模式
type AssignmentMode struct {
	ZT bool `json:"zt"`
}

// Member 网络成员
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
		return fmt.Errorf("解析成员JSON失败: %w", err)
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

// MemberConfig 成员配置
type MemberConfig struct {
	Authorized      bool     `json:"authorized"`
	ActiveBridge    bool     `json:"activeBridge"`
	IPAssignments   []string `json:"ipAssignments"`
	Tags            []Tag    `json:"tags"`
	NATTraversal    bool     `json:"natTraversal"`
	Capabilities    []int    `json:"capabilities"`
	NoAutoAssignIPs bool     `json:"noAutoAssignIps"`
}

// Status ZeroTier状态
type Status struct {
	Version     string `json:"version"`
	Address     string `json:"address"`
	Online      bool   `json:"online"`
	TCPFallback bool   `json:"tcpFallbackAvailable"`
	APIReady    bool   `json:"apiReady"`
}

// NewClient 创建新的ZeroTier客户端
func NewClient() (*Client, error) {
	cfg, err := config.Current()
	if err != nil {
		return nil, err
	}

	return NewClientWithConfig(cfg)
}

func NewClientWithConfig(cfg *config.Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("配置未加载")
	}

	if err := config.LoadTokenFromPathInto(cfg, cfg.ZeroTier.TokenPath); err != nil && strings.TrimSpace(cfg.ZeroTier.Token) == "" {
		return nil, fmt.Errorf("加载ZeroTier令牌失败: %w", err)
	}

	token, err := config.GetZTTokenFrom(cfg)
	if err != nil {
		return nil, fmt.Errorf("获取ZeroTier令牌失败: %w", err)
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	baseURL := cfg.ZeroTier.URL
	if baseURL == "" {
		baseURL = "http://localhost:9993"
	}

	return &Client{
		BaseURL:    baseURL,
		Token:      token,
		HTTPClient: httpClient,
	}, nil
}

// doRequest 执行HTTP请求
func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	// 构建URL
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)

	// 创建请求体
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	// 创建请求
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ZT1-Auth", c.Token)

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("请求失败 (状态码: %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetStatus 获取ZeroTier控制器状态
func (c *Client) GetStatus() (*Status, error) {
	respBody, err := c.doRequest("GET", "/status", nil)
	if err != nil {
		return nil, err
	}

	var status Status
	if err := json.Unmarshal(respBody, &status); err != nil {
		return nil, fmt.Errorf("解析状态响应失败: %w; 响应预览: %s", err, responsePreview(respBody))
	}

	return &status, nil
}

// GetNetworks 获取所有网络列表
func (c *Client) GetNetworks() ([]Network, error) {
	networkIDs, err := c.GetNetworkIDs()
	if err != nil {
		return nil, err
	}

	networks := make([]Network, 0)
	for _, id := range networkIDs {
		network, err := c.GetNetwork(id)
		if err != nil {
			return nil, fmt.Errorf("获取网络 %s 详情失败: %w", id, err)
		}
		if network != nil {
			networks = append(networks, *network)
		}
	}

	return networks, nil
}

// GetNetworkIDs 只获取网络ID列表（轻量级）
func (c *Client) GetNetworkIDs() ([]string, error) {
	respBody, err := c.doRequest("GET", "/controller/network", nil)
	if err != nil {
		return nil, err
	}

	networkIDs, err := parseNetworkIDs(respBody)
	if err != nil {
		return nil, fmt.Errorf("解析网络ID列表失败: %w; 响应预览: %s", err, responsePreview(respBody))
	}

	return networkIDs, nil
}

func parseNetworkIDs(data []byte) ([]string, error) {
	trimmed := strings.TrimSpace(string(data))
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
		return nil, fmt.Errorf("不支持的网络列表格式")
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

// GetNetworkStatus 获取网络状态
func (c *Client) GetNetworkStatus(networkID string) (string, error) {
	endpoint := fmt.Sprintf("/controller/network/%s/status", networkID)
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	// 解析响应获取状态
	var status struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &status); err != nil {
		return "", fmt.Errorf("解析网络状态失败: %w; 响应预览: %s", err, responsePreview(respBody))
	}

	return status.Status, nil
}

// GetNetwork 获取单个网络详情
func (c *Client) GetNetwork(networkID string) (*Network, error) {
	endpoint := fmt.Sprintf("/controller/network/%s", networkID)
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var network Network
	if err := json.Unmarshal(respBody, &network); err != nil {
		return nil, fmt.Errorf("解析网络详情失败: %w; 响应预览: %s", err, responsePreview(respBody))
	}

	return &network, nil
}

// CreateNetwork 创建新网络
func (c *Client) CreateNetwork(network *Network) (*Network, error) {
	respBody, err := c.doRequest("POST", "/controller/network", network)
	if err != nil {
		return nil, err
	}

	var createdNetwork Network
	if err := json.Unmarshal(respBody, &createdNetwork); err != nil {
		return nil, fmt.Errorf("解析创建网络响应失败: %w; 响应预览: %s", err, responsePreview(respBody))
	}

	return &createdNetwork, nil
}

// UpdateNetwork 更新网络配置
func (c *Client) UpdateNetwork(networkID string, network *Network) (*Network, error) {
	endpoint := fmt.Sprintf("/controller/network/%s", networkID)
	respBody, err := c.doRequest("POST", endpoint, network)
	if err != nil {
		return nil, err
	}

	var updatedNetwork Network
	if err := json.Unmarshal(respBody, &updatedNetwork); err != nil {
		return nil, fmt.Errorf("解析更新网络响应失败: %w; 响应预览: %s", err, responsePreview(respBody))
	}

	return &updatedNetwork, nil
}

// PartialUpdateNetwork 部分更新网络配置
func (c *Client) PartialUpdateNetwork(networkID string, updateReq *NetworkUpdateRequest) (*Network, error) {
	endpoint := fmt.Sprintf("/controller/network/%s", networkID)
	respBody, err := c.doRequest("POST", endpoint, updateReq)
	if err != nil {
		return nil, err
	}

	var updatedNetwork Network
	if err := json.Unmarshal(respBody, &updatedNetwork); err != nil {
		return nil, fmt.Errorf("解析更新网络响应失败: %w; 响应预览: %s", err, responsePreview(respBody))
	}

	return &updatedNetwork, nil
}

// DeleteNetwork 删除网络
func (c *Client) DeleteNetwork(networkID string) error {
	endpoint := fmt.Sprintf("/controller/network/%s", networkID)
	_, err := c.doRequest("DELETE", endpoint, nil)
	return err
}

// GetMembers 获取网络成员列表
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
			return nil, fmt.Errorf("获取成员 %s 详情失败: %w", memberID, detailErr)
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

// GetMember 获取单个成员详情
func (c *Client) GetMember(networkID, memberID string) (*Member, error) {
	endpoint := fmt.Sprintf("/controller/network/%s/member/%s", networkID, memberID)
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var member Member
	if err := json.Unmarshal(respBody, &member); err != nil {
		return nil, fmt.Errorf("解析成员详情失败: %w; 响应预览: %s", err, responsePreview(respBody))
	}

	return &member, nil
}

// UpdateMember 更新成员配置
func (c *Client) UpdateMember(networkID, memberID string, member *MemberUpdateRequest) (*Member, error) {
	endpoint := fmt.Sprintf("/controller/network/%s/member/%s", networkID, memberID)
	respBody, err := c.doRequest("POST", endpoint, member)
	if err != nil {
		return nil, err
	}

	var updatedMember Member
	if err := json.Unmarshal(respBody, &updatedMember); err != nil {
		return nil, fmt.Errorf("解析更新成员响应失败: %w; 响应预览: %s", err, responsePreview(respBody))
	}

	return &updatedMember, nil
}

// DeleteMember 移除网络成员
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
		return nil, fmt.Errorf("解析成员列表失败: %w; 响应预览: %s", err, responsePreview(respBody))
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
		return nil, fmt.Errorf("解析成员索引失败: %w; 响应预览: %s", err, responsePreview(respBody))
	}

	memberIDs := make([]string, 0, len(memberIndex))
	for memberID := range memberIndex {
		memberIDs = append(memberIDs, memberID)
	}
	sort.Strings(memberIDs)

	return memberIDs, nil
}

func responsePreview(respBody []byte) string {
	preview := strings.TrimSpace(string(respBody))
	if preview == "" {
		return "<empty>"
	}
	if len(preview) > responsePreviewLimit {
		return preview[:responsePreviewLimit] + "..."
	}
	return preview
}
