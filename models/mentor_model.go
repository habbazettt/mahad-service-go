package models

import (
	"time"
)

type Mentor struct {
	ID                uint                `gorm:"primaryKey" json:"id"`
	Nama              string              `gorm:"type:varchar(255);not null" json:"nama"`
	Email             string              `gorm:"type:varchar(100);not null;unique" json:"email"`
	Password          string              `gorm:"not null" json:"-"`
	Gender            string              `gorm:"type:varchar(10);not null" json:"gender"`
	Mahasantri        []Mahasantri        `gorm:"foreignKey:MentorID;constraint:OnDelete:CASCADE;" json:"mahasantri"`
	JadwalRekomendasi []JadwalRekomendasi `gorm:"foreignKey:MentorID;constraint:OnDelete:CASCADE;" json:"jadwal_rekomendasi"`
	Absensi           []Absensi           `gorm:"foreignKey:MentorID;constraint:OnDelete:CASCADE;" json:"absensi"`
	Hafalan           []Hafalan           `gorm:"foreignKey:MentorID;constraint:OnDelete:CASCADE;" json:"hafalan"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
}
