package config

import (
	"fmt"
	"gotask-backend/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	// 1. Load the .env file
	// If this fails, the app should probably crash because it can't run without config
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// 2. Read variables from the environment
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	// 3. Construct the DSN (Data Source Name) string
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	// 4. Connect
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}

	// 5. Migrate
	database.AutoMigrate(&models.User{}, &models.Project{}, &models.Task{})

	DB = database
	fmt.Println("Database connected successfully using .env!")
}
