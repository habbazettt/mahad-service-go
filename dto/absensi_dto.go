package dto

import "time"

type AbsensiRequestDTO struct {
	MahasantriID uint   `json:"mahasantri_id" validate:"required"`                 // Validasi untuk memastikan MahasantriID ada
	Waktu        string `json:"waktu" validate:"required,oneof=shubuh isya"`       // "Shubuh" atau "Isya"
	Status       string `json:"status" validate:"required,oneof=hadir absen izin"` // "Hadir", "Absen", atau "Izin"
}

type AbsensiResponseDTO struct {
	ID           uint                  `json:"id"`
	MahasantriID uint                  `json:"mahasantri_id"`
	MentorID     uint                  `json:"mentor_id"`
	Waktu        string                `json:"waktu"`
	Status       string                `json:"status"`
	Tanggal      string                `json:"tanggal"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	Mentor       MentorResponseDTO     `json:"mentor"`
	Mahasantri   MahasantriResponseDTO `json:"mahasantri"`
}

type MentorResponseDTO struct {
	ID     uint   `json:"id"`
	Nama   string `json:"nama"`
	Email  string `json:"email"`
	Gender string `json:"gender"`
}

type MahasantriResponseDTO struct {
	ID      uint   `json:"id"`
	Nama    string `json:"nama"`
	NIM     string `json:"nim"`
	Jurusan string `json:"jurusan"`
	Gender  string `json:"gender"`
}

type AbsensiDailySummaryDTO struct {
	Tanggal string `json:"tanggal"` // Format: dd-mm-yyyy
	Shubuh  string `json:"subuh"`   // hadir / absen / izin / belum-absen
	Isya    string `json:"isya"`    // hadir / absen / izin / belum-absen
}

type AbsensiMonthlySummaryDTO struct {
	Month      string         `json:"month"`
	Year       int            `json:"year"`
	TotalHadir int            `json:"total_hadir"`
	TotalIzin  int            `json:"total_izin"`
	TotalAlpa  int            `json:"total_alpa"`
	Shubuh     StatusCountDTO `json:"shubuh"`
	Isya       StatusCountDTO `json:"isya"`
}

type StatusCountDTO struct {
	Hadir int `json:"hadir"`
	Izin  int `json:"izin"`
	Alpa  int `json:"alpa"`
}
