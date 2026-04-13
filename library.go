package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
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

type libraryCreator interface {
	Exec(query string, args ...any) (sql.Result, error)
}

func listLibrariesHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// name 쿼리 파라미터로 라이브러리 이름 부분 검색을 지원한다.
		// 예: /libraries?name=movie
		nameQuery := strings.TrimSpace(c.Query("name"))

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
			writeJSON(c, 500, messageResponse{
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
				writeJSON(c, 500, messageResponse{
					Message: "failed to scan library row",
				})
				return
			}

			libraries = append(libraries, item)
		}

		// 반복 중에 발생한 숨은 에러가 없는지 마지막으로 확인한다.
		if err := rows.Err(); err != nil {
			writeJSON(c, 500, messageResponse{
				Message: "failed while reading library rows",
			})
			return
		}

		// 최종 목록을 JSON 배열로 응답한다.
		writeJSON(c, 200, libraries)
	}
}

func createLibraryHandler(db libraryCreator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 요청 본문 JSON을 읽고, 공백 제거와 필수값 검사를 함께 수행한다.
		req, err := decodeCreateLibraryReq(c)
		if err != nil {
			writeJSON(c, 400, messageResponse{
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
			// PostgreSQL unique violation(23505)은 중복 데이터 요청이므로 409로 응답한다.
			if isUniqueViolation(err) {
				writeJSON(c, http.StatusConflict, messageResponse{
					Message: "folder_path already exists",
				})
				return
			}

			writeJSON(c, http.StatusInternalServerError, messageResponse{
				Message: "failed to create library",
			})
			return
		}

		// 등록이 끝나면 생성 성공 상태 코드와 메시지를 반환한다.
		writeJSON(c, http.StatusCreated, messageResponse{
			Message: "library created",
		})
	}
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func decodeCreateLibraryReq(c *gin.Context) (createLibraryRequest, error) {
	var req createLibraryRequest
	rawBody, err := c.GetRawData()
	if err != nil {
		return createLibraryRequest{}, errors.New("failed to read request body")
	}

	// JSON 디코더를 사용해 요청 본문을 구조체로 변환한다.
	decoder := json.NewDecoder(strings.NewReader(string(rawBody)))

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
