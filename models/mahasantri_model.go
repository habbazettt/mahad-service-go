package models

import (
	"time"
)

type Mahasantri struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Nama      string    `gorm:"type:varchar(255);not null" json:"nama"`
	NIM       string    `gorm:"type:varchar(50);not null;unique" json:"nim"`
	Jurusan   string    `gorm:"type:varchar(100);not null" json:"jurusan"`
	Password  string    `gorm:"not null" json:"-"`
	Gender    string    `gorm:"type:varchar(10);not null" json:"gender"`
	MentorID  uint      `gorm:"not null" json:"mentor_id"`
	Mentor    Mentor    `gorm:"foreignKey:MentorID;constraint:OnDelete:CASCADE;" json:"mentor"`
	Hafalan   []Hafalan `gorm:"foreignKey:MahasantriID;constraint:OnDelete:CASCADE;" json:"hafalan,omitempty"`
	Absensi   []Absensi `gorm:"foreignKey:MahasantriID;constraint:OnDelete:CASCADE;" json:"absensi"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
