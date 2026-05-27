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
├── auth/           # Authentication (signup, login, token)
├── organizations/    # Organization management
├── projects/        # Project CRUD
├── tasks/           # Task, Status, Label
├── clients/         # Client management
├── invoices/        # Invoice & billing
└── licenses/        # License activation
```

### Layer Pattern: Handler → Service → Repository

```
Handler (HTTP) → Service (Business Logic) → Repository (Database)
```

- Handler: Parse request, validate input, return response
- Service: Business logic, transactions
- Repository: Database queries

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
4. Resolve organization context (dari X-Organization-ID header atau personal workspace)
5. Set `c.Set("user", minimalUser)
6. Set `c.Set("org_id", orgID)`

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
    plan VARCHAR(20) DEFAULT 'free',
    license_key VARCHAR(100),
    license_status VARCHAR(20),
    created_at TIMESTAMP
);

-- Organizations (Workspaces)
CREATE TABLE organizations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    owner_id BIGINT REFERENCES users(id),
    org_type VARCHAR(20) DEFAULT 'personal',
    created_at TIMESTAMP
);

-- organization_users (Membership)
CREATE TABLE organization_users (
    organization_id BIGINT REFERENCES organizations(id),
    user_id BIGINT REFERENCES users(id),
    role VARCHAR(20), -- owner, admin, member
    joined_at TIMESTAMP
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
    status_id UUID REFERENCES project_statuses(id), -- Q19
    version INT DEFAULT 1, -- optimistic locking
    created_at TIMESTAMP
);

-- project_statuses (Q19)
CREATE TABLE project_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50),
    color VARCHAR(7),
    created_at TIMESTAMP
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
    created_at TIMESTAMP
);

-- statuses (Task statuses per project)
CREATE TABLE statuses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50),
    index_order INT,
    project_id BIGINT REFERENCES projects(id),
    created_at TIMESTAMP
);

-- labels (Task labels per project)
CREATE TABLE labels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    color VARCHAR(7),
    project_id BIGINT REFERENCES projects(id),
    created_at TIMESTAMP
);

-- task_users (Assignees)
CREATE TABLE task_users (
    task_id BIGINT REFERENCES tasks(id),
    user_id BIGINT REFERENCES users(id)
);

-- task_labels (Labels assignment)
CREATE TABLE task_labels (
    task_id BIGINT REFERENCES tasks(id),
    label_id BIGINT REFERENCES labels(id)
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
    organization_id BIGINT REFERENCES organizations(id)
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
    paid_at TIMESTAMP,
    created_at TIMESTAMP
);

-- licenses
CREATE TABLE licenses (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE,
    type VARCHAR(50), -- free, pro, team, enterprise
    status VARCHAR(20), -- available, activated, revoked, expired
    activated_by BIGINT REFERENCES users(id),
    activated_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP
);

-- invitations
CREATE TABLE invitations (
    id SERIAL PRIMARY KEY,
    organization_id BIGINT REFERENCES organizations(id),
    email VARCHAR(100),
    role VARCHAR(20),
    token VARCHAR(100) UNIQUE,
    expires_at TIMESTAMP,
    created_at TIMESTAMP
);

-- password_reset_tokens
CREATE TABLE password_reset_tokens (
    token VARCHAR(100) PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    expires_at TIMESTAMP,
    used_at TIMESTAMP,
    created_at TIMESTAMP
);
```

---

## 🌐 API Endpoints

### Public (No Auth Required)

| Method | Endpoint | Description |
|---|---|---|
| POST | `/signup` | Register user |
| POST | `/login` | Login user |
| POST | `/forgot-password` | Request password reset |
| POST | `/reset-password` | Reset password with token |
| GET | `/health` | Health check |
| GET | `/ready` | Readiness check |
| POST | `/api/licenses/validate` | Validate license key |

### Protected (JWT Required)

#### Auth & Profile
| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/auth/me` | Get current user profile |
| PATCH | `/api/users/me` | Update profile |
| PATCH | `/api/users/me/password` | Change password |
| POST | `/api/users/me/switch-organization` | Switch active workspace (M11) |

#### Projects
| Method | Endpoint | Description |
|---|---|---|
| GET | `/projects` | List projects |
| POST | `/projects` | Create project (Q18 auto-create labels) |
| GET | `/projects/:id` | Get project (includes project_status) |
| PATCH | `/projects/:id` | Update project (M12 status update) |
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
| PATCH | `/tasks/:id` | Update task |
| DELETE | `/tasks/:id` | Delete task |
| GET | `/tasks/search` | Search tasks |

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

#### Clients & Invoices
| Method | Endpoint | Description |
|---|---|---|
| GET | `/clients` | List clients |
| POST | `/clients` | Create client |
| GET | `/clients/:id` | Get client |
| PATCH | `/clients/:id` | Update client |
| DELETE | `/clients/:id` | Delete client |
| GET | `/clients/stats` | Client statistics |
| GET | `/invoices` | List invoices |
| POST | `/invoices` | Create invoice |
| GET | `/invoices/:id` | Get invoice |
| PATCH | `/invoices/:id` | Update invoice |
| DELETE | `/invoices/:id` | Delete invoice |
| PATCH | `/invoices/:id/mark-paid` | Mark invoice as paid |

#### Licenses
| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/licenses/activate` | Activate license |
| POST | `/api/licenses` | Create license (admin) |

---

## 🔑 Key Features

### Q17: License Warning
- Middleware `RequireAuth` check license status on every request
- License warning injected ke response context
- Response include `license_warning` field (frontend show banner)

### Q18: Auto-generate Labels
- Saat project dibuat, auto-create 5 label:
  - Todo (gray #E2E8F0)
  - On Going (blue #3B82F6)
  - Done (green #22C55E)
  - Delivered (purple #A855F7)
  - Canceled (red #EF4444)

### Q19: Project Status
- Seed 4 project_statuses saat startup:
  - Active (green #22C55E) - DEFAULT
  - On Hold (amber #F59E0B)
  - Completed (blue #3B82F6)
  - Archived (gray #6B7280)

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
  "message": "Success",
  "data": { ... },
  "license_warning": { ... }
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

### License Warning (Q17)
```json
"license_warning": {
  "expired": true,
  "days_remaining": -7,
  "message": "License expired. Please upgrade to continue premium features."
}
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
```

---

## 📝 Environment Variables

```bash
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=gotaskdb
DB_PORT=5432
SECRET_KEY=your-secret-key
MIGRATION_URL=postgres://user:pass@localhost:5432/dbname
```

---

## 🔒 Security Notes

- Password hashed dengan bcrypt
- JWT token expires 30 hari
- Rate limiting: 100 req/min (IP) + 500 req/min (user)
- CORS open di development, locked di production
- No hardcoded secrets (always use env vars)
- Optimistic locking via version column

---

*Dokumen ini untuk developer. Untuk panduan pengguna non-teknis, lihat FLOW-USER.md*
