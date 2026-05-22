package invoices

import (
	"encoding/json"
	"time"
)

type Invoice struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	OrganizationID uint         `json:"organization_id"`
	InvoiceNumber  string       `json:"invoice_number"`
	ClientID       *uint        `json:"client_id"`
	ProjectID      *uint        `json:"project_id"`
	Title          *string      `json:"title"`
	Amount         float64      `json:"amount"`
	Tax            float64      `gorm:"default:0" json:"tax"`
	Discount       float64      `gorm:"default:0" json:"discount"`
	AmountPaid     float64      `gorm:"default:0" json:"amount_paid"`
	Status         string       `gorm:"default:draft" json:"status"`
	DueDate        *time.Time   `json:"due_date"`
	PaidAt         *time.Time   `json:"paid_at"`
	Notes          *string      `json:"notes"`
	Items          InvoiceItems `gorm:"type:jsonb" json:"items"`
	Version        int          `gorm:"default:1" json:"version"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

type InvoiceItem struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
}

type InvoiceItems []InvoiceItem

func (ij *InvoiceItems) Scan(value interface{}) error {
	if value == nil {
		*ij = InvoiceItems{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		*ij = InvoiceItems{}
		return nil
	}
	return json.Unmarshal(bytes, ij)
}

// DTOs
type CreateInvoiceInput struct {
	OrganizationID uint
	ClientID       *uint
	ProjectID      *uint
	Title          *string
	Amount         float64
	Tax            float64
	Discount       float64
	DueDate        *time.Time
	Notes          *string
	Items          InvoiceItems
}

type UpdateInvoiceInput struct {
	ClientID   *uint
	ProjectID  *uint
	Title      *string
	Amount     *float64
	Tax        *float64
	Discount   *float64
	Status     *string
	DueDate    *time.Time
	PaidAt     *time.Time
	Notes      *string
	Items      InvoiceItems
}