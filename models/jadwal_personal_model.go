package models

import (
	"time"
)

type JadwalPersonal struct {
	ID                uint        `gorm:"primaryKey" json:"id"`
	MahasantriID      *uint       `gorm:"uniqueIndex" json:"mahasantri_id"`
	MentorID          *uint       `gorm:"uniqueIndex" json:"mentor_id"`
	TotalHafalan      int         `gorm:"not null" json:"total_hafalan"`
	Kesibukan         string      `gorm:"type:varchar(255);not null" json:"kesibukan"`
	Jadwal            string      `gorm:"type:varchar(255)" json:"jadwal"`
	EfektifitasJadwal int         `gorm:"not null" json:"efektifitas_jadwal"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	Mahasantri        *Mahasantri `gorm:"foreignKey:MahasantriID;constraint:OnDelete:CASCADE;" json:"-"`
	Mentor            *Mentor     `gorm:"foreignKey:MentorID;constraint:OnDelete:CASCADE;" json:"-"`
}
