package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupJadwalPersonalRoutes(app *fiber.App, db *gorm.DB) {
	service := services.NewJadwalPersonalService(db)

	jadwalRoutes := app.Group("/api/v1/jadwal-personal", middleware.JWTMiddleware)
	{
		jadwalRoutes.Get("/all", middleware.RoleMiddleware("mentor"), service.GetAllJadwalPersonal)
		jadwalRoutes.Get("/", middleware.RoleMiddleware("mahasantri", "mentor"), service.GetJadwalPersonal)
		jadwalRoutes.Post("/", middleware.RoleMiddleware("mahasantri", "mentor"), service.CreateJadwalPersonal)
		jadwalRoutes.Put("/", middleware.RoleMiddleware("mahasantri", "mentor"), service.UpdateJadwalPersonal)
	}
}
