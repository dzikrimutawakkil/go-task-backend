# ACTIVE_SPEC.md — N6: Label → Urgency Level

## 📋 Overview

**Task:** N6 — Label → Urgency Level (Urgent, Normal, Low)
**Status:** ⏳ In Progress
**Effort:** 2 jam
**Goal:** Ubah label task dari workflow stages (Todo, On Going, Done...) menjadi urgency levels (Urgent, Normal, Low)

---

## 🔄 Current State vs Target State

| Aspect | Current | Target |
|---|---|---|
| **Label values** | Todo, On Going, Done, Delivered, Canceled | Urgent, Normal, Low |
| **Label function** | Kanban column / workflow stage | Urgency level / priority indicator |
| **Auto-generate count** | 5 labels | 3 labels |
| **User can change** | Via PATCH /tasks/:id | ✅ Same, no change needed |
| **Status (separate)** | Todo, On Progress, Done, Pending, Cancel | ✅ No change |

---

## 🎯 Technical Scope

### 1. Seeder Update
**File:** `config/db.go` (or wherever label seeder lives)

**Before:**
```go
labels := []string{"Todo", "On Going", "Done", "Delivered", "Canceled"}
```

**After:**
```go
labels := []string{"Urgent", "Normal", "Low"}
```

### 2. Service Layer Validation
**File:** `modules/tasks/service.go`

Ensure that when creating/updating a task, the `label_id` must reference a valid label belonging to the same organization/project. Invalid label values should be rejected with a clear error message.

### 3. API Response
Ensure that when fetching tasks, the `label` field returns the urgency level name (Urgent/Normal/Low) instead of the old workflow values.

### 4. Documentation Updates
- Update CLAUDE.md (already done during planning phase)
- Regenerate Swagger docs

---

## ✅ Acceptance Criteria

| # | Criteria | Test Method |
|---|---|---|
| AC1 | When creating a new project, exactly 3 labels are auto-generated: "Urgent", "Normal", "Low" | Create project via API, check labels |
| AC2 | Label values are validated — invalid values are rejected with HTTP 400 | Test with random label_id |
| AC3 | User can update task label via PATCH /tasks/:id | Test with valid label_id |
| AC4 | Task list response includes label with urgency value | GET /tasks or GET /tasks/:id |
| AC5 | Swagger docs reflect new label values | Visit /swagger/index.html |
| AC6 | `go build` passes without errors | Run `go build -o main .` |
| AC7 | All existing tests pass | Run `go test ./...` |

---

## 📁 Files to Modify

1. `config/db.go` — Update label seeder values
2. `modules/tasks/service.go` — Ensure label validation exists
3. `CLAUDE.md` — Already updated during planning
4. Run `swag init` to regenerate docs

---

## 🔍 Verification Commands

```bash
# Build
go build -o main .

# Test
go test ./...

# Regenerate Swagger
go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -o docs --parseDependency
```
