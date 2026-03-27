package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv" // .env 파일을 Go에서 쉽게 읽게 해주는 외부 도구
	_ "github.com/lib/pq"      // PostgreSQL과 통신하기 위한 드라이버
)

func main() {
	err := godotenv.Load()	// .env 파일을 읽어서 환경 변수로 설정

	// 만약 .env 파일을 읽는 데 실패했다면 (파일이 없거나 오타가 났다면)
	if err != nil {
		log.Fatal("Error: .env 파일을 찾을 수 없습니다.")	// 에러 메시지를 출력하고 프로그램을 즉시 멈춥니다.
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	
	// PostgreSQL 접속 주소 조립
	connStr := fmt.Sprintf("host=localhost port=5432 user=%s password=%s dbname=media_db sslmode=disable", dbUser, dbPassword)

	// 조립한 주소로 데이터베이스에 접속 준비
	db, err := sql.Open("postgres", connStr)

	// 준비 과정에서 에러시 프로그램 중단
	if err != nil {
		log.Fatal("Error: DB 접속 설정 오류:", err)
	}
	
	// defer: 메인 함수가 완전히 끝나기 직전에 DB 연결을 닫도록 예약하는 키워드
	defer db.Close()

	// 실제 DB 서버와 연결이 되는지 확인
	err = db.Ping()

	// 도커가 켜져 있지 않거나 접속 정보가 틀렸다면 에러가 나고 프로그램 중단
	if err != nil {
		log.Fatal("Error: DB 연결 실패! 도커로 DB를 켰는지 확인하세요:", err)
	}

	// 연결에 성공했다면 화면에 성공 메시지를 출력
	fmt.Println("Success: DB 연결 성공! 이제 젤리핀 구조의 테이블을 생성합니다...")

	// os.ReadFile을 사용해 분리해 둔 init.sql 파일의 내용을 읽어옴
	// 읽어온 데이터는 컴퓨터가 이해하는 바이트(byte) 형태로 저장됨
	sqlBytes, err := os.ReadFile("sql/init.sql")
	if err != nil {
		log.Fatal("Error: SQL 파일을 읽어오는 데 실패했습니다:", err)
	}

	// 바이트 덩어리를 문자열로 변환
	createTablesSQL := string(sqlBytes)


	// db.Exec() 함수를 써서 위에 길게 적어둔 SQL 명령어들을 DB로 보내고 실행합니다.
	_, err = db.Exec(createTablesSQL) // _는 버리는 값
	// 테이블 생성 중에 에러가 났다면 (문법 오류 등) 프로그램 중단
	if err != nil {
		log.Fatal("Error: 테이블 생성 실패:", err)
	}

	// 테이블 생성이 끝났다면 최종 성공 메시지를 출력
	fmt.Println("Success: 외부 SQL 파일을 읽어 데이터베이스 테이블 생성을 완료했습니다!")
}