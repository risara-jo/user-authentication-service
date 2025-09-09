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

    "github.com/gin-gonic/gin"
    jwt "github.com/golang-jwt/jwt/v4"
    "github.com/MicahParks/keyfunc"
)

// Auth holds the verification state and configuration.
type Auth struct {
    issuer   string
    audience string // optional
    jwks     *keyfunc.JWKS
    once     sync.Once
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

    // Fetch discovery
    req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, discURL, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("fetch discovery: %w", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("discovery status %d", resp.StatusCode)
    }
    var dd discoveryDoc
    if err := json.NewDecoder(resp.Body).Decode(&dd); err != nil {
        return nil, fmt.Errorf("decode discovery: %w", err)
    }
    if dd.JWKSURI == "" {
        return nil, errors.New("jwks_uri not found in discovery")
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

    return &Auth{issuer: iss, audience: audience, jwks: jwks}, nil
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

// Middleware verifies the bearer token and injects claims into context.
func (a *Auth) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authz := c.GetHeader("Authorization")
        if authz == "" || !strings.HasPrefix(strings.ToLower(authz), "bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
            return
        }
        tokenStr := strings.TrimSpace(authz[len("Bearer "):])

        parser := &jwt.Parser{ValidMethods: []string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512"}}
        parsed, err := parser.Parse(tokenStr, a.jwks.Keyfunc)
        if err != nil || !parsed.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }

        // Extract claims into map
        m, ok := parsed.Claims.(jwt.MapClaims)
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
            return
        }
        // Validate issuer, audience (optional), and time-based claims
        if !m.VerifyIssuer(a.issuer, true) {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid issuer"})
            return
        }
        if a.audience != "" && !m.VerifyAudience(a.audience, true) {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid audience"})
            return
        }
        if err := m.Valid(); err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "expired or not yet valid"})
            return
        }

        claims := Claims(m)
        c.Set(ContextClaimsKey, claims)
        c.Next()
    }
}

// RequireScopes ensures the token has all required scopes.
func RequireScopes(required ...string) gin.HandlerFunc {
    reqSet := make(map[string]struct{}, len(required))
    for _, s := range required {
        reqSet[s] = struct{}{}
    }
    return func(c *gin.Context) {
        val, ok := c.Get(ContextClaimsKey)
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no auth context"})
            return
        }
        claims, _ := val.(Claims)
        got := make(map[string]struct{})
        for _, s := range claims.Scopes() {
            got[s] = struct{}{}
        }
        for s := range reqSet {
            if _, ok := got[s]; !ok {
                c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "missing scope: " + s})
                return
            }
        }
        c.Next()
    }
}

const ContextClaimsKey = "authClaims"

// FromContext retrieves claims from Gin context.
func FromContext(c *gin.Context) (Claims, bool) {
    v, ok := c.Get(ContextClaimsKey)
    if !ok {
        return nil, false
    }
    cl, ok := v.(Claims)
    return cl, ok
}
