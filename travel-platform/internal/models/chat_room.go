package models

import (
	"time"

	"gorm.io/gorm"
)

type ChatRoom struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"unique;size:100" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	// Cascade delete messages when room is deleted
	Messages []ChatMessage `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE" json:"messages,omitempty"`
}
