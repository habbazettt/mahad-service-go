package config

import (
	"log"
	"os"

	"github.com/habbazettt/mahad-service-go/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DB_URL")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Connected to database!")
	DB = db
}

func MigrateDB() {
	DB.AutoMigrate(&models.Mentor{}, &models.Mahasantri{}, &models.Hafalan{}, &models.Absensi{})
	log.Println("Database migrated successfully!")
}
