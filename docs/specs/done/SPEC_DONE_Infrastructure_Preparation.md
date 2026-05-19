# ACTIVE_SPEC.md - Infrastructure & Documentation Preparation

**Project**: GoTask Backend - SaaS Task Management Platform
**Status**: In Progress
**Phase**: 1.5 - Waiting for Frontend API Contract Review
**Created**: 2026-05-19
**CEO**: Dzikri

---

## 📋 Context / Background

Backend GoTask sudah jalan di lokal dengan Swagger accessible. File `swagger.json` sudah dikirim ke teman (frontend developer) untuk review API contract. Teman menggunakan React + Firebase (MVP stage).

**Goal**: Siapkan infrastruktur dan dokumentasi SELAGI MENUNGGU feedback dari frontend developer tentang:
- API apa yang needed
- API apa yang missing dari GoTask
- Data apa yang akan dipindahkan dari Firebase ke PostgreSQL

---

## 🎯 Objectives

### Quick Wins (High Impact, Low Effort) - PRIORITAS UTAMA 🚀

1. **Setup CI/CD Pipeline**
   - GitHub Actions untuk automated testing & build
   - Automated deployment trigger

2. **Setup Docker Production Ready**
   - Multi-stage Dockerfile untuk production
   - Docker Compose untuk development + production
   - Environment variable management

3. **Buat README Lengkap**
   - Project overview
   - Setup instructions
   - API documentation summary
   - Deployment guide

4. **Unit Tests Foundation**
   - Test coverage untuk critical paths (auth, task CRUD)
   - Integration tests untuk database operations

### Major Projects (High Impact, High Effort) - PRIORITAS KEDUA 🏗️

5. **Deployment Preparation**
   - Setup deployment scripts (Render/Railway/VPS)
   - SSL/TLS configuration
   - Monitoring & logging setup

6. **API Documentation Enhancement**
   - Detailed request/response examples
   - Error codes documentation
   - Postman collection

---

## 🔧 Technical Specification

### 1. CI/CD Pipeline (GitHub Actions)

**File**: `.github/workflows/ci.yml`

```yaml
name: GoTask Backend CI/CD

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - name: Run tests
        run: go test -v ./...
      - name: Build
        run: go build -o gotask-backend .

  docker:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build Docker image
        run: docker build -t gotask-backend:${{ github.sha }} .
```

**Acceptance Criteria**:
- [ ] Tests run automatically on push/PR
- [ ] Docker image built successfully
- [ ] No manual deployment needed for changes to main branch

### 2. Docker Production Ready

**Files**:
- `Dockerfile` (multi-stage)
- `docker-compose.yml` (existing, update for production)
- `.dockerignore`

**Dockerfile Requirements**:
- Multi-stage build (build → production)
- Non-root user for security
- Healthcheck included
- Minimal image size

**Acceptance Criteria**:
- [ ] `docker build -t gotask-backend .` succeeds
- [ ] Container starts and healthcheck passes
- [ ] All env vars loaded from .env

### 3. README Documentation

**File**: `README.md`

**Sections**:
1. Project Overview
2. Features
3. Tech Stack
4. Prerequisites
5. Quick Start
6. API Documentation (summary)
7. Deployment Guide
8. Contributing
9. License

**Acceptance Criteria**:
- [ ] README bisa dimengerti orang baru
- [ ] Setup instructions lengkap dan tested
- [ ] Deployment guide step-by-step

### 4. Unit Tests Foundation

**Target Files**:
- `modules/auth/*_test.go`
- `modules/tasks/*_test.go`
- `modules/projects/*_test.go`

**Test Coverage Target**: 70%+ untuk critical modules

**Acceptance Criteria**:
- [ ] Auth tests (signup, login, JWT validation)
- [ ] Task CRUD tests
- [ ] Project CRUD tests
- [ ] All tests pass: `go test ./...`

---

## 📁 Deliverables

| Deliverable | Location | Priority |
|------------|----------|----------|
| CI/CD Pipeline | `.github/workflows/ci.yml` | HIGH |
| Docker Production | `Dockerfile` | HIGH |
| README | `README.md` | HIGH |
| Unit Tests | `modules/**/*_test.go` | MEDIUM |
| Postman Collection | `docs/GoTask-API.postman_collection.json` | MEDIUM |
| Deployment Guide | `docs/DEPLOYMENT.md` | MEDIUM |

---

## 🔜 Next Steps (After Frontend Feedback)

1. **Phase 2: API Development**
   - Build missing API endpoints based on frontend needs
   - Refactor existing APIs if needed
   - Data migration planning (Firebase → PostgreSQL)

2. **Phase 3: Integration**
   - Connect frontend with backend
   - Test end-to-end flows
   - Fix any integration issues

3. **Phase 4: Deployment**
   - Deploy to production server
   - Setup monitoring
   - Go live

---

## 📝 Notes

- Frontend: React + Firebase MVP (sedang direview swagger.json)
- Backend: Go + PostgreSQL (working, local)
- Integration: Belum terhubung, menunggu API contract final

---

**Last Updated**: 2026-05-19
**Status**: Ready for Engineering Execution