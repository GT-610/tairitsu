package models

import (
	"time"
)

// Network represents a managed network record in the database.
type Network struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName returns the database table name for Network.
func (Network) TableName() string {
	return "networks"
}
