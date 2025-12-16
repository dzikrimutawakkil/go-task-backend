package models

import "gorm.io/gorm"

// ByOrg is a reusable filter for any table that has an 'organization_id' column
func ByOrg(orgID string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("organization_id = ?", orgID)
	}
}

// ByUser is a reusable filter for any table that has a 'user_id' column
func ByUser(userID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	}
}
