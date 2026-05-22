# GoTask Backend — Complete Setup Guide

Panduan lengkap untuk menjalankan backend dan menyimpannya ke frontend `freelance-os`.

---

## Table of Contents

1. [Quick Overview](#1-quick-overview)
2. [Backend Setup (Two Options)](#2-backend-setup-two-options)
3. [Frontend Setup: Connecting freelance-os](#3-frontend-setup-connecting-freelance-os)
4. [Verifikasi: Semua Endpoint Work](#4-verifikasi-semua-endpoint-work)
5. [API Reference Quick Ref](#5-api-reference-quick-ref)
6. [Troubleshooting](#6-troubleshooting)

---

## 1. Quick Overview

```
┌──────────────────────────┐
│   Frontend (freelance-os) │
│   Next.js App Router      │
│   Port: 3000 (default)     │
└──────────────┬─────────────┘
               │ REST API
               ▼
┌──────────────────────────┐
│   Backend (go-task-backend)│
│   Go + Gin + PostgreSQL     │
│   Port: 8080 (default)      │
└──────────────┬─────────────┘
               │ PostgreSQL
               ▼
┌──────────────────────────┐
│   PostgreSQL Database     │
│   Port: 5432 (default)    │
└──────────────────────────┘
```

**Yang sudah di-implementasi:**
- Auth (Signup, Login, Get Me, Forgot Password)
- Projects (CRUD + Get Single + Update)
- Tasks (CRUD + Search + Status/Label management)
- Clients (CRUD + Stats)
- Invoices (CRUD + Auto-invoice-number + Revenue sync)
- 30 total endpoints

---

## 2. Backend Setup (Two Options)

### Option A: Docker Compose (Recommended — 1 Command)

**Ini cara paling cepat. Semua dependencies (PostgreSQL + Backend) jalan dalam 1 command.**

#### Langkah 1: Edit Environment Variable

```bash
# Di folder go-task-backend/
copy .env.example .env
```

Edit `.env`:
```env
# Database
DB_HOST=db                    # Ganti dari localhost → db (untuk docker-compose)
DB_USER=postgres
DB_PASSWORD=rahasia
DB_NAME=gotaskdb
DB_PORT=5432

# JWT
SECRET_KEY=my_super_secret_key_at_least_32_chars

# App
APP_URL=http://localhost:8080
LOG_LEVEL=debug
```

#### Langkah 2: Jalankan

```bash
# Di folder go-task-backend/
docker-compose up --build -d
```

Tunggu ~30 detik, lalu cek:
```bash
curl http://localhost:8080/health
# Response: {"status":"ok"}
```

**That's it!** PostgreSQL + Backend sudah running. Lanjut ke [Bagian 3](#3-frontend-setup-connecting-freelance-os).

---

### Option B: Manual (Tanpa Docker)

#### Langkah 1: Install Dependencies

- **Go 1.24+**: https://go.dev/dl/
- **PostgreSQL 15+**: https://www.postgresql.org/download/

Cek install:
```bash
go version    # harus >= 1.24
psql --version
```

#### Langkah 2: Setup Environment

```bash
# Di folder go-task-backend/
copy .env.example .env
```

Edit `.env`:
```env
# Database
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=rahasia
DB_NAME=gotaskdb
DB_PORT=5432

# JWT
SECRET_KEY=my_super_secret_key_at_least_32_chars

# App
APP_URL=http://localhost:8080
LOG_LEVEL=debug
```

#### Langkah 3: Buat Database

```bash
# Login ke PostgreSQL
psql -U postgres -h 127.0.0.1

# Buat database (di dalam psql)
CREATE DATABASE gotaskdb;
\q
```

#### Langkah 4: Install Dependencies & Run

```bash
# Install Go dependencies
go mod tidy

# Run migrations + start server
go run main.go
```

Cek log — kalau muncul:
```
Database connected and migrated!
Starting server addr: :8080
```

Berarti success. Buka `http://localhost:8080/health` untuk verifikasi.

---

## 3. Frontend Setup: Connecting freelance-os

### Langkah 1: Set Environment Variable

Di folder `freelance-os/`, buat atau edit `.env.local`:

```bash
# Di folder freelance-os/
copy .env.example .env.local   # jika ada .env.example
```

Tambahkan/ubah:
```bash
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

> **Penting:** Port default Next.js adalah `3000`, jadi backend harus di port `8080`. Kalau mau ubah port frontend, pakai `PORT=3001 npm run dev`.

### Langkah 2: Verifikasi Connection

Jalankan frontend:
```bash
cd freelance-os
npm run dev
```

Buka `http://localhost:3000`. Kalau:
- Halaman login muncul → ✅ Berarti frontend berhasil connect ke backend
- Backend error 401/404 → Cek `NEXT_PUBLIC_API_BASE_URL` apakah sudah bener

### Langkah 3: Test Auth Flow

Buka browser DevTools (F12) → Network tab.

1. **Register** user baru:
   - Form register → Submit
   - Di Network tab, cari request ke `POST /signup`
   - Response harus contain `{ success: true, data: { user, token } }`
   - Token harus tersimpan di `localStorage` key `gotask_token`

2. **Login** dengan user yang sudah ada:
   - Form login → Submit
   - Response harus contain `{ success: true, data: { user, token } }`

3. **Reload page** — session harus tetap ada:
   - Cek localStorage `gotask_token` ada isinya
   - Page reload → tidak di-redirect ke `/login`

### Langkah 4: Test Projects, Tasks, Clients, Invoices

1. **Buat Project baru** (dari halaman Projects):
   - Form create project → Submit
   - Project harus muncul di list
   - Default statuses otomatis dibuat (Todo, On Progress, Done, Pending, Cancel)

2. **Buat Task** di dalam project:
   - Pilih project → Add task
   - Task harus muncul dengan status "Todo" (ID=1)

3. **Buat Client** baru:
   - Halaman Clients → Add client
   - Client harus muncul di list

4. **Buat Invoice**:
   - Halaman Invoices → Create invoice
   - Invoice number auto-generated format `INV-2026-XXX`
   - Invoice muncul di list

5. **Mark Invoice Paid**:
   - Ubah invoice status → "paid"
   - Buka client details → `totalRevenue` harus bertambah sesuai amount invoice

---

## 4. Verifikasi: Semua Endpoint Work

### 4a. Auth Endpoints

```bash
# Ganti <TOKEN> dengan token hasil signup/login

# 1. Signup → returns { user, token }
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"test@example.com","password":"password123","phone":"+6281234567890"}'
# Expected: { "success":true, "data":{ "user":{...}, "token":"eyJ..." } }

# 2. Login → returns { user, token }
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
# Expected: { "success":true, "data":{ "user":{...}, "token":"eyJ..." } }

# 3. Get Me → returns current user
curl http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer <TOKEN>"
# Expected: { "success":true, "data":{ "user":{ "id":1, "name":"John Doe", ... } } }

# 4. Forgot Password (always returns 200 to prevent email enumeration)
curl -X POST http://localhost:8080/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'
# Expected: { "success":true, "message":"If the email exists, a reset link has been sent" }
```

### 4b. Project Endpoints

```bash
# 5. List Projects
curl http://localhost:8080/projects \
  -H "Authorization: Bearer <TOKEN>"
# Expected: { "success":true, "data":[] }

# 6. Create Project
curl -X POST http://localhost:8080/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"name":"Website Redesign","description":"Complete redesign of company website","status":"in_progress","priority":"high"}'
# Expected: { "success":true, "data":{ "id":1, "name":"Website Redesign", ... } }
# Note: Default statuses (Todo, On Progress, Done, Pending, Cancel) automatically created!

# 7. Get Single Project
curl http://localhost:8080/projects/1 \
  -H "Authorization: Bearer <TOKEN>"
# Expected: { "success":true, "data":{ "id":1, "name":"Website Redesign", ... } }

# 8. Update Project
curl -X PATCH http://localhost:8080/projects/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"name":"Updated Name","progress":50}'
# Expected: { "success":true, "data":{ "id":1, "name":"Updated Name", "progress":50, ... } }

# 9. Delete Project
curl -X DELETE http://localhost:8080/projects/1 \
  -H "Authorization: Bearer <TOKEN>"
# Expected: { "success":true }
```

### 4c. Task Endpoints

```bash
# 10. List Tasks by Project
curl "http://localhost:8080/projects/1/tasks?page=1&limit=50" \
  -H "Authorization: Bearer <TOKEN>"
# Expected: { "success":true, "data":{ "tasks":[], "meta":{...} } }

# 11. Create Task
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"title":"Build homepage","description":"Create the landing page","project_id":1,"status_id":1,"priority_id":2}'
# Expected: { "success":true, "data":{ "id":1, "title":"Build homepage", "status":{ "name":"Todo" }, ... } }

# 12. Update Task (including description)
curl -X PATCH http://localhost:8080/tasks/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"title":"Updated task","description":"Updated description","status_id":2}'
# Expected: { "success":true, "data":{ "id":1, "title":"Updated task", "description":"Updated description", ... } }

# 13. Search Tasks
curl "http://localhost:8080/tasks/search?q=homepage&status_id=2" \
  -H "Authorization: Bearer <TOKEN>"
# Expected: { "success":true, "data":{ "tasks":[], "meta":{...} } }

# 14. Delete Task
curl -X DELETE http://localhost:8080/tasks/1 \
  -H "Authorization: Bearer <TOKEN>"
# Expected: { "success":true }
```

### 4d. Client Endpoints

```bash
# 15. List Clients
curl http://localhost:8080/clients \
  -H "Authorization: Bearer <TOKEN>"
# Expected: { "success":true, "data":{ "clients":[], "meta":{...} } }

# 16. Create Client
curl -X POST http://localhost:8080/clients \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"name":"Acme Corp","email":"contact@acme.com","whatsapp":"+6281234567890","company":"Acme Corp","website":"https://acme.com"}'
# Expected: { "success":true, "data":{ "id":1, "name":"Acme Corp", "total_revenue":0, ... } }

# 17. Get Client Stats
curl http://localhost:8080/clients/stats \
  -H "Authorization: Bearer <TOKEN>"
# Expected: { "success":true, "data":{ "total":1, "totalRevenue":0, "avgRevenue":0 } }
# Note: Returns flat object directly (no nested "stats" wrapper)

# 18. Update Client
curl -X PATCH http://localhost:8080/clients/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"email":"newemail@acme.com"}'
# Expected: { "success":true, "data":{ "id":1, "email":"newemail@acme.com", ... } }

# 19. Delete Client
curl -X DELETE http://localhost:8080/clients/1 \
  -H "Authorization: Bearer <TOKEN>"
# Expected: { "success":true }
```

### 4e. Invoice Endpoints

```bash
# 20. List Invoices
curl http://localhost:8080/invoices \
  -H "Authorization: Bearer <TOKEN>"
# Expected: { "success":true, "data":{ "invoices":[], "meta":{...} } }

# 21. Create Invoice (auto-generates invoice number)
curl -X POST http://localhost:8080/invoices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"client_id":1,"title":"Website Redesign","amount":5000000,"tax":500000,"discount":0,"due_date":"2026-06-30T00:00:00Z","notes":"Payment within 30 days","items":[{"description":"Web Development","quantity":1,"unit_price":5000000,"total":5000000}]}'
# Expected: { "success":true, "data":{ "id":1, "invoice_number":"INV-2026-ABC","status":"draft", ... } }
# Note: Invoice number format: INV-YYYY-XXX (e.g. INV-2026-X7K)

# 22. Update Invoice (mark as paid → triggers revenue sync)
curl -X PATCH http://localhost:8080/invoices/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"status":"paid"}'
# Expected: { "success":true, "data":{ "id":1, "status":"paid", "amount_paid":5000000, "paid_at":"2026-...", ... } }
# Note: When status becomes "paid", client's totalRevenue increases automatically!

# 23. Delete Invoice
curl -X DELETE http://localhost:8080/invoices/1 \
  -H "Authorization: Bearer <TOKEN>"
# Expected: { "success":true }
```

### 4f. Status & Label Endpoints

```bash
# 24. List Project Statuses
curl http://localhost:8080/projects/1/status \
  -H "Authorization: Bearer <TOKEN>"
# Expected: { "success":true, "data":[{ "id":1, "name":"Todo", "index":0, ... }, ...] }

# 25. Create Label
curl -X POST http://localhost:8080/projects/1/labels \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"name":"Bug","color":"#FF0000"}'
# Expected: { "success":true, "data":{ "id":1, "name":"Bug", "color":"#FF0000", ... } }
```

---

## 5. API Reference Quick Ref

### Base URL
```
http://localhost:8080
```

### Auth Headers
```
Authorization: Bearer <jwt_token>
```

### Response Format (Semua Endpoint)
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

### Key Points

| Item | Detail |
|---|---|
| **Organization ID** | Header `X-Organization-ID` **optional** — middleware auto-resolves ke personal org. Frontend TIDAK PERLU kirim header ini. |
| **Token Storage** | JWT disimpan sebagai `gotask_token` di `localStorage` |
| **Task Status** | Backend pakai `status_id` (numeric ID). Default per project: Todo(ID=1), On Progress(ID=2), Done(ID=3), Pending(ID=4), Cancel(ID=5) |
| **Task Priority** | Backend pakai `priority_id`. Seeded values: Low(ID=1), Medium(ID=2), High(ID=3), Urgent(ID=4) |
| **401 Handling** | Axios interceptor auto-redirect ke `/login` saat HTTP 401 |
| **Invoice Number** | Auto-generated format `INV-YYYY-XXX` (e.g. `INV-2026-M2P`). Unique per database. |
| **Revenue Sync** | Saat invoice `status` diubah ke `paid`, `total_revenue` client bertambah otomatis |

### Interactive API Docs
```
http://localhost:8080/swagger/index.html
```
Bisa coba langsung semua endpoint dari browser tanpa curl.

---

## 6. Troubleshooting

### ❌ "connection refused" saat curl

Backend belum jalan. Jalankan:
```bash
# Docker
docker-compose up -d

# Manual
go run main.go
```

### ❌ "401 Unauthorized"

1. Token belum diset atau expired
2. Cek localStorage:
   ```javascript
   // Browser console
   localStorage.getItem('gotask_token')
   ```
3. Kalau null → user belum login
4. Kalau ada → token corrupt atau server restart (token reset)

### ❌ "database connection failed"

1. PostgreSQL belum jalan:
   ```bash
   # Docker
   docker-compose ps

   # Manual - cek port 5432
   netstat -ano | findstr :5432
   ```

2. Credentials salah di `.env`:
   ```
   DB_HOST=localhost      # Docker: db
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=rahasia
   DB_NAME=gotaskdb
   ```

### ❌ "project not found" (404)

Kemungkinan:
1. Project sudah dihapus
2. Salah ID — coba GET `/projects` dulu untuk lihat list project
3. Project milik org lain — middleware auto-resolve ke personal org, jadi pastikan user punya project tersebut

### ❌ Frontend tidak bisa login

1. Cek Network tab di DevTools — request ke `/login` kirim ke mana?
2. Pastikan `NEXT_PUBLIC_API_BASE_URL=http://localhost:8080` (tanpa trailing slash)
3. Cek `.env.local` di folder `freelance-os`, bukan di folder `go-task-backend`
4. `npm run dev` jalan di port berapa? Default `3000`

### ❌ Task description tidak kesimpan

Pastikan field name adalah `description` (lowercase):
```json
{
  "description": "Task description text"
}
```

Backend `UpdateTaskRequest` sudah support `description`. Pastikan frontend kirim field name yang benar.

### 🔄 Reset Everything (Clean Start)

Kalau mau mulai dari 0:

```bash
# 1. Stop everything
docker-compose down        # Docker
# atau
# kill proses go run

# 2. Hapus database volumes (Docker)
docker-compose down -v

# 3. Hapus database manual
psql -U postgres -h 127.0.0.1 -d gotaskdb -c "DROP DATABASE gotaskdb;"
psql -U postgres -h 127.0.0.1 -c "CREATE DATABASE gotaskdb;"

# 4. Restart
docker-compose up --build -d    # Docker
# atau
go run main.go                 # Manual
```

### ✅ Quick Verification Checklist

| Test | Command | Expected |
|---|---|---|
| Backend running? | `curl http://localhost:8080/health` | `{"status":"ok"}` |
| DB connected? | Cek logs saat `go run main.go` | "Database connected and migrated!" |
| Signup works? | `curl -X POST /signup ...` | `{ "success":true, "data":{ "user", "token" } }` |
| Swagger available? | Buka `/swagger/index.html` | API docs page |
| Frontend connected? | Buka `http://localhost:3000` | App loaded |
| Projects CRUD? | Create then GET | Project appears in list |
| Invoices paid → revenue? | Mark paid → GET /clients/stats | `totalRevenue` increased |

---

## 📞 Kalau Still Stuck

1. **Cek logs backend** — semua error akan keliatan di terminal tempat backend jalan
2. **Cek Network tab** di browser DevTools — lihat request/response JSON
3. **Swagger UI** — coba endpoint langsung dari `http://localhost:8080/swagger/index.html`
4. **Cek `.env`** — semua variabel harus bener antara backend dan frontend