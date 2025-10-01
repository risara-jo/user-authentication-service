package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"smart-transit-system/internal/audit"
	"smart-transit-system/internal/auth"
	"smart-transit-system/internal/models"
	"smart-transit-system/internal/repository"
	"smart-transit-system/internal/scim"
)

func TestListProfiles_OrgIsolation(t *testing.T) {
	repo := &stubProfileRepo{profiles: []models.UserProfile{}, total: 0}
	svc := NewService(repo, stubScimClient{}, audit.NewLogger(nil))

	actor := auth.ContextValues{
		Scopes:    []string{"org.manage"},
		CompanyID: "comp-1",
		LoungeID:  "lounge-1",
	}

	_, err := svc.ListProfiles(context.Background(), actor, ListProfilesInput{})
	require.NoError(t, err)
	require.Equal(t, "comp-1", repo.lastFilter.CompanyID)
	require.Equal(t, "lounge-1", repo.lastFilter.LoungeID)

	// When users.manage present, no forced override.
	actor = auth.ContextValues{Scopes: []string{"users.manage"}}
	repo.lastFilter = repository.ProfileFilter{}
	_, err = svc.ListProfiles(context.Background(), actor, ListProfilesInput{CompanyID: "custom"})
	require.NoError(t, err)
	require.Equal(t, "custom", repo.lastFilter.CompanyID)
}

type stubProfileRepo struct {
	profiles   []models.UserProfile
	total      int64
	err        error
	lastFilter repository.ProfileFilter
}

func (s *stubProfileRepo) List(ctx context.Context, filter repository.ProfileFilter) ([]models.UserProfile, int64, error) {
	s.lastFilter = filter
	return s.profiles, s.total, s.err
}

func (s *stubProfileRepo) GetByID(ctx context.Context, id string) (*models.UserProfile, error) {
	return nil, repository.ErrProfileNotFound
}

func (s *stubProfileRepo) GetByEmail(ctx context.Context, email string) (*models.UserProfile, error) {
	return nil, repository.ErrProfileNotFound
}

func (s *stubProfileRepo) GetBySub(ctx context.Context, sub string) (*models.UserProfile, error) {
	return nil, repository.ErrProfileNotFound
}

func (s *stubProfileRepo) Create(ctx context.Context, profile *models.UserProfile) error {
	return errors.New("not implemented")
}

func (s *stubProfileRepo) UpdateProfile(ctx context.Context, id string, update repository.ProfileUpdate) (*models.UserProfile, error) {
	return nil, errors.New("not implemented")
}

func (s *stubProfileRepo) ReplaceRoles(ctx context.Context, id string, roles []string) (*models.UserProfile, error) {
	return nil, errors.New("not implemented")
}

func (s *stubProfileRepo) UpdateStatus(ctx context.Context, id string, status string) (*models.UserProfile, error) {
	return nil, errors.New("not implemented")
}

type stubScimClient struct{}

func (stubScimClient) CreateUser(ctx context.Context, req scim.CreateUserRequest) (*scim.User, error) {
	return nil, errors.New("not implemented")
}

func (stubScimClient) UpdateUser(ctx context.Context, id string, req scim.UpdateUserRequest) error {
	return errors.New("not implemented")
}

func (stubScimClient) ReplaceRoles(ctx context.Context, id string, roles []string) error {
	return errors.New("not implemented")
}

func (stubScimClient) UpdateStatus(ctx context.Context, id string, active bool) error {
	return errors.New("not implemented")
}
