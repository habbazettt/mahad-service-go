package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupMahasantriRoutes(app *fiber.App, db *gorm.DB) {
	service := services.MahasantriService{DB: db}

	mahasantriLimiter := limiter.New(limiter.Config{
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
			return mahasantriLimiter(c)
		}
		return c.Next()
	}

	mahasantriRoutes := app.Group("/api/v1/mahasantri", methodLimiter)
	{
		mahasantriRoutes.Get("/", service.GetAllMahasantri)
		mahasantriRoutes.Get("/:id", middleware.JWTMiddleware, middleware.RoleMiddleware("mentor", "mahasantri"), service.GetMahasantriByID)
		mahasantriRoutes.Get("/mentor/:mentor_id", middleware.JWTMiddleware, middleware.RoleMiddleware("mentor"), service.GetMahasantriByMentorID)
		mahasantriRoutes.Put("/:id", middleware.JWTMiddleware, middleware.RoleMiddleware("mentor", "mahasantri"), service.UpdateMahasantri)
		mahasantriRoutes.Delete("/:id", middleware.JWTMiddleware, middleware.RoleMiddleware("mentor"), service.DeleteMahasantri)
	}
}
