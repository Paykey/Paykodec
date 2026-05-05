package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func TestDecodeRegisterReq(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(`{"username":" user1 ","password":" secret "}`),
	)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	got, err := decodeRegisterReq(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.Username != "user1" {
		t.Fatalf("expected trimmed username, got %q", got.Username)
	}

	if got.Password != "secret" {
		t.Fatalf("expected trimmed password, got %q", got.Password)
	}
}

func TestRegisterRejectsInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(`{"username":`))
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	handler := registerHandler(&fakeUserCreator{})
	handler(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestRegisterRejectsMissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(`{"username":"user1","password":"  "}`),
	)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	handler := registerHandler(&fakeUserCreator{})
	handler(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestRegisterReturnsConflictOnDuplicateUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(`{"username":"user1","password":"secret"}`),
	)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	handler := registerHandler(&fakeUserCreator{
		err: &pq.Error{Code: "23505"},
	})
	handler(c)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestRegisterStoresHashedPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(`{"username":"user1","password":"secret"}`),
	)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	spy := fakeUserCreator{}
	handler := registerHandler(&spy)
	handler(c)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if spy.gotUsername != "user1" {
		t.Fatalf("expected username user1, got %q", spy.gotUsername)
	}

	if spy.gotPassword == "secret" {
		t.Fatal("expected hashed password, got plain text")
	}

	if !isBcryptHash(spy.gotPassword) {
		t.Fatalf("expected bcrypt hash format, got %q", spy.gotPassword)
	}
}

func TestDecodeLoginReq(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(`{"username":" user1 ","password":" secret "}`),
	)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	got, err := decodeLoginReq(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.Username != "user1" {
		t.Fatalf("expected trimmed username, got %q", got.Username)
	}

	if got.Password != "secret" {
		t.Fatalf("expected trimmed password, got %q", got.Password)
	}
}

func TestLoginReturnsUnauthorizedWhenUserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(`{"username":"user1","password":"secret"}`),
	)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	handler := loginHandlerWithDeps(
		func(username string) (int, string, bool, error) {
			return 0, "", false, sql.ErrNoRows
		},
		func(userID int, username string, isAdmin bool) (string, int64, error) {
			return "", 0, nil
		},
	)
	handler(c)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestLoginReturnsTokenOnSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-secret")

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to prepare hash: %v", err)
	}
	hashed := string(hashedBytes)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(`{"username":"user1","password":"secret"}`),
	)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	handler := loginHandlerWithDeps(
		func(username string) (int, string, bool, error) {
			return 7, hashed, true, nil
		},
		generateJWT,
	)
	handler(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body authResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Token == "" {
		t.Fatal("expected token to be set")
	}

	if body.TokenType != "Bearer" {
		t.Fatalf("expected token type Bearer, got %q", body.TokenType)
	}

	if !body.IsAdmin {
		t.Fatal("expected is_admin true in login response")
	}
}

func TestAuthMiddlewareRejectsMissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(authMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		writeJSON(c, http.StatusOK, messageResponse{Message: "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthMiddlewareAcceptsValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-secret")
	token, _, err := generateJWT(3, "user1", false)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	router := gin.New()
	router.Use(authMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		writeJSON(c, http.StatusOK, messageResponse{Message: "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

type fakeUserCreator struct {
	err error

	gotUsername string
	gotPassword string
}

func (f *fakeUserCreator) Exec(query string, args ...any) (sql.Result, error) {
	if len(args) >= 2 {
		if username, ok := args[0].(string); ok {
			f.gotUsername = username
		}
		if password, ok := args[1].(string); ok {
			f.gotPassword = password
		}
	}

	if f.err != nil {
		return nil, f.err
	}

	return fakeSQLResult{rows: 1}, nil
}
