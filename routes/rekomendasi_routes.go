package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupRekomendasiRoutes(app *fiber.App, db *gorm.DB) {
	service := services.NewRekomendasiService(db)

	rekomendasiLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many write requests, please try again later",
			})
		},
		SkipSuccessfulRequests: true,
	})

	methodLimiter := func(c *fiber.Ctx) error {
		if c.Method() == fiber.MethodPost {
			return rekomendasiLimiter(c)
		}
		return c.Next()
	}

	rekomendasiRoutes := app.Group("/api/v1/rekomendasi", middleware.JWTMiddleware, methodLimiter)
	{
		rekomendasiRoutes.Post("/", middleware.RoleMiddleware("mentor", "mahasantri"), service.GetRecommendation)
		rekomendasiRoutes.Get("/", middleware.RoleMiddleware("mentor", "mahasantri"), service.GetAllRekomendasi)
		rekomendasiRoutes.Get("/kesibukan", middleware.RoleMiddleware("mentor", "mahasantri"), service.GetAllKesibukan)
	}
}
