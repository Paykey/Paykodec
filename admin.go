package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type userSummary struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	IsAdmin   bool   `json:"is_admin"`
	CreatedAt string `json:"created_at"`
}

type setAdminRequest struct {
	IsAdmin bool `json:"is_admin"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type createUserByAdminRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func adminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, ok := c.Get("is_admin")
		if !ok {
			writeJSON(c, http.StatusForbidden, messageResponse{Message: "admin only"})
			c.Abort()
			return
		}

		adminValue, ok := isAdmin.(bool)
		if !ok || !adminValue {
			writeJSON(c, http.StatusForbidden, messageResponse{Message: "admin only"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func createUserByAdminHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		req, err := decodeCreateUserByAdminReq(c)
		if err != nil {
			writeJSON(c, http.StatusBadRequest, messageResponse{Message: err.Error()})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeJSON(c, http.StatusInternalServerError, messageResponse{Message: "failed to hash password"})
			return
		}

		_, err = db.Exec(`
			INSERT INTO users (username, password_hash, is_admin)
			VALUES ($1, $2, FALSE)
		`, req.Username, string(hashedPassword))
		if err != nil {
			if isUniqueViolation(err) {
				writeJSON(c, http.StatusConflict, messageResponse{Message: "username already exists"})
				return
			}

			writeJSON(c, http.StatusInternalServerError, messageResponse{Message: "failed to create user"})
			return
		}

		writeJSON(c, http.StatusCreated, messageResponse{Message: "user created with general role"})
	}
}

func listUsersHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT id, username, is_admin, created_at::text
			FROM users
			ORDER BY id ASC
		`)
		if err != nil {
			writeJSON(c, http.StatusInternalServerError, messageResponse{Message: "failed to list users"})
			return
		}
		defer rows.Close()

		users := make([]userSummary, 0)
		for rows.Next() {
			var item userSummary
			if err := rows.Scan(&item.ID, &item.Username, &item.IsAdmin, &item.CreatedAt); err != nil {
				writeJSON(c, http.StatusInternalServerError, messageResponse{Message: "failed to scan users"})
				return
			}
			users = append(users, item)
		}

		if err := rows.Err(); err != nil {
			writeJSON(c, http.StatusInternalServerError, messageResponse{Message: "failed while reading users"})
			return
		}

		writeJSON(c, http.StatusOK, users)
	}
}

func setUserAdminHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetID, err := parsePositiveID(c.Param("id"), "user id")
		if err != nil {
			writeJSON(c, http.StatusBadRequest, messageResponse{Message: err.Error()})
			return
		}

		req, err := decodeSetAdminReq(c)
		if err != nil {
			writeJSON(c, http.StatusBadRequest, messageResponse{Message: err.Error()})
			return
		}

		if targetID == 1 && !req.IsAdmin {
			writeJSON(c, http.StatusBadRequest, messageResponse{Message: "default admin cannot lose admin role"})
			return
		}

		result, err := db.Exec(`
			UPDATE users
			SET is_admin = $1
			WHERE id = $2
		`, req.IsAdmin, targetID)
		if err != nil {
			writeJSON(c, http.StatusInternalServerError, messageResponse{Message: "failed to update user role"})
			return
		}

		affected, err := result.RowsAffected()
		if err != nil {
			writeJSON(c, http.StatusInternalServerError, messageResponse{Message: "failed to confirm update result"})
			return
		}
		if affected == 0 {
			writeJSON(c, http.StatusNotFound, messageResponse{Message: "user not found"})
			return
		}

		writeJSON(c, http.StatusOK, messageResponse{Message: "user role updated"})
	}
}

func deleteUserHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetID, err := parsePositiveID(c.Param("id"), "user id")
		if err != nil {
			writeJSON(c, http.StatusBadRequest, messageResponse{Message: err.Error()})
			return
		}

		currentUserIDRaw, exists := c.Get("user_id")
		if !exists {
			writeJSON(c, http.StatusUnauthorized, messageResponse{Message: "missing user context"})
			return
		}
		currentUserID, ok := currentUserIDRaw.(int)
		if !ok {
			writeJSON(c, http.StatusUnauthorized, messageResponse{Message: "invalid user context"})
			return
		}

		if targetID == currentUserID {
			writeJSON(c, http.StatusBadRequest, messageResponse{Message: "cannot delete current login account"})
			return
		}
		if targetID == 1 {
			writeJSON(c, http.StatusBadRequest, messageResponse{Message: "default admin cannot be deleted"})
			return
		}

		result, err := db.Exec(`
			DELETE FROM users
			WHERE id = $1
		`, targetID)
		if err != nil {
			writeJSON(c, http.StatusInternalServerError, messageResponse{Message: "failed to delete user"})
			return
		}

		affected, err := result.RowsAffected()
		if err != nil {
			writeJSON(c, http.StatusInternalServerError, messageResponse{Message: "failed to confirm delete result"})
			return
		}
		if affected == 0 {
			writeJSON(c, http.StatusNotFound, messageResponse{Message: "user not found"})
			return
		}

		writeJSON(c, http.StatusOK, messageResponse{Message: "user deleted"})
	}
}

func changeOwnPasswordHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDRaw, exists := c.Get("user_id")
		if !exists {
			writeJSON(c, http.StatusUnauthorized, messageResponse{Message: "missing user context"})
			return
		}
		userID, ok := userIDRaw.(int)
		if !ok {
			writeJSON(c, http.StatusUnauthorized, messageResponse{Message: "invalid user context"})
			return
		}

		req, err := decodeChangePasswordReq(c)
		if err != nil {
			writeJSON(c, http.StatusBadRequest, messageResponse{Message: err.Error()})
			return
		}

		var currentHash string
		err = db.QueryRow(`
			SELECT password_hash
			FROM users
			WHERE id = $1
		`, userID).Scan(&currentHash)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSON(c, http.StatusNotFound, messageResponse{Message: "user not found"})
				return
			}
			writeJSON(c, http.StatusInternalServerError, messageResponse{Message: "failed to load current password"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.CurrentPassword)); err != nil {
			writeJSON(c, http.StatusUnauthorized, messageResponse{Message: "current password is incorrect"})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			writeJSON(c, http.StatusInternalServerError, messageResponse{Message: "failed to hash password"})
			return
		}

		result, err := db.Exec(`
			UPDATE users
			SET password_hash = $1
			WHERE id = $2
		`, string(hashedPassword), userID)
		if err != nil {
			writeJSON(c, http.StatusInternalServerError, messageResponse{Message: "failed to update password"})
			return
		}

		affected, err := result.RowsAffected()
		if err != nil {
			writeJSON(c, http.StatusInternalServerError, messageResponse{Message: "failed to confirm update result"})
			return
		}
		if affected == 0 {
			writeJSON(c, http.StatusNotFound, messageResponse{Message: "user not found"})
			return
		}

		writeJSON(c, http.StatusOK, messageResponse{Message: "password updated"})
	}
}

func decodeSetAdminReq(c *gin.Context) (setAdminRequest, error) {
	var req setAdminRequest
	rawBody, err := c.GetRawData()
	if err != nil {
		return setAdminRequest{}, errors.New("failed to read request body")
	}

	decoder := json.NewDecoder(strings.NewReader(string(rawBody)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return setAdminRequest{}, errors.New("request body must be valid JSON")
	}

	return req, nil
}

func decodeCreateUserByAdminReq(c *gin.Context) (createUserByAdminRequest, error) {
	var req createUserByAdminRequest
	rawBody, err := c.GetRawData()
	if err != nil {
		return createUserByAdminRequest{}, errors.New("failed to read request body")
	}

	decoder := json.NewDecoder(strings.NewReader(string(rawBody)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return createUserByAdminRequest{}, errors.New("request body must be valid JSON")
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	if req.Username == "" || req.Password == "" {
		return createUserByAdminRequest{}, errors.New("username and password are required")
	}

	return req, nil
}

func decodeChangePasswordReq(c *gin.Context) (changePasswordRequest, error) {
	var req changePasswordRequest
	rawBody, err := c.GetRawData()
	if err != nil {
		return changePasswordRequest{}, errors.New("failed to read request body")
	}

	decoder := json.NewDecoder(strings.NewReader(string(rawBody)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return changePasswordRequest{}, errors.New("request body must be valid JSON")
	}

	req.CurrentPassword = strings.TrimSpace(req.CurrentPassword)
	req.NewPassword = strings.TrimSpace(req.NewPassword)
	if req.CurrentPassword == "" {
		return changePasswordRequest{}, errors.New("current_password is required")
	}
	if req.NewPassword == "" {
		return changePasswordRequest{}, errors.New("new_password is required")
	}

	return req, nil
}

func parsePositiveID(raw string, fieldName string) (int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, errors.New(fieldName + " is required")
	}

	id, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New(fieldName + " must be a number")
	}
	if id <= 0 {
		return 0, errors.New(fieldName + " must be greater than zero")
	}

	return id, nil
}
