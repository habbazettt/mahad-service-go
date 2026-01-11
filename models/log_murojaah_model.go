package models

import "time"

// StatusDetailLog untuk status sesi murojaah
type StatusDetailLog string

const (
	StatusSesiBelumSelesai StatusDetailLog = "Belum Selesai"
	StatusSesiBerjalan     StatusDetailLog = "Berjalan"
	StatusSesiSelesai      StatusDetailLog = "Selesai"
)

// WaktuMurojaah - Opsi waktu murojaah yang cover 24 jam
type WaktuMurojaahOption struct {
	Kode      string `json:"kode"`
	Label     string `json:"label"`
	JamMulai  string `json:"jam_mulai"`
	JamAkhir  string `json:"jam_akhir"`
	FullLabel string `json:"full_label"`
}

// Daftar opsi waktu murojaah (24 jam)
var WaktuMurojaahOptions = []WaktuMurojaahOption{
	{Kode: "QIYAMUL_LAIL", Label: "Qiyamul Lail", JamMulai: "00:00", JamAkhir: "03:00", FullLabel: "Qiyamul Lail (00:00 - 03:00)"},
	{Kode: "TAHAJJUD", Label: "Tahajjud", JamMulai: "03:00", JamAkhir: "04:30", FullLabel: "Tahajjud (03:00 - 04:30)"},
	{Kode: "SHUBUH", Label: "Ba'da Shubuh", JamMulai: "04:30", JamAkhir: "06:00", FullLabel: "Ba'da Shubuh (04:30 - 06:00)"},
	{Kode: "DHUHA_AWAL", Label: "Dhuha Awal", JamMulai: "06:00", JamAkhir: "09:00", FullLabel: "Dhuha Awal (06:00 - 09:00)"},
	{Kode: "DHUHA_AKHIR", Label: "Dhuha Akhir", JamMulai: "09:00", JamAkhir: "11:30", FullLabel: "Dhuha Akhir (09:00 - 11:30)"},
	{Kode: "DZUHUR", Label: "Ba'da Dzuhur", JamMulai: "11:30", JamAkhir: "15:00", FullLabel: "Ba'da Dzuhur (11:30 - 15:00)"},
	{Kode: "ASHAR", Label: "Ba'da Ashar", JamMulai: "15:00", JamAkhir: "17:30", FullLabel: "Ba'da Ashar (15:00 - 17:30)"},
	{Kode: "MAGHRIB", Label: "Ba'da Maghrib", JamMulai: "17:30", JamAkhir: "19:00", FullLabel: "Ba'da Maghrib (17:30 - 19:00)"},
	{Kode: "ISYA", Label: "Ba'da Isya", JamMulai: "19:00", JamAkhir: "21:00", FullLabel: "Ba'da Isya (19:00 - 21:00)"},
	{Kode: "MALAM", Label: "Malam", JamMulai: "21:00", JamAkhir: "00:00", FullLabel: "Malam (21:00 - 00:00)"},
}

// Validasi kode waktu murojaah
func IsValidWaktuMurojaah(kode string) bool {
	for _, w := range WaktuMurojaahOptions {
		if w.Kode == kode {
			return true
		}
	}
	return false
}

// Get label lengkap dari kode
func GetWaktuMurojaahLabel(kode string) string {
	for _, w := range WaktuMurojaahOptions {
		if w.Kode == kode {
			return w.FullLabel
		}
	}
	return kode
}

// LogHarian - Log harian per tanggal per mahasantri
type LogHarian struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	MahasantriID        uint      `gorm:"not null;uniqueIndex:idx_mahasantri_tanggal" json:"mahasantri_id"`
	Tanggal             time.Time `gorm:"type:date;not null;uniqueIndex:idx_mahasantri_tanggal" json:"tanggal"`
	TotalTargetHalaman  int       `gorm:"default:0" json:"total_target_halaman"`
	TotalSelesaiHalaman int       `gorm:"default:0" json:"total_selesai_halaman"`

	Mahasantri Mahasantri  `gorm:"foreignKey:MahasantriID;constraint:OnDelete:CASCADE;" json:"-"`
	DetailLogs []DetailLog `gorm:"foreignKey:LogHarianID;constraint:OnDelete:CASCADE;" json:"detail_logs"`
}

// DetailLog - Sesi murojaah (target utama)
type DetailLog struct {
	ID                  uint            `gorm:"primaryKey" json:"id"`
	LogHarianID         uint            `gorm:"not null;index" json:"log_harian_id"`
	WaktuMurojaah       string          `gorm:"type:varchar(50);not null" json:"waktu_murojaah"` // Kode waktu (SHUBUH, ISYA, dll)
	TargetStartJuz      int             `gorm:"not null" json:"target_start_juz"`
	TargetStartHalaman  int             `gorm:"not null" json:"target_start_halaman"`
	TargetEndJuz        int             `gorm:"not null" json:"target_end_juz"`
	TargetEndHalaman    int             `gorm:"not null" json:"target_end_halaman"`
	TotalTargetHalaman  int             `gorm:"default:0" json:"total_target_halaman"`
	TotalSelesaiHalaman int             `gorm:"default:0" json:"total_selesai_halaman"` // Agregat dari ProgressLogs
	Status              StatusDetailLog `gorm:"type:varchar(50);default:'Belum Selesai'" json:"status"`
	Catatan             string          `gorm:"type:text" json:"catatan"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`

	ProgressLogs []ProgressLog `gorm:"foreignKey:DetailLogID;constraint:OnDelete:CASCADE;" json:"progress_logs"`
}

// ProgressLog - Entry progress (BARU)
// Setiap kali user melaporkan progress, akan dibuat entry baru
type ProgressLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	DetailLogID   uint      `gorm:"not null;index" json:"detail_log_id"`
	JumlahHalaman int       `gorm:"not null" json:"jumlah_halaman"`
	Catatan       string    `gorm:"type:text" json:"catatan"`
	CreatedAt     time.Time `json:"created_at"`
}
