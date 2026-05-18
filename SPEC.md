# SPEC.md — Phase 1: Quick Wins (Production-Grade Foundation)

**Project:** go-task-backend → Collaborative Kanban Board
**Phase:** Phase 1 — Quick Wins
**Target:** Production-grade, siap di-scale, observable, secure
**Last Updated:** 2026-05-18
**Status:** Draft — **PENDING CEO APPROVAL**

---

## Overview

Phase 1 Quick Wins adalah fondasi production-grade. Semua task di bawah ini berdampak tinggi (High Impact) tapi effort rendah (Low Effort, total ~35-40 jam). Fokusnya: **reliability, observability, security, dan basic team management**.

> ⚠️ **Constraint:** Setiap feature wajib mengikuti arsitektur `Handler → Service → Repository`. Semua endpoint baru harus menggunakan `utils.SendSuccess` / `utils.SendError`. Semua route baru masuk middleware `RequireAuth` (kecuali health check).

---

## Q1: Structured Logging (slog) + Request ID

### 1. Problem Statement
Backend currently uses `log.Println` yang tidak terstruktur. Di production, tidak bisa trace request spesifik atau aggregate error. Semua log campur aduk tanpa correlation.

### 2. Goals / Success Criteria
- [ ] Semua endpoint logging pakai structured JSON format
- [ ] Setiap request punya unique `request_id` (UUID) di headers & log
- [ ] Log fields: `timestamp`, `level`, `request_id`, `method`, `path`, `user_id`, `org_id`, `duration_ms`, `status_code`, `error`
- [ ] Request ID di-return di response header `X-Request-ID`
- [ ] Sensitive data (password, token) TIDAK pernah di-log

### 3. Technical Approach
- Pakai `log/slog` (Go 1.21+ built-in) dengan JSON handler
- Middleware `RequestIDMiddleware` inject `request_id` ke `gin.Context`
- Middleware `StructuredLoggerMiddleware` log request start & end
- Buat package `utils/logger.go` — singleton logger

### 4. File Changes
- **New:** `utils/logger.go` — slog setup
- **New:** `middlewares/request_id.go` — inject UUID request ID
- **New:** `middlewares/structured_logger.go` — structured request logging
- **Modify:** `main.go` — register middleware, remove `gin.Default()` noise

### 5. API Changes
```
# No new API endpoints — ini internal infrastructure
```

### 6. Definition of Done
- [ ] `go build` passes
- [ ] Log output is valid JSON
- [ ] Request ID muncul di setiap log line yang related
- [ ] Password/token TIDAK muncul di log

---

## Q2: Health Check Endpoints

### 1. Problem Statement
Tidak ada endpoint untuk load balancer / orchestration tool (Docker Healthcheck, Kubernetes liveness/readiness probe). App akan dianggap "down" padahal masih jalan.

### 2. Goals / Success Criteria
- [ ] `GET /health` → returns `200 OK` dengan `{"status": "ok"}` — untuk basic liveness check
- [ ] `GET /ready` → returns `200 OK` jika DB connected, `503` jika DB down — untuk readiness probe
- [ ] Endpoint ini TIDAK require JWT auth (public)
- [ ] Response time < 50ms

### 3. Technical Approach
- Buat `handlers/health.go` atau inline di `main.go`
- `GET /health` → selalu 200 jika process alive
- `GET /ready` → ping DB (`config.DB.Raw("SELECT 1")`)
- Register SEBELUM middleware auth

### 4. File Changes
- **New:** `handlers/health.go` — health & readiness handlers
- **Modify:** `main.go` — register health routes (public, before auth middleware)

### 5. API Changes
```
GET /health        → 200 { "status": "ok" }
GET /ready         → 200 { "status": "ready", "db": "connected" }
                   → 503 { "status": "not ready", "db": "disconnected" }
```

### 6. Definition of Done
- [ ] `go build` passes
- [ ] `curl localhost:8080/health` returns 200
- [ ] `curl localhost:8080/ready` returns 200 (DB up) atau 503 (DB down)
- [ ] No auth required for both endpoints

---

## Q3: RBAC — Role & Permission Scopes

### 1. Problem Statement
Saat ini semua user dalam 1 organization dianggap setara. Tidak ada perbedaan antara Owner (pembuat org), Admin, dan Member. Semua bisa delete project, invite member, dll — risk keamanan dan abuse.

### 2. Goals / Success Criteria
- [ ] Enum role: `owner`, `admin`, `member`
- [ ] Tabel `organization_users` punya kolom `role` (default: `member`)
- [ ] Setiap action punya permission scope:

| Action | owner | admin | member |
|---|---|---|---|
| Delete organization | ✅ | ❌ | ❌ |
| Invite member | ✅ | ✅ | ❌ |
| Remove member | ✅ | ✅ | ❌ |
| Update member role | ✅ | ✅ | ❌ |
| Delete project | ✅ | ✅ | ❌ |
| Create project | ✅ | ✅ | ✅ |
| Create/update/delete task | ✅ | ✅ | ✅ |
| View project | ✅ | ✅ | ✅ |

- [ ] Middleware check role BEFORE eksekusi action
- [ ] Owner CANNOT be removed or demoted (self-protect)

### 3. Technical Approach
- **New Model:** `models/role.go` — enum string constants
- **Modify:** `organization_models.go` — tambah kolom `role` + enum
- **New:** `services/permission.go` — `CheckPermission(userID, orgID, action)` → bool
- **Middleware approach:** Di service layer, cek permission sebelum operasi write
- **Database migration:** Add `role` column to `organization_users` with default `member`

### 4. File Changes
- **New:** `models/role.go` — Role enum
- **Modify:** `modules/organizations/organization_models.go` — add role field
- **Modify:** `modules/organizations/organization_repository.go` — update member query
- **Modify:** `modules/organizations/organization_service.go` — permission check di invite/remove
- **Modify:** `modules/projects/project_service.go` — permission check di delete
- **Modify:** `config/db.go` — golang-migrate setup (see Q10)

### 5. Database Changes
```sql
ALTER TABLE organization_users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'member';

INSERT INTO roles (name) VALUES ('owner'), ('admin'), ('member'); -- if roles table exists
-- or just use string enum directly in application code
```

### 6. API Changes
```
# No new endpoints — ini validasi tambahan di existing endpoints
# Error response jika tidak punya permission:
→ 403 { "success": false, "message": "Insufficient permission", "data": null }
```

### 7. Definition of Done
- [ ] `go build` passes
- [ ] Owner tidak bisa di-kick dari org
- [ ] Member tidak bisa invite/remove orang
- [ ] Admin tidak bisa delete org
- [ ] Service test: happy path + unauthorized path

---

## Q4: Email Invite Flow with Token Expiry + Resend

### 1. Problem Statement
Fitur `POST /organizations/invite` yang sekarang hanya buat user langsung join tanpa token, email, atau expiry. Freelancer perlu bisa invite client/colleague via email link yang expire.

### 2. Goals / Success Criteria
- [ ] `POST /organizations/invite` generate invite token (UUID), simpan ke tabel `organization_invitations`
- [ ] Invitation record: `org_id`, `invited_email`, `token`, `role`, `expires_at` (default 7 days), `created_by`
- [ ] Generate invite URL: `${APP_URL}/invite/${token}`
- [ ] Kirim email via SMTP (body text sederhana + link)
- [ ] `POST /invite/accept` — accept invite dengan token → auto-join org
- [ ] `GET /organizations/invitations` — list pending invitations (owner + admin only)
- [ ] `POST /invite/resend` — resend email invite (reset expiry, regenerate token)
- [ ] `DELETE /invite/:token` — cancel/revoke invitation
- [ ] Token expired → return 410 Gone
- [ ] Token already used → return 409 Conflict
- [ ] Email already in org → return 400 Bad Request

### 3. Technical Approach
- **New Model:** `modules/organizations/invitation_model.go` — `OrganizationInvitation`
- **New Repository:** `modules/organizations/invitation_repository.go`
- **New Service:** `modules/organizations/invitation_service.go`
- **Email:** Pakai `gopkg.in/gomail.v2` atau `github.com/go-mail/mail` — SMTP config dari env
- **Token:** UUID v4, stored as string
- **Expiry:** `time.Now().Add(7 * 24 * time.Hour)`
- **DI di main.go:** Register invitation handler + service

### 4. Environment Variables (NEW)
```
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your@email.com
SMTP_PASSWORD=your_app_password
APP_URL=https://yourdomain.com
INVITE_EXPIRY_HOURS=168  # 7 days
```

### 5. File Changes
- **New:** `modules/organizations/invitation_model.go`
- **New:** `modules/organizations/invitation_repository.go`
- **New:** `modules/organizations/invitation_service.go`
- **New:** `modules/organizations/invitation_handler.go`
- **New:** `utils/email.go` — SMTP sender
- **Modify:** `modules/organizations/organization_models.go` — update invite flow
- **Modify:** `main.go` — register invitation module, update invite route
- **Modify:** `.env.example` — add SMTP vars

### 6. API Changes
```
POST /organizations/invite
Body: { "email": "client@example.com", "role": "member" }
→ 201 { "success": true, "message": "Invitation sent", "data": { "invitation_url": "..." } }

GET /organizations/invitations
Header: X-Organization-ID required
→ 200 { "success": true, "data": [ { "id", "email", "role", "expires_at", "status" } ] }

POST /invite/accept
Body: { "token": "uuid-v4-token" }
→ 200 { "success": true, "message": "Joined organization successfully" }
→ 410 { "success": false, "message": "Invitation expired" }
→ 409 { "success": false, "message": "Invitation already used" }

POST /invite/resend
Body: { "invitation_id": 123 }
→ 200 { "success": true, "message": "Invitation resent" }

DELETE /invite/:token
→ 200 { "success": true, "message": "Invitation revoked" }
```

### 7. Definition of Done
- [ ] `go build` passes
- [ ] Invite → email terkirim dengan link
- [ ] Click link → user bisa join org
- [ ] Expired token → proper error message
- [ ] Resend → new token generated, old one invalidated
- [ ] Only owner/admin bisa see/list/cancel invitations

---

## Q5: Remove / Update Member from Organization

### 1. Problem Statement
Tidak ada cara untuk remove member atau update role member. Freelancer yang collaborate dengan client temporer perlu bisa remove akses kapan saja.

### 2. Goals / Success Criteria
- [ ] `DELETE /organizations/members/:user_id` — remove member dari org
- [ ] `PATCH /organizations/members/:user_id` — update member role
- [ ] Owner TIDAK bisa di-remove atau di-demote
- [ ] Admin TIDAK bisa demote owner atau promote someone to owner
- [ ] Member TIDAK bisa remove/promote anyone
- [ ] User tidak bisa remove themselves (owner protection)

### 3. Technical Approach
- Extend `modules/organizations/organization_handler.go` — add new handlers
- Extend `modules/organizations/organization_service.go` — permission check
- Extend `modules/organizations/organization_repository.go` — update member role, delete member
- Pakai existing middleware auth + RBAC dari Q3

### 4. File Changes
- **Modify:** `modules/organizations/organization_handler.go` — add RemoveMember, UpdateMemberRole
- **Modify:** `modules/organizations/organization_service.go` — add RemoveMember, UpdateMemberRole methods
- **Modify:** `modules/organizations/organization_repository.go` — update query
- **Modify:** `main.go` — register routes

### 5. API Changes
```
DELETE /organizations/members/:user_id
Header: X-Organization-ID required
→ 200 { "success": true, "message": "Member removed" }
→ 403 { "success": false, "message": "Cannot remove owner" }
→ 403 { "success": false, "message": "Insufficient permission" }

PATCH /organizations/members/:user_id
Body: { "role": "admin" | "member" }
→ 200 { "success": true, "message": "Role updated", "data": { "user_id", "role" } }
→ 403 { "success": false, "message": "Cannot change owner role" }
```

### 6. Definition of Done
- [ ] `go build` passes
- [ ] Admin bisa remove member lain
- [ ] Owner tidak bisa di-remove
- [ ] Member tidak bisa akses endpoint ini

---

## Q6: Full-Text Task Search

### 1. Problem Statement
Freelancer dengan 50+ tasks tidak bisa cari task spesifik. Harus scroll semua project satu per satu.

### 2. Goals / Success Criteria
- [ ] `GET /tasks/search` — search tasks across current organization
- [ ] Search by: `title` (partial match), `description` (partial match)
- [ ] Filter by: `project_id`, `assignee_id`, `status_id`, `priority_id`
- [ ] Sort by: `created_at`, `updated_at`, `title`
- [ ] Pagination: `page`, `limit` (max 100)
- [ ] Response time < 200ms untuk 10k tasks (pakai PostgreSQL full-text index)
- [ ] HANYA return task dari org user saat ini (security)

### 3. Technical Approach
- Pakai PostgreSQL `ILIKE` untuk simple search ATAU `tsvector` untuk advanced full-text
- Add composite index: `(organization_id, title)` untuk speed
- **New Repository:** `modules/tasks/task_search_repository.go` — dedicated search method
- **Extend Service:** `modules/tasks/task_service.go` — add SearchTasks method
- **Extend Handler:** `modules/tasks/task_handler.go` — add SearchTasks handler
- Route: `GET /tasks/search`

### 4. File Changes
- **Modify:** `modules/tasks/task_repository.go` — add search query
- **Modify:** `modules/tasks/task_service.go` — add SearchTasks
- **Modify:** `modules/tasks/task_handler.go` — add SearchTasks handler
- **Modify:** `main.go` — register `GET /tasks/search`
- **Database:** Add composite index via migration (Q10)

### 5. API Changes
```
GET /tasks/search?q=design&project_id=1&assignee_id=2&status_id=3&priority_id=1&sort=created_at&order=desc&page=1&limit=20
Header: X-Organization-ID required
→ 200 {
  "success": true,
  "data": {
    "tasks": [...],
    "meta": { "current_page": 1, "limit": 20, "total_data": 5, "total_pages": 1 }
  }
}
```

### 6. Definition of Done
- [ ] `go build` passes
- [ ] Search "des" returns tasks dengan "design", "design task"
- [ ] Search scoped ke organization (tidak bisa see other org tasks)
- [ ] Pagination works correctly

---

## Q7: Task Filter — Assignee, Status, Priority, Date Range

### 1. Problem Statement
Endpoint `GET /projects/:id/tasks` saat ini hanya support pagination. Tidak bisa filter berdasarkan assignee, status, priority, atau date range.

### 2. Goals / Success Criteria
- [ ] Extend existing `GET /projects/:id/tasks` dengan query params:
  - `?assignee_id=1` — filter by assignee (single)
  - `?status_id=1` — filter by status
  - `?priority_id=1` — filter by priority
  - `?due_from=2026-05-01` — filter due date >= from
  - `?due_to=2026-05-31` — filter due date <= to
  - `?created_from=...`, `?created_to=...` — filter by created date
  - `?label_ids=1,2,3` — filter by labels (COMING WITH Q8 — prepare column)
- [ ] Filters bisa dikombinasikan (AND logic)
- [ ] Empty filter → return all (maintain backward compatibility)
- [ ] HANYA return task dari org user saat ini

### 3. Technical Approach
- Modify `modules/tasks/task_repository.go` — add filter params ke `GetTasksByProject`
- Modify `modules/tasks/task_service.go` — pass filter params
- Modify `modules/tasks/task_handler.go` — extract query params
- Add index di `end_date` column for date range queries

### 4. File Changes
- **Modify:** `modules/tasks/task_repository.go` — add dynamic WHERE clause
- **Modify:** `modules/tasks/task_service.go` — add FilterInput struct
- **Modify:** `modules/tasks/task_handler.go` — parse query params
- **Database:** Add index on `end_date`, `start_date`

### 5. API Changes
```
GET /projects/1/tasks?assignee_id=2&status_id=3&due_from=2026-05-01&due_to=2026-05-31&page=1&limit=50
→ 200 {
  "success": true,
  "data": {
    "tasks": [...],
    "meta": { ... }
  }
}
```

### 6. Definition of Done
- [ ] `go build` passes
- [ ] Filter assignee_id works
- [ ] Filter status_id works
- [ ] Filter date range works
- [ ] Combined filters work (AND logic)
- [ ] No filter → returns all (backward compatible)

---

## Q8: Labels/Tags System per Organization

### 1. Problem Statement
Tidak ada cara untuk categorize task selain status dan priority. Freelancer perlu tag task dengan label kustom (e.g., "Design", "Development", "Client Review", "Urgent", "Billable").

### 2. Goals / Success Criteria
- [ ] `POST /projects/:id/labels` — create label (name, color)
- [ ] `GET /projects/:id/labels` — list all labels in project
- [ ] `PATCH /labels/:id` — update label name/color
- [ ] `DELETE /labels/:id` — delete label
- [ ] `POST /tasks` — bisa assign label_ids saat create
- [ ] `PATCH /tasks/:id` — bisa update label_ids
- [ ] Label belongs to PROJECT (bukan global)
- [ ] Label punya: `id`, `project_id`, `name`, `color` (hex), `created_at`
- [ ] Task-Label relationship: Many-to-Many via `task_labels` table
- [ ] Admin/Owner of org bisa CRUD label

### 3. Technical Approach
- **New Model:** `modules/tasks/label_model.go` — `Label`, `TaskLabel` (join table)
- **New Repository:** `modules/tasks/label_repository.go`
- **New Service:** `modules/tasks/label_service.go`
- **New Handler:** `modules/tasks/label_handler.go`
- Many-to-Many: GORM `many2many:task_labels`
- Update Task Create/Update service untuk handle label_ids

### 4. File Changes
- **New:** `modules/tasks/label_model.go`
- **New:** `modules/tasks/label_repository.go`
- **New:** `modules/tasks/label_service.go`
- **New:** `modules/tasks/label_handler.go`
- **Modify:** `modules/tasks/task_models.go` — add Labels relation
- **Modify:** `modules/tasks/task_service.go` — handle label_ids in create/update
- **Modify:** `modules/tasks/task_handler.go` — add label_ids in request body
- **Modify:** `main.go` — register label routes

### 5. Database Changes
```sql
CREATE TABLE labels (
    id SERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) NOT NULL DEFAULT '#6366F1',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE task_labels (
    task_id BIGINT REFERENCES tasks(id) ON DELETE CASCADE,
    label_id BIGINT REFERENCES labels(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, label_id)
);

CREATE INDEX idx_labels_project_id ON labels(project_id);
```

### 6. API Changes
```
# Labels
POST /projects/:id/labels        → 201 { name, color }
GET  /projects/:id/labels       → 200 [ { id, name, color } ]
PATCH /labels/:id               → 200 { id, name, color }
DELETE /labels/:id              → 200

# Task with labels
POST /tasks                     → Body: { ..., "label_ids": [1, 2] }
PATCH /tasks/:id                → Body: { ..., "label_ids": [1, 3] }
GET /projects/:id/tasks         → Response: tasks include labels: [ { id, name, color } ]
```

### 7. Definition of Done
- [ ] `go build` passes
- [ ] CRUD label works
- [ ] Task bisa di-assign multiple labels
- [ ] Label filter (Q7) ready to use (Q7 already prepare column)

---

## Q9: Optimistic Locking (Version Column)

### 1. Problem Statement
Race condition: 2 users edit task bersamaan. User A save, User B save — User B's save OVERWRITES User A's changes tanpa ada warning.

### 2. Goals / Success Criteria
- [ ] Tabel `tasks` punya kolom `version` (integer, default 1, auto-increment on update)
- [ ] `PATCH /tasks/:id` BODY wajib terima `expected_version` (uint)
- [ ] Service: `UPDATE tasks SET ... WHERE id = ? AND version = ?`
- [ ] Jika `version` mismatch → return `409 Conflict` dengan message: "Task was modified by another user. Please refresh and try again."
- [ ] Response include `current_version` agar client bisa retry dengan version baru
- [ ] `POST /tasks` → initial version = 1
- [ ] Optimistic locking juga untuk: Projects, Labels, Statuses

### 3. Technical Approach
- Add `version` column to tasks, projects, statuses tables
- Modify repository `UpdateTask` — add `expectedVersion` param, check affected rows
- Extend `UpdateTaskInput` struct di service — add `ExpectedVersion`
- Extend request body di handler — parse `expected_version`
- Return `409` if row not updated (version mismatch)

### 4. File Changes
- **Modify:** `modules/tasks/task_models.go` — add `Version` field
- **Modify:** `modules/tasks/task_repository.go` — add version check in UPDATE
- **Modify:** `modules/tasks/task_service.go` — add version to UpdateTaskInput
- **Modify:** `modules/tasks/task_handler.go` — parse expected_version
- **Modify:** `modules/projects/project_models.go` — add Version field
- **Modify:** `modules/projects/project_repository.go` — version check
- **Modify:** `modules/tasks/status_models.go` — add Version field
- **Modify:** `modules/tasks/status_repository.go` — version check

### 5. API Changes
```
PATCH /tasks/:id
Body: {
  "title": "New title",
  "expected_version": 3
}
→ 200 { "success": true, "data": { ...task with version: 4 } }
→ 409 { "success": false, "message": "Task was modified by another user", "data": { "current_version": 5 } }
```

### 6. Definition of Done
- [ ] `go build` passes
- [ ] Version increment on every update
- [ ] 409 returned when expected_version mismatch
- [ ] Client bisa recover dari 409 dan retry

---

## Q10: Database Migration (golang-migrate)

### 1. Problem Statement
Saat ini pakai `config.DB.AutoMigrate()` — ini DEV ONLY. Di production, auto-migrate bisa cause data loss atau unexpected schema changes. Butuh versioned, reversible migrations.

### 2. Goals / Success Criteria
- [ ] Setup `github.com/golang-migrate/migrate/v4` dengan PostgreSQL driver
- [ ] Buat folder `migrations/` dengan timestamp-based versioned SQL files
- [ ] Migration files:
  - `000001_init_schema.up.sql` — initial tables (users, organizations, projects, tasks, statuses, priorities, organization_users)
  - `000001_init_schema.down.sql` — rollback
  - `000002_add_organization_roles.up.sql` — role column (Q3)
  - `000002_add_organization_roles.down.sql` — rollback
  - `000003_add_task_version.up.sql` — version column (Q9)
  - `000004_add_labels.up.sql` — labels table (Q8)
  - `000005_add_invitations.up.sql` — invitation table (Q4)
- [ ] Hapus `config/db.go` AutoMigrate calls
- [ ] Setup migration di startup (`config/db.go` — migrate up on startup)
- [ ] Migration di-run SEBELUM app start (atau on app start via migrate package)
- [ ] Add seed data di migration atau separate seeder

### 3. Technical Approach
```bash
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/postgres
go get github.com/golang-migrate/migrate/v4/source/file
```

- `config/db.go` — call `migrate.Up()` on startup
- Fallback: jika migration fails, log error dan EXIT (don't start with bad schema)
- Buat manual SQL migration files (NOT use GORM AutoMigrate output — manual karena kita mau control exact schema)

### 4. File Changes
- **New:** `migrations/000001_init_schema.up.sql`
- **New:** `migrations/000001_init_schema.down.sql`
- **New:** `migrations/000002_add_organization_roles.up.sql`
- **New:** `migrations/000002_add_organization_roles.down.sql`
- **New:** `migrations/000003_add_task_version.up.sql`
- **New:** `migrations/000003_add_task_version.down.sql`
- **New:** `migrations/000004_add_labels.up.sql`
- **New:** `migrations/000004_add_labels.down.sql`
- **New:** `migrations/000005_add_invitations.up.sql`
- **New:** `migrations/000005_add_invitations.down.sql`
- **Modify:** `config/db.go` — replace AutoMigrate dengan golang-migrate
- **Modify:** `go.mod` — add migrate dependencies

### 5. Database Schema (Migration 000001)

```sql
-- Users
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    avatar_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Organizations
CREATE TABLE organizations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Organization Users (join table)
CREATE TABLE organization_users (
    organization_id BIGINT REFERENCES organizations(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'member',  -- owner, admin, member
    joined_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (organization_id, user_id)
);

-- Projects
CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    organization_id BIGINT REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_projects_org_id ON projects(organization_id);

-- Statuses
CREATE TABLE statuses (
    id SERIAL PRIMARY KEY,
    project_id BIGINT REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) DEFAULT '#6366F1',
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_statuses_project_id ON statuses(project_id);

-- Priorities (seeded, static)
CREATE TABLE priorities (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7) NOT NULL,
    sort_order INT NOT NULL
);

-- Tasks
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    project_id BIGINT REFERENCES projects(id) ON DELETE CASCADE,
    status_id BIGINT REFERENCES statuses(id) ON DELETE SET NULL,
    priority_id BIGINT REFERENCES priorities(id) ON DELETE SET NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    start_date DATE,
    end_date DATE,
    version INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_tasks_status_id ON tasks(status_id);
CREATE INDEX idx_tasks_assignee_id ON task_users(user_id); -- after task_users table
CREATE INDEX idx_tasks_end_date ON tasks(end_date);
CREATE INDEX idx_tasks_title_search ON tasks USING gin(to_tsvector('simple', title));

-- Task Users (assignee)
CREATE TABLE task_users (
    task_id BIGINT REFERENCES tasks(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, user_id)
);

CREATE INDEX idx_task_users_user_id ON task_users(user_id);
```

### 6. Definition of Done
- [ ] `go build` passes
- [ ] `migrate -up` runs tanpa error
- [ ] `migrate -down` rollback correctly
- [ ] App auto-run migration on startup
- [ ] No `AutoMigrate` call remains in codebase

---

## Q11: Graceful Shutdown

### 1. Problem Statement
App saat ini pakai `r.Run(":8080")` yang tidak handle SIGTERM. Di Docker/Kubernetes, container di-terminate tapi request in-flight gagal tanpa graceful response.

### 2. Goals / Success Criteria
- [ ] Ganti `r.Run()` dengan manual HTTP server dengan `graceful shutdown`
- [ ] Listen SIGINT (Ctrl+C) dan SIGTERM
- [ ] On signal: stop accept new requests, wait for in-flight requests (max 30 detik)
- [ ] Log shutdown event: `"Shutting down server..."` dan `"Server exited gracefully"`
- [ ] Flush structured logs (slog) before exit

### 3. Technical Approach
```go
srv := &http.Server{Addr: ":8080", Handler: r}

go func() {
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        slog.Error("Server error", "error", err)
    }
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

slog.Info("Shutting down server...")
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := srv.Shutdown(ctx); err != nil {
    slog.Error("Server forced to shutdown", "error", err)
}
slog.Info("Server exited")
```

### 4. File Changes
- **Modify:** `main.go` — replace `r.Run()` dengan graceful shutdown pattern

### 5. Definition of Done
- [ ] `go build` passes
- [ ] `Ctrl+C` triggers graceful shutdown
- [ ] In-flight requests complete before exit
- [ ] Log shows shutdown sequence

---

## Q12: Rate Limiting (Per IP + Per User)

### 1. Problem Statement
Tidak ada rate limiting. Attacker atau runaway script bisa DoS server dengan unlimited requests. Tidak ada proteksi untuk tier Free dari abuse.

### 2. Goals / Success Criteria
- [ ] Rate limit by IP (unauthenticated): `100 requests/minute` per IP
- [ ] Rate limit by User (authenticated): `500 requests/minute` per user
- [ ] Menggunakan in-memory sliding window atau token bucket
- [ ] Response header: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`
- [ ] Exceed limit → `429 Too Many Requests` dengan `Retry-After` header
- [ ] Health check endpoints EXCLUDED dari rate limiting
- [ ] Library suggestion: `golang.org/x/time/rate` atau `github.com/ulule/limiter`

### 3. Technical Approach
- **New Middleware:** `middlewares/rate_limiter.go`
- Per-IP: extract from `c.ClientIP()`
- Per-User: extract user_id dari JWT (setelah RequireAuth middleware)
- Use sync.Map atau IP-based map untuk store request count
- For production scaling: recommend Redis-backed rate limiter (but keep in-memory as v1)

### 4. File Changes
- **New:** `middlewares/rate_limiter.go`
- **Modify:** `main.go` — register rate limiter middleware AFTER RequireAuth
- **Modify:** `go.mod` — add rate limiter library

### 5. API Changes
```
# Response headers on every response:
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 97
X-RateLimit-Reset: 1747569600

# When exceeded:
HTTP/1.1 429 Too Many Requests
Retry-After: 60
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1747569600
Body: { "success": false, "message": "Rate limit exceeded. Try again in 60 seconds." }
```

### 6. Definition of Done
- [ ] `go build` passes
- [ ] Burst of 150 requests in 1 second → 429 on requests after 100
- [ ] Headers present on every response
- [ ] Health endpoints unaffected

---

## Q13: CI/CD Pipeline (GitHub Actions)

### 1. Problem Statement
Tidak ada automated testing, linting, atau build pipeline. Setiap release perlu manual build dan test, risk human error dan inkonsistensi.

### 2. Goals / Success Criteria
- [ ] **Lint:** `golangci-lint run` — catch common issues
- [ ] **Test:** `go test ./... -v -race` — run all tests dengan race detector
- [ ] **Build:** `go build -o main .` — ensure compile
- [ ] **Docker:** `docker build .` dan `docker-compose up -d`
- [ ] **Triggers:** On every `push` ke branch `main` dan on every `PR`
- [ ] **Secrets:** `DOCKER_HUB_TOKEN`, `SSH_DEPLOY_KEY` injected via GitHub Secrets
- [ ] **On PR:** Comment hasil test + lint di PR
- [ ] **On merge to main:** Build + push Docker image + deploy to staging (optional)

### 3. Technical Approach
- **New File:** `.github/workflows/ci.yml` — lint + test
- **New File:** `.github/workflows/build-push.yml` — build + push image
- **Dockerfile:** Already exists — verify multi-stage build

### 4. File Changes
- **New:** `.github/workflows/ci.yml`
- **New:** `.github/workflows/build-push.yml` (optional — for future deployment)
- **New:** `.golangci.yml` — golangci-lint configuration
- **Modify:** `Dockerfile` — verify multi-stage, add LDFLAGS untuk version

### 5. Pipeline Stages (ci.yml)
```yaml
name: CI
on: [push, pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: go-version: '1.24.0'
      - run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
      - run: golangci-lint run ./...

  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: test
          POSTGRES_DB: gotaskdb_test
        ports: ['5432:5432']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: go-version: '1.24.0'
      - run: go test ./... -v -race -coverprofile=coverage.out
      - uses: codecov/codecov-action@v4

  build:
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: go-version: '1.24.0'
      - run: go build -ldflags "-X main.version=${{ github.sha }}" -o main .
      - run: docker build -t gotask-backend:${{ github.sha }} .
```

### 6. Definition of Done
- [ ] `golangci-lint run` passes locally
- [ ] `go test ./...` passes
- [ ] GitHub Actions runs on PR
- [ ] Docker image builds successfully
- [ ] Coverage report generated

---

## ⚠️ Implementation Order (Quick Wins)

Urutan eksekusi sudah dioptimasi agar dependencies respected:

```
PRE-REQ CHAIN:
Q10 (Migrations) ──→ Q3 (RBAC) ──→ Q4 (Invite Flow)
                              └──→ Q5 (Remove Member)
Q10 ──→ Q9 (Version) ──→ Q6 (Search) ──→ Q7 (Filter)
Q10 ──→ Q8 (Labels)
Q1  (Logger) selalu pertama
Q2  (Health) selalu pertama
Q11 (Graceful) bisa kapan saja
Q12 (Rate Limit) sebelum Q4/Invite Flow
Q13 (CI/CD) selalu terakhir
```

### Recommended Order:
```
Week 1, Day 1-2:  Q1 (Logger) + Q2 (Health) + Q10 (Migrations) [Foundation]
Week 1, Day 3-4:  Q3 (RBAC) + Q9 (Version Locking) [Security + Integrity]
Week 2, Day 1-2:  Q5 (Remove Member) + Q4 (Invite Flow) [Team Management]
Week 2, Day 3:    Q8 (Labels) [Categorization]
Week 2, Day 4-5:  Q6 (Search) + Q7 (Filter) [Discovery]
Week 3, Day 1:    Q11 (Graceful Shutdown) + Q12 (Rate Limiting) [Reliability]
Week 3, Day 2-3:  Q13 (CI/CD) [Automation]
```

---

## ✅ Acceptance Criteria per Task

Setiap task QUICK WIN WAJIB memenuhi:
1. ✅ `go build -o main .` passes tanpa error
2. ✅ Unit test / integration test covers happy path + error path
3. ✅ Route terdaftar di `main.go` dengan middleware yang sesuai
4. ✅ Response format konsisten: `{ "success": bool, "message": string, "data": any }`
5. ✅ Error cases return appropriate HTTP status code
6. ✅ Security: no SQL injection, no hardcoded secrets, no data leak antar org
7. ✅ Backward compatible dengan existing API contracts

---

**Ready for CEO Approval → `/project:dev`**
