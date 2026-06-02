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
├── organizations/      # Organization management + tier plans + invitations
├── projects/          # Project CRUD + labels
├── tasks/             # Task, status, priority, labels
├── clients/           # Client management + stats
└── invoices/          # Invoice + billing + revenue sync
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
Request → CORSMiddleware → EnsureJSON → RateLimiter → RequireAuth → Handler
```

### RequireAuth Middleware

1. Parse JWT token dari header
2. Validasi token expiry
3. Load user dari database
4. Resolve organization context (dari `X-Organization-ID` header atau personal workspace)
5. Set `c.Set("user", minimalUser)`
6. Set `c.Set("org_id", orgID)`
7. Inject `tier_info` ke response context

---

## 📊 Database Schema

### Core Tables

```sql
-- Users (with tier)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(100) UNIQUE,
    password_hash VARCHAR(255),
    name VARCHAR(100),
    phone VARCHAR(20),
    address TEXT,
    tier VARCHAR(20) DEFAULT 'free',          -- free, pro, ultimate
    tier_expires_at TIMESTAMP WITH TIME ZONE,
    tier_activated_at TIMESTAMP WITH TIME ZONE,
    tier_activated_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Organizations (Workspaces)
CREATE TABLE organizations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    owner_id BIGINT REFERENCES users(id),
    org_type VARCHAR(20) DEFAULT 'personal',  -- personal, team
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- organization_users (Membership)
CREATE TABLE organization_users (
    id SERIAL PRIMARY KEY,
    organization_id BIGINT REFERENCES organizations(id),
    user_id BIGINT REFERENCES users(id),
    role VARCHAR(20), -- owner, admin, member
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id, user_id)
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
    organization_id BIGINT REFERENCES organizations(id),
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
    total_revenue DECIMAL DEFAULT 0,
    organization_id BIGINT REFERENCES organizations(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- invoices
CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    client_id BIGINT REFERENCES clients(id),
    organization_id BIGINT REFERENCES organizations(id),
    amount DECIMAL,
    amount_paid DECIMAL,
    status VARCHAR(20) DEFAULT 'sent', -- sent, paid, cancelled
    invoice_number VARCHAR(20), -- auto-generated: INV-YYYY-XXX
    paid_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- invitations
CREATE TABLE invitations (
    id SERIAL PRIMARY KEY,
    organization_id BIGINT REFERENCES organizations(id),
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
    tier VARCHAR(20) NOT NULL UNIQUE,
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

#### Auth & Profile
| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/auth/me` | Get current user profile |
| PATCH | `/api/users/me` | Update profile |
| PATCH | `/api/users/me/password` | Change password |
| POST | `/api/users/me/switch-organization` | Switch active workspace |
| GET | `/api/users/me/tier` | Get tier info + usage + limits |

#### Projects
| Method | Endpoint | Description |
|---|---|---|
| GET | `/projects` | List projects |
| POST | `/projects` | Create project (auto-generate labels + status) |
| GET | `/projects/:id` | Get project (includes project_status) |
| PATCH | `/projects/:id` | Update project (update status_id) |
| DELETE | `/projects/:id` | Delete project |
| GET | `/projects/:id/labels` | Get project labels |
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
| GET | `/tasks/search` | Full-text search tasks |
| PATCH | `/tasks/:id` | Update task |
| DELETE | `/tasks/:id` | Delete task |

#### Organizations
| Method | Endpoint | Description |
|---|---|---|
| GET | `/organizations` | List user's organizations |
| POST | `/organizations` | Create organization |
| POST | `/organizations/invite` | Invite member |
| GET | `/organizations/members` | List members |
| PATCH | `/organizations/members/:id` | Update member role |
| DELETE | `/organizations/members/:id` | Remove member |
| GET | `/organizations/invitations` | List pending invitations |
| POST | `/organizations/invitations/:id/resend` | Resend invitation email |

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
| PATCH | `/api/admin/users/:id/tier` | Activate/extend user tier (admin only) |

---

## 🔑 Key Features

### Subscription Tiers (M5)
- Tier diikat ke **user** (bukan organisasi)
- Semua organisasi milik user mewarisi tier yang sama
- Aktivasi manual oleh admin (tanpa Stripe/payment gateway)
- Quota enforcement di service layer

### Quota Enforcement
Quota di-check sebelum operasi **Create**:

| Resource | Free | Pro | Ultimate |
|---|---|---|---|
| Workspaces | 1 | 2 | 4 |
| Projects per workspace | 3 | Unlimited | Unlimited |
| Tasks per project | 50 | Unlimited | Unlimited |
| Members per workspace | 1 | 3 | 15 |
| Clients | 5 | Unlimited | Unlimited |
| Invoices per bulan | 10 | Unlimited | Unlimited |

Jika quota exceeded → HTTP 403 `quota exceeded: workspace limit is 1 on free tier. Please upgrade.`

### Tier Feature Gate
Middleware `RequireTierFeature(feature)` blocking akses fitur berdasarkan tier:

```go
// Route example
tasks.POST("/:id/comments", RequireTierFeature("comment"), commentHandler.Create)
```

### Q17: Tier Info
Middleware `RequireAuth` inject `tier_info` ke setiap response:

```json
"tier_info": {
  "tier": "free",
  "is_active": true,
  "expires_at": null,
  "days_remaining": 0,
  "warning": "Upgrade untuk lebih banyak fitur."
}
```

### Q18: Auto-generate Urgency Labels
Saat project dibuat, auto-create 3 urgency labels:
- **Urgent** (red #EF4444) — harus selesai secepatnya
- **Normal** (yellow #F59E0B) — prioritas standar
- **Low** (blue #3B82F6) — bisa nanti

### Q19: Project Status Workflow
Seed 4 project_statuses saat startup:
- **Active** (green #22C55E) — DEFAULT
- **On Hold** (amber #F59E0B)
- **Completed** (blue #3B82F6)
- **Archived** (gray #6B7280)

Saat project dibuat, auto-generate 5 task statuses:
- Todo, On Progress, Done, Pending, Cancel

### M11: Workspace Switch
- Endpoint: `POST /api/users/me/switch-organization`
- Body: `{ "organization_id": 1 }`
- Validates membership before switching

### M12: Project Status Workflow
- `PATCH /projects/:id` accept `status_id` field
- Response include `project_status` object (id, name, color)

---

## 📋 API Response Format

### Success
```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... },
  "tier_info": {
    "tier": "pro",
    "is_active": true,
    "expires_at": "2027-05-28T00:00:00Z",
    "days_remaining": 365,
    "warning": null
  }
}
```

### Error
```json
{
  "success": false,
  "message": "Error description",
  "data": null,
  "tier_info": { ... }
}
```

### Quota Exceeded (HTTP 403)
```json
{
  "success": false,
  "message": "quota exceeded: workspace limit is 1 on free tier. Please upgrade.",
  "data": null,
  "tier_info": { ... }
}
```

---

## 🧩 Utility Functions

### utils/quota.go
Helper functions untuk tier management:

```go
// GetTierLimits(tier string) → TierLimits struct
// IsTierActive(tier string, expiresAt *time.Time) → bool
// GetEffectiveTier(tier string, expiresAt *time.Time) → string (fallback to free if expired)
// ErrQuotaExceeded(resource, limit, tier) → *QuotaError
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
```

---

## 🔒 Security Notes

- Password hashed dengan bcrypt
- JWT token expires 30 hari
- Rate limiting: 100 req/min (IP) + 500 req/min (user)
- CORS open di development, locked di production
- No hardcoded secrets (always use env vars)
- Optimistic locking via version column
- Tier feature gate middleware

---

*Dokumen ini untuk developer. Untuk panduan pengguna non-teknis, lihat FLOW-USER.md*