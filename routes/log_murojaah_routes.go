package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/middleware"
	"github.com/habbazettt/mahad-service-go/services"
	"gorm.io/gorm"
)

func SetupLogMurojaahRoutes(app *fiber.App, db *gorm.DB) {
	service := services.NewLogMurojaahService(db)

	mahasantriLogRoutes := app.Group("/api/v1/log-harian", middleware.JWTMiddleware, middleware.RoleMiddleware("mahasantri"))
	{
		mahasantriLogRoutes.Get("/", service.GetOrCreateLogHarian)
		mahasantriLogRoutes.Post("/detail", service.AddDetailToLog)
		mahasantriLogRoutes.Put("/detail/:detailID", service.UpdateDetailLog)
		mahasantriLogRoutes.Delete("/detail/:detailID", service.DeleteDetailLog)
		mahasantriLogRoutes.Get("/rekap/mingguan", service.GetRecapMingguan)
		mahasantriLogRoutes.Get("/statistik", service.GetStatistikMurojaah)
		mahasantriLogRoutes.Post("/detail/dari-rekomendasi", service.ApplyAIRekomendasi)
	}

	mentorLogRoutes := app.Group("/api/v1/mentor", middleware.JWTMiddleware, middleware.RoleMiddleware("mentor"))
	{
		mentorLogRoutes.Get("/mahasantri/:mahasantriID/log-harian", service.GetOrCreateLogHarian)
		mentorLogRoutes.Get("/log-harian-mahasantri", service.GetAllLogsForMentorDashboard)
		mentorLogRoutes.Get("/rekap-bimbingan/mingguan", service.GetRekapBimbinganMingguan)
	}
}
