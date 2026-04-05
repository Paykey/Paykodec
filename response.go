package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type healthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type messageResponse struct {
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	// 응답 형식을 JSON으로 지정하고 상태 코드를 먼저 쓴다.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	// 어떤 구조체든 JSON으로 인코딩해서 응답 본문에 기록한다.
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("error: failed to write JSON response:", err)
	}
}
