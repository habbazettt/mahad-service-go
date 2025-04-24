package models

import (
	"time"
)

type Absensi struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	MahasantriID uint      `gorm:"not null" json:"mahasantri_id"`
	MentorID     uint      `gorm:"not null" json:"mentor_id"`
	Waktu        string    `gorm:"type:varchar(10);not null" json:"waktu"`
	Status       string    `gorm:"type:varchar(10);not null" json:"status"`
	Tanggal      time.Time `gorm:"type:date;not null" json:"tanggal"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Mentor     Mentor     `gorm:"foreignKey:MentorID;constraint:OnDelete:CASCADE;" json:"mentor"`
	Mahasantri Mahasantri `gorm:"foreignKey:MahasantriID;constraint:OnDelete:CASCADE;" json:"mahasantri"`
}

func (a *Absensi) GetFormattedTanggal() string {
	return a.Tanggal.Format("02-01-2006")
}
