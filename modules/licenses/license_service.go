package licenses

import (
	"errors"
	"strings"
	"time"
)

type LicenseService interface {
	ValidateKey(key string) (*ValidateResult, error)
	ActivateKey(key string, userID uint) (*License, error)
	CreateLicenseKeys(keys []LicenseInput) ([]LicenseResult, error)
}

type ValidateResult struct {
	Valid   bool         `json:"valid"`
	Message string       `json:"message,omitempty"`
	License *LicenseInfo `json:"license,omitempty"`
}

type LicenseInfo struct {
	ID   uint   `json:"id"`
	Type string `json:"type"`
}

type LicenseInput struct {
	Key  string `json:"key"`
	Type string `json:"type"`
}

type licenseService struct {
	repo LicenseRepository
}

func NewLicenseService(repo LicenseRepository) LicenseService {
	return &licenseService{repo: repo}
}

// ValidateKey validates a license key format and status (no auth required)
func (s *licenseService) ValidateKey(key string) (*ValidateResult, error) {
	// Normalize key
	key = strings.ToUpper(strings.TrimSpace(key))

	// Validate format: XXXX-XXXX-XXXX-XXXX
	if !isValidKeyFormat(key) {
		return nil, errors.New("invalid license key format")
	}

	// Find the license
	license, err := s.repo.FindByKey(key)
	if err != nil {
		return &ValidateResult{
			Valid:   false,
			Message: "License key not found",
		}, nil
	}

	// Check status
	switch license.Status {
	case LicenseStatusAvailable:
		return &ValidateResult{
			Valid: true,
			License: &LicenseInfo{
				ID:   license.ID,
				Type: license.Type,
			},
		}, nil
	case LicenseStatusActivated:
		return &ValidateResult{
			Valid:   false,
			Message: "Already activated",
		}, nil
	case LicenseStatusRevoked:
		return &ValidateResult{
			Valid:   false,
			Message: "License has been revoked",
		}, nil
	case LicenseStatusExpired:
		return &ValidateResult{
			Valid:   false,
			Message: "License has expired",
		}, nil
	default:
		return &ValidateResult{
			Valid:   false,
			Message: "License status is invalid",
		}, nil
	}
}

// ActivateKey activates a license key for a user
func (s *licenseService) ActivateKey(key string, userID uint) (*License, error) {
	// Normalize key
	key = strings.ToUpper(strings.TrimSpace(key))

	// Validate format first
	if !isValidKeyFormat(key) {
		return nil, errors.New("invalid license key format")
	}

	// Find the license
	license, err := s.repo.FindByKey(key)
	if err != nil {
		return nil, errors.New("license key not found")
	}

	// Check if already activated
	if license.Status == LicenseStatusActivated {
		return nil, errors.New("license already activated")
	}

	// Check if revoked
	if license.Status == LicenseStatusRevoked {
		return nil, errors.New("license has been revoked")
	}

	// Check if expired
	if license.ExpiresAt != nil && time.Now().After(*license.ExpiresAt) {
		return nil, errors.New("license has expired")
	}

	// Mark as activated
	now := time.Now()
	license.Status = LicenseStatusActivated
	license.ActivatedBy = &userID
	license.ActivatedAt = &now

	// Update license and user in transaction
	if err := s.repo.UpdateUserLicense(userID, license); err != nil {
		return nil, errors.New("failed to activate license")
	}

	return license, nil
}

// CreateLicenseKeys creates multiple license keys (admin only)
func (s *licenseService) CreateLicenseKeys(inputs []LicenseInput) ([]LicenseResult, error) {
	var licenses []License

	for _, input := range inputs {
		license := License{
			Key:    strings.ToUpper(strings.TrimSpace(input.Key)),
			Type:   input.Type,
			Status: LicenseStatusAvailable,
		}
		licenses = append(licenses, license)
	}

	return s.repo.CreateMany(licenses)
}

// isValidKeyFormat validates the license key format (XXXX-XXXX-XXXX-XXXX)
func isValidKeyFormat(key string) bool {
	if len(key) != 19 { // XXXX-XXXX-XXXX-XXXX = 19 chars
		return false
	}

	// Check pattern: 4 alphanumeric - 4 alphanumeric - 4 alphanumeric - 4 alphanumeric
	parts := strings.Split(key, "-")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if len(part) != 4 {
			return false
		}
		// Validate each char is alphanumeric
		for _, c := range part {
			if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
				return false
			}
		}
	}

	return true
}
