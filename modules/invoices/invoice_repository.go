package invoices

import (
	"gorm.io/gorm"
)

type InvoiceRepository interface {
	FindAllByOrg(orgID string) ([]Invoice, error)
	FindByID(id uint) (*Invoice, error)
	FindByInvoiceNumber(number string) (*Invoice, error)
	Create(invoice *Invoice) error
	Update(invoice *Invoice) error
	Delete(invoice *Invoice) error
}

type invoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &invoiceRepository{db}
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