package repository

import (
	"context"

	"gorm.io/gorm"

	"smart-transit-system/internal/models"
)

type AuditRepository interface {
	Create(ctx context.Context, log *models.AuditLog) error
}

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, log *models.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
