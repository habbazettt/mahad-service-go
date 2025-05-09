package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupAbsensiRoutes(app *fiber.App, db *gorm.DB) {
	absensiService := services.AbsensiService{DB: db}

	absensiLimiter := limiter.New(limiter.Config{
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
		if c.Method() == fiber.MethodPost ||
			c.Method() == fiber.MethodPut ||
			c.Method() == fiber.MethodDelete {
			return absensiLimiter(c)
		}
		return c.Next()
	}

	absensiRoutes := app.Group("/api/v1/absensi", middleware.JWTMiddleware, methodLimiter)
	{
		absensiRoutes.Post("/", middleware.RoleMiddleware("mentor"), absensiService.CreateAbsensi)
		absensiRoutes.Get("/", middleware.RoleMiddleware("mentor"), absensiService.GetAbsensi)
		absensiRoutes.Get("/:id", middleware.RoleMiddleware("mentor"), absensiService.GetAbsensiByID)
		absensiRoutes.Put("/:id", middleware.RoleMiddleware("mentor"), absensiService.UpdateAbsensi)
		absensiRoutes.Get("/mahasantri/:mahasantri_id/daily-summary", middleware.RoleMiddleware("mentor", "mahasantri"), absensiService.GetAbsensiDailySummary)
		absensiRoutes.Delete("/:id", middleware.RoleMiddleware("mentor"), absensiService.DeleteAbsensi)
	}
}
