package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// 데이터베이스 핸들러
var db *sql.DB

func main() {
	// Gin 라우터 설정
	r := gin.Default()

	// MySQL 데이터베이스 연결
	if err := connectDatabase(); err != nil {
		log.Fatal("Database connection failed:", err)
	}

	// GET 엔드포인트 정의
	r.GET("/workload_info", handleGetRequest)

	// POST 엔드포인트 정의
	r.POST("/submit", handlePostRequest)

	// 서버 실행
	r.Run("0.0.0.0:8080")
}

// 데이터베이스 연결 함수
func connectDatabase() error {
	dsn := loadDatabaseConfig()
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	// 데이터베이스 연결 확인
	if err = db.Ping(); err != nil {
		return err
	}
	return nil
}

// 환경 변수에서 DB 설정 불러오는 함수
func loadDatabaseConfig() string {
	return os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") +
		"@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" + os.Getenv("DB_NAME")
}
