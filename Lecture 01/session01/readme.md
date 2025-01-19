# GORM Tutorial 교안

---

## **전체 학습 목표**

1. **GORM의 기본 개념과 주요 기능 이해**: ORM의 목적과 GORM이 제공하는 기능에 대한 이해를 바탕으로, Go 환경에서 데이터베이스 작업을 간소화하는 방법 학습.
2. **실제 데이터베이스와 연동**: GORM 설치부터 테이블 정의, 데이터 생성 및 조회 작업을 통해 데이터베이스와 상호작용하는 환경을 구축.
3. **데이터 관리 작업 수행**: CRUD 작업, 조건부 쿼리, 트랜잭션 등 실무에서 자주 사용되는 기능을 적용하여 생산성을 높이는 방법 익히기.

---

## **1. GORM 소개**

### **1.1 GORM의 특징과 설치**

- **ORM(Object-Relational Mapping) 이란?**\
  객체지향 프로그래밍에서 데이터를 RDB(Relational Database) 테이블과 객체로 매핑해주는 기술.

- **GORM의 주요 특징**

  - 간단한 CRUD 메서드 제공: `Create`, `Find`, `Save`, `Delete` 등.
  - 자동 마이그레이션: 구조체를 기반으로 테이블 자동 생성.
  - 관계 설정 지원: 1:1, 1\:N, N\:N 관계 매핑.
  - 다양한 DB 드라이버 지원: MySQL, PostgreSQL, SQLite 등.

- **설치 방법**:

  ```bash
  go get -u gorm.io/gorm
  go get -u gorm.io/driver/sqlite
  ```

---

## **2. GORM 기본 설정 및 모델 정의**

### **2.1 기본 설정** (`infra/db.go`)

- SQLite를 활용한 데이터베이스 초기화 및 연결:
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

### **2.2 모델 정의** (`internal/models/user.go`)

- 구조체를 사용해 데이터베이스 테이블 정의:

  ```go
  package models

  import "gorm.io/gorm"

  type User struct {
    gorm.Model        // ID, CreatedAt, UpdatedAt, DeletedAt 필드를 포함
    Name       string `gorm:"size:255"` // Name 필드의 최대 길이를 255로 제한
    Email      string `gorm:"unique"`   // Email 필드는 유일해야 함 (인덱스 생성)
    Age        int    // Age 필드는 정수형
  }
  ```

- **`gorm.Model`**** 포함 필드**: `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`.

---

## **3. 기본 CRUD 실습**

### **3.1 데이터 생성 (Create)**

- **예제 코드** (`cmd/gorm-practice/main.go`):
  ```go
  func main() {
      infra.InitDB()
      infra.DB.AutoMigrate(&models.User{})

      user := models.User{Name: "Alice", Email: "alice@example.com", Age: 25}
      infra.DB.Create(&user)
  }
  ```

### **3.2 데이터 조회 (Read)**

- **단일 데이터 조회** (`cmd/gorm-practice/main.go`):

  ```go
  var user models.User
  infra.DB.First(&user)
  ```

- **조건부 데이터 조회** (`cmd/gorm-practice/main.go`):

  ```go
  var users []models.User
  infra.DB.Where("age > ?", 20).Find(&users)
  ```

### **3.3 데이터 수정 (Update)**

- **예제 코드** (`cmd/gorm-practice/main.go`):
  ```go
  var user models.User
  infra.DB.First(&user)
  user.Name = "Updated Name"
  infra.DB.Save(&user)
  ```

### **3.4 데이터 삭제 (Delete)**

- **예제 코드** (`cmd/gorm-practice/main.go`):
  ```go
  infra.DB.Delete(&user)
  ```

---

## **4. 패키지 구조 및 프로젝트 설계**

### **4.0 internal 디렉토리의 특수성**
- **internal 디렉토리의 목적**: 
  `internal` 디렉토리는 Go에서 제공하는 모듈 범위 접근 제한 기능을 구현하기 위해 사용됩니다. 
  `internal` 내부의 패키지는 동일 모듈 내에서만 접근 가능하며, 외부 모듈에서 직접 사용할 수 없습니다.
  이를 통해 내부 구현 세부 사항을 캡슐화하고 외부로부터 의도하지 않은 접근을 방지할 수 있습니다.
- **예시**: 
  만약 `gorm-practice`가 다른 프로젝트에 포함되더라도, `internal/repositories`나 `internal/services`에 정의된 코드는 해당 프로젝트 외부에서 접근할 수 없습니다.

---

## **4.1 디렉토리 구조**

### **4.1 디렉토리 구조**

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

### **4.2 코드 설명**

#### **cmd/gorm-practice/main.go**

```go
package main

import (
    "gorm-practice/infra"
    "gorm-practice/internal/models"
    "gorm-practice/internal/services"
)

func main() {
    infra.InitDB()
    infra.DB.AutoMigrate(&models.User{})

    // 사용자 생성
    user, _ := services.RegisterUser("Alice", "alice@example.com", 25)

    // 사용자 조회
    fetchedUser, _ := services.GetUserDetails(user.ID)
    fmt.Println(fetchedUser)
}
```

#### **internal/repositories/user\_repository.go**

```go
package repositories

import (
    "gorm-practice/infra"
    "gorm-practice/internal/models"
)

func CreateUser(user *models.User) error {
    return infra.DB.Create(user).Error
}

func GetUserByID(id uint) (*models.User, error) {
    var user models.User
    result := infra.DB.First(&user, id)
    return &user, result.Error
}
```

---

## **5. 테스트 환경 구성 및 유닛 테스트**

### **5.1 테스트용 데이터베이스 설정** (`infra/db_test.go`)

- SQLite 메모리 데이터베이스를 사용:
  ```go
  func InitTestDB() {
      var err error
      DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
      if err != nil {
          panic("테스트 DB 연결 실패")
      }
  }
  ```

### **5.2 유닛 테스트 코드**

#### **user\_repository\_test.go** (`internal/repositories/user_repository_test.go`)

```go
package repositories

import (
    "gorm-practice/infra"
    "gorm-practice/internal/models"
    "testing"
)

func TestCreateUser(t *testing.T) {
    infra.InitTestDB()
    infra.DB.AutoMigrate(&models.User{})

    user := models.User{Name: "Test User", Email: "test@example.com", Age: 30}
    err := CreateUser(&user)
    if err != nil {
        t.Fatalf("유저 생성 실패: %v", err)
    }
}
```

---

## **결론**

- **학습 목표 달성**: GORM을 사용해 데이터베이스와의 기본 상호작용 및 확장 가능한 프로젝트 구조 설계.
- **확장 가능성**: 패키지 분리를 통해 실무 환경에서도 쉽게 적용 가능.

