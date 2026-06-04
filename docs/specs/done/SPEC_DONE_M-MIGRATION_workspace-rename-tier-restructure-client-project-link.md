# ACTIVE_SPEC.md — Workspace Migration + Tier Restructure + Client-Project Link

## 📋 Overview

**Task:** M-MIGRATION — Breaking Changes: Workspace Rename + Tier Per-Workspace + Tier Baru + Client-Project Link
**Status:** ⏳ Pending
**Effort:** 1-2 hari
**Priority:** 🔴 Harus selesai sebelum ada paying customer

### Ringkasan perubahan
1. **Rename `organization` → `workspace`** di seluruh codebase (folder, variable, endpoint, DB column)
2. **Migrasi tier dari `users` → `organizations`** — tier sekarang per-workspace, bukan per-user
3. **Tier baru: Free / Starter / Business / Enterprise** — menggantikan Free/Pro/Ultimate
4. **Client ↔ Project link** — project bisa punya client, bisa inline create client saat buat project

---

## 🔄 Current State vs Target State

| Aspect | Current | Target |
|---|---|---|
| Naming | `organization` di kode, `workspace` di docs | `workspace` konsisten di semua tempat |
| Tier location | Kolom `tier` di tabel `users` | Kolom `tier` di tabel `organizations` (rename → `workspaces`) |
| Tier names | Free / Pro / Ultimate | Free / Starter / Business / Enterprise |
| Tier scope | Per user — semua workspace user inherit tier | Per workspace — tiap workspace punya tier sendiri |
| Client-Project | Tidak terhubung | Project punya optional `client_id`, bisa inline create |
| Admin endpoint | `PATCH /admin/users/:id/tier` | `PATCH /admin/workspaces/:id/tier` |

---

## 🎯 Tier Definition Final

| Fitur | Free | Starter | Business | Enterprise |
|---|---|---|---|---|
| **Workspace** | 1 | 1 | 3 | Unlimited |
| **Project per workspace** | 3 | Unlimited | Unlimited | Unlimited |
| **Task per project** | 50 | Unlimited | Unlimited | Unlimited |
| **Member per workspace** | 1 | 5 | 15 | Unlimited |
| **Clients** | 5 | Unlimited | Unlimited | Unlimited |
| **Invoices per bulan** | 10 | Unlimited | Unlimited | Unlimited |
| **Comments** | ❌ | ✅ | ✅ | ✅ |
| **Real-time (SSE)** | ❌ | ✅ | ✅ | ✅ |
| **Audit Log** | ❌ | ❌ | ✅ | ✅ |
| **Priority Support** | ❌ | ❌ | ✅ | ✅ |
| **Custom Branding** | ❌ | ❌ | ❌ | ✅ |
| **Harga/bulan** | Rp0 | Rp49.000 | Rp99.000 | Custom |

### Logika tier per workspace
- Tier menempel ke **workspace**, bukan ke user
- Owner workspace yang upgrade — semua member di workspace itu otomatis dapat fitur tier tersebut
- Satu user bisa punya workspace dengan tier berbeda (workspace A = Starter, workspace B = Free)
- Upgrade = workspace aktif saat ini yang di-upgrade (Opsi A — simpel, tidak perlu pilih)

---

## 🎯 Technical Scope

---

### BAGIAN 1: Rename Organization → Workspace

#### 1.1 Rename Folder & Package

```
modules/organizations/ → modules/workspaces/
```

Semua file di dalamnya:
- `organization_handler.go` → `workspace_handler.go`
- `organization_service.go` → `workspace_service.go`
- `organization_repository.go` → `workspace_repository.go`
- `organization_model.go` → `workspace_model.go`

Package name di semua file: `organizations` → `workspaces`

#### 1.2 Rename Database Table & Columns

**Migration baru:**

```sql
-- Rename table
ALTER TABLE organizations RENAME TO workspaces;

-- Rename foreign key columns di tabel lain
ALTER TABLE projects RENAME COLUMN organization_id TO workspace_id;
ALTER TABLE clients RENAME COLUMN organization_id TO workspace_id;
ALTER TABLE invoices RENAME COLUMN organization_id TO workspace_id;
ALTER TABLE tasks RENAME COLUMN organization_id TO workspace_id;

-- Rename di organization_members jika ada
ALTER TABLE organization_members RENAME TO workspace_members;
ALTER TABLE workspace_members RENAME COLUMN organization_id TO workspace_id;

-- Rename di invitations jika ada
ALTER TABLE invitations RENAME COLUMN organization_id TO workspace_id;
```

#### 1.3 Rename di Model

**File:** `modules/workspaces/workspace_model.go`

```go
type Workspace struct {
    // gorm:"table:workspaces"
    // Semua field yang sebelumnya di Organization
    // organization_id di relasi lain → workspace_id
}
```

#### 1.4 Rename Endpoint URLs

| Before | After |
|---|---|
| `GET /organizations` | `GET /workspaces` |
| `POST /organizations` | `POST /workspaces` |
| `POST /organizations/invite` | `POST /workspaces/invite` |
| `GET /organizations/members` | `GET /workspaces/members` |
| `PATCH /organizations/members/:id` | `PATCH /workspaces/members/:id` |
| `DELETE /organizations/members/:id` | `DELETE /workspaces/members/:id` |
| `GET /organizations/invitations` | `GET /workspaces/invitations` |
| `POST /api/users/me/switch-organization` | `POST /api/users/me/switch-workspace` |
| `PATCH /admin/users/:id/tier` | `PATCH /admin/workspaces/:id/tier` |

#### 1.5 Rename Header

| Before | After |
|---|---|
| `X-Organization-ID` | `X-Workspace-ID` |

Update di:
- `middlewares/RequireAuth` — semua reference ke `X-Organization-ID`
- `CLAUDE.md` — Architecture Notes
- Swagger annotations di semua handler

#### 1.6 Rename Variable & Field di Seluruh Codebase

Lakukan find & replace (case-sensitive) untuk:

| Before | After |
|---|---|
| `organizationID` / `orgID` | `workspaceID` |
| `organization_id` | `workspace_id` |
| `OrganizationID` | `WorkspaceID` |
| `orgRepo` | `workspaceRepo` |
| `orgService` | `workspaceService` |
| `orgHandler` | `workspaceHandler` |
| `FindByOrganizationID` | `FindByWorkspaceID` |
| `ByOrg` scope | `ByWorkspace` scope |

**File yang pasti perlu diubah:**
- `main.go` — semua DI wiring
- `middlewares/auth.go` — header extraction
- `modules/projects/` — semua file
- `modules/tasks/` — semua file
- `modules/clients/` — semua file
- `modules/invoices/` — semua file
- `models/` — shared scopes
- `utils/` — jika ada reference ke org

---

### BAGIAN 2: Migrasi Tier dari Users → Workspaces

#### 2.1 Database Migration

```sql
-- Tambah kolom tier ke workspaces (setelah rename organizations → workspaces)
ALTER TABLE workspaces
  ADD COLUMN tier VARCHAR(20) NOT NULL DEFAULT 'free',
  ADD COLUMN tier_expires_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN tier_activated_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN tier_activated_by BIGINT REFERENCES users(id);

-- Migrasi data: copy tier dari owner user ke workspace
UPDATE workspaces w
SET
  tier = u.tier,
  tier_expires_at = u.tier_expires_at,
  tier_activated_at = u.tier_activated_at,
  tier_activated_by = u.tier_activated_by
FROM workspace_members wm
JOIN users u ON u.id = wm.user_id
WHERE wm.workspace_id = w.id
  AND wm.role = 'owner';

-- Hapus kolom tier dari users (setelah data berhasil dimigrasikan)
ALTER TABLE users
  DROP COLUMN tier,
  DROP COLUMN tier_expires_at,
  DROP COLUMN tier_activated_at,
  DROP COLUMN tier_activated_by;
```

**File migration:**
```
migrations/XXXX_migrate_tier_to_workspaces.up.sql
migrations/XXXX_migrate_tier_to_workspaces.down.sql
```

#### 2.2 Update Model

**`modules/workspaces/workspace_model.go`** — tambah field tier:

```go
type Workspace struct {
    // ... existing fields ...
    Tier            string     `json:"tier" gorm:"default:free"`
    TierExpiresAt   *time.Time `json:"tier_expires_at"`
    TierActivatedAt *time.Time `json:"tier_activated_at"`
    TierActivatedBy *uint      `json:"tier_activated_by"`
}
```

**`modules/auth/model.go`** — hapus field tier:

```go
type User struct {
    // HAPUS: Tier, TierExpiresAt, TierActivatedAt, TierActivatedBy
}
```

#### 2.3 Update Quota Helper

**`utils/quota.go`** — update tier names dan limits:

```go
func GetTierLimits(tier string) TierLimits {
    switch tier {
    case "starter":
        return TierLimits{
            MaxWorkspaces: 1, MaxProjects: -1, MaxTasksPerProject: -1,
            MaxMembers: 5, MaxClients: -1, MaxInvoicesPerMonth: -1,
            CanComment: true, CanSSE: true, CanAuditLog: false,
            CanPrioritySupport: false, CanCustomBranding: false,
        }
    case "business":
        return TierLimits{
            MaxWorkspaces: 3, MaxProjects: -1, MaxTasksPerProject: -1,
            MaxMembers: 15, MaxClients: -1, MaxInvoicesPerMonth: -1,
            CanComment: true, CanSSE: true, CanAuditLog: true,
            CanPrioritySupport: true, CanCustomBranding: false,
        }
    case "enterprise":
        return TierLimits{
            MaxWorkspaces: -1, MaxProjects: -1, MaxTasksPerProject: -1,
            MaxMembers: -1, MaxClients: -1, MaxInvoicesPerMonth: -1,
            CanComment: true, CanSSE: true, CanAuditLog: true,
            CanPrioritySupport: true, CanCustomBranding: true,
        }
    default: // free
        return TierLimits{
            MaxWorkspaces: 1, MaxProjects: 3, MaxTasksPerProject: 50,
            MaxMembers: 1, MaxClients: 5, MaxInvoicesPerMonth: 10,
            CanComment: false, CanSSE: false, CanAuditLog: false,
            CanPrioritySupport: false, CanCustomBranding: false,
        }
    }
}
```

#### 2.4 Update Quota Check di Service Layer

Semua quota check yang sebelumnya baca `user.Tier` harus diganti baca dari `workspace.Tier`.

Pattern baru:

```go
// BEFORE (per user)
user, _ := s.authService.FindByID(userID)
effectiveTier := utils.GetEffectiveTier(user) // baca dari user

// AFTER (per workspace)
workspace, _ := s.workspaceRepo.FindByID(workspaceID)
effectiveTier := utils.GetEffectiveTier(workspace) // baca dari workspace
```

`GetEffectiveTier` di `utils/quota.go` perlu diupdate parameternya dari `*auth.User` → `*workspaces.Workspace`.

#### 2.5 Update Admin Endpoint

**Endpoint lama:** `PATCH /admin/users/:id/tier`
**Endpoint baru:** `PATCH /admin/workspaces/:id/tier`

Request body tidak berubah:
```json
{
  "tier": "starter",
  "duration_months": 12
}
```

Validasi tier — update valid values:
```go
validTiers := map[string]bool{
    "free": true,
    "starter": true,
    "business": true,
    "enterprise": true,
}
```

Logic extend tidak berubah — extend dari `tier_expires_at` jika masih aktif, mulai dari sekarang jika sudah expired.

#### 2.6 Update `GetMyTierInfo`

Endpoint: `GET /users/me/tier` (atau rename ke `GET /workspaces/me/tier`)

Sekarang baca tier dari workspace aktif user (dari context `X-Workspace-ID`), bukan dari user.

```go
func (s *workspaceService) GetMyTierInfo(workspaceID uint) (*TierInfoResult, error) {
    workspace, err := s.repo.FindByID(workspaceID)
    // ... baca tier dari workspace, bukan dari user
}
```

#### 2.7 Update TierPlans di DB

Tabel `tier_plans` — update nama tier dan harga:

```sql
UPDATE tier_plans SET
  name = 'starter',
  price_monthly = 49000,
  price_yearly = 490000
WHERE name = 'pro';

UPDATE tier_plans SET
  name = 'business',
  price_monthly = 99000,
  price_yearly = 990000
WHERE name = 'ultimate';

INSERT INTO tier_plans (name, price_monthly, price_yearly)
VALUES ('enterprise', 0, 0); -- 0 = custom pricing, diisi manual
```

---

### BAGIAN 3: Client ↔ Project Link

#### 3.1 Database Migration

```sql
-- Tambah foreign key client_id ke projects
ALTER TABLE projects
  ADD COLUMN client_id BIGINT REFERENCES clients(id) ON DELETE SET NULL;
```

**File migration:**
```
migrations/XXXX_add_client_to_projects.up.sql
migrations/XXXX_add_client_to_projects.down.sql
```

#### 3.2 Update Project Model

**`modules/projects/model.go`**

```go
type Project struct {
    // ... existing fields ...
    ClientID *uint   `json:"client_id" gorm:"index"`
    Client   *Client `json:"client,omitempty" gorm:"foreignKey:ClientID"`
}
```

#### 3.3 Update Create Project Request

**`modules/projects/service.go`**

```go
type CreateProjectRequest struct {
    // ... existing fields ...

    // Opsi 1: pilih client yang sudah ada
    ClientID *uint `json:"client_id"`

    // Opsi 2: inline create client baru
    NewClient *InlineCreateClientRequest `json:"new_client"`
}

type InlineCreateClientRequest struct {
    Name      string `json:"name" binding:"required"`
    Email     string `json:"email"`
    WhatsApp  string `json:"whatsapp"`
    Phone     string `json:"phone"`
    Company   string `json:"company"`
}
```

**Validasi:**
- `client_id` dan `new_client` tidak boleh diisi bersamaan → HTTP 400
- Keduanya boleh kosong — client di project bersifat opsional
- Jika `new_client` diisi, buat client baru dulu (dalam transaksi yang sama), lalu gunakan ID-nya sebagai `client_id`

#### 3.4 Logic di Service Layer

**`modules/projects/service.go`** — update `CreateProject`:

```go
func (s *projectService) CreateProject(workspaceID uint, userID uint, req CreateProjectRequest) (*Project, error) {
    // Validasi: tidak boleh isi keduanya
    if req.ClientID != nil && req.NewClient != nil {
        return nil, errors.New("cannot specify both client_id and new_client")
    }

    // Jika inline create client
    var clientID *uint
    if req.NewClient != nil {
        // Cek quota client dulu
        // Buat client baru via clientService atau langsung repo
        newClient, err := s.clientService.CreateClient(workspaceID, CreateClientRequest{
            Name:     req.NewClient.Name,
            Email:    req.NewClient.Email,
            WhatsApp: req.NewClient.WhatsApp,
            Phone:    req.NewClient.Phone,
            Company:  req.NewClient.Company,
        })
        if err != nil {
            return nil, fmt.Errorf("failed to create client: %w", err)
        }
        clientID = &newClient.ID
    } else {
        clientID = req.ClientID
    }

    // Validasi client_id milik workspace yang sama (jika diisi)
    if clientID != nil {
        client, err := s.clientRepo.FindByID(*clientID)
        if err != nil || client.WorkspaceID != workspaceID {
            return nil, errors.New("client not found or does not belong to this workspace")
        }
    }

    // ... proceed with create project, pass clientID ...
}
```

#### 3.5 Update Project Response

Saat `GET /projects` dan `GET /projects/:id`, sertakan client info jika ada:

```json
{
  "id": 1,
  "name": "Website Redesign",
  "client_id": 3,
  "client": {
    "id": 3,
    "name": "PT Maju Jaya",
    "company": "PT Maju Jaya"
  },
  "...": "..."
}
```

Jika tidak ada client: `"client_id": null, "client": null`

#### 3.6 Update Project (PATCH)

`PATCH /projects/:id` juga bisa update `client_id`:

```json
{
  "client_id": 5
}
```

Atau unlink client:
```json
{
  "client_id": null
}
```

---

## ✅ Acceptance Criteria

### Rename Organization → Workspace

| # | Criteria | Test Method |
|---|---|---|
| AC1 | Header `X-Organization-ID` tidak lagi diterima — diganti `X-Workspace-ID` | Hit endpoint dengan header lama → HTTP 400/403 |
| AC2 | Semua endpoint `/organizations/...` sudah menjadi `/workspaces/...` | Cek routing di main.go |
| AC3 | `POST /api/users/me/switch-workspace` berfungsi | Hit endpoint, cek response |
| AC4 | Tabel DB sudah bernama `workspaces`, kolom `organization_id` sudah `workspace_id` | Cek DB schema |
| AC5 | `go build` passes tanpa error setelah rename | Run `go build -o main .` |

### Migrasi Tier ke Workspace

| # | Criteria | Test Method |
|---|---|---|
| AC6 | Kolom `tier` tidak lagi ada di tabel `users` | Cek DB schema |
| AC7 | Kolom `tier` ada di tabel `workspaces` dengan default `free` | Cek DB schema |
| AC8 | Workspace baru auto-dapat tier `free` saat dibuat | Create workspace, cek DB |
| AC9 | Quota check baca dari workspace, bukan dari user | Test buat project ke-4 di Free workspace → HTTP 403 |
| AC10 | `PATCH /admin/workspaces/:id/tier` berfungsi dengan tier baru | Hit endpoint dengan tier `starter` |
| AC11 | Tier `pro` dan `ultimate` tidak lagi valid | Hit endpoint dengan tier `pro` → HTTP 400 |
| AC12 | Tier `starter`, `business`, `enterprise` valid | Hit endpoint dengan masing-masing → berhasil |
| AC13 | `GET /users/me/tier` return tier dari workspace aktif, bukan dari user | Cek response |
| AC14 | `GET /tier/plans` return 4 tier dengan nama dan harga baru | Hit endpoint |
| AC15 | Semua member workspace dapat fitur sesuai tier workspace (bukan tier user sendiri) | Login sebagai member, cek quota |

### Client ↔ Project

| # | Criteria | Test Method |
|---|---|---|
| AC16 | `POST /projects` bisa terima `client_id` (existing client) | Create project dengan client_id valid |
| AC17 | `POST /projects` bisa terima `new_client` (inline create) | Create project dengan new_client object |
| AC18 | Isi keduanya (`client_id` + `new_client`) → HTTP 400 | Test kombinasi keduanya |
| AC19 | Kosongkan keduanya → project berhasil dibuat tanpa client | Create project tanpa client field |
| AC20 | Inline create client menghormati quota client (Free max 5) | Test di workspace Free yang sudah punya 5 client |
| AC21 | `client_id` dari workspace lain → HTTP 400 | Test cross-workspace client |
| AC22 | `GET /projects/:id` menyertakan `client` object jika ada | Cek response |
| AC23 | `PATCH /projects/:id` bisa update atau unlink client | Test set `client_id: null` |

### General

| # | Criteria | Test Method |
|---|---|---|
| AC24 | `go build` passes tanpa error | Run `go build -o main .` |
| AC25 | `go test ./...` passes | Run `go test ./...` |
| AC26 | Swagger docs diupdate — tidak ada lagi mention `organization` | Run `swag init`, cek `/swagger/index.html` |
| AC27 | `CLAUDE.md` diupdate — schema, endpoints, architecture notes | Review manual |

---

## 📁 Files to Create / Modify

### Migration Files (Buat Baru)
- `migrations/XXXX_rename_organizations_to_workspaces.up.sql`
- `migrations/XXXX_rename_organizations_to_workspaces.down.sql`
- `migrations/XXXX_migrate_tier_to_workspaces.up.sql`
- `migrations/XXXX_migrate_tier_to_workspaces.down.sql`
- `migrations/XXXX_add_client_to_projects.up.sql`
- `migrations/XXXX_add_client_to_projects.down.sql`

### Rename Folder
- `modules/organizations/` → `modules/workspaces/` (rename semua file di dalamnya)

### Modifikasi
- `modules/workspaces/workspace_model.go` — tambah tier fields, hapus org naming
- `modules/workspaces/workspace_repository.go` — rename semua method, update queries
- `modules/workspaces/workspace_service.go` — rename, update tier logic (baca dari workspace)
- `modules/workspaces/workspace_handler.go` — rename, update endpoints
- `modules/auth/model.go` — hapus tier fields dari User struct
- `modules/auth/service.go` — hapus tier-related logic
- `modules/projects/model.go` — tambah ClientID, Client relation
- `modules/projects/service.go` — tambah inline client create logic
- `modules/projects/repository.go` — update queries untuk include client
- `modules/projects/handler.go` — update request/response
- `modules/clients/service.go` — tidak banyak berubah, pastikan workspace naming konsisten
- `modules/invoices/service.go` — pastikan workspace naming konsisten
- `modules/tasks/service.go` — pastikan workspace naming konsisten
- `middlewares/auth.go` — ganti `X-Organization-ID` → `X-Workspace-ID`
- `models/scopes.go` — rename `ByOrg` → `ByWorkspace`
- `utils/quota.go` — update tier names, update GetEffectiveTier parameter
- `main.go` — update semua DI wiring, update semua route paths
- `CLAUDE.md` — update schema overview, architecture notes, endpoints
- `docs/FLOW-USER.md` — update semua mention "organization" → "workspace"
- `docs/README.md` — update endpoints table, tier table

---

## ⚠️ Urutan Eksekusi yang Aman

Kerjakan dalam urutan ini untuk menghindari build error:

```
1. Buat semua migration files (SQL only, belum dijalankan)
2. Rename folder modules/organizations → modules/workspaces
3. Update semua model files (struct fields)
4. Update utils/quota.go (tier names + parameter types)
5. Update repository files
6. Update service files
7. Update handler files
8. Update middlewares/auth.go (header rename)
9. Update models/scopes.go
10. Update main.go (DI + routing)
11. go build → fix semua compile error
12. Jalankan migrations
13. go test ./...
14. swag init
15. Update dokumentasi (CLAUDE.md, README, FLOW-USER)
```

---

## 🔍 Verification Commands

```bash
# Pastikan tidak ada sisa kata "organization" di kode (kecuali di docs lama)
grep -r "organization" --include="*.go" .

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

- **Breaking change pada header** — frontend harus update dari `X-Organization-ID` ke `X-Workspace-ID`. Koordinasi dengan frontend sebelum deploy.
- **Breaking change pada endpoints** — semua `/organizations/...` berubah. Jika ada client API yang sudah integrate, perlu dikabari.
- **Data migration aman** — migration tier dari users ke workspaces menggunakan UPDATE berdasarkan owner role, bukan hapus data. Down migration juga harus disiapkan untuk rollback.
- **Enterprise tier harga 0** — di DB, enterprise disimpan sebagai harga 0 yang artinya custom. Frontend harus handle tampilkan "Hubungi Kami" bukan "Rp0".
- **Inline create client dalam transaksi** — idealnya create client dan create project dalam satu DB transaction supaya tidak ada orphan client jika project gagal dibuat.
