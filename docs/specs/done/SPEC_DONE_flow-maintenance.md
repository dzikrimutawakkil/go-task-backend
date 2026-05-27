# ACTIVE SPEC — Flow Maintenance & New Features

**Created:** 2026-05-26
**Status:** ✅ DISETUJUI — Ready to Execute
**Phase:** Phase 1e — Flow Maintenance & Project Status

---

## 📋 Scope: Backend Only (No Frontend)

Semua implementasi fokus di sisi backend API. Frontend handling dilakukan oleh tim terpisah.

---

## 🎯 Tasks Overview

| # | Task | Type | Effort | Priority |
|---|---|---|---|---|
| **Q17** | License Expiry → Soft Warning Banner | Quick Win | 2 jam | 🚀 High |
| **Q18** | Auto-generate default labels saat project dibuat | Quick Win | 2 jam | 🚀 High |
| **Q19** | Seed project status + set default | Quick Win | 1 jam | 🚀 High |
| **M11** | Workspace Switch Endpoint | Major | 2 jam | 🏗️ Medium |
| **M12** | Project Status Workflow | Major | 3 jam | 🏗️ Medium |

---

## 🔍 Task Details

### Q17: License Expiry — Soft Warning Banner

#### Requirements
- Saat user login atau request API manapun, check apakah license organization sudah expired
- Jika expired, tambahkan field `license_warning` di response JSON dan header `X-License-Warning: expired`
- Format response warning:
  ```json
  {
    "success": true,
    "message": "Request successful",
    "data": { ... },
    "license_warning": {
      "expired": true,
      "expired_at": "2026-05-19T00:00:00Z",
      "days_remaining": -7,
      "message": "License expired. Please upgrade to continue premium features."
    }
  }
  ```
- **Behavior:** User tetap bisa akses semua fitur, tapi frontend akan show banner berdasarkan field ini

#### Technical Approach
1. Modify `RequireAuth` middleware atau buat middleware baru `LicenseCheckMiddleware`
2. Setelah validasi user + org, check `organization.license_expires_at` vs current time
3. Inject warning ke Gin context → di-extract di response wrapper atau handler
4. Response util `SendSuccess` perlu di-modify untuk accept warning parameter

#### Database Changes
- None (license_expires_at sudah ada di organizations table)

#### Files to Modify
- `middlewares/auth.go` atau buat `middlewares/license.go` baru
- `utils/response.go` — add `SendSuccessWithWarning` variant
- Potentially handler response wrappers

#### Acceptance Criteria
- [ ] API response includes `license_warning` when license expired
- [ ] API response includes `license_warning` with `expired: false` when license active
- [ ] Header `X-License-Warning` present on all authenticated responses
- [ ] No performance degradation (check should be O(1) using cached org data)

---

### Q18: Auto-generate Default Labels Saat Project Dibuat

#### Requirements
- Saat project baru dibuat, auto-generate 5 labels:
  1. `Todo` (index: 0, color: `#E2E8F0` / gray)
  2. `On Going` (index: 1, color: `#3B82F6` / blue)
  3. `Done` (index: 2, color: `#22C55E` / green)
  4. `Delivered` (index: 3, color: `#A855F7` / purple)
  5. `Canceled` (index: 4, color: `#EF4444` / red)
- Label dibuat per-project (bukan global)
- Default status task yang dibuat = label pertama (`Todo`)

#### Technical Approach
1. Modify `projects.CreateProject` service method
2. Setelah project berhasil dibuat, buat 5 label records dengan project_id yang baru
3. Gunakan transaction untuk ensure atomicity
4. Update task creation logic untuk default ke label dengan index=0

#### Database Changes
- None (labels table sudah ada dengan project_id foreign key)

#### Files to Modify
- `modules/projects/service.go` — modify `CreateProject`
- `modules/tasks/service.go` — modify `CreateTask` untuk default label
- Potentially repository layer if needed

#### Acceptance Criteria
- [ ] POST /projects creates project + 5 labels atomically
- [ ] Labels have correct names, colors, and indices
- [ ] GET /projects/:id/projects includes labels
- [ ] Task created without explicit label_id → auto-assigned to label index=0

---

### Q19: Seed Project Status + Set Default

#### Requirements
- Seed 4 project statuses (if not exists):
  1. `Active` (name: "Active", color: `#22C55E` / green) — DEFAULT
  2. `On Hold` (name: "On Hold", color: `#F59E0B` / amber)
  3. `Completed` (name: "Completed", color: `#3B82F6` / blue)
  4. `Archived` (name: "Archived", color: `#6B7280` / gray)
- New projects auto-set to `Active` status (seeding default)
- Seed runs via existing seeder mechanism in `config/db.go`

#### Technical Approach
1. Add project status seeding in `config/db.go`
2. Add `ProjectStatus` model in `models/project_status.go` or extend existing project model
3. Add `status_id` foreign key to projects table via migration

#### Database Changes
```sql
-- Migration: add project_statuses table
CREATE TABLE IF NOT EXISTS project_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7) NOT NULL DEFAULT '#6B7280',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Migration: add status_id to projects table
ALTER TABLE projects ADD COLUMN IF NOT EXISTS status_id UUID REFERENCES project_statuses(id);

-- Update projects SET status_id = <active_status_id> WHERE status_id IS NULL;
```

#### Files to Modify
- `config/db.go` — add seeder for project_statuses
- `models/project.go` — add `StatusID` field and relation
- Migration file for project_statuses table + alter projects table

#### Acceptance Criteria
- [ ] Seed creates 4 project statuses if not exists
- [ ] Projects table has `status_id` column
- [ ] New projects auto-assigned to "Active" status
- [ ] GET /projects returns project with `status` object

---

### M11: Workspace Switch Endpoint

#### Requirements
- Endpoint: `POST /users/me/switch-organization`
- Request body:
  ```json
  {
    "organization_id": "uuid-of-target-org"
  }
  ```
- Validasi:
  - User harus punya membership di organization target
  - Jika valid, set organization tersebut sebagai "active" context
- Response: Return user object dengan organization yang baru

#### Technical Approach
1. Buat handler baru di `modules/users/handler.go`
2. Buat service method `SwitchOrganization(userID, orgID)`
3. Validasi membership via existing `GetMembership` or similar
4. Update user's current organization context (may need to track "last_active_org" in user record or just validate membership exists)

#### Database Changes
- Potentially add `last_active_org_id` to users table (optional, untuk analytics)

#### Files to Modify
- `modules/users/handler.go` — add `POST /users/me/switch-organization`
- `modules/users/service.go` — add `SwitchOrganization`
- `modules/users/repository.go` — add if needed
- `main.go` — register route

#### Acceptance Criteria
- [ ] POST /users/me/switch-organization with valid org_id returns success
- [ ] POST /users/me/switch-organization with invalid org_id returns 404
- [ ] POST /users/me/switch-organization with org user is not member of returns 403
- [ ] Response includes user's current organization context

---

### M12: Project Status Workflow

#### Requirements
- Project memiliki status independent dari task labels
- Endpoint untuk update project status:
  - `PATCH /projects/:id` sudah ada → perlu support `status_id` field
- Project status bisa di-read via:
  - `GET /projects` — includes status
  - `GET /projects/:id` — includes status detail
- Validation: Status yang di-assign harus ada di project_statuses table

#### Technical Approach
1. Extend existing project handlers/services to support `status_id`
2. Add validation: only allow valid project_status IDs
3. Add `status` relation to Project model

#### Database Changes
- Already covered in Q19 (status_id column + project_statuses table)

#### Files to Modify
- `modules/projects/handler.go` — update PATCH to accept status_id
- `modules/projects/service.go` — add status update logic
- `modules/projects/repository.go` — add if needed
- `models/project.go` — ensure Status relation is defined

#### Acceptance Criteria
- [ ] PATCH /projects/:id with status_id updates project status
- [ ] GET /projects returns projects with status object
- [ ] GET /projects/:id returns project with full status detail
- [ ] PATCH with invalid status_id returns 400

---

## 📊 Architecture Summary

```
Organization
├── license_expires_at (for Q17 warning check)
├── members (users with roles)
└── projects
    ├── status (Active/On Hold/Completed/Archived) ← M12
    └── tasks
        └── label (Todo/On Going/Done/Delivered/Canceled) ← Q18 auto-generate
            └── assignees (from org members)
```

---

## 🔗 Dependencies

| Task | Depends On | Blocking? |
|---|---|---|
| Q17 | License system (T10-T14) — sudah ada | No |
| Q18 | Labels table sudah ada | No |
| Q19 | Project status migration | No (standalone) |
| M11 | N1-N2 (personal org) + membership check | No |
| M12 | Q19 (project_statuses table) | Yes |

**Execution Order:** Q19 → M12 (karena M12 depend pada Q19), Q17/Q18/M11 bisa parallel.

---

## 📝 Notes

1. **Scope:** Backend only — frontend akan handle banner display sendiri berdasarkan response field
2. **MVP:** M13 (Role-based task permissions) di-skip — semua member bisa edit semua task
3. **Workspace Limit:** Unlimited — tidak ada batasan untuk saat ini
4. **Existing Features:** Tidak ada yang di-break — semua enhancement/addition