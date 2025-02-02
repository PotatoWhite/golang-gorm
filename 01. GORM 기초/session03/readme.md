# 두 사용자 간 잔액 이체 실습

이번 Session 03에서는 두 사용자 간 잔액 이체(포인트 송금) 예제를 단계별로 구현합니다.  
이 실습은 Session 02에서 배운 트랜잭션 기법을 활용하여, 한 사용자의 잔액이 부족할 경우 전체 작업이 롤백되는 것을 확인합니다.

---

## Step 1: 문제 정의

- **시나리오:**  
  UserA가 UserB에게 50 포인트를 송금합니다.
- **조건:**
    - UserA의 잔액이 50 미만이면 송금이 불가하여 전체 작업이 롤백됩니다.
    - 모든 작업은 하나의 트랜잭션 단위로 처리됩니다.

---

## Step 2: 프로젝트 및 환경 준비

Session 03에서는 Session 02의 환경(Go, GORM, SQLite, User 모델 등)을 그대로 사용합니다.  
따라서 이전에 설정한 `infra/db.go` 및 `internal/models/user.go` 파일은 그대로 활용합니다.

---

## Step 3: 서비스 로직 작성

`internal/services/transfer_service.go` 파일을 생성하고, 두 사용자 간 이체를 처리하는 로직을 작성합니다.

```go
package services

import (
	"errors"
	"fmt"
	"gorm-transaction/infra"
	"gorm-transaction/internal/models"
	"gorm.io/gorm"
)

// TransferPoints는 senderID의 사용자로부터 receiverID의 사용자에게 amount만큼 포인트를 이체합니다.
// 조건: sender의 잔액이 amount 미만이면 에러를 반환하여 전체 트랜잭션을 롤백합니다.
func TransferPoints(senderID, receiverID uint, amount int) error {
	return infra.DB.Transaction(func(tx *gorm.DB) error {
		var sender, receiver models.User

		// 1. 송신자(UserA) 조회
		if err := tx.First(&sender, senderID).Error; err != nil {
			return fmt.Errorf("송신자 조회 실패: %w", err)
		}
		// 2. 수신자(UserB) 조회
		if err := tx.First(&receiver, receiverID).Error; err != nil {
			return fmt.Errorf("수신자 조회 실패: %w", err)
		}
		// 3. 잔액 확인
		if sender.Balance < amount {
			return errors.New("잔액 부족으로 이체 불가")
		}
		// 4. 이체 처리: 송신자 잔액 차감, 수신자 잔액 증가
		sender.Balance -= amount
		receiver.Balance += amount

		// 5. DB 업데이트
		if err := tx.Save(&sender).Error; err != nil {
			return fmt.Errorf("송신자 잔액 갱신 실패: %w", err)
		}
		if err := tx.Save(&receiver).Error; err != nil {
			return fmt.Errorf("수신자 잔액 갱신 실패: %w", err)
		}
		return nil
	})
}
```

---

## Step 4: 메인 함수 작성

`cmd/gorm-practice/main.go` 파일을 생성하고, 잔액 이체 서비스를 호출하는 메인 함수를 작성합니다.

```go
package main

import (
	"fmt"
	"log"
	"gorm-transaction/infra"
	"gorm-transaction/internal/models"
	"gorm-transaction/internal/services"
)

func main() {
	// DB 초기화
	infra.InitDB()

	// 기존 User 테이블 삭제 후 재생성 (테스트 초기화)
	infra.DB.Migrator().DropTable(&models.User{})
	if err := infra.DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("마이그레이션 실패: %v", err)
	}

	// 테스트용 사용자 생성
	userA := models.User{Name: "Alice", Email: "alice@example.com", Balance: 100}
	userB := models.User{Name: "Bob", Email: "bob@example.com", Balance: 10}

	if err := infra.DB.Create(&userA).Error; err != nil {
		log.Fatalf("사용자 A 생성 실패: %v", err)
	}
	if err := infra.DB.Create(&userB).Error; err != nil {
		log.Fatalf("사용자 B 생성 실패: %v", err)
	}

	fmt.Printf("초기 상태: A(%d), B(%d)\n", userA.Balance, userB.Balance)

	// 50 포인트 이체 시도
	if err := services.TransferPoints(userA.ID, userB.ID, 50); err != nil {
		log.Printf("이체 실패: %v", err)
	} else {
		fmt.Println("이체 성공")
	}

	// 결과 확인: DB에서 다시 조회하여 잔액 확인
	var updatedA, updatedB models.User
	infra.DB.First(&updatedA, userA.ID)
	infra.DB.First(&updatedB, userB.ID)

	fmt.Printf("최종 상태: A(%d), B(%d)\n", updatedA.Balance, updatedB.Balance)
}
```

---

## Step 5: 프로그램 실행 및 결과 확인

1. **프로젝트 실행**  
   터미널에서 프로젝트 폴더로 이동 후 다음 명령어 실행:
   ```bash
   go run cmd/gorm-practice/main.go
   ```

2. **출력 결과 확인**
    - **정상 이체 시:**
        - 예를 들어, UserA의 잔액이 100에서 50으로, UserB의 잔액이 10에서 60으로 변경됩니다.
    - **잔액 부족 시:**
        - UserA의 잔액이 50 미만인 경우 이체가 실패하고 전체 작업이 롤백됨을 확인할 수 있습니다.

---

## Step 6: 추가 주의사항 및 팁

- **트랜잭션 범위 내 작업:**  
  트랜잭션 시작 후에는 반드시 해당 `tx *gorm.DB` 객체를 사용하여 작업합니다.

- **에러 상황 테스트:**  
  일부러 잔액 부족 등의 조건을 만들어 롤백 동작을 확인해보세요.

---

## Step 7: 교안 요약

- **문제 정의:** 두 사용자 간 잔액 이체를 하나의 트랜잭션으로 처리하여, 송신자의 잔액이 부족할 경우 전체 작업을 롤백합니다.
- **서비스 로직:** 송신자와 수신자를 조회한 후, 잔액을 조정하고 업데이트 작업을 수행합니다.
- **전체 트랜잭션:** 에러 발생 시 전체 트랜잭션이 롤백되어, 데이터 일관성을 보장합니다.
