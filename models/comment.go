package models

import "gorm.io/gorm"

type Comment struct {
	gorm.Model

	Content string `json:"content" gorm:"type:text;not null"`
	UserID  uint   `json:"user_id" gorm:"not null"`
	PostID  uint   `json:"post_id" gorm:"not null"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Post Post `json:"post,omitempty" gorm:"foreignKey:PostID"`
}
