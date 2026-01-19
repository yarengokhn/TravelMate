package models

import (
	"time"

	"gorm.io/gorm"
)

type ChatMessage struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	RoomID    uint           `gorm:"index;not null" json:"room_id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Room      ChatRoom       `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	Message   string         `gorm:"type:text;not null" json:"message"` //VARCHAR'dan farkı: Uzun metinler için
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
