package models

import "time"

// NotificationType - Jenis notifikasi
type NotificationType string

const (
	// Notifikasi untuk Mahasantri - Internal
	NotifReminderMurojaah NotificationType = "REMINDER_MUROJAAH"  // Pengingat waktu murojaah
	NotifSesiBelumSelesai NotificationType = "SESI_BELUM_SELESAI" // Sesi murojaah belum complete
	NotifTargetHarian     NotificationType = "TARGET_HARIAN"      // Progress target harian
	NotifTargetSemester   NotificationType = "TARGET_SEMESTER"    // Progress target semester
	NotifHafalanTercatat  NotificationType = "HAFALAN_TERCATAT"   // Hafalan baru tercatat
	NotifAbsensiTercatat  NotificationType = "ABSENSI_TERCATAT"   // Absensi tercatat
	NotifRekapMingguan    NotificationType = "REKAP_MINGGUAN"     // Rekap progress mingguan
	NotifMotivasi         NotificationType = "MOTIVASI"           // Motivasi/pencapaian
	NotifRekomendasiAI    NotificationType = "REKOMENDASI_AI"     // Rekomendasi jadwal dari AI

	// Notifikasi dari Mentor
	NotifPesanMentor NotificationType = "PESAN_MENTOR" // Pesan dari mentor

	// Notifikasi untuk Mentor
	NotifMahasantriProgress NotificationType = "MAHASANTRI_PROGRESS" // Update progress mahasantri
	NotifMahasantriBaru     NotificationType = "MAHASANTRI_BARU"     // Mahasantri baru terdaftar

	// Notifikasi Sistem
	NotifSystem     NotificationType = "SYSTEM"     // Notifikasi sistem umum
	NotifPengumuman NotificationType = "PENGUMUMAN" // Pengumuman dari admin
)

// Notification - Model notifikasi
type Notification struct {
	ID          uint             `gorm:"primaryKey" json:"id"`
	UserID      uint             `gorm:"not null;index" json:"user_id"`                    // ID penerima
	UserType    string           `gorm:"type:varchar(20);not null;index" json:"user_type"` // "mahasantri", "mentor", atau "admin"
	Type        NotificationType `gorm:"type:varchar(50);not null;index" json:"type"`
	Title       string           `gorm:"type:varchar(255);not null" json:"title"`
	Message     string           `gorm:"type:text;not null" json:"message"`
	IsRead      bool             `gorm:"default:false;index" json:"is_read"`
	RelatedID   *uint            `gorm:"index" json:"related_id,omitempty"`              // ID terkait (optional)
	RelatedType string           `gorm:"type:varchar(50)" json:"related_type,omitempty"` // Tipe relasi: "hafalan", "absensi", "log_harian", dll
	CreatedAt   time.Time        `json:"created_at"`
}

// NotificationTypeInfo - Info untuk setiap tipe notifikasi
type NotificationTypeInfo struct {
	Type        NotificationType `json:"type"`
	Label       string           `json:"label"`
	Description string           `json:"description"`
	ForUserType string           `json:"for_user_type"` // "mahasantri", "mentor", "all"
}

// Daftar semua tipe notifikasi dengan info
var NotificationTypes = []NotificationTypeInfo{
	// Mahasantri Internal
	{Type: NotifReminderMurojaah, Label: "Pengingat Murojaah", Description: "Pengingat waktu murojaah berdasarkan jadwal", ForUserType: "mahasantri"},
	{Type: NotifSesiBelumSelesai, Label: "Sesi Belum Selesai", Description: "Reminder sesi murojaah yang belum complete", ForUserType: "mahasantri"},
	{Type: NotifTargetHarian, Label: "Target Harian", Description: "Notifikasi progress target harian", ForUserType: "mahasantri"},
	{Type: NotifTargetSemester, Label: "Target Semester", Description: "Notifikasi progress target semester", ForUserType: "mahasantri"},
	{Type: NotifHafalanTercatat, Label: "Hafalan Tercatat", Description: "Konfirmasi hafalan baru tercatat", ForUserType: "mahasantri"},
	{Type: NotifAbsensiTercatat, Label: "Absensi Tercatat", Description: "Konfirmasi absensi tercatat", ForUserType: "mahasantri"},
	{Type: NotifRekapMingguan, Label: "Rekap Mingguan", Description: "Ringkasan progress mingguan", ForUserType: "mahasantri"},
	{Type: NotifMotivasi, Label: "Motivasi", Description: "Pesan motivasi dan pencapaian", ForUserType: "mahasantri"},
	{Type: NotifRekomendasiAI, Label: "Rekomendasi AI", Description: "Rekomendasi jadwal murojaah dari AI", ForUserType: "mahasantri"},

	// Dari Mentor
	{Type: NotifPesanMentor, Label: "Pesan Mentor", Description: "Pesan personal dari mentor", ForUserType: "mahasantri"},

	// Untuk Mentor
	{Type: NotifMahasantriProgress, Label: "Progress Mahasantri", Description: "Update progress mahasantri bimbingan", ForUserType: "mentor"},
	{Type: NotifMahasantriBaru, Label: "Mahasantri Baru", Description: "Mahasantri baru terdaftar di bimbingan", ForUserType: "mentor"},

	// Sistem
	{Type: NotifSystem, Label: "Sistem", Description: "Notifikasi sistem", ForUserType: "all"},
	{Type: NotifPengumuman, Label: "Pengumuman", Description: "Pengumuman dari admin", ForUserType: "all"},
}

// GetNotificationTypeInfo - Get info untuk tipe notifikasi
func GetNotificationTypeInfo(t NotificationType) *NotificationTypeInfo {
	for _, info := range NotificationTypes {
		if info.Type == t {
			return &info
		}
	}
	return nil
}
