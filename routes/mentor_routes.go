package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupMentorRoutes(app *fiber.App, db *gorm.DB) {
	service := services.MentorService{DB: db}

	mentorRoutes := app.Group("/api/v1/mentors")
	{
		mentorRoutes.Get("/", service.GetAllMentors)
		mentorRoutes.Get("/:id", service.GetMentorByID)
		mentorRoutes.Put("/:id", middleware.JWTMiddleware, middleware.RoleMiddleware("mentor"), service.UpdateMentor)
		mentorRoutes.Delete("/:id", middleware.JWTMiddleware, middleware.RoleMiddleware("mentor"), service.DeleteMentor)
	}
}
