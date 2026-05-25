# ACTIVE SPEC — Backend Alignment for Frontend Migration

## Context

Frontend project `freelance-os` sudah bisa berjalan baik. Ia mengonsumsi `go-task-backend` REST API. Tapi ada **gap** antara apa yang frontend harapkan vs apa yang backend currently serve. Spec ini mendokumentasikan semua perubahan yang diperlukan **di backend saja** agar frontend bisa berfungsi penuh tanpa perubahan.

> ⚠️ **CATATAN PENTING:** Semua perubahan yang tertera di spec ini adalah **100% backend-side only**. Frontend TIDAK boleh diubah.

---

## Gap Analysis Summary

### Yang SUDAH match (tidak perlu diubah)

| Endpoint | Status |
|---|---|
| `POST /api/auth/signup` | ✅ Endpoint ada, request/response format match |
| `POST /api/auth/login` | ✅ Endpoint ada, request format match |
| `GET /api/projects` | ✅ Berfungsi |
| `POST /api/projects` | ✅ Berfungsi |
| `DELETE /api/projects/:id` | ✅ Berfungsi |
| `GET /api/projects/:id/tasks` | ✅ Berfungsi |
| `POST /api/tasks` | ✅ Berfungsi |
| `PATCH /api/tasks/:id` | ✅ Berfungsi |
| `DELETE /api/tasks/:id` | ✅ Berfungsi |

### Yang BELUM match / TIDAK ADA (perlu diperbaiki)

| Endpoint | Issue | Priority |
|---|---|---|
| `POST /api/auth/signup` | Response missing `token` | 🚨 CRITICAL |
| `POST /api/auth/login` | Response missing `user` | 🚨 CRITICAL |
| `GET /api/auth/me` | Endpoint tidak ada | 🚨 CRITICAL |
| `POST /api/auth/forgot-password` | Endpoint tidak ada | 🚨 CRITICAL |
| `GET /api/projects/:id` | Endpoint tidak ada | 🚨 CRITICAL |
| `PATCH /api/projects/:id` | Handler method tidak ada | 🚨 CRITICAL |
| `modules/clients/` | Module sama sekali tidak ada | 🚨 CRITICAL |
| `modules/invoices/` | Module sama sekali tidak ada | 🚨 CRITICAL |

---

## Spec: Auth Responses Fix

### A1. Signup — Return `{ user, token }`

**File:** `modules/auth/auth_handler.go`, `modules/auth/auth_service.go`

**Before:**
```go
// auth_handler.go — Signup returns only user, no token
utils.SendSuccess(c, "Signup successful", gin.H{
    "user": user,
})
```

**After:**
```go
// auth_handler.go — Generate and return token alongside user
import "gotask-backend/modules/auth"

// In Signup handler — get token after creating user
token, err := h.authService.Login(auth.LoginInput{
    Email:    req.Email,
    Password: req.Password,
})
if err != nil {
    utils.SendError(c, http.StatusInternalServerError, "Signup succeeded but token generation failed")
    return
}

utils.SendSuccess(c, "Signup successful", gin.H{
    "user":  user,
    "token": token,
})
```

**Rationale:** Frontend `authService.ts` reads `response.data.data.token` and `response.data.data.user` from signup. Token diperlukan agar user langsung logged-in setelah register.

---

### A2. Login — Return `{ user, token }`

**File:** `modules/auth/auth_handler.go`

**Before:**
```go
// auth_handler.go — Login returns only token
utils.SendSuccess(c, "Login successful", gin.H{"token": token})
```

**After:**
```go
// auth_handler.go — Fetch user and return both
user, err := h.authService.GetUserByEmail(req.Email)
if err != nil {
    utils.SendError(c, http.StatusInternalServerError, "Login succeeded but user fetch failed")
    return
}

utils.SendSuccess(c, "Login successful", gin.H{
    "user":  user,
    "token": token,
})
```

**Rationale:** Frontend `authService.ts` reads `response.data.data.user` from login response (line 64). Currently user is missing.

**Also needed in auth_service.go:**
```go
// Add GetUserByEmail method already exists, verify it's exposed
func (s *authService) GetUserByEmail(email string) (*User, error) {
    return s.repo.FindUserByEmail(email)
}
```

---

### A3. Add `GET /api/auth/me`

**File:** `modules/auth/auth_handler.go`, `main.go`

**Implementation:**
```go
// auth_handler.go
func (h *Handler) Me(c *gin.Context) {
    user := c.MustGet("user").(models.MinimalUser)
    utils.SendSuccess(c, "success", gin.H{
        "user": user,
    })
}
```

**Route in main.go (protected group):**
```go
protected.GET("/auth/me", authHandler.Me)
```

**Rationale:** Frontend `authService.ts getMe()` calls `GET /api/auth/me` to restore session on page reload. Currently returns 404.

---

### A4. Add `POST /api/auth/forgot-password`

**File:** `modules/auth/auth_handler.go`, `modules/auth/auth_service.go`, `main.go`

**Handler:**
```go
// auth_handler.go
type ForgotPasswordRequest struct {
    Email string `json:"email" binding:"required"`
}

func (h *Handler) ForgotPassword(c *gin.Context) {
    var req ForgotPasswordRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.SendError(c, http.StatusBadRequest, err.Error())
        return
    }
    if err := h.authService.ForgotPassword(req.Email); err != nil {
        // Always return 200 to prevent email enumeration
        utils.SendSuccess(c, "If the email exists, a reset link has been sent")
        return
    }
    utils.SendSuccess(c, "If the email exists, a reset link has been sent")
}
```

**Service:**
```go
// auth_service.go
func (s *authService) ForgotPassword(email string) error {
    user, err := s.repo.FindUserByEmail(email)
    if err != nil {
        return nil // Silently ignore — prevent email enumeration
    }
    // TODO: Send reset email via utils.SendEmail (if email service configured)
    // For MVP: just log the request (email service may be added later)
    return nil
}
```

**Route in main.go (public):**
```go
r.POST("/forgot-password", authHandler.ForgotPassword)
```

**Rationale:** Frontend calls `POST /api/auth/forgot-password`. Currently returns 404.

---

## Spec: Project Endpoints Fix

### A5. Add `GET /api/projects/:id`

**File:** `modules/projects/project_handler.go`, `modules/projects/project_service.go`, `modules/projects/project_repository.go`

**Repository:**
```go
// project_repository.go
func (r *projectRepository) FindByID(id string) (*Project, error) {
    var project Project
    if err := r.db.First(&project, id).Error; err != nil {
        return nil, err
    }
    return &project, nil
}
```

**Service:**
```go
// project_service.go
func (s *projectService) GetProject(id string) (*Project, error) {
    return s.repo.FindByID(id)
}
```

**Handler:**
```go
// project_handler.go
func (h *ProjectHandler) GetProject(c *gin.Context) {
    id := c.Param("id")
    project, err := h.service.GetProject(id)
    if err != nil {
        utils.SendError(c, http.StatusNotFound, "Project not found")
        return
    }
    utils.SendSuccess(c, "Success", project)
}
```

**Route in main.go:**
```go
protected.GET("/projects/:id", projectHandler.GetProject)
```

**Rationale:** Frontend `projectService.ts getProject(id)` calls `GET /api/projects/{id}`. Currently this route doesn't exist.

---

### A6. Add `PATCH /api/projects/:id` (Update Project)

**File:** `modules/projects/project_handler.go`, `modules/projects/project_service.go`

**Handler:**
```go
// project_handler.go

// UpdateProjectRequest maps frontend field names
type UpdateProjectRequest struct {
    Name        *string `json:"name"`
    Description *string `json:"description"`
    Status      *string `json:"status"`
    Priority    *string `json:"priority"`
    Budget      *float64 `json:"budget"`
    Deadline    *string `json:"deadline"`
    Progress    *int    `json:"progress"`
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
    id := c.Param("id")

    var req UpdateProjectRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.SendError(c, http.StatusBadRequest, err.Error())
        return
    }

    input := UpdateProjectInput{}
    if req.Name != nil {
        input.Name = req.Name
    }
    if req.Description != nil {
        input.Description = req.Description
    }
    if req.Status != nil {
        input.Status = req.Status
    }
    if req.Priority != nil {
        input.Priority = req.Priority
    }
    if req.Budget != nil {
        input.Budget = req.Budget
    }
    if req.Deadline != nil {
        input.Deadline = req.Deadline
    }
    if req.Progress != nil {
        input.Progress = req.Progress
    }

    project, err := h.service.UpdateProject(id, input)
    if err != nil {
        if err.Error() == "project not found" {
            utils.SendError(c, http.StatusNotFound, "Project not found")
            return
        }
        utils.SendError(c, http.StatusInternalServerError, "Failed to update project")
        return
    }

    utils.SendSuccess(c, "Project updated successfully", project)
}
```

**Service:**
```go
// project_service.go
type UpdateProjectInput struct {
    Name        *string
    Description *string
    Status      *string
    Priority    *string
    Budget      *float64
    Deadline    *string
    Progress    *int
}

func (s *projectService) UpdateProject(id string, input UpdateProjectInput) (*Project, error) {
    project, err := s.repo.FindByID(id)
    if err != nil {
        return nil, errors.New("project not found")
    }

    if input.Name != nil {
        project.Name = *input.Name
    }
    if input.Description != nil {
        project.Description = *input.Description
    }
    if input.Status != nil {
        project.Status = *input.Status
    }
    if input.Priority != nil {
        project.Priority = *input.Priority
    }
    if input.Budget != nil {
        project.Budget = input.Budget
    }
    if input.Deadline != nil {
        project.Deadline = input.Deadline
    }
    if input.Progress != nil {
        project.Progress = int64(*input.Progress)
    }

    if err := s.repo.Update(project); err != nil {
        return nil, err
    }
    return project, nil
}
```

**Route in main.go:**
```go
protected.PATCH("/projects/:id", projectHandler.UpdateProject)
```

**Rationale:** Frontend `projectService.ts updateProject(id, data)` calls `PATCH /api/projects/{id}`. The route is registered but the handler method is missing.

---

## Spec: New Modules

### B1. Clients Module

**File:** New files in `modules/clients/`

**Scope:** CRUD for clients (Freelance OS "Clients" feature — contact management)

**Design decisions:**

#### Database Schema

```sql
-- Create clients table
CREATE TABLE IF NOT EXISTS clients (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    whatsapp VARCHAR(50),
    phone VARCHAR(50),
    company VARCHAR(255),
    website VARCHAR(255),
    address TEXT,
    notes TEXT,
    total_revenue DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_clients_org ON clients(organization_id);
```

**Note:** Table created via migration file in `config/` (golang-migrate).

#### Module Structure

```
modules/clients/
├── client_models.go      # Client model + DTO structs
├── client_repository.go  # GORM operations
├── client_service.go     # Business logic
└── client_handler.go     # HTTP handlers
```

#### API Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/clients` | List all clients for current org |
| `POST` | `/clients` | Create client |
| `GET` | `/clients/:id` | Get single client |
| `PATCH` | `/clients/:id` | Update client |
| `DELETE` | `/clients/:id` | Delete client |
| `GET` | `/clients/stats` | Get client stats `{ total, totalRevenue, avgRevenue }` |

#### Data Model

```go
type Client struct {
    ID             uint      `json:"id"`
    OrganizationID uint      `json:"organization_id"`
    Name           string    `json:"name"`
    Email          *string   `json:"email"`
    WhatsApp       *string   `json:"whatsapp"`
    Phone          *string   `json:"phone"`
    Company        *string   `json:"company"`
    Website        *string   `json:"website"`
    Address        *string   `json:"address"`
    Notes          *string   `json:"notes"`
    TotalRevenue   float64   `json:"total_revenue"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}
```

#### API Response Format

All responses use `utils/response.go` — `{ success: bool, message: string, data: any }`

List response:
```json
{
  "success": true,
  "message": "success",
  "data": {
    "clients": [...],
    "meta": { "total": 10 }
  }
}
```

Stats response:
```json
{
  "success": true,
  "message": "success",
  "data": {
    "total": 10,
    "totalRevenue": 50000000,
    "avgRevenue": 5000000
  }
}
```

---

### B2. Invoices Module

**File:** New files in `modules/invoices/`

**Scope:** CRUD for invoices + invoice number generation + revenue sync on mark-as-paid

**Design decisions:**

#### Database Schema

```sql
CREATE TABLE IF NOT EXISTS invoices (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id),
    invoice_number VARCHAR(50) NOT NULL UNIQUE,
    client_id INTEGER REFERENCES clients(id),
    project_id INTEGER REFERENCES projects(id),
    title VARCHAR(255),
    amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    tax DECIMAL(15,2) NOT NULL DEFAULT 0,
    discount DECIMAL(15,2) NOT NULL DEFAULT 0,
    amount_paid DECIMAL(15,2) DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    due_date TIMESTAMP,
    paid_at TIMESTAMP,
    notes TEXT,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_invoices_org ON invoices(organization_id);
CREATE INDEX idx_invoices_client ON invoices(client_id);
```

#### Module Structure

```
modules/invoices/
├── invoice_models.go      # Invoice model + InvoiceItem + DTOs
├── invoice_repository.go   # GORM operations
├── invoice_service.go     # Business logic + invoice number generation + revenue sync
└── invoice_handler.go     # HTTP handlers
```

#### Invoice Number Generation

Format: `INV-{YEAR}-{RAND}` where RAND = 3 uppercase alphanumeric chars.

Example: `INV-2026-X7K`, `INV-2026-M2P`

Implementation:
```go
func generateInvoiceNumber() string {
    year := time.Now().Year()
    chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    rand := make([]byte, 3)
    for i := range rand {
        rand[i] = chars[rand.Intn(len(chars))]
    }
    return fmt.Sprintf("INV-%d-%s", year, string(rand))
}
```

#### Revenue Sync on Mark-as-Paid

When invoice status transitions to `paid`:
1. Set `paid_at = now()`
2. Increment `invoice.amount_paid = invoice.amount`
3. Call `clients.UpdateRevenue(client_id, amount)` — add to client's `total_revenue`

#### API Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/invoices` | List all invoices for current org |
| `POST` | `/invoices` | Create invoice (auto-generate invoice number) |
| `GET` | `/invoices/:id` | Get single invoice |
| `PATCH` | `/invoices/:id` | Update invoice |
| `DELETE` | `/invoices/:id` | Delete invoice |

#### Data Model

```go
type Invoice struct {
    ID             uint         `json:"id"`
    OrganizationID uint         `json:"organization_id"`
    InvoiceNumber  string       `json:"invoice_number"`
    ClientID       *uint        `json:"client_id"`
    ProjectID      *uint        `json:"project_id"`
    Title          *string      `json:"title"`
    Amount         float64      `json:"amount"`
    Tax            float64      `json:"tax"`
    Discount       float64      `json:"discount"`
    AmountPaid     float64      `json:"amount_paid"`
    Status         string      `json:"status"` // draft|sent|pending|paid|overdue|cancelled
    DueDate        *time.Time   `json:"due_date"`
    PaidAt         *time.Time   `json:"paid_at"`
    Notes          *string      `json:"notes"`
    Items          []InvoiceItem `json:"items"`
    Version        int         `json:"version"`
    CreatedAt      time.Time   `json:"created_at"`
    UpdatedAt      time.Time   `json:"updated_at"`
}

type InvoiceItem struct {
    Description string  `json:"description"`
    Quantity    float64 `json:"quantity"`
    UnitPrice   float64 `json:"unit_price"`
    Total       float64 `json:"total"`
}
```

---

## Implementation Order

### Sprint 1: Critical Auth Fixes (Day 1)
1. [ ] Fix `POST /signup` → return `{ user, token }`
2. [ ] Fix `POST /login` → return `{ user, token }`
3. [ ] Add `GET /api/auth/me`
4. [ ] Add `POST /forgot-password`

### Sprint 2: Project Completeness (Day 1-2)
5. [ ] Add `GET /api/projects/:id`
6. [ ] Add `PATCH /api/projects/:id`

### Sprint 3: Clients Module (Day 2-3)
7. [ ] Create `modules/clients/` (models, repo, service, handler)
8. [ ] Wire up routes in `main.go`
9. [ ] Create migration for `clients` table
10. [ ] Run build verification

### Sprint 4: Invoices Module (Day 3-4)
11. [ ] Create `modules/invoices/` (models, repo, service, handler)
12. [ ] Wire up routes in `main.go`
13. [ ] Create migration for `invoices` table
14. [ ] Run build verification

### Sprint 5: Validation (Day 4)
15. [ ] Generate Swagger docs
16. [ ] Verify all endpoints match frontend expectations
17. [ ] Run full test suite

---

## Verification Checklist

After implementation, verify each frontend call works:

- [ ] `POST /api/auth/signup` → returns `{ user, token }` ✅ Frontend reads both
- [ ] `POST /api/auth/login` → returns `{ user, token }` ✅ Frontend reads both
- [ ] `GET /api/auth/me` → returns current user ✅ Frontend restores session
- [ ] `POST /forgot-password` → returns 200 ✅ Frontend shows success message
- [ ] `GET /api/projects/:id` → returns project ✅ Frontend loads project detail
- [ ] `PATCH /api/projects/:id` → updates and returns project ✅ Frontend saves project
- [ ] `GET /clients` → returns clients array ✅ Frontend lists clients
- [ ] `POST /clients` → creates and returns client ✅ Frontend adds client
- [ ] `GET /clients/stats` → returns `{ total, totalRevenue, avgRevenue }` ✅ Frontend shows stats
- [ ] `GET /invoices` → returns invoices array ✅ Frontend lists invoices
- [ ] `POST /invoices` → creates with auto invoice number ✅ Frontend generates INV-2026-XXX
- [ ] Mark invoice `paid` → updates client `totalRevenue` ✅ Revenue sync works

---

## Constraints

1. **No frontend changes** — This spec is backend-only
2. **STRICT LAYERING** — Handler → Service → Repository (no direct GORM in handlers)
3. **API Response format** — All endpoints use `utils/response.go` (`SendSuccess`/`SendError`)
4. **Personal org auto-resolve** — Clients/Invoices automatically scoped to user's personal org via middleware
5. **Optimistic locking** — Invoices table includes `version` column for future concurrency protection
6. **Build before commit** — `go build -o main .` must succeed before marking task complete