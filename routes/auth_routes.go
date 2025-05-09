package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupAuthRoutes(app *fiber.App, db *gorm.DB) {
	services := services.AuthService{DB: db}

	authLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests, please try again later",
			})
		},
	})

	auth := app.Group("/api/v1/auth")

	auth.Use(authLimiter)

	{
		auth.Post("/register/mahasantri", services.RegisterMahasantri)
		auth.Post("/register/mentor", services.RegisterMentor)
		auth.Post("/login/mahasantri", services.LoginMahasantri)
		auth.Post("/login/mentor", services.LoginMentor)
		auth.Post("/forget-password", services.ForgotPassword)
		auth.Post("/logout", middleware.JWTMiddleware, services.Logout)
		auth.Get("/me", middleware.JWTMiddleware, services.GetCurrentUser)
	}
}
