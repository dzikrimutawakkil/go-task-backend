package licenses

import (
	"crypto/rand"
	"errors"
	"strings"

	"gotask-backend/modules/auth"

	"gorm.io/gorm"
)

type LicenseRepository interface {
	FindByKey(key string) (*License, error)
	Create(license *License) error
	CreateMany(licenses []License) ([]LicenseResult, error)
	Update(license *License) error
	UpdateUserLicense(userID uint, license *License) error
}

type LicenseResult struct {
	Key    string `json:"key"`
	Status string `json:"status"` // created or error
	ID     uint   `json:"id,omitempty"`
	Error  string `json:"error,omitempty"`
}

type licenseRepository struct {
	db *gorm.DB
}

func NewLicenseRepository(db *gorm.DB) LicenseRepository {
	return &licenseRepository{db: db}
}

func (r *licenseRepository) FindByKey(key string) (*License, error) {
	var license License
	// Always normalize key to uppercase
	key = strings.ToUpper(strings.TrimSpace(key))
	err := r.db.Where("key = ?", key).First(&license).Error
	return &license, err
}

func (r *licenseRepository) Create(license *License) error {
	// Normalize key
	license.Key = strings.ToUpper(strings.TrimSpace(license.Key))
	return r.db.Create(license).Error
}

// CreateMany creates multiple licenses and returns results per key
func (r *licenseRepository) CreateMany(licenses []License) ([]LicenseResult, error) {
	// Normalize all keys
	for i := range licenses {
		licenses[i].Key = strings.ToUpper(strings.TrimSpace(licenses[i].Key))
	}

	var results []LicenseResult

	err := r.db.Transaction(func(tx *gorm.DB) error {
		for _, lic := range licenses {
			result := LicenseResult{Key: lic.Key}

			// Check if key already exists
			var existing License
			err := tx.Where("key = ?", lic.Key).First(&existing).Error
			if err == nil {
				// Key already exists
				result.Status = "error"
				result.Error = "Key already exists"
				results = append(results, result)
				continue
			}

			// Create the license
			if err := tx.Create(&lic).Error; err != nil {
				result.Status = "error"
				result.Error = "Failed to create license"
				results = append(results, result)
				continue
			}

			result.Status = "created"
			result.ID = lic.ID
			results = append(results, result)
		}

		return nil
	})

	return results, err
}

func (r *licenseRepository) Update(license *License) error {
	return r.db.Save(license).Error
}

// UpdateUserLicense updates both user and license in a transaction
func (r *licenseRepository) UpdateUserLicense(userID uint, license *License) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update the license record
		if err := tx.Save(license).Error; err != nil {
			return err
		}

		// Update the user record
		if err := tx.Model(&auth.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
			"plan":           license.Type,
			"license_key":    license.Key,
			"license_status": LicenseStatusActivated,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

// GenerateLicenseKey generates a secure random license key
// Format: XXXX-XXXX-XXXX-XXXX (uppercase alphanumeric)
func GenerateLicenseKey() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 16
	const segmentLength = 4

	bytes := make([]byte, keyLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", errors.New("failed to generate license key")
	}

	var key strings.Builder
	for i, b := range bytes {
		key.WriteByte(charset[int(b)%len(charset)])
		if (i+1)%segmentLength == 0 && i < keyLength-1 {
			key.WriteByte('-')
		}
	}

	return key.String(), nil
}
