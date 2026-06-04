# TECHNICAL DOCUMENTATION — Backend API
## Freelance OS Backend — Developer Guide

> Dokumen ini untuk developer/engineer yang mau memahami cara kerja backend secara teknis.

---

## 🔧 Tech Stack

| Komponen | Teknologi |
|---|---|
| Language | Go 1.24.0 |
| Framework | Gin v1.11.0 |
| Database | PostgreSQL 15 |
| ORM | GORM v1.31.1 |
| Auth | JWT (golang-jwt/jwt/v5) |
| Migration | golang-migrate |
| Container | Docker & Docker Compose |

---

## 🏗️ Architecture

### Pattern: Modular Monolith

```
modules/
├── auth/              # Authentication (signup, login, token, password reset)
├── workspaces/         # Workspace management + tier plans + invitations
├── projects/           # Project CRUD + client link
├── tasks/              # Task, status, priority, labels
├── clients/            # Client management + stats
├── invoices/           # Invoice + billing + revenue sync
└── health/             # Health check handlers
```

### Layer Pattern: Handler → Service → Repository

```
Handler (HTTP) → Service (Business Logic) → Repository (Database)
```

- **Handler:** Parse request, validate input, return response
- **Service:** Business logic, transactions, quota enforcement
- **Repository:** Database queries via GORM

**Tidak boleh** langsung query database dari Handler. **Wajib** pakai Service → Repository.

---

## 🔐 Authentication Flow

### JWT Token Structure

```go
type JWTClaims struct {
    sub  uint   // User ID
    exp  int64  // Expiry timestamp (30 days from login)
}
```

Token disimpan di header: `Authorization: Bearer <token>`

### Middleware Chain

```
Request → RequestID → StructuredLogger → CORS → EnsureJSON → RateLimiter → RequireAuth → Handler
```

### RequireAuth Middleware

1. Parse JWT token dari header
2. Validasi token expiry
3. Load user dari database
4. Resolve workspace context:
   - Jika `X-Workspace-ID` header ada → validasi membership
   - Jika tidak ada → auto-resolve ke personal workspace user
5. Set `c.Set("user", minimalUser)`
6. Set `c.Set("workspace_id", workspaceID)`
7. Berikan tier_info di setiap response

---

## 📊 Database Schema

### Core Tables

```sql
-- Users
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(100) UNIQUE,
    password_hash VARCHAR(255),
    name VARCHAR(100),
    phone VARCHAR(20),
    address TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Workspaces (formerly Organizations)
CREATE TABLE workspaces (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    owner_id BIGINT REFERENCES users(id),
    workspace_type VARCHAR(20) DEFAULT 'personal',  -- personal, team
    tier VARCHAR(20) DEFAULT 'free',                 -- free, pro, ultimate
    tier_expires_at TIMESTAMP WITH TIME ZONE,
    tier_activated_at TIMESTAMP WITH TIME ZONE,
    tier_activated_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- workspace_members (formerly organization_users)
CREATE TABLE workspace_members (
    id SERIAL PRIMARY KEY,
    workspace_id BIGINT REFERENCES workspaces(id),
    user_id BIGINT REFERENCES users(id),
    role VARCHAR(20), -- owner, admin, member
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(workspace_id, user_id)
);

-- project_statuses (Q19)
CREATE TABLE project_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50),
    color VARCHAR(7),
    index_order INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Projects
CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    description TEXT,
    workspace_id BIGINT REFERENCES workspaces(id),
    client_id BIGINT REFERENCES clients(id) ON DELETE SET NULL,
    priority VARCHAR(20) DEFAULT 'medium',
    progress BIGINT DEFAULT 0,
    budget DECIMAL,
    deadline DATE,
    status_id UUID REFERENCES project_statuses(id),
    version INT DEFAULT 1, -- optimistic locking
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- tasks
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200),
    description TEXT,
    project_id BIGINT REFERENCES projects(id),
    workspace_id BIGINT REFERENCES workspaces(id),
    status_id INT REFERENCES statuses(id),
    priority_id INT REFERENCES priorities(id),
    version INT DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- statuses (Task statuses per project)
CREATE TABLE statuses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50),
    color VARCHAR(7),
    index_order INT,
    project_id BIGINT REFERENCES projects(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- labels (Urgency labels per project)
CREATE TABLE labels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    color VARCHAR(7),
    project_id BIGINT REFERENCES projects(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- task_users (Assignees)
CREATE TABLE task_users (
    task_id BIGINT REFERENCES tasks(id),
    user_id BIGINT REFERENCES users(id),
    PRIMARY KEY (task_id, user_id)
);

-- task_labels (Labels assignment)
CREATE TABLE task_labels (
    task_id BIGINT REFERENCES tasks(id),
    label_id BIGINT REFERENCES labels(id),
    PRIMARY KEY (task_id, label_id)
);

-- priorities
CREATE TABLE priorities (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50),
    color VARCHAR(7),
    index_order INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- clients
CREATE TABLE clients (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(100),
    phone VARCHAR(20),
    whatsapp VARCHAR(20),
    company VARCHAR(100),
    website VARCHAR(255),
    address TEXT,
    notes TEXT,
    total_revenue DECIMAL DEFAULT 0,
    workspace_id BIGINT REFERENCES workspaces(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- invoices
CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    workspace_id BIGINT REFERENCES workspaces(id),
    client_id BIGINT REFERENCES clients(id) ON DELETE SET NULL,
    project_id BIGINT REFERENCES projects(id) ON DELETE SET NULL,
    invoice_number VARCHAR(50) UNIQUE,
    title VARCHAR(255),
    amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    tax DECIMAL(15,2) DEFAULT 0,
    discount DECIMAL(15,2) DEFAULT 0,
    amount_paid DECIMAL(15,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'draft', -- draft, sent, paid, cancelled
    due_date TIMESTAMP WITH TIME ZONE,
    paid_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    items JSONB,
    version INT DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- invitations
CREATE TABLE invitations (
    id SERIAL PRIMARY KEY,
    workspace_id BIGINT REFERENCES workspaces(id),
    email VARCHAR(100),
    role VARCHAR(20),
    token VARCHAR(100) UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- password_reset_tokens
CREATE TABLE password_reset_tokens (
    token VARCHAR(100) PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    expires_at TIMESTAMP WITH TIME ZONE,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- tier_plans (M5)
CREATE TABLE tier_plans (
    id SERIAL PRIMARY KEY,
    tier VARCHAR(20) NOT NULL UNIQUE, -- free, pro, ultimate
    name VARCHAR(50) NOT NULL,
    description TEXT,
    price_monthly INTEGER NOT NULL DEFAULT 0,
    price_yearly INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- tier_limits (M5)
CREATE TABLE tier_limits (
    id SERIAL PRIMARY KEY,
    tier VARCHAR(20) NOT NULL UNIQUE,
    max_workspaces INTEGER NOT NULL DEFAULT 1,
    max_projects INTEGER NOT NULL DEFAULT 3,      -- -1 = unlimited
    max_tasks_per_project INTEGER NOT NULL DEFAULT 50,
    max_members INTEGER NOT NULL DEFAULT 1,
    max_clients INTEGER NOT NULL DEFAULT 5,
    max_invoices_per_month INTEGER NOT NULL DEFAULT 10,
    can_comment BOOLEAN NOT NULL DEFAULT false,
    can_sse BOOLEAN NOT NULL DEFAULT false,
    can_audit_log BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

---

## 🌐 API Endpoints

### Public Endpoints (No Auth Required)

| Method | Endpoint | Description |
|---|---|---|
| POST | `/signup` | Register user + auto-login |
| POST | `/login` | Login user |
| POST | `/forgot-password` | Request password reset |
| POST | `/reset-password` | Reset password with token |
| GET | `/health` | Health check |
| GET | `/ready` | Readiness check |
| GET | `/tier/plans` | List all tier plans + pricing (public) |

### Protected Endpoints (JWT Required)

> **Note:** Semua endpoint dilindungi middleware `RequireAuth`. Jika header `X-Workspace-ID` tidak diberikan, sistem otomatis menggunakan personal workspace user.

#### Auth & Profile
| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/auth/me` | Get current user profile |
| PATCH | `/api/users/me` | Update profile (name, phone, address) |
| PATCH | `/api/users/me/password` | Change password |
| POST | `/api/users/me/switch-workspace` | Switch active workspace |
| GET | `/users/me/tier` | Get tier info + usage + limits |

#### Projects
| Method | Endpoint | Description |
|---|---|---|
| GET | `/projects` | List projects (includes `project_status`) |
| POST | `/projects` | Create project (auto-generate labels + status) |
| GET | `/projects/:id` | Get project (includes `project_status` and `client`) |
| PATCH | `/projects/:id` | Update project (update status_id, client_id) |
| DELETE | `/projects/:id` | Delete project |
| GET | `/projects/:id/labels` | Get project urgency labels |
| POST | `/projects/:id/labels` | Create label |
| PATCH | `/labels/:id` | Update label |
| DELETE | `/labels/:id` | Delete label |
| GET | `/projects/:id/status` | Get task statuses |
| POST | `/projects/:id/status` | Create task status |
| PATCH | `/status/:id` | Update task status |
| DELETE | `/status/:id` | Delete task status |

#### Tasks
| Method | Endpoint | Description |
|---|---|---|
| GET | `/projects/:id/tasks` | List tasks in project |
| POST | `/tasks` | Create task |
| GET | `/tasks/search` | Full-text search + filters (assignee, status, priority, date) |
| PATCH | `/tasks/:id` | Update task |
| DELETE | `/tasks/:id` | Delete task |

#### Workspaces
| Method | Endpoint | Description |
|---|---|---|
| GET | `/workspaces` | List user's workspaces |
| POST | `/workspaces` | Create workspace |
| POST | `/workspaces/invite` | Invite member (email) |
| GET | `/workspaces/members` | List members |
| PATCH | `/workspaces/members/:user_id` | Update member role |
| DELETE | `/workspaces/members/:user_id` | Remove member |
| GET | `/workspaces/invitations` | List pending invitations |

#### Clients & Invoices
| Method | Endpoint | Description |
|---|---|---|
| GET | `/clients` | List clients |
| POST | `/clients` | Create client |
| GET | `/clients/stats` | Revenue statistics |
| GET | `/clients/:id` | Get client detail |
| PATCH | `/clients/:id` | Update client |
| DELETE | `/clients/:id` | Delete client |
| GET | `/invoices` | List invoices |
| POST | `/invoices` | Create invoice (auto-generate INV-YYYY-XXX) |
| GET | `/invoices/:id` | Get invoice detail |
| PATCH | `/invoices/:id` | Update invoice |
| DELETE | `/invoices/:id` | Delete invoice |
| PATCH | `/invoices/:id/mark-paid` | Mark paid + sync client revenue |

#### Admin
| Method | Endpoint | Description |
|---|---|---|
| PATCH | `/admin/workspaces/:id/tier` | Activate/extend workspace tier (admin only) |

---

## 🔑 Key Features

### Subscription Tiers (Per Workspace)
- Tier diikat ke **workspace** (bukan ke user)
- User bisa punya workspace dengan tier berbeda
- Semua member workspace mewarisi fitur sesuai tier workspace
- Aktivasi manual oleh admin (tanpa Stripe/payment gateway)
- Quota enforcement di service layer berdasarkan workspace tier

### Quota Enforcement
Quota di-check sebelum operasi **Create**:

| Resource | Free | Pro | Ultimate |
|---|---|---|---|
| Workspaces | 1 | 2 | 4 |
| Projects per workspace | 3 | Unlimited | Unlimited |
| Tasks per project | 50 | Unlimited | Unlimited |
| Members per workspace | 1 | 5 | 15 |
| Clients | 5 | Unlimited | Unlimited |
| Invoices per bulan | 10 | Unlimited | Unlimited |

Jika quota exceeded → HTTP 403 `quota exceeded: project limit is 3 on free tier. Please upgrade.`

### Client-Project Link
- Project bisa linked ke sebuah client (opsional)
- Create project: accept `client_id` atau `new_client` (inline create)
- Update project: bisa ubah atau unlink `client_id`
- Validasi: client harus milik workspace yang sama

### Tier Feature Gate
Middleware `RequireTierFeature(feature)` blocking akses fitur berdasarkan workspace tier:

```go
// Route example
tasks.POST("/:id/comments", RequireTierFeature("comment"), commentHandler.Create)
```

### Tier Info Banner
Middleware `RequireAuth` inject `tier_info` ke setiap response:

```json
"tier_info": {
  "tier": "free",
  "effective_tier": "free",
  "is_active": true,
  "expires_at": null,
  "days_remaining": 0,
  "limits": { ... },
  "features": { ... },
  "usage": { ... }
}
```

### Auto-generate Urgency Labels
Saat project dibuat, auto-create 3 urgency labels:
- **Urgent** (red #EF4444) — harus selesai secepatnya
- **Normal** (yellow #F59E0B) — prioritas standar
- **Low** (green #22C55E) — bisa nanti

### Project Status Workflow
Seed 4 project_statuses saat startup:
- **Active** (green #22C55E) — DEFAULT
- **On Hold** (amber #F59E0B)
- **Completed** (blue #3B82F6)
- **Archived** (gray #6B7280)

Saat project dibuat, auto-generate 5 task statuses:
- Todo, On Progress, Done, Pending, Cancel

### Workspace Switch
- Endpoint: `POST /api/users/me/switch-workspace`
- Body: `{ "workspace_id": 1 }`
- Validates membership before switching

---

## 📋 API Response Format

### Success
```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... }
}
```

### Error
```json
{
  "success": false,
  "message": "Error description",
  "data": null
}
```

### Quota Exceeded (HTTP 403)
```json
{
  "success": false,
  "message": "quota exceeded: project limit is 3 on free tier. Please upgrade.",
  "data": null
}
```

---

## 🧩 Utility Functions

### utils/quota.go
Helper functions untuk tier management:

```go
// GetTierLimits(tier string) → TierLimits struct (free/pro/ultimate)
// IsTierActive(tier string, expiresAt *time.Time) → bool
// GetEffectiveTier(tier string, expiresAt *time.Time) → string (fallback to free if expired)
// GetEffectiveTierFromString(tier string, tierExpiresAtStr *string) → string (DB string format)
// ErrQuotaExceeded(resource, limit, tier) → *QuotaError
// DaysRemaining(expiresAt *time.Time) → int
```

### utils/response.go
Standard API response helpers:

```go
// SendSuccess(c, message, data)
// SendError(c, statusCode, message)
```

---

## 🚀 Running the Server

```bash
# Development
go run main.go

# Docker
docker-compose up --build

# Build
go build -o main .

# Test
go test ./...

# Generate Swagger docs
go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -o docs --parseDependency

# Migration
migrate -path ./migrations -database "postgres://..." up
```

---

## 📝 Environment Variables

```bash
# Database
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=gotaskdb
DB_PORT=5432

# Auth
SECRET_KEY=your-secret-key-min-32-chars

# SMTP (email)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=email@gmail.com
SMTP_PASSWORD=app-password
SMTP_FROM=noreply@gotask.app

# Logging
LOG_LEVEL=debug
```

---

## 🔒 Security Notes

- Password hashed dengan bcrypt
- JWT token expires 30 hari
- Rate limiting: 100 req/min (IP) + 500 req/min (user)
- CORS open di development, locked di production
- No hardcoded secrets (always use env vars)
- Optimistic locking via version column
- Tier feature gate middleware (per workspace)
- Workspace membership validation on every protected request

---

*Dokumen ini untuk developer. Untuk panduan pengguna non-teknis, lihat FLOW-USER.md*