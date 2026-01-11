package dto

import "time"

// ==================== REQUEST DTOs ====================

// AddDetailLogRequest - Request untuk membuat sesi murojaah baru
type AddDetailLogRequest struct {
	WaktuMurojaah      string `json:"waktu_murojaah" validate:"required"`
	TargetStartJuz     int    `json:"target_start_juz" validate:"required,min=1,max=30"`
	TargetStartHalaman int    `json:"target_start_halaman" validate:"required,min=1,max=20"`
	TargetEndJuz       int    `json:"target_end_juz" validate:"required,min=1,max=30"`
	TargetEndHalaman   int    `json:"target_end_halaman" validate:"required,min=1,max=20"`
	Catatan            string `json:"catatan"`
}

// AddProgressRequest - Request untuk menambah progress ke DetailLog (BARU)
type AddProgressRequest struct {
	JumlahHalaman int    `json:"jumlah_halaman" validate:"required,min=1"`
	Catatan       string `json:"catatan"`
}

// ApplyAIRekomendasiRequest - Request untuk apply rekomendasi AI
type ApplyAIRekomendasiRequest struct {
	RekomendasiID      uint   `json:"rekomendasi_id" validate:"required"`
	TargetStartJuz     int    `json:"target_start_juz" validate:"required,min=1,max=30"`
	TargetStartHalaman int    `json:"target_start_halaman" validate:"required,min=1,max=20"`
	TargetEndJuz       int    `json:"target_end_juz" validate:"required,min=1,max=30"`
	TargetEndHalaman   int    `json:"target_end_halaman" validate:"required,min=1,max=20"`
	Catatan            string `json:"catatan"`
}

// ==================== RESPONSE DTOs ====================

// ProgressLogResponse - Response untuk entry progress
type ProgressLogResponse struct {
	ID            uint      `json:"id"`
	JumlahHalaman int       `json:"jumlah_halaman"`
	Catatan       string    `json:"catatan"`
	CreatedAt     time.Time `json:"created_at"`
}

// DetailLogResponse - Response untuk sesi murojaah
type DetailLogResponse struct {
	ID                  uint                  `json:"id"`
	WaktuMurojaah       string                `json:"waktu_murojaah"`       // Kode waktu
	WaktuMurojaahLabel  string                `json:"waktu_murojaah_label"` // Label lengkap dengan jam
	TargetStartJuz      int                   `json:"target_start_juz"`
	TargetStartHalaman  int                   `json:"target_start_halaman"`
	TargetEndJuz        int                   `json:"target_end_juz"`
	TargetEndHalaman    int                   `json:"target_end_halaman"`
	TotalTargetHalaman  int                   `json:"total_target_halaman"`
	TotalSelesaiHalaman int                   `json:"total_selesai_halaman"`
	PersentaseSelesai   float64               `json:"persentase_selesai"`
	Status              string                `json:"status"`
	Catatan             string                `json:"catatan"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
	ProgressLogs        []ProgressLogResponse `json:"progress_logs"`
}

// LogHarianResponse - Response untuk log harian
type LogHarianResponse struct {
	ID                  uint                `json:"id"`
	Tanggal             string              `json:"tanggal"`
	TotalTargetHalaman  int                 `json:"total_target_halaman"`
	TotalSelesaiHalaman int                 `json:"total_selesai_halaman"`
	PersentaseSelesai   float64             `json:"persentase_selesai"`
	DetailLogs          []DetailLogResponse `json:"detail_logs"`
}

// WaktuMurojaahOptionResponse - Response untuk opsi waktu murojaah
type WaktuMurojaahOptionResponse struct {
	Kode      string `json:"kode"`
	Label     string `json:"label"`
	JamMulai  string `json:"jam_mulai"`
	JamAkhir  string `json:"jam_akhir"`
	FullLabel string `json:"full_label"`
}

// ==================== MENTOR DASHBOARD DTOs ====================

// MahasantriInfoForLog - Info mahasantri untuk log
type MahasantriInfoForLog struct {
	ID   uint   `json:"id"`
	Nama string `json:"nama"`
}

// LogHarianForMentorResponse - Response log harian untuk mentor
type LogHarianForMentorResponse struct {
	LogID               uint                 `json:"log_id"`
	Tanggal             string               `json:"tanggal"`
	TotalTargetHalaman  int                  `json:"total_target_halaman"`
	TotalSelesaiHalaman int                  `json:"total_selesai_halaman"`
	PersentaseSelesai   float64              `json:"persentase_selesai"`
	Mahasantri          MahasantriInfoForLog `json:"mahasantri"`
	DetailLogs          []DetailLogResponse  `json:"detail_logs"`
}
