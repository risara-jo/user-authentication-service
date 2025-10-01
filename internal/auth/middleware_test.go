package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
)

func TestMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth, token := newTestAuth(t, map[string]any{
		"sub":   "user-123",
		"email": "user@example.com",
		"iss":   "https://issuer.example.com",
		"scope": "users.manage org.manage",
		"exp":   time.Now().Add(time.Minute).Unix(),
	})

	router := gin.New()
	router.GET("/protected", auth.Middleware(), func(c *gin.Context) {
		ctx, ok := ValuesFromContext(c)
		require.True(t, ok)
		require.Equal(t, "user-123", ctx.Subject)
		require.Contains(t, ctx.Scopes, "users.manage")
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)
}

func TestMiddleware_InvalidIssuer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth, token := newTestAuth(t, map[string]any{
		"sub": "user-123",
		"iss": "https://wrong.example.com",
		"exp": time.Now().Add(time.Minute).Unix(),
	})

	// Override expected issuer to something else.
	auth.issuer = "https://issuer.example.com"

	router := gin.New()
	router.GET("/protected", auth.Middleware(), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestScopeGuard(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth, token := newTestAuth(t, map[string]any{
		"sub":   "user-123",
		"iss":   "https://issuer.example.com",
		"scope": "org.manage",
		"exp":   time.Now().Add(time.Minute).Unix(),
	})
	auth.issuer = "https://issuer.example.com"

	router := gin.New()
	router.GET("/needs-scope",
		auth.Middleware(),
		ScopeGuard("users.manage"),
		func(c *gin.Context) { c.String(http.StatusOK, "ok") },
	)

	req := httptest.NewRequest(http.MethodGet, "/needs-scope", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	require.Equal(t, http.StatusForbidden, resp.Code)

	// Now test ScopeGuardAny allowing org.manage
	router = gin.New()
	router.GET("/any",
		auth.Middleware(),
		ScopeGuardAny("users.manage", "org.manage"),
		func(c *gin.Context) { c.String(http.StatusOK, "ok") },
	)
	anyReq := httptest.NewRequest(http.MethodGet, "/any", nil)
	anyReq.Header.Set("Authorization", "Bearer "+token)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, anyReq)
	require.Equal(t, http.StatusOK, resp.Code)
}

func newTestAuth(t *testing.T, claims map[string]any) (*Auth, string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	given := keyfunc.NewGiven(map[string]keyfunc.GivenKey{
		"test-kid": keyfunc.NewGivenRSA(&key.PublicKey),
	})

	a := &Auth{
		issuer: "https://issuer.example.com",
		jwks:   given,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{})
	token.Header["kid"] = "test-kid"
	for k, v := range claims {
		token.Claims.(jwt.MapClaims)[k] = v
	}
	if _, ok := token.Claims.(jwt.MapClaims)["iss"]; !ok {
		token.Claims.(jwt.MapClaims)["iss"] = a.issuer
	}
	if _, ok := token.Claims.(jwt.MapClaims)["exp"]; !ok {
		token.Claims.(jwt.MapClaims)["exp"] = time.Now().Add(time.Minute).Unix()
	}
	signed, err := token.SignedString(key)
	require.NoError(t, err)

	return a, signed
}
