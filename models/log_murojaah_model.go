package models

import "time"

type StatusDetailLog string

const (
	StatusSesiBelumSelesai StatusDetailLog = "Belum Selesai"
	StatusSesiSelesai      StatusDetailLog = "Selesai"
)

type LogHarian struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	MahasantriID        uint      `gorm:"not null;uniqueIndex:idx_mahasantri_tanggal" json:"mahasantri_id"`
	Tanggal             time.Time `gorm:"type:date;not null;uniqueIndex:idx_mahasantri_tanggal" json:"tanggal"`
	TotalTargetHalaman  int       `gorm:"default:0" json:"total_target_halaman"`
	TotalSelesaiHalaman int       `gorm:"default:0" json:"total_selesai_halaman"`

	Mahasantri Mahasantri  `gorm:"foreignKey:MahasantriID;constraint:OnDelete:CASCADE;" json:"-"`
	DetailLogs []DetailLog `gorm:"foreignKey:LogHarianID;constraint:OnDelete:CASCADE;" json:"detail_logs"`
}

type DetailLog struct {
	ID                  uint            `gorm:"primaryKey"`
	LogHarianID         uint            `gorm:"not null"`
	WaktuMurojaah       string          `gorm:"not null"`
	TargetStartJuz      int             `gorm:"not null"`
	TargetStartHalaman  int             `gorm:"not null"`
	TargetEndJuz        int             `gorm:"not null"`
	TargetEndHalaman    int             `gorm:"not null"`
	SelesaiEndJuz       int             `gorm:"default:0"`
	SelesaiEndHalaman   int             `gorm:"default:0"`
	TotalTargetHalaman  int             `gorm:"default:0"`
	TotalSelesaiHalaman int             `gorm:"default:0"`
	Status              StatusDetailLog `gorm:"type:varchar(50);default:'Belum Selesai'"`
	Catatan             string          `gorm:"type:text"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
