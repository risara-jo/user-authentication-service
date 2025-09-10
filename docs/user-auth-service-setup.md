# User & Authentication Service – Setup Guide

This guide explains how to configure Asgardeo and run the User & Authentication Service locally, then outlines next steps for deployment to Choreo.

## Prerequisites

- Asgardeo tenant and console access
- Go 1.21+
- PostgreSQL (local or Docker)
- cURL or Postman

## 1) Configure Asgardeo (OIDC)

Create one OIDC application per client app (Passenger, Driver/Conductor, Company Owner, Lounge Owner, Admin). Start with one app to test locally.

Steps in Asgardeo Console:
- Applications → New Application → choose app type (Web, SPA, Native for mobile).
- Grant types: Authorization Code + PKCE. Enable Refresh Token for mobile apps.
- Redirect URIs:
  - Web dev: `http://localhost:3000/callback` (adjust to your client)
  - Mobile dev: e.g., `com.passenger.app://callback` (your app scheme)
- Scopes: include at least `openid`, `profile`, `email`.
- Copy the tenant issuer (see Testing below for correct value).

Roles and claims (high level):
- Define roles: `passenger`, `driver`, `conductor`, `company_owner`, `lounge_owner`, `admin`.
- Ensure access tokens include a role/group claim (e.g., `roles` or `groups`).
- Optional custom claims: `org_id`, `company_id`, `lounge_id`.

Service application (later):
- Create one Client Credentials app for server-to-server operations (SCIM/provisioning). Store its credentials as secrets in Choreo.

## 2) Configure the API

Set environment variables (copy `.env.example` to `.env` and fill values):

- `ASGARDEO_ISSUER`: `https://api.asgardeo.io/t/<tenant>/oauth2`
- `ASGARDEO_AUDIENCE`: optional (leave empty unless you enforce audience)
- Database: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`

The service discovers JWKS from `/.well-known/openid-configuration` and validates JWTs.

## 3) Run Locally

Option A: Go directly
```
go run ./cmd/api
```

Option B: Docker Compose (dev)
```
docker-compose -f docker-compose.dev.yml up --build
```

Migrations: GORM auto-migrates minimal tables on startup. A SQL migration exists at `migrations/100_user_auth.sql` for reference.

## 4) Test

1) Health check
```
curl http://localhost:8080/health
```

2) Obtain an access token for your Asgardeo application (via your client app or OAuth tool).

3) Call `/api/v1/me`
```
curl -H "Authorization: Bearer <access_token>" \
  http://localhost:8080/api/v1/me | jq
```
Expected: JSON with `sub`, `email`, `scopes`, `roles`, and raw claims.

If you receive `invalid issuer`, verify `ASGARDEO_ISSUER` matches the `issuer` field from:
```
GET https://api.asgardeo.io/t/<tenant>/oauth2/.well-known/openid-configuration
```

## 5) Roles, Scopes, Authorization

- Roles are attached to users in Asgardeo. Ensure they are included in access tokens (roles/groups claim).
- Add fine-grained scopes (e.g., `user.read`, `user.write`, `users.manage`, `org.manage`) and require them on protected endpoints using the included `RequireScopes` helper.

## 6) Provisioning (Next)

- Add admin endpoints to create/update users and assign roles/org memberships.
- Configure a Client Credentials app in Asgardeo for SCIM and admin operations.
- Implement inbound sync for user updates (polling or webhook) to keep local profile store in sync.

## 7) Choreo Deployment (High Level)

- Build container using `docker/Dockerfile` and publish to a registry.
- Create a Choreo component (HTTP service on port `8080`).
- Configure environment variables/secrets: DB connection, `ASGARDEO_ISSUER`, optional `ASGARDEO_AUDIENCE`.
- Enable observability and set rate limits for registration/admin endpoints.
- Expose `/api/v1/*` via Choreo API; enforce scopes/roles at the gateway if desired.

## Troubleshooting

- 401 invalid token: Check signature, issuer, and token expiry.
- Missing roles in `/me`: Ensure roles/groups are included in access tokens.
- Audience failures: Leave `ASGARDEO_AUDIENCE` empty or set it to the expected value configured in Asgardeo.

## Repository Pointers

- API entry: `cmd/api/main.go`
- JWT middleware: `internal/auth/middleware.go`
- `/me` handler: `internal/handlers/me.go`
- Models: `internal/models/user.go`
- Config: `internal/config/config.go`
- Example env: `.env.example`
- SQL baseline: `migrations/100_user_auth.sql`

