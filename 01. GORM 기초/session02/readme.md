# Session 02: GORM 트랜잭션 및 Savepoint 실습

이 문서(Session 02)에서는 GORM을 이용한 트랜잭션의 기본 개념과 환경 설정, User 모델 정의, 그리고 트랜잭션 처리 기법(명시적 Begin/Commit/Rollback 방식과 `db.Transaction` 헬퍼 함수 방식)을 다루며, Savepoint를 활용한 중첩 트랜잭션 기법에 대해 설명합니다.

---

## Step 1: 트랜잭션의 기본 개념 이해하기

1. **트랜잭션이란?**  
   여러 데이터베이스 작업(INSERT, UPDATE, DELETE 등)을 하나의 논리적 단위로 묶어,
    - **모두 성공하면 커밋(Commit)** 하고
    - **하나라도 실패하면 롤백(Rollback)** 하는 기능입니다.

2. **트랜잭션의 4가지 특징 (ACID)**
    - **원자성 (Atomicity):** 모든 작업이 전부 성공하거나 전부 실패합니다.
    - **일관성 (Consistency):** 트랜잭션 완료 후 데이터는 항상 유효한 상태를 유지합니다.
    - **격리성 (Isolation):** 각 트랜잭션은 서로 영향을 주지 않고 독립적으로 실행됩니다.
    - **지속성 (Durability):** 커밋된 변경사항은 영구적으로 저장됩니다.

---

## Step 2: 프로젝트 및 환경 설정

1. **Go 설치 및 프로젝트 폴더 구성**
    - Go가 설치되어 있어야 합니다. ([Go 다운로드](https://go.dev/dl/))
    - 예시 프로젝트 폴더: `gorm_transaction_practice`

2. **필요 패키지 설치**
   ```bash
   go get -u gorm.io/gorm
   go get -u gorm.io/driver/sqlite
   ```

3. **프로젝트 폴더 구조 예시**
   ```
   gorm_transaction_practice/
   ├── cmd/
   │   └── gorm-practice/
   │       └── main.go
   ├── infra/
   │   └── db.go            // DB 초기화 및 글로벌 DB 객체
   ├── internal/
   │   └── models/
   │       └── user.go      // User 모델 정의
   └── go.mod
   ```

4. **infra 패키지 설정 (DB 연결 설정)**
    - `infra/db.go` 파일을 생성 후 다음과 같이 작성합니다:
      ```go
      package infra
 
      import (
          "log"
          "gorm.io/driver/sqlite"
          "gorm.io/gorm"
      )
 
      var DB *gorm.DB
 
      // InitDB는 DB 연결을 초기화합니다.
      func InitDB() {
          var err error
          DB, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
          if err != nil {
              log.Fatalf("데이터베이스 연결 실패: %v", err)
          }
          log.Println("데이터베이스 연결 성공")
      }
      ```

---

## Step 3: User 모델 정의하기

1. **모델 파일 생성**
    - `internal/models/user.go` 파일 생성

2. **User 모델 코드 작성**
   ```go
   package models

   import "gorm.io/gorm"

   // User 모델은 사용자 정보를 저장합니다.
   type User struct {
       gorm.Model
       Name    string `gorm:"size:255"`
       Email   string `gorm:"unique"`
       Balance int    // 잔액(포인트)
   }
   ```

---

## Step 4: GORM 트랜잭션 사용법 알아보기

GORM에서는 크게 두 가지 방식으로 트랜잭션을 처리할 수 있습니다.

### 4.1 명시적 Begin/Commit/Rollback 방식

- **설명:**  
  직접 `Begin()`으로 트랜잭션을 시작하고, 작업 도중 에러가 발생하면 `Rollback()`,  
  모든 작업이 성공하면 `Commit()`을 호출하여 변경사항을 확정합니다.

- **예제 코드:**
   ```go
   package main

   import (
       "gorm-transaction/infra"
       "gorm-transaction/internal/models"
       "log"
   )

   func main() {
       // DB 초기화 및 마이그레이션
       infra.InitDB()
       if err := infra.DB.AutoMigrate(&models.User{}); err != nil {
           log.Fatalf("AutoMigrate 실패: %v", err)
       }

       // 예제 사용자 생성
       user := models.User{
           Name:    "Potato",
           Email:   "potato@example.com",
           Balance: 0,
       }

       // 기존 사용자 삭제 (hard delete)
       if err := infra.DB.Unscoped().Where("1 = 1").Delete(&models.User{}).Error; err != nil {
           log.Fatalf("Delete 실패: %v", err)
           return
       }

       // 트랜잭션 시작
       tx := infra.DB.Begin()

       if err := tx.Create(&user).Error; err != nil {
           tx.Rollback()
           return
       }

       // 사용자 생성 후 잔액 업데이트
       user.Balance = 1000
       if err := tx.Save(&user).Error; err != nil {
           tx.Rollback()
           return
       }

       // balance 업데이트 (2000으로 변경)
       if err := tx.Model(&user).Update("balance", 2000).Error; err != nil {
           tx.Rollback()
           return
       }

       // 특정 조건 업데이트: "Potato" 사용자의 balance를 3000으로 변경
       updateResult := tx.Model(&models.User{}).Where("name = ?", "Potato").Update("balance", 3000)
       if updateResult.Error != nil {
           tx.Rollback()
           return
       }
       if updateResult.RowsAffected == 0 {
           log.Println("변경된 레코드가 없으므로 롤백합니다.")
           tx.Rollback()
           return
       }

       // 특정 사용자 삭제: "tomato" 사용자를 삭제 (삭제된 레코드가 없으면 롤백)
       result := tx.Where("name = ?", "tomato").Delete(&models.User{})
       if result.Error != nil {
           tx.Rollback()
           return
       }
       if result.RowsAffected == 0 {
           log.Println("삭제된 레코드가 없으므로 롤백합니다.")
           tx.Rollback()
           return
       }

       tx.Commit()
   }
   ```

### 4.2 `db.Transaction` 헬퍼 함수 방식 및 Savepoint 활용

- **설명:**  
  `db.Transaction` 헬퍼 함수를 사용하면 내부에서 Savepoint를 자동 생성하여 작업을 처리합니다.  
  Savepoint는 이미 시작된 트랜잭션 내에서 임시 저장 지점을 설정해, 해당 Savepoint 이후의 작업만 롤백할 수 있도록 해줍니다.

- **예제 코드:**
   ```go
   package main

   import (
       "fmt"
       "log"
       "gorm-transaction/infra"
       "gorm-transaction/internal/models"
       "gorm.io/gorm"
   )

   func main() {
       // DB 초기화 및 마이그레이션
       infra.InitDB()
       if err := infra.DB.AutoMigrate(&models.User{}); err != nil {
           log.Fatalf("AutoMigrate 실패: %v", err)
       }

       // 예제 사용자 생성
       user := models.User{
           Name:    "Potato",
           Email:   "potato@example.com",
           Balance: 0,
       }

       // 기존 사용자 삭제
       if err := infra.DB.Unscoped().Where("1 = 1").Delete(&models.User{}).Error; err != nil {
           log.Fatalf("Delete 실패: %v", err)
           return
       }

       // db.Transaction 헬퍼 함수를 사용하여 트랜잭션 처리
       err := infra.DB.Transaction(func(tx *gorm.DB) error {
           // 사용자 생성
           if err := tx.Create(&user).Error; err != nil {
               return err
           }

           // 중첩 Savepoint 내에서 사용자 잔액 업데이트 작업 수행
           if err := tx.Transaction(func(nestedTx *gorm.DB) error {
               // Savepoint 이후 작업: 잔액 업데이트
               user.Balance = 1000
               if err := nestedTx.Save(&user).Error; err != nil {
                   return err
               }
               // balance 업데이트 (2000으로 변경)
               if err := nestedTx.Model(&user).Update("balance", 2000).Error; err != nil {
                   return err
               }
               // "Potato" 사용자의 balance를 3000으로 변경
               updateResult := nestedTx.Model(&models.User{}).Where("name = ?", "Potato").Update("balance", 3000)
               if updateResult.Error != nil {
                   return updateResult.Error
               }
               if updateResult.RowsAffected == 0 {
                   return fmt.Errorf("업데이트된 레코드가 없으므로 롤백")
               }
               return nil
           }); err != nil {
               // nestedTx 내 에러 발생 시 해당 Savepoint까지 롤백
               return err
           }

           // 또 다른 중첩 Savepoint 내에서 사용자 삭제 작업 수행
           if err := tx.Transaction(func(nestedTx *gorm.DB) error {
               result := nestedTx.Where("name = ?", "tomato").Delete(&models.User{})
               if result.Error != nil {
                   return result.Error
               }
               if result.RowsAffected == 0 {
                   return fmt.Errorf("삭제된 레코드가 없으므로 롤백")
               }
               return nil
           }); err != nil {
               return err
           }

           // 모든 작업이 성공하면 자동 커밋됨
           return nil
       })

       if err != nil {
           log.Fatalf("트랜잭션 실패: %v", err)
       } else {
           log.Println("트랜잭션 성공, 커밋 완료")
       }
   }
   ```

### 4.3 명시적 SavePoint vs 묵시적 SavePoint 예제 비교

GORM에서는 중첩 트랜잭션에서 Savepoint를 활용하는 두 가지 방식이 있습니다.

- **명시적 SavePoint 예제:**  
  외부 트랜잭션 내에서 사용자 생성 후, 명시적으로 `SavePoint("sp_explicit")`를 생성하고,  
  중첩 트랜잭션에서 업데이트 작업을 시도합니다.  
  내부 블록에서 에러가 발생하면 `RollbackTo("sp_explicit")`를 호출하여 SavePoint 이후 작업만 롤백하고,  
  사용자 생성 작업은 유지됩니다.

- **묵시적 SavePoint 예제:**  
  외부 트랜잭션 내에서 사용자 생성 후, 별도의 SavePoint 호출 없이 `db.Transaction` 헬퍼 함수로  
  중첩 트랜잭션을 실행하면 내부에서 Savepoint가 자동 생성되어 에러 발생 시 롤백되지만,  
  에러가 외부로 전파되면 전체 트랜잭션이 롤백될 수 있습니다.  
  (예제에서는 에러를 로깅하고 무시하여 외부 트랜잭션은 커밋하도록 처리합니다.)

**예제 코드:**

```go
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
```

### 실행결과 
```bash
2025/02/02 20:37:27 데이터베이스 연결 성공
2025/02/02 20:37:27 ==== Explicit SavePoint Example ====
2025/02/02 20:37:27 Explicit: 내부 트랜잭션 에러 발생: Explicit: 업데이트된 레코드가 없으므로 롤백
2025/02/02 20:37:27 Explicit: 외부 트랜잭션 커밋 완료 (사용자 생성은 유지되고, 업데이트는 롤백됨)
2025/02/02 20:37:27 ==== Implicit SavePoint Example ====
2025/02/02 20:37:27 Implicit: 내부 트랜잭션 에러 발생: Implicit: 업데이트된 레코드가 없으므로 롤백
2025/02/02 20:37:27 Implicit: 에러를 무시하고 외부 트랜잭션을 계속 진행합니다.
2025/02/02 20:37:27 Implicit: 외부 트랜잭션 커밋 완료 (사용자 생성은 유지되고, 업데이트는 롤백됨)
최종 사용자 레코드:
{Model:{ID:11 CreatedAt:2025-02-02 20:37:27.33105457 +0900 +0900 UpdatedAt:2025-02-02 20:37:27.33105457 +0900 +0900 DeletedAt:{Time:0001-01-01 00:00:00 +0000 UTC Valid:false}} Name:ExplicitUser Email:explicit@example.com Balance:100}
{Model:{ID:12 CreatedAt:2025-02-02 20:37:27.357309279 +0900 +0900 UpdatedAt:2025-02-02 20:37:27.357309279 +0900 +0900 DeletedAt:{Time:0001-01-01 00:00:00 +0000 UTC Valid:false}} Name:ImplicitUser Email:implicit@example.com Balance:100}
```
---

## 요약 : 트랜잭션 전파(Propagation) 및 Savepoint 상세 설명

1. **트랜잭션 전파(Propagation)**
    - 여러 함수가 하나의 트랜잭션 컨텍스트 내에서 실행되도록 하려면, 상위에서 생성한 `tx *gorm.DB` 객체를 하위 함수에 전달합니다.

2. **Savepoint에 대한 추가 설명**
    - **Savepoint란?**  
      이미 시작된 트랜잭션 내에서 임시 저장 지점을 설정하는 기능입니다.  
      Savepoint 이후의 작업만 롤백할 수 있어, 마치 중첩 트랜잭션처럼 부분 롤백 효과를 얻을 수 있습니다.
    - **GORM의 Savepoint 활용**
        - `db.Transaction` 헬퍼 함수는 내부에서 Savepoint를 자동으로 생성합니다.
        - nested 트랜잭션에서 에러가 발생하면 해당 Savepoint까지 롤백되고, 그 이전의 작업은 외부 트랜잭션에 남게 됩니다.
        - 단, 개발자가 nested 트랜잭션의 에러를 외부로 전파하면 전체 트랜잭션이 롤백될 수 있으므로 주의해야 합니다.

---


## 활용 팁: GORM 트랜잭션 사용 시 주의사항

1. **트랜잭션 범위 내 작업**  
   트랜잭션 시작 후에는 반드시 해당 `tx *gorm.DB` 객체를 사용하여 모든 작업을 처리합니다.

2. **전파(Propagation) 및 Savepoint 활용**
    - 상위 트랜잭션에서 생성한 `tx` 객체를 하위 함수에 전달하여 하나의 트랜잭션 컨텍스트를 유지합니다.
    - GORM은 Savepoint를 활용하여 중첩 트랜잭션 효과를 낼 수 있습니다.  
      Savepoint는 내부 트랜잭션 블록에서 에러 발생 시 해당 블록의 작업만 롤백하도록 도와줍니다.  
      단, nested 트랜잭션에서 발생한 **에러를 상위로 전달**하면 전체 트랜잭션(outer 포함)이 롤백될 수 있습니다.