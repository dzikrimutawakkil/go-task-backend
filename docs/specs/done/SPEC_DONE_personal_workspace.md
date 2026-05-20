# ACTIVE_SPEC: Personal Workspace Mode

**Status:** IN REVIEW
**Author:** PM
**Date:** 2026-05-20
**CEO Approval:** Pending

---

## 1. Goal

User can manage **projects, tasks, statuses, priorities, labels** WITHOUT needing to create or join an organization first. Every user automatically gets a personal workspace upon registration. When they want to collaborate, they can invite members — just like Trello/Notion personal mode.

---

## 2. Architecture Decision

**Approach:** Auto-seed personal org on registration, NOT nullable organization_id.

> Every user gets 1 automatic "Personal Workspace" organization upon signup. They can use the app immediately. When they're ready to collaborate, they invite members to that org. The existing invite flow (Q4) already handles this. Per-project RBAC is Phase 2.

**Why not nullable?**
- No schema breaking change
- API remains backward-compatible (no `org_id` required, but explicit `X-Organization-ID` still works)
- Existing RBAC/invite infrastructure already works on top of orgs

---

## 3. Database Migration

**File:** `migrations/000007_add_org_type.up.sql`

```sql
-- Add organization type for personal vs team workspaces
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS org_type VARCHAR(20) NOT NULL DEFAULT 'personal';
```

- `personal` — auto-created workspace (user cannot delete this org)
- `team` — user-created shared workspace

---

## 4. Model Changes

### 4.1 `modules/organizations/organization_models.go`

Add `OrgType` field to `Organization` struct:

```go
type Organization struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Name      string    `gorm:"unique" json:"name"`
    OwnerID   uint      `json:"owner_id"`
    OrgType   string    `gorm:"type:varchar(20);default:'personal'" json:"org_type"` // personal, team
    CreatedAt time.Time `json:"created_at"`
}
```

### 4.2 `modules/organizations/organization_models.go`

Add constants:

```go
const (
    OrgTypePersonal = "personal"
    OrgTypeTeam     = "team"
)
```

---

## 5. Service Changes

### 5.1 `modules/auth/auth_service.go`

Modify `Signup` to auto-create personal org:

```go
func (s *authService) Signup(input SignupInput) (*User, error) {
    // 1. Hash Password
    hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
    if err != nil {
        return nil, errors.New("failed to hash password")
    }

    // 2. Create User
    user := User{...}
    if err := s.repo.CreateUser(&user); err != nil {
        return nil, errors.New("email already registered")
    }

    // 3. Auto-create personal workspace
    personalOrg := organizations.Organization{
        Name:    input.Name + "'s Workspace",
        OwnerID: user.ID,
        OrgType: organizations.OrgTypePersonal,
    }
    if err := s.orgRepo.Create(&personalOrg); err != nil {
        return nil, errors.New("failed to create personal workspace")
    }
    // Add owner as member
    if err := s.orgRepo.AddMember(personalOrg.ID, user.ID, models.RoleOwner); err != nil {
        return nil, errors.New("failed to setup personal workspace")
    }

    return &user, nil
}
```

**Dependency change:** `authService` needs `OrganizationRepository` injected. Update `NewAuthService` signature and `main.go` wiring.

### 5.2 `modules/auth/auth_repository.go`

Add `CreateOrganization` method if not already present (or reuse existing org repo method).

---

## 6. Middleware Changes

### 6.1 `middlewares/auth_middleware.go`

Modify `RequireAuth` to **always set** `org_id` — using the user's personal workspace as fallback when no `X-Organization-ID` header is provided:

```go
func RequireAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // ... existing JWT parsing ...

        orgIDHeader := c.GetHeader("X-Organization-ID")

        if orgIDHeader != "" {
            // Existing: validate membership, set org_id
            ...
        } else {
            // NEW: Auto-resolve to user's personal workspace
            orgID, err := resolvePersonalOrgID(user.ID)
            if err != nil {
                c.AbortWithStatusJSON(500, gin.H{"success": false, "message": "Personal workspace not found"})
                return
            }
            c.Set("org_id", strconv.FormatUint(uint64(orgID), 10))
        }
    }
}
```

Need to add `resolvePersonalOrgID` helper that queries:
```sql
SELECT id FROM organizations WHERE owner_id = ? AND org_type = 'personal' LIMIT 1
```

> **Note:** This function is called on EVERY authenticated request. Cache the result per user (in-memory map with mutex) to avoid N+1 queries.

---

## 7. Handler Changes (Fallback Logic)

With middleware always setting `org_id`, most handlers will work without changes. However, `MustGet` panics if context value is missing — since middleware now always sets it, this is safer. But we should still make handlers defensive.

### 7.1 `modules/projects/project_handler.go`

`FindProjects` and `CreateProject` already use `c.MustGet("org_id")`. After middleware change, these will work automatically without header.

### 7.2 `modules/tasks/task_handler.go`

`FindTasksByProject` and `SearchTasks` use manual `c.Get("org_id")` check (returns 400 if missing). After middleware change, these will always pass — no more 400 errors when header is absent.

**No code changes needed** in handlers — the middleware fix propagates automatically.

---

## 8. Repository Changes

### 8.1 `modules/organizations/organization_repository.go`

Add method to resolve personal org by owner:

```go
func (r *organizationRepository) FindPersonalOrgByOwnerID(ownerID uint) (*Organization, error) {
    var org Organization
    err := r.db.Where("owner_id = ? AND org_type = ?", ownerID, OrgTypePersonal).First(&org).Error
    if err != nil {
        return nil, err
    }
    return &org, nil
}
```

Also add `Create` method if not already there.

---

## 9. API Behavior Summary

| Scenario | Before | After |
|---|---|---|
| User registers | No org | Auto-creates personal workspace |
| User calls `GET /projects` (no header) | ❌ 500 panic | ✅ Returns personal projects |
| User calls `POST /projects` (no header) | ❌ 500 panic | ✅ Creates in personal org |
| User calls `POST /tasks` (no header) | ✅ No check (bug) | ✅ Creates in project (via org) |
| User calls with `X-Organization-ID` | ✅ Works | ✅ Works (explicit org) |
| User invites teammate | ❌ No org | ✅ Can invite to personal org |

---

## 10. Out of Scope (Phase 2+)

- Per-project RBAC (default: all invited members can CRUD all tasks)
- Org deletion (personal orgs cannot be deleted)
- Org type display in API responses (Phase 2)
- Migration to move existing users' projects to auto-personal org (deferred)

---

## 11. File Changes Checklist

| # | File | Change |
|---|---|---|
| 1 | `migrations/000007_add_org_type.up.sql` | New file — add `org_type` column |
| 2 | `modules/organizations/organization_models.go` | Add `OrgType` field + constants |
| 3 | `modules/organizations/organization_repository.go` | Add `Create`, `FindPersonalOrgByOwnerID` |
| 4 | `modules/auth/auth_service.go` | Auto-create personal org in `Signup` |
| 5 | `modules/auth/auth_models.go` | Add `OrganizationRepository` dependency |
| 6 | `middlewares/auth_middleware.go` | Always set `org_id`, resolve personal org fallback + cache |
| 7 | `main.go` | Wire `OrganizationRepository` into `AuthService` |
| 8 | `CLAUDE.md` | Update backlog priorities |

---

## 12. Testing Checklist

- [ ] `POST /signup` creates user + personal org + org_users entry
- [ ] `GET /projects` (no header) returns personal projects
- [ ] `POST /projects` (no header) creates in personal org
- [ ] `GET /projects/:id/tasks` (no header) works
- [ ] `GET /tasks/search` (no header) works
- [ ] `POST /tasks` (no header) creates task in personal org project
- [ ] With `X-Organization-ID` header → uses explicit org (backward compatible)
- [ ] `POST /organizations/invite` with personal org works
- [ ] Build passes: `go build -o main .`

---

## 13. Effort Estimate

**Total: ~6-8 jam**

| Task | Effort |
|---|---|
| Migration + model | 1 jam |
| Repository method | 1 jam |
| Auth service (auto-create org) | 2 jam |
| Middleware fallback + cache | 2 jam |
| Main.go wiring + build | 1 jam |
| Testing + swagger | 1 jam |