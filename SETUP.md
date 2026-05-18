# 🚀 GoTask Backend - Setup Guide

Panduan lengkap untuk menjalankan GoTask Backend dari awal (setup database hingga test Swagger API).

---

## 📋 Prasyarat

Pastikan komputer kamu sudah install:

| Tools | Version | Cara Install |
|-------|---------|-------------|
| Go | 1.24.0+ | [go.dev/dl](https://go.dev/dl) |
| Docker | latest | [docker.com/get-started](https://docker.com/get-started) |
| Docker Compose | v2+ | Sudah termasuk di Docker Desktop |

---

## 🎯 Langkah 1: Copy Environment File

Copy file `.env.example` menjadi `.env`:

```bash
# Dari folder project
copy .env.example .env
```

Atau buat manual dengan isi:

```env
# DATABASE
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gotaskdb
DB_PORT=5432

# JWT AUTHENTICATION
SECRET_KEY=ganti_dengan_secret_key_rahasia_kamu

# MIGRATIONS
MIGRATION_URL=
MIGRATIONS_PATH=

# APPLICATION
APP_URL=http://localhost:8080
LOG_LEVEL=info

# EMAIL / SMTP (optional untuk invite email)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=
INVITE_EXPIRY_HOURS=168
```

---

## 🐳 Langkah 2: Setup Database dengan Docker

### Opsi A: Start Database dengan Docker Compose (Recommended)

```bash
# Start PostgreSQL saja (tanpa build app)
docker-compose up db -d

# Cek apakah container running
docker ps | grep postgres
```

### Opsi B: Start Semua (App + Database)

```bash
# Build dan start semua service
docker-compose up --build

# Atau di background
docker-compose up --build -d
```

### Opsi C: Local Development (Tanpa Docker)

Kalau mau jalanin app langsung di local (tanpa container):

1. Install PostgreSQL 15 di komputer kamu
2. Create database:
   ```sql
   CREATE DATABASE gotaskdb;
   CREATE USER postgres WITH PASSWORD 'postgres';
   ```
3. Update `.env`:
   ```env
   DB_HOST=localhost
   DB_PASSWORD=postgres
   ```

---

## 🔧 Langkah 3: Setup Database (Create Database)

Sebelum jalanin app, database `gotaskdb` harus sudah ada:

### Kalau pakai Docker Compose (otomatis dibuat):
Sudah dihandle sama `docker-compose.yml` - database otomatis dibuat saat container start.

### Kalau setup manual di local:
```bash
# Login ke PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE gotaskdb;

# Keluar
\q
```

---

## ▶️ Langkah 4: Jalankan Aplikasi

### Development Mode (Tanpa Docker):

```bash
go run main.go
```

### Dengan Docker Compose:

```bash
docker-compose up --build
```

### Tanpa Docker (Langsung Go):

```bash
go run main.go
```

---

## ✅ Langkah 5: Verifikasi Server Running

Cek apakah server sudah jalan:

```bash
# Liveness check
curl http://localhost:8080/health

# Response harusnya:
# {"status":"ok"}
```

```bash
# Readiness check (cek DB connection)
curl http://localhost:8080/ready

# Response kalau sukses:
# {"status":"ready","db":"connected"}
```

---

## 📖 Langkah 6: Test Swagger API

Buka browser dan visit:

```
http://localhost:8080/swagger/index.html
```

Kamu akan lihat Swagger UI dengan semua endpoint terdokumentasi.

### Cara Test Endpoint dengan JWT:

1. **Signup** (buat akun baru):
   - Expand `POST /signup`
   - Klik **"Try it out"**
   - Isi `email` dan `password`
   - Klik **"Execute"**
   - Response: `{ "success": true, "data": { "user": {...} } }`

2. **Login** (dapat JWT token):
   - Expand `POST /login`
   - Klik **"Try it out"**
   - Isi email dan password yang sama
   - Klik **"Execute"**
   - Copy token dari response: `data.token`

3. **Authorize JWT Token di Swagger**:
   - Klik tombol **"Authorize"** (lock icon) di atas
   - Isi: `Bearer <token_yang_di_copy>`
   - Klik **"Authorize"** → **"Close"**

4. **Test Endpoint Protected**:
   - Sekarang semua endpoint yang butuh auth bisa di-test!
   - Contoh: `POST /organizations` - untuk bikin organization baru

---

## 🎨 Langkah 7: Test Flow Lengkap

### 1. Buat User & Login
```bash
# Signup
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
# Copy token dari response
```

### 2. Create Organization
```bash
curl -X POST http://localhost:8080/organizations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -H "X-Organization-ID: 1" \
  -d '{"name":"Tim Alpha"}'
```

### 3. Create Project
```bash
curl -X POST http://localhost:8080/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -H "X-Organization-ID: 1" \
  -d '{"name":"Website Redesign","description":"Redesign company website"}'
```

### 4. Create Task
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -H "X-Organization-ID: 1" \
  -d '{
    "title":"Design landing page",
    "project_id":1,
    "priority_id":2
  }'
```

---

## 🛠️ Troubleshooting

### Problem: "Failed to connect to database!"

**Solusi:**
```bash
# Cek apakah PostgreSQL container running
docker ps | grep postgres

# Kalau gak running, start ulang
docker-compose up db -d

# Tunggu 5 detik, baru jalanin app
go run main.go
```

### Problem: Port 8080 sudah dipake

**Solusi:**
```bash
# Cari process yang pake port 8080
netstat -ano | findstr :8080

# Kill process (ganti PID dengan angka yang muncul)
taskkill /PID <PID> /F

# Atau pakai Docker port lain, edit docker-compose.yml
ports:
  - "3000:8080"  # Akses di http://localhost:3000
```

### Problem: "No .env file found"

**Solusi:**
```bash
# Pastikan .env ada di root folder
dir .env

# Kalau gak ada, copy dari example
copy .env.example .env
```

### Problem: Migration failed

**Solusi:**
```bash
# Drop dan recreate database
docker-compose down -v   # Hapus semua data!
docker-compose up db -d   # Start ulang DB
go run main.go           # Run ulang app (migration otomatis)
```

### Problem: Swagger gak muncul

**Solusi:**
```bash
# Cek endpoint JSON spec
curl http://localhost:8080/swagger/doc.json

# Kalau error, regenerate docs
go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -o docs --parseDependency
go run main.go
```

---

## 📁 Struktur Project

```
go-task-backend/
├── config/
│   └── db.go           # Koneksi DB & migrations
├── docs/               # Swagger docs (auto-generated)
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── handlers/
│   └── health.go       # /health, /ready endpoints
├── middlewares/        # Auth, CORS, Rate limiting
├── modules/
│   ├── auth/           # Signup, Login
│   ├── organizations/   # Organization management
│   ├── projects/       # Project management
│   └── tasks/          # Task, Status, Label management
├── utils/
│   └── response.go     # Standard API response
├── docker-compose.yml  # Docker services config
├── Dockerfile          # App container config
├── main.go            # Entry point
└── .env               # Environment variables (DIJAGA RAHASIA!)
```

---

## 🔐 Catatan Keamanan

- **JANGAN** commit `.env` ke GitHub! Already ada di `.gitignore`
- **JANGAN** hardcode secret/password di source code
- **GANTI** `SECRET_KEY` dengan nilai yang kuat di production

---

## 🎉 Selamat! 🎉

Kamu sudah berhasil setup dan jalanin GoTask Backend!

下一步? Buka Swagger UI dan mulai test semua endpoint:

```
http://localhost:8080/swagger/index.html
```