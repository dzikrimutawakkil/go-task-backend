# GoTask Backend - Quick Setup

## Prasyarat
- Go 1.24+
- PostgreSQL running on port 5432

## Setup Steps

### 1. Setup Environment
```bash
cd gotask-backend
copy .env.example .env
```

### 2. Buat Database (jika belum ada)
```bash
PGPASSWORD=<your_db_password> psql -U postgres -h 127.0.0.1 -c "CREATE DATABASE gotask;"
```

### 3. Jalankan Aplikasi
```bash
go run main.go
```

## Verifikasi

### Cek Server Hidup
```bash
curl http://localhost:8080/health
# Response: {"status":"ok"}
```

### Buka Swagger API Docs
```
http://localhost:8080/swagger/index.html
```

## Test API (Signup & Login)

```bash
# 1. Signup
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# 2. Login (copy token dari response)
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# 3. Test Protected Endpoint (ganti <TOKEN> dengan token dari login)
curl http://localhost:8080/projects \
  -H "Authorization: Bearer <TOKEN>"
```

## Troubleshooting

**Database connection failed?**
```bash
# Pastikan PostgreSQL running
netstat -ano | findstr :5432
```

**Port 8080 sudah dipake?**
```bash
# Kill process di port 8080
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

**Migration error?**
```bash
# Reset database
PGPASSWORD=<your_db_password> psql -U postgres -h 127.0.0.1 -d gotask -c "DROP TABLE IF EXISTS task_users, tasks, statuses, projects, organization_users, project_users, organizations, priorities, labels, invitations, users, schema_migrations CASCADE;"
go run main.go
```