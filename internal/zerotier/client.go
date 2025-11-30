package zerotier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/GT-610/tairitsu/internal/app/config"
)

// Client ZeroTier API client
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// Network Network structure
type Network struct {
	ID          string        `json:"id"`
	Name        string        `json:"name" binding:"required"`
	Description string        `json:"description"`
	Config      NetworkConfig `json:"config" binding:"required"`
	Created     int64         `json:"creationTime"`
	Modified    int64         `json:"lastModifiedTime"`
}

// NetworkConfig Network configuration
type NetworkConfig struct {
	Private                    bool               `json:"private"`
	AllowPassivePortForwarding bool               `json:"allowPassivePortForwarding"`
	IPAssignmentPools          []IPAssignmentPool `json:"ipAssignmentPools"`
	Routes                     []Route            `json:"routes"`
	Tags                       []Tag              `json:"tags"`
	Rules                      []Rule             `json:"rules"`
	V4AssignMode               AssignmentMode     `json:"v4AssignMode"`
	V6AssignMode               AssignmentMode     `json:"v6AssignMode"`
}

// IPAssignmentPool IP assignment pool
type IPAssignmentPool struct {
	IPRangeStart string `json:"ipRangeStart"`
	IPRangeEnd   string `json:"ipRangeEnd"`
}

// Route Route
type Route struct {
	Target string `json:"target"`
	Via    string `json:"via,omitempty"`
}

// Tag Tag
type Tag struct {
	ID    int `json:"id"`
	Value int `json:"value"`
}

// Rule Rule
type Rule struct {
	Not       bool   `json:"not"`
	Or        []Rule `json:"or,omitempty"`
	Type      string `json:"type"`
	Metric    int    `json:"metric"`
	PortMin   int    `json:"portMin,omitempty"`
	PortMax   int    `json:"portMax,omitempty"`
	EthType   int    `json:"etherType,omitempty"`
	IPVersion int    `json:"ipVersion,omitempty"`
	Action    string `json:"action"`
}

// AssignmentMode Assignment mode
type AssignmentMode struct {
	ZT bool `json:"zt"`
}

// Member Network member
type Member struct {
	ID            string       `json:"id"`
	Address       string       `json:"address"`
	Config        MemberConfig `json:"config"`
	Identity      string       `json:"identity"`
	Name          string       `json:"name"`
	Description   string       `json:"description"`
	ClientVersion string       `json:"clientVersion"`
	Online        bool         `json:"online"`
	LastSeen      int64        `json:"lastOnline"`
	CreationTime  int64        `json:"creationTime"`
}

// MemberConfig Member configuration
type MemberConfig struct {
	Authorized      bool     `json:"authorized"`
	ActiveBridge    bool     `json:"activeBridge"`
	IPAssignments   []string `json:"ipAssignments"`
	Tags            []Tag    `json:"tags"`
	NATTraversal    bool     `json:"natTraversal"`
	Capabilities    []int    `json:"capabilities"`
	NoAutoAssignIPs bool     `json:"noAutoAssignIps"`
}

// Status ZeroTier status
type Status struct {
	Version     string `json:"version"`
	Address     string `json:"address"`
	Online      bool   `json:"online"`
	TCPFallback bool   `json:"tcpFallbackAvailable"`
	APIReady    bool   `json:"apiReady"`
}

// NewClient Create a new ZeroTier client
func NewClient() (*Client, error) {
	// Get configuration
	if config.AppConfig == nil {
		return nil, fmt.Errorf("配置未初始化")
	}

	// Try to load token from TokenPath into configuration
	err := config.LoadTokenFromPath(config.AppConfig.ZeroTier.TokenPath)

	// Get token (may be loaded from file or already exist in configuration)
	token, err := config.GetZTToken()
	if err != nil {
		return nil, fmt.Errorf("获取ZeroTier令牌失败: %w", err)
	}

	// Create HTTP client
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create client using URL from configuration
	baseURL := config.AppConfig.ZeroTier.URL
	if baseURL == "" {
		baseURL = "http://localhost:9993"
	}

	return &Client{
		BaseURL:    baseURL,
		Token:      token,
		HTTPClient: httpClient,
	}, nil
}

// doRequest Execute HTTP request
func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	// Build URL
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)

	// Create request body
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	// Create request
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ZT1-Auth", c.Token)

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("请求失败 (状态码: %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetStatus Get ZeroTier controller status
func (c *Client) GetStatus() (*Status, error) {
	respBody, err := c.doRequest("GET", "/status", nil)
	if err != nil {
		return nil, err
	}

	var status Status
	if err := json.Unmarshal(respBody, &status); err != nil {
		return nil, fmt.Errorf("解析状态响应失败: %w", err)
	}

	return &status, nil
}

// GetNetworks Get all networks list
func (c *Client) GetNetworks() ([]Network, error) {
	// Step 1: Get network ID array
	respBody, err := c.doRequest("GET", "/controller/network", nil)
	if err != nil {
		return nil, err
	}

	// Parse network ID array
	var networkIDs []string
	if err := json.Unmarshal(respBody, &networkIDs); err != nil {
		return nil, fmt.Errorf("解析网络ID列表失败: %w", err)
	}

	// Step 2: Iterate network IDs, get detailed information for each network
	// Critical fix: Initialize empty slice with make([]Network, 0) instead of var networks []Network
	networks := make([]Network, 0)
	for _, id := range networkIDs {
		// 调用GetNetwork获取单个网络的详细信息
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

// GetNetwork Get single network details
func (c *Client) GetNetwork(networkID string) (*Network, error) {
	endpoint := fmt.Sprintf("/controller/network/%s", networkID)
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var network Network
	if err := json.Unmarshal(respBody, &network); err != nil {
		return nil, fmt.Errorf("解析网络详情失败: %w", err)
	}

	return &network, nil
}

// CreateNetwork Create new network
func (c *Client) CreateNetwork(network *Network) (*Network, error) {
	respBody, err := c.doRequest("POST", "/controller/network", network)
	if err != nil {
		return nil, err
	}

	var createdNetwork Network
	if err := json.Unmarshal(respBody, &createdNetwork); err != nil {
		return nil, fmt.Errorf("解析创建网络响应失败: %w", err)
	}

	return &createdNetwork, nil
}

// UpdateNetwork Update network configuration
func (c *Client) UpdateNetwork(networkID string, network *Network) (*Network, error) {
	endpoint := fmt.Sprintf("/controller/network/%s", networkID)
	respBody, err := c.doRequest("POST", endpoint, network)
	if err != nil {
		return nil, err
	}

	var updatedNetwork Network
	if err := json.Unmarshal(respBody, &updatedNetwork); err != nil {
		return nil, fmt.Errorf("解析更新网络响应失败: %w", err)
	}

	return &updatedNetwork, nil
}

// DeleteNetwork Delete network
func (c *Client) DeleteNetwork(networkID string) error {
	endpoint := fmt.Sprintf("/controller/network/%s", networkID)
	_, err := c.doRequest("DELETE", endpoint, nil)
	return err
}

// GetMembers Get network members list
func (c *Client) GetMembers(networkID string) ([]Member, error) {
	endpoint := fmt.Sprintf("/controller/network/%s/member", networkID)
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var members []Member
	if err := json.Unmarshal(respBody, &members); err != nil {
		return nil, fmt.Errorf("解析成员列表失败: %w", err)
	}

	return members, nil
}

// GetMember Get single member details
func (c *Client) GetMember(networkID, memberID string) (*Member, error) {
	endpoint := fmt.Sprintf("/controller/network/%s/member/%s", networkID, memberID)
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var member Member
	if err := json.Unmarshal(respBody, &member); err != nil {
		return nil, fmt.Errorf("解析成员详情失败: %w", err)
	}

	return &member, nil
}

// UpdateMember Update member configuration
func (c *Client) UpdateMember(networkID, memberID string, member *Member) (*Member, error) {
	endpoint := fmt.Sprintf("/controller/network/%s/member/%s", networkID, memberID)
	respBody, err := c.doRequest("POST", endpoint, member)
	if err != nil {
		return nil, err
	}

	var updatedMember Member
	if err := json.Unmarshal(respBody, &updatedMember); err != nil {
		return nil, fmt.Errorf("解析更新成员响应失败: %w", err)
	}

	return &updatedMember, nil
}

// DeleteMember Remove network member
func (c *Client) DeleteMember(networkID, memberID string) error {
	endpoint := fmt.Sprintf("/controller/network/%s/member/%s", networkID, memberID)
	_, err := c.doRequest("DELETE", endpoint, nil)
	return err
}
