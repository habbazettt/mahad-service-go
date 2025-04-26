package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupTargetSemesterRoutes(app *fiber.App, db *gorm.DB) {
	service := services.TargetSemesterService{DB: db}

	targetSemesterRoutes := app.Group("/api/v1/target_semester", middleware.JWTMiddleware)
	{
		targetSemesterRoutes.Post("/", middleware.RoleMiddleware("mentor"), service.CreateTargetSemester)
		targetSemesterRoutes.Get("/", middleware.RoleMiddleware("mentor"), service.GetAllTargetSemesters)
		targetSemesterRoutes.Get("/mahasantri/:mahasantri_id", middleware.RoleMiddleware("mentor", "mahasantri"), service.GetTargetSemesterByMahasantriID)
		targetSemesterRoutes.Put("/:id", middleware.RoleMiddleware("mentor"), service.UpdateTargetSemester)
		targetSemesterRoutes.Delete("/:id", middleware.RoleMiddleware("mentor"), service.DeleteTargetSemester)
	}
}
