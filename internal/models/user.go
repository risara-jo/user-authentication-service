package models

import (
    "time"
)

type User struct {
    ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    Sub       string    `gorm:"uniqueIndex;not null"` // Asgardeo subject
    Email     string    `gorm:"index"`
    Phone     string
    FirstName string
    LastName  string
    Status    string    `gorm:"default:'active'"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Organization struct {
    ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    Type      string    `gorm:"not null"` // company|lounge|system
    Name      string    `gorm:"not null"`
    Status    string    `gorm:"default:'active'"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type UserOrgMembership struct {
    ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    UserID    string    `gorm:"type:uuid;index;not null"`
    OrgID     string    `gorm:"type:uuid;index;not null"`
    Role      string    `gorm:"not null"`
    AssignedAt time.Time `gorm:"autoCreateTime"`
}

type UserAudit struct {
    ID       string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    UserID   string    `gorm:"type:uuid;index"`
    ActorID  string    `gorm:"type:uuid;index"`
    Action   string    `gorm:"not null"`
    Details  string    `gorm:"type:text"`
    Ts       time.Time `gorm:"autoCreateTime"`
}

