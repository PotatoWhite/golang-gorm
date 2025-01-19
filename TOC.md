
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