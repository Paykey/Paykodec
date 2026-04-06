package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
)

const serverAddr = ":8080"

func main() {
	// 프로그램 시작 전에 .env 파일을 읽어 환경변수를 불러온다.
	loadEnv()

	// 데이터베이스 연결을 열고, 프로그램 종료 시 닫히도록 한다.
	db, err := openDB()
	if err != nil {
		log.Fatal("error:", err)
	}
	defer db.Close()

	// 프로그램 실행에 필요한 테이블을 시작 시점에 준비한다.
	if err := runInitSQL(db); err != nil {
		log.Fatal("error:", err)
	}

	// Gin 엔진을 만들고 상태 확인용 API와 라이브러리 API를 등록한다.
	router := setupRouter(db)

	// 서버 실행 전에 접속 가능한 주소를 로그로 남긴다.
	log.Println("server started on", serverAddr)
	log.Println("health check: http://localhost" + serverAddr + "/health")
	log.Println("db health check: http://localhost" + serverAddr + "/health/db")
	log.Println("libraries: http://localhost" + serverAddr + "/libraries")

	// 8080 포트에서 Gin HTTP 서버를 시작한다.
	if err := router.Run(serverAddr); err != nil {
		log.Fatal("error: Gin server failed:", err)
	}
}

func setupRouter(db *sql.DB) *gin.Engine {
	// Default 엔진은 로거와 복구 미들웨어를 포함한 라우터를 만든다.
	router := gin.Default()

	router.GET("/health", healthHandler)
	router.GET("/health/db", healthDBHandler(db))
	router.GET("/libraries", listLibrariesHandler(db))
	router.POST("/libraries", createLibraryHandler(db))

	return router
}
