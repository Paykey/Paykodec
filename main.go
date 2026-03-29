package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const serverAddr = ":8080"

type healthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type errorResponse struct {
	Message string `json:"message"`
}

type library struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	FolderPath string `json:"folder_path"`
}

type createLibraryRequest struct {
	Name       string `json:"name"`
	FolderPath string `json:"folder_path"`
}

func loadEnv() {
	// .env 파일을 읽어서 환경변수로 등록한다.
	if err := godotenv.Load(); err != nil {
		log.Fatal("오류: .env 파일을 찾을 수 없습니다.")
	}
}

func openDB() (*sql.DB, error) {
	// 환경변수에서 데이터베이스 접속 정보를 읽는다.
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	// PostgreSQL 접속 문자열을 만든다.
	connStr := fmt.Sprintf(
		"host=localhost port=5432 user=%s password=%s dbname=media_db sslmode=disable",
		dbUser,
		dbPassword,
	)

	// 데이터베이스 연결 객체를 만든다.
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("DB 연결 준비 실패: %w", err)
	}

	// 실제로 데이터베이스와 통신이 되는지 확인한다.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("DB 연결 확인 실패: %w", err)
	}

	return db, nil
}

func runInitSQL(db *sql.DB) error {
	// 테이블 생성 SQL 파일을 읽는다.
	sqlBytes, err := os.ReadFile("sql/init.sql")
	if err != nil {
		return fmt.Errorf("SQL 파일 읽기 실패: %w", err)
	}

	// 읽어온 SQL을 실행해서 필요한 테이블을 만든다.
	if _, err := db.Exec(string(sqlBytes)); err != nil {
		return fmt.Errorf("테이블 생성 실패: %w", err)
	}

	return nil
}

func writeJSON(w http.ResponseWriter, statusCode int, data healthResponse) {
	// 응답 형식을 JSON으로 지정한다.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	// JSON 인코딩 결과를 응답으로 보낸다.
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("오류: JSON 응답 작성 실패:", err)
	}
}

func writeJSONMessage(w http.ResponseWriter, statusCode int, data errorResponse) {
	// 간단한 메시지 응답을 JSON으로 보낸다.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	// JSON 인코딩 결과를 응답으로 보낸다.
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("오류: JSON 응답 작성 실패:", err)
	}
}

func writeJSONLibraries(w http.ResponseWriter, statusCode int, data []library) {
	// 라이브러리 목록 응답을 JSON으로 보낸다.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	// JSON 인코딩 결과를 응답으로 보낸다.
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("오류: JSON 응답 작성 실패:", err)
	}
}

func handleLibraries(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 요청 메서드에 따라 목록 조회와 등록 기능을 나눈다.
		switch r.Method {
		case http.MethodGet:
			listLibrariesHandler(db, w)
		case http.MethodPost:
			createLibraryHandler(db, w, r)
		default:
			writeJSONMessage(w, http.StatusMethodNotAllowed, errorResponse{
				Message: "GET 또는 POST 요청만 허용됩니다.",
			})
		}
	}
}

func listLibrariesHandler(db *sql.DB, w http.ResponseWriter) {
	// 데이터베이스에서 라이브러리 목록을 조회한다.
	rows, err := db.Query(`
		SELECT id, name, folder_path
		FROM libraries
		ORDER BY id ASC
	`)
	if err != nil {
		writeJSONMessage(w, http.StatusInternalServerError, errorResponse{
			Message: "라이브러리 목록 조회에 실패했습니다.",
		})
		return
	}
	defer rows.Close()

	// 조회한 결과를 슬라이스에 차례대로 담는다.
	libraries := make([]library, 0)

	for rows.Next() {
		var item library

		if err := rows.Scan(&item.ID, &item.Name, &item.FolderPath); err != nil {
			writeJSONMessage(w, http.StatusInternalServerError, errorResponse{
				Message: "라이브러리 데이터 읽기에 실패했습니다.",
			})
			return
		}

		libraries = append(libraries, item)
	}

	// 반복 처리 중 생긴 에러가 있는지 마지막으로 확인한다.
	if err := rows.Err(); err != nil {
		writeJSONMessage(w, http.StatusInternalServerError, errorResponse{
			Message: "라이브러리 목록 처리에 실패했습니다.",
		})
		return
	}

	writeJSONLibraries(w, http.StatusOK, libraries)
}

func createLibraryHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// 요청 본문에서 JSON 데이터를 읽는다.
	var req createLibraryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONMessage(w, http.StatusBadRequest, errorResponse{
			Message: "JSON 형식이 올바르지 않습니다.",
		})
		return
	}

	// 빈 값이 들어오지 않도록 가장 기본적인 검사를 한다.
	if req.Name == "" || req.FolderPath == "" {
		writeJSONMessage(w, http.StatusBadRequest, errorResponse{
			Message: "name과 folder_path는 필수입니다.",
		})
		return
	}

	// 라이브러리 정보를 데이터베이스에 저장한다.
	_, err := db.Exec(`
		INSERT INTO libraries (name, folder_path)
		VALUES ($1, $2)
	`, req.Name, req.FolderPath)
	if err != nil {
		writeJSONMessage(w, http.StatusInternalServerError, errorResponse{
			Message: "라이브러리 등록에 실패했습니다. 같은 경로가 이미 존재하는지 확인하세요.",
		})
		return
	}

	writeJSONMessage(w, http.StatusCreated, errorResponse{
		Message: "라이브러리가 등록되었습니다.",
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// GET 요청만 허용한다.
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, healthResponse{
			Status:  "error",
			Message: "GET 요청만 허용됩니다.",
		})
		return
	}

	writeJSON(w, http.StatusOK, healthResponse{
		Status:  "ok",
		Message: "서버가 정상적으로 실행 중입니다.",
	})
}

func healthDBHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// GET 요청만 허용한다.
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, healthResponse{
				Status:  "error",
				Message: "GET 요청만 허용됩니다.",
			})
			return
		}

		// 요청 시점에 데이터베이스 연결 상태를 다시 확인한다.
		if err := db.Ping(); err != nil {
			writeJSON(w, http.StatusInternalServerError, healthResponse{
				Status:  "error",
				Message: "데이터베이스 연결에 실패했습니다.",
			})
			return
		}

		writeJSON(w, http.StatusOK, healthResponse{
			Status:  "ok",
			Message: "데이터베이스 연결이 정상입니다.",
		})
	}
}

func main() {
	// 프로그램 시작 전에 환경변수를 불러온다.
	loadEnv()

	// 데이터베이스에 연결한다.
	db, err := openDB()
	if err != nil {
		log.Fatal("오류:", err)
	}
	defer db.Close()

	// 프로그램 실행 시 필요한 테이블을 생성한다.
	if err := runInitSQL(db); err != nil {
		log.Fatal("오류:", err)
	}

	// 상태 확인용 API를 등록한다.
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/health/db", healthDBHandler(db))
	http.HandleFunc("/libraries", handleLibraries(db))

	log.Println("서버가 시작되었습니다:", serverAddr)
	log.Println("헬스 체크 주소: http://localhost" + serverAddr + "/health")
	log.Println("DB 체크 주소: http://localhost" + serverAddr + "/health/db")
	log.Println("라이브러리 목록 주소: http://localhost" + serverAddr + "/libraries")

	// 8080 포트에서 HTTP 서버를 실행한다.
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatal("오류: HTTP 서버 실행 실패:", err)
	}
}
