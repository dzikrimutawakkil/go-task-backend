package invoices

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"gotask-backend/modules/clients"
)

type InvoiceService interface {
	GetInvoices(orgID string) ([]Invoice, error)
	GetInvoice(id uint) (*Invoice, error)
	CreateInvoice(input CreateInvoiceInput) (*Invoice, error)
	UpdateInvoice(id uint, input UpdateInvoiceInput) (*Invoice, error)
	DeleteInvoice(id uint) error
	MarkAsPaid(id uint, input MarkAsPaidInput) (*Invoice, error)
	GenerateInvoiceNumber() string
}

type invoiceService struct {
	repo       InvoiceRepository
	clientRepo clients.ClientRepository
}

func NewInvoiceService(repo InvoiceRepository, clientRepo clients.ClientRepository) InvoiceService {
	return &invoiceService{repo: repo, clientRepo: clientRepo}
}

func (s *invoiceService) GetInvoices(orgID string) ([]Invoice, error) {
	return s.repo.FindAllByOrg(orgID)
}

func (s *invoiceService) GetInvoice(id uint) (*Invoice, error) {
	return s.repo.FindByID(id)
}

// GenerateInvoiceNumber creates INV-YYYY-XXX format
// Example: INV-2026-A7K
func (s *invoiceService) GenerateInvoiceNumber() string {
	year := time.Now().Year()
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randBytes := make([]byte, 3)
	rand.Read(randBytes)
	var result strings.Builder
	for _, b := range randBytes {
		result.WriteByte(chars[int(b)%len(chars)])
	}
	return fmt.Sprintf("INV-%d-%s", year, result.String())
}

func (s *invoiceService) CreateInvoice(input CreateInvoiceInput) (*Invoice, error) {
	// Generate unique invoice number
	invoiceNumber := s.GenerateInvoiceNumber()
	for {
		_, err := s.repo.FindByInvoiceNumber(invoiceNumber)
		if err != nil {
			break // unique
		}
		invoiceNumber = s.GenerateInvoiceNumber()
	}

	invoice := Invoice{
		OrganizationID: input.OrganizationID,
		InvoiceNumber:  invoiceNumber,
		ClientID:       input.ClientID,
		ProjectID:      input.ProjectID,
		Title:          input.Title,
		Amount:         input.Amount,
		Tax:            input.Tax,
		Discount:       input.Discount,
		AmountPaid:     0,
		Status:         "draft",
		DueDate:        input.DueDate,
		Notes:          input.Notes,
		Items:          input.Items,
		Version:        1,
	}

	if err := s.repo.Create(&invoice); err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (s *invoiceService) UpdateInvoice(id uint, input UpdateInvoiceInput) (*Invoice, error) {
	invoice, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		invoice.Title = input.Title
	}
	if input.ClientID != nil {
		invoice.ClientID = input.ClientID
	}
	if input.ProjectID != nil {
		invoice.ProjectID = input.ProjectID
	}
	if input.Amount != nil {
		invoice.Amount = *input.Amount
	}
	if input.Tax != nil {
		invoice.Tax = *input.Tax
	}
	if input.Discount != nil {
		invoice.Discount = *input.Discount
	}
	if input.Status != nil {
		oldStatus := invoice.Status
		invoice.Status = *input.Status

		// Revenue sync: when marking as paid
		if oldStatus != "paid" && *input.Status == "paid" {
			invoice.AmountPaid = invoice.Amount
			now := time.Now()
			invoice.PaidAt = &now

			// Sync revenue to client
			if invoice.ClientID != nil && s.clientRepo != nil {
				s.clientRepo.UpdateRevenue(*invoice.ClientID, invoice.Amount)
			}
		}
	}
	if input.DueDate != nil {
		invoice.DueDate = input.DueDate
	}
	if input.PaidAt != nil {
		invoice.PaidAt = input.PaidAt
	}
	if input.Notes != nil {
		invoice.Notes = input.Notes
	}
	if input.Items != nil {
		invoice.Items = input.Items
	}

	if err := s.repo.Update(invoice); err != nil {
		return nil, err
	}
	return invoice, nil
}

func (s *invoiceService) DeleteInvoice(id uint) error {
	invoice, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(invoice)
}

// MarkAsPaidInput represents the input for marking an invoice as paid
type MarkAsPaidInput struct {
	AmountPaid float64
	PaidAt     *string
}

// MarkAsPaid marks an invoice as paid with atomic revenue sync
func (s *invoiceService) MarkAsPaid(id uint, input MarkAsPaidInput) (*Invoice, error) {
	// Validate amount_paid is not negative
	if input.AmountPaid < 0 {
		return nil, errors.New("amount_paid cannot be negative")
	}

	// Get the invoice first to check status and get client_id
	invoice, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Check if already paid
	if invoice.Status == "paid" {
		return nil, ErrInvoiceAlreadyPaid
	}

	// Check if already cancelled
	if invoice.Status == "cancelled" {
		return nil, ErrInvoiceCancelled
	}

	var clientID uint
	if invoice.ClientID != nil {
		clientID = *invoice.ClientID
	}

	// Use the transactional method
	return s.repo.MarkAsPaidAndSyncRevenue(id, clientID, input.AmountPaid, input.PaidAt)
}
