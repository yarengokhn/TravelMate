package models

import (
	"time"
)

type Expense struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TripID      uint      `gorm :"not null" json:"trip_id"`
	Trip        Trip      `gorm:"foreignKey:TripID" json:"trip,omitempty"`
	Category    string    `gorm:"not null" json:"category"`
	Amount      float64   `gorm:"not null" json:"amount"`
	ExpenseDate time.Time `gorm:"not null" json:"expense_date"`
}
