package dto

import "time"

type AbsensiRequestDTO struct {
	MahasantriID uint   `json:"mahasantri_id" validate:"required"`
	MentorID     uint   `json:"mentor_id" validate:"required"`
	Waktu        string `json:"waktu" validate:"required,oneof=shubuh isya"`       // "Shubuh" atau "Isya"
	Status       string `json:"status" validate:"required,oneof=hadir absen izin"` // "Hadir", "Absen", "Izin"
	Tanggal      string `json:"tanggal" validate:"required"`                       // Format: dd-mm-yyyy
}

type UpdateAbsensiRequestDTO struct {
	Waktu   *string `json:"waktu,omitempty"`   // "Shubuh" atau "Isya"
	Status  *string `json:"status,omitempty"`  // "Hadir", "Absen", "Izin"
	Tanggal *string `json:"tanggal,omitempty"` // Format: dd-mm-yyyy
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
	Hari    string `json:"hari"`    // Senin, Selasa, Rabu, Kamis, Jumat, Sabtu, Minggu
	Tanggal string `json:"tanggal"` // Format: dd-mm-yyyy
	Shubuh  string `json:"shubuh"`  // hadir / absen / izin / libur / belum-absen
	Isya    string `json:"isya"`    // hadir / absen / izin / libur / belum-absen
}
