package models

import (
	"time"

	"gorm.io/gorm"
)

type Absensi struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	MahasantriID uint           `gorm:"not null" json:"mahasantri_id"`           // Relasi ke Mahasantri
	MentorID     uint           `gorm:"not null" json:"mentor_id"`               // Relasi ke Mentor
	Waktu        string         `gorm:"type:varchar(10);not null" json:"waktu"`  // "Shubuh" atau "Isya"
	Status       string         `gorm:"type:varchar(10);not null" json:"status"` // "Hadir", "Tidak Hadir"
	Tanggal      time.Time      `gorm:"type:date;not null" json:"tanggal"`       // Tanggal absensi (otomatis berdasarkan hari)
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relasi ke Mentor
	Mentor Mentor `gorm:"foreignKey:MentorID" json:"mentor"`

	// Relasi ke Mahasantri
	Mahasantri Mahasantri `gorm:"foreignKey:MahasantriID" json:"mahasantri"`
}

// Custom method untuk format tanggal
func (a *Absensi) GetFormattedTanggal() string {
	return a.Tanggal.Format("02-01-2006") // dd-mm-yyyy
}
