package models

import "time"

// NetworkViewer 表示网络的只读成员设备查看授权
type NetworkViewer struct {
	NetworkID  string    `json:"network_id" gorm:"primaryKey;index"`
	UserID     string    `json:"user_id" gorm:"primaryKey;index"`
	GrantedBy  string    `json:"granted_by" gorm:"index"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (NetworkViewer) TableName() string {
	return "network_viewers"
}
