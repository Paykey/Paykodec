package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDecodeCreateLibraryRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 공백이 포함된 정상 JSON을 넣고 trim 처리 결과를 확인한다.
	req := httptest.NewRequest(
		http.MethodPost,
		"/libraries",
		strings.NewReader(`{"name":" Movies ","folder_path":" D:/media/movies "}`),
	)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	got, err := decodeCreateLibraryRequest(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.Name != "Movies" {
		t.Fatalf("expected trimmed name, got %q", got.Name)
	}

	if got.FolderPath != "D:/media/movies" {
		t.Fatalf("expected trimmed folder path, got %q", got.FolderPath)
	}
}

func TestDecodeCreateLibraryRequestRejectsUnknownField(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 정의되지 않은 extra 필드가 있으면 거부해야 한다.
	req := httptest.NewRequest(
		http.MethodPost,
		"/libraries",
		strings.NewReader(`{"name":"Movies","folder_path":"D:/media/movies","extra":"x"}`),
	)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	_, err := decodeCreateLibraryRequest(c)
	if err == nil {
		t.Fatal("expected an error for unknown field")
	}
}

func TestCreateLibraryHandlerRejectsInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 문법이 깨진 JSON이면 DB에 닿기 전에 400을 반환해야 한다.
	req := httptest.NewRequest(http.MethodPost, "/libraries", strings.NewReader(`{"name":`))
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	handler := createLibraryHandler(nil)
	handler(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateLibraryHandlerRejectsMissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 필수값이 비어 있어도 잘못된 요청으로 처리해야 한다.
	req := httptest.NewRequest(
		http.MethodPost,
		"/libraries",
		strings.NewReader(`{"name":"Movies","folder_path":"   "}`),
	)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	handler := createLibraryHandler(nil)
	handler(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}
