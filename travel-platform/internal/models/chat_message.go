package models

type ChatMessage struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	UserID  uint   `gorm:"not null" json:"user_id"`
	User    User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Message string `gorm:"not null" json:"message"`
	SentAt  int64  `gorm:"not null" json:"sent_at"`
}
