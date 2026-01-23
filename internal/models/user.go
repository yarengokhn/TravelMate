package models

type User struct {
	ID           uint          `gorm:"primaryKey" json:"id"`
	FirstName    string        `gorm:"not null" json:"first_name"`   //mandotary fields
	LastName     string        `gorm:"not null" json:"last_name"`    //mandotary fields
	Email        string        `gorm:"unique;not null" json:"email"` //mandotary fields
	Password     string        `gorm:"not null" json:"-"`            //mandotary fields,hides from json
	Trips        []Trip        `gorm:"foreignKey:UserID" json:"trips,omitempty"`
	ChatMessages []ChatMessage `gorm:"foreignKey:UserID" json:"chat_messages,omitempty"`
	//omitempty :if this field is empty it won't be shown in the json
}
