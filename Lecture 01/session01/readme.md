# GORM Tutorial: 기본 실습

## 1. 전체 학습 목표

1. **프로젝트 구조 설계**  
   - Go 모듈 내에서 `cmd`, `infra`, `internal` 등 디렉토리를 분리하여 유지보수에 유리한 구조를 만든다.

2. **데이터베이스 연동**  
   - SQLite를 예시로, GORM을 활용해 DB와 연동하는 방법을 익힌다.

3. **CRUD 실습**  
   - 간단한 User 모델을 정의하고, 생성(Create), 조회(Read), 수정(Update), 삭제(Delete) 작업을 수행해 본다.  
   - 특히, **여러 명의 사용자**를 한 번에 삽입(Bulk Create)하는 기능을 살펴본다.

4. **테스트 환경 구성**  
   - 테스트용 DB를 사용해 레포지토리 함수를 유닛 테스트하는 방법을 체득한다.

---

## 2. 패키지 구조 및 `internal` 디렉토리

### 2.1 기본 디렉토리 구성

```
gorm-practice/
├── cmd/
│   └── gorm-practice/
│       └── main.go
├── infra/
│   └── db.go
├── internal/
│   ├── models/
│   │   └── user.go
│   ├── repositories/
│   │   └── user_repository.go
│   └── services/
│       └── user_service.go
└── go.mod
```

1. **cmd/gorm-practice**  
   - 프로그램 실행의 진입점(`main.go`)을 두는 디렉토리  
2. **infra**  
   - DB 연결 설정, 외부 인프라 관련 설정 등을 배치  
3. **internal/models**  
   - GORM 모델 정의(테이블 구조)  
4. **internal/repositories**  
   - 실제 DB 쿼리 로직(CRUD 등)  
5. **internal/services**  
   - 비즈니스 로직과 도메인 규칙 처리

### 2.2 `internal` 디렉토리의 특성, 한계, 그리고 소스코드 공개 여부

- **Go 언어 차원의 접근 제한**  
  - `internal` 디렉토리 내부 코드는 **동일 모듈** 내 다른 패키지에서는 자유롭게 import할 수 있지만,  
    **모듈을 벗어난 외부 프로젝트**에서는 import 시도 시 컴파일 에러(“use of internal package not allowed”)가 발생합니다.
- **캡슐화 및 구조적 안정성**  
  - 의도치 않은 외부 접근을 막아, 내부 로직 변경 시 외부 의존성을 최소화할 수 있습니다.
- **한계**  
  1. **보안 기능은 아님**:  
     - `internal` 디렉토리에 있다고 해서, Git 리포지토리나 소스 자체가 감춰지는 것은 아닙니다.  
     - `git clone`으로 레포지토리를 내려받으면 `internal` 내부 코드를 그대로 열람할 수 있습니다.
  2. **외부 프로젝트에서 재사용 불가**:  
     - 모듈 바깥에서 `go get`으로 받아도 `internal` 내 코드를 직접 import할 수는 없습니다.  
     - 공개 API로 사용하려면 `internal` 밖(`pkg`나 루트 등)에 배치해야 합니다.
  3. **프로젝트 규모가 커질 때**:  
     - `internal` 디렉토리 내 서비스/레포지토리가 방대해지면, 내부 구조를 더 세분화하거나 모듈을 분리해야 할 수도 있습니다.

---

## 3. GORM 기본 설정 및 모델 정의

### 3.1 DB 연결 설정 (`infra/db.go`)

아래 예시는 **SQLite**를 활용한 간단한 DB 연결 예시입니다.

```go
package infra

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
    var err error
    DB, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    if err != nil {
        panic("DB 연결 실패")
    }
}
```

### 3.2 모델 정의 (`internal/models/user.go`)

```go
package models

import "gorm.io/gorm"

type User struct {
    gorm.Model         // ID, CreatedAt, UpdatedAt, DeletedAt 포함
    Name  string `gorm:"size:255"` // 최대 길이 설정
    Email string `gorm:"unique"`   // 유니크 인덱스
    Age   int
}
```

- `gorm.Model`에는 `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt` 필드가 자동 포함됩니다.  
- 태그(`gorm:"..."`)로 각 필드 설정을 세부 제어할 수 있습니다.

---

## 4. 기본 CRUD 실습

아래 코드는 `main.go`에서 **Bulk Create(여러 사용자 일괄 생성)**를 포함해 CRUD를 순서대로 시연합니다.

### 4.1 `cmd/gorm-practice/main.go`

```go
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

    // 5) Update 예시 - Potato의 이름 수정
    var potatoUser models.User
    if err := infra.DB.Where("email = ?", "potato@example.com").First(&potatoUser).Error; err != nil {
        log.Printf("[ERROR] Potato 조회 실패: %v", err)
    } else {
        potatoUser.Name = "Updated Potato"
        if err := infra.DB.Save(&potatoUser).Error; err != nil {
            log.Printf("[ERROR] 사용자 수정 실패: %v", err)
        } else {
            fmt.Printf("Potato 사용자 이름이 [%s]로 수정되었습니다.\n", potatoUser.Name)
        }
    }

    // 6) Delete 예시 - Carrot 삭제
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

    // 7) 최종 결과 확인
    var finalUsers []models.User
    if err := infra.DB.Find(&finalUsers).Error; err != nil {
        log.Printf("[ERROR] 최종 사용자 조회 실패: %v", err)
    } else {
        fmt.Printf("최종 사용자 목록: %v\n", finalUsers)
    }
}
```

---

## 5. 테스트 환경 구성 및 유닛 테스트

### 5.1 테스트용 DB 설정 (`infra/db_test.go`)

실습이나 자동화 테스트에는 **메모리 기반** SQLite를 사용하면 빠르고 편리합니다.

```go
package infra

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func InitTestDB() {
    var err error
    DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        panic("테스트 DB 연결 실패")
    }
}
```

### 5.2 유닛 테스트 예시 (`internal/repositories/user_repository_test.go`)

```go
package repositories

import (
    "testing"

    "gorm-practice/infra"
    "gorm-practice/internal/models"
)

// 예: CreateUser 함수가 있다고 가정
func TestCreateUser(t *testing.T) {
    // 1) 테스트용 DB 초기화
    infra.InitTestDB()
    infra.DB.AutoMigrate(&models.User{})

    // 2) 테스트 로직 실행
    user := models.User{Name: "TestUser", Email: "test@example.com", Age: 30}
    err := CreateUser(&user)
    if err != nil {
        t.Fatalf("유저 생성 실패: %v", err)
    }

    // 3) 필요 시 infra.DB에서 user를 다시 조회해 검증 가능
}
```

---

## 6. 결론

1. **프로젝트 구조**  
   - `cmd`, `infra`, `internal` 디렉토리를 통해 Go 모듈 내 코드를 명확히 구분했습니다.  
   - `internal` 디렉토리는 모듈 범위를 벗어난 **import**를 불가능하게 하여, 내부 로직을 안전히 캡슐화합니다.  
   - **단, 내부 코드가 아예 숨겨지는 것은 아니며**, `git clone` 시 누구든 `internal` 코드를 열람할 수 있습니다.

2. **Bulk Create 및 기본 CRUD**  
   - **여러 사용자를 한 번에 생성**하는 방법(`db.Create(&slice)`)을 포함해 `Create`, `Find`, `Save`, `Delete` 메서드를 실습하였습니다.

3. **테스트 구성**  
   - 메모리 기반 DB(SQLite)를 활용해 빠르게 테스트 환경을 구성하고,  
     레포지토리 단위 테스트를 진행할 수 있습니다.

