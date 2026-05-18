# SPEC: API Documentation (Swagger / OpenAPI 3.0)

## Context

Task Management RESTful API Backend membutuhkan dokumentasi API yang komprehensif untuk konsumsi frontend developer dan API consumer lainnya. Selama ini belum ada dokumentasi terstruktur — developer harus membaca source code untuk memahami kontrak API.

## Goals

- Generate OpenAPI 3.0 spec dari Go annotations menggunakan `swaggo/swag`
- Serve Swagger UI di endpoint `/swagger/*`
- Document semua endpoint dengan tag, descriptions, schemas, dan security schemes
- Support JWT Bearer authentication di Swagger UI

## Technical Approach

### Stack
- **swaggo/swag** (`github.com/swaggo/swag/cmd/swag`) — CLI untuk generate spec dari annotations
- **swaggo/gin-swagger** (`github.com/swaggo/gin-swagger`) — Gin middleware untuk serve Swagger UI
- **swaggo/files** — embedded static files untuk Swagger UI assets

### File Structure Changes
```
go-task-backend/
├── docs/                    # AUTO-GENERATED (swag init)
│   ├── docs.go             # Generated swagger docs
│   └── swagger.json        # Optional: untuk hosting external
├── main.go                 # + swagger embed & route
├── go.mod                  # + swag dependencies
└── [modules]/[ *_handler.go] # + swagger annotations
```

### Annotations Pattern

 Setiap handler function di-annotate dengan:
```go
// @Summary     Short description
// @Description Detailed description of the endpoint
// @Tags        grouping (e.g., "Auth", "Projects", "Tasks")
// @Accept      json
// @Produce     json
// @Param       body body RequestStruct true "description"
// @Success     200 {object} utils.APIResponse "description"
// @Failure     400,401,403,500 {object} utils.APIResponse "description"
// @Security    BearerAuth
// @Router      /path [method]
```

### Security Scheme
- Type: `http` (Bearer JWT)
- Scheme: `bearer`
- BearerFormat: `JWT`

### Tags Organization
| Tag | Endpoints |
|---|---|
| `Auth` | signup, login |
| `Health` | health, ready |
| `Organizations` | create, invite, members, remove, update role |
| `Invitations` | accept, resend, revoke, list |
| `Projects` | find, create, delete |
| `Tasks` | find by project, search, create, update, delete |
| `Statuses` | find by project, create, update, delete |
| `Labels` | find, create, update, delete |

### Response Schema
```go
// APIResponse represents the standard API response envelope
// @Description Standard API response envelope
type APIResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

## Endpoints to Document

### Public Endpoints (No Auth)
- `GET  /health` — Liveness check
- `GET  /ready` — Readiness probe
- `POST /signup` — User registration
- `POST /login` — Authentication
- `POST /invite/accept` — Accept invitation
- `POST /invite/resend` — Resend invitation
- `DELETE /invite/:token` — Revoke invitation

### Protected Endpoints (JWT Required)
- `GET    /projects` — List projects
- `POST   /projects` — Create project
- `DELETE /projects/:id` — Delete project
- `GET    /projects/:id/tasks` — List tasks by project
- `GET    /tasks/search` — Search tasks
- `POST   /tasks` — Create task
- `PATCH  /tasks/:id` — Update task
- `DELETE /tasks/:id` — Delete task
- `GET    /projects/:id/status` — List statuses
- `POST   /projects/:id/status` — Create status
- `PATCH  /status/:id` — Update status
- `DELETE /status/:id` — Delete status
- `GET    /projects/:id/labels` — List labels
- `POST   /projects/:id/labels` — Create label
- `PATCH  /labels/:id` — Update label
- `DELETE /labels/:id` — Delete label
- `POST   /organizations` — Create organization
- `POST   /organizations/invite` — Invite member
- `GET    /organizations/members` — Get members
- `DELETE /organizations/members/:user_id` — Remove member
- `PATCH  /organizations/members/:user_id` — Update member role
- `GET    /organizations/invitations` — List invitations

### Header Requirements
- Protected endpoints: `Authorization: Bearer <JWT>`
- Organization-scoped endpoints: `X-Organization-ID: <org_id>`

## Acceptance Criteria

1. `swag init` menghasilkan `docs/docs.go` tanpa error
2. `GET /swagger/*` serving Swagger UI works
3. Semua endpoint terdokumentasi dengan baik di Swagger UI
4. JWT Bearer auth berfungsi di Swagger UI (try-out feature)
5. Build `go build -o main .` tetap berhasil setelah perubahan
6. Swagger annotations tidak违反 clean-code rules (self-documenting)

## Effort Estimate

3-4 hari (sesuai backlog M9)