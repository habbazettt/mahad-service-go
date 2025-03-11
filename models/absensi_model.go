package models

import (
	"time"

	"gorm.io/gorm"
)

type Absensi struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	MahasantriID uint           `gorm:"not null" json:"mahasantri_id"`
	Waktu        string         `gorm:"type:varchar(10);not null" json:"waktu"` // "Shubuh" atau "Isya"
	Status       string         `gorm:"type:varchar(10);not null"`              // "Hadir", "Tidak Hadir"
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
