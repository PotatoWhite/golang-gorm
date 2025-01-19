
## GORM

### 1 기초

1. **GORM 소개**
    - GORM의 특징과 설치
    - 기본 설정(DB 연결, DB 드라이버 사용)
2. **모델 정의**
    - 구조체를 이용한 테이블 매핑
    - 필드 태그(`gorm:"column:..."`, `primaryKey`, `unique` 등)
3. **기본 CRUD**
    - Create(생성), Read(조회), Update(갱신), Delete(삭제)
    - `db.First`, `db.Find`, `db.Create`, `db.Save`, `db.Delete` 등 주요 메서드
4. **쿼리 기초**
    - 조건(`Where`), 정렬(`Order`), 제한(`Limit`), 오프셋(`Offset`) 등
    - Migrations(마이그레이션)과 AutoMigrate
5. **관계 맵핑 기초**
    - 1:1, 1:N 관계(예: `User` - `Profile`, `User` - `Posts`)
    - 기본 Association 사용법
6. **간단한 예제 프로젝트**
    - 작은 프로젝트에서 DB 스키마 설계 & CRUD 로직 구현
    - 예: 게시판, TODO 리스트에 DB 적용

---

### 2 심화

1. **고급 쿼리**
    - 복잡한 WHERE 조건, OR, AND 결합
    - Select 절 조정
    - Group By, Having
    - SubQuery, 조인(Join) 활용
2. **트랜잭션(Transactions)**
    - `db.Transaction` 문법
    - 커밋(commit), 롤백(rollback) 제어
3. **관계 매핑 심화**
    - N:N 관계(중간 테이블), Polymorphic 관계
    - Association(연관 데이터) 처리 기법
    - Preload, Eager Loading, Lazy Loading
4. **마이그레이션 심화**
    - AutoMigrate의 한계와 직접 SQL 마이그레이션
    - 마이그레이션 도구(Migrate, Goose 등)와 연동
5. **성능 최적화**
    - N+1 문제 해결(Preload 적절히 사용)
    - 인덱스 설계, 캐싱 전략 개요
    - 로깅, 쿼리 최적화
6. **GORM Hooks(콜백)**
    - `BeforeCreate`, `AfterCreate` 등 훅 사용
    - 데이터 무결성 검증, 자동 처리
7. **프로젝트 구성**
    - Gin, Echo 등 웹 프레임워크와의 연동
    - 계층형 아키텍처(Service/Repository Layer)에서의 GORM 활용

---

### 3 실무

1. **Raw Query 및 고급 SQL 활용**
    - GORM Raw SQL(`db.Raw`) 사용법
    - Subquery, Window Functions, DB 벤더별 확장 기능
2. **다중 DB & 샤딩(Sharding)**
    - Master/Slave 구성, 읽기/쓰기 분리
    - 샤딩 전략, GORM에서의 구현
3. **고급 성능 튜닝**
    - 대규모 트래픽 대응을 위한 DB 구조 설계
    - 인덱스 튜닝, 캐시 레이어(Redis 등) 연동
    - 고성능 쿼리 작성과 실행 계획 분석
4. **GORM 플러그인/확장**
    - GORM 소스 코드 분석
    - 직접 플러그인(Callback) 작성, 커스텀 기능 추가
5. **엔터프라이즈 환경에서의 GORM**
    - 대규모 프로젝트에서의 모범 사례
    - 마이크로서비스 아키텍처에서의 DB 통합 전략
    - DDD(Domain Driven Design) 관점에서의 엔티티 매핑
6. **Zero-Downtime Migration**
    - 롤링 업데이트 환경에서의 DB 마이그레이션
    - 배포 자동화와 연계
7. **보안 및 안정성**
    - SQL 인젝션 방어
    - 동시성 충돌 처리(Optimistic Locking 등)
    - 대규모 운영 시 모니터링, 로깅, 알림
8. **실전 사례 분석**
    - 대형 서비스(전자상거래, SNS 등)에서의 GORM 사용 사례
    - 오픈소스 프로젝트(GORM 추가 확장 기능, 예제 코드) 분석






# 트랜잭션 (심화) - 이거 따로 정리할 필요있음

## 1. 기본(Basic)

1. **트랜잭션 개념과 필요성**  
   - 트랜잭션 정의 (ACID, 원자성/일관성/격리성/지속성)  
   - ORM 관점에서 트랜잭션이 중요한 이유

2. **GORM 트랜잭션 기초**  
   - `Begin`/`Commit`/`Rollback` 메서드 개요  
   - 간단한 예제: 한 함수 내에서 `tx := db.Begin()`, 여러 CRUD 후 `tx.Commit() / tx.Rollback()`
   - `db.Transaction(func(tx *gorm.DB) error { ... })` 함수형 트랜잭션 소개

3. **에러 처리와 트랜잭션 흐름**  
   - `if err != nil` 시점에서 `tx.Rollback()`  
   - `nil` 반환 시 `Commit`  
   - `RowsAffected` 등 결과값 점검 방식

4. **트랜잭션 범위 및 스코프**  
   - 함수 단위, 요청 단위로 트랜잭션 묶기  
   - 동일한 `tx *gorm.DB` 객체를 여러 함수에 전달해서 사용하는 방법

5. **실습 예제**  
   - 단순 시나리오: “한 사용자 생성 → 다른 사용자 생성 → 오류 발생 시 롤백”  
   - `db.Transaction()` vs. `Begin()`/`Commit()`

---

## 2. 심화(Advanced)

1. **중첩 트랜잭션(Nested Transaction) 이슈**  
   - GORM에서 중첩 트랜잭션을 권장하지 않는 이유  
   - 트랜잭션 객체를 상위에서 만들어 하위 함수로 주입하는 패턴  
   - 중첩 트랜잭션이 필요한 상황과 대안

2. **트랜잭션 격리 수준(Isolation Level)**  
   - DB별 기본 격리 수준(READ COMMITTED, REPEATABLE READ 등)  
   - GORM에서 Isolation Level 설정(가능 시 raw SQL, driver-specific 옵션)

3. **동시성 제어, 데드락(Deadlock) 방지**  
   - InnoDB 락 매커니즘(MySQL) 예시  
   - Optimistic Lock vs. Pessimistic Lock  
   - 트랜잭션이 긴 경우에 발생할 수 있는 문제들(Timeout, Lock Wait Timeout)

4. **트랜잭션 성능 최적화**  
   - 쿼리 수 줄이기(Batch Insert/Update)  
   - 트랜잭션 내에서 최소 로직만 수행, 빠른 Commit  
   - Index 활용, Deadlock 모니터링

5. **트랜잭션과 Raw SQL**  
   - `tx.Raw()`, `tx.Exec()` 사용 시 주의점  
   - OR마저도 안 되는 복잡한 쿼리를 트랜잭션 범위에서 실행할 때 패턴

6. **트랜잭션 로깅 및 감사(Auditing)**  
   - 트랜잭션이 언제 시작/종료되었는지 로깅  
   - 에러 상황에서 Rollback 이슈 추적

7. **심화 예제**  
   - 복수 테이블 갱신 로직에서 동시성 이슈 처리  
   - Isolation Level을 변경하여 테스트해보기

---

## 3. 실무(Production / Real Practice)

1. **대규모 시스템에서의 트랜잭션 설계**  
   - 마이크로서비스 환경에서의 분산 트랜잭션(Saga, 2PC, Outbox 패턴 간략 소개)  
   - GORM은 주로 단일 DB 트랜잭션에 초점을 둠 → 분산 트랜잭션 시 대안 고찰

2. **트랜잭션 모니터링 및 운영**  
   - 트랜잭션 모니터링 툴, DB에서의 Lock/Deadlock 모니터링  
   - Slow query, Deadlock 발생 시 로그 분석

3. **트랜잭션 에러 핸들링과 재시도(Retry) 전략**  
   - DB Deadlock 시 재시도 로직(Retry Pattern)  
   - GORM에서의 재시도 구현 예시(어떻게 트랜잭션을 재시작?)

4. **테스트 및 QA**  
   - 트랜잭션 단위 테스트 vs 통합 테스트(실제 DB)  
   - Mock DB 적용할 때 주의 사항(트랜잭션 모킹)

5. **트랜잭션 베스트 프랙티스**  
   - “짧고 굵게” 트랜잭션 유지 → 락 경쟁 줄이기  
   - 비즈니스 로직(계산, 파일 처리 등)을 트랜잭션 밖에서 수행

6. **실무 사례 연구**  
   - 실제 구현: “이체(Transfer) 로직”에서 동시성 문제 해결  
   - “주문 생성 → 재고 감소” 시나리오 등에서 트랜잭션 범위 설정

7. **결론 및 참고자료**  
   - GORM 오피셜 문서, driver별 트랜잭션 이슈, DB-specific 문서(MySQL, PostgreSQL 등)  
   - 마이크로서비스/분산 환경에서의 GORM 트랜잭션 한계와 대안

---

## 요약

- **기본(Basic)** 단계에서는 GORM 트랜잭션의 함수형 사용법(`db.Transaction`)과 에러 처리, `Begin/Commit/Rollback` 흐름을 소개합니다.  
- **심화(Advanced)** 단계에서는 Isolation Level, Lock/Deadlock 대응, 중첩 트랜잭션 문제, Raw SQL과 혼합 사용 등 더욱 복잡한 상황을 다룹니다.  
- **실무(Production)** 단계에서는 대규모 시스템/마이크로서비스 관점에서의 트랜잭션 설계, 모니터링/재시도, 베스트 프랙티스와 실전 시나리오를 확인합니다.

