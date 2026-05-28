# GoTask Backend

> High-performance Task Management RESTful API Backend built with Go, PostgreSQL, and GORM.

![Go Version](https://img.shields.io/badge/Go-1.24-blue)
![License](https://img.shields.io/badge/License-MIT-green)
![Build Status](https://github.com/your-org/gotask-backend/actions/workflows/ci.yml/badge.svg)

## 🎯 Overview

GoTask is a SaaS-ready task management backend API with multi-tenant organization support, role-based access control, and complete task lifecycle management — dari registration hingga invoice billing.

## ✨ Features

### Core Features

| Feature | Status |
|---|---|
| **User Registration** — Signup + auto-login dengan token JWT | ✅ |
| **Personal Workspace** — Auto-create workspace saat register | ✅ |
| **Workspace Switching** — Pindah antar organisasi dengan 1 endpoint | ✅ |
| **Project Management** — CRUD dengan auto-generate labels | ✅ |
| **Task Management** — Full CRUD dengan assignee, labels, priorities | ✅ |
| **Auto-generated Labels** — 5 label dibuat otomatis saat project dibuat | ✅ |
| **Project Status Workflow** — Active, On Hold, Completed, Archived | ✅ |
| **License Warning Banner** — Soft warning di setiap response API | ✅ |
| **Clients** — Kontak + revenue tracking | ✅ |
| **Invoices** — Auto-generate nomor + revenue sync saat lunas | ✅ |

### Security Features

| Feature | Status |
|---|---|
| **JWT Authentication** — Stateless login | ✅ |
| **RBAC** — Owner/Admin/Member scopes | ✅ |
| **Rate Limiting** — Per-IP + per-user | ✅ |
| **Password Hashing** — bcrypt | ✅ |
| **Optimistic Locking** — Version column untuk race condition | ✅ |
| **Graceful Shutdown** — SIGTERM handling | ✅ |
| **Structured Logging** — JSON slog | ✅ |
| **Health Checks** — `/health` + `/ready` | ✅ |

### API Features

| Feature | Status |
|---|---|
| **Swagger/OpenAPI** — Interactive docs di `/swagger` | ✅ |
| **Search** — Full-text search task (GIN index) | ✅ |
| **Task Filters** — By assignee, status, priority, date | ✅ |
| **Invite Members** — Email dengan token expiry + resend | ✅ |
| **Docker** — Multi-stage production build | ✅ |
| **CI/CD** — GitHub Actions | ✅ |

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
```

## 📋 User Flow

### Registration → Project → Task

```
1. User daftar (signup)
   → Sistem buat akun + personal workspace + role Owner
   → User langsung login otomatis

2. User buat project
   → Sistem auto-generate 3 urgency label: Urgent, Normal, Low
   → Sistem auto-generate project status: Active
   → Project siap dipakai

3. User invite member
   → Kirim email invitation
   → Member terima link → auto join workspace

4. User buat task
   → Default status: Todo (label pertama)
   → Assignee dari member workspace

5. User update task
   → Pindahkan antar label
   → Ganti assignee, priority, deadline
```

## 🔌 API Reference

### Public Endpoints

| Method | Endpoint | Description |
|---|---|---|
| POST | `/signup` | Register + auto-login (returns user + token) |
| POST | `/login` | Login (returns user + token) |
| POST | `/forgot-password` | Request reset link |
| POST | `/reset-password` | Reset dengan token |
| GET | `/health` | Liveness check |
| GET | `/ready` | Readiness check |

### Protected Endpoints (Bearer Token Required)

#### Auth & Profile
| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/auth/me` | Get current user |
| PATCH | `/api/users/me` | Update profile |
| PATCH | `/api/users/me/password` | Change password |
| POST | `/api/users/me/switch-organization` | Pindah workspace |

#### Projects
| Method | Endpoint | Description |
|---|---|---|
| GET | `/projects` | List projects (with project_status) |
| POST | `/projects` | Create project (auto-generate labels) |
| GET | `/projects/:id` | Get project detail |
| PATCH | `/projects/:id` | Update project (+ status_id) |
| DELETE | `/projects/:id` | Delete project |

#### Labels
| Method | Endpoint | Description |
|---|---|---|
| GET | `/projects/:id/labels` | List labels |
| POST | `/projects/:id/labels` | Create label |
| PATCH | `/labels/:id` | Update label |
| DELETE | `/labels/:id` | Delete label |

#### Tasks
| Method | Endpoint | Description |
|---|---|---|
| GET | `/projects/:id/tasks` | List tasks |
| GET | `/tasks/search` | Search tasks |
| POST | `/tasks` | Create task |
| PATCH | `/tasks/:id` | Update task |
| DELETE | `/tasks/:id` | Delete task |

#### Organizations
| Method | Endpoint | Description |
|---|---|---|
| GET | `/organizations` | List user's orgs |
| POST | `/organizations` | Create org |
| POST | `/organizations/invite` | Invite member |
| GET | `/organizations/members` | List members |
| PATCH | `/organizations/members/:id` | Update role |
| DELETE | `/organizations/members/:id` | Remove member |
| GET | `/organizations/invitations` | Pending invitations |

#### Clients & Invoices
| Method | Endpoint | Description |
|---|---|---|
| GET | `/clients` | List clients |
| POST | `/clients` | Create client |
| GET | `/clients/stats` | Revenue statistics |
| PATCH | `/clients/:id` | Update client |
| GET | `/invoices` | List invoices |
| POST | `/invoices` | Create invoice |
| PATCH | `/invoices/:id/mark-paid` | Mark lunas + sync revenue |

### Response Format

#### Success
```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... },
  "license_warning": { ... }  // Q17: jika expired/free plan
}
```

#### Error
```json
{
  "success": false,
  "message": "Error description",
  "data": null,
  "license_warning": { ... }
}
```

#### License Warning (Q17)
```json
{
  "license_warning": {
    "expired": true,
    "days_remaining": -7,
    "message": "License expired. Please upgrade to continue premium features."
  }
}
```

## 📁 Project Structure

```
gotask-backend/
├── config/             # Database connection + seeders
├── docs/               # Swagger docs + user flows
│   ├── FLOW-USER.md    # Panduan pengguna (non-technical)
│   ├── TECHNICAL.md    # Developer guide
│   └── specs/done/     # Archived specs
├── middlewares/         # Auth, CORS, rate limit, logging
├── models/             # Shared models (scopes, roles)
├── modules/            # Modular monolith
│   ├── auth/           # Signup, login, password
│   ├── clients/        # Client CRUD + stats
│   ├── invoices/       # Invoice + auto-numbering + revenue sync
│   ├── licenses/       # License validation + activation
│   ├── organizations/  # Org + members + invitations
│   ├── projects/       # Project CRUD + labels
│   └── tasks/          # Task, status, priority, labels
├── utils/               # Response helpers, logger
├── migrations/           # Database migrations
├── main.go              # Entry point + DI + routing
├── docker-compose.yml    # Development environment
└── Dockerfile           # Production build
```

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

## 🔒 Security

- Password di-hash dengan bcrypt
- JWT expires 30 hari
- Rate limiting: 100 req/min (IP) + 500 req/min (user)
- CORS configurable untuk development
- RBAC Owner/Admin/Member

## 📚 Documentation

| File | Audience |
|---|---|
| `docs/FLOW-USER.md` | Pengguna biasa (non-technical) |
| `docs/TECHNICAL.md` | Developer (technical) |
| `/swagger` | API consumers |
| `CLAUDE.md` | AI Engineer (project context) |

## 🚢 Deployment

```bash
# Docker Compose (development)
docker-compose up -d

# Production
docker build -t gotask-backend .
docker run -d -p 8080:8080 --env-file .env gotask-backend
```

---

Built with Go + PostgreSQL | All phases complete | Ready for production
