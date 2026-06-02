# ACTIVE_SPEC.md — Subscription Tier System (Manual Activation)

## 📋 Overview

**Task:** M5 — Subscription Tier dengan Manual Activation
**Status:** ⏳ Pending
**Effort:** 4-6 jam
**Goal:** Implementasi sistem tier (Free/Pro/Ultimate) per-user dengan aktivasi manual oleh admin — tanpa Stripe, tanpa payment gateway. Admin mengaktifkan tier setelah konfirmasi pembayaran manual (transfer bank, QRIS, dll).

---

## 🔑 Decision Summary (Final)

| Decision | Value |
|---|---|
| **Tier binding** | Per-user (bukan per-org) |
| **Upgrade scope** | ALL orgs yang user adalah owner-nya |
| **Member benefit** | Member yang di-invite menikmati fitur premium org tsb |
| **Tier names** | Free, Pro, Ultimate (bukan Team) |

---

## 🔄 Current State vs Target State

| Aspect | Current | Target |
|---|---|---|
| **License system** | License key + expiry per user (modules/licenses) | Tier per user (free/pro/ultimate) + expiry |
| **Tier scope** | Tidak ada | Per-user — semua org owner-nya inherit tier tsb |
| **Activation** | Admin generate license key | Admin PATCH tier ke user |
| **Quota enforcement** | Tidak ada | Hard limit sebelum `Create` di service layer |
| **Response warning** | `license_warning` di semua response | `tier_info` menggantikan `license_warning` |
| **Pricing** | Hardcoded | Database-driven (`tier_plans` + `tier_limits` tables) |

---

## 🎯 Tier Definition

| Fitur | Free | Pro | Ultimate |
|---|---|---|---|
| **Workspace** | 1 (personal only) | 2 (personal + 1 team) | 4 (personal + 3 team) |
| **Project per workspace** | 3 | Unlimited | Unlimited |
| **Task per project** | 50 | Unlimited | Unlimited |
| **Member per workspace** | 1 (solo) | 3 | 15 |
| **Clients** | 5 | Unlimited | Unlimited |
| **Invoices per bulan** | 10 | Unlimited | Unlimited |
| **Comments & Notifikasi** | ❌ | ✅ | ✅ |
| **Real-time (SSE)** | ❌ | ✅ | ✅ |
| **Audit Log** | ❌ | ❌ | ✅ |

> **Catatan:**
> - Workspace limit = MAKS org yang bisa dibuat/dimiliki user
> - Personal org (auto-created saat register) TETAP ada di semua tier
> - Member yang di-invite ke org upgraded menikmati fitur premium org tsb
> - Comments, Notifikasi, SSE, Audit Log di-gate di tier (implementasi ada di spec terpisah: M1, M2, M4, M10)

---

## 🎯 User Flow

```
User A (Free) register:
├── Auto-create Org Personal (Free)
└── Limit: hanya 1 org, tidak bisa buat org baru

User A upgraded to Pro:
├── Org Personal → Pro ✅ (benefit: unlimited projects, tasks, dll)
├── ✅ Boleh buat Org Team #1 (Pro) ← tambahan 1 workspace
└── Invite User B, C ke Org Personal → User B, C nikmati fitur Pro di org tsb

User A upgraded to Ultimate:
├── Org Personal → Ultimate ✅
├── Org Team #1 → Ultimate ✅
├── ✅ Boleh buat Org Team #2, Team #3 (Ultimate) ← tambahan 3 workspace
└── Total: 4 orgs (1 personal + 3 team)
```

---

## 🎯 Technical Scope

### 1. Database Migration

**HAPUS:**
- `modules/licenses/` (entire folder — belum release, belum ada user)
- Table `licenses` (jika ada)

**Tambah kolom ke tabel `users`:**

```sql
ALTER TABLE users
  ADD COLUMN tier VARCHAR(20) NOT NULL DEFAULT 'free',
  ADD COLUMN tier_expires_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN tier_activated_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN tier_activated_by BIGINT REFERENCES users(id);
```

**Buat table `tier_plans`:**

```sql
CREATE TABLE tier_plans (
  id SERIAL PRIMARY KEY,
  tier VARCHAR(20) NOT NULL UNIQUE, -- free, pro, ultimate
  name VARCHAR(50) NOT NULL,
  description TEXT,
  price_monthly INTEGER NOT NULL DEFAULT 0, -- dalam Rupiah
  price_yearly INTEGER NOT NULL DEFAULT 0,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Buat table `tier_limits`:**

```sql
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

**Seeder untuk default tier plans & limits:**

```sql
INSERT INTO tier_plans (tier, name, description, price_monthly, price_yearly) VALUES
  ('free', 'Free', 'Untuk freelancer solo yang baru mulai', 0, 0),
  ('pro', 'Pro', 'Untuk freelancer yang berkembang dengan tim kecil', 79000, 790000),
  ('ultimate', 'Ultimate', 'Untuk agensi dan komunitas dengan tim besar', 199000, 1990000);

INSERT INTO tier_limits (tier, max_workspaces, max_projects, max_tasks_per_project, max_members, max_clients, max_invoices_per_month, can_comment, can_sse, can_audit_log) VALUES
  ('free', 1, 3, 50, 1, 5, 10, false, false, false),
  ('pro', 2, -1, -1, 3, -1, -1, true, true, false),
  ('ultimate', 4, -1, -1, 15, -1, -1, true, true, true);
```

**File migration:**

```
migrations/XXXX_add_tier_to_users.up.sql
migrations/XXXX_add_tier_to_users.down.sql
migrations/XXXX_create_tier_tables.up.sql
migrations/XXXX_create_tier_tables.down.sql
```

---

### 2. Hapus Modules Licenses

**Hapus entire folder:** `modules/licenses/`
- `modules/licenses/license_model.go`
- `modules/licenses/license_repository.go`
- `modules/licenses/license_service.go`
- `modules/licenses/license_handler.go`

**Hapus routes dari `main.go`** yang merefer ke license handlers.

---

### 3. Model Update

**File:** `modules/auth/auth_models.go` — tambah field tier ke User struct

```go
type User struct {
    // ... existing fields ...
    Tier            string     `json:"tier" gorm:"default:free"`
    TierExpiresAt   *time.Time `json:"tier_expires_at"`
    TierActivatedAt *time.Time `json:"tier_activated_at"`
    TierActivatedBy *uint      `json:"tier_activated_by"`
}
```

**File baru:** `modules/organizations/tier_plan_model.go`

```go
type TierPlan struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    Tier         string    `gorm:"unique;column:tier" json:"tier"`
    Name         string    `json:"name"`
    Description  string    `json:"description"`
    PriceMonthly int       `json:"price_monthly"`
    PriceYearly  int       `json:"price_yearly"`
    IsActive     bool      `json:"is_active"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

type TierLimit struct {
    ID                  uint   `gorm:"primaryKey" json:"id"`
    Tier                string `gorm:"unique;column:tier" json:"tier"`
    MaxWorkspaces       int    `json:"max_workspaces"`
    MaxProjects         int    `json:"max_projects"`
    MaxTasksPerProject  int    `json:"max_tasks_per_project"`
    MaxMembers          int    `json:"max_members"`
    MaxClients          int    `json:"max_clients"`
    MaxInvoicesPerMonth int    `json:"max_invoices_per_month"`
    CanComment          bool   `json:"can_comment"`
    CanSSE              bool   `json:"can_sse"`
    CanAuditLog         bool   `json:"can_audit_log"`
}
```

---

### 4. Quota Helper

**File:** `utils/quota.go`

```go
package utils

import (
    "time"
)

// TierLimits represents quota limits for each tier
type TierLimits struct {
    MaxWorkspaces       int
    MaxProjects         int
    MaxTasksPerProject  int
    MaxMembers          int
    MaxClients          int
    MaxInvoicesPerMonth int
    CanComment          bool
    CanSSE              bool
    CanAuditLog         bool
}

// GetTierLimits returns default limits (fallback if DB unavailable)
func GetTierLimits(tier string) TierLimits {
    defaults := map[string]TierLimits{
        "free": {
            MaxWorkspaces: 1, MaxProjects: 3, MaxTasksPerProject: 50,
            MaxMembers: 1, MaxClients: 5, MaxInvoicesPerMonth: 10,
            CanComment: false, CanSSE: false, CanAuditLog: false,
        },
        "pro": {
            MaxWorkspaces: 2, MaxProjects: -1, MaxTasksPerProject: -1,
            MaxMembers: 3, MaxClients: -1, MaxInvoicesPerMonth: -1,
            CanComment: true, CanSSE: true, CanAuditLog: false,
        },
        "ultimate": {
            MaxWorkspaces: 4, MaxProjects: -1, MaxTasksPerProject: -1,
            MaxMembers: 15, MaxClients: -1, MaxInvoicesPerMonth: -1,
            CanComment: true, CanSSE: true, CanAuditLog: true,
        },
    }

    if limits, ok := defaults[tier]; ok {
        return limits
    }
    return defaults["free"]
}

// IsTierActive checks if tier is still valid (not expired)
func IsTierActive(tier string, expiresAt *time.Time) bool {
    if tier == "free" {
        return true // free never expires
    }
    if expiresAt == nil {
        return false
    }
    return time.Now().Before(*expiresAt)
}

// GetEffectiveTier returns actual tier (fallback to free if expired)
func GetEffectiveTier(tier string, expiresAt *time.Time) string {
    if IsTierActive(tier, expiresAt) {
        return tier
    }
    return "free"
}

// QuotaError for tier limit exceeded
type QuotaError struct {
    Resource    string
    Limit       int
    CurrentTier string
}

func (e *QuotaError) Error() string {
    return fmt.Sprintf("quota exceeded: %s limit is %d on %s tier. Please upgrade.", e.Resource, e.Limit, e.CurrentTier)
}

func ErrQuotaExceeded(resource string, limit int, tier string) *QuotaError {
    return &QuotaError{Resource: resource, Limit: limit, CurrentTier: tier}
}
```

---

### 5. Quota Check di Service Layer

**Tambah quota check sebelum Create di setiap service yang relevan.**

**A. Organization Service (CreateOrganization):**

```go
func (s *organizationService) CreateOrganization(name string, ownerID uint) (*Organization, error) {
    // Check workspace limit per user
    user, err := s.authService.FindByID(ownerID)
    if err != nil {
        return nil, err
    }

    effectiveTier := utils.GetEffectiveTier(user.Tier, user.TierExpiresAt)
    limits := utils.GetTierLimits(effectiveTier)

    // Count user's owned orgs
    ownedOrgs, err := s.repo.CountByOwner(ownerID)
    if err != nil {
        return nil, err
    }

    if ownedOrgs >= limits.MaxWorkspaces {
        return nil, utils.ErrQuotaExceeded("workspace", limits.MaxWorkspaces, effectiveTier)
    }

    // Create org with inherited tier
    org := Organization{
        Name:    name,
        OwnerID: ownerID,
        OrgType: OrgTypeTeam, // manual created = team org
    }
    // Note: tier is derived from user, not stored in org
    // All orgs owned by the user inherit the user's tier

    if err := s.repo.Create(&org); err != nil {
        return nil, err
    }

    // Add owner as member
    if err := s.repo.AddMember(org.ID, ownerID, models.RoleOwner); err != nil {
        return nil, err
    }

    return &org, nil
}
```

**B. Project Service (CreateProject):**

```go
func (s *ProjectService) CreateProject(orgID uint, req CreateProjectRequest) (*Project, error) {
    // Get org (to find owner)
    org, err := s.orgRepo.FindByID(orgID)
    if err != nil {
        return nil, err
    }

    // Get user's tier (owner inherits tier)
    user, err := s.userRepo.FindByID(org.OwnerID)
    if err != nil {
        return nil, err
    }

    effectiveTier := utils.GetEffectiveTier(user.Tier, user.TierExpiresAt)
    limits := utils.GetTierLimits(effectiveTier)

    // Check project limit per org
    if limits.MaxProjects != -1 {
        count, err := s.projectRepo.CountByOrg(orgID)
        if err != nil {
            return nil, err
        }
        if count >= limits.MaxProjects {
            return nil, utils.ErrQuotaExceeded("project", limits.MaxProjects, effectiveTier)
        }
    }

    // ... proceed with create ...
}
```

**C. Task Service (CreateTask):**

```go
func (s *TaskService) CreateTask(orgID uint, req CreateTaskRequest) (*Task, error) {
    // Get user's tier from org owner
    org, err := s.orgRepo.FindByID(orgID)
    if err != nil {
        return nil, err
    }

    user, err := s.userRepo.FindByID(org.OwnerID)
    if err != nil {
        return nil, err
    }

    effectiveTier := utils.GetEffectiveTier(user.Tier, user.TierExpiresAt)
    limits := utils.GetTierLimits(effectiveTier)

    // Check task limit per project
    if limits.MaxTasksPerProject != -1 {
        count, err := s.taskRepo.CountByProject(req.ProjectID)
        if err != nil {
            return nil, err
        }
        if count >= limits.MaxTasksPerProject {
            return nil, utils.ErrQuotaExceeded("task", limits.MaxTasksPerProject, effectiveTier)
        }
    }

    // ... proceed with create ...
}
```

**D. Organization Service (InviteMember):**

```go
func (s *organizationService) InviteMember(orgID uint, email string, requesterID uint) error {
    // ... existing permission checks ...

    // Check member limit per org
    org, err := s.repo.FindByID(orgID)
    if err != nil {
        return err
    }

    user, err := s.authService.FindByID(org.OwnerID)
    if err != nil {
        return err
    }

    effectiveTier := utils.GetEffectiveTier(user.Tier, user.TierExpiresAt)
    limits := utils.GetTierLimits(effectiveTier)

    if limits.MaxMembers != -1 {
        count, err := s.repo.CountMembers(orgID)
        if err != nil {
            return err
        }
        if count >= limits.MaxMembers {
            return utils.ErrQuotaExceeded("member", limits.MaxMembers, effectiveTier)
        }
    }

    // ... proceed with invite ...
}
```

**E. Client Service (CreateClient):**

```go
func (s *ClientService) CreateClient(orgID uint, req CreateClientRequest) (*Client, error) {
    // Get user's tier
    org, err := s.orgRepo.FindByID(orgID)
    if err != nil {
        return nil, err
    }

    user, err := s.userRepo.FindByID(org.OwnerID)
    if err != nil {
        return nil, err
    }

    effectiveTier := utils.GetEffectiveTier(user.Tier, user.TierExpiresAt)
    limits := utils.GetTierLimits(effectiveTier)

    if limits.MaxClients != -1 {
        count, err := s.clientRepo.CountByOrg(orgID)
        if err != nil {
            return nil, err
        }
        if count >= limits.MaxClients {
            return nil, utils.ErrQuotaExceeded("client", limits.MaxClients, effectiveTier)
        }
    }

    // ... proceed with create ...
}
```

**F. Invoice Service (CreateInvoice):**

```go
func (s *InvoiceService) CreateInvoice(orgID uint, req CreateInvoiceRequest) (*Invoice, error) {
    // Get user's tier
    org, err := s.orgRepo.FindByID(orgID)
    if err != nil {
        return nil, err
    }

    user, err := s.userRepo.FindByID(org.OwnerID)
    if err != nil {
        return nil, err
    }

    effectiveTier := utils.GetEffectiveTier(user.Tier, user.TierExpiresAt)
    limits := utils.GetTierLimits(effectiveTier)

    if limits.MaxInvoicesPerMonth != -1 {
        count, err := s.invoiceRepo.CountThisMonth(orgID)
        if err != nil {
            return nil, err
        }
        if count >= limits.MaxInvoicesPerMonth {
            return nil, utils.ErrQuotaExceeded("invoice", limits.MaxInvoicesPerMonth, effectiveTier)
        }
    }

    // ... proceed with create ...
}
```

**Handler error handling:**

```go
if quotaErr, ok := err.(*utils.QuotaError); ok {
    utils.SendError(c, http.StatusForbidden, quotaErr.Error())
    return
}
```

---

### 6. Public Endpoint: List Tier Plans

**Endpoint:**
```
GET /tier/plans
```

**Middleware:** Public — tidak perlu auth. Untuk halaman pricing di frontend.

**Response:**
```json
{
  "success": true,
  "message": "OK",
  "data": [
    {
      "tier": "free",
      "name": "Free",
      "description": "Untuk freelancer solo yang baru mulai",
      "price_monthly": 0,
      "price_yearly": 0,
      "limits": {
        "max_workspaces": 1,
        "max_projects_per_workspace": 3,
        "max_tasks_per_project": 50,
        "max_members_per_workspace": 1,
        "max_clients": 5,
        "max_invoices_per_month": 10
      },
      "features": {
        "comments": false,
        "realtime": false,
        "audit_log": false
      }
    },
    {
      "tier": "pro",
      "name": "Pro",
      "description": "Untuk freelancer yang berkembang dengan tim kecil",
      "price_monthly": 79000,
      "price_yearly": 790000,
      "limits": {
        "max_workspaces": 2,
        "max_projects_per_workspace": -1,
        "max_tasks_per_project": -1,
        "max_members_per_workspace": 3,
        "max_clients": -1,
        "max_invoices_per_month": -1
      },
      "features": {
        "comments": true,
        "realtime": true,
        "audit_log": false
      }
    },
    {
      "tier": "ultimate",
      "name": "Ultimate",
      "description": "Untuk agensi dan komunitas dengan tim besar",
      "price_monthly": 199000,
      "price_yearly": 1990000,
      "limits": {
        "max_workspaces": 4,
        "max_projects_per_workspace": -1,
        "max_tasks_per_project": -1,
        "max_members_per_workspace": 15,
        "max_clients": -1,
        "max_invoices_per_month": -1
      },
      "features": {
        "comments": true,
        "realtime": true,
        "audit_log": true
      }
    }
  ]
}
```

---

### 7. User Endpoint: Get My Tier Info

**Endpoint:**
```
GET /users/me/tier
```

**Middleware:** `RequireAuth`

**Response (tier aktif):**
```json
{
  "success": true,
  "message": "OK",
  "data": {
    "tier": "pro",
    "effective_tier": "pro",
    "is_active": true,
    "activated_at": "2026-05-28T00:00:00Z",
    "expires_at": "2027-05-28T00:00:00Z",
    "days_remaining": 365,
    "limits": {
      "max_workspaces": 2,
      "max_projects_per_workspace": -1,
      "max_tasks_per_project": -1,
      "max_members_per_workspace": 3,
      "max_clients": -1,
      "max_invoices_per_month": -1
    },
    "features": {
      "comments": true,
      "realtime": true,
      "audit_log": false
    },
    "usage": {
      "owned_workspaces": 1,
      "projects": 7,
      "members": 2,
      "clients": 12,
      "invoices_this_month": 4
    }
  }
}
```

**Response (tier expired, fallback ke free):**
```json
{
  "success": true,
  "message": "OK",
  "data": {
    "tier": "pro",
    "effective_tier": "free",
    "is_active": false,
    "activated_at": "2026-05-28T00:00:00Z",
    "expires_at": "2026-05-20T00:00:00Z",
    "days_remaining": -8,
    "limits": { ...free limits... },
    "features": { ...free features... },
    "usage": { ... }
  }
}
```

---

### 8. Admin Endpoint: Activate Tier

**Endpoint:**
```
PATCH /admin/users/:id/tier
```

**Middleware:** `RequireAuth` + Admin only

**Request body:**
```json
{
  "tier": "pro",
  "duration_months": 12
}
```

**Response:**
```json
{
  "success": true,
  "message": "Tier activated successfully",
  "data": {
    "user_id": 5,
    "tier": "pro",
    "tier_expires_at": "2027-05-28T00:00:00Z",
    "tier_activated_at": "2026-05-28T00:00:00Z",
    "affected_organizations": ["Personal Workspace", "Team Workspace #1"]
  }
}
```

**Validasi:**
- `tier` harus salah satu dari: `free`, `pro`, `ultimate`
- `duration_months` harus antara 1–24
- Jika user sudah punya tier aktif, `tier_expires_at` di-extend (bukan di-reset)

**Logic extend:**
```go
startFrom := time.Now()
if user.TierExpiresAt != nil && user.TierExpiresAt.After(time.Now()) {
    startFrom = *user.TierExpiresAt
}
newExpiry := startFrom.AddDate(0, req.DurationMonths, 0)
```

---

### 9. Update Response: `tier_info`

**File:** `utils/response.go`

```go
// APIResponse represents the standard API response envelope.
type APIResponse struct {
    Success  bool      `json:"success"`
    Message  string    `json:"message"`
    Data     interface{} `json:"data,omitempty"`
    TierInfo *TierInfo `json:"tier_info,omitempty"`
}

// TierInfo represents tier status in the API response
type TierInfo struct {
    Tier          string `json:"tier"`
    IsActive      bool   `json:"is_active"`
    ExpiresAt     string `json:"expires_at,omitempty"`
    DaysRemaining int    `json:"days_remaining"`
    Warning       string `json:"warning,omitempty"`
}
```

**Middleware update** (`middlewares/require_auth.go`):
- Hapus `setLicenseWarning` + `license_warning` context
- Ganti dengan `setTierInfo` yang baca dari user tier fields

---

### 10. Feature Gate Middleware

**File:** `middlewares/require_tier.go` (baru)

```go
func RequireTierFeature(feature string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := getUserFromContext(c) // dari RequireAuth middleware
        effectiveTier := utils.GetEffectiveTier(user.Tier, user.TierExpiresAt)
        limits := utils.GetTierLimits(effectiveTier)

        allowed := false
        switch feature {
        case "comment":
            allowed = limits.CanComment
        case "sse":
            allowed = limits.CanSSE
        case "audit_log":
            allowed = limits.CanAuditLog
        }

        if !allowed {
            utils.SendError(c, http.StatusForbidden, "This feature requires Pro or Ultimate tier.")
            return
        }
        c.Next()
    }
}
```

**Contoh pemakaian:**
```go
tasks.POST("/:id/comments", RequireTierFeature("comment"), commentHandler.Create)
```

---

## ✅ Acceptance Criteria

| # | Criteria | Test Method |
|---|---|---|
| AC1 | User baru auto-dapat tier `free` saat register | Check DB setelah signup |
| AC2 | Create org ke-2 di Free tier → HTTP 403 | POST /organizations saat sudah punya 1 org |
| AC3 | Create project ke-4 di Free tier → HTTP 403 | POST /projects saat sudah ada 3 |
| AC4 | Create task ke-51 di project (Free tier) → HTTP 403 | POST /tasks saat sudah ada 50 |
| AC5 | Admin dapat upgrade user ke Pro via `PATCH /admin/users/:id/tier` | Panggil endpoint, cek response |
| AC6 | Setelah upgrade Pro, user bisa buat 2 org total | Create org kedua, harus berhasil |
| AC7 | ALL orgs user upgrade ke tier tsb | Check semua org punya fitur premium |
| AC8 | Member yang di-invite nikmati fitur premium org | User B (Free) join org Pro → dapat unlimited tasks |
| AC9 | Extend tier aktif menambah durasi dari expire date saat ini, bukan dari sekarang | Cek `tier_expires_at` setelah extend |
| AC10 | Tier expired → efektif jadi Free, quota Free berlaku | Set `tier_expires_at` ke masa lalu, coba create org kedua |
| AC11 | Semua response API menyertakan `tier_info` | Cek semua endpoint |
| AC12 | `license_warning` lama dihapus dan digantikan `tier_info` | Pastikan tidak ada duplikasi field |
| AC13 | Non-admin tidak bisa akses `PATCH /admin/users/:id/tier` → HTTP 403 | Test dengan user biasa |
| AC14 | `GET /tier/plans` bisa diakses tanpa auth, return 3 tier dengan limit & harga | Hit endpoint tanpa token |
| AC15 | `GET /users/me/tier` return tier + usage + limits user | Hit endpoint dengan token valid |
| AC16 | `effective_tier` fallback ke `free` jika tier expired | Set `tier_expires_at` ke masa lalu, hit endpoint |
| AC17 | Field `usage` di `/me/tier` menampilkan jumlah aktual | Bandingkan dengan data di DB |
| AC18 | `go build` passes tanpa error | Run `go build -o main .` |
| AC19 | Semua existing test pass | Run `go test ./...` |
| AC20 | Swagger docs diupdate | Run `swag init`, visit `/swagger/index.html` |

---

## 📁 Files to Create / Modify

### Buat Baru
- `migrations/XXXX_add_tier_to_users.up.sql`
- `migrations/XXXX_add_tier_to_users.down.sql`
- `migrations/XXXX_create_tier_tables.up.sql`
- `migrations/XXXX_create_tier_tables.down.sql`
- `modules/organizations/tier_plan_model.go` — TierPlan, TierLimit
- `modules/organizations/tier_plan_repository.go` — CRUD untuk tier_plans & tier_limits
- `utils/quota.go` — TierLimits, GetTierLimits, IsTierActive, GetEffectiveTier, ErrQuotaExceeded
- `middlewares/require_tier.go` — RequireTierFeature middleware

### Modifikasi
- `modules/auth/auth_models.go` — tambah Tier, TierExpiresAt, TierActivatedAt, TierActivatedBy ke User
- `modules/organizations/organization_repository.go` — tambah CountByOwner, CountMembers method
- `modules/organizations/organization_service.go` — tambah ActivateTier, GetMyTierInfo, CountUsage, quota check di CreateOrganization
- `modules/organizations/organization_handler.go` — tambah 3 endpoint + update model reference
- `modules/projects/service.go` — tambah quota check sebelum CreateProject
- `modules/tasks/service.go` — tambah quota check sebelum CreateTask
- `modules/clients/service.go` — tambah quota check sebelum CreateClient
- `modules/invoices/service.go` — tambah quota check sebelum CreateInvoice
- `utils/response.go` — ganti `license_warning` → `tier_info`
- `middlewares/require_auth.go` — hapus setLicenseWarning, ganti dengan setTierInfo
- `main.go` — register 3 route baru + hapus license routes

### Hapus Entirely
- `modules/licenses/` (semua file)

---

## 🔍 Verification Commands

```bash
# Build
go build -o main .

# Test
go test ./...

# Migration
migrate -path ./migrations -database "postgres://..." up

# Regenerate Swagger
go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -o docs --parseDependency
```

---

## 📝 Notes

- **Tidak ada Stripe/payment gateway** — semua aktivasi tier dilakukan manual oleh admin via endpoint setelah konfirmasi pembayaran di luar sistem.
- **License key lama dihapus** — karena belum release dan belum ada user, seluruh `modules/licenses` dihapus entirely.
- **Pricing dari database** — tier_plans dan tier_limits table memungkinkan perubahan harga tanpa redeploy.
- **Unlimited = -1** — sentinel value `-1` untuk limit tidak terbatas.
- **Invoice limit** dihitung per bulan kalender (bukan rolling 30 hari).
- **Personal org** tetap auto-created saat register, tidak dihitung sebagai "team org"
- **Team org** = org yang dibuat manual oleh user (OrgType = "team")