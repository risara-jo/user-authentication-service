package audit

import (
	"context"

	"github.com/lib/pq"
	"gorm.io/datatypes"

	"smart-transit-system/internal/models"
	"smart-transit-system/internal/repository"
)

type Actor struct {
	Subject string
	Email   string
	Roles   []string
}

type Event struct {
	Actor    Actor
	TargetID string
	Action   string
	Before   map[string]any
	After    map[string]any
}

type Logger struct {
	repo repository.AuditRepository
}

func NewLogger(repo repository.AuditRepository) *Logger {
	return &Logger{repo: repo}
}

func (l *Logger) Record(ctx context.Context, evt Event) error {
	if l == nil || l.repo == nil {
		return nil
	}
	log := &models.AuditLog{
		ActorSub:     evt.Actor.Subject,
		ActorEmail:   evt.Actor.Email,
		ActorRoles:   pq.StringArray(evt.Actor.Roles),
		TargetUserID: evt.TargetID,
		Action:       evt.Action,
	}
	if evt.Before != nil {
		log.Before = datatypes.JSONMap(evt.Before)
	}
	if evt.After != nil {
		log.After = datatypes.JSONMap(evt.After)
	}
	return l.repo.Create(ctx, log)
}
