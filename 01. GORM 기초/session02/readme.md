# Transaction 실습

## 1. 트랜잭션이란?
- **정의**: 다수의 데이터베이스 작업을 하나의 논리적 단위로 묶어, **모두 성공하면 커밋(Commit)** 하고, **하나라도 실패하면 롤백(Rollback)** 하는 것을 보장하는 기능.
- **특징**:   
  - **원자성(Atomicity)**: 작업 단위가 전부 성공하거나, 전부 실패(롤백).
  - **일관성(Consistency)**: 트랜잭션 완료 시, 데이터는 유효한 상태를 유지.
  - **격리성(Isolation)**: 트랜잭션 간에 상호 간섭 없이 독립적으로 동작.
  - **지속성(Durability)**: 커밋된 데이터는 영구적으로 반영.

## 2. GORM에서 트랜잭션 사용 방법
GORM에서는 크게 두 가지 방법으로 트랜잭션을 처리할 수 있습니다.

### 2.1 명시적 Begin/Commit/Rollback
직접 `Begin()`, `Commit()`, `Rollback()`을 호출하여 트랜잭션을 제어합니다.

```go
tx := db.Begin()

// 트랜잭션 내 여러 작업 수행
if err := tx.Create(&someModel).Error; err != nil {
    tx.Rollback()
    return
}

if err := tx.Model(&someModel).Update("field", "new value").Error; err != nil {
    tx.Rollback()
    return
}

if err := tx.Commit().Error; err != nil {
    // 커밋 과정에서 문제 발생 시에도, 재차 Rollback 고려
    tx.Rollback()
    return
}
```

### 2.2 GORM의 `db.Transaction` 헬퍼 함수
`db.Transaction(func(tx *gorm.DB) error {...})`를 사용하면, 중간에 에러를 반환하는 즉시 자동 롤백, 에러가 없으면 자동 커밋됩니다.

```go
err := db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&someModel).Error; err != nil {
        return err // 반환 = 즉시 롤백
    }

    if err := tx.Model(&someModel).Update("field", "new value").Error; err != nil {
        return err
    }

    return nil // nil 반환 시 자동 커밋
})

if err != nil {
    fmt.Println("트랜잭션 실패, 롤백됨:", err)
} else {
    fmt.Println("트랜잭션 성공, 커밋 완료")
}
```

---

## 3. 실습 예제: 두 사용자 간 잔액 이체

### 3.1 시나리오
- User 모델에 `Balance(잔액)` 필드를 추가.
- **UserA**와 **UserB**가 있다고 할 때, UserA → UserB에게 50 포인트를 송금(이체)한다고 가정.
- 만약 UserA의 잔액이 50 미만이면, 이체가 불가능하므로 **롤백**되어야 함.
- 이 모든 과정을 하나의 트랜잭션으로 처리해보기.

### 3.2 모델 정의 예시 (`internal/models/user.go`)

```go
package models

import "gorm.io/gorm"

type User struct {
    gorm.Model
    Name    string `gorm:"size:255"`
    Email   string `gorm:"unique"`
    Age     int
    Balance int // 잔액(포인트)
}
```

### 3.3 서비스 로직: 이체 기능 구현 (`internal/services/transfer_service.go`)

```go
package services

import (
    "errors"
    "fmt"
    "gorm-practice/infra"
    "gorm-practice/internal/models"
)

// TransferPoints는 sender → receiver 로 amount만큼 이체한다.
// 만약 sender 잔액이 부족하면 에러를 반환하며 롤백된다.
func TransferPoints(senderID, receiverID uint, amount int) error {
    return infra.DB.Transaction(func(tx *gorm.DB) error {
        var sender, receiver models.User

        // 1. 송신자 조회
        if err := tx.First(&sender, senderID).Error; err != nil {
            return fmt.Errorf("송신자 조회 실패: %w", err)
        }

        // 2. 수신자 조회
        if err := tx.First(&receiver, receiverID).Error; err != nil {
            return fmt.Errorf("수신자 조회 실패: %w", err)
        }

        // 3. 송신자 잔액 확인
        if sender.Balance < amount {
            return errors.New("잔액 부족으로 이체 불가")
        }

        // 4. 송신자 잔액 차감, 수신자 잔액 추가
        sender.Balance -= amount
        receiver.Balance += amount

        // 5. DB에 업데이트
        if err := tx.Save(&sender).Error; err != nil {
            return fmt.Errorf("송신자 잔액 갱신 실패: %w", err)
        }
        if err := tx.Save(&receiver).Error; err != nil {
            return fmt.Errorf("수신자 잔액 갱신 실패: %w", err)
        }

        // 모든 작업 정상 완료 → 자동 Commit
        return nil
    })
}
```

- **설명**  
  - `infra.DB.Transaction(...)` 을 사용해 트랜잭션을 시작.
  - 송신자와 수신자를 각각 조회하고, 송신자 잔액이 `amount`보다 적으면 에러 반환 → 즉시 롤백.
  - 두 `Save` 메서드가 모두 성공하면 `nil`을 반환 → **자동 커밋**.

### 3.4 실습 실행: (`cmd/gorm-practice/main.go`)

```go
package main

import (
    "fmt"
    "log"

    "gorm-practice/infra"
    "gorm-practice/internal/models"
    "gorm-practice/internal/services"
)

func main() {
    infra.InitDB()

    // (선택) 기존 테이블 삭제 후 재생성
    infra.DB.Migrator().DropTable(&models.User{})
    if err := infra.DB.AutoMigrate(&models.User{}); err != nil {
        log.Fatalf("마이그레이션 실패: %v", err)
    }

    // 1. 테스트용 사용자 생성
    userA := models.User{Name: "Alice", Email: "alice@example.com", Balance: 100}
    userB := models.User{Name: "Bob", Email: "bob@example.com", Balance: 10}

    if err := infra.DB.Create(&userA).Error; err != nil {
        log.Fatalf("사용자 A 생성 실패: %v", err)
    }
    if err := infra.DB.Create(&userB).Error; err != nil {
        log.Fatalf("사용자 B 생성 실패: %v", err)
    }

    fmt.Printf("처음 상태: A(%d), B(%d)\n", userA.Balance, userB.Balance)

    // 2. 트랜잭션 실습: 50 포인트 이체
    if err := services.TransferPoints(userA.ID, userB.ID, 50); err != nil {
        log.Printf("이체 실패: %v", err)
    } else {
        fmt.Println("이체 성공")
    }

    // 3. 결과 확인
    //    - DB에서 다시 조회
    var updatedA, updatedB models.User
    infra.DB.First(&updatedA, userA.ID)
    infra.DB.First(&updatedB, userB.ID)

    fmt.Printf("결과 상태: A(%d), B(%d)\n", updatedA.Balance, updatedB.Balance)
}
```

#### 결과 예시
1. **잔액이 충분할 때**  
   - `A`의 Balance가 100 → 50을 이체 → 성공 시 `A` 잔액은 50, `B` 잔액은 60으로 업데이트됩니다.
2. **잔액이 부족할 때**(예: A가 40인데 50 이체 시도)  
   - `A` 잔액이 `50` 미만 → 에러 반환 → 두 사용자 모두 잔액 변동 없음 → **롤백** 확인.

---

## 4. 트랜잭션 주의 사항
1. **트랜잭션 범위**  
   - `Begin()` ~ `Commit()`/`Rollback()` 사이에서만 유효.  
   - `db.Transaction`을 사용할 때는 콜백 함수 안에서 리턴이 발생해야 정확히 롤백/커밋이 동작함.
2. **중첩 트랜잭션**  
   - GORM에서 중첩 트랜잭션은 권장되지 않음.  
   - 트랜잭션이 필요한 로직을 상위에서 한 번에 관리하거나, 트랜잭션 객체(`tx *gorm.DB`)를 하위로 넘기는 방식을 고려.
3. **데드락(Deadlock) 및 락 경합 주의**  
   - 한 번의 트랜잭션에 너무 많은 DB 연산을 담으면, 다른 트랜잭션과 충돌해 DB 락 경합이 발생할 수 있음.  
   - 가급적 **트랜잭션 내에서 최소한의 연산**을 처리하도록 설계.

---

## 5. 요약
- **트랜잭션**은 원자적(Atomic)으로 여러 작업을 처리해, 데이터 무결성을 보장하는 핵심 기능.
- GORM은 **명시적 Begin/Commit/Rollback** 방식과, **`db.Transaction` 헬퍼 함수** 방식을 모두 지원.
- 실무에서 주로 **`db.Transaction(func(tx *gorm.DB) error {...})`** 패턴을 많이 사용:
  - **오류(return err)** → 자동 롤백  
  - **정상(return nil)** → 자동 커밋
- **이체 시나리오** 같은 다중 레코드 갱신 로직을 통해, 트랜잭션의 중요성과 사용 방법을 확실히 익힐 수 있음.

---

이상으로 **트랜잭션** 실습 교안이 완료되었습니다.  
해당 예제를 통해 실무에서도 **서로 연관된 작업들을 일괄 처리**하고,  
**오류 상황에서 전체 작업을 되돌리는(롤백) 원리**를 체득할 수 있길 바랍니다.  