package models

import "time"

type TargetSemester struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	MahasantriID uint      `gorm:"not null" json:"mahasantri_id"`
	Target       int       `gorm:"not null" json:"target"`
	Semester     string    `gorm:"type:varchar(10);not null" json:"semester"`
	TahunAjaran  string    `gorm:"type:varchar(10);not null" json:"tahun_ajaran"`
	Keterangan   string    `gorm:"type:varchar(255)" json:"keterangan,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Mahasantri Mahasantri `gorm:"foreignKey:MahasantriID;constraint:OnDelete:CASCADE;" json:"-"`
}
