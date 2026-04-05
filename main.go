package main

import (
	"log"
	"net/http"
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

	// 상태 확인용 API와 라이브러리 API를 라우트에 등록한다.
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/health/db", healthDBHandler(db))
	http.HandleFunc("/libraries", handleLibraries(db))

	// 서버 실행 전에 접속 가능한 주소를 로그로 남긴다.
	log.Println("server started on", serverAddr)
	log.Println("health check: http://localhost" + serverAddr + "/health")
	log.Println("db health check: http://localhost" + serverAddr + "/health/db")
	log.Println("libraries: http://localhost" + serverAddr + "/libraries")

	// 8080 포트에서 HTTP 서버를 시작한다.
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatal("error: HTTP server failed:", err)
	}
}
