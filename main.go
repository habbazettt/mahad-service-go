package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/config"
)

func main() {
	config.ConnectDB()
	config.MigrateDB()

	app := fiber.New()

	log.Fatal(app.Listen(":8080"))
}
