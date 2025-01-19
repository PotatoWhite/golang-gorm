package main

import (
	"fmt"
	"log"

	"gorm-practice/infra"
	"gorm-practice/internal/models"
)

func main() {
	// 1) DB 초기화
	infra.InitDB()

	// 2) (실습용) 기존 User 테이블 삭제 후 재생성
	if err := infra.DB.Migrator().DropTable(&models.User{}); err != nil {
		log.Printf("[WARN] 테이블 삭제 실패(존재하지 않을 수 있음): %v", err)
	}
	if err := infra.DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("[FATAL] 테이블 마이그레이션 실패: %v", err)
	}

	// 3) Bulk Create (여러 사용자 생성)
	newUsers := []models.User{
		{Name: "Potato", Email: "potato@example.com", Age: 11},
		{Name: "Tomato", Email: "tomato@example.com", Age: 12},
		{Name: "Carrot", Email: "carrot@example.com", Age: 13},
	}
	if err := infra.DB.Create(&newUsers).Error; err != nil {
		log.Fatalf("[FATAL] 사용자 일괄 생성 실패: %v", err)
	}
	fmt.Printf("생성된 사용자들: %#v\n", newUsers)

	// 4) Read - 전체 조회
	var allUsers []models.User
	if err := infra.DB.Find(&allUsers).Error; err != nil {
		log.Printf("[ERROR] 사용자 전체 조회 실패: %v", err)
	} else {
		fmt.Printf("전체 사용자: %v\n", allUsers)
	}

	// ---------------------------------------------
	// Update 예시 1) Save: 조회 후 특정 필드 수정 & Save
	// ---------------------------------------------
	var potatoUser models.User
	if err := infra.DB.Where("email = ?", "potato@example.com").First(&potatoUser).Error; err != nil {
		log.Printf("[ERROR] Potato 조회 실패: %v", err)
	} else {
		potatoUser.Name = "Updated Potato"
		if err := infra.DB.Save(&potatoUser).Error; err != nil {
			log.Printf("[ERROR] 사용자 Save 실패: %v", err)
		} else {
			fmt.Printf("Save로 Potato의 Name 갱신: [%s]\n", potatoUser.Name)
		}
	}

	// --------------------------------------------------
	// Update 예시 2) 특정 필드만 Update
	//    - 조건이 맞는 하나의 레코드에 대해 한 필드만 변경
	// --------------------------------------------------
	if err := infra.DB.Model(&models.User{}).
		Where("email = ?", "tomato@example.com").
		Update("Age", 20).Error; err != nil {
		log.Printf("[ERROR] Tomato Age 업데이트 실패: %v", err)
	} else {
		fmt.Println("Tomato Age가 20으로 업데이트되었습니다.")
	}

	// ---------------------------------------------------
	// Update 예시 3) 다중 필드 Updates
	//    - 조건이 맞는 레코드에 대해 여러 필드를 한 번에 변경
	// ---------------------------------------------------
	if err := infra.DB.Model(&models.User{}).
		Where("email = ?", "carrot@example.com").
		Updates(map[string]interface{}{
			"Name": "Fresh Carrot",
			"Age":  15,
		}).Error; err != nil {
		log.Printf("[ERROR] Carrot 다중 업데이트 실패: %v", err)
	} else {
		fmt.Println("Carrot Name과 Age가 동시 업데이트되었습니다.")
	}

	// 5) Delete 예시 - Carrot 삭제
	var carrotUser models.User
	if err := infra.DB.Where("email = ?", "carrot@example.com").First(&carrotUser).Error; err != nil {
		log.Printf("[ERROR] Carrot 조회 실패(이미 삭제됐을 수도 있음): %v", err)
	} else {
		if err := infra.DB.Delete(&carrotUser).Error; err != nil {
			log.Printf("[ERROR] Carrot 삭제 실패: %v", err)
		} else {
			fmt.Printf("Carrot 사용자(ID=%d) 삭제 완료\n", carrotUser.ID)
		}
	}

	// 6) 최종 결과 확인
	var finalUsers []models.User
	if err := infra.DB.Find(&finalUsers).Error; err != nil {
		log.Printf("[ERROR] 최종 사용자 조회 실패: %v", err)
	} else {
		fmt.Printf("최종 사용자 목록: %v\n", finalUsers)
	}
}
