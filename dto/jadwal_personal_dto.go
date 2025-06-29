package dto

import "time"

type CreateJadwalPersonalRequest struct {
	TotalHafalan      int    `json:"total_hafalan" validate:"required,min=1,max=30"`
	Jadwal            string `json:"jadwal" validate:"required"`
	Kesibukan         string `json:"kesibukan" validate:"required"`
	EfektifitasJadwal int    `json:"efektifitas_jadwal" validate:"required,min=1,max=5"`
}

type UpdateJadwalPersonalRequest struct {
	TotalHafalan      *int    `json:"total_hafalan" validate:"required,min=1,max=30"`
	Jadwal            *string `json:"jadwal" validate:"required"`
	Kesibukan         *string `json:"kesibukan" validate:"required"`
	EfektifitasJadwal *int    `json:"efektifitas_jadwal" validate:"required,min=1,max=5"`
}

type JadwalPersonalResponse struct {
	ID                uint   `json:"id"`
	MahasantriID      *uint  `json:"mahasantri_id,omitempty"`
	MentorID          *uint  `json:"mentor_id,omitempty"`
	TotalHafalan      int    `json:"total_hafalan"`
	Jadwal            string `json:"jadwal"`
	Kesibukan         string `json:"kesibukan"`
	EfektifitasJadwal int    `json:"efektifitas_jadwal"`
}

type JadwalPersonalDetailResponse struct {
	ID                uint      `json:"id"`
	OwnerName         string    `json:"owner_name"` // Nama pemilik jadwal (mentor atau mahasantri)
	OwnerRole         string    `json:"owner_role"` // Peran pemilik jadwal
	TotalHafalan      int       `json:"total_hafalan"`
	Jadwal            string    `json:"jadwal"`
	Kesibukan         string    `json:"kesibukan"`
	EfektifitasJadwal int       `json:"efektifitas_jadwal"`
	UpdatedAt         time.Time `json:"updated_at"`
}
