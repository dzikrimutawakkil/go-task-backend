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

	database.AutoMigrate(&models.User{}, &models.Project{}, &models.Task{}, &models.Status{}, &models.Priority{})

	DB = database

	seedStatuses()
	seedPriority()

	fmt.Println("Database connected and seeded!")
}

func seedStatuses() {
	statuses := []string{"Todo", "In Progress", "Done", "Pending", "Canceled"}

	for _, name := range statuses {
		var status models.Status
		slug := strings.ToLower(strings.ReplaceAll(name, " ", "_"))

		DB.FirstOrCreate(&status, models.Status{Name: name, Slug: slug})
	}
}

func seedPriority() {
	priorities := []models.Priority{
		{Name: "Low", Level: 1, Color: "#808080"},    // Gray
		{Name: "Medium", Level: 2, Color: "#0000FF"}, // Blue
		{Name: "High", Level: 3, Color: "#FFA500"},   // Orange
		{Name: "Urgent", Level: 4, Color: "#FF0000"}, // Red
	}

	for _, p := range priorities {
		var exists models.Priority
		if DB.Where("name = ?", p.Name).First(&exists).Error != nil {
			DB.Create(&p)
		}
	}
}
