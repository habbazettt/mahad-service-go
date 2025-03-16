package models

import (
	"time"

	"gorm.io/gorm"
)

type Hafalan struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	MahasantriID uint           `gorm:"not null" json:"mahasantri_id"`
	Mahasantri   Mahasantri     `gorm:"foreignKey:MahasantriID" json:"-"` // Relasi ke Mahasantri
	Juz          int            `gorm:"not null" json:"juz"`
	Halaman      string         `gorm:"type:varchar(20);not null" json:"halaman"`
	TotalSetoran float32        `gorm:"not null" json:"total_setoran"`
	Kategori     string         `gorm:"type:varchar(20);not null" json:"kategori" validate:"oneof=Ziyadah Murojaah"`
	Waktu        string         `gorm:"type:varchar(10);not null" json:"waktu" validate:"oneof=Shubuh Isya"`
	Catatan      string         `gorm:"type:varchar(255)" json:"catatan,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
