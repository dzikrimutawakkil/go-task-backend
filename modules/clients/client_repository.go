package clients

import (
	"gorm.io/gorm"
)

type ClientRepository interface {
	FindAllByOrg(orgID string) ([]Client, error)
	FindByID(id uint) (*Client, error)
	Create(client *Client) error
	Update(client *Client) error
	Delete(client *Client) error
	UpdateRevenue(clientID uint, amount float64) error
	CountByOrg(orgID string) (int64, error)
	SumRevenueByOrg(orgID string) (float64, error)
}

type clientRepository struct {
	db *gorm.DB
}

func NewClientRepository(db *gorm.DB) ClientRepository {
	return &clientRepository{db}
}

func (r *clientRepository) FindAllByOrg(orgID string) ([]Client, error) {
	var clients []Client
	err := r.db.Where("organization_id = ?", orgID).
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

func (r *clientRepository) CountByOrg(orgID string) (int64, error) {
	var count int64
	err := r.db.Model(&Client{}).Where("organization_id = ?", orgID).Count(&count).Error
	return count, err
}

func (r *clientRepository) SumRevenueByOrg(orgID string) (float64, error) {
	var sum float64
	err := r.db.Model(&Client{}).
		Where("organization_id = ?", orgID).
		Select("COALESCE(SUM(total_revenue), 0)").
		Scan(&sum).Error
	return sum, err
}
