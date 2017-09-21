package vegeta

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	Name     string `gorm:"not null"`
	Password string `gorm:"not null"`
	Salt     string `gorm:"not null"`
	Token    string `gorm:"not null"`
	Tags     []Tag  `gorm:"ForeignKey:UserID"`
}

type Tag struct {
	gorm.Model
	UserID   uint   `gorm:"not null"`
	Name     string `gorm:"not null"`
	SomeData []Data `gorm:"ForeignKey:TagID"`
}

type Data struct {
	gorm.Model
	TagID      uint   `gorm:"not null"`
	RemoteAddr string `gorm:"not null"`
	Serialized string `gorm:"not null" sql:"type:text;"`
}
