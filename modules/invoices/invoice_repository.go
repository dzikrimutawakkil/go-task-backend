package invoices

import (
	"errors"
	"time"

	"gotask-backend/modules/clients"
	"gotask-backend/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Custom errors
var (
	ErrInvoiceCancelled   = errors.New("cannot mark a cancelled invoice as paid")
	ErrInvoiceAlreadyPaid = errors.New("invoice is already paid")
)

// M-MIGRATION: Updated interface to use workspace instead of organization
type InvoiceRepository interface {
	FindAllByWorkspace(workspaceID string) ([]Invoice, error)
	FindByID(id uint) (*Invoice, error)
	FindByInvoiceNumber(number string) (*Invoice, error)
	Create(invoice *Invoice) error
	Update(invoice *Invoice) error
	Delete(invoice *Invoice) error
	// Transactional operations
	MarkAsPaidAndSyncRevenue(invoiceID uint, clientID uint, amountPaid float64, paidAt *string) (*Invoice, error)

	// M5: Quota check helpers
	CountThisMonth(workspaceID string) (int, error)
}

type invoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &invoiceRepository{db: db}
}

// M-MIGRATION: Renamed from FindAllByOrg to FindAllByWorkspace
func (r *invoiceRepository) FindAllByWorkspace(workspaceID string) ([]Invoice, error) {
	var invoices []Invoice

	// Use Raw SQL to avoid GORM automatic scanning issues with JSONB
	// This explicitly handles the items JSONB column
	rows, err := r.db.Raw(`
		SELECT id, workspace_id, invoice_number, client_id, project_id, title,
		       amount, tax, discount, amount_paid, status, due_date, paid_at,
		       notes, version, created_at, updated_at
		FROM invoices
		WHERE workspace_id = ?
		ORDER BY created_at DESC
	`, workspaceID).Rows()
	if err != nil {
		utils.GetLogger().Error("FindAllByWorkspace query failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var inv Invoice
		var clientID, projectID *uint
		var title, notes *string
		var dueDate, paidAt *time.Time

		err := rows.Scan(
			&inv.ID, &inv.WorkspaceID, &inv.InvoiceNumber, &clientID, &projectID,
			&title, &inv.Amount, &inv.Tax, &inv.Discount, &inv.AmountPaid,
			&inv.Status, &dueDate, &paidAt, &notes, &inv.Version,
			&inv.CreatedAt, &inv.UpdatedAt,
		)
		if err != nil {
			utils.GetLogger().Error("Row scan failed", "error", err)
			continue
		}
		inv.ClientID = clientID
		inv.ProjectID = projectID
		inv.Title = title
		inv.Notes = notes
		inv.DueDate = dueDate
		inv.PaidAt = paidAt

		// Items JSONB column is handled by GORM for create/update operations
		// For list queries, items are left as nil (parsed separately when needed)

		invoices = append(invoices, inv)
	}

	return invoices, nil
}

func (r *invoiceRepository) FindByID(id uint) (*Invoice, error) {
	var invoice Invoice
	if err := r.db.First(&invoice, id).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) FindByInvoiceNumber(number string) (*Invoice, error) {
	var invoice Invoice
	if err := r.db.Where("invoice_number = ?", number).First(&invoice).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) Create(invoice *Invoice) error {
	return r.db.Create(invoice).Error
}

func (r *invoiceRepository) Update(invoice *Invoice) error {
	return r.db.Save(invoice).Error
}

func (r *invoiceRepository) Delete(invoice *Invoice) error {
	return r.db.Delete(invoice).Error
}

// MarkAsPaidAndSyncRevenue marks an invoice as paid and syncs revenue to client atomically
func (r *invoiceRepository) MarkAsPaidAndSyncRevenue(invoiceID uint, clientID uint, amountPaid float64, paidAt *string) (*Invoice, error) {
	var invoice Invoice
	var client clients.Client

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Lock the invoice for update
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&invoice, invoiceID).Error; err != nil {
			return err
		}

		// Check if already cancelled
		if invoice.Status == "cancelled" {
			return ErrInvoiceCancelled
		}

		// Set paid fields
		invoice.AmountPaid = amountPaid
		invoice.Status = "paid"
		if paidAt != nil {
			// Parse the paid_at timestamp
			paidAtTime, err := time.Parse(time.RFC3339, *paidAt)
			if err == nil {
				invoice.PaidAt = &paidAtTime
			}
		} else {
			now := tx.NowFunc()
			invoice.PaidAt = &now
		}

		// Update invoice
		if err := tx.Save(&invoice).Error; err != nil {
			return err
		}

		// Sync revenue to client in the same transaction
		if clientID > 0 {
			if err := tx.First(&client, clientID).Error; err != nil {
				return err
			}
			client.TotalRevenue += amountPaid
			if err := tx.Save(&client).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return &invoice, err
}

// CountThisMonth returns the number of invoices created this month for a workspace.
// M5: Subscription Tiers — Phase 5: Service Layer — Quota check for invoice limit.
// M-MIGRATION: Renamed from CountThisMonth
func (r *invoiceRepository) CountThisMonth(workspaceID string) (int, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var count int64
	err := r.db.Model(&Invoice{}).
		Where("workspace_id = ? AND created_at >= ?", workspaceID, monthStart).
		Count(&count).Error
	return int(count), err
}
