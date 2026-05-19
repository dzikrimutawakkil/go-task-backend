# GoTask Backend

> High-performance Task Management RESTful API Backend built with Go, PostgreSQL, and GORM.

![Go Version](https://img.shields.io/badge/Go-1.24-blue)
![License](https://img.shields.io/badge/License-MIT-green)
![Build Status](https://github.com/your-org/gotask-backend/actions/workflows/ci.yml/badge.svg)

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [API Documentation](#api-documentation)
- [Project Structure](#project-structure)
- [Deployment Guide](#deployment-guide)
- [Contributing](#contributing)
- [License](#license)

## 🎯 Overview

GoTask is a SaaS-ready task management backend API designed to support multi-tenant organizations with role-based access control. Built with modern Go practices, it provides secure authentication, comprehensive task management, and scalable architecture.

**Key Characteristics:**
- JWT-based authentication
- Organization & Project hierarchy
- Task CRUD with labels, statuses, and priorities
- Optimistic locking for concurrent updates
- Rate limiting middleware
- Structured JSON logging

## ✨ Features

### Core Features
- [x] **Authentication**: Signup, Login, JWT tokens
- [x] **Organizations**: Create, invite members, role management
- [x] **Projects**: CRUD within organizations
- [x] **Tasks**: Full CRUD with filtering, search, pagination
- [x] **Statuses**: Auto-created with projects, reorderable
- [x] **Labels**: Custom labels per project
- [x] **Assignees**: Multiple users per task

### Security Features
- [x] **RBAC**: Role-based access control (Admin, Manager, Member)
- [x] **Rate Limiting**: Per-IP and per-user limits
- [x] **Password Hashing**: bcrypt with cost factor 10
- [x] **CORS**: Configurable cross-origin policies
- [x] **Non-root Docker**: Runs as non-root user

### Developer Experience
- [x] **Swagger UI**: Interactive API documentation at `/swagger/*any`
- [x] **Graceful Shutdown**: Clean server termination
- [x] **Structured Logging**: JSON logs for production
- [x] **Health Checks**: `/health` and `/ready` endpoints

## 🛠 Tech Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| Language | Go 1.24 | Core application |
| Framework | Gin | HTTP routing & middleware |
| ORM | GORM | Database operations |
| Database | PostgreSQL 16+ | Primary data store |
| Auth | JWT (golang-jwt) | Stateless authentication |
| Migrations | golang-migrate | Schema version control |
| Documentation | swaggo | OpenAPI/Swagger generation |
| Container | Docker, Docker Compose | Deployment & dev environment |

## 📦 Prerequisites

| Requirement | Version | Notes |
|------------|---------|-------|
| Go | 1.24+ | [Install Guide](https://go.dev/doc/install) |
| PostgreSQL | 16+ | [Install Guide](https://www.postgresql.org/download/) |
| Docker | Latest | Optional, for containerized setup |
| Git | Any | Clone repository |

## 🚀 Quick Start

### 1. Clone & Setup Environment

```bash
git clone https://github.com/your-org/gotask-backend.git
cd gotask-backend
```

### 2. Configure Environment Variables

```bash
# Copy example environment file
cp .env.example .env

# Edit with your values
nano .env
```

**Required Variables:**

```env
# Database
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=gotaskdb
DB_PORT=5432

# JWT Authentication
SECRET_KEY=your_super_secret_key_at_least_32_chars

# Application
APP_URL=http://localhost:8080
LOG_LEVEL=debug
```

### 3. Start Database

```bash
# Using Docker Compose (recommended)
docker-compose up -d postgres

# Or use existing PostgreSQL instance
```

### 4. Run Migrations

```bash
# Auto-migrate on startup (default behavior)
# OR run manually:
migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/gotaskdb?sslmode=disable" up
```

### 5. Run the Application

```bash
# Development mode (with live reload)
air

# OR production build
go build -o gotask-backend .
./gotask-backend
```

### 6. Verify Installation

```bash
# Health check
curl http://localhost:8080/health

# Swagger UI
open http://localhost:8080/swagger/index.html
```

## 📚 API Documentation

### Base URL

```
http://localhost:8080
```

### Authentication Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/signup` | Register new user |
| POST | `/login` | Authenticate & get JWT |

### Project Endpoints (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects` | List all projects |
| POST | `/projects` | Create new project |
| DELETE | `/projects/:id` | Delete project |

### Task Endpoints (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects/:id/tasks` | List tasks by project |
| GET | `/tasks/search` | Search tasks |
| POST | `/tasks` | Create new task |
| PATCH | `/tasks/:id` | Update task |
| DELETE | `/tasks/:id` | Delete task |

### Status Endpoints (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects/:id/status` | List project statuses |
| POST | `/projects/:id/status` | Create status |
| PATCH | `/status/:id` | Update status |
| DELETE | `/status/:id` | Delete status |

### Label Endpoints (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects/:id/labels` | List project labels |
| POST | `/projects/:id/labels` | Create label |
| PATCH | `/labels/:id` | Update label |
| DELETE | `/labels/:id` | Delete label |

### Organization Endpoints (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/organizations` | Create organization |
| POST | `/organizations/invite` | Invite member |
| GET | `/organizations/members` | List members |
| DELETE | `/organizations/members/:user_id` | Remove member |

### Request Headers

All protected endpoints require:

```
Authorization: Bearer <jwt_token>
```

### Response Format

```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... }
}
```

### Error Response

```json
{
  "success": false,
  "message": "Error description",
  "data": null
}
```

## 📁 Project Structure

```
gotask-backend/
├── config/              # Database connection & configuration
├── docs/                # Swagger generated documentation
├── handlers/             # HTTP handlers (health, etc.)
├── middlewares/          # Gin middlewares (auth, logging, CORS, rate limit)
├── migrations/           # Database migration files
├── models/              # Database models
├── modules/             # Feature modules (Clean Architecture)
│   ├── auth/            # Authentication module
│   ├── organizations/    # Organization module
│   ├── projects/        # Project module
│   └── tasks/           # Task, Status, Label module
├── routes/              # Route definitions
├── utils/               # Utility functions (logger, response helpers)
├── docs/
│   ├── DEPLOYMENT.md     # Deployment guide
│   └── SPEC.md          # Technical specification
├── main.go              # Application entry point
├── Dockerfile           # Multi-stage production Dockerfile
├── docker-compose.yml  # Development environment
├── .env.example        # Environment variable template
├── go.mod / go.sum     # Go dependencies
└── README.md           # This file
```

### Module Structure (Clean Architecture)

Each module follows the same pattern:

```
modules/<module>/
├── *_handler.go    # HTTP handlers (controllers)
├── *_service.go    # Business logic
├── *_repository.go # Data access layer
└── *_model.go      # Data models
```

## 🚢 Deployment Guide

See [DEPLOYMENT.md](docs/DEPLOYMENT.md) for detailed deployment instructions.

### Quick Deploy Options

#### Docker Compose (Development/Staging)

```bash
docker-compose up -d
```

#### Render

1. Connect GitHub repository
2. Configure environment variables
3. Set build command: `go build -o gotask-backend .`
4. Set start command: `./gotask-backend`

#### Railway

1. Import project from GitHub
2. Add PostgreSQL database
3. Configure environment variables
4. Deploy automatically

#### VPS (Ubuntu)

```bash
# 1. Install dependencies
sudo apt update && sudo apt upgrade -y
sudo apt install -y docker.io docker-compose

# 2. Clone and configure
git clone https://github.com/your-org/gotask-backend.git
cd gotask-backend
cp .env.example .env
nano .env

# 3. Run with Docker
docker build -t gotask-backend .
docker run -d -p 8080:8080 --env-file .env gotask-backend
```

### Environment Variables for Production

```env
# Database (Production)
DB_HOST=production-db-host
DB_USER=production_user
DB_PASSWORD=secure_production_password
DB_NAME=gotask_production
DB_PORT=5432

# JWT
SECRET_KEY=very_long_random_secret_key_at_least_32_characters

# Application
GIN_MODE=release
LOG_LEVEL=warn
APP_URL=https://api.gotask.app

# Optional: SMTP for email notifications
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=notifications@yourapp.com
SMTP_PASSWORD=app_password
```

## 🤝 Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/your-feature`
3. **Commit changes**: `git commit -m 'Add new feature'`
4. **Push to branch**: `git push origin feature/your-feature`
5. **Open a Pull Request**

### Development Setup

```bash
# Install development tools
go install github.com/cosmtrek/air@latest  # Live reload
go install github.com/swaggo/swag/cmd/swag@latest  # Swagger

# Run with live reload
air

# Generate Swagger docs
swag init -g main.go -o docs/generated
```

### Code Standards

- Follow Go idioms and `gofmt`
- Run `golangci-lint run` before committing
- Write tests for new features
- Update Swagger docs when changing API

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Built with ❤️ for the GoTask community**