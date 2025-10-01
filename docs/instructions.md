Go User Authentication Service (Asgardeo) — Copilot Build Script (Aligned with Smart Transit DB)

This document instructs GitHub Copilot (and any developer) to implement the User Authentication Service (UAS) in Go, integrated with Asgardeo for authentication + provisioning (SCIM 2.0).
It is adapted to the Smart Transit database schema (users, bus_owners, bus_staff, lounges, audit_logs, …).

Core idea: Asgardeo handles identity (login UI, tokens, users, roles, SCIM).
The Go service enforces authorization, validates tokens, provisions users via SCIM, and mirrors them into the existing Smart Transit database.

0) Constraints & Tech Choices

Language/Framework: Go 1.23+, Gin, GORM, PostgreSQL.

Database schema: Provided in 004_smart_transit_complete_v2.sql.

Identity provider: Asgardeo (OIDC + SCIM).

Deployment: Docker + Air dev reload.

Principles: Config-driven, security-first, consistent APIs.

1) Dependencies
go get github.com/gin-gonic/gin@v1.10.0
go get gorm.io/gorm@v1.25.12
go get gorm.io/driver/postgres@v1.5.9

# OIDC/JWT
go get github.com/golang-jwt/jwt/v5
go get github.com/MicahParks/keyfunc@v1.9.2

# HTTP utilities
go get github.com/hashicorp/go-retryablehttp@v0.7.7

2) Environment variables
PORT=8080
GIN_MODE=debug

DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=smart_transit_system

ASGARDEO_ISSUER_URL=
ASGARDEO_JWKS_URL=
ASGARDEO_CLIENT_ID=
ASGARDEO_CLIENT_SECRET=
ASGARDEO_SCIM_BASE_URL=
ASGARDEO_SCIM_BEARER=

OIDC_EXPECTED_AUDIENCE=smart-transit-api
OIDC_ACCEPTED_SCOPES=openid,profile,email,user.read,user.write,users.manage,org.manage,org.staff

3) Project structure
internal/
  auth/
    jwks.go
    validator.go
    middleware.go
  scim/
    client.go
    mapper.go
  services/
    user_service.go
    audit_service.go
  handlers/
    me.go
    roles.go
    profiles.go
    admin_users.go
    org.go
  middleware/
    responses.go
    ratelimit.go
internal/models/
    users.go
    audit_logs.go
    bus_owners.go
    bus_staff.go
    lounges.go

4) Data models (GORM)

Use the existing DB schema.

users
type User struct {
  ID             string    `gorm:"primaryKey;type:varchar(128)" json:"id"`
  Email          string    `gorm:"uniqueIndex;not null" json:"email"`
  FirstName      string    `gorm:"size:100;not null" json:"first_name"`
  LastName       string    `gorm:"size:100;not null" json:"last_name"`
  Phone          string    `gorm:"size:20" json:"phone"`
  NIC            string    `gorm:"size:20;unique" json:"nic"`
  Role           string    `gorm:"type:user_role;not null" json:"role"`
  Status         string    `gorm:"type:user_status;default:'active'" json:"status"`
  EmailVerified  bool      `gorm:"default:false" json:"email_verified"`
  FirebaseClaims string    `gorm:"type:jsonb" json:"firebase_custom_claims"`
  CreatedAt      time.Time `json:"created_at"`
  UpdatedAt      time.Time `json:"updated_at"`
}

audit_logs
type AuditLog struct {
  ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
  UserID      *string   `gorm:"type:varchar(128)" json:"user_id"`
  ActionType  string    `gorm:"type:audit_action;not null" json:"action_type"`
  TableName   string    `gorm:"size:100" json:"table_name"`
  RecordID    string    `gorm:"size:36" json:"record_id"`
  OldValues   string    `gorm:"type:jsonb" json:"old_values"`
  NewValues   string    `gorm:"type:jsonb" json:"new_values"`
  IPAddress   string    `gorm:"size:45" json:"ip_address"`
  UserAgent   string    `json:"user_agent"`
  SessionID   string    `json:"session_id"`
  Additional  string    `gorm:"type:jsonb" json:"additional_info"`
  CreatedAt   time.Time `json:"created_at"`
}

Relationships

bus_owners.user_id → maps bus companies to users with role bus_owner.

bus_staff.user_id → maps drivers/conductors to users with role driver or conductor.

lounges.lounge_owner_user_id → maps lounges to users with role lounge_owner.

5) Roles & Scopes

Roles (user_role ENUM):

admin

passenger

bus_owner

driver

conductor

lounge_owner

Scopes (OIDC):

users.manage → Admin

org.manage → Bus/Lounge owners

org.staff → Drivers/Conductors

Self-service (no scopes) → Passengers

6) Enforcement strategy

Admin: users.manage + role=admin

Bus owner: org.manage + role=bus_owner → filter by bus_owners.user_id

Lounge owner: org.manage + role=lounge_owner → filter by lounges.lounge_owner_user_id

Driver/Conductor: org.staff + role=driver/conductor → filter by bus_staff.user_id

Passenger: role=passenger, only self endpoints

7) SCIM → DB Mapping
SCIM Attribute	DB Field / Table
id (sub)	users.id
emails[0].value	users.email
name.givenName	users.first_name
name.familyName	users.last_name
phoneNumbers[0]	users.phone
roles[0].value	users.role
active	users.status
company_id (custom)	bus_owners.user_id
lounge_id (custom)	lounges.lounge_owner_user_id
8) Services

UserService:

FindByID, FindByEmail, CreateFromSCIM, UpdateFromSCIM, Deactivate.

Always mirror SCIM changes into users.

AuditService:

Insert into audit_logs on every action (create/update/delete/login).

9) Handlers & Routes

Example:

GET  /api/v1/me
GET  /api/v1/roles
PUT  /api/v1/profiles/self

# Admin
GET    /api/v1/admin/users
POST   /api/v1/admin/users
PUT    /api/v1/admin/users/:id
DELETE /api/v1/admin/users/:id

# Org (bus_owner, lounge_owner)
GET    /api/v1/org/users
POST   /api/v1/org/users
PUT    /api/v1/org/users/:id
DELETE /api/v1/org/users/:id

10) Summary

No custom tables → use users, audit_logs.

Org scoping → enforced via bus_owners, bus_staff, lounges.

Roles & scopes → align with ENUMs.

SCIM sync → ensures external Asgardeo user → internal users row.

11) Acceptance Criteria

Tokens validated (issuer, aud, exp, scopes).

Users mirrored to users.

Audit trail in audit_logs.

Authorization enforced by role + scope + DB FK scoping.

Endpoints return consistent JSON.