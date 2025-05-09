package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupTargetSemesterRoutes(app *fiber.App, db *gorm.DB) {
	service := services.TargetSemesterService{DB: db}

	targetLimiter := limiter.New(limiter.Config{
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
			return targetLimiter(c)
		}
		return c.Next()
	}

	targetSemesterRoutes := app.Group("/api/v1/target_semester", middleware.JWTMiddleware, methodLimiter)
	{
		targetSemesterRoutes.Post("/", middleware.RoleMiddleware("mentor"), service.CreateTargetSemester)
		targetSemesterRoutes.Get("/", middleware.RoleMiddleware("mentor"), service.GetAllTargetSemesters)
		targetSemesterRoutes.Get("/:id", middleware.RoleMiddleware("mentor"), service.GetTargetSemesterByID)
		targetSemesterRoutes.Get("/mahasantri/:mahasantri_id", middleware.RoleMiddleware("mentor", "mahasantri"), service.GetTargetSemesterByMahasantriID)
		targetSemesterRoutes.Put("/:id", middleware.RoleMiddleware("mentor"), service.UpdateTargetSemester)
		targetSemesterRoutes.Delete("/:id", middleware.RoleMiddleware("mentor"), service.DeleteTargetSemester)
	}
}
