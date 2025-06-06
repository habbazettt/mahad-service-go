package models

import "time"

type JadwalRekomendasi struct {
	ID                        uint      `gorm:"primaryKey" json:"id"`
	MahasantriID              *uint     `gorm:"index" json:"mahasantri_id,omitempty"`
	MentorID                  *uint     `gorm:"index" json:"mentor_id,omitempty"`
	State                     string    `gorm:"not null" json:"state"`
	RekomendasiJadwal         string    `gorm:"not null" json:"rekomendasi_jadwal"`
	TipeRekomendasi           string    `gorm:"not null" json:"tipe_rekomendasi"`
	EstimasiQValue            *float64  `gorm:"null" json:"estimasi_q_value"`
	PersentaseEfektifHistoris *float64  `gorm:"null" json:"persentase_efektif_historis"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`

	Mahasantri Mahasantri `gorm:"foreignKey:MahasantriID;constraint:OnDelete:SET NULL;" json:"-"`
	Mentor     Mentor     `gorm:"foreignKey:MentorID;constraint:OnDelete:SET NULL;" json:"-"`
}
