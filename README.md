# 🎬 Paykodec

> Go와 PostgreSQL로 만들어보는 나만의 미디어 서버 백엔드

## 💡 프로젝트 소개

Paykodec은 Jellyfin 같은 거대한 오픈소스 미디어 서버의 핵심 동작 원리를 공부하고, 이를 Go 언어의 높은 성능과 결합해 직접 구현해 보기 위해 시작한 프로젝트입니다.

**현재 진행 상태:** 초기 데이터베이스(PostgreSQL) 환경 구성 및 핵심 테이블(Schema) 설계 완료

## 🛠 기술 스택

- **Language**: Go
- **Database**: PostgreSQL 15, `database/sql`
- **Infrastructure**: Docker, Docker Compose

## 🗄️ 데이터베이스 구조 (진행 중)

- `users`: 사용자 계정
- `libraries`: 미디어 카테고리(폴더)
- `media_items`: 미디어 파일 원본 정보
- `user_media_data`: 사용자별 시청 기록 (이어보기 등)

## 🚀 실행 방법

### 요구 사항

- Go 설치 완료
- Docker Desktop 실행 중

### 로컬 테스트

1. 환경 변수 파일(`.env`) 생성

```env
DB_USER=본인설정아이디
DB_PASSWORD=본인설정비밀번호
```

2. 데이터베이스 컨테이너 실행

```bash
docker compose up -d
```

3. 데이터베이스 테이블 초기화 및 연결 확인

```bash
go run main.go
```
