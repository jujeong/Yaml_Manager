package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// 데이터베이스 핸들러
var db *sql.DB

func main() {
	// Gin 라우터 설정
	r := gin.Default()
	// 데이터베이스 초기화
	initDatabase()
	// CORS 설정
	r.Use(setupCORS())
	// 라우트 등록
	registerRoutes(r)
	// 서버 실행
	r.Run("0.0.0.0:8080")
}

// 데이터베이스 초기화
func initDatabase() {
	var err error
	dsn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") +
		"@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" + os.Getenv("DB_NAME")

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
}

// CORS 설정 함수
func setupCORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},                      // 허용할 도메인
		AllowMethods:     []string{"GET", "POST", "OPTIONS"}, // 허용할 HTTP 메서드
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "*"},
		AllowCredentials: true, // 쿠키 허용 여부
	})
}

// 라우트 등록 함수
func registerRoutes(r *gin.Engine) {
	// 기본 엔드포인트
	r.GET("/workload_info", handleGetWorkloadinfoRequest)
	r.GET("/strato", handleGetStratoRequest)
	r.POST("/submit", handleSubmitRequest)
	r.POST("/submit_resource", handleSubmitResourceRequest)

	// History 관련 라우트 등록
	RegisterHistoryRoutes(r)
}
