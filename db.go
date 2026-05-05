package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func loadEnv() {
	// .env 파일을 읽어서 DB 접속 정보 같은 환경변수를 등록한다.
	if err := godotenv.Load(); err != nil {
		log.Fatal("error: .env file not found.")
	}
}

func openDB() (*sql.DB, error) {
	// 환경변수에서 DB 사용자명과 비밀번호를 가져온다.
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	// PostgreSQL 접속 문자열을 만든다.
	connStr := fmt.Sprintf(
		"host=localhost port=5432 user=%s password=%s dbname=media_db sslmode=disable",
		dbUser,
		dbPassword,
	)

	// sql.Open은 연결 정보를 가진 DB 객체를 준비한다.
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare DB connection: %w", err)
	}

	// Ping으로 실제 연결이 가능한지 한 번 확인한다.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	return db, nil
}

func runInitSQL(db *sql.DB) error {
	// 테이블 생성 SQL 파일을 읽어온다.
	sqlBytes, err := os.ReadFile("sql/init.sql")
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	// 읽어온 SQL을 실행해서 필요한 테이블을 만든다.
	if _, err := db.Exec(string(sqlBytes)); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

func ensureDefaultAdmin(db *sql.DB) error {
	const defaultAdminUsername = "admin"
	const defaultAdminPassword = "admin"

	var existingID int
	err := db.QueryRow(`
		SELECT id
		FROM users
		WHERE username = $1
	`, defaultAdminUsername).Scan(&existingID)
	if err == nil {
		// 이미 admin 계정이 있으면 초기 비밀번호를 다시 덮어쓰지 않는다.
		_, promoteErr := db.Exec(`
			UPDATE users
			SET is_admin = TRUE
			WHERE id = $1
		`, existingID)
		if promoteErr != nil {
			return fmt.Errorf("failed to ensure admin role for default admin: %w", promoteErr)
		}
		return nil
	}
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check default admin user: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash default admin password: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO users (username, password_hash, is_admin)
		VALUES ($1, $2, TRUE)
	`, defaultAdminUsername, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to ensure default admin user: %w", err)
	}

	return nil
}
