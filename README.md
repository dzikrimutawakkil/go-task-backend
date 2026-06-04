# GoTask Backend

> High-performance Task Management RESTful API Backend built with Go, PostgreSQL, and GORM.

![Go Version](https://img.shields.io/badge/Go-1.24-blue)
![License](https://img.shields.io/badge/License-MIT-green)

## 🎯 Overview

GoTask is a SaaS-ready task management backend API with multi-tenant workspace support, role-based access control, and subscription tier system — dari registration hingga invoice billing.

---

## ✨ Features

### Core Features

| Feature | Status | Notes |
|---|---|---|
| **User Registration** | ✅ | Signup + auto-login dengan token JWT |
| **Personal Workspace** | ✅ | Auto-create workspace saat register |
| **Subscription Tiers** | ✅ | Free/Pro/Ultimate dengan quota enforcement |
| **Quota Enforcement** | ✅ | Hard limit di service layer sebelum Create |
| **Workspace Switching** | ✅ | Pindah antar workspace dengan 1 endpoint |
| **Project Management** | ✅ | CRUD dengan auto-generate urgency labels |
| **Task Management** | ✅ | Full CRUD dengan assignee, labels, priorities |
| **Urgency Labels** | ✅ | 3 labels (Urgent, Normal, Low) auto-generated per project |
| **Project Status Workflow** | ✅ | Active, On Hold, Completed, Archived |
| **Tier Info Banner** | ✅ | Soft warning `tier_info` di setiap response API |
| **Clients** | ✅ | Kontak + revenue tracking |
| **Invoices** | ✅ | Auto-generate nomor + revenue sync saat lunas |
| **Invite Members** | ✅ | Email invitation dengan token expiry + resend |

### Security Features

| Feature | Status |
|---|---|
| **JWT Authentication** | ✅ |
| **RBAC** | ✅ | Owner / Admin / Member |
| **Rate Limiting** | ✅ | Per-IP + per-user |
| **Password Hashing** | ✅ | bcrypt |
| **Optimistic Locking** | ✅ | Version column |
| **Graceful Shutdown** | ✅ | SIGTERM handling |
| **Structured Logging** | ✅ | JSON slog |
| **Health Checks** | ✅ | `/health` + `/ready` |

### API Features

| Feature | Status |
|---|---|
| **Swagger/OpenAPI** | ✅ | `/swagger/index.html` |
| **Full-text Search** | ✅ | Task by title/description (GIN index) |
| **Task Filters** | ✅ | By assignee, status, priority, date range |
| **Docker** | ✅ | Multi-stage production build |
| **CI/CD** | ✅ | GitHub Actions |

---

## 🚀 Quick Start

### 1. Clone & Setup

```bash
git clone https://github.com/your-org/gotask-backend
cd gotask-backend
cp .env.example .env
# Edit .env dengan nilai Anda
```

### 2. Environment Variables

```env
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=gotaskdb
DB_PORT=5432
SECRET_KEY=your_secret_key_minimum_32_chars
LOG_LEVEL=debug
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=noreply@gotask.app
```

### 3. Jalankan

```bash
# Development
go run main.go

# Docker
docker-compose up --build

# Production build
go build -o main .
./main
```

### 4. Verifikasi

```bash
# Health check
curl http://localhost:8080/health

# Swagger docs
open http://localhost:8080/swagger/index.html

# List tier plans (public)
curl http://localhost:8080/tier/plans
```

---

## 📋 Subscription Tiers

| Fitur | Free | Pro | Ultimate |
|---|---|---|---|
| **Workspace** | 1 (personal only) | 2 | 4 |
| **Project per workspace** | 3 | Unlimited | Unlimited |
| **Task per project** | 50 | Unlimited | Unlimited |
| **Member per workspace** | 1 | 3 | 15 |
| **Clients** | 5 | Unlimited | Unlimited |
| **Invoices per bulan** | 10 | Unlimited | Unlimited |
| **Comments** | ❌ | ✅ | ✅ |
| **Real-time (SSE)** | ❌ | ✅ | ✅ |
| **Audit Log** | ❌ | ❌ | ✅ |

- Tier diikat ke **workspace** (bukan user)
- Aktivasi dilakukan **manual** oleh admin (tanpa Stripe)

---

## 🔌 API Reference

### Public Endpoints (No Auth Required)

| Method | Endpoint | Description |
|---|---|---|
| POST | `/signup` | Register + auto-login (returns `user` + `token`) |
| POST | `/login` | Login (returns `user` + `token`) |
| POST | `/forgot-password` | Request password reset email |
| POST | `/reset-password` | Reset password dengan token |
| GET | `/health` | Liveness check |
| GET | `/ready` | Readiness check |
| GET | `/tier/plans` | List all tiers with pricing + limits (public) |

### Protected Endpoints (Bearer Token Required)

> **Note:** Semua endpoint dilindungi middleware `RequireAuth`. Jika header `X-Workspace-ID` tidak diberikan, sistem otomatis menggunakan personal workspace user.

#### Auth & Profile
| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/auth/me` | Get current user profile |
| PATCH | `/api/users/me` | Update profile (name, phone, address) |
| PATCH | `/api/users/me/password` | Change password |
| POST | `/api/users/me/switch-workspace` | Switch active workspace |
| GET | `/users/me/tier` | Get current tier + usage + limits |

#### Projects
| Method | Endpoint | Description |
|---|---|---|
| GET | `/projects` | List projects (includes `project_status`) |
| POST | `/projects` | Create project (auto-generate labels + status) |
| GET | `/projects/:id` | Get project detail (includes `project_status`) |
| PATCH | `/projects/:id` | Update project (includes `status_id`) |
| DELETE | `/projects/:id` | Delete project |

#### Status
| Method | Endpoint | Description |
|---|---|---|
| GET | `/projects/:id/status` | List project statuses |
| POST | `/projects/:id/status` | Create status |
| PATCH | `/status/:id` | Update status |
| DELETE | `/status/:id` | Delete status |

#### Labels (Urgency Level)
| Method | Endpoint | Description |
|---|---|---|
| GET | `/projects/:id/labels` | List urgency labels |
| POST | `/projects/:id/labels` | Create label |
| PATCH | `/labels/:id` | Update label |
| DELETE | `/labels/:id` | Delete label |

#### Tasks
| Method | Endpoint | Description |
|---|---|---|
| GET | `/projects/:id/tasks` | List tasks by project |
| GET | `/tasks/search` | Full-text search + filters |
| POST | `/tasks` | Create task |
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
| GET | `/workspaces/invitations` | Pending invitations |

#### Clients
| Method | Endpoint | Description |
|---|---|---|
| GET | `/clients` | List clients |
| POST | `/clients` | Create client |
| GET | `/clients/stats` | Revenue statistics |
| GET | `/clients/:id` | Get client detail |
| PATCH | `/clients/:id` | Update client |
| DELETE | `/clients/:id` | Delete client |

#### Invoices
| Method | Endpoint | Description |
|---|---|---|
| GET | `/invoices` | List invoices |
| POST | `/invoices` | Create invoice |
| GET | `/invoices/:id` | Get invoice detail |
| PATCH | `/invoices/:id` | Update invoice |
| DELETE | `/invoices/:id` | Delete invoice |
| PATCH | `/invoices/:id/mark-paid` | Mark paid + sync client revenue |

#### Admin
| Method | Endpoint | Description |
|---|---|---|
| PATCH | `/admin/workspaces/:id/tier` | Activate/extend workspace tier |

---

## 📦 Response Format

### Standard Success
```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... },
  "tier_info": {
    "tier": "free",
    "is_active": true,
    "days_remaining": 0,
    "warning": "Upgrade untuk lebih banyak fitur."
  }
}
```

### Standard Error
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
  "message": "quota exceeded: project limit is 3 on free tier. Please upgrade.",
  "data": null
}
```

---

## 📁 Project Structure

```
go-task-backend/
├── config/                 # Database connection + seeders
├── docs/                  # Documentation
│   ├── FLOW-USER.md       # Panduan pengguna (non-technical)
│   ├── TECHNICAL.md       # Developer guide
│   └── specs/done/        # Archived specs
├── middlewares/           # Auth, CORS, rate limit, tier feature gate
├── models/                 # Shared models (scopes, roles)
├── modules/                # Modular monolith
│   ├── auth/              # Signup, login, password reset
│   ├── clients/           # Client CRUD + stats
│   ├── health/            # Health check handlers
│   ├── invoices/           # Invoice + auto-numbering + revenue sync
│   ├── projects/          # Project CRUD + auto-generate labels/status
│   ├── tasks/             # Task, status, priority, labels
│   └── workspaces/        # Workspace + members + invitations + tier plans
├── utils/                  # Response helpers, logger, quota helpers
├── internal/               # Interfaces
├── migrations/              # Versioned database migrations (golang-migrate)
├── main.go                  # Entry point + DI + routing
├── docker-compose.yml       # Development environment
└── Dockerfile              # Production multi-stage build
```

---

## 🛠 Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.24 |
| Framework | Gin v1.11 |
| Database | PostgreSQL 15+ |
| ORM | GORM v1.31 |
| Auth | JWT (golang-jwt/jwt/v5) |
| Migrations | golang-migrate |
| Container | Docker |
| Docs | Swagger/OpenAPI |

---

## 🔒 Security

- Password di-hash dengan bcrypt
- JWT expires 30 hari
- Rate limiting: 100 req/min (IP) + 500 req/min (user)
- CORS configurable untuk development
- RBAC Owner / Admin / Member
- Tier feature gate middleware
- Workspace membership validation on every protected request

---

## 📚 Documentation

| File | Audience |
|---|---|
| `docs/FLOW-USER.md` | Pengguna biasa (non-technical) |
| `docs/TECHNICAL.md` | Developer (technical) |
| `/swagger/index.html` | API consumers |
| `CLAUDE.md` | AI Engineer (project context) |

---

## 🚢 Deployment

```bash
# Docker Compose (development)
docker-compose up -d

# Production
docker build -t gotask-backend .
docker run -d -p 8080:8080 --env-file .env gotask-backend
```

---

## 🔄 User Flow

```
Pengguna → Daftar → Dapat Workspace + Tier Free → Buat Project → Buat Task
                              ↓
                        Invite Member (opsional)
                        Bergabung ke Workspace lain
```

### Tier Activation (Admin)

```bash
# Activate pro tier untuk 1 bulan
curl -X PATCH http://localhost:8080/admin/workspaces/1/tier \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tier":"pro","duration_months":1}'
```

---

Built with Go + PostgreSQL | Production-ready