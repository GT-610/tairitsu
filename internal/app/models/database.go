package models

// DatabaseConfig is the API request shape for database configuration.
type DatabaseConfig struct {
	Type string `json:"type"` // Database type
	Path string `json:"path,omitempty"` // SQLite database path
	Host string `json:"host,omitempty"` // PostgreSQL/MySQL host
	Port int    `json:"port,omitempty"` // PostgreSQL/MySQL port
	User string `json:"user,omitempty"` // PostgreSQL/MySQL user
	Pass string `json:"pass,omitempty"` // PostgreSQL/MySQL password
	Name string `json:"name,omitempty"` // PostgreSQL/MySQL database name
}