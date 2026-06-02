package clients

import (
	"gotask-backend/internal/interfaces"
	"gotask-backend/utils"
	"strconv"
)

type ClientService interface {
	GetClients(orgID string) ([]Client, error)
	GetClient(id uint) (*Client, error)
	CreateClient(input CreateClientInput) (*Client, error)
	UpdateClient(id uint, input UpdateClientInput) (*Client, error)
	DeleteClient(id uint) error
	GetClientStats(orgID string) (*ClientStats, error)
	AddRevenue(clientID uint, amount float64) error
}

type clientService struct {
	repo    ClientRepository
	orgRepo interfaces.OrgFinder
	authS   interfaces.AuthService
}

// M5: Subscription Tiers — added orgRepo and authS for quota checks.
func NewClientService(repo ClientRepository, orgRepo interfaces.OrgFinder, authS interfaces.AuthService) ClientService {
	return &clientService{repo: repo, orgRepo: orgRepo, authS: authS}
}

func (s *clientService) GetClients(orgID string) ([]Client, error) {
	return s.repo.FindAllByOrg(orgID)
}

func (s *clientService) GetClient(id uint) (*Client, error) {
	return s.repo.FindByID(id)
}

func (s *clientService) CreateClient(input CreateClientInput) (*Client, error) {
	// M5: Quota check — check client limit based on org owner's tier
	effectiveTier := "free"
	limits := utils.GetTierLimits("free")

	if orgInfo, err := s.orgRepo.FindOrgInfoByID(input.OrganizationID); err == nil {
		if owner, err := s.authS.FindByID(orgInfo.OwnerID); err == nil {
			effectiveTier = utils.GetEffectiveTier(owner.Tier, owner.TierExpiresAt)
			limits = utils.GetTierLimits(effectiveTier)
		}
	}

	if limits.MaxClients != -1 {
		count, err := s.repo.CountByOrg(strconv.FormatUint(uint64(input.OrganizationID), 10))
		if err != nil {
			return nil, err
		}
		if int(count) >= limits.MaxClients {
			return nil, utils.ErrQuotaExceeded("client", limits.MaxClients, effectiveTier)
		}
	}

	client := Client{
		OrganizationID: input.OrganizationID,
		Name:           input.Name,
		Email:          input.Email,
		WhatsApp:       input.WhatsApp,
		Phone:          input.Phone,
		Company:        input.Company,
		Website:        input.Website,
		Address:        input.Address,
		Notes:          input.Notes,
		TotalRevenue:   0,
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

func (s *clientService) GetClientStats(orgID string) (*ClientStats, error) {
	total, err := s.repo.CountByOrg(orgID)
	if err != nil {
		return nil, err
	}

	totalRevenue, err := s.repo.SumRevenueByOrg(orgID)
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
