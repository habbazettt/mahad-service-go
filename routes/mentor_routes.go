package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupMentorRoutes(app *fiber.App, db *gorm.DB) {
	service := services.MentorService{DB: db}

	mentorLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many modification requests, please try again later",
			})
		},
	})

	methodLimiter := func(c *fiber.Ctx) error {
		if c.Method() == fiber.MethodPut || c.Method() == fiber.MethodDelete {
			return mentorLimiter(c)
		}
		return c.Next()
	}

	mentorRoutes := app.Group("/api/v1/mentors", methodLimiter)
	{
		mentorRoutes.Get("/", service.GetAllMentors)
		mentorRoutes.Get("/:id", service.GetMentorByID)
		mentorRoutes.Put("/:id", middleware.JWTMiddleware, middleware.RoleMiddleware("mentor"), service.UpdateMentor)
		mentorRoutes.Delete("/:id", middleware.JWTMiddleware, middleware.RoleMiddleware("mentor"), service.DeleteMentor)
	}
}
