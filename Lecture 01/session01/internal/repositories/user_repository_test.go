package repositories

import (
	"gorm-practice/infra"
	"gorm-practice/internal/models"
	"testing"
)

func TestCreateUser(t *testing.T) {
	// 1) 테스트용 DB 초기화
	InitTestDB()
	infra.DB.AutoMigrate(&models.User{})

	// 2) 테스트 로직 실행
	user := models.User{Name: "TestUser", Email: "test@example.com", Age: 30}
	err := CreateUser(&user)
	if err != nil {
		t.Fatalf("유저 생성 실패: %v", err)
	}

	// 3) 필요 시 infra.DB에서 user를 다시 조회해 검증 가능
}
