package dto

import "time"

type AddDetailLogRequest struct {
	WaktuMurojaah      string `json:"waktu_murojaah" validate:"required"`
	TargetStartJuz     int    `json:"target_start_juz" validate:"required,min=1,max=30"`
	TargetStartHalaman int    `json:"target_start_halaman" validate:"required,min=1,max=20"`
	TargetEndJuz       int    `json:"target_end_juz" validate:"required,min=1,max=30"`
	TargetEndHalaman   int    `json:"target_end_halaman" validate:"required,min=1,max=20"`
	Catatan            string `json:"catatan"`
}

type UpdateDetailLogRequest struct {
	SelesaiEndJuz     int    `json:"selesai_end_juz" validate:"required,min=1,max=30"`
	SelesaiEndHalaman int    `json:"selesai_end_halaman" validate:"required,min=1,max=20"`
	Catatan           string `json:"catatan"`
}

type DetailLogResponse struct {
	ID                  uint      `json:"id"`
	WaktuMurojaah       string    `json:"waktu_murojaah"`
	TargetStartJuz      int       `json:"target_start_juz"`
	TargetStartHalaman  int       `json:"target_start_halaman"`
	TargetEndJuz        int       `json:"target_end_juz"`
	TargetEndHalaman    int       `json:"target_end_halaman"`
	TotalTargetHalaman  int       `json:"total_target_halaman"`
	SelesaiEndJuz       int       `json:"selesai_end_juz"`
	SelesaiEndHalaman   int       `json:"selesai_end_halaman"`
	TotalSelesaiHalaman int       `json:"total_selesai_halaman"`
	Status              string    `json:"status"`
	Catatan             string    `json:"catatan"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type LogHarianResponse struct {
	ID                  uint                `json:"id"`
	Tanggal             string              `json:"tanggal"`
	TotalTargetHalaman  int                 `json:"total_target_halaman"`
	TotalSelesaiHalaman int                 `json:"total_selesai_halaman"`
	DetailLogs          []DetailLogResponse `json:"detail_logs"`
}

type MahasantriInfoForLog struct {
	ID   uint   `json:"id"`
	Nama string `json:"nama"`
}

type LogHarianForMentorResponse struct {
	LogID               uint                 `json:"log_id"`
	Tanggal             string               `json:"tanggal"`
	TotalTargetHalaman  int                  `json:"total_target_halaman"`
	TotalSelesaiHalaman int                  `json:"total_selesai_halaman"`
	Mahasantri          MahasantriInfoForLog `json:"mahasantri"`
	DetailLogs          []DetailLogResponse  `json:"detail_logs"`
}

type ApplyAIRekomendasiRequest struct {
	RekomendasiID      uint   `json:"rekomendasi_id" validate:"required"`
	TargetStartJuz     int    `json:"target_start_juz" validate:"required,min=1,max=30"`
	TargetStartHalaman int    `json:"target_start_halaman" validate:"required,min=1,max=20"`
	TargetEndJuz       int    `json:"target_end_juz" validate:"required,min=1,max=30"`
	TargetEndHalaman   int    `json:"target_end_halaman" validate:"required,min=1,max=20"`
	Catatan            string `json:"catatan"`
}
