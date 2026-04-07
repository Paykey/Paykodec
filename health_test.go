package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthHandlerReturnsOK(t *testing.T) {
	// 테스트에서는 디버그 로그를 줄이기 위해 테스트 모드를 사용한다.
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", healthHandler)

	// 테스트용 GET 요청과 응답 기록기를 준비한다.
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// 실제 Gin 라우터를 통해 핸들러를 호출한다.
	router.ServeHTTP(rec, req)

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

func TestHealthHandlerRejectsPostByRoute(t *testing.T) {
	// GET만 등록된 라우트에 POST를 보내면 Gin이 404를 반환한다.
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", healthHandler)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestHealthDBHandlerReturnsOK(t *testing.T) {
	// DB Ping 성공 상황을 가짜 객체로 만들어 200 응답을 확인한다.
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health/db", healthDBHandler(fakeDBPinger{}))

	req := httptest.NewRequest(http.MethodGet, "/health/db", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("expected status field to be ok, got %q", body.Status)
	}
}

func TestHealthDBHandlerReturnsInternalServerError(t *testing.T) {
	// DB Ping 실패 상황을 가짜 객체로 만들어 500 응답을 확인한다.
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health/db", healthDBHandler(fakeDBPinger{err: errors.New("db down")}))

	req := httptest.NewRequest(http.MethodGet, "/health/db", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

type fakeDBPinger struct {
	err error
}

func (f fakeDBPinger) Ping() error {
	return f.err
}
