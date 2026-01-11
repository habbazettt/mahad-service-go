package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupNotificationRoutes(app *fiber.App, db *gorm.DB) {
	service := services.NewNotificationService(db)

	// Public route untuk get notification types
	app.Get("/api/v1/notification-types", service.GetNotificationTypes)

	// Routes untuk semua user yang sudah login (mahasantri & mentor)
	notifRoutes := app.Group("/api/v1/notifications", middleware.JWTMiddleware)
	{
		notifRoutes.Get("/", service.GetMyNotifications)
		notifRoutes.Get("/count", service.GetUnreadCount)
		notifRoutes.Put("/:id/read", service.MarkAsRead)
		notifRoutes.Put("/read-all", service.MarkAllAsRead)
		notifRoutes.Delete("/:id", service.DeleteNotification)
	}

	// Routes untuk Mentor - kirim notifikasi ke mahasantri
	mentorNotifRoutes := app.Group("/api/v1/mentor/notifications", middleware.JWTMiddleware, middleware.RoleMiddleware("mentor"))
	{
		mentorNotifRoutes.Post("/send", service.MentorSendNotification)
	}

	// Routes untuk Admin - kirim notifikasi ke siapa saja
	adminNotifRoutes := app.Group("/api/v1/admin/notifications", middleware.JWTMiddleware, middleware.RoleMiddleware("admin"))
	{
		adminNotifRoutes.Get("/", service.AdminGetAllNotifications)
		adminNotifRoutes.Post("/send", service.AdminSendNotification)
		adminNotifRoutes.Post("/broadcast", service.AdminBroadcastNotification)
	}
}
