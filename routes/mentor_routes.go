package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupMentorRoutes(app *fiber.App, db *gorm.DB) {
	service := services.MentorService{DB: db}

	mentorRoutes := app.Group("/api/v1/mentors", middleware.JWTMiddleware)
	{
		mentorRoutes.Get("/", middleware.RoleMiddleware("mentor"), service.GetAllMentors)
		mentorRoutes.Get("/:id", middleware.RoleMiddleware("mentor"), service.GetMentorByID)
		mentorRoutes.Put("/:id", middleware.RoleMiddleware("mentor"), service.UpdateMentor)
		mentorRoutes.Delete("/:id", middleware.RoleMiddleware("mentor"), service.DeleteMentor)
	}
}
