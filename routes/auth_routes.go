package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupAuthRoutes(app *fiber.App, db *gorm.DB) {
	services := services.AuthService{DB: db}

	auth := app.Group("/api/v1/auth")
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
