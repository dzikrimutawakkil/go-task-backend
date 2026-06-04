package models

import "gorm.io/gorm"

// ByWorkspace is a reusable filter for any table that has a 'workspace_id' column.
// M-MIGRATION: Renamed from ByOrg to ByWorkspace.
func ByWorkspace(workspaceID string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("workspace_id = ?", workspaceID)
	}
}

// ByUser is a reusable filter for any table that has a 'user_id' column
func ByUser(userID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	}
}
