package repositories

import (
	"gorm-practice/infra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitTestDB() {
	var err error
	infra.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("테스트 DB 연결 실패")
	}
}
