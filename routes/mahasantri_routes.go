package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupMahasantriRoutes(app *fiber.App, db *gorm.DB) {
	service := services.MahasantriService{DB: db}

	mahasantriRoutes := app.Group("/api/v1/mahasantri")
	{
		mahasantriRoutes.Get("/", service.GetAllMahasantri)
		mahasantriRoutes.Get("/:id", middleware.JWTMiddleware, middleware.RoleMiddleware("mentor", "mahasantri"), service.GetMahasantriByID)
		mahasantriRoutes.Get("/mentor/:mentor_id", middleware.JWTMiddleware, middleware.RoleMiddleware("mentor"), service.GetMahasantriByMentorID)
		mahasantriRoutes.Put("/:id", middleware.JWTMiddleware, middleware.RoleMiddleware("mentor", "mahasantri"), service.UpdateMahasantri)
		mahasantriRoutes.Delete("/:id", middleware.JWTMiddleware, middleware.RoleMiddleware("mentor"), service.DeleteMahasantri)
	}
}
