package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/config"
	"github.com/habbazettt/mahad-service-go/routes"
	"github.com/sirupsen/logrus"
)

func main() {
	config.LoadEnv()
	config.InitLogger()

	db := config.ConnectDB()
	config.MigrateDB()

	app := fiber.New(fiber.Config{
		EnablePrintRoutes: true,
	})

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

	logrus.Infof("Server berjalan di http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
