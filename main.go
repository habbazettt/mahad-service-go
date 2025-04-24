package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/habbazettt/mahad-service-go/config"
	_ "github.com/habbazettt/mahad-service-go/docs"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/routes"
	"github.com/sirupsen/logrus"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// @title Mahad Service API
// @version 1.0
// @description API untuk sistem Mahad (Absensi, Hafalan, dll)
// @host localhost:8080
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Masukkan token dengan format: Bearer {token}
func main() {
	config.LoadEnv()
	config.InitLogger()

	db := config.ConnectDB()
	config.MigrateDB()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173,https://t74dj4s7-5173.asse.devtunnels.ms",
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	app.Use(middleware.Logger())

	app.Get("/", func(c *fiber.Ctx) error {
		logrus.Info("🚀 Mahad Service API is running!")
		return c.SendString("🚀 Mahad Service API is running!")
	})

	routes.SetupAuthRoutes(app, db)
	routes.SetupMentorRoutes(app, db)
	routes.SetupMahasantriRoutes(app, db)
	routes.SetupHafalanRoutes(app, db)
	routes.SetupAbsensiRoutes(app, db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	logrus.Infof("Server berjalan di http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
