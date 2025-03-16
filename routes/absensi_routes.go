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

	absensiRoutes.Post("/", middleware.RoleMiddleware("mentor"), absensiService.CreateAbsensi)
	absensiRoutes.Get("/mahasantri/:mahasantri_id", middleware.RoleMiddleware("mentor", "mahasantri"), absensiService.GetAbsensiByMahasantriID)
	absensiRoutes.Get("/mahasantri/:mahasantri_id/per-month", middleware.RoleMiddleware("mentor", "mahasantri"), absensiService.GetAttendancePerMonth)
}
