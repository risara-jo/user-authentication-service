package admin

import (
	"context"
	"errors"
	"strings"

	"github.com/lib/pq"

	"smart-transit-system/internal/apperror"
	"smart-transit-system/internal/audit"
	"smart-transit-system/internal/auth"
	"smart-transit-system/internal/models"
	"smart-transit-system/internal/repository"
	"smart-transit-system/internal/scim"
)

var roleCatalog = []string{
	"passenger",
	"driver",
	"conductor",
	"company_owner",
	"lounge_owner",
	"admin",
}

type scimUserClient interface {
	CreateUser(ctx context.Context, req scim.CreateUserRequest) (*scim.User, error)
	UpdateUser(ctx context.Context, id string, req scim.UpdateUserRequest) error
	ReplaceRoles(ctx context.Context, id string, roles []string) error
	UpdateStatus(ctx context.Context, id string, active bool) error
}

type Service struct {
	repo  repository.ProfileRepository
	scim  scimUserClient
	audit *audit.Logger
}

func NewService(repo repository.ProfileRepository, scimClient scimUserClient, auditLogger *audit.Logger) *Service {
	return &Service{repo: repo, scim: scimClient, audit: auditLogger}
}

func (s *Service) ListRoles() []string {
	return append([]string(nil), roleCatalog...)
}

type ListProfilesInput struct {
	Query     string
	Role      string
	CompanyID string
	LoungeID  string
	Status    string
	Page      int
	Limit     int
}

type ListProfilesResult struct {
	Profiles []models.UserProfile
	Total    int64
	Page     int
	Limit    int
}

func (s *Service) ListProfiles(ctx context.Context, actor auth.ContextValues, input ListProfilesInput) (*ListProfilesResult, error) {
	if s.repo == nil {
		return nil, apperror.Internal("repository not configured", nil)
	}

	filter := repository.ProfileFilter{
		Query:     input.Query,
		Role:      input.Role,
		CompanyID: input.CompanyID,
		LoungeID:  input.LoungeID,
		Status:    input.Status,
		Page:      input.Page,
		Limit:     input.Limit,
	}

	hasUsersManage := contains(actor.Scopes, "users.manage")
	hasOrgManage := contains(actor.Scopes, "org.manage")
	if !hasUsersManage && hasOrgManage {
		if actor.CompanyID != "" {
			filter.CompanyID = actor.CompanyID
			filter.CompanyIDs = []string{actor.CompanyID}
		}
		if actor.LoungeID != "" {
			filter.LoungeID = actor.LoungeID
			filter.LoungeIDs = []string{actor.LoungeID}
		}
	}

	profiles, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	return &ListProfilesResult{Profiles: profiles, Total: total, Page: page, Limit: limit}, nil
}

type CreateUserInput struct {
	Email             string
	Name              string
	Roles             []string
	CompanyID         *string
	LoungeID          *string
	TemporaryPassword *string
}

func (s *Service) CreateUser(ctx context.Context, actor auth.ContextValues, input CreateUserInput) (*models.UserProfile, error) {
	if s.repo == nil || s.scim == nil {
		return nil, apperror.Internal("service dependencies missing", nil)
	}

	normalizedRoles := normalizeRoles(input.Roles)
	if len(normalizedRoles) == 0 {
		normalizedRoles = []string{"passenger"}
	}

	if existing, err := s.repo.GetByEmail(ctx, input.Email); err == nil && existing != nil {
		return nil, apperror.Conflict("user.email_exists", "user with email already exists", input.Email)
	} else if err != nil && !errors.Is(err, repository.ErrProfileNotFound) {
		return nil, err
	}

	scimUser, err := s.scim.CreateUser(ctx, scim.CreateUserRequest{
		Email:             input.Email,
		Name:              input.Name,
		Roles:             normalizedRoles,
		CompanyID:         input.CompanyID,
		LoungeID:          input.LoungeID,
		TemporaryPassword: input.TemporaryPassword,
	})
	if err != nil {
		return nil, apperror.Internal("scim create failed", err.Error())
	}

	profile := &models.UserProfile{
		Sub:    scimUser.ID,
		Email:  strings.ToLower(strings.TrimSpace(input.Email)),
		Name:   strings.TrimSpace(input.Name),
		Roles:  pq.StringArray(normalizedRoles),
		Status: models.UserStatusActive,
	}
	if input.CompanyID != nil && strings.TrimSpace(*input.CompanyID) != "" {
		cid := strings.TrimSpace(*input.CompanyID)
		profile.CompanyID = &cid
	}
	if input.LoungeID != nil && strings.TrimSpace(*input.LoungeID) != "" {
		lid := strings.TrimSpace(*input.LoungeID)
		profile.LoungeID = &lid
	}

	if err := s.repo.Create(ctx, profile); err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil, apperror.Conflict("user.email_exists", "user with email already exists", input.Email)
		}
		return nil, err
	}

	s.audit.Record(ctx, audit.Event{
		Actor:    actorToAudit(actor),
		TargetID: profile.ID,
		Action:   "admin.user.create",
		After:    profileToMap(profile),
	})

	return profile, nil
}

type UpdateUserInput struct {
	Name        *string
	Email       *string
	CompanyID   *string
	LoungeID    *string
	PrimaryRole *string
}

func (s *Service) UpdateUser(ctx context.Context, actor auth.ContextValues, id string, input UpdateUserInput) (*models.UserProfile, error) {
	profile, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			return nil, apperror.NotFound("user.not_found", "user profile not found")
		}
		return nil, err
	}

	before := profileToMap(profile)

	scimReq := scim.UpdateUserRequest{}
	repoUpdate := repository.ProfileUpdate{}

	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		scimReq.Name = &trimmed
		repoUpdate.Name = &trimmed
	}
	if input.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*input.Email))
		scimReq.Email = &email
		repoUpdate.Email = &email
	}
	if input.CompanyID != nil {
		cid := strings.TrimSpace(*input.CompanyID)
		if cid == "" {
			repoUpdate.CompanyID = ptrString("")
			scimReq.CompanyID = ptrString("")
		} else {
			repoUpdate.CompanyID = &cid
			scimReq.CompanyID = &cid
		}
	}
	if input.LoungeID != nil {
		lid := strings.TrimSpace(*input.LoungeID)
		if lid == "" {
			repoUpdate.LoungeID = ptrString("")
			scimReq.LoungeID = ptrString("")
		} else {
			repoUpdate.LoungeID = &lid
			scimReq.LoungeID = &lid
		}
	}

	if scimReq.Name != nil || scimReq.Email != nil || scimReq.CompanyID != nil || scimReq.LoungeID != nil {
		if err := s.scim.UpdateUser(ctx, profile.Sub, scimReq); err != nil {
			return nil, apperror.Internal("scim update failed", err.Error())
		}
	}

	if repoUpdate != (repository.ProfileUpdate{}) {
		updated, err := s.repo.UpdateProfile(ctx, id, repoUpdate)
		if err != nil {
			if errors.Is(err, repository.ErrDuplicateEmail) {
				return nil, apperror.Conflict("user.email_exists", "user with email already exists", input.Email)
			}
			return nil, err
		}
		profile = updated
	}

	if input.PrimaryRole != nil {
		newRoles := reorderRoles(profile.Roles, *input.PrimaryRole)
		if err := s.scim.ReplaceRoles(ctx, profile.Sub, newRoles); err != nil {
			return nil, apperror.Internal("scim replace roles failed", err.Error())
		}
		updated, err := s.repo.ReplaceRoles(ctx, id, newRoles)
		if err != nil {
			return nil, err
		}
		profile = updated
	}

	s.audit.Record(ctx, audit.Event{
		Actor:    actorToAudit(actor),
		TargetID: profile.ID,
		Action:   "admin.user.update",
		Before:   before,
		After:    profileToMap(profile),
	})

	return profile, nil
}

func (s *Service) ReplaceRoles(ctx context.Context, actor auth.ContextValues, id string, roles []string) (*models.UserProfile, error) {
	profile, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			return nil, apperror.NotFound("user.not_found", "user profile not found")
		}
		return nil, err
	}

	before := profileToMap(profile)
	clean := normalizeRoles(roles)
	if len(clean) == 0 {
		return nil, apperror.BadRequest("user.roles_invalid", "roles must not be empty", nil)
	}

	if err := s.scim.ReplaceRoles(ctx, profile.Sub, clean); err != nil {
		return nil, apperror.Internal("scim replace roles failed", err.Error())
	}

	updated, err := s.repo.ReplaceRoles(ctx, id, clean)
	if err != nil {
		return nil, err
	}

	s.audit.Record(ctx, audit.Event{
		Actor:    actorToAudit(actor),
		TargetID: updated.ID,
		Action:   "admin.user.roles.replace",
		Before:   before,
		After:    profileToMap(updated),
	})

	return updated, nil
}

func (s *Service) UpdateStatus(ctx context.Context, actor auth.ContextValues, id string, status string) (*models.UserProfile, error) {
	profile, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			return nil, apperror.NotFound("user.not_found", "user profile not found")
		}
		return nil, err
	}

	before := profileToMap(profile)
	normalized := strings.ToLower(strings.TrimSpace(status))
	if normalized != models.UserStatusActive && normalized != models.UserStatusDisabled {
		return nil, apperror.BadRequest("user.status_invalid", "status must be active or disabled", status)
	}

	active := normalized == models.UserStatusActive
	if err := s.scim.UpdateStatus(ctx, profile.Sub, active); err != nil {
		return nil, apperror.Internal("scim status update failed", err.Error())
	}

	updated, err := s.repo.UpdateStatus(ctx, id, normalized)
	if err != nil {
		return nil, err
	}

	s.audit.Record(ctx, audit.Event{
		Actor:    actorToAudit(actor),
		TargetID: updated.ID,
		Action:   "admin.user.status.update",
		Before:   before,
		After:    profileToMap(updated),
	})

	return updated, nil
}

func normalizeRoles(roles []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(roles))
	for _, r := range roles {
		r = strings.TrimSpace(strings.ToLower(r))
		if r == "" {
			continue
		}
		if _, exists := seen[r]; exists {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	return out
}

func reorderRoles(existing pq.StringArray, primary string) []string {
	clean := normalizeRoles([]string(existing))
	primary = strings.TrimSpace(strings.ToLower(primary))
	if primary == "" {
		return clean
	}
	result := []string{primary}
	seen := make(map[string]struct{})
	seen[primary] = struct{}{}
	for _, r := range clean {
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		result = append(result, r)
	}
	return result
}

func actorToAudit(ctx auth.ContextValues) audit.Actor {
	return audit.Actor{Subject: ctx.Subject, Email: ctx.Email, Roles: append([]string(nil), ctx.Roles...)}
}

func profileToMap(p *models.UserProfile) map[string]any {
	if p == nil {
		return nil
	}
	m := map[string]any{
		"id":         p.ID,
		"sub":        p.Sub,
		"email":      p.Email,
		"name":       p.Name,
		"roles":      []string(p.Roles),
		"status":     p.Status,
		"company_id": p.CompanyID,
		"lounge_id":  p.LoungeID,
	}
	return m
}

func contains(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func ptrString(v string) *string {
	return &v
}
