package models

import "gorm.io/gorm"

type Post struct {
	gorm.Model

	Title   string `json:"title" gorm:"not null"`
	Content string `json:"content" gorm:"type:text;not null"`
	UserID  uint   `json:"user_id" gorm:"not null"`

	Comments []Comment `json:"comments,omitempty" gorm:"foreignKey:PostID"`
	User     User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
