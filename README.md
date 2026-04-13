# 🎬 Paykodec

> Go와 PostgreSQL로 만들어보는 나만의 미디어 서버 백엔드

## 💡 프로젝트 소개

Paykodec은 Jellyfin 같은 거대한 오픈소스 미디어 서버의 핵심 동작 원리를 공부하고, 이를 Go 언어의 높은 성능과 결합해 직접 구현해 보기 위해 시작한 프로젝트입니다.

**현재 진행 상태:** 초기 데이터베이스(PostgreSQL) 환경 구성 및 핵심 테이블(Schema) 설계 완료

## 🛠 기술 스택

- Language: Go
- Database: PostgreSQL 15
- HTTP Framework: Gin (`github.com/gin-gonic/gin`)
- DB Access: `database/sql`
- Infrastructure: Docker

## 🗄️ 데이터베이스 구조 (진행 중)

- `users`: 사용자 계정
- `libraries`: 미디어 카테고리(폴더)
- `media_items`: 미디어 파일 원본 정보
- `user_media_data`: 사용자별 시청 기록 (이어보기 등)

## 🚀 실행 방법

### 요구 사항

- Go 설치 완료
- Docker Desktop 실행 중

## 프로그램 시작

1. `.env.example` 파일을 참고해서 `.env` 파일을 만듭니다.
2. `docker compose up -d`로 PostgreSQL을 실행합니다.
3. `go test ./...`로 테스트를 실행합니다.
4. `go run .`으로 서버를 실행합니다.

## API 예시

### 서버 상태 확인

```bash
curl http://localhost:8080/health
```

### DB 상태 확인

```bash
curl http://localhost:8080/health/db
```

### 라이브러리 목록 조회

```bash
curl http://localhost:8080/libraries
```

### 라이브러리 생성

```bash
curl -X POST http://localhost:8080/libraries \
  -H "Content-Type: application/json" \
  -d '{"name":"Movies","folder_path":"D:/media/movies"}'
```

### 현재 테스트

```bash
go test ./...
```

현재 포함된 테스트:

- `/health` 정상 응답 테스트
- `/health`에 잘못된 메서드를 보냈을 때의 동작 테스트
- `/health/db`의 DB Ping 성공/실패 테스트
- `POST /libraries` 요청 본문 검증 테스트

### 현재 지원 API

- `GET /health`
- `GET /health/db`
- `GET /libraries`
- `POST /libraries`

### API 응답 예시

`GET /health`

```json
{
  "status": "ok",
  "message": "server is running"
}
```

`GET /health/db`

```json
{
  "status": "ok",
  "message": "database connection is healthy"
}
```

`GET /libraries`

```json
[
  {
    "id": 1,
    "name": "Movies",
    "folder_path": "D:/media/movies"
  }
]
```

`POST /libraries`

성공 시:

```json
{
  "message": "library created"
}
```

실패 시 예시:

```json
{
  "message": "name and folder_path are required"
}
```

폴더 경로 중복 에러(HTTP 409):

```json
{
  "message": "folder_path already exists"
}
```

### `POST /libraries` 요청 규칙

- 요청은 JSON이어야 합니다.
- `name`, `folder_path`는 필수입니다.
- 앞뒤 공백은 제거한 뒤 검사합니다.
- 정의되지 않은 추가 필드가 들어오면 거부합니다.
- 중복된 `folder_path`는 DB 저장 과정에서 실패할 수 있습니다.
