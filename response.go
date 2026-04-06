package main

import (
	"github.com/gin-gonic/gin"
)

type healthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type messageResponse struct {
	Message string `json:"message"`
}

func writeJSON(c *gin.Context, statusCode int, data any) {
	// Gin의 JSON 응답 기능으로 상태 코드와 본문을 함께 기록한다.
	c.JSON(statusCode, data)
}
