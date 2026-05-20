package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDatabase initializes the GORM database connection and runs migrations.
func ConnectDatabase() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system env variables")
	}

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}

	DB = database

	// Q10: Run golang-migrate migrations on startup
	runMigrations()

	fmt.Println("Database connected and migrated!")
}

// runMigrations runs all pending SQL migrations using golang-migrate.
// Exits the process if migrations fail (fail-fast — do not start with bad schema).
func runMigrations() {
	migrationURL := os.Getenv("MIGRATION_URL")
	if migrationURL == "" {
		// Build migration URL from env vars
		host := os.Getenv("DB_HOST")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		port := os.Getenv("DB_PORT")
		migrationURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			user, password, host, port, dbname,
		)
	}

	// Get the path to the migrations folder (relative to working directory)
	execDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Cannot determine working directory for migrations")
	}
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		// Default: look for migrations/ folder in project root
		migrationsPath = filepath.Join(execDir, "migrations")
	}

	// IMPORTANT: file:// URLs MUST use forward slashes, NOT Windows backslashes
	migrationsPathClean := filepath.ToSlash(migrationsPath)

	// Use the file driver to open the migrations directory
	fSource, err := (&file.File{}).Open(migrationsPathClean)
	if err != nil {
		log.Fatalf("Failed to open migrations source directory: %v", err)
	}

	m, err := migrate.NewWithSourceInstance("file", fSource, migrationURL)
	if err != nil {
		log.Fatalf("Failed to create migration instance: %v", err)
	}
	defer func() {
		// Perbaikan: m.Close() mengembalikan dua error (source error, database error)
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			log.Printf("Warning: failed to close migration source: %v", sourceErr)
		}
		if dbErr != nil {
			log.Printf("Warning: failed to close migration database: %v", dbErr)
		}
	}()

	// Run up migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	// Log current version
	version, dirty, verErr := m.Version()
	if verErr == nil {
		if dirty {
			log.Printf("WARNING: Database is at version %d and marked DIRTY", version)
		} else {
			log.Printf("Database migration: version=%d, dirty=false", version)
		}
	}
}