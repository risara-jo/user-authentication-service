package handlers

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"smart-transit-system/internal/httpx"
)

// AuthLogin returns an authorize URL template for SPA PKCE login.
// Note: SPA must generate code_challenge and state client-side.
func AuthLogin(c *gin.Context) {
	baseIssuer := strings.TrimRight(os.Getenv("ASGARDEO_ISSUER"), "/")
	clientID := os.Getenv("ASGARDEO_CLIENT_ID")
	redirectURI := os.Getenv("ASGARDEO_REDIRECT_URI")

	if baseIssuer == "" || clientID == "" || redirectURI == "" {
		httpx.RespondError(c, http.StatusBadRequest, "auth.misconfigured", "missing ASGARDEO_ISSUER, ASGARDEO_CLIENT_ID, or ASGARDEO_REDIRECT_URI", nil)
		return
	}

	authEndpoint := baseIssuer + "/authorize"
	// Build a template authorize URL (without PKCE values)
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", "openid profile email")
	// SPA should add: state, code_challenge, code_challenge_method=S256

	// Optional inputs for convenience/testing
	state := c.Query("state")
	codeChallenge := c.Query("code_challenge")
	codeMethod := c.DefaultQuery("code_challenge_method", "S256")

	fullURL := ""
	if state != "" && codeChallenge != "" {
		qp := url.Values{}
		for k, v := range q {
			qp[k] = v
		}
		qp.Set("state", state)
		qp.Set("code_challenge", codeChallenge)
		qp.Set("code_challenge_method", codeMethod)
		fullURL = authEndpoint + "?" + qp.Encode()
	}

	httpx.RespondSuccess(c, http.StatusOK, gin.H{
		"authorize_endpoint":          authEndpoint,
		"base_params":                 q.Encode(),
		"full_url_if_params_provided": fullURL,
		"notes":                       "SPA must add 'state', 'code_challenge', and 'code_challenge_method=S256' before redirecting.",
	}, "login metadata fetched")
}

// AuthAuthorize redirects to Asgardeo authorize endpoint when provided with PKCE params.
func AuthAuthorize(c *gin.Context) {
	baseIssuer := strings.TrimRight(os.Getenv("ASGARDEO_ISSUER"), "/")
	clientID := os.Getenv("ASGARDEO_CLIENT_ID")
	redirectURI := os.Getenv("ASGARDEO_REDIRECT_URI")
	if baseIssuer == "" || clientID == "" || redirectURI == "" {
		httpx.RespondError(c, http.StatusBadRequest, "auth.misconfigured", "missing ASGARDEO_ISSUER/CLIENT_ID/REDIRECT_URI", nil)
		return
	}
	state := c.Query("state")
	codeChallenge := c.Query("code_challenge")
	method := c.DefaultQuery("code_challenge_method", "S256")
	if state == "" || codeChallenge == "" {
		httpx.RespondError(c, http.StatusBadRequest, "auth.missing_pkce", "state and code_challenge are required", nil)
		return
	}
	authEndpoint := baseIssuer + "/authorize"
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", "openid profile email")
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", method)

	c.Redirect(http.StatusFound, authEndpoint+"?"+q.Encode())
}
