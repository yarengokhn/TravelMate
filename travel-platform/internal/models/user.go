package models

type User struct {
	ID           uint          `gorm:"primaryKey" json:"id"`
	FirstName    string        `gorm:"not null" json:"first_name"`
	LastName     string        `gorm:"not null" json:"last_name"`
	Email        string        `gorm:"unique;not null" json:"email"`
	Password     string        `gorm:"not null" json:"-"`
	Trips        []Trip        `gorm:"foreignKey:UserID" json:"trips,omitempty"`
	ChatMessages []ChatMessage `gorm:"foreignKey:UserID" json:"chat_messages,omitempty"`
}
