package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/tairitsu/tairitsu/internal/app/crypto"
	"github.com/spf13/viper"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type DatabaseType `json:"type"`
	Path string       `json:"path"`
	Host string       `json:"host"`
	Port int          `json:"port"`
	User string       `json:"user"`
	Pass string       `json:"pass"` // 加密后的密码
	Name string       `json:"name"`
}

// DatabaseType 数据库类型
type DatabaseType string

const (
	SQLite     DatabaseType = "sqlite"
	PostgreSQL DatabaseType = "postgresql"
	MySQL      DatabaseType = "mysql"
)

// ZeroTierConfig ZeroTier配置
type ZeroTierConfig struct {
	URL       string `json:"url"`
	Token     string `json:"token"`     // 加密后的令牌
	TokenPath string `json:"tokenPath"` // 令牌文件路径
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWTSecret     string `json:"jwt_secret"`
	SessionSecret string `json:"session_secret"`
}

// Config 应用程序配置结构体
type Config struct {
	Initialized bool           `json:"initialized"`          // 初始化状态标识
	Database    DatabaseConfig `json:"database"`             // 数据库配置
	ZeroTier    ZeroTierConfig `json:"zerotier"`             // ZeroTier配置
	Server      ServerConfig   `json:"server"`               // 服务器配置
	Security    SecurityConfig `json:"security"`             // 安全配置
}

// AppConfig 全局配置实例
var AppConfig *Config

// tempSettings 内存中的临时设置，用于存储不需要持久化的配置
var tempSettings = make(map[string]string)
var tempSettingsMutex sync.RWMutex

const configFilePath = "./config.json"

// LoadConfig 加载配置（从config.json）
func LoadConfig() (*Config, error) {
	// 首先尝试从config.json加载
	cfg, err := loadConfigFromJSON()
	if err == nil {
		AppConfig = cfg
		return cfg, nil
	}

	// 如果config.json不存在或读取失败，创建默认配置
	cfg = createDefaultConfig()

	// 尝试从.env文件或环境变量加载部分配置
	loadEnvConfig(cfg)

	AppConfig = cfg
	return cfg, nil
}

// loadConfigFromJSON 从JSON文件加载配置
func loadConfigFromJSON() (*Config, error) {
	// 检查配置文件是否存在
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在")
	}

	// 读取配置文件
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析JSON
	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return cfg, nil
}

// SaveConfig 保存配置到JSON文件
func SaveConfig(cfg *Config) error {
	// 序列化配置
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 写入文件，设置适当的权限（仅所有者可读写）
	return os.WriteFile(configFilePath, data, 0600)
}

// createDefaultConfig 创建默认配置
func createDefaultConfig() *Config {
	return &Config{
		Initialized: false,
		Database: DatabaseConfig{
			Type: SQLite,
			Path: "tairitsu.db",
		},
		ZeroTier: ZeroTierConfig{
			URL:       "http://localhost:9993",
			TokenPath: "/var/lib/zerotier-one/authtoken.secret",
		},
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
		Security: SecurityConfig{
			JWTSecret:     "default_jwt_secret_key_change_in_production",
			SessionSecret: "default_session_secret_change_in_production",
		},
	}
}

// loadEnvConfig 从环境变量加载部分配置（仅读取.env中存在的配置）
func loadEnvConfig(cfg *Config) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath("./")
	viper.AutomaticEnv()

	// 尝试读取.env文件
	_ = viper.ReadInConfig() // 忽略错误，仅尝试读取

	// 从环境变量更新配置，只读取.env中存在的配置项
	if url := viper.GetString("ZT_CONTROLLER_URL"); url != "" {
		cfg.ZeroTier.URL = url
	}
	cfg.Server.Port = viper.GetInt("SERVER_PORT")
	if jwt := viper.GetString("JWT_SECRET"); jwt != "" {
		cfg.Security.JWTSecret = jwt
	}
	if session := viper.GetString("SESSION_SECRET"); session != "" {
		cfg.Security.SessionSecret = session
	}
	
	// 读取ZT_TOKEN_PATH并尝试从文件中读取令牌
	if tokenPath := viper.GetString("ZT_TOKEN_PATH"); tokenPath != "" {
		cfg.ZeroTier.TokenPath = tokenPath
		// 尝试读取令牌文件
		_ = LoadTokenFromPath(tokenPath)
	}
}

// GetZTToken 获取ZeroTier令牌（自动解密）
func GetZTToken() (string, error) {
	if AppConfig == nil {
		return "", fmt.Errorf("配置未加载")
	}

	// 解密令牌
	if AppConfig.ZeroTier.Token != "" {
		decryptedToken, err := decryptSensitiveData(AppConfig.ZeroTier.Token)
		if err != nil {
			// 如果解密失败，可能是未加密的数据，尝试直接返回
			// 这是为了兼容可能的未加密数据
			return AppConfig.ZeroTier.Token, nil
		}
		return decryptedToken, nil
	}

	// 如果配置中没有，尝试从环境变量获取
	token := viper.GetString("ZT_TOKEN")
	if token != "" {
		return token, nil
	}

	return "", fmt.Errorf("ZeroTier令牌未配置")
}

// SetZTToken 设置ZeroTier令牌（自动加密）
func SetZTToken(token string) error {
	if AppConfig == nil {
		return fmt.Errorf("配置未加载")
	}

	// 加密令牌
	encryptedToken, err := encryptSensitiveData(token)
	if err != nil {
		return fmt.Errorf("加密令牌失败: %w", err)
	}

	AppConfig.ZeroTier.Token = encryptedToken
	return nil
}

// SetZTConfig 设置ZeroTier配置
func SetZTConfig(url, tokenPath string) error {
	if AppConfig == nil {
		return fmt.Errorf("配置未加载")
	}

	AppConfig.ZeroTier.URL = url
	AppConfig.ZeroTier.TokenPath = tokenPath
	
	// 尝试从新的tokenPath读取令牌
	err := LoadTokenFromPath(tokenPath)
	if err != nil {
		return fmt.Errorf("读取令牌文件失败: %w", err)
	}
	
	return SaveConfig(AppConfig)
}

// LoadTokenFromPath 从指定路径加载ZeroTier令牌
func LoadTokenFromPath(path string) error {
	if AppConfig == nil {
		return fmt.Errorf("配置未加载")
	}

	// 尝试读取令牌文件
	tokenBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取令牌文件失败: %w", err)
	}

	// 成功读取文件，去除换行符并设置令牌
	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return fmt.Errorf("令牌文件为空")
	}

	return SetZTToken(token)
}

// GetDatabasePassword 获取数据库密码（自动解密）
func GetDatabasePassword() (string, error) {
	if AppConfig == nil {
		return "", fmt.Errorf("配置未加载")
	}

	if AppConfig.Database.Pass != "" {
		decryptedPass, err := decryptSensitiveData(AppConfig.Database.Pass)
		if err != nil {
			// 如果解密失败，可能是未加密的数据，尝试直接返回
			return AppConfig.Database.Pass, nil
		}
		return decryptedPass, nil
	}

	return "", nil
}

// SetDatabasePassword 设置数据库密码（自动加密）
func SetDatabasePassword(password string) error {
	if AppConfig == nil {
		return fmt.Errorf("配置未加载")
	}

	// 加密密码
	encryptedPass, err := encryptSensitiveData(password)
	if err != nil {
		return fmt.Errorf("加密密码失败: %w", err)
	}

	AppConfig.Database.Pass = encryptedPass
	return nil
}

// encryptSensitiveData 加密敏感数据
func encryptSensitiveData(data string) (string, error) {
	if AppConfig == nil {
		return "", fmt.Errorf("配置未加载")
	}

	// 使用JWT密钥作为加密密钥（简化方案）
	// 在生产环境中，建议使用单独的加密密钥
	key := AppConfig.Security.JWTSecret

	// 导入crypto包
	encrypted, err := crypto.Encrypt(data, key)
	if err != nil {
		return "", err
	}

	// 添加加密标识前缀
	return "encrypted:" + encrypted, nil
}

// decryptSensitiveData 解密敏感数据
func decryptSensitiveData(data string) (string, error) {
	if AppConfig == nil {
		return "", fmt.Errorf("配置未加载")
	}

	// 检查是否带有加密标识前缀
	if !strings.HasPrefix(data, "encrypted:") {
		// 如果没有加密标识，直接返回原始数据
		return data, nil
	}

	// 移除加密标识前缀
	encryptedData := strings.TrimPrefix(data, "encrypted:")

	// 使用JWT密钥作为解密密钥
	key := AppConfig.Security.JWTSecret

	// 导入crypto包
	return crypto.Decrypt(encryptedData, key)
}

// SetInitialized 设置初始化状态
func SetInitialized(initialized bool) error {
	if AppConfig == nil {
		return fmt.Errorf("配置未加载")
	}

	AppConfig.Initialized = initialized
	return SaveConfig(AppConfig)
}

// IsInitialized 检查是否已初始化
func IsInitialized() bool {
	if AppConfig == nil {
		return false
	}
	return AppConfig.Initialized
}

// GetTempSetting 获取临时设置
// 临时设置存储在内存中，不会持久化到配置文件
func GetTempSetting(key string) string {
	tempSettingsMutex.RLock()
	defer tempSettingsMutex.RUnlock()
	return tempSettings[key]
}

// SetTempSetting 设置临时设置
// 临时设置存储在内存中，不会持久化到配置文件
func SetTempSetting(key, value string) {
	tempSettingsMutex.Lock()
	defer tempSettingsMutex.Unlock()
	tempSettings[key] = value
}