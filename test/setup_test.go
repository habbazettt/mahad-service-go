package test

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/models"
	"github.com/habbazettt/mahad-service-go/services"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupTestApp() (*fiber.App, *gorm.DB) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dsn := os.Getenv("TEST_DB_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	db.Migrator().DropTable(&models.Mentor{}, &models.Mahasantri{})
	db.AutoMigrate(&models.Mentor{}, &models.Mahasantri{})

	app := fiber.New()
	authService := services.AuthService{DB: db}

	api := app.Group("/api/v1")
	auth := api.Group("/auth")
	auth.Post("/login/mentor", authService.LoginMentor)
	auth.Post("/login/mahasantri", authService.LoginMahasantri)

	return app, db
}
