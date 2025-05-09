package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupHafalanRoutes(app *fiber.App, db *gorm.DB) {
	service := services.HafalanService{DB: db}

	hafalanLimiter := limiter.New(limiter.Config{
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
		if c.Method() == fiber.MethodPost || c.Method() == fiber.MethodPut || c.Method() == fiber.MethodDelete {
			return hafalanLimiter(c)
		}
		return c.Next()
	}

	hafalanRoutes := app.Group("/api/v1/hafalan", middleware.JWTMiddleware, methodLimiter)
	{
		hafalanRoutes.Post("/", middleware.RoleMiddleware("mentor"), service.CreateHafalan)
		hafalanRoutes.Get("/", middleware.RoleMiddleware("mentor"), service.GetAllHafalan)
		hafalanRoutes.Get("/:id", middleware.RoleMiddleware("mentor"), service.GetHafalanByID)
		hafalanRoutes.Get("/mahasantri/:mahasantri_id", middleware.RoleMiddleware("mentor", "mahasantri"), service.GetHafalanByMahasantriID)
		hafalanRoutes.Get("/mentor/:mentor_id", middleware.RoleMiddleware("mentor"), service.GetHafalanByMentorID)
		hafalanRoutes.Get("/:mahasantri_id/kategori", middleware.RoleMiddleware("mentor"), service.GetHafalanByKategori)
		hafalanRoutes.Put("/:id", middleware.RoleMiddleware("mentor"), service.UpdateHafalan)
		hafalanRoutes.Delete("/:id", middleware.RoleMiddleware("mentor"), service.DeleteHafalan)
	}
}
