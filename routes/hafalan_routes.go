package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupHafalanRoutes(app *fiber.App, db *gorm.DB) {
	service := services.HafalanService{DB: db}

	hafalanRoutes := app.Group("/api/v1/hafalan", middleware.JWTMiddleware)

	hafalanRoutes.Post("/", middleware.RoleMiddleware("mentor"), service.CreateHafalan)
	hafalanRoutes.Get("/", middleware.RoleMiddleware("mentor"), service.GetAllHafalan)
	hafalanRoutes.Get("/:id", middleware.RoleMiddleware("mentor"), service.GetHafalanByID)
	hafalanRoutes.Get("/mahasantri/:mahasantri_id", middleware.RoleMiddleware("mentor", "mahasantri"), service.GetHafalanByMahasantriID)
	hafalanRoutes.Get("/:mahasantri_id/kategori", middleware.RoleMiddleware("mentor"), service.GetHafalanByKategori)
	hafalanRoutes.Put("/:id", middleware.RoleMiddleware("mentor"), service.UpdateHafalan)
	hafalanRoutes.Delete("/:id", middleware.RoleMiddleware("mentor"), service.DeleteHafalan)
}
