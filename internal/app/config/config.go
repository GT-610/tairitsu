package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用程序配置结构体
type Config struct {
	// ZeroTier配置
	ZeroTier struct {
		URL       string `mapstructure:"ZT_CONTROLLER_URL"`
		TokenPath string `mapstructure:"ZT_TOKEN_PATH"`
	}

	// 服务器配置
	Server struct {
		Port int    `mapstructure:"SERVER_PORT"`
		Host string `mapstructure:"SERVER_HOST"`
	}

	// 安全配置
	Security struct {
		JWTSecret    string `mapstructure:"JWT_SECRET"`
		SessionSecret string `mapstructure:"SESSION_SECRET"`
	}

	// 数据库配置（可选）
	DatabaseURL string `mapstructure:"DATABASE_URL"`
}

// AppConfig 全局配置实例
var AppConfig *Config

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath("./")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取环境变量文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果.env文件不存在，尝试从环境变量加载
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
		fmt.Println("未找到配置文件，将从环境变量加载配置")
	} else {
		fmt.Println("成功读取配置文件")
	}

	// 创建配置实例
	cfg := &Config{}

	// 设置默认值
	viper.SetDefault("ZT_CONTROLLER_URL", "http://localhost:9993")
	viper.SetDefault("ZT_TOKEN_PATH", "/var/lib/zerotier-one/authtoken.secret")
	viper.SetDefault("SERVER_PORT", 8080)
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("JWT_SECRET", "default_jwt_secret_key_change_in_production")
	viper.SetDefault("SESSION_SECRET", "default_session_secret_change_in_production")

	// 解析配置到结构体
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	
	// 手动设置配置值
	cfg.ZeroTier.URL = viper.GetString("ZT_CONTROLLER_URL")
	cfg.ZeroTier.TokenPath = viper.GetString("ZT_TOKEN_PATH")
	cfg.Server.Port = viper.GetInt("SERVER_PORT")
	cfg.Server.Host = viper.GetString("SERVER_HOST")
	cfg.Security.JWTSecret = viper.GetString("JWT_SECRET")
	cfg.Security.SessionSecret = viper.GetString("SESSION_SECRET")

	// 确保AppConfig更新
	AppConfig = cfg

	return cfg, nil
}

// GetZTToken 获取ZeroTier令牌
func GetZTToken(tokenPath string) (string, error) {
	// 首先尝试从环境变量获取
	token := viper.GetString("ZT_TOKEN")
	if token != "" {
		return token, nil
	}

	// 如果没有直接配置，则尝试从文件读取
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		return "", fmt.Errorf("ZT令牌文件不存在: %s", tokenPath)
	}

	content, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", fmt.Errorf("读取ZT令牌文件失败: %w", err)
	}

	return strings.TrimSpace(string(content)), nil
}