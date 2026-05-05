# 🎬 Paykodec

> Go와 PostgreSQL로 만들어보는 나만의 미디어 서버 백엔드

## 💡 프로젝트 소개

Paykodec은 Jellyfin 같은 거대한 오픈소스 미디어 서버의 핵심 동작 원리를 공부하고, 이를 Go 언어의 높은 성능과 결합해 직접 구현해 보기 위해 시작한 프로젝트입니다.

**현재 진행 상태:** JWT 권한 분리(admin/general), URL 기반 페이지 라우팅(`/`, `/settings`, `/dashboard`, `/libraries/:id`)과 라이브러리 카드 UI 적용

## 🛠 기술 스택

- Language: Go
- Database: PostgreSQL 15
- HTTP Framework: Gin (`github.com/gin-gonic/gin`)
- DB Access: `database/sql`
- Infrastructure: Docker
- Frontend: React + Vite + TypeScript

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

## 프론트엔드 시작

1. `frontend` 폴더로 이동합니다.
2. `npm install`로 의존성을 설치합니다.
3. `npm run dev`로 프론트엔드 개발 서버를 실행합니다.

```bash
cd frontend
npm install
npm run dev
```

프론트엔드 기본 주소: `http://localhost:5173`

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

### 라이브러리 미디어 목록 조회

```bash
curl http://localhost:8080/libraries/1/media
```

### 라이브러리 생성

```bash
curl -X POST http://localhost:8080/libraries \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{"name":"Movies","folder_path":"D:/media/movies"}'
```

### 라이브러리 삭제

```bash
curl -X DELETE http://localhost:8080/libraries/1 \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### 회원가입

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"user1","password":"password123"}'
```

### 로그인

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user1","password":"password123"}'
```

### 로그인 후 토큰으로 라이브러리 생성

```bash
curl -X POST http://localhost:8080/libraries \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
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
- `POST /libraries` 중복 경로(409) 처리 테스트
- `DELETE /libraries/:id`의 400/404/200 분기 테스트
- `POST /auth/register` 요청 본문/중복 username 처리/해시 저장 테스트
- `POST /auth/login` 인증 실패/성공(JWT 발급) 테스트
- JWT 인증 미들웨어의 허용/거부 동작 테스트

### 현재 지원 API

- 공용
- `GET /health`
- `GET /health/db`
- `POST /auth/register`
- `POST /auth/login`
- `GET /libraries`
- `GET /libraries/:id`
- `GET /libraries/:id/media`
- 로그인 사용자(JWT 필요)
- `PATCH /me/password`
- 관리자 전용(JWT + admin)
- `POST /libraries`
- `DELETE /libraries/:id`
- `POST /admin/users`
- `GET /admin/users`
- `PATCH /admin/users/:id/admin`
- `DELETE /admin/users/:id`

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

`GET /libraries/:id/media`

```json
[
  {
    "id": 1,
    "library_id": 1,
    "title": "Inception",
    "file_path": "D:/media/movies/Inception.mkv",
    "container": "mkv",
    "duration_seconds": 8880
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

로그인 없이 요청 시 예시(HTTP 401):

```json
{
  "message": "missing authorization header"
}
```

폴더 경로 중복 에러(HTTP 409):

```json
{
  "message": "folder_path already exists"
}
```

`DELETE /libraries/:id`

성공 시:

```json
{
  "message": "library deleted"
}
```

실패 시 예시(잘못된 id, HTTP 400):

```json
{
  "message": "library id must be a number"
}
```

실패 시 예시(대상 없음, HTTP 404):

```json
{
  "message": "library not found"
}
```

`POST /auth/register`

성공 시:

```json
{
  "message": "user registered"
}
```

실패 시 예시(중복 username, HTTP 409):

```json
{
  "message": "username already exists"
}
```

`POST /auth/login`

성공 시:

```json
{
  "token": "<JWT_TOKEN>",
  "expires_at": 1760000000,
  "token_type": "Bearer",
  "username": "user1",
  "is_admin": false
}
```

실패 시 예시(인증 실패, HTTP 401):

```json
{
  "message": "invalid username or password"
}
```

### `POST /libraries` 요청 규칙

- 요청은 JSON이어야 합니다.
- `name`, `folder_path`는 필수입니다.
- 앞뒤 공백은 제거한 뒤 검사합니다.
- 정의되지 않은 추가 필드가 들어오면 거부합니다.
- 중복된 `folder_path`는 DB 저장 과정에서 실패할 수 있습니다.
- `Authorization: Bearer <JWT_TOKEN>` 헤더가 필요합니다.

### `DELETE /libraries/:id` 요청 규칙

- `id`는 양의 정수여야 합니다.
- 숫자가 아니거나 0 이하이면 `400`을 반환합니다.
- 삭제 대상이 없으면 `404`를 반환합니다.
- 삭제에 성공하면 `200`을 반환합니다.
- `Authorization: Bearer <JWT_TOKEN>` 헤더가 필요합니다.
- 관리자 권한(`is_admin=true`)이 필요합니다.

### `PATCH /me/password` 요청 규칙

- `Authorization: Bearer <JWT_TOKEN>` 헤더가 필요합니다.
- 요청 본문에 `current_password`, `new_password`가 모두 필요합니다.
- 현재 비밀번호 검증이 성공해야 변경됩니다.

### 인증 요청 규칙

- `POST /auth/register`, `POST /auth/login` 요청 본문은 JSON이어야 합니다.
- `username`, `password`는 필수이며 앞뒤 공백을 제거한 뒤 검사합니다.
- 로그인 성공 시 JWT 토큰을 반환합니다.
- JWT는 기본 24시간 동안 유효합니다.
- 서버 환경변수 `JWT_SECRET`이 없으면 개발용 기본 시크릿을 사용합니다.
- 서버 시작 시 기본 관리자 계정 `admin/admin`을 보장하지만, 이미 계정이 있으면 비밀번호는 덮어쓰지 않습니다.

### 권한 모델

- `general`: 라이브러리 읽기(조회) 전용
- `admin`: 라이브러리 쓰기(생성/삭제), 사용자 생성/권한 부여/삭제 가능

### CORS 설정

- 프론트엔드 개발 서버(`http://localhost:5173`)에서의 API 호출을 허용합니다.
- 허용 메서드: `GET, POST, PATCH, DELETE, OPTIONS`
- 허용 헤더: `Content-Type, Authorization`
