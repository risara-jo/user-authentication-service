package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"

	"smart-transit-system/internal/httpx"
)

// Auth holds the verification state and configuration.
type Auth struct {
	issuer   string
	audience string // optional
	jwks     *keyfunc.JWKS
	once     sync.Once
	tenant   string // extracted from issuer path (/t/{tenant}) for tolerant checks
}

type discoveryDoc struct {
	Issuer  string `json:"issuer"`
	JWKSURI string `json:"jwks_uri"`
}

// New creates an Auth instance by discovering the JWKS from the issuer.
func New(issuer string, audience string, cacheMinutes int) (*Auth, error) {
	if issuer == "" {
		return nil, errors.New("issuer is required")
	}

	// Normalize issuer: trim trailing slash for consistency
	iss := strings.TrimRight(issuer, "/")
	discURL := iss + "/.well-known/openid-configuration"

	// Fetch discovery (with fallback to issuer + "/jwks")
	var dd discoveryDoc
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, discURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err == nil && resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			if decErr := json.NewDecoder(resp.Body).Decode(&dd); decErr == nil && dd.JWKSURI != "" {
				// ok
			}
		}
	}
	// Fallback if jwks_uri missing
	if dd.JWKSURI == "" {
		dd.Issuer = iss
		dd.JWKSURI = iss + "/jwks"
	}

	// Prefer the issuer value reported by discovery when available.
	if dd.Issuer != "" {
		iss = strings.TrimRight(dd.Issuer, "/")
	}

	// Build JWKS with background refresh
	refreshInt := time.Duration(cacheMinutes) * time.Minute
	if refreshInt <= 0 {
		refreshInt = 60 * time.Minute
	}
	jwks, err := keyfunc.Get(dd.JWKSURI, keyfunc.Options{
		RefreshErrorHandler: func(err error) {
			// no-op logging; could integrate with a real logger
		},
		RefreshInterval: refreshInt,
		RefreshTimeout:  10 * time.Second,
		// Follow any headers recommending refresh
		RefreshUnknownKID: true,
	})
	if err != nil {
		return nil, fmt.Errorf("load jwks: %w", err)
	}

	// Extract tenant from issuer path (first segment after /t/)
	tenant := ""
	if i := strings.Index(iss, "/t/"); i >= 0 {
		rest := iss[i+3:]
		if j := strings.Index(rest, "/"); j >= 0 {
			tenant = rest[:j]
		} else {
			tenant = rest
		}
	}

	return &Auth{issuer: iss, audience: audience, jwks: jwks, tenant: tenant}, nil
}

// Claims is a permissive map of token claims with helpers.
type Claims map[string]any

func (c Claims) Subject() string {
	if v, ok := c["sub"].(string); ok {
		return v
	}
	return ""
}
func (c Claims) Email() string {
	if v, ok := c["email"].(string); ok {
		return v
	}
	return ""
}

func (c Claims) Scopes() []string {
	// scope as space-delimited string
	if v, ok := c["scope"].(string); ok && v != "" {
		parts := strings.Fields(v)
		return parts
	}
	// scp as array
	if arr, ok := c["scp"].([]any); ok {
		out := make([]string, 0, len(arr))
		for _, x := range arr {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return nil
}

func (c Claims) Roles() []string {
	// roles as array
	if arr, ok := c["roles"].([]any); ok {
		out := make([]string, 0, len(arr))
		for _, x := range arr {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	// roles as string (space or comma delimited)
	if s, ok := c["roles"].(string); ok && s != "" {
		sep := ","
		if strings.Contains(s, " ") {
			sep = " "
		}
		parts := strings.Split(s, sep)
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}
	// groups fallback
	if arr, ok := c["groups"].([]any); ok {
		out := make([]string, 0, len(arr))
		for _, x := range arr {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return nil
}

func (c Claims) CompanyID() string {
	if v, ok := c["company_id"].(string); ok {
		return strings.TrimSpace(v)
	}
	if v, ok := c["companyId"].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

func (c Claims) LoungeID() string {
	if v, ok := c["lounge_id"].(string); ok {
		return strings.TrimSpace(v)
	}
	if v, ok := c["loungeId"].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

// Middleware verifies the bearer token and injects claims into context.
func (a *Auth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authz := c.GetHeader("Authorization")
		if authz == "" || !strings.HasPrefix(strings.ToLower(authz), "bearer ") {
			httpx.RespondError(c, http.StatusUnauthorized, "auth.missing_token", "missing bearer token", nil)
			c.Abort()
			return
		}
		tokenStr := strings.TrimSpace(authz[len("Bearer "):])

		parser := &jwt.Parser{ValidMethods: []string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512"}}
		parsed, err := parser.Parse(tokenStr, a.jwks.Keyfunc)
		if err != nil || !parsed.Valid {
			detail := any("token validation failed")
			if err != nil {
				detail = err.Error()
			}
			httpx.RespondError(c, http.StatusUnauthorized, "auth.invalid_token", "invalid token", detail)
			c.Abort()
			return
		}

		// Extract claims into map
		m, ok := parsed.Claims.(jwt.MapClaims)
		if !ok {
			httpx.RespondError(c, http.StatusUnauthorized, "auth.invalid_claims", "invalid claims payload", nil)
			c.Abort()
			return
		}
		// Validate issuer, audience (optional), and time-based claims
		// Issuer check with tolerance for trailing slash differences
		issClaim, _ := m["iss"].(string)
		expected := strings.TrimRight(a.issuer, "/")
		got := strings.TrimRight(issClaim, "/")
		// Normalize to tolerate trailing slash, optional /oauth2 or /oidc segment,
		// and host aliases (api/sts). Also accept any issuer that clearly targets
		// the same tenant (/t/{tenant}).
		canon := func(s string) string {
			s = strings.TrimSpace(s)
			s = strings.TrimRight(s, "/")
			s = strings.TrimSuffix(s, "/oauth2")
			s = strings.TrimSuffix(s, "/oidc")
			s = strings.ReplaceAll(s, "api.asgardeo.io", "asgardeo.io")
			s = strings.ReplaceAll(s, "sts.asgardeo.io", "asgardeo.io")
			return s
		}
		eg := canon(expected)
		gg := canon(got)
		valid := eg != "" && gg == eg
		if !valid && a.tenant != "" {
			// Tenant-aware relaxed check
			// Accept if token issuer contains /t/{tenant} after canonicalization.
			if strings.Contains(gg, "/t/"+a.tenant) {
				valid = true
			}
		}
		if !valid {
			httpx.RespondError(c, http.StatusUnauthorized, "auth.invalid_issuer", "invalid issuer", gin.H{"expected": eg, "got": gg})
			c.Abort()
			return
		}
		if a.audience != "" && !m.VerifyAudience(a.audience, true) {
			httpx.RespondError(c, http.StatusUnauthorized, "auth.invalid_audience", "invalid audience", gin.H{"expected": a.audience})
			c.Abort()
			return
		}
		if err := m.Valid(); err != nil {
			httpx.RespondError(c, http.StatusUnauthorized, "auth.token_expired", "token expired or not yet valid", err.Error())
			c.Abort()
			return
		}

		claims := Claims(m)
		c.Set(ContextClaimsKey, claims)
		c.Set(ContextValuesKey, buildContextValues(claims))
		c.Next()
	}
}

// ScopeGuard ensures the token has all required scopes.
func ScopeGuard(required ...string) gin.HandlerFunc {
	reqSet := make(map[string]struct{}, len(required))
	for _, s := range required {
		reqSet[strings.TrimSpace(s)] = struct{}{}
	}

	return func(c *gin.Context) {
		ctx, ok := ValuesFromContext(c)
		if !ok {
			httpx.RespondError(c, http.StatusUnauthorized, "auth.no_context", "authentication required", nil)
			c.Abort()
			return
		}

		got := make(map[string]struct{}, len(ctx.Scopes))
		for _, s := range ctx.Scopes {
			got[s] = struct{}{}
		}

		var missing []string
		for s := range reqSet {
			if _, ok := got[s]; !ok {
				missing = append(missing, s)
			}
		}
		if len(missing) > 0 {
			httpx.RespondError(c, http.StatusForbidden, "auth.missing_scope", "missing required scopes", missing)
			c.Abort()
			return
		}

		c.Next()
	}
}

// ScopeGuardAny ensures the token has at least one of the required scopes.
func ScopeGuardAny(required ...string) gin.HandlerFunc {
	req := make([]string, 0, len(required))
	for _, s := range required {
		s = strings.TrimSpace(s)
		if s != "" {
			req = append(req, s)
		}
	}
	return func(c *gin.Context) {
		ctx, ok := ValuesFromContext(c)
		if !ok {
			httpx.RespondError(c, http.StatusUnauthorized, "auth.no_context", "authentication required", nil)
			c.Abort()
			return
		}
		got := make(map[string]struct{}, len(ctx.Scopes))
		for _, s := range ctx.Scopes {
			got[s] = struct{}{}
		}
		for _, s := range req {
			if _, ok := got[s]; ok {
				c.Next()
				return
			}
		}
		httpx.RespondError(c, http.StatusForbidden, "auth.missing_scope", "missing required scope", req)
		c.Abort()
	}
}

const (
	ContextClaimsKey = "authClaims"
	ContextValuesKey = "authContext"
)

// ContextValues represents normalized data extracted from JWT claims.
type ContextValues struct {
	Claims    Claims
	Subject   string
	Email     string
	Roles     []string
	Scopes    []string
	CompanyID string
	LoungeID  string
}

// FromContext retrieves raw claims from Gin context.
func FromContext(c *gin.Context) (Claims, bool) {
	v, ok := c.Get(ContextClaimsKey)
	if !ok {
		return nil, false
	}
	cl, ok := v.(Claims)
	return cl, ok
}

// ValuesFromContext retrieves normalized context values set by the middleware.
func ValuesFromContext(c *gin.Context) (ContextValues, bool) {
	v, ok := c.Get(ContextValuesKey)
	if !ok {
		return ContextValues{}, false
	}
	ctx, ok := v.(ContextValues)
	return ctx, ok
}

func buildContextValues(cl Claims) ContextValues {
	roles := cl.Roles()
	scopes := cl.Scopes()
	return ContextValues{
		Claims:    cl,
		Subject:   cl.Subject(),
		Email:     cl.Email(),
		Roles:     append([]string(nil), roles...),
		Scopes:    append([]string(nil), scopes...),
		CompanyID: cl.CompanyID(),
		LoungeID:  cl.LoungeID(),
	}
}
