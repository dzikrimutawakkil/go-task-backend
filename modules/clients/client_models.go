package clients

import (
	"time"
)

// Client represents a client/customer.
// M-MIGRATION: Renamed organization_id to workspace_id
type Client struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	WorkspaceID  uint      `json:"workspace_id"`
	Name         string    `json:"name"`
	Email        *string   `json:"email"`
	WhatsApp     *string   `gorm:"column:whatsapp" json:"whatsapp"`
	Phone        *string   `json:"phone"`
	Company      *string   `json:"company"`
	Website      *string   `json:"website"`
	Address      *string   `json:"address"`
	Notes        *string   `json:"notes"`
	TotalRevenue float64   `gorm:"default:0" json:"total_revenue"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// DTOs
type CreateClientInput struct {
	Name        string
	Email       *string
	WhatsApp    *string
	Phone       *string
	Company     *string
	Website     *string
	Address     *string
	Notes       *string
	WorkspaceID uint
}

type UpdateClientInput struct {
	Name         *string
	Email        *string
	WhatsApp     *string
	Phone        *string
	Company      *string
	Website      *string
	Address      *string
	Notes        *string
	TotalRevenue *float64
}

type ClientStats struct {
	Total        int     `json:"total"`
	TotalRevenue float64 `json:"totalRevenue"`
	AvgRevenue   int     `json:"avgRevenue"`
}
