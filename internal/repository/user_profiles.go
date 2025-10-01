package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"smart-transit-system/internal/models"
)

var (
	ErrProfileNotFound = errors.New("user profile not found")
	ErrDuplicateEmail  = errors.New("user profile with email already exists")
)

type ProfileFilter struct {
	Query      string
	Role       string
	CompanyID  string
	LoungeID   string
	CompanyIDs []string
	LoungeIDs  []string
	Status     string
	Page       int
	Limit      int
}

type ProfileRepository interface {
	List(ctx context.Context, filter ProfileFilter) ([]models.UserProfile, int64, error)
	GetByID(ctx context.Context, id string) (*models.UserProfile, error)
	GetByEmail(ctx context.Context, email string) (*models.UserProfile, error)
	GetBySub(ctx context.Context, sub string) (*models.UserProfile, error)
	Create(ctx context.Context, profile *models.UserProfile) error
	UpdateProfile(ctx context.Context, id string, update ProfileUpdate) (*models.UserProfile, error)
	ReplaceRoles(ctx context.Context, id string, roles []string) (*models.UserProfile, error)
	UpdateStatus(ctx context.Context, id string, status string) (*models.UserProfile, error)
}

type userProfileRepository struct {
	db *gorm.DB
}

func NewUserProfileRepository(db *gorm.DB) ProfileRepository {
	return &userProfileRepository{db: db}
}

type ProfileUpdate struct {
	Name      *string
	Email     *string
	CompanyID *string
	LoungeID  *string
}

func (r *userProfileRepository) List(ctx context.Context, filter ProfileFilter) ([]models.UserProfile, int64, error) {
	base := applyProfileFilters(r.db.WithContext(ctx).Model(&models.UserProfile{}), filter)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []models.UserProfile{}, 0, nil
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	var profiles []models.UserProfile
	query := applyProfileFilters(r.db.WithContext(ctx).Model(&models.UserProfile{}), filter)
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&profiles).Error; err != nil {
		return nil, 0, err
	}

	return profiles, total, nil
}

func (r *userProfileRepository) GetByID(ctx context.Context, id string) (*models.UserProfile, error) {
	var profile models.UserProfile
	if err := r.db.WithContext(ctx).First(&profile, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (r *userProfileRepository) GetByEmail(ctx context.Context, email string) (*models.UserProfile, error) {
	var profile models.UserProfile
	if err := r.db.WithContext(ctx).First(&profile, "LOWER(email) = LOWER(?)", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (r *userProfileRepository) GetBySub(ctx context.Context, sub string) (*models.UserProfile, error) {
	var profile models.UserProfile
	if err := r.db.WithContext(ctx).First(&profile, "sub = ?", sub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (r *userProfileRepository) Create(ctx context.Context, profile *models.UserProfile) error {
	if err := r.db.WithContext(ctx).Create(profile).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateEmail
		}
		return err
	}
	return nil
}

func (r *userProfileRepository) UpdateProfile(ctx context.Context, id string, update ProfileUpdate) (*models.UserProfile, error) {
	changes := make(map[string]any)
	if update.Name != nil {
		changes["name"] = strings.TrimSpace(*update.Name)
	}
	if update.Email != nil {
		changes["email"] = strings.TrimSpace(*update.Email)
	}
	if update.CompanyID != nil {
		if *update.CompanyID == "" {
			changes["company_id"] = nil
		} else {
			changes["company_id"] = strings.TrimSpace(*update.CompanyID)
		}
	}
	if update.LoungeID != nil {
		if *update.LoungeID == "" {
			changes["lounge_id"] = nil
		} else {
			changes["lounge_id"] = strings.TrimSpace(*update.LoungeID)
		}
	}
	if len(changes) == 0 {
		return r.GetByID(ctx, id)
	}

	res := r.db.WithContext(ctx).Model(&models.UserProfile{}).Where("id = ?", id).Updates(changes)
	if res.Error != nil {
		if isUniqueViolation(res.Error) {
			return nil, ErrDuplicateEmail
		}
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrProfileNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *userProfileRepository) ReplaceRoles(ctx context.Context, id string, roles []string) (*models.UserProfile, error) {
	clean := make([]string, 0, len(roles))
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role != "" {
			clean = append(clean, role)
		}
	}
	res := r.db.WithContext(ctx).Model(&models.UserProfile{}).Where("id = ?", id).Update("roles", pq.StringArray(clean))
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrProfileNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *userProfileRepository) UpdateStatus(ctx context.Context, id string, status string) (*models.UserProfile, error) {
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		return nil, fmt.Errorf("status required")
	}
	res := r.db.WithContext(ctx).Model(&models.UserProfile{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrProfileNotFound
	}
	return r.GetByID(ctx, id)
}

func applyProfileFilters(db *gorm.DB, filter ProfileFilter) *gorm.DB {
	if filter.Query != "" {
		q := "%" + strings.TrimSpace(filter.Query) + "%"
		db = db.Where("email ILIKE ? OR name ILIKE ?", q, q)
	}
	if filter.Role != "" {
		db = db.Where("? = ANY(roles)", strings.TrimSpace(filter.Role))
	}
	if filter.CompanyID != "" {
		db = db.Where("company_id = ?", strings.TrimSpace(filter.CompanyID))
	}
	if len(filter.CompanyIDs) > 0 {
		db = db.Where("company_id IN ?", filter.CompanyIDs)
	}
	if filter.LoungeID != "" {
		db = db.Where("lounge_id = ?", strings.TrimSpace(filter.LoungeID))
	}
	if len(filter.LoungeIDs) > 0 {
		db = db.Where("lounge_id IN ?", filter.LoungeIDs)
	}
	if filter.Status != "" {
		db = db.Where("status = ?", strings.TrimSpace(filter.Status))
	}
	return db
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}
