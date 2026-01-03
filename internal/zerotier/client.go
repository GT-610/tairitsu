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

// Client ZeroTier API客户端
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

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
	V4AssignMode               AssignmentMode     `json:"v4AssignMode"`
	V6AssignMode               V6AssignmentMode   `json:"v6AssignMode"`
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
	// 获取配置
	if config.AppConfig == nil {
		return nil, fmt.Errorf("配置未初始化")
	}

	// 尝试从TokenPath加载令牌到配置中
	err := config.LoadTokenFromPath(config.AppConfig.ZeroTier.TokenPath)

	// 获取令牌（可能是从文件加载的，也可能是已存在于配置中的）
	token, err := config.GetZTToken()
	if err != nil {
		return nil, fmt.Errorf("获取ZeroTier令牌失败: %w", err)
	}

	// 创建HTTP客户端
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 使用配置中的URL创建客户端
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
		return nil, fmt.Errorf("解析状态响应失败: %w", err)
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

	var networkIDs []string
	if err := json.Unmarshal(respBody, &networkIDs); err != nil {
		return nil, fmt.Errorf("解析网络ID列表失败: %w", err)
	}

	return networkIDs, nil
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
		return "", fmt.Errorf("解析网络状态失败: %w", err)
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
		return nil, fmt.Errorf("解析网络详情失败: %w", err)
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
		return nil, fmt.Errorf("解析创建网络响应失败: %w", err)
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
		return nil, fmt.Errorf("解析更新网络响应失败: %w", err)
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
		return nil, fmt.Errorf("解析更新网络响应失败: %w", err)
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

	var members []Member
	if err := json.Unmarshal(respBody, &members); err != nil {
		return nil, fmt.Errorf("解析成员列表失败: %w", err)
	}

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
		return nil, fmt.Errorf("解析成员详情失败: %w", err)
	}

	return &member, nil
}

// UpdateMember 更新成员配置
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

// DeleteMember 移除网络成员
func (c *Client) DeleteMember(networkID, memberID string) error {
	endpoint := fmt.Sprintf("/controller/network/%s/member/%s", networkID, memberID)
	_, err := c.doRequest("DELETE", endpoint, nil)
	return err
}
