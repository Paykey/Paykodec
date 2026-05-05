package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const tokenTTL = 24 * time.Hour

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	TokenType string `json:"token_type"`
	Username  string `json:"username"`
	IsAdmin   bool   `json:"is_admin"`
}

type userCreator interface {
	Exec(query string, args ...any) (sql.Result, error)
}

type userFetcher interface {
	QueryRow(query string, args ...any) *sql.Row
}

func registerHandler(db userCreator) gin.HandlerFunc {
	return func(c *gin.Context) {
		req, err := decodeRegisterReq(c)
		if err != nil {
			writeJSON(c, http.StatusBadRequest, messageResponse{
				Message: err.Error(),
			})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeJSON(c, http.StatusInternalServerError, messageResponse{
				Message: "failed to hash password",
			})
			return
		}

		_, err = db.Exec(`
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
	`, req.Username, string(hashedPassword))
		if err != nil {
			if isUniqueViolation(err) {
				writeJSON(c, http.StatusConflict, messageResponse{
					Message: "username already exists",
				})
				return
			}

			writeJSON(c, http.StatusInternalServerError, messageResponse{
				Message: "failed to register user",
			})
			return
		}

		writeJSON(c, http.StatusCreated, messageResponse{
			Message: "user registered",
		})
	}
}

func loginHandler(db userFetcher) gin.HandlerFunc {
	return loginHandlerWithDeps(
		func(username string) (int, string, bool, error) {
			return fetchUserCredentialsByUsername(db, username)
		},
		generateJWT,
	)
}

func loginHandlerWithDeps(
	fetchCredentials func(username string) (int, string, bool, error),
	signToken func(userID int, username string, isAdmin bool) (string, int64, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		req, err := decodeLoginReq(c)
		if err != nil {
			writeJSON(c, http.StatusBadRequest, messageResponse{
				Message: err.Error(),
			})
			return
		}

		userID, passwordHash, isAdmin, err := fetchCredentials(req.Username)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSON(c, http.StatusUnauthorized, messageResponse{
					Message: "invalid username or password",
				})
				return
			}

			writeJSON(c, http.StatusInternalServerError, messageResponse{
				Message: "failed to login",
			})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			writeJSON(c, http.StatusUnauthorized, messageResponse{
				Message: "invalid username or password",
			})
			return
		}

		token, expiresAt, err := signToken(userID, req.Username, isAdmin)
		if err != nil {
			writeJSON(c, http.StatusInternalServerError, messageResponse{
				Message: "failed to create token",
			})
			return
		}

		writeJSON(c, http.StatusOK, authResponse{
			Token:     token,
			ExpiresAt: expiresAt,
			TokenType: "Bearer",
			Username:  req.Username,
			IsAdmin:   isAdmin,
		})
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			writeJSON(c, http.StatusUnauthorized, messageResponse{
				Message: "missing authorization header",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeJSON(c, http.StatusUnauthorized, messageResponse{
				Message: "invalid authorization header",
			})
			c.Abort()
			return
		}

		tokenText := strings.TrimSpace(parts[1])
		if tokenText == "" {
			writeJSON(c, http.StatusUnauthorized, messageResponse{
				Message: "missing token",
			})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenText, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(getJWTSecret()), nil
		})
		if err != nil || !token.Valid {
			writeJSON(c, http.StatusUnauthorized, messageResponse{
				Message: "invalid or expired token",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if ok {
			if username, ok := claims["username"].(string); ok {
				c.Set("username", username)
			}
			if userID, ok := claims["uid"].(float64); ok {
				c.Set("user_id", int(userID))
			}
			if isAdmin, ok := claims["is_admin"].(bool); ok {
				c.Set("is_admin", isAdmin)
			}
		}

		c.Next()
	}
}

func decodeRegisterReq(c *gin.Context) (registerRequest, error) {
	var req registerRequest
	rawBody, err := c.GetRawData()
	if err != nil {
		return registerRequest{}, errors.New("failed to read request body")
	}

	decoder := json.NewDecoder(strings.NewReader(string(rawBody)))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		return registerRequest{}, errors.New("request body must be valid JSON")
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" || req.Password == "" {
		return registerRequest{}, errors.New("username and password are required")
	}

	return req, nil
}

func decodeLoginReq(c *gin.Context) (loginRequest, error) {
	var req loginRequest
	rawBody, err := c.GetRawData()
	if err != nil {
		return loginRequest{}, errors.New("failed to read request body")
	}

	decoder := json.NewDecoder(strings.NewReader(string(rawBody)))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		return loginRequest{}, errors.New("request body must be valid JSON")
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" || req.Password == "" {
		return loginRequest{}, errors.New("username and password are required")
	}

	return req, nil
}

func fetchUserCredentialsByUsername(db userFetcher, username string) (int, string, bool, error) {
	var (
		userID       int
		passwordHash string
		isAdmin      bool
	)

	err := db.QueryRow(`
		SELECT id, password_hash, is_admin
		FROM users
		WHERE username = $1
	`, username).Scan(&userID, &passwordHash, &isAdmin)
	if err != nil {
		return 0, "", false, err
	}

	return userID, passwordHash, isAdmin, nil
}

func generateJWT(userID int, username string, isAdmin bool) (string, int64, error) {
	expiresAt := time.Now().Add(tokenTTL).Unix()
	claims := jwt.MapClaims{
		"uid":      userID,
		"username": username,
		"is_admin": isAdmin,
		"exp":      expiresAt,
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(getJWTSecret()))
	if err != nil {
		return "", 0, err
	}

	return signed, expiresAt, nil
}

func getJWTSecret() string {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		return "dev-secret-change-me"
	}
	return secret
}

func isBcryptHash(text string) bool {
	return strings.HasPrefix(text, "$2a$") || strings.HasPrefix(text, "$2b$") || strings.HasPrefix(text, "$2y$")
}
