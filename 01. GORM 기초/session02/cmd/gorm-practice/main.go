package main

import (
	"fmt"
	"log"

	"gorm-transaction/infra"
	"gorm-transaction/internal/models"
	"gorm.io/gorm"
)

// explicitSavepointExample:
// 명시적으로 SavePoint를 생성한 후, 내부 블록에서 실패하면 RollbackTo를 호출하여 SavePoint 이전 상태를 보존.
func explicitSavepointExample() {
	log.Println("==== Explicit SavePoint Example ====")
	outerTx := infra.DB.Begin()
	if outerTx.Error != nil {
		log.Fatalf("Explicit: 외부 트랜잭션 시작 실패: %v", outerTx.Error)
	}

	// 사용자 생성
	user := models.User{
		Name:    "ExplicitUser",
		Email:   "explicit@example.com",
		Balance: 100,
	}
	if err := outerTx.Create(&user).Error; err != nil {
		outerTx.Rollback()
		log.Fatalf("Explicit: 사용자 생성 실패: %v", err)
	}

	// 명시적 SavePoint 생성
	if err := outerTx.SavePoint("sp_explicit").Error; err != nil {
		outerTx.Rollback()
		log.Fatalf("Explicit: SavePoint 생성 실패: %v", err)
	}

	// 중첩 트랜잭션에서 업데이트 작업 시도 (존재하지 않는 사용자 업데이트로 에러 유도)
	err := outerTx.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.User{}).Where("name = ?", "NonExistentUser").Update("balance", 500)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("Explicit: 업데이트된 레코드가 없으므로 롤백")
		}
		return nil
	})
	if err != nil {
		log.Printf("Explicit: 내부 트랜잭션 에러 발생: %v", err)
		if rbErr := outerTx.RollbackTo("sp_explicit").Error; rbErr != nil {
			outerTx.Rollback()
			log.Fatalf("Explicit: SavePoint 롤백 실패: %v", rbErr)
		}
	}

	if err := outerTx.Commit().Error; err != nil {
		log.Fatalf("Explicit: 커밋 실패: %v", err)
	}
	log.Println("Explicit: 외부 트랜잭션 커밋 완료 (사용자 생성은 유지되고, 업데이트는 롤백됨)")
}

// implicitSavepointExample:
// 묵시적 SavePoint 사용: 별도의 SavePoint 호출 없이 중첩 트랜잭션 헬퍼 함수가 내부에서 자동으로 SavePoint를 생성합니다.
// 내부에서 에러 발생 시 에러가 외부로 전파될 수 있으므로, 예제에서는 에러를 로깅한 후 외부 트랜잭션은 커밋하도록 처리합니다.
func implicitSavepointExample() {
	log.Println("==== Implicit SavePoint Example ====")
	outerTx := infra.DB.Begin()
	if outerTx.Error != nil {
		log.Fatalf("Implicit: 외부 트랜잭션 시작 실패: %v", outerTx.Error)
	}

	// 사용자 생성
	user := models.User{
		Name:    "ImplicitUser",
		Email:   "implicit@example.com",
		Balance: 100,
	}
	if err := outerTx.Create(&user).Error; err != nil {
		outerTx.Rollback()
		log.Fatalf("Implicit: 사용자 생성 실패: %v", err)
	}

	// 묵시적 SavePoint 사용: 별도 SavePoint 없이 중첩 트랜잭션 호출
	err := outerTx.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.User{}).Where("name = ?", "NonExistentUser").Update("balance", 500)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("Implicit: 업데이트된 레코드가 없으므로 롤백")
		}
		return nil
	})
	if err != nil {
		log.Printf("Implicit: 내부 트랜잭션 에러 발생: %v", err)
		log.Println("Implicit: 에러를 무시하고 외부 트랜잭션을 계속 진행합니다.")
	}

	if err := outerTx.Commit().Error; err != nil {
		log.Fatalf("Implicit: 커밋 실패: %v", err)
	}
	log.Println("Implicit: 외부 트랜잭션 커밋 완료 (사용자 생성은 유지되고, 업데이트는 롤백됨)")
}

func main() {
	infra.InitDB()
	if err := infra.DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("AutoMigrate 실패: %v", err)
	}

	// 기존 사용자 삭제 (테스트 초기화)
	if err := infra.DB.Unscoped().Where("1 = 1").Delete(&models.User{}).Error; err != nil {
		log.Fatalf("기존 사용자 삭제 실패: %v", err)
	}

	// 명시적 SavePoint 예제 실행
	explicitSavepointExample()
	// 묵시적 SavePoint 예제 실행
	implicitSavepointExample()

	// 최종 결과 확인
	var users []models.User
	if err := infra.DB.Find(&users).Error; err != nil {
		log.Fatalf("사용자 조회 실패: %v", err)
	}
	fmt.Println("최종 사용자 레코드:")
	for _, u := range users {
		fmt.Printf("%+v\n", u)
	}
}
