package models

import (
	"time"

	"gorm.io/gorm"
)

type Mentor struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Email     string         `gorm:"type:varchar(100);not null;unique" json:"email"`
	Password  string         `gorm:"not null" json:"-"`
	Gender    string         `gorm:"type:varchar(10);not null" json:"gender"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
