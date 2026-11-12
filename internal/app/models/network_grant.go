package models

import "time"

// NetworkGrantSubjectUser is the subject type for account grants. The subject
// type column is intentionally modelled now so reusable user groups can be
// added later without changing the table shape.
const NetworkGrantSubjectUser = "user"

// NetworkGrant grants one account a role on one network. Ownership is not a
// grant: it lives on Network.OwnerID, so the table only holds delegated roles.
type NetworkGrant struct {
	NetworkID   string    `json:"network_id" gorm:"primaryKey;index"`
	SubjectType string    `json:"subject_type" gorm:"primaryKey;size:16"`
	SubjectID   string    `json:"subject_id" gorm:"primaryKey;index"`
	Role        string    `json:"role" gorm:"size:16;index"`
	GrantedBy   string    `json:"granted_by" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (NetworkGrant) TableName() string {
	return "network_grants"
}
