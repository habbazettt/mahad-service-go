package dto

import "time"

// ==================== REQUEST DTOs ====================

// SendNotificationRequest - Admin mengirim notifikasi ke siapa saja
type SendNotificationRequest struct {
	RecipientID   uint   `json:"recipient_id" validate:"required"`
	RecipientType string `json:"recipient_type" validate:"required,oneof=mahasantri mentor"`
	Type          string `json:"type" validate:"required"`
	Title         string `json:"title" validate:"required,max=255"`
	Message       string `json:"message" validate:"required"`
}

// MentorSendNotificationRequest - Mentor mengirim ke mahasantri bimbingan
type MentorSendNotificationRequest struct {
	MahasantriID uint   `json:"mahasantri_id" validate:"required"`
	Title        string `json:"title" validate:"required,max=255"`
	Message      string `json:"message" validate:"required"`
}

// BroadcastNotificationRequest - Admin broadcast ke semua user
type BroadcastNotificationRequest struct {
	TargetType string `json:"target_type" validate:"required,oneof=mahasantri mentor all"` // Target: mahasantri, mentor, atau all
	Type       string `json:"type" validate:"required"`
	Title      string `json:"title" validate:"required,max=255"`
	Message    string `json:"message" validate:"required"`
}

// ==================== RESPONSE DTOs ====================

// NotificationResponse - Response untuk satu notifikasi
type NotificationResponse struct {
	ID          uint      `json:"id"`
	Type        string    `json:"type"`
	TypeLabel   string    `json:"type_label"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	IsRead      bool      `json:"is_read"`
	RelatedID   *uint     `json:"related_id,omitempty"`
	RelatedType string    `json:"related_type,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	TimeAgo     string    `json:"time_ago"` // "5 menit yang lalu", "2 jam yang lalu", dll
}

// NotificationListResponse - Response untuk daftar notifikasi
type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	UnreadCount   int64                  `json:"unread_count"`
	TotalCount    int64                  `json:"total_count"`
}

// NotificationCountResponse - Response untuk jumlah notifikasi
type NotificationCountResponse struct {
	UnreadCount int64 `json:"unread_count"`
	TotalCount  int64 `json:"total_count"`
}

// NotificationTypeResponse - Response untuk tipe notifikasi
type NotificationTypeResponse struct {
	Type        string `json:"type"`
	Label       string `json:"label"`
	Description string `json:"description"`
	ForUserType string `json:"for_user_type"`
}
