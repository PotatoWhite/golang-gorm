package models

import "gorm.io/gorm"

type User struct {
	gorm.Model        // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 포함
	Name       string `gorm:"size:255"` // Name 필드의 최대 길이를 255로 제한
	Email      string `gorm:"unique"`   // Email 필드는 유일해야 함 (인덱스 생성)
	Age        int    // Age 필드는 정수형
}
