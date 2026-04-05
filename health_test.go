package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandlerReturnsOK(t *testing.T) {
	// 테스트용 GET 요청과 응답 기록기를 준비한다.
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// 실제 핸들러를 호출해 응답을 기록한다.
	healthHandler(rec, req)

	// 상태 코드와 본문이 기대값과 같은지 확인한다.
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body healthResponse
	// JSON 응답을 다시 구조체로 풀어서 필드 값을 검사한다.
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("expected status field to be ok, got %q", body.Status)
	}
}

func TestHealthHandlerRejectsPost(t *testing.T) {
	// 허용되지 않은 메서드를 보냈을 때 405가 나오는지 확인한다.
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	healthHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}
