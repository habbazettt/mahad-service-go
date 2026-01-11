package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupAdminRoutes(app *fiber.App, db *gorm.DB) {
	adminService := services.AdminService{DB: db}

	// Auth routes untuk admin login
	auth := app.Group("/api/v1/auth")
	auth.Post("/login/admin", adminService.LoginAdmin)

	// Admin routes (protected - admin only)
	admin := app.Group("/api/v1/admin")
	admin.Use(middleware.JWTMiddleware)
	admin.Use(middleware.RoleMiddleware("admin"))

	{
		// Admin management
		admin.Post("/register", adminService.RegisterAdmin)
		admin.Get("/dashboard", adminService.GetDashboardStats)

		// Mentor management
		admin.Get("/mentors", adminService.GetAllMentors)
		admin.Post("/mentors", adminService.CreateMentor)
		admin.Put("/mentors/:id", adminService.UpdateMentor)
		admin.Delete("/mentors/:id", adminService.DeleteMentor)

		// Mahasantri management
		admin.Get("/mahasantri", adminService.GetAllMahasantri)
		admin.Post("/mahasantri", adminService.CreateMahasantri)
		admin.Put("/mahasantri/:id", adminService.UpdateMahasantri)
		admin.Delete("/mahasantri/:id", adminService.DeleteMahasantri)
	}
}
