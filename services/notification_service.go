package services

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/dto"
	"github.com/habbazettt/mahad-service-go/models"
	"github.com/habbazettt/mahad-service-go/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type NotificationService struct {
	DB *gorm.DB
}

func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{DB: db}
}

// Helper: Format time ago
func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff.Minutes() < 1 {
		return "Baru saja"
	} else if diff.Minutes() < 60 {
		return fmt.Sprintf("%d menit yang lalu", int(diff.Minutes()))
	} else if diff.Hours() < 24 {
		return fmt.Sprintf("%d jam yang lalu", int(diff.Hours()))
	} else if diff.Hours() < 48 {
		return "Kemarin"
	} else if diff.Hours() < 168 { // 7 days
		return fmt.Sprintf("%d hari yang lalu", int(diff.Hours()/24))
	} else {
		return t.Format("02 Jan 2006")
	}
}

// Helper: Build notification response
func buildNotificationResponse(notif models.Notification) dto.NotificationResponse {
	typeLabel := string(notif.Type)
	if info := models.GetNotificationTypeInfo(notif.Type); info != nil {
		typeLabel = info.Label
	}

	return dto.NotificationResponse{
		ID:          notif.ID,
		Type:        string(notif.Type),
		TypeLabel:   typeLabel,
		Title:       notif.Title,
		Message:     notif.Message,
		IsRead:      notif.IsRead,
		RelatedID:   notif.RelatedID,
		RelatedType: notif.RelatedType,
		CreatedAt:   notif.CreatedAt,
		TimeAgo:     formatTimeAgo(notif.CreatedAt),
	}
}

// GetNotificationTypes - Get all notification types
func (s *NotificationService) GetNotificationTypes(c *fiber.Ctx) error {
	var responses []dto.NotificationTypeResponse
	for _, info := range models.NotificationTypes {
		responses = append(responses, dto.NotificationTypeResponse{
			Type:        string(info.Type),
			Label:       info.Label,
			Description: info.Description,
			ForUserType: info.ForUserType,
		})
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Tipe notifikasi berhasil diambil", responses)
}

// GetMyNotifications - Get all notifications for current user
func (s *NotificationService) GetMyNotifications(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	userID := claims.ID
	userType := claims.Role

	log := logrus.WithFields(logrus.Fields{"handler": "GetMyNotifications", "userID": userID, "userType": userType})

	var notifications []models.Notification
	if err := s.DB.Where("user_id = ? AND user_type = ?", userID, userType).
		Order("created_at DESC").
		Limit(100).
		Find(&notifications).Error; err != nil {
		log.WithError(err).Error("Failed to fetch notifications")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal mengambil notifikasi", err.Error())
	}

	var unreadCount, totalCount int64
	s.DB.Model(&models.Notification{}).Where("user_id = ? AND user_type = ? AND is_read = false", userID, userType).Count(&unreadCount)
	s.DB.Model(&models.Notification{}).Where("user_id = ? AND user_type = ?", userID, userType).Count(&totalCount)

	responses := make([]dto.NotificationResponse, len(notifications))
	for i, notif := range notifications {
		responses[i] = buildNotificationResponse(notif)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Notifikasi berhasil diambil", dto.NotificationListResponse{
		Notifications: responses,
		UnreadCount:   unreadCount,
		TotalCount:    totalCount,
	})
}

// GetUnreadCount - Get unread notification count
func (s *NotificationService) GetUnreadCount(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	userID := claims.ID
	userType := claims.Role

	var unreadCount, totalCount int64
	s.DB.Model(&models.Notification{}).Where("user_id = ? AND user_type = ? AND is_read = false", userID, userType).Count(&unreadCount)
	s.DB.Model(&models.Notification{}).Where("user_id = ? AND user_type = ?", userID, userType).Count(&totalCount)

	return utils.SuccessResponse(c, fiber.StatusOK, "Jumlah notifikasi berhasil diambil", dto.NotificationCountResponse{
		UnreadCount: unreadCount,
		TotalCount:  totalCount,
	})
}

// MarkAsRead - Mark single notification as read
func (s *NotificationService) MarkAsRead(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	userID := claims.ID
	userType := claims.Role
	notifID := c.Params("id")

	log := logrus.WithFields(logrus.Fields{"handler": "MarkAsRead", "userID": userID, "notifID": notifID})

	result := s.DB.Model(&models.Notification{}).
		Where("id = ? AND user_id = ? AND user_type = ?", notifID, userID, userType).
		Update("is_read", true)

	if result.Error != nil {
		log.WithError(result.Error).Error("Failed to mark notification as read")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal menandai notifikasi", result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return utils.ResponseError(c, fiber.StatusNotFound, "Notifikasi tidak ditemukan", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Notifikasi ditandai sudah dibaca", nil)
}

// MarkAllAsRead - Mark all notifications as read
func (s *NotificationService) MarkAllAsRead(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	userID := claims.ID
	userType := claims.Role

	log := logrus.WithFields(logrus.Fields{"handler": "MarkAllAsRead", "userID": userID})

	result := s.DB.Model(&models.Notification{}).
		Where("user_id = ? AND user_type = ? AND is_read = false", userID, userType).
		Update("is_read", true)

	if result.Error != nil {
		log.WithError(result.Error).Error("Failed to mark all notifications as read")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal menandai semua notifikasi", result.Error.Error())
	}

	log.WithField("count", result.RowsAffected).Info("Marked all notifications as read")
	return utils.SuccessResponse(c, fiber.StatusOK, fmt.Sprintf("%d notifikasi ditandai sudah dibaca", result.RowsAffected), nil)
}

// DeleteNotification - Delete a notification
func (s *NotificationService) DeleteNotification(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	userID := claims.ID
	userType := claims.Role
	notifID := c.Params("id")

	log := logrus.WithFields(logrus.Fields{"handler": "DeleteNotification", "userID": userID, "notifID": notifID})

	result := s.DB.Where("id = ? AND user_id = ? AND user_type = ?", notifID, userID, userType).
		Delete(&models.Notification{})

	if result.Error != nil {
		log.WithError(result.Error).Error("Failed to delete notification")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal menghapus notifikasi", result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return utils.ResponseError(c, fiber.StatusNotFound, "Notifikasi tidak ditemukan", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Notifikasi berhasil dihapus", nil)
}

// MentorSendNotification - Mentor sends notification to mahasantri
func (s *NotificationService) MentorSendNotification(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	mentorID := claims.ID

	log := logrus.WithFields(logrus.Fields{"handler": "MentorSendNotification", "mentorID": mentorID})

	var req dto.MentorSendNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Request body tidak valid", err.Error())
	}

	// Verify mahasantri belongs to this mentor
	var mahasantri models.Mahasantri
	if err := s.DB.Where("id = ? AND mentor_id = ?", req.MahasantriID, mentorID).First(&mahasantri).Error; err != nil {
		return utils.ResponseError(c, fiber.StatusForbidden, "Mahasantri bukan bimbingan Anda", nil)
	}

	// Create notification
	notification := models.Notification{
		UserID:   req.MahasantriID,
		UserType: "mahasantri",
		Type:     models.NotifPesanMentor,
		Title:    req.Title,
		Message:  req.Message,
	}

	if err := s.DB.Create(&notification).Error; err != nil {
		log.WithError(err).Error("Failed to create notification")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal mengirim notifikasi", err.Error())
	}

	log.WithField("mahasantriID", req.MahasantriID).Info("Mentor sent notification to mahasantri")
	return utils.SuccessResponse(c, fiber.StatusCreated, "Notifikasi berhasil dikirim", buildNotificationResponse(notification))
}

// AdminSendNotification - Admin sends notification to specific user
func (s *NotificationService) AdminSendNotification(c *fiber.Ctx) error {
	log := logrus.WithField("handler", "AdminSendNotification")

	var req dto.SendNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Request body tidak valid", err.Error())
	}

	// Verify recipient exists
	switch req.RecipientType {
	case "mahasantri":
		var m models.Mahasantri
		if err := s.DB.First(&m, req.RecipientID).Error; err != nil {
			return utils.ResponseError(c, fiber.StatusNotFound, "Mahasantri tidak ditemukan", nil)
		}
	case "mentor":
		var m models.Mentor
		if err := s.DB.First(&m, req.RecipientID).Error; err != nil {
			return utils.ResponseError(c, fiber.StatusNotFound, "Mentor tidak ditemukan", nil)
		}
	}

	notification := models.Notification{
		UserID:   req.RecipientID,
		UserType: req.RecipientType,
		Type:     models.NotificationType(req.Type),
		Title:    req.Title,
		Message:  req.Message,
	}

	if err := s.DB.Create(&notification).Error; err != nil {
		log.WithError(err).Error("Failed to create notification")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal mengirim notifikasi", err.Error())
	}

	log.WithFields(logrus.Fields{
		"recipientID":   req.RecipientID,
		"recipientType": req.RecipientType,
	}).Info("Admin sent notification")

	return utils.SuccessResponse(c, fiber.StatusCreated, "Notifikasi berhasil dikirim", buildNotificationResponse(notification))
}

// AdminBroadcastNotification - Admin broadcasts notification to all users of a type
func (s *NotificationService) AdminBroadcastNotification(c *fiber.Ctx) error {
	log := logrus.WithField("handler", "AdminBroadcastNotification")

	var req dto.BroadcastNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Request body tidak valid", err.Error())
	}

	var notifications []models.Notification

	// Get all target users
	if req.TargetType == "mahasantri" || req.TargetType == "all" {
		var mahasantris []models.Mahasantri
		s.DB.Select("id").Find(&mahasantris)
		for _, m := range mahasantris {
			notifications = append(notifications, models.Notification{
				UserID:   m.ID,
				UserType: "mahasantri",
				Type:     models.NotificationType(req.Type),
				Title:    req.Title,
				Message:  req.Message,
			})
		}
	}

	if req.TargetType == "mentor" || req.TargetType == "all" {
		var mentors []models.Mentor
		s.DB.Select("id").Find(&mentors)
		for _, m := range mentors {
			notifications = append(notifications, models.Notification{
				UserID:   m.ID,
				UserType: "mentor",
				Type:     models.NotificationType(req.Type),
				Title:    req.Title,
				Message:  req.Message,
			})
		}
	}

	if len(notifications) == 0 {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Tidak ada penerima yang valid", nil)
	}

	// Batch insert
	if err := s.DB.CreateInBatches(notifications, 100).Error; err != nil {
		log.WithError(err).Error("Failed to broadcast notifications")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal broadcast notifikasi", err.Error())
	}

	log.WithFields(logrus.Fields{
		"targetType": req.TargetType,
		"count":      len(notifications),
	}).Info("Admin broadcasted notification")

	return utils.SuccessResponse(c, fiber.StatusCreated, fmt.Sprintf("Notifikasi berhasil dikirim ke %d penerima", len(notifications)), nil)
}

// AdminGetAllNotifications - Admin gets all notifications in system
func (s *NotificationService) AdminGetAllNotifications(c *fiber.Ctx) error {
	log := logrus.WithField("handler", "AdminGetAllNotifications")

	var notifications []models.Notification
	if err := s.DB.Order("created_at DESC").Limit(500).Find(&notifications).Error; err != nil {
		log.WithError(err).Error("Failed to fetch all notifications")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal mengambil notifikasi", err.Error())
	}

	responses := make([]dto.NotificationResponse, len(notifications))
	for i, notif := range notifications {
		responses[i] = buildNotificationResponse(notif)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Semua notifikasi berhasil diambil", responses)
}

// ==================== HELPER FUNCTIONS FOR CREATING SYSTEM NOTIFICATIONS ====================

// CreateNotification - Helper to create a notification
func (s *NotificationService) CreateNotification(userID uint, userType string, notifType models.NotificationType, title, message string, relatedID *uint, relatedType string) error {
	notification := models.Notification{
		UserID:      userID,
		UserType:    userType,
		Type:        notifType,
		Title:       title,
		Message:     message,
		RelatedID:   relatedID,
		RelatedType: relatedType,
	}
	return s.DB.Create(&notification).Error
}

// CreateNotificationForMahasantri - Quick helper for mahasantri notifications
func (s *NotificationService) CreateNotificationForMahasantri(mahasantriID uint, notifType models.NotificationType, title, message string) error {
	return s.CreateNotification(mahasantriID, "mahasantri", notifType, title, message, nil, "")
}

// CreateNotificationForMentor - Quick helper for mentor notifications
func (s *NotificationService) CreateNotificationForMentor(mentorID uint, notifType models.NotificationType, title, message string) error {
	return s.CreateNotification(mentorID, "mentor", notifType, title, message, nil, "")
}
