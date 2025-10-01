package scim

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	schemaCoreUser   = "urn:ietf:params:scim:schemas:core:2.0:User"
	schemaPatch      = "urn:ietf:params:scim:api:messages:2.0:PatchOp"
	schemaEnterprise = "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
)

// Client handles SCIM 2.0 interactions with Asgardeo.
type Client struct {
	baseURL      string
	http         *http.Client
	token        string
	clientID     string
	clientSecret string
}

type Config struct {
	BaseURL      string
	StaticToken  string
	ClientID     string
	ClientSecret string
}

// New creates a SCIM client.
func New(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, errors.New("SCIM base URL required")
	}
	if _, err := url.Parse(cfg.BaseURL); err != nil {
		return nil, fmt.Errorf("invalid SCIM base URL: %w", err)
	}
	return &Client{
		baseURL:      strings.TrimRight(cfg.BaseURL, "/"),
		http:         &http.Client{Timeout: 15 * time.Second},
		token:        strings.TrimSpace(cfg.StaticToken),
		clientID:     strings.TrimSpace(cfg.ClientID),
		clientSecret: strings.TrimSpace(cfg.ClientSecret),
	}, nil
}

type User struct {
	ID        string `json:"id"`
	UserName  string `json:"userName"`
	Active    bool   `json:"active"`
	SubjectID string `json:"-"`
	Email     string `json:"-"`
}

type CreateUserRequest struct {
	Email             string
	Name              string
	Roles             []string
	CompanyID         *string
	LoungeID          *string
	TemporaryPassword *string
}

// CreateUser provisions a user in Asgardeo SCIM.
func (c *Client) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	payload := map[string]any{
		"schemas":     []string{schemaCoreUser},
		"userName":    strings.ToLower(strings.TrimSpace(req.Email)),
		"displayName": strings.TrimSpace(req.Name),
		"name": map[string]any{
			"givenName": strings.TrimSpace(req.Name),
		},
		"emails": []map[string]any{{"value": strings.TrimSpace(req.Email), "primary": true}},
		"active": true,
	}
	if req.TemporaryPassword != nil && strings.TrimSpace(*req.TemporaryPassword) != "" {
		payload["password"] = strings.TrimSpace(*req.TemporaryPassword)
	}

	ext := map[string]any{}
	if req.CompanyID != nil && strings.TrimSpace(*req.CompanyID) != "" {
		ext["department"] = strings.TrimSpace(*req.CompanyID)
	}
	if req.LoungeID != nil && strings.TrimSpace(*req.LoungeID) != "" {
		ext["costCenter"] = strings.TrimSpace(*req.LoungeID)
	}
	if len(ext) > 0 {
		payload[schemaEnterprise] = ext
	}
	if len(req.Roles) > 0 {
		payload["roles"] = req.Roles
	}

	var resp User
	if err := c.do(ctx, http.MethodPost, "/Users", payload, &resp); err != nil {
		return nil, err
	}
	resp.Email = req.Email
	resp.SubjectID = resp.ID
	return &resp, nil
}

type UpdateUserRequest struct {
	Name      *string
	Email     *string
	CompanyID *string
	LoungeID  *string
}

// UpdateUser patches user attributes.
func (c *Client) UpdateUser(ctx context.Context, id string, req UpdateUserRequest) error {
	operations := make([]map[string]any, 0)
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		operations = append(operations,
			map[string]any{"op": "replace", "path": "displayName", "value": name},
			map[string]any{"op": "replace", "path": "name.givenName", "value": name},
		)
	}
	if req.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*req.Email))
		operations = append(operations,
			map[string]any{"op": "replace", "path": "userName", "value": email},
			map[string]any{"op": "replace", "path": "emails", "value": []map[string]any{{"value": email, "primary": true}}},
		)
	}
	enterprise := map[string]any{}
	if req.CompanyID != nil {
		enterprise["department"] = strings.TrimSpace(*req.CompanyID)
	}
	if req.LoungeID != nil {
		enterprise["costCenter"] = strings.TrimSpace(*req.LoungeID)
	}
	if len(enterprise) > 0 {
		operations = append(operations, map[string]any{"op": "replace", "path": schemaEnterprise, "value": enterprise})
	}
	if len(operations) == 0 {
		return nil
	}
	payload := map[string]any{
		"schemas":    []string{schemaPatch},
		"Operations": operations,
	}
	return c.do(ctx, http.MethodPatch, "/Users/"+url.PathEscape(id), payload, nil)
}

// ReplaceRoles sets the roles array on the SCIM user.
func (c *Client) ReplaceRoles(ctx context.Context, id string, roles []string) error {
	clean := make([]string, 0, len(roles))
	for _, r := range roles {
		r = strings.TrimSpace(r)
		if r != "" {
			clean = append(clean, r)
		}
	}
	payload := map[string]any{
		"schemas": []string{schemaPatch},
		"Operations": []map[string]any{
			{"op": "replace", "path": "roles", "value": clean},
		},
	}
	return c.do(ctx, http.MethodPatch, "/Users/"+url.PathEscape(id), payload, nil)
}

// UpdateStatus toggles the active flag.
func (c *Client) UpdateStatus(ctx context.Context, id string, active bool) error {
	payload := map[string]any{
		"schemas": []string{schemaPatch},
		"Operations": []map[string]any{
			{"op": "replace", "path": "active", "value": active},
		},
	}
	return c.do(ctx, http.MethodPatch, "/Users/"+url.PathEscape(id), payload, nil)
}

func (c *Client) do(ctx context.Context, method string, path string, body any, out any) error {
	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	} else if c.clientID != "" && c.clientSecret != "" {
		req.SetBasicAuth(c.clientID, c.clientSecret)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var apiErr map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		detail, _ := apiErr["detail"].(string)
		if detail == "" {
			detail = resp.Status
		}
		return fmt.Errorf("scim error: %s (status %d)", detail, resp.StatusCode)
	}

	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
