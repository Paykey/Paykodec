package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type library struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	FolderPath string `json:"folder_path"`
}

type createLibraryRequest struct {
	Name       string `json:"name"`
	FolderPath string `json:"folder_path"`
}

func handleLibraries(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// /libraries는 메서드에 따라 목록 조회와 등록 기능으로 나눈다.
		switch r.Method {
		case http.MethodGet:
			listLibrariesHandler(db, w, r)
		case http.MethodPost:
			createLibraryHandler(db, w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, messageResponse{
				Message: "only GET or POST is allowed",
			})
		}
	}
}

func listLibrariesHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// name 쿼리 파라미터로 라이브러리 이름 부분 검색을 지원한다.
	// 예: /libraries?name=movie
	nameQuery := strings.TrimSpace(r.URL.Query().Get("name"))

	// 공통 SELECT 구문을 분리해두면 조건별 분기가 읽기 쉽다.
	baseQuery := `
		SELECT id, name, folder_path
		FROM libraries
	`

	var (
		rows *sql.Rows
		err  error
	)

	if nameQuery == "" {
		// 검색어가 없으면 전체 목록을 ID 순서로 조회한다.
		rows, err = db.Query(baseQuery + ` ORDER BY id ASC`)
	} else {
		// 검색어가 있으면 ILIKE로 대소문자 구분 없이 부분 검색한다.
		// $1 같은 바인딩을 사용하면 SQL injection 위험을 줄일 수 있다.
		rows, err = db.Query(baseQuery+`
			WHERE name ILIKE $1
			ORDER BY id ASC
		`, "%"+nameQuery+"%")
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, messageResponse{
			Message: "failed to fetch libraries",
		})
		return
	}
	defer rows.Close()

	// 조회 결과를 담을 슬라이스를 준비한다.
	libraries := make([]library, 0)

	for rows.Next() {
		var item library

		// 현재 행의 컬럼 값을 구조체 필드에 읽어 넣는다.
		if err := rows.Scan(&item.ID, &item.Name, &item.FolderPath); err != nil {
			writeJSON(w, http.StatusInternalServerError, messageResponse{
				Message: "failed to scan library row",
			})
			return
		}

		libraries = append(libraries, item)
	}

	// 반복 중에 발생한 숨은 에러가 없는지 마지막으로 확인한다.
	if err := rows.Err(); err != nil {
		writeJSON(w, http.StatusInternalServerError, messageResponse{
			Message: "failed while reading library rows",
		})
		return
	}

	// 최종 목록을 JSON 배열로 응답한다.
	writeJSON(w, http.StatusOK, libraries)
}

func createLibraryHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// 요청 본문 JSON을 읽고, 공백 제거와 필수값 검사를 함께 수행한다.
	req, err := decodeCreateLibraryRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, messageResponse{
			Message: err.Error(),
		})
		return
	}

	// 라이브러리 정보를 DB에 저장한다.
	_, err = db.Exec(`
		INSERT INTO libraries (name, folder_path)
		VALUES ($1, $2)
	`, req.Name, req.FolderPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, messageResponse{
			Message: "failed to create library; check for duplicate folder paths",
		})
		return
	}

	// 등록이 끝나면 생성 성공 상태 코드와 메시지를 반환한다.
	writeJSON(w, http.StatusCreated, messageResponse{
		Message: "library created",
	})
}

func decodeCreateLibraryRequest(r *http.Request) (createLibraryRequest, error) {
	// 요청 처리가 끝나면 Body를 닫아 리소스를 정리한다.
	defer r.Body.Close()

	var req createLibraryRequest

	// JSON 디코더를 사용해 요청 본문을 구조체로 변환한다.
	decoder := json.NewDecoder(r.Body)

	// 정의하지 않은 필드가 들어오면 에러를 내도록 설정한다.
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		return createLibraryRequest{}, errors.New("request body must be valid JSON")
	}

	// 앞뒤 공백만 있는 입력도 빈 값처럼 처리하기 위해 trim을 적용한다.
	req.Name = strings.TrimSpace(req.Name)
	req.FolderPath = strings.TrimSpace(req.FolderPath)

	// 필수값이 비어 있으면 잘못된 요청으로 처리한다.
	if req.Name == "" || req.FolderPath == "" {
		return createLibraryRequest{}, errors.New("name and folder_path are required")
	}

	return req, nil
}
