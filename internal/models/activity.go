package models

import (
	"time"
)

type Activity struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TripID      uint      `gorm:"not null" json:"trip_id"`
	Trip        Trip      `gorm:"foreignKey:TripID" json:"trip,omitempty"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Date        time.Time `gorm:"not null" json:"date"`
	StartTime   time.Time `json:"start_time"`
	Rating      uint      `json:"rating"`
}
