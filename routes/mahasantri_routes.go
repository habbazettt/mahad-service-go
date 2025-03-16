package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupMahasantriRoutes(app *fiber.App, db *gorm.DB) {
	service := services.MahasantriService{DB: db}

	mahasantriRoutes := app.Group("/api/v1/mahasantri", middleware.JWTMiddleware)

	mahasantriRoutes.Get("/", middleware.RoleMiddleware("mentor"), service.GetAllMahasantri)
	mahasantriRoutes.Get("/:id", middleware.RoleMiddleware("mentor", "mahasantri"), service.GetMahasantriByID)
	mahasantriRoutes.Get("/mentor/:mentor_id", middleware.RoleMiddleware("mentor"), service.GetMahasantriByMentorID)
	mahasantriRoutes.Put("/:id", middleware.RoleMiddleware("mentor", "mahasantri"), service.UpdateMahasantri)
	mahasantriRoutes.Delete("/:id", middleware.RoleMiddleware("mentor"), service.DeleteMahasantri)
}
