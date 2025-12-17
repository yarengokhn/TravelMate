package models

import (
	"time"

	"gorm.io/gorm"
)

type Activity struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	TripID uint `gorm:"not null" json:"trip_id"`
	Trip   Trip `gorm:"foreignKey:TripID" json:"trip,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartTime   time.Time `gorm:"not null" json:"start_time"`
	EndTime     time.Time `gorm:"not null" json:"end_time"`
	Rating      uint      `gorm:"default:0"json:"rating"`
}
