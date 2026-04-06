package main

import (
	"github.com/gin-gonic/gin"
)

type dbPinger interface {
	Ping() error
}

func healthHandler(c *gin.Context) {
	// 서버 프로세스가 살아 있으면 정상 응답을 준다.
	writeJSON(c, 200, healthResponse{
		Status:  "ok",
		Message: "server is running",
	})
}

func healthDBHandler(db dbPinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 요청 시점마다 DB에 Ping을 보내 실제 연결 상태를 확인한다.
		if err := db.Ping(); err != nil {
			writeJSON(c, 500, healthResponse{
				Status:  "error",
				Message: "database connection failed",
			})
			return
		}

		// DB 연결이 정상이면 성공 응답을 반환한다.
		writeJSON(c, 200, healthResponse{
			Status:  "ok",
			Message: "database connection is healthy",
		})
	}
}
