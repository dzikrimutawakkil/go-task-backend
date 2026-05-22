package clients

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
	repo ClientRepository
}

func NewClientService(repo ClientRepository) ClientService {
	return &clientService{repo: repo}
}

func (s *clientService) GetClients(orgID string) ([]Client, error) {
	return s.repo.FindAllByOrg(orgID)
}

func (s *clientService) GetClient(id uint) (*Client, error) {
	return s.repo.FindByID(id)
}

func (s *clientService) CreateClient(input CreateClientInput) (*Client, error) {
	client := Client{
		OrganizationID: input.OrganizationID,
		Name:          input.Name,
		Email:         input.Email,
		WhatsApp:      input.WhatsApp,
		Phone:         input.Phone,
		Company:       input.Company,
		Website:       input.Website,
		Address:       input.Address,
		Notes:         input.Notes,
		TotalRevenue:  0,
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