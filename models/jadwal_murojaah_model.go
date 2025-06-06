package models

type JadwalMurojaah struct {
	ID           uint `gorm:"primaryKey" json:"id"`
	MahasantriID uint `gorm:"not null" json:"mahasantri_id"`
	// Lengkapi Sesuai Hasil Q-Learning

	Mahasantri Mahasantri `gorm:"foreignKey:MahasantriID;constraint:OnDelete:CASCADE;" json:"mahasantri"`
}
