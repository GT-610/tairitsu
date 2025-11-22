package models

// DatabaseConfig 数据库配置模型
type DatabaseConfig struct {
	Type string `json:"type" binding:"required,oneof=sqlite postgresql mysql"` // 数据库类型
	Path string `json:"path,omitempty"`                                           // SQLite数据库路径
	Host string `json:"host,omitempty"`                                           // PostgreSQL/MySQL主机
	Port int    `json:"port,omitempty"`                                           // PostgreSQL/MySQL端口
	User string `json:"user,omitempty"`                                           // PostgreSQL/MySQL用户
	Pass string `json:"pass,omitempty"`                                           // PostgreSQL/MySQL密码
	Name string `json:"name,omitempty"`                                           // PostgreSQL/MySQL数据库名
}