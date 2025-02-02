package models

import "gorm.io/gorm"

// User 모델은 사용자 정보를 저장합니다.
type User struct {
	gorm.Model
	Name  string `gorm:"size:255"`
	Email string `gorm:"unique"`

	Balance int // 잔액(포인트)
}
