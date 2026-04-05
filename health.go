package main

import (
	"database/sql"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// /health는 GET 요청만 허용한다.
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, healthResponse{
			Status:  "error",
			Message: "only GET is allowed",
		})
		return
	}

	// 서버 프로세스가 살아 있으면 정상 응답을 준다.
	writeJSON(w, http.StatusOK, healthResponse{
		Status:  "ok",
		Message: "server is running",
	})
}

func healthDBHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// /health/db도 GET 요청만 허용한다.
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, healthResponse{
				Status:  "error",
				Message: "only GET is allowed",
			})
			return
		}

		// 요청 시점마다 DB에 Ping을 보내 실제 연결 상태를 확인한다.
		if err := db.Ping(); err != nil {
			writeJSON(w, http.StatusInternalServerError, healthResponse{
				Status:  "error",
				Message: "database connection failed",
			})
			return
		}

		// DB 연결이 정상이면 성공 응답을 반환한다.
		writeJSON(w, http.StatusOK, healthResponse{
			Status:  "ok",
			Message: "database connection is healthy",
		})
	}
}
