package models

import "time"

type Mentor struct {
	ID                   uint                `gorm:"primaryKey" json:"id"`
	Nama                 string              `gorm:"type:varchar(255);not null" json:"nama"`
	Email                string              `gorm:"type:varchar(100);not null;unique" json:"email"`
	Password             string              `gorm:"not null" json:"-"`
	Gender               string              `gorm:"type:varchar(10);not null" json:"gender"`
	IsDataMurojaahFilled bool                `gorm:"default:false" json:"is_data_murojaah_filled"`
	JadwalPersonal       *JadwalPersonal     `gorm:"foreignKey:MentorID;constraint:OnDelete:CASCADE;" json:"jadwal_personal,omitempty"`
	Mahasantri           []Mahasantri        `gorm:"foreignKey:MentorID;constraint:OnDelete:CASCADE;" json:"mahasantri,omitempty"` // omitempty lebih baik
	JadwalRekomendasi    []JadwalRekomendasi `gorm:"foreignKey:MentorID;constraint:OnDelete:CASCADE;" json:"jadwal_rekomendasi,omitempty"`
	Absensi              []Absensi           `gorm:"foreignKey:MentorID;constraint:OnDelete:CASCADE;" json:"absensi,omitempty"`
	Hafalan              []Hafalan           `gorm:"foreignKey:MentorID;constraint:OnDelete:CASCADE;" json:"hafalan,omitempty"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
}
