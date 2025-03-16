package models

import (
	"time"

	"gorm.io/gorm"
)

type Mentor struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Nama       string         `gorm:"type:varchar(255);not null" json:"nama"`
	Email      string         `gorm:"type:varchar(100);not null;unique" json:"email"`
	Password   string         `gorm:"not null" json:"-"`
	Gender     string         `gorm:"type:varchar(10);not null" json:"gender"`
	Mahasantri []Mahasantri   `gorm:"foreignKey:MentorID" json:"mahasantri"` // Relasi ke Mahasantri
	Absensi    []Absensi      `gorm:"foreignKey:MentorID" json:"absensi"`    // Relasi ke Absensi (yang diinputkan oleh Mentor)
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
