# Active Spec: Bug Fixes + Workspace Naming Migration

## 📋 Overview
- **Task:** Fix QA bugs + full codebase rename from "org/organization" to "workspace"
- **Priority:** High
- **Status:** In Progress
- **Date:** 2026-06-04

---

## 🐛 Bugs to Fix

### Bug #1: GET /invoices returns 500
- **File:** `modules/invoices/invoice_handler.go`
- **Symptom:** `ListInvoices` returns HTTP 500 with "Failed to fetch invoices"
- **Root Cause:** Investigate `FindAllByWorkspace` in invoice_repository.go
- **Expected:** Should return list of invoices for the workspace

### Bug #2: Handlers read `org_id` instead of `workspace_id`
- **Issue:** Middleware sets `workspace_id` but handlers read `org_id`
- **Affected Endpoints:**
  - `GET /projects/:id/tasks` → task_handler.go
  - `GET /tasks/search` → task_handler.go
  - `GET /projects/:id/labels` → label_handler.go
  - `PATCH /labels/:id` → label_handler.go
- **Fix:** Change all `c.Get("org_id")` → `c.Get("workspace_id")`

---

## 🔄 Full Codebase Sweep

### Naming Convention Changes
| Old | New |
|-----|-----|
| `org_id` | `workspace_id` |
| `organization` | `workspace` |
| `Organization` | `Workspace` |
| `org` | `workspace` |
| `Org` | `Workspace` |

### Files to Check and Update
1. `modules/tasks/task_handler.go` — `org_id` → `workspace_id`
2. `modules/tasks/label_handler.go` — `org_id` → `workspace_id`
3. `modules/tasks/task_service.go` — any org references
4. `modules/tasks/task_repository.go` — any org references
5. `modules/tasks/task_models.go` — model field names
6. `modules/projects/project_handler.go` — org references
7. `modules/projects/project_service.go` — org references
8. `modules/projects/project_repository.go` — org references
9. `modules/workspaces/workspace_service.go` — check for org remnants
10. `modules/invoices/invoice_handler.go` — check org references
11. `modules/invoices/invoice_service.go` — check org references
12. `modules/invoices/invoice_repository.go` — check org references
13. `middlewares/require_auth.go` — ensure consistent
14. `main.go` — route comments, variable names
15. `docs/docs.go` — swagger annotations
16. `models/*.go` — scope names, model fields
17. Any test files

---

## ✅ Acceptance Criteria

1. **Bug Fix #1:** `GET /invoices` returns HTTP 200 with invoice list
2. **Bug Fix #2:** All endpoints work without needing explicit `X-Workspace-ID` header
3. **Naming Migration:** Zero occurrences of `org_id`, `organization`, `Org` in Go source files (except comments explaining the migration)
4. **Build:** `go build -o main .` succeeds with no errors
5. **Tests:** `go test ./...` passes all tests
6. **Verify:** All QA test cases pass with curl

---

## 🧪 Verification Steps

```bash
# 1. Test invoice list
curl -s http://localhost:8080/invoices -H "Authorization: Bearer TOKEN"

# 2. Test task search (should work without X-Workspace-ID)
curl -s "http://localhost:8080/tasks/search?q=Test" -H "Authorization: Bearer TOKEN"

# 3. Test project tasks
curl -s http://localhost:8080/projects/12/tasks -H "Authorization: Bearer TOKEN"

# 4. Test labels
curl -s http://localhost:8080/projects/12/labels -H "Authorization: Bearer TOKEN"

# 5. Verify no org references remain
grep -r "org_id" --include="*.go" modules/ middlewares/ || echo "Clean!"
grep -r "Organization" --include="*.go" modules/ | grep -v "_test.go" || echo "Clean!"
```