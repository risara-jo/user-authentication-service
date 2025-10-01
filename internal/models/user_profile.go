package models

import (
	"crypto/rand"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	UserStatusActive   = "active"
	UserStatusDisabled = "disabled"
)

// UserProfile represents a transit user managed via Asgardeo + local DB.
type UserProfile struct {
	ID        string         `gorm:"type:char(26);primaryKey"`
	Sub       string         `gorm:"uniqueIndex;not null"`
	Email     string         `gorm:"uniqueIndex;not null"`
	Name      string         `gorm:"not null"`
	Roles     pq.StringArray `gorm:"type:text[];not null;default:'{}'"`
	CompanyID *string        `gorm:"type:text"`
	LoungeID  *string        `gorm:"type:text"`
	Status    string         `gorm:"type:varchar(16);not null;default:'active'"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
}

func (UserProfile) TableName() string {
	return "user_profiles"
}

func (u *UserProfile) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = newULID()
	}
	if u.Status == "" {
		u.Status = UserStatusActive
	}
	return nil
}

// AuditLog captures administrative changes for compliance.
type AuditLog struct {
	ID           string            `gorm:"type:char(26);primaryKey"`
	ActorSub     string            `gorm:"index;not null"`
	ActorEmail   string            `gorm:"not null"`
	ActorRoles   pq.StringArray    `gorm:"type:text[];default:'{}'"`
	TargetUserID string            `gorm:"index"`
	Action       string            `gorm:"not null"`
	Before       datatypes.JSONMap `gorm:"type:jsonb"`
	After        datatypes.JSONMap `gorm:"type:jsonb"`
	CreatedAt    time.Time         `gorm:"autoCreateTime"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = newULID()
	}
	return nil
}

var (
	ulidMutex   sync.Mutex
	ulidEntropy = ulid.Monotonic(rand.Reader, 0)
)

func newULID() string {
	ulidMutex.Lock()
	defer ulidMutex.Unlock()
	return ulid.MustNew(ulid.Timestamp(time.Now()), ulidEntropy).String()
}
