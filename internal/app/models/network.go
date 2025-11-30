package models

import (
	"time"
)

// Network represents a ZeroTier network with ownership information
type Network struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for the Network model
func (Network) TableName() string {
	return "networks"
}
