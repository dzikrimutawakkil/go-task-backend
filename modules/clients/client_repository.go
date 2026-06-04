package clients

import (
	"gorm.io/gorm"
)

// M-MIGRATION: Updated interface to use workspace instead of organization
type ClientRepository interface {
	FindAllByWorkspace(workspaceID string) ([]Client, error)
	FindByID(id uint) (*Client, error)
	Create(client *Client) error
	Update(client *Client) error
	Delete(client *Client) error
	UpdateRevenue(clientID uint, amount float64) error
	CountByWorkspace(workspaceID string) (int64, error)
	SumRevenueByWorkspace(workspaceID string) (float64, error)
}

type clientRepository struct {
	db *gorm.DB
}

func NewClientRepository(db *gorm.DB) ClientRepository {
	return &clientRepository{db}
}

// M-MIGRATION: Renamed from FindAllByOrg to FindAllByWorkspace
func (r *clientRepository) FindAllByWorkspace(workspaceID string) ([]Client, error) {
	var clients []Client
	err := r.db.Where("workspace_id = ?", workspaceID).
		Order("created_at DESC").
		Find(&clients).Error
	return clients, err
}

func (r *clientRepository) FindByID(id uint) (*Client, error) {
	var client Client
	if err := r.db.First(&client, id).Error; err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *clientRepository) Create(client *Client) error {
	return r.db.Create(client).Error
}

func (r *clientRepository) Update(client *Client) error {
	return r.db.Save(client).Error
}

func (r *clientRepository) Delete(client *Client) error {
	return r.db.Delete(client).Error
}

func (r *clientRepository) UpdateRevenue(clientID uint, amount float64) error {
	return r.db.Model(&Client{}).
		Where("id = ?", clientID).
		UpdateColumn("total_revenue", gorm.Expr("total_revenue + ?", amount)).Error
}

// M-MIGRATION: Renamed from CountByOrg to CountByWorkspace
func (r *clientRepository) CountByWorkspace(workspaceID string) (int64, error) {
	var count int64
	err := r.db.Model(&Client{}).Where("workspace_id = ?", workspaceID).Count(&count).Error
	return count, err
}

// M-MIGRATION: Renamed from SumRevenueByOrg to SumRevenueByWorkspace
func (r *clientRepository) SumRevenueByWorkspace(workspaceID string) (float64, error) {
	var sum float64
	err := r.db.Model(&Client{}).
		Where("workspace_id = ?", workspaceID).
		Select("COALESCE(SUM(total_revenue), 0)").
		Scan(&sum).Error
	return sum, err
}
