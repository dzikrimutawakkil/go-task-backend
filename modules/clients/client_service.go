package clients

import (
	"gotask-backend/internal/interfaces"
	"gotask-backend/utils"
	"strconv"
)

// M-MIGRATION: Updated to use workspace-based tier
type ClientService interface {
	GetClients(workspaceID string) ([]Client, error)
	GetClient(id uint) (*Client, error)
	CreateClient(input CreateClientInput) (*Client, error)
	UpdateClient(id uint, input UpdateClientInput) (*Client, error)
	DeleteClient(id uint) error
	GetClientStats(workspaceID string) (*ClientStats, error)
	AddRevenue(clientID uint, amount float64) error
}

type clientService struct {
	repo   ClientRepository
	wsRepo interfaces.WorkspaceFinder
	authS  interfaces.AuthService
}

// M5: Subscription Tiers — M-MIGRATION: uses workspace-based tier for quota checks
func NewClientService(repo ClientRepository, wsRepo interfaces.WorkspaceFinder, authS interfaces.AuthService) ClientService {
	return &clientService{repo: repo, wsRepo: wsRepo, authS: authS}
}

func (s *clientService) GetClients(workspaceID string) ([]Client, error) {
	return s.repo.FindAllByWorkspace(workspaceID)
}

func (s *clientService) GetClient(id uint) (*Client, error) {
	return s.repo.FindByID(id)
}

func (s *clientService) CreateClient(input CreateClientInput) (*Client, error) {
	// M-MIGRATION: Quota check — check client limit based on workspace's tier
	effectiveTier := "free"
	limits := utils.GetTierLimits("free")

	wsInfo, err := s.wsRepo.FindWorkspaceInfoByID(input.WorkspaceID)
	if err == nil {
		effectiveTier = utils.GetEffectiveTier(wsInfo.Tier, wsInfo.TierExpiresAt)
		limits = utils.GetTierLimits(effectiveTier)
	}

	if limits.MaxClients != -1 {
		count, err := s.repo.CountByWorkspace(strconv.FormatUint(uint64(input.WorkspaceID), 10))
		if err != nil {
			return nil, err
		}
		if int(count) >= limits.MaxClients {
			return nil, utils.ErrQuotaExceeded("client", limits.MaxClients, effectiveTier)
		}
	}

	client := Client{
		WorkspaceID:  input.WorkspaceID,
		Name:         input.Name,
		Email:        input.Email,
		WhatsApp:     input.WhatsApp,
		Phone:        input.Phone,
		Company:      input.Company,
		Website:      input.Website,
		Address:      input.Address,
		Notes:        input.Notes,
		TotalRevenue: 0,
	}
	if err := s.repo.Create(&client); err != nil {
		return nil, err
	}
	return &client, nil
}

func (s *clientService) UpdateClient(id uint, input UpdateClientInput) (*Client, error) {
	client, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		client.Name = *input.Name
	}
	if input.Email != nil {
		client.Email = input.Email
	}
	if input.WhatsApp != nil {
		client.WhatsApp = input.WhatsApp
	}
	if input.Phone != nil {
		client.Phone = input.Phone
	}
	if input.Company != nil {
		client.Company = input.Company
	}
	if input.Website != nil {
		client.Website = input.Website
	}
	if input.Address != nil {
		client.Address = input.Address
	}
	if input.Notes != nil {
		client.Notes = input.Notes
	}

	if err := s.repo.Update(client); err != nil {
		return nil, err
	}
	return client, nil
}

func (s *clientService) DeleteClient(id uint) error {
	client, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(client)
}

// M-MIGRATION: Renamed from GetClientStats
func (s *clientService) GetClientStats(workspaceID string) (*ClientStats, error) {
	total, err := s.repo.CountByWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	totalRevenue, err := s.repo.SumRevenueByWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	avgRevenue := 0
	if total > 0 {
		avgRevenue = int(totalRevenue / float64(total))
	}

	return &ClientStats{
		Total:        int(total),
		TotalRevenue: totalRevenue,
		AvgRevenue:   avgRevenue,
	}, nil
}

func (s *clientService) AddRevenue(clientID uint, amount float64) error {
	return s.repo.UpdateRevenue(clientID, amount)
}
