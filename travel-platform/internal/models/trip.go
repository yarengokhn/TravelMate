package models

import (
	"time"

	"gorm.io/gorm"
)

type Trip struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Title       string    `gorm:"not null" json:"title"`
	Destination string    `gorm:"not null" json:"destination"`
	StartDate   time.Time `gorm:"not null" json:"start_date"`
	EndDate     time.Time `gorm:"not null" json:"end_date"`
	Description string    `json:"description"`
	Budget      float64   `json:"budget"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Expenses []Expense `gorm:"foreignKey:TripID" json:"expenses,omitempty"`

	// Itineraries []Itinerary `gorm:"foreignKey:TripID" json:"itineraries,omitempty"`

	Activities []Activity `gorm:"foreignKey:TripID" json:"activities,omitempty"`
}
