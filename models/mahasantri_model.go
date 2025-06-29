package models

import "time"

type Mahasantri struct {
	ID                   uint                `gorm:"primaryKey" json:"id"`
	Nama                 string              `gorm:"type:varchar(255);not null" json:"nama"`
	NIM                  string              `gorm:"type:varchar(50);not null;unique" json:"nim"`
	Jurusan              string              `gorm:"type:varchar(100);not null" json:"jurusan"`
	Password             string              `gorm:"not null" json:"-"`
	Gender               string              `gorm:"type:varchar(10);not null" json:"gender"`
	IsDataMurojaahFilled bool                `gorm:"default:false" json:"is_data_murojaah_filled"`
	MentorID             uint                `gorm:"not null" json:"mentor_id"`
	Mentor               *Mentor             `gorm:"foreignKey:MentorID;constraint:OnDelete:CASCADE;" json:"mentor,omitempty"`
	JadwalPersonal       *JadwalPersonal     `gorm:"foreignKey:MahasantriID;constraint:OnDelete:CASCADE;" json:"jadwal_personal,omitempty"`
	LogHarians           []LogHarian         `gorm:"foreignKey:MahasantriID;constraint:OnDelete:CASCADE;" json:"log_harians,omitempty"`
	JadwalRekomendasis   []JadwalRekomendasi `gorm:"foreignKey:MahasantriID;constraint:OnDelete:CASCADE;" json:"jadwal_rekomendasi,omitempty"`
	Hafalan              []Hafalan           `gorm:"foreignKey:MahasantriID;constraint:OnDelete:CASCADE;" json:"hafalan,omitempty"`
	Absensi              []Absensi           `gorm:"foreignKey:MahasantriID;constraint:OnDelete:CASCADE;" json:"absensi,omitempty"`
	TargetSemester       []TargetSemester    `gorm:"foreignKey:MahasantriID;constraint:OnDelete:CASCADE;" json:"target_semester,omitempty"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
}
