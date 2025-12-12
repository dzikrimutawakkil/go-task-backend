package config

import (
	"fmt"
	"gotask-backend/models"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	// ... (Your existing .env loading code) ...
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
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

	// 1. AutoMigrate the new Status struct
	database.AutoMigrate(&models.User{}, &models.Project{}, &models.Task{}, &models.Status{})

	DB = database

	// 2. Run Seeder
	seedStatuses()

	fmt.Println("Database connected and seeded!")
}

// 3. Seeder Logic
func seedStatuses() {
	statuses := []string{"Todo", "In Progress", "Done", "Pending", "Canceled"}

	for _, name := range statuses {
		var status models.Status
		slug := strings.ToLower(strings.ReplaceAll(name, " ", "_"))

		// Create status if it doesn't exist
		DB.FirstOrCreate(&status, models.Status{Name: name, Slug: slug})
	}
}
