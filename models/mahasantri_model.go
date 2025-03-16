package models

import (
	"time"

	"gorm.io/gorm"
)

type Mahasantri struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Nama      string         `gorm:"type:varchar(255);not null" json:"nama"`
	NIM       string         `gorm:"type:varchar(50);not null;unique" json:"nim"`
	Jurusan   string         `gorm:"type:varchar(100);not null" json:"jurusan"`
	Password  string         `gorm:"not null" json:"-"`
	Gender    string         `gorm:"type:varchar(10);not null" json:"gender"`
	MentorID  uint           `gorm:"not null" json:"mentor_id"`         // Foreign Key ke Mentor
	Mentor    Mentor         `gorm:"foreignKey:MentorID" json:"mentor"` // Relasi ke Mentor
	Hafalan   []Hafalan      `gorm:"foreignKey:MahasantriID;constraint:OnDelete:CASCADE;" json:"hafalan,omitempty"`
	Absensi   []Absensi      `gorm:"foreignKey:MahasantriID" json:"absensi"` // Relasi ke Absensi (untuk Mahasantri)
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
