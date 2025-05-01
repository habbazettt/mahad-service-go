package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupAbsensiRoutes(app *fiber.App, db *gorm.DB) {
	absensiService := services.AbsensiService{DB: db}

	absensiRoutes := app.Group("/api/v1/absensi", middleware.JWTMiddleware)
	{
		absensiRoutes.Post("/", middleware.RoleMiddleware("mentor"), absensiService.CreateAbsensi)
		absensiRoutes.Get("/", middleware.RoleMiddleware("mentor"), absensiService.GetAbsensi)
		absensiRoutes.Put("/:id", middleware.RoleMiddleware("mentor"), absensiService.UpdateAbsensi)
		absensiRoutes.Get("/mahasantri/:mahasantri_id/daily-summary", middleware.RoleMiddleware("mentor", "mahasantri"), absensiService.GetAbsensiDailySummary)
		absensiRoutes.Delete("/:id", middleware.RoleMiddleware("mentor"), absensiService.DeleteAbsensi)
	}
}
