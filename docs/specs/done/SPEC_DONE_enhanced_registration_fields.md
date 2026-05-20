# ACTIVE_SPEC: Enhanced Registration Fields

## 📋 Overview

Menambah field registrasi user untuk mendukung collaborative task management dan kebutuhan bisnis (contingency contact, WhatsApp community).

## 🎯 Goals

1. **Collaborative Branding** — Tampilkan nama user di task assignment, @mentions, comments
2. **Contingency Contact** — Nomor telpon untuk contingency contact + future WhatsApp community
3. **Record Keeping** — Alamat opsional untuk dokumentasi

## 📦 Scope

### Registration Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | ✅ Yes | Full name untuk collaboration |
| `email` | string | ✅ Yes | Email untuk autentikasi |
| `password` | string | ✅ Yes | Password untuk keamanan |
| `phone` | string | ✅ Yes | Nomor telpon untuk contingency + WhatsApp |
| `address` | string | ❌ No | Alamat opsional untuk dokumentasi |

### Files to Modify

- `modules/auth/auth_handler.go` — Tambah field di `SignupRequest`
- `modules/auth/auth_service.go` — Update `SignupInput` struct
- `modules/auth/auth_models.go` — Tambah field di `User` model
- `modules/auth/auth_repository.go` — Update create logic

## 🔄 Changes

### 1. `auth_models.go` — User Model
```go
type User struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Email     string    `gorm:"unique" json:"email"`
    Name      string    `json:"name"`                    // BARU
    Phone     string    `json:"phone"`                  // BARU
    Address   string    `json:"address"`                // BARU (optional)
    Password  string    `json:"-"`
    CreatedAt time.Time `json:"created_at"`
}
```

### 2. `auth_handler.go` — SignupRequest
```go
type SignupRequest struct {
    Email    string `json:"email" binding:"required"`
    Password string `json:"password" binding:"required"`
    Name     string `json:"name" binding:"required"`     // BARU
    Phone    string `json:"phone" binding:"required"`    // BARU
    Address  string `json:"address"`                      // BARU (optional)
}
```

### 3. `auth_service.go` — SignupInput & Logic
```go
type SignupInput struct {
    Email    string
    Password string
    Name     string  // BARU
    Phone    string  // BARU
    Address  string  // BARU (optional)
}
```

## ✅ Acceptance Criteria

- [ ] Signup menerima `name`, `phone`, `address`
- [ ] `name` dan `phone` wajib diisi (validation)
- [ ] `address` opsional (tidak ada validation error jika kosong)
- [ ] Response user object mengembalikan semua field baru
- [ ] Swagger docs ter-update otomatis

## 🧪 Test Scenarios

1. **Signup with all fields** — Success, semua field tersimpan
2. **Signup without name** — Error 400, "name is required"
3. **Signup without phone** — Error 400, "phone is required"
4. **Signup without address** — Success, address = empty string
5. **Login returns user with all fields** — Verify data integrity

## 📊 Effort Estimate

| Task | Effort |
|------|--------|
| Modify User model | 15 min |
| Update SignupRequest struct | 15 min |
| Update SignupInput & service | 30 min |
| Update repository | 15 min |
| Generate Swagger | 10 min |
| Test | 30 min |
| **Total** | **~2 jam** |

## 📅 Status

**Started:** 2026-05-20
**Status:** 🚀 In Progress
**Quadrant:** Quick Win (High Impact, Low Effort)