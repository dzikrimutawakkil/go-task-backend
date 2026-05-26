package invoices

import (
	"errors"
	"time"

	"gotask-backend/modules/clients"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Custom errors
var (
	ErrInvoiceCancelled   = errors.New("cannot mark a cancelled invoice as paid")
	ErrInvoiceAlreadyPaid = errors.New("invoice is already paid")
)

type InvoiceRepository interface {
	FindAllByOrg(orgID string) ([]Invoice, error)
	FindByID(id uint) (*Invoice, error)
	FindByInvoiceNumber(number string) (*Invoice, error)
	Create(invoice *Invoice) error
	Update(invoice *Invoice) error
	Delete(invoice *Invoice) error
	// Transactional operations
	MarkAsPaidAndSyncRevenue(invoiceID uint, clientID uint, amountPaid float64, paidAt *string) (*Invoice, error)
}

type invoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) FindAllByOrg(orgID string) ([]Invoice, error) {
	var invoices []Invoice
	err := r.db.Where("organization_id = ?", orgID).
		Order("created_at DESC").
		Find(&invoices).Error
	return invoices, err
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
