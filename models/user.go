package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	// 显式指定列名：GORM 默认会把 UserName/PassWord 映射为 user_name/pass_word，
	// 与 controllers 中手写的 username 查询不一致
	UserName string `json:"username" gorm:"column:username;uniqueIndex;not null;size:100"`
	PassWord string `json:"-" gorm:"column:password;not null"`
	Email    string `json:"email" gorm:"uniqueIndex;not null;size:255"`

	Posts    []Post    `json:"posts,omitempty" gorm:"foreignKey:UserID"`
	Comments []Comment `json:"comments,omitempty" gorm:"foreignKey:UserID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.PassWord), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.PassWord = string(hashedPassword)

	return nil
}
